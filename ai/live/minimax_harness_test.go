//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestMiniMaxResearch(t *testing.T)  { driveMiniMax(t) }
func TestMiniMaxCandidate(t *testing.T) { driveMiniMax(t) }
func TestMiniMaxFormal(t *testing.T)    { driveMiniMax(t) }
func TestMiniMaxReplay(t *testing.T)    { driveMiniMax(t) }

type minimaxSession struct {
	client        *ai.Client
	rec           *recorder
	gate          *minimaxHTTPGate
	ctx           context.Context
	tool          ai.Result
	toolRequest   ai.Request
	requestReruns int
}

func driveMiniMax(t *testing.T) {
	if minimaxReport == nil {
		t.Skip("isolated MiniMax identity not selected")
	}
	r := minimaxReport
	if minimaxConfigError != "" {
		minimaxFail(minimaxConfigError)
		t.Fatal(minimaxConfigError)
	}
	if r.Mode != "replay" && (os.Getenv(envLive) != "1" || minimaxKey == "" || r.Entitlement != "confirmed") {
		t.Skip("NOT_RUN: live switch, selected key or independently confirmed entitlement absent")
	}
	sources, e := minimaxlive.Sources()
	if e != nil {
		minimaxFail("fixed sources unavailable")
		t.Fatal(e)
	}
	if e = run.Record("minimax-sources", sources); e != nil {
		t.Fatal(e)
	}
	if e = run.Record("minimax-catalog", minimaxCatalog); e != nil {
		t.Fatal(e)
	}
	rec := newRecorder()
	transport := rec.next.(*http.Transport)
	defer transport.CloseIdleConnections()
	if proxy := os.Getenv("BARNESS_MINIMAX_PROXY"); proxy != "" {
		u, e := smokeProxy(proxy)
		if e != nil {
			minimaxFail("invalid explicit proxy")
			t.Fatal(e)
		}
		transport.Proxy = http.ProxyURL(u)
	}
	if r.Mode == "controlled" {
		installMiniMaxControlled(rec, os.Getenv("BARNESS_MINIMAX_CONTROLLED"))
	}
	if r.Mode == "replay" {
		if e = installMiniMaxReplay(rec); e != nil {
			minimaxFail("capture replay rejected")
			t.Fatal(e)
		}
	}
	gate := &minimaxHTTPGate{next: rec.next}
	rec.next = gate
	policy := miniMaxPolicy()
	key := minimaxKey
	if r.Mode == "replay" {
		key = "synthetic-capture-replay"
	}
	client, e := newClientWithPolicy(minimaxSelected, key, rec, minimaxCatalog, policy)
	if e != nil {
		minimaxFail("public Client assembly failed")
		t.Fatal(e)
	}
	testDeadline, _ := t.Deadline()
	ctx, cancel := context.WithDeadline(context.Background(), miniMaxContextDeadline(r, testDeadline))
	defer cancel()
	s := &minimaxSession{client: client, rec: rec, gate: gate, ctx: ctx}
	for _, p := range r.Expected {
		retry := false
		t.Run(p.ID, func(t *testing.T) { retry = s.execute(t, p, 0) })
		if retry && r.Kind != minimaxlive.ThinkingKind {
			if r.Kind == "minimax-research-only" {
				t.Errorf("research own tool/reasoning remains unproven")
			} else {
				t.Run(p.ID+"-rerun-1", func(t *testing.T) { s.execute(t, p, 1) })
			}
		}
		if r.Kind == "minimax-research-only" && p.ID == "research-tool-call" &&
			(len(r.Calls) == 0 || r.Calls[len(r.Calls)-1].Outcome != "PASS") {
			r.Unknown = append(r.Unknown, "research-tool-result-native NOT_RUN: own tool/reasoning prerequisite did not pass")
			break
		}
		if r.Failure == "known-input-overrun" {
			break
		}
	}
	r.Outcome = "PASS"
	if t.Failed() || len(r.Calls) < len(r.Expected) || len(r.Calls) > len(r.Expected)+1+minimaxlive.MaxRequestReruns {
		r.Outcome = "FAIL"
	}
	for i, c := range r.Calls {
		if c.Outcome != "PASS" && !minimaxlive.Recovered(r.Calls, i) {
			r.Outcome = "FAIL"
		}
	}
}

