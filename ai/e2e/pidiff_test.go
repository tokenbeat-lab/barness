package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestPiDifferential is the frozen pi-ai differential (spec Testing Decisions
// §2, §5): it reuses E2E scenarios — same fixture, same logical input, same
// response script — and runs them through both pi-ai and barness-ai. Each
// protocol subtest is that protocol's differential gate; a pending finding
// fails it.
//
// Opt-in: BARNESS_AI_PIDIFF=1 with Node.js and the frozen copy installed
// (npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node).
// Without it the cases are recorded NOT_RUN, never PASS.
func TestPiDifferential(t *testing.T) {
	ledger := loadLedger(t)
	t.Run(string(ai.APIOpenAIResponses), func(t *testing.T) {
		for _, entry := range []string{"stream", "streamSimple"} {
			t.Run(entry, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E01-text-"+entry)
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differentialText(t, ev, o, ledger, entry)
			})
		}
	})
}

func openOracle(t *testing.T) *pioracle.Oracle {
	t.Helper()
	if !pioracle.Enabled() {
		t.Skipf("frozen pi differential not requested (set %s=1)", pioracle.EnableEnv)
	}
	o, err := pioracle.Open()
	if err != nil {
		// Requested but unusable is a failure, not a skip.
		t.Fatal(err)
	}
	return o
}

