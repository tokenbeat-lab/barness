//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var glmAdd = ai.Tool{Name: "add", Description: "Add the two given integers and return their sum.", Parameters: json.RawMessage(`{"type":"object","properties":{"a":{"type":"integer"},"b":{"type":"integer"}},"required":["a","b"],"additionalProperties":false}`)}

type glmSession struct {
	ctx      context.Context
	client   *ai.Client
	rec      *recorder
	results  map[string]ai.Result
	requests map[string]ai.Request
	bodies   map[string]map[string]any
}

func TestGLMCandidate(t *testing.T) {
	if glmReportState == nil {
		t.Skip("isolated GLM route not selected")
	}
	r := glmReportState
	if glmConfigError != "" {
		glmFail(t, glmConfigError)
		return
	}
	cs := run.Case(t, "GLM11-sources")
	setGLMReplay(cs)
	for _, n := range []string{"sources.json", "candidate.json", "plan.json"} {
		raw, e := readDecisionsFile("testdata/glm-candidate/"+n, 2<<20)
		if !cs.Check("fixed source copied", e == nil, "%v", e) {
			return
		}
		cs.Fixture(n, raw)
	}
	if os.Getenv("BARNESS_AI_LIVE") != "1" || glmKey == "" {
		t.Skip("NOT_RUN: explicit live switch or selected key absent")
	}
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	if err := configureDecisionsProxy(rec); err != nil {
		glmFail(t, err.Error())
		return
	}
	if name := os.Getenv("BARNESS_GLM_CONTROLLED"); name != "" {
		r.Mode = "controlled"
		installGLMControlled(rec, name)
	}
	rec.next = &glmHTTPGate{next: rec.next}
	client, e := newGLMCandidateClient(glmSelected, glmKey, rec, glmCandidateCatalog)
	if e != nil {
		glmFail(t, "public Client assembly failed")
		return
	}
	r.ProcessDeadline = started.Add(1800 * time.Second)
	if d, ok := t.Deadline(); ok && d.Before(r.ProcessDeadline) {
		r.ProcessDeadline = d
	}
	r.ExecutionDeadline = r.ProcessDeadline.Add(-20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), r.ExecutionDeadline)
	defer cancel()
	s := &glmSession{ctx: ctx, client: client, rec: rec, results: map[string]ai.Result{}, requests: map[string]ai.Request{}, bodies: map[string]map[string]any{}}
	switch os.Getenv("BARNESS_GLM_CONTROLLED") {
	case "calls-budget":
		r.Budget.CallsUsed = 24
	case "attempts-budget":
		r.Budget.AttemptsReserved = 32
	case "actual-budget":
		r.Budget.HTTPAttempts = 32
	case "deadline":
		cancel()
	}
	for i, sc := range glmFixedPlan.Routes[r.Route] {
		t.Run(sc.ID, func(t *testing.T) { s.execute(t, sc, &r.Scenarios[i]) })
	}
	r.Outcome = "PASS"
	for _, m := range []string{"glm-5.3", "glm-5.2"} {
		r.Models[m] = "PASS"
		for _, sc := range r.Scenarios {
			if sc.Model == m && sc.Outcome != "PASS" {
				r.Models[m] = "FAIL"
				r.Outcome = "FAIL"
			}
		}
	}
}

func glmFullOptions(sc glmCasePlan) ai.Options {
	kind := sc.Kind
	if strings.HasPrefix(kind, "control-") {
		kind = "full-stream"
		if sc.Kind == "control-auto" {
			kind = "tool-call"
		}
		if sc.Kind == "control-unsigned" {
			kind = "native-history"
		}
	}
	switch glmSelected.api {
	case ai.APIOpenAICompletions:
		effort := ai.ThinkingLevel("")
		if kind == "full-stream" {
			effort = ai.ThinkingHigh
		}
		if kind == "full-complete" {
			effort = ai.ThinkingLow
		}
		o := ai.ChatOptions{MaxTokens: ai.Value(4096), ReasoningEffort: effort}
		if sc.Kind == "control-temperature" {
			o.Temperature = ai.Value(0.0)
		}
		if sc.Kind == "control-usage" {
			o.SamplingParams = map[string]json.RawMessage{"stream_options": json.RawMessage(`{"include_usage":true}`)}
		}
		return o
	case ai.APIOpenAIResponses:
		return ai.ResponsesOptions{MaxTokens: ai.Value(4096), ReasoningEffort: ai.ThinkingHigh}
	default:
		o := ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
		if sc.Model == "glm-5.2" && (kind == "full-stream" || kind == "native-history") && sc.Kind != "control-default" {
			o.ThinkingEnabled = ai.Value(true)
			o.ThinkingBudgetTokens = 1024
		}
		if sc.Kind == "control-auto" {
			o.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
		}
		return o
	}
}