func miniMaxPolicy() *ai.ResourcePolicy {
	policy := livePolicy()
	policy.CallTimeout = 120 * time.Second
	// This Host transport validates the complete native authority before it
	// dials. That local work exceeded the ordinary 15s connect clock (issue24).
	// Keep each fresh validation in place and let the unchanged call/header
	// limits bound the combined local and network work instead.
	policy.ConnectTimeout = policy.CallTimeout
	return policy
}

func (s *minimaxSession) execute(t *testing.T, p minimaxlive.Case, rerun int) (needsRerun bool) {
	cs := run.Case(t, minimaxlive.CaseID(p.ID, rerun))
	cs.ReplayRun("^TestMiniMaxProcess$/^success-" + minimaxReport.Identity.Route + "$")
	out := minimaxlive.Call{ID: p.ID, CaseID: cs.ID(), Rerun: rerun, Outcome: "FAIL"}
	originalID := p.ID
	p.ID = strings.TrimPrefix(p.ID, "research-")
	defer func() {
		if t.Failed() {
			out.Outcome = "FAIL"
		}
		minimaxReport.Calls = append(minimaxReport.Calls, out)
	}()
	req, e := s.request(p)
	if !cs.Check("own complete history prerequisite", e == nil, "%v", e) {
		return
	}
	provenance := grokhistory.Provenance{Before: grokhistory.Envelopes(req)}
	raw, _ := json.Marshal(req)
	input := len(raw) + 1024
	if input < 2048 {
		input = 2048
	}
	out.InputBytes = len(raw)
	if s.ctx.Err() != nil {
		cs.Check("process deadline", false, "execution deadline exhausted")
		return
	}
	reservation, e := minimaxLedger.Reserve(originalID, p.Output, input)
	if !cs.Check("durable cumulative reservation precedes Client", e == nil, "%v", e) {
		return
	}
	out.Reservation = reservation
	b := &minimaxReport.Budget
	b.Calls++
	b.ReservedAttempts++
	b.RequestedOutput += p.Output
	b.ReservedInput += input
	s.gate.caseID = originalID
	s.gate.receipts = nil
	switch os.Getenv("BARNESS_MINIMAX_CONTROLLED") {
	case "output-budget":
		b.RequestedOutput = 40001
	case "input-budget":
		b.ReservedInput = 120001
	}
	if !cs.Check("process cumulative output and input bounded", b.Calls <= 20 && b.RequestedOutput <= 40000 && b.ReservedInput <= 120000, "budget exceeded") {
		return
	}
	cs.Record("public-request", req)
	cs.Record("options", miniMaxOptions(p))
	cs.Record("plan", p)
	cs.Record("budget-before", *b)
	ctx, cancel := context.WithTimeout(s.ctx, 120*time.Second)
	defer cancel()
	scope := newScope()
	out.RequestID = scope.RequestID
	target := ai.Target{BindingID: minimaxSelected.name, ModelID: minimaxReport.Identity.Model}
	hookCalls := 0
	hook := s.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
		hookCalls++
		return ai.KeepPayload(), nil
	}})
	simple := ai.SimpleOptions{MaxTokens: ai.Value(p.Output)}
	capture := minimaxlive.Capture{}
	var err error
	var events []json.RawMessage
	visible := false
	began := time.Now()
	switch p.ID {
	case "full-complete":
		capture.Result, err = hook.Complete(ctx, scope, target, req, miniMaxOptions(p))
	case "simple-complete":
		capture.Result, err = hook.CompleteSimple(ctx, scope, target, req, simple)
	default:
		var stream *ai.Stream
		if p.ID == "simple-stream" {
			stream = hook.StreamSimple(ctx, scope, target, req, simple)
		} else {
			stream = hook.Stream(ctx, scope, target, req, miniMaxOptions(p))
		}
		for stream.Next() {
			event := stream.Event()
			raw, _ := json.Marshal(map[string]any{"type": event.Type(), "envelope": stream.Envelope()})
			events = append(events, raw)
			if p.ID == "cancel" && !visible && (event.Type() == ai.EventTextDelta || event.Type() == ai.EventThinkingDelta) {
				visible = true
				cancel()
			}
		}
		capture.Result, err = stream.Result()
		stream.Close()
	}
	out.DurationMS = time.Since(began).Milliseconds()
	var ae *ai.Error
	if errors.As(err, &ae) {
		capture.ErrorCode = ae.Code
		capture.ErrorPhase = ae.Phase
	}
	obs := s.rec.observationsOf(scope.RequestID)
	ex := s.rec.take()
	if p.ID == grokhistory.Scenario {
		after, _ := json.Marshal(req)
		cs.Record("immutable-input", req)
		provenance.After = grokhistory.Envelopes(req)
		cs.Record("history-provenance", provenance)
		rawProof, _ := json.Marshal(provenance)
		cs.Check("original trusted envelopes unchanged", grokhistory.VerifyProvenance(rawProof, scope.TenantID) == nil, "source envelopes changed")
		cs.Check("fixed Host input immutable", grokhistory.VerifyHost(after), "input changed")
		cs.Check("two foreign assistant downgrades", capture.Result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{CrossModel: 2}), "downgrades differ")
		cs.Check("actual xAI collision/reuse wire", len(ex) == 1 && grokhistory.Verify(minimaxReport.Identity.API, []byte(ex[0].RequestBody)) == nil, "wire differs")
	}
	out.Attempts = append(out.Attempts, s.gate.receipts...)
	cs.Record("result", capture)
	cs.Record("events", events)
	cs.Record("observations", obs)
	cs.Record("exchanges", ex)
	cs.Record("metrics", map[string]any{"activeBodies": s.gate.active.Load(), "observer": s.client.ObserverStats(), "hookCalls": hookCalls})
	cs.Record("budget-after", *b)
	if minimaxReport.Kind == minimaxlive.ThinkingKind {
		cs.Record("thinking-budget", s.gate.thinkingBudget)
	}
	compareMiniMaxReplay(cs, cs.ID(), capture, events)
	if p.ID == "preflight-error" {
		if cs.Check("preflight refuses with zero attempts", err != nil && len(ex) == 0 && len(capture.Result.Metadata.Attempts) == 0, "code=%s", capture.ErrorCode) {
			out.Outcome = "PASS"
		}
		return
	}
	if !cs.Check("exactly one actual Attempt with settled Observer", len(ex) == 1 && len(capture.Result.Metadata.Attempts) == 1 && len(out.Attempts) == 1 && len(obs) == 4 && obs[3].Kind == ai.ObservationCallFinished, "exchanges=%d attempts=%d observations=%d", len(ex), len(capture.Result.Metadata.Attempts), len(obs)) {
		return
	}
	res := capture.Result
	cs.Check("exact identity snapshot", res.Metadata.ProviderID == ai.ProviderMiniMax && res.Metadata.API == minimaxReport.Identity.API && res.Metadata.ModelID == minimaxReport.Identity.Model && res.Metadata.CatalogHash == minimaxReport.CatalogHash, "identity mismatch")
	cs.Check("body and Observer resources settled", s.gate.active.Load() == 0 && s.client.ObserverStats().Dropped == 0 && s.client.ObserverStats().Failed == 0, "resource not settled")
	if rerun == 0 && minimaxReport.Kind != "minimax-research-only" && minimaxReport.Kind != minimaxlive.ThinkingKind && s.requestReruns < minimaxlive.MaxRequestReruns && minimaxlive.RequestRerunnable(capture, events, ex[0].Status, ex[0].ResponseBody) {
		for _, x := range ex {
			cs.Check("bounded initial request failure capture", !x.Truncated, "truncated")
			cs.Fixture("request.json", []byte(x.RequestBody))
			cs.Fixture("response.sse", []byte(x.ResponseBody))
			out.RawModels, out.UsagePresence = minimaxlive.RawFacts(x.ResponseBody)
		}
		s.requestReruns++
		out.Failure = "transient-request"
		cs.Record("request-rerun", map[string]any{"outcome": "FAIL", "reason": "initial transient REQUEST only; no successful stream response", "ordinal": s.requestReruns, "limit": minimaxlive.MaxRequestReruns})
		return true
	}
	for _, x := range ex {
		cs.Check("complete bounded capture", !x.Truncated, "truncated")
		cs.Fixture("request.json", []byte(x.RequestBody))
		cs.Fixture("response.sse", []byte(x.ResponseBody))
		out.RawModels, out.UsagePresence = minimaxlive.RawFacts(x.ResponseBody)
		sources := glmRequestIDSources(res, x)
		out.RequestIDSources = append(out.RequestIDSources, sources...)
		for _, model := range out.RawModels {
			cs.Check("raw actual model matches selected identity", model == minimaxReport.Identity.Model, "actual model differs")
		}
		cs.Check("actual model present", len(out.RawModels) > 0 || p.ID == "cancel", "missing actual model")
		var body map[string]any
		json.Unmarshal([]byte(x.RequestBody), &body)
		field := "max_tokens"
		if minimaxSelected.api == ai.APIOpenAICompletions {
			field = "max_completion_tokens"
		}
		if minimaxSelected.api == ai.APIOpenAIResponses {
			field = "max_output_tokens"
		}
		cs.Check("fixed final output reservation", body[field] == float64(p.Output), "field=%s value=%v", field, body[field])
	}
	for _, a := range res.Metadata.Attempts {
		known := a.Usage.Input + a.Usage.CacheRead + a.Usage.CacheWrite
		if known > int64(input) {
			minimaxReport.Failure = "known-input-overrun"
			cs.Check("known input within conservative reservation", false, "reported known input exceeded reservation")
			return
		}
	}
	if p.ID == "cancel" {
		cs.Check("cancel after visible content", visible && capture.ErrorCode == ai.CodeCanceled && res.Message.StopReason == ai.StopReasonAborted, "code=%s stop=%s", capture.ErrorCode, res.Message.StopReason)
	} else if p.ID == "truncate" {
		cs.Check("length terminal", err == nil && res.Message.StopReason == ai.StopReasonLength, "code=%s stop=%s", capture.ErrorCode, res.Message.StopReason)
	} else {
		cs.Check("public call succeeds", err == nil, "code=%s phase=%s", capture.ErrorCode, capture.ErrorPhase)
		if p.ID == "tool-call" {
			if err != nil {
				return
			}
			tools, thoughts := 0, 0
			for i, part := range res.Message.Content {
				switch v := part.(type) {
				case ai.ToolCall:
					_, e := ai.ValidateToolCall(res.Message, i, req.Tools)
					cs.Check("Host validates own tool", e == nil, "%v", e)
					tools++
				case ai.Thinking:
					if v.Thinking != "" {
						thoughts++
					}
				}
			}
			if tools == 0 || thoughts == 0 || res.Message.StopReason != ai.StopReasonToolUse {
				out.Failure = "tool-capability"
				cs.Record("incomplete-capability", map[string]any{"tools": tools, "thoughts": thoughts, "stop": res.Message.StopReason, "outcome": "FAIL", "note": "only tool-call can rerun once in the same ledger"})
				if rerun > 0 {
					cs.Check("own tool and reasoning coexist after allowed rerun", false, "tools=%d thoughts=%d stop=%s", tools, thoughts, res.Message.StopReason)
				}
				return true
			}
			cs.Check("own tool and reasoning coexist", true, "tools=%d thoughts=%d", tools, thoughts)
			s.tool = res
			s.toolRequest = req
		} else {
			cs.Check("successful terminal", res.Message.StopReason == ai.StopReasonStop, "stop=%s", res.Message.StopReason)
		}
		if p.ID != "tool-call" {
			text := ""
			for _, part := range res.Message.Content {
				if value, ok := part.(ai.Text); ok {
					text += value.Text
				}
			}
			cs.Check("required successful text", strings.TrimSpace(text) != "", "empty text")
		}
		if p.ID == "thinking-control" && minimaxReport.Identity.Model == minimaxlive.M3 {
			thoughts := 0
			for _, part := range res.Message.Content {
				if value, ok := part.(ai.Thinking); ok && value.Thinking != "" {
					thoughts++
				}
			}
			cs.Check("binary off returns no reasoning", thoughts == 0, "thoughts=%d", thoughts)
		}
		if p.ID == "thinking-on" {
			thoughts := 0
			for _, part := range res.Message.Content {
				if v, ok := part.(ai.Thinking); ok && v.Thinking != "" {
					thoughts++
				}
			}
			cs.Check("own reasoning actually observed", thoughts > 0, "no reasoning")
		}
	}
	out.Outcome = "PASS"
	return false
}
