//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

var regionalTool = ai.Tool{Name: "inspect_regional", Description: "Return the synthetic regional probe value for label alpha.", Parameters: json.RawMessage(`{"type":"object","properties":{"label":{"type":"string","enum":["alpha"]}},"required":["label"],"additionalProperties":false}`)}

func TestKimiHistorySupplement(t *testing.T) { driveRegional(t) }

func TestRegionalCandidate(t *testing.T) { driveRegional(t) }
func TestRegionalFormal(t *testing.T)    { driveRegional(t) }

type regionalSession struct {
	p      *regionalProcess
	client *ai.Client
	rec    *recorder
	gate   *regionalHTTPGate
	probe  *probe.Probe
	ctx    context.Context
	tools  map[string]ai.AssistantMessage
	native map[string]ai.AssistantMessage
}

func driveRegional(t *testing.T) {
	p := regionalState
	if p == nil {
		t.Skip("NOT_RUN: explicit new regional route not selected")
	}
	if p.problem != "" {
		t.Fatal(p.problem)
	}
	if p.notRun != "" {
		t.Skip("NOT_RUN: " + p.notRun)
	}
	processCtx := context.Background()
	if supportmatrix.RegionalIsKimi(p.report.Combo) {
		var stop context.CancelFunc
		processCtx, stop = signal.NotifyContext(processCtx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}
	if err := p.consumeGrant(); err != nil {
		p.problem = "regional grant reservation failed"
		t.Fatal(p.problem)
	}
	_ = run.Record("regional-execution-source", p.source)
	_ = run.Record("regional-catalog", p.cat)
	_ = run.Record("regional-candidate-facts", supportmatrix.RegionalCandidateFacts(p.report.Combo))
	rec := newRecorder()
	transport := rec.next.(*http.Transport)
	defer transport.CloseIdleConnections()
	if proxy := os.Getenv("BARNESS_REGIONAL_PROXY"); proxy != "" {
		u, err := smokeProxy(proxy)
		if err != nil {
			t.Fatal("invalid explicit regional proxy")
		}
		transport.Proxy = http.ProxyURL(u)
	}
	if p.controlled != "" {
		if err := installRegionalControlled(rec, p.controlled); err != nil {
			t.Fatal(err)
		}
	}
	gate := &regionalHTTPGate{next: rec.next, process: p}
	rec.next = gate
	gauges := probe.New()
	models := supportmatrix.RegionalModels(p.report.Combo)
	resolver := &resolver{binding: ai.Binding{TenantID: liveTenant, BindingID: p.report.Combo, Version: "regional-v1", Enabled: true, ProviderID: ai.ProviderID(p.report.Provider), API: ai.API(p.report.API), Endpoint: p.report.Endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: "regional:" + p.report.AccountAlias, CredentialRef: "regional-own-credential", AllowedModels: models}, credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: "regional-own-credential", Version: "regional-v1", BindingVersion: "regional-v1", AccountScopeID: "regional:" + p.report.AccountAlias, Active: true, APIKey: ai.NewSecret(p.key)}}
	policy := livePolicy()
	policy.CallTimeout = 120 * time.Second
	client, err := ai.NewClient(ai.Config{Policy: policy, Catalog: &p.cat, Bindings: resolver, Credentials: resolver, Transport: rec, Observer: rec, Probe: gauges})
	if err != nil {
		t.Fatal("regional public Client assembly failed")
	}
	deadline := p.report.StartedAt.Add(time.Duration(p.report.Budget.ProcessSeconds-20) * time.Second)
	if supportmatrix.RegionalIsKimi(p.report.Combo) && p.grant.ExpiresAt.Before(deadline) {
		deadline = p.grant.ExpiresAt
	}
	ctx, cancel := context.WithDeadline(processCtx, deadline)
	defer cancel()
	s := regionalSession{p: p, client: client, rec: rec, gate: gate, probe: gauges, ctx: ctx, tools: map[string]ai.AssistantMessage{}, native: map[string]ai.AssistantMessage{}}
	for i := range p.report.Scenarios {
		i := i
		t.Run(p.report.Scenarios[i].ID, func(t *testing.T) { s.execute(t, i) })
		if supportmatrix.RegionalIsKimi(p.report.Combo) && t.Failed() {
			break
		}
	}
}