func (s *glmSession) execute(t *testing.T, sc glmCasePlan, out *glmOutcome) {
	cs := run.Case(t, "GLM11-"+strings.ReplaceAll(sc.ID, "/", "-"))
	setGLMReplay(cs)
	out.Outcome = "FAIL"
	defer func() {
		if t.Failed() {
			out.Outcome = "FAIL"
			if out.Failure == "" {
				out.Failure = "assertion"
			}
		}
	}()
	req := ai.Request{Messages: []ai.Message{ai.UserText(glmFixedPlan.Prompts["text"])}}
	kind := sc.Kind
	if kind == "tool-call" || kind == "control-auto" {
		req = ai.Request{Messages: []ai.Message{ai.UserText(glmFixedPlan.Prompts["tool"])}, Tools: []ai.Tool{glmAdd}}
	}
	if kind == "tool-result" {
		prior, ok := s.results[sc.Model+"/tool-call"]
		if !cs.Check("own tool prerequisite", ok && prior.Message.StopReason == ai.StopReasonToolUse, "missing own tool turn") {
			return
		}
		req = s.requests[sc.Model+"/tool-call"]
		req.Messages = append([]ai.Message{}, req.Messages...)
		req.Messages = append(req.Messages, prior.Message)
		count := 0
		for i, c := range prior.Message.Content {
			call, ok := c.(ai.ToolCall)
			if !ok {
				continue
			}
			v, e := ai.ValidateToolCall(prior.Message, i, req.Tools)
			if !cs.Check("Host validates complete tool", e == nil, "%v", e) {
				return
			}
			var args struct{ A, B int }
			e = v.Decode(&args)
			if !cs.Check("Host executes known inputs", e == nil && args.A == 17 && args.B == 25, "unexpected arguments") {
				return
			}
			req.Messages = append(req.Messages, ai.ToolResultText(call, fmt.Sprint(args.A+args.B), false))
			count++
		}
		if !cs.Check("Host paired tool results", count > 0, "no tool results") {
			return
		}
	}
	if kind == "native-history" || kind == "control-unsigned" {
		prior, ok := s.results[sc.Model+"/full-stream"]
		if !cs.Check("own native prerequisite", ok && prior.Message.StopReason == ai.StopReasonStop, "missing successful own full stream") {
			return
		}
		msg := prior.Message
		env, trusted := msg.NativeState.Envelope()
		if !cs.Check("own native provenance", trusted, "not trusted") {
			return
		}
		msg.NativeState = ai.TrustedNativeState{}
		msg.Content = append([]ai.AssistantContent{}, msg.Content...)
		thoughts := 0
		signed := 0
		for i, c := range msg.Content {
			if v, ok := c.(ai.Thinking); ok && v.Thinking != "" {
				thoughts++
				if v.Signature != "" {
					signed++
				}
				if kind == "control-unsigned" {
					v.Signature = ""
					msg.Content[i] = v
				}
			}
		}
		if !cs.Check("own observable reasoning", thoughts > 0, "own full stream returned no reasoning") {
			return
		}
		if glmSelected.api == ai.APIAnthropicMessages && !cs.Check("own signed native baseline", signed > 0, "no signed thinking") {
			return
		}
		msg.NativeState, _ = ai.TrustNativeState(newScope(), env)
		req.Messages = append(req.Messages, msg, ai.UserText(glmFixedPlan.Prompts["continuation"]))
	}
	if kind == "cancel" {
		req.Messages = []ai.Message{ai.UserText(glmFixedPlan.Prompts["cancel"])}
	}
	if sc.Baseline != "" && kind != "control-unsigned" {
		baseline, ok := s.requests[sc.Baseline]
		if !cs.Check("own control prerequisite", ok, "baseline missing") {
			return
		}
		req = baseline
	}
	if s.ctx.Err() != nil {
		cs.Check("process execution deadline", false, "deadline exhausted")
		out.Failure = "deadline"
		return
	}
	if e := glmReserve(); !cs.Check("finite reservation before public call", e == nil, "%v", e) {
		out.Failure = "budget"
		return
	}
	out.Calls = 1
	s.requests[sc.ID] = req
	cs.Record("public-request", req)
	cs.Record("plan", sc)
	cs.Record("options", glmFullOptions(sc))
	cs.Record("budget-before", glmReportState.Budget)
	ctx, cancel := context.WithTimeout(s.ctx, 120*time.Second)
	defer cancel()
	scope := newScope()
	target := ai.Target{BindingID: glmSelected.name, ModelID: sc.Model}
	hook := s.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		if kind == "control-low" || kind == "control-max" {
			effort := "low"
			if kind == "control-max" {
				effort = "max"
			}
			obj, _ := p.Body["reasoning"].(map[string]any)
			if obj == nil {
				return ai.KeepPayload(), fmt.Errorf("missing reasoning baseline")
			}
			obj["effort"] = effort
		}
		return ai.KeepPayload(), nil
	}})
	beforeHTTP := glmReportState.Budget.HTTPAttempts
	began := time.Now()
	var result ai.Result
	var err error
	var events []json.RawMessage
	seenCancel := false
	simple := ai.SimpleOptions{MaxTokens: ai.Value(4096)}
	if glmSelected.api == ai.APIOpenAIResponses {
		simple.Reasoning = ai.ThinkingHigh
	}
	if glmSelected.api == ai.APIOpenAICompletions && kind == "simple-complete" {
		simple.Reasoning = ai.ThinkingMax
	}
	if kind == "full-complete" {
		result, err = hook.Complete(ctx, scope, target, req, glmFullOptions(sc))
	} else if kind == "simple-complete" {
		result, err = hook.CompleteSimple(ctx, scope, target, req, simple)
	} else {
		var stream *ai.Stream
		if kind == "simple-stream" {
			stream = hook.StreamSimple(ctx, scope, target, req, simple)
		} else {
			stream = hook.Stream(ctx, scope, target, req, glmFullOptions(sc))
		}
		for stream.Next() {
			event := stream.Event()
			raw, _ := json.Marshal(stream.Envelope())
			events = append(events, raw)
			if kind == "cancel" && !seenCancel && (event.Type() == ai.EventTextDelta || event.Type() == ai.EventThinkingDelta) {
				seenCancel = true
				cancel()
			}
		}
		result, err = stream.Result()
		stream.Close()
	}
	out.DurationMS = time.Since(began).Milliseconds()
	obs := s.rec.observationsOf(scope.RequestID)
	ex := s.rec.take()
	out.HTTPAttempts = glmReportState.Budget.HTTPAttempts - beforeHTTP
	cs.Record("attempt-1-exchanges", ex)
	cs.Record("attempt-1-result", map[string]any{"result": result, "error": errText(err)})
	cs.Record("events", events)
	cs.Record("observations-1", obs)
	cs.Record("budget-after", glmReportState.Budget)
	if !cs.Check("single actual attempt and finite capture", len(ex) == 1 && len(result.Metadata.Attempts) == 1 && glmReportState.Budget.HTTPAttempts <= 32, "actual=%d attempts=%d", len(ex), len(result.Metadata.Attempts)) {
		return
	}
	cs.Check("exact snapshot attribution", result.Metadata.CatalogHash == glmReportState.CatalogHash && result.Metadata.ProviderID == ai.ProviderZhipu && result.Metadata.API == glmSelected.api && result.Metadata.ModelID == sc.Model, "wrong attribution")
	cs.Check("finished Observer metadata", len(obs) == 4 && obs[len(obs)-1].Kind == ai.ObservationCallFinished, "observations=%d", len(obs))
	out.Usage = result.Message.Usage
	for _, a := range result.Metadata.Attempts {
		out.UsageReporting = append(out.UsageReporting, a.UsageReporting)
		out.ProviderRequestIDs = append(out.ProviderRequestIDs, a.ProviderRequestID)
	}
	for _, e := range ex {
		out.Truncated = out.Truncated || e.Truncated
		out.RawUsagePresence = append(out.RawUsagePresence, glmUsagePresence(e.ResponseBody)...)
		if !cs.Check("capture has no accidental truncation", !e.Truncated, "capture truncated") {
			return
		}
		var b map[string]any
		json.Unmarshal([]byte(e.RequestBody), &b)
		s.bodies[sc.ID] = b
		cs.Fixture("request.json", []byte(e.RequestBody))
		cs.Fixture("response.sse", []byte(e.ResponseBody))
		cs.Record("usage-presence", glmUsagePresence(e.ResponseBody))
		n := b["max_tokens"]
		if glmSelected.api == ai.APIOpenAIResponses {
			n = b["max_output_tokens"]
		}
		cs.Check("final output budget exact4096", n == float64(4096), "budget=%v", n)
		cs.Record("request-id-source", map[string]any{"responseHeaders": e.ResponseHeaders, "providerRequestIDs": out.ProviderRequestIDs, "responseID": result.Message.ResponseID, "actualSource": glmRequestIDSources(result, e)})
	}
	if kind == "cancel" {
		var ae *ai.Error
		if cs.Check("cancel after observable content", seenCancel && errors.As(err, &ae) && ae.Code == ai.CodeCanceled && result.Message.StopReason == ai.StopReasonAborted, "classification=%s", glmSafeFailure(err)) {
			out.Outcome = "PASS"
		}
		cs.Record("capture-scope", "intentional partial stream read after Host cancellation; absent usage remains unknown")
		return
	}
	if !cs.Check("own public call succeeds", err == nil, "classification=%s", glmSafeFailure(err)) {
		out.Failure = "public-call"
		return
	}
	if !cs.Check("complete required terminal", result.Message.StopReason == ai.StopReasonStop || (kind == "tool-call" || kind == "control-auto") && result.Message.StopReason == ai.StopReasonToolUse, "stop=%s", result.Message.StopReason) {
		return
	}
	if kind != "tool-call" && kind != "control-auto" {
		text := ""
		for _, part := range result.Message.Content {
			if v, ok := part.(ai.Text); ok {
				text += v.Text
			}
		}
		cs.Check("required text entry produces final text", strings.TrimSpace(text) != "", "no final text")
	}
	if kind == "tool-call" || kind == "control-auto" {
		count := 0
		for i, c := range result.Message.Content {
			if _, ok := c.(ai.ToolCall); ok {
				_, e := ai.ValidateToolCall(result.Message, i, req.Tools)
				cs.Check("Host tool call executable", e == nil, "%v", e)
				count++
			}
		}
		cs.Check("own automatic function call", count > 0, "no function call")
	}
	if kind == "native-history" && glmSelected.api == ai.APIOpenAIResponses {
		input, _ := s.bodies[sc.ID]["input"].([]any)
		found := false
		for _, v := range input {
			item, _ := v.(map[string]any)
			if item["type"] == "reasoning" {
				_, found = item["content"].(map[string]any)
			}
		}
		cs.Check("own Responses InputReasoning object replay", found, "object reasoning absent")
	}
	if sc.Baseline != "" {
		glmCheckControl(cs, sc, s.bodies)
	}
	cs.Record("estimated-cost-provenance", map[string]any{"usage": result.Message.Usage, "catalogHash": glmReportState.CatalogHash, "sourceManifestHash": glmSourcePin, "currency": "USD estimate; reported known counters only, not invoice"})
	s.results[sc.ID] = result
	out.Outcome = "PASS"
}