func loadLedger(t *testing.T) pioracle.Ledger {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "pidiff", "ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	l, err := pioracle.ParseLedger(raw)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// differentialText runs the P01 plain-text scenario (text-basic.json, as in
// TestResponsesText) on both sides and judges the difference.
func differentialText(t *testing.T, ev *evidence.Case, o *pioracle.Oracle, ledger pioracle.Ledger, entry string) {
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("fixture.json", raw)
	reply := sseReply(t, f, provider.FramingLF)
	ev.Record("response-script", reply)
	req := textRequest(f)

	// barness-ai through the public Client.
	w := newWorld(t, tenantA)
	w.provider.Enqueue(reply)
	var s *ai.Stream
	if entry == "stream" {
		s = w.client.Stream(ctxFor(t), textScope("req-pidiff-"+entry), textTarget(f), req, nil)
	} else {
		s = w.client.StreamSimple(ctxFor(t), textScope("req-pidiff-"+entry), textTarget(f), req, ai.SimpleOptions{})
	}
	barness, barnessReq := barnessObservation(t, ev, w, s)

	// Frozen pi-ai against its own local controlled Provider with the same script.
	srv := provider.New(map[string]string{tenantA.secret: tenantA.alias})
	defer srv.Close()
	srv.Enqueue(reply)
	piRun, err := o.Run(ctxFor(t), pioracle.Case{
		API:      string(ai.APIOpenAIResponses),
		Provider: string(ai.ProviderOpenAI),
		Model:    f.Model,
		BaseURL:  srv.URL() + "/v1",
		APIKey:   tenantA.secret,
		Entry:    entry,
		Context:  piContext(t, req),
	})
	if !ev.Check("pi runner completed", err == nil, "%v", err) {
		return
	}
	pi, piReq := piObservation(ev, srv, piRun)
	if !ev.Check("each side sent exactly one inference request", piReq != nil && barnessReq != nil, "pi=%v barness=%v", piReq != nil, barnessReq != nil) {
		return
	}
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", barness)

	diffs, err := pioracle.Compare(pi, barness)
	if !ev.Check("observations compare", err == nil, "%v", err) {
		return
	}
	protocol := string(ai.APIOpenAIResponses)
	verdict := ledger.Classify(protocol, diffs)
	rec := o.Record(ev.ID(), protocol, verdict, pioracle.Sides{
		BarnessVersions: map[string]string{
			"go":        runtime.Version(),
			"openai-go": evidence.ModuleVersion("github.com/openai/openai-go/v3"),
		},
		NodeVersion:      piRun.Node,
		ModelCatalogHash: hashJSON(ai.BuiltinCatalog()),
		PiRequest:        mustMarshal(t, piReq),
		BarnessRequest:   mustMarshal(t, barnessReq),
		Frames:           []byte(reply.Script),
	})
	ev.Record("pidiff", rec)

	var pending []string
	for _, f := range verdict.Findings {
		if f.Classification == pioracle.Pending {
			pending = append(pending, fmt.Sprintf("%s %s (%s)", f.Kind, f.Path, f.Decision))
		}
	}
	ev.Check("differential gate: no pending differences", verdict.Pass, "%d pending:\n  %s", len(pending), strings.Join(pending, "\n  "))
}

// barnessObservation consumes s and projects barness-ai's public results onto
// pi-ai's JSON shape (see piEvent/piMessage). A failed call is still an
// observation: its error message and stop reason are compared like any result.
func barnessObservation(t *testing.T, ev *evidence.Case, w *world, s *ai.Stream) (pioracle.Observation, *provider.Request) {
	t.Helper()
	defer s.Close()
	var obs pioracle.Observation
	var events []ai.Event
	for s.Next() {
		events = append(events, s.Event())
		obs.Events = append(obs.Events, mustMarshal(t, piEvent(t, s.Event())))
	}
	res, err := s.Result()
	ev.Record("barness-events", eventsJSON(events))
	ev.Record("barness-result", map[string]any{"result": res, "error": errString(err)})
	obs.Result = mustMarshal(t, piMessage(t, res.Message))
	req := onlyRequest(ev, "barness-requests", w.provider, &obs)
	return obs, req
}

func piObservation(ev *evidence.Case, srv *provider.Server, r pioracle.Run) (pioracle.Observation, *provider.Request) {
	ev.Record("pi-run", r)
	obs := pioracle.Observation{Events: r.Events, Result: r.Result}
	req := onlyRequest(ev, "pi-requests", srv, &obs)
	return obs, req
}

// onlyRequest records what srv received and, when it was exactly one
// inference request, sets it as obs's request.
func onlyRequest(ev *evidence.Case, name string, srv *provider.Server, obs *pioracle.Observation) *provider.Request {
	reqs := srv.Requests()
	ev.Record(name, reqs)
	if len(reqs) != 1 {
		return nil
	}
	r := reqs[0]
	obs.Request = pioracle.Request{Method: r.Method, Path: r.Path, Query: r.Query, Auth: r.KeyAlias, Headers: r.Header, Body: r.Body}
	return &r
}

// piContext is the pi Context for the same logical input barness receives.
// User messages carry pi's required timestamp; it never reaches the wire.
func piContext(t *testing.T, req ai.Request) json.RawMessage {
	t.Helper()
	messages := []map[string]any{}
	for _, m := range req.Messages {
		um, ok := m.(ai.UserMessage)
		if !ok {
			t.Fatalf("piContext: unsupported message %T", m)
		}
		content := []map[string]any{}
		for _, c := range um.Content {
			text, ok := c.(ai.Text)
			if !ok {
				t.Fatalf("piContext: unsupported user content %T", c)
			}
			content = append(content, map[string]any{"type": "text", "text": text.Text})
		}
		messages = append(messages, map[string]any{"role": "user", "content": content, "timestamp": 0})
	}
	ctx := map[string]any{"messages": messages}
	if req.SystemPrompt != "" {
		ctx["systemPrompt"] = req.SystemPrompt
	}
	return mustMarshal(t, ctx)
}

// piEvent is a barness event in pi-ai's event shape. Barness-only extensions
// (ErrorEvent.Err with Code/Phase) are not part of the pi golden (spec §5) and
// are asserted by the offline E2E instead. An event kind without a mapping
// shows up as a difference rather than being dropped.
func piEvent(t *testing.T, e ai.Event) map[string]any {
	switch e := e.(type) {
	case ai.StartEvent:
		return map[string]any{"type": "start"}
	case ai.TextStartEvent:
		return map[string]any{"type": "text_start", "contentIndex": e.ContentIndex}
	case ai.TextDeltaEvent:
		return map[string]any{"type": "text_delta", "contentIndex": e.ContentIndex, "delta": e.Delta}
	case ai.TextEndEvent:
		return map[string]any{"type": "text_end", "contentIndex": e.ContentIndex, "content": e.Content}
	case ai.DoneEvent:
		return map[string]any{"type": "done", "reason": e.Reason, "message": piMessage(t, e.Message)}
	case ai.ErrorEvent:
		return map[string]any{"type": "error", "reason": e.Reason, "error": piMessage(t, e.Message)}
	default:
		return map[string]any{"type": string(e.Type()), "barnessUnmapped": fmt.Sprintf("%T", e)}
	}
}

// piMessage is a barness AssistantMessage in pi-ai's message shape. Every
// field barness serializes is kept, so a field pi lacks surfaces as a
// difference; only pi's role and content-block type discriminators are added.
func piMessage(t *testing.T, m ai.AssistantMessage) map[string]any {
	t.Helper()
	out := jsonObject(t, m)
	out["role"] = "assistant"
	content := make([]any, 0, len(m.Content))
	for _, c := range m.Content {
		block := jsonObject(t, c)
		switch c.(type) {
		case ai.Text:
			block["type"] = "text"
		default:
			block["type"] = fmt.Sprintf("barness-unmapped:%T", c)
		}
		content = append(content, block)
	}
	out["content"] = content
	return out
}

// jsonObject is v's JSON serialization as a generic object.
func jsonObject(t *testing.T, v any) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(mustMarshal(t, v), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