func regionalOptions(model ai.Model, budget int) ai.Options {
	switch model.API {
	case ai.APIOpenAICompletions:
		return ai.ChatOptions{MaxTokens: ai.Value(budget), ReasoningEffort: ai.ThinkingHigh}
	case ai.APIOpenAIResponses:
		return ai.ResponsesOptions{MaxTokens: ai.Value(budget), ReasoningEffort: ai.ThinkingHigh}
	default:
		o := ai.AnthropicOptions{MaxTokens: ai.Value(budget)}
		if model.Provider == ai.ProviderMiniMaxCN {
			o.ThinkingEnabled = ai.Value(true)
		}
		// Only the independently evidenced ZAI GLM-5.2 candidate opts into
		// this exact full shape. GLM-5.3/default and simple keep omission;
		// this is no canonical intensity map or budget enforcement claim.
		if model.Provider == ai.ProviderZAI && model.Compat.MessagesThinkingBudgetTokens == 1024 {
			o.ThinkingEnabled, o.ThinkingBudgetTokens = ai.Value(true), 1024
		}
		return o
	}
}

func (s *regionalSession) request(model, kind string) (ai.Request, error) {
	if kind == grokhistory.Scenario {
		return grokhistory.Request(ai.CallScope{TenantID: liveTenant})
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText("Reply with the exact short text regional ready.")}}
	if kind == "cancel" {
		req.Messages = []ai.Message{ai.UserText("Write a detailed long explanation of ocean currents in at least 1000 words.")}
	}
	if kind == "tool-call" {
		req.Messages = []ai.Message{ai.UserText("Call inspect_regional with label alpha. You must use this tool before answering. Think through the request first.")}
		req.Tools = []ai.Tool{regionalTool}
	}
	if kind == "native-history" && supportmatrix.RegionalIsKimi(s.p.report.Combo) {
		first, ok := s.native[model]
		if !ok {
			return ai.Request{}, fmt.Errorf("fixed own native source prerequisite absent")
		}
		req.Messages = append(req.Messages, first, ai.UserText("Continue that exact conversation with a short answer."))
		return req, nil
	}
	if kind == "tool-result" || kind == "native-history" {
		m, ok := s.tools[model]
		if !ok {
			return ai.Request{}, fmt.Errorf("own model tool prerequisite absent")
		}
		req.Messages = []ai.Message{ai.UserText("Call inspect_regional with label alpha. You must use this tool before answering. Think through the request first."), m}
		if kind == "native-history" && supportmatrix.RegionalNativeSourceCase(s.p.report.Combo, model) == supportmatrix.RegionalCaseID(s.p.report.Combo, model+"/full-stream") {
			first, ok := s.native[model]
			if !ok {
				return ai.Request{}, fmt.Errorf("fixed own native source prerequisite absent")
			}
			// Predeclared full-stream source, never a successful-turn search.
			// Keep the complete tool/Host-result chain after that own turn.
			req.Messages = append([]ai.Message{ai.UserText("Reply with the exact short text regional ready."), first}, req.Messages...)
		}
		for _, b := range m.Content {
			if c, ok := b.(ai.ToolCall); ok {
				req.Messages = append(req.Messages, ai.ToolResultText(c, "REGIONAL42", false))
			}
		}
		req.Messages = append(req.Messages, ai.UserText("Reply with exactly REGIONAL42 using that result, without calling another tool."))
		req.Tools = []ai.Tool{regionalTool}
	}
	return req, nil
}