func glmCheckControl(cs *evidence.Case, sc glmCasePlan, bodies map[string]map[string]any) {
	clone := func(v any) map[string]any {
		raw, _ := json.Marshal(v)
		var b map[string]any
		json.Unmarshal(raw, &b)
		return b
	}
	a, b := clone(bodies[sc.Baseline]), clone(bodies[sc.ID])
	originalA, _ := json.Marshal(a)
	originalB, _ := json.Marshal(b)
	cs.Check("declared field actually changes", string(originalA) != string(originalB), "control identical to baseline")
	switch sc.Kind {
	case "control-temperature":
		cs.Check("temperature0 present", b["temperature"] == float64(0), "missing")
		delete(a, "temperature")
		delete(b, "temperature")
	case "control-usage":
		x, _ := b["stream_options"].(map[string]any)
		cs.Check("include_usage true", x["include_usage"] == true, "missing")
		delete(a, "stream_options")
		delete(b, "stream_options")
	case "control-low", "control-max":
		for _, o := range []map[string]any{a, b} {
			if x, ok := o["reasoning"].(map[string]any); ok {
				delete(x, "effort")
			}
		}
	case "control-auto":
		delete(a, "tool_choice")
		delete(b, "tool_choice")
	case "control-default":
		delete(a, "thinking")
		delete(b, "thinking")
	case "control-unsigned":
		for _, o := range []map[string]any{a, b} {
			messages, _ := o["messages"].([]any)
			for _, v := range messages {
				msg, _ := v.(map[string]any)
				content, _ := msg["content"].([]any)
				for _, v := range content {
					part, _ := v.(map[string]any)
					if part["type"] == "thinking" {
						delete(part, "signature")
					}
				}
			}
		}
	}
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	cs.Record("single-variable-comparison", map[string]any{"baseline": sc.Baseline, "soleField": sc.OnlyChangedField, "equalOtherFields": string(aa) == string(bb)})
	cs.Check("only preregistered field differs", string(aa) == string(bb), "unexpected body difference")
}
