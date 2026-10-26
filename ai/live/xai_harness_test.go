//go:build live

package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type xaiRoundTripper func(*http.Request) (*http.Response, error)

func (f xaiRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type xaiScenario struct {
	id  string
	run func(*xaiSession)
}
type xaiSession struct {
	t      *testing.T
	cs     *evidence.Case
	p      *xaiProcess
	client *ai.Client
	rec    *recorder
	ctx    context.Context
	row    *supportmatrix.ScenarioResult
	last   []exchange
}

func TestXAICandidate(t *testing.T) { runXAI(t) }
func TestXAIFormal(t *testing.T)    { runXAI(t) }

func runXAI(t *testing.T) {
	p := xaiState
	if p == nil {
		t.Skip("NOT_RUN: explicitly select xai-responses")
	}
	cs := run.Case(t, "E17-xai-process-gate")
	if p.problem != "" {
		cs.Check("selected xAI configuration", false, "%s", p.problem)
		return
	}
	if p.notRun != "" {
		t.Skip("NOT_RUN: " + p.notRun)
	}
	if err := p.claim(); err != nil {
		p.problem = "grant already consumed or claim unavailable"
		cs.Check("claim before requests", false, "%s", p.problem)
		return
	}
	p.report.Outcome = "RUNNING"
	if err := run.Record(p.reportName(), p.report); err != nil {
		t.Fatal("selected xAI initial report unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1770*time.Second)
	defer cancel()
	rec := newRecorder()
	transport := rec.next.(*http.Transport)
	defer transport.CloseIdleConnections()
	if p.proxy != nil {
		transport.Proxy = http.ProxyURL(p.proxy)
	}
	if p.controlled != "" {
		rec.next = xaiRoundTripper(func(r *http.Request) (*http.Response, error) {
			if p.controlled == "cancel" {
				cancel()
				return nil, r.Context().Err()
			}
			return xaiFixtureResponse(r)
		})
	}
	next := rec.next
	rec.next = xaiRoundTripper(func(r *http.Request) (*http.Response, error) {
		b := &p.report.Budget
		if r.URL.String() != xaiEndpoint+"/responses" || b.HTTPAttempts >= b.MaxAttempts || !p.grant.ExpiresAt.After(time.Now().UTC()) {
			return nil, fmt.Errorf("selected xAI transport scope exhausted")
		}
		b.HTTPAttempts++
		if p.controlled == "" {
			b.VendorHTTPAttempts++
		}
		return next.RoundTrip(r)
	})
	c := &combo{name: xaiCombo, operation: ai.OperationChat, provider: ai.ProviderXAI, api: ai.APIOpenAIResponses, endpoint: xaiEndpoint, model: "grok-4.7"}
	policy := livePolicy()
	policy.MaxRequestBytes = 128 << 10
	policy.MaxImageBytes = 4096
	policy.MaxFrameBytes = 256 << 10
	policy.MaxOutputBytes = 512 << 10
	policy.CallTimeout = 65 * time.Second
	policy.ResponseHeaderTimeout = 60 * time.Second
	policy.ReadIdleTimeout = 60 * time.Second
	policy.Image.MaxOutputImageBytes = 512 << 10
	policy.Image.MaxTotalOutputImageBytes = 512 << 10
	client, err := newClientWithPolicy(c, p.key, rec, p.cat, policy)
	if !cs.Check("existing Host Client assembly", err == nil, "Client assembly failed: %v", err) {
		return
	}
	for i, sc := range xaiScenarios() {
		if ctx.Err() != nil || t.Failed() {
			break
		}
		row := &p.report.Scenarios[i]
		began := time.Now()
		beforeCalls, beforeHTTP := p.report.Budget.CallsUsed, p.report.Budget.HTTPAttempts
		ok := t.Run(sc.id, func(t *testing.T) {
			s := &xaiSession{t: t, cs: run.Case(t, "E17-xai-"+sc.id), p: p, client: client, rec: rec, ctx: ctx, row: row}
			s.cs.ReplayRun(p.entry())
			sc.run(s)
		})
		row.DurationMS = time.Since(began).Milliseconds()
		row.Calls = p.report.Budget.CallsUsed - beforeCalls
		row.Attempts = p.report.Budget.HTTPAttempts - beforeHTTP
		row.Outcome = supportmatrix.Pass
		if !ok {
			row.Outcome = supportmatrix.Fail
			if row.ErrorCategory == "" {
				row.ErrorCategory = supportmatrix.CategoryAssertion
			}
			break
		}
	}
}

func (s *xaiSession) call(label, entry string, req ai.Request, o ai.ResponsesOptions) ai.Result {
	s.t.Helper()
	b := &s.p.report.Budget
	if !s.cs.Check("finite call reservation "+label, s.ctx.Err() == nil && b.CallsUsed < b.MaxCalls && b.RequestedOutputTokens+xaiCallOutput <= b.MaxRequestedOutputTokens && b.ReservedInputTokens+8192 <= b.MaxReservedInputTokens && b.ObservedMaximumUSD < s.p.grant.MaxUSD, "selected xAI budget exhausted") {
		return ai.Result{}
	}
	b.CallsUsed++
	b.RequestedOutputTokens += xaiCallOutput
	b.ReservedInputTokens += 8192
	ctx, cancel := context.WithTimeout(s.ctx, 65*time.Second)
	defer cancel()
	target := ai.Target{BindingID: xaiCombo, ModelID: "grok-4.7"}
	scope := newScope()
	o.MaxTokens = ai.Value(xaiCallOutput)
	simple := ai.SimpleOptions{MaxTokens: ai.Value(xaiCallOutput), Reasoning: ai.ThinkingLow}
	var res ai.Result
	var err error
	switch entry {
	case "Stream":
		d := drain(s.client.Stream(ctx, scope, target, req, o), nil)
		res, err = d.res, d.err
		s.recordEvents(label, d.events)
		s.streamProof(label, d)
	case "StreamSimple":
		d := drain(s.client.StreamSimple(ctx, scope, target, req, simple), nil)
		res, err = d.res, d.err
		s.recordEvents(label, d.events)
		s.streamProof(label, d)
	case "CompleteSimple":
		res, err = s.client.CompleteSimple(ctx, scope, target, req, simple)
	default:
		res, err = s.client.Complete(ctx, scope, target, req, o)
	}
	s.last = s.rec.take()
	s.cs.Record(label+"-http", s.last)
	s.cs.Record(label+"-result", res)
	if len(res.Metadata.Attempts) == 1 {
		s.row.ProviderRequestIDs = append(s.row.ProviderRequestIDs, res.Metadata.Attempts[0].ProviderRequestID)
	}
	obs := s.rec.observationsOf(scope.RequestID)
	s.cs.Record(label+"-observer", obs)
	if !s.cs.Check("successful call "+label, err == nil && res.Message.StopReason != ai.StopReasonError && res.Message.StopReason != ai.StopReasonAborted, "call failed") {
		var callErr *ai.Error
		if errors.As(err, &callErr) {
			s.row.ErrorCategory = string(callErr.Code)
		}
		return res
	}
	if !s.cs.Check("one complete capture and attempt "+label, len(s.last) == 1 && !s.last[0].Truncated && len(res.Metadata.Attempts) == 1, "missing complete capture") {
		return res
	}
	s.wireProof(res)
	s.cs.Check("terminal Observer "+label, len(obs) > 0 && obs[len(obs)-1].Kind == ai.ObservationCallFinished, "missing terminal observation")
	u := res.Message.Usage
	if !s.cs.Check("cumulative reported tokens remain exact "+label, u.TotalTokens >= 0 && u.TotalTokens <= math.MaxInt64-b.ReportedTokens, "token accounting overflow") {
		return res
	}
	b.ReportedTokens += u.TotalTokens
	s.p.totalInput += u.Input + u.CacheRead + u.CacheWrite
	s.p.totalOutput += u.Output
	a := res.Metadata.Attempts[0]
	maximum, known := a.CostEstimate.Maximum.Get()
	if known {
		nextUSD := b.ObservedMaximumUSD + maximum.Total
		if !s.cs.Check("finite cumulative USD "+label, !math.IsInf(nextUSD, 0) && !math.IsNaN(nextUSD), "USD accounting overflow") {
			return res
		}
		b.ObservedMaximumUSD = nextUSD
	}
	s.cs.Check("complete usage arithmetic "+label, a.UsageReporting == ai.UsageComplete && u.Input+u.CacheRead+u.CacheWrite+u.Output == u.TotalTokens && u.Input+u.CacheRead <= 8192, "usage incomplete or inconsistent")
	s.cs.Check("finite observed USD "+label, known && (a.CostEstimate.Status == ai.CostResolved || a.CostEstimate.Status == ai.CostBounded) && b.ObservedMaximumUSD <= s.p.grant.MaxUSD, "unknown or exhausted spend; no later calls allowed")
	return res
}

func xaiFixtureResponse(r *http.Request) (*http.Response, error) {
	// This is an ordinary synthetic peer behind the existing Transport seam.
	// Its counters and PASS remain controlled, with zero vendor attempts.
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var req map[string]any
	if json.Unmarshal(raw, &req) != nil {
		return nil, fmt.Errorf("invalid synthetic request")
	}
	suffix := strconv.Itoa(xaiState.report.Budget.HTTPAttempts)
	output := []any{map[string]any{"type": "reasoning", "id": "rs-synthetic-" + suffix, "summary": []any{map[string]any{"type": "summary_text", "text": "Synthetic reasoning."}}, "encrypted_content": "synthetic-ciphertext-" + suffix, "future_field": "retained"}}
	choice, _ := req["tool_choice"].(string)
	tools, _ := req["tools"].([]any)
	if len(tools) > 0 && choice != "none" {
		count := 1
		if req["parallel_tool_calls"] == true {
			count = 2
		}
		for i := 0; i < count; i++ {
			city := "Paris"
			if i == 1 {
				city = "Tokyo"
			}
			args, _ := json.Marshal(map[string]string{"city": city})
			output = append(output, map[string]any{"type": "function_call", "id": "item|synthetic-" + suffix + "-" + strconv.Itoa(i), "call_id": "call|synthetic-" + suffix + "-" + strconv.Itoa(i), "name": "get_weather", "arguments": string(args), "status": "completed"})
		}
	} else {
		output = append(output, map[string]any{"type": "message", "id": "msg-synthetic-" + suffix, "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": "391. Red. Hello.", "annotations": []any{}}}})
	}
	reasoning, total := 110, 151
	if xaiState.report.Stage == "formal" {
		reasoning, total = 1100, 1141
	} // Reasoning is not part of the visible1024 cap.
	response := map[string]any{"id": "resp-synthetic-" + suffix, "object": "response", "status": "completed", "model": "grok-4.7", "output": output, "usage": map[string]any{"input_tokens": 32, "input_tokens_details": map[string]any{"cached_tokens": 0}, "output_tokens": 9, "output_tokens_details": map[string]any{"reasoning_tokens": reasoning}, "total_tokens": total}}
	event, _ := json.Marshal(map[string]any{"type": "response.completed", "response": response})
	body := "event: response.completed\ndata: " + string(event) + "\n\ndata: [DONE]\n\n"
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
}