func (s *regionalSession) execute(t *testing.T, i int) {
	row := &s.p.report.Scenarios[i]
	sourceCase := supportmatrix.RegionalNativeSourceCase(s.p.report.Combo, row.Model)
	_, kind, _ := strings.Cut(row.ID, "/")
	req, err := s.request(row.Model, kind)
	if err != nil {
		t.Skip("NOT_RUN: own tool prerequisite missing")
	}
	provenance := grokhistory.Provenance{Before: grokhistory.Envelopes(req)}
	cs := run.Case(t, row.CaseID)
	cs.ReplayRun(regionalEntry())
	modelInfo, _ := s.p.cat.Lookup(ai.ProviderID(s.p.report.Provider), ai.API(s.p.report.API), row.Model)
	if err := s.p.waitInterval(s.ctx); err != nil {
		row.Outcome, row.ErrorCategory = supportmatrix.Fail, supportmatrix.CategoryBudget
		cs.Check("pacing wait honors process cancellation and grant expiry", false, "%v", err)
		return
	}
	outputBudget := min(4096, modelInfo.MaxTokens)
	if err := s.p.reserve(s.ctx, outputBudget); err != nil {
		row.Outcome, row.ErrorCategory = supportmatrix.Fail, supportmatrix.CategoryBudget
		cs.Check("own finite budget available", false, "reservation exhausted")
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, 120*time.Second)
	defer cancel()
	scope := newScope()
	target := ai.Target{BindingID: s.p.report.Combo, ModelID: row.Model}
	opts := regionalOptions(modelInfo, outputBudget)
	if supportmatrix.RegionalIsKimi(s.p.report.Combo) {
		opts = kimiRegionalOptions(modelInfo, outputBudget, kind)
	}
	if s.p.report.API == "anthropic-messages" && s.p.report.Provider == "zai" && (kind == "simple-stream" || kind == "simple-complete") {
		// options.json is the effective full shape used by independent raw
		// replay. Simple omits thinking even when full explicitly opts in.
		opts = ai.AnthropicOptions{MaxTokens: ai.Value(outputBudget)}
	}
	simple := ai.SimpleOptions{MaxTokens: ai.Value(outputBudget)}
	if supportmatrix.RegionalIsKimi(s.p.report.Combo) {
		simple.Reasoning = ai.ThinkingLow
		if kind == "simple-complete" {
			simple.Reasoning = ai.ThinkingMedium
		}
		opts = kimiRegionalEffectiveOptions(modelInfo, outputBudget, kind)
	}
	if !supportmatrix.RegionalIsKimi(s.p.report.Combo) && (s.p.report.API != "anthropic-messages" || s.p.report.Provider == "minimax-cn") {
		simple.Reasoning = ai.ThinkingHigh
	}
	began := time.Now()
	before := s.gate.attempts.Load()
	beforeInput := s.p.report.Budget.ReservedInputTokens
	var result ai.Result
	var events []map[string]any
	seenCancel := false
	switch kind {
	case "full-complete":
		result, err = s.client.Complete(ctx, scope, target, req, opts)
	case "simple-complete":
		result, err = s.client.CompleteSimple(ctx, scope, target, req, simple)
	default:
		var stream *ai.Stream
		if kind == "simple-stream" {
			stream = s.client.StreamSimple(ctx, scope, target, req, simple)
		} else {
			stream = s.client.Stream(ctx, scope, target, req, opts)
		}
		for stream.Next() {
			e := stream.Event()
			if len(events) < 4096 {
				events = append(events, map[string]any{"type": e.Type(), "content": e})
			} else {
				cancel()
			}
			if kind == "cancel" && !seenCancel && (e.Type() == ai.EventTextDelta || e.Type() == ai.EventThinkingDelta) {
				seenCancel = true
				cancel()
			}
		}
		result, err = stream.Result()
		stream.Close()
	}
	row.DurationMS = time.Since(began).Milliseconds()
	row.Calls = 1
	row.Attempts = int(s.gate.attempts.Load() - before)
	s.p.report.Budget.HTTPAttempts += row.Attempts
	s.p.report.Budget.InputTokensUsed += int(result.Message.Usage.Input + result.Message.Usage.CacheRead + result.Message.Usage.CacheWrite)
	s.p.report.Budget.OutputTokensUsed += int(result.Message.Usage.Output)
	for _, a := range result.Metadata.Attempts {
		if a.ProviderRequestID != "" {
			row.ProviderRequestIDs = append(row.ProviderRequestIDs, a.ProviderRequestID)
		}
	}
	cs.Record("request", req)
	cs.Record("tools", req.Tools)
	cs.Record("options", opts)
	cs.Record("result", result)
	cs.Record("events", events)
	exchanges := s.rec.take()
	cs.Record("exchanges", exchanges)
	if kind == grokhistory.Scenario {
		raw, _ := json.Marshal(req)
		cs.Record("immutable-input", req)
		provenance.After = grokhistory.Envelopes(req)
		cs.Record("history-provenance", provenance)
		rawProof, _ := json.Marshal(provenance)
		cs.Check("original trusted envelopes unchanged", grokhistory.VerifyProvenance(rawProof, scope.TenantID) == nil, "source envelopes changed")
		cs.Check("fixed Host history unchanged after call", grokhistory.VerifyHost(raw), "history changed")
		cs.Check("foreign source downgrades exactly two assistant turns", result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{CrossModel: 2}), "%+v", result.Metadata.NativeStateDowngrades)
		proof := len(exchanges) == 1 && grokhistory.Verify(ai.API(s.p.report.API), []byte(exchanges[0].RequestBody)) == nil
		cs.Check("actual wire preserves foreign collision and reuse", proof, "fixed actual wire proof")
	}
	cs.Record("observations", s.rec.observationsOf(scope.RequestID))
	deadline := time.Now().Add(time.Second)
	for s.probe.ActiveCalls() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	resources := map[string]int64{"calls": s.probe.ActiveCalls(), "events": s.probe.QueuedEvents(), "permits": s.probe.Permits(), "waiters": s.probe.AdmissionWaiters(), "bodies": s.gate.bodies.Load()}
	cs.Record("resources", resources)
	for name, n := range resources {
		cs.Check("released "+name, n == 0, "%d", n)
	}
	if supportmatrix.RegionalIsKimi(s.p.report.Combo) {
		u := result.Message.Usage
		cs.Check("actual token counts fit this call reservation", u.Input+u.CacheRead+u.CacheWrite <= int64(s.p.report.Budget.ReservedInputTokens-beforeInput) && u.Output <= int64(outputBudget), "output=%d", u.Output)
	}
	cs.Check("actual own attempt and attribution", row.Attempts == 1 && len(result.Metadata.Attempts) == 1 && result.Metadata.ProviderID == ai.ProviderID(s.p.report.Provider) && result.Metadata.ModelID == row.Model, "attempts=%d", row.Attempts)
	if kind == "cancel" {
		cs.Check("cancel after content retains classification", seenCancel && errors.Is(err, &ai.Error{Code: ai.CodeCanceled}) && result.Message.StopReason == ai.StopReasonAborted, "cancel=%t", seenCancel)
	} else {
		cs.Check("own legal turn succeeds", err == nil && (result.Message.StopReason == ai.StopReasonStop || kind == "tool-call" && result.Message.StopReason == ai.StopReasonToolUse), "stop=%s", result.Message.StopReason)
		cs.Check("reported usage retained", len(result.Metadata.Attempts) == 1 && result.Metadata.Attempts[0].UsageReporting != ai.UsageUnreported, "attempts=%d", len(result.Metadata.Attempts))
		if supportmatrix.RegionalIsKimi(s.p.report.Combo) {
			priced := len(result.Metadata.Attempts) == 1 && (result.Metadata.Attempts[0].CostEstimate.Status == ai.CostResolved || result.Metadata.Attempts[0].CostEstimate.Status == ai.CostBounded)
			cs.Check("finite reliable cost bounds available", priced, "attempts=%d", len(result.Metadata.Attempts))
		}
	}
	if kind == "tool-call" {
		valid := false
		for index, b := range result.Message.Content {
			if _, ok := b.(ai.ToolCall); ok {
				_, e := ai.ValidateToolCall(result.Message, index, req.Tools)
				valid = e == nil
				break
			}
		}
		cs.Check("Host validates complete own tool", valid, "tool validation")
		if valid && err == nil {
			s.tools[row.Model] = result.Message
			if sourceCase == row.CaseID {
				s.native[row.Model] = result.Message
			}
		}
	}
	if kind == "full-stream" && sourceCase == row.CaseID && err == nil {
		if supportmatrix.RegionalIsKimi(s.p.report.Combo) {
			native := false
			for _, b := range result.Message.Content {
				if thought, ok := b.(ai.Thinking); ok && thought.Thinking != "" {
					native = true
				}
			}
			cs.Check("fixed L01 native reasoning exists before any continuation", native, "source=%s", sourceCase)
		}
		s.native[row.Model] = result.Message
	}
	if kind == "tool-result" || kind == "native-history" && !supportmatrix.RegionalIsKimi(s.p.report.Combo) {
		text := ""
		for _, b := range result.Message.Content {
			if b, ok := b.(ai.Text); ok {
				text += b.Text
			}
		}
		cs.Check("visible answer uses Host result", strings.Contains(text, "REGIONAL42"), "result marker present")
	}
	if kind == "native-history" {
		native := false
		for _, b := range s.native[row.Model].Content {
			if v, ok := b.(ai.Thinking); ok && v.Thinking != "" {
				native = true
			}
		}
		cs.Check("own native reasoning retained without downgrade", native && result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "native=%t", native)
	}
	row.Outcome = supportmatrix.Pass
	if t.Failed() {
		row.Outcome, row.ErrorCategory = supportmatrix.Fail, supportmatrix.CategoryAssertion
	}
}
