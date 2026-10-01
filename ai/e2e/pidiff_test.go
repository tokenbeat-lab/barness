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
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
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
		for _, scenario := range []string{"text", "interleaved", "tool-call", "tool-results", "reasoning-call", "reasoning-replay", "reasoning-cross-model"} {
			for _, entry := range []string{"stream", "streamSimple"} {
				t.Run(scenario+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "PIDIFF-P01-E01-"+scenario+"-"+entry)
					ev.ReplayEnv(pioracle.EnableEnv + "=1")
					o := openOracle(t)
					differential(t, ev, o, ledger, entry, loadPidiffScenario(t, scenario, entry))
				})
			}
		}
		// tool_choice's object form, on the full entry that offers it.
		t.Run("tool-call-function/stream", func(t *testing.T) {
			ev := run.Case(t, "PIDIFF-P01-E03-tool-call-function-stream")
			ev.ReplayEnv(pioracle.EnableEnv + "=1")
			o := openOracle(t)
			differential(t, ev, o, ledger, "stream", toolPidiffScenario(t, "tool-call-function", "stream"))
		})
		// E03 history normalization, downgrade and images (TestHistoryNormalization).
		hf, hraw := loadHistoryFixture(t)
		for _, sc := range hf.Scenarios {
			for _, entry := range []string{"stream", "streamSimple"} {
				t.Run("history/"+sc.ID+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "PIDIFF-P01-E03-history-"+sc.ID+"-"+entry)
					ev.ReplayEnv(pioracle.EnableEnv + "=1")
					o := openOracle(t)
					differential(t, ev, o, ledger, entry, hf.pidiff(t, hraw, sc))
				})
			}
		}
		// E04 full and simple options, reasoning mapping and presence
		// (TestOptions); each scenario runs on the entry its options belong to.
		of, oraw := loadOptionsFixture(t)
		for _, sc := range of.Scenarios {
			t.Run("options/"+sc.ID+"/"+sc.Entry, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E04-options-"+sc.ID+"-"+sc.Entry)
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, sc.Entry, of.pidiff(t, oraw, sc))
			})
		}
		// E02 failure terminals reuse the offline scenarios. Failure handling
		// does not depend on the entry point's option mapping, so they run on
		// the full entry only.
		f, raw := loadFailureFixture(t)
		for _, sc := range f.Scenarios {
			t.Run("failure/"+sc.ID, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E02-"+sc.ID+"-stream")
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, "stream", failurePidiffScenario(t, f, raw, sc))
			})
		}
		// E05 explicit retry (TestExplicitRetry): with the same retry settings
		// both sides send the same attempts and end the same way.
		rf, rraw := loadRetryFixture(t)
		text, _ := loadTextFixture(t, "text-basic.json")
		for _, sc := range rf.Scenarios {
			if sc.PidiffSkip != "" {
				continue
			}
			t.Run("retry/"+sc.ID, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E05-"+sc.ID+"-stream")
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, "stream", retryPidiffScenario(t, rf, rraw, text, sc))
			})
		}
		// E08 byte limits (TestByteLimits): at its boundary each limit changes
		// nothing pi would do; above it barness ends at resource_limit, an
		// extension registered in the ledger.
		for _, c := range limitPidiffScenarios(t) {
			t.Run("limits/"+c.id, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E08-"+c.id+"-stream")
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, "stream", c.sc)
			})
		}
		// Tool calls cut off by truncation or a failed stream (06).
		tf, traw := loadToolFixture(t)
		for _, sc := range tf.Truncated {
			t.Run("tool/"+sc.ID, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P01-E02-tool-"+sc.ID+"-stream")
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, "stream", pidiffScenario{fixture: "tool-round-trip.json", raw: traw, model: tf.Model,
					req: tf.request(), reply: tf.reply(t, sc.Events, provider.FramingLF, sc.End), requests: 1})
			})
		}
	})
}

// pidiffScenario is one E2E scenario's logical input and response script.
type pidiffScenario struct {
	fixture string
	raw     []byte
	model   string
	req     ai.Request
	reply   provider.Reply
	// replies, when set, script several attempts in place of reply.
	replies []provider.Reply
	// retry is barness's binding retry policy; piOptions carry pi's. clock
	// replaces barness's backoff clock; pi waits for real.
	retry ai.RetryPolicy
	clock *clock.Clock
	// unreachable points both sides at an endpoint nothing listens on.
	unreachable bool
	// abortAfter > 0 makes both callers abort once they received that many
	// events.
	abortAfter int
	// requests is how many inference requests each side must send.
	requests int
	// full and simple are barness's options for the stream and streamSimple
	// entries; piOptions are the same options in pi's shape.
	full      ai.Options
	simple    ai.SimpleOptions
	piOptions map[string]any
	// piOptionsRaw, when set, are pi's options as JSON in place of piOptions.
	piOptionsRaw json.RawMessage
	// modelCompat replaces the model's compat flags and modelPatch other
	// top-level model fields on both sides, as a custom model does; nil
	// keeps the catalogs'.
	modelCompat *ai.ModelCompat
	modelPatch  json.RawMessage
	// policy changes barness's resource policy; pi has none.
	policy func(*ai.ResourcePolicy)
}

// loadPidiffScenario reuses the offline E2E fixtures: P01 plain text
// (TestResponsesText), E01 interleaved blocks (TestStreamLifecycle), both
// rounds of the tool round trip (TestToolRoundTrip) and of the reasoning
// replay (TestNativeStateReplay).
func loadPidiffScenario(t *testing.T, name, entry string) pidiffScenario {
	t.Helper()
	switch name {
	case "text":
		f, raw := loadTextFixture(t, "text-basic.json")
		return pidiffScenario{fixture: "text-basic.json", raw: raw, model: f.Model, req: textRequest(f),
			reply: sseEvents(t, f.Events, provider.FramingLF), requests: 1}
	case "interleaved":
		f, raw := loadLifecycleFixture(t)
		return pidiffScenario{fixture: "interleaved-blocks.json", raw: raw, model: f.Model, req: lifecycleRequest(f),
			reply: sseEvents(t, f.Events, provider.FramingLF), requests: 1}
	case "tool-call", "tool-results":
		return toolPidiffScenario(t, name, entry)
	case "reasoning-call", "reasoning-replay", "reasoning-cross-model":
		return nativePidiffScenario(t, name)
	}
	t.Fatalf("unknown pidiff scenario %q", name)
	return pidiffScenario{}
}

// failurePidiffScenario is E02 scenario sc of TestFailureTerminals.
func failurePidiffScenario(t *testing.T, f failureFixture, raw []byte, sc failureScenario) pidiffScenario {
	t.Helper()
	p := pidiffScenario{fixture: "failures.json", raw: raw, model: f.Model, req: f.request(),
		unreachable: sc.Unreachable, abortAfter: sc.CancelAfterEvents, requests: 1}
	if sc.Expect.Requests != nil {
		p.requests = *sc.Expect.Requests
	}
	if !sc.Unreachable {
		p.reply = f.reply(t, sc)
	}
	return p
}

// replyScripts concatenates the scripted bodies of every attempt.
func replyScripts(replies []provider.Reply) string {
	var b strings.Builder
	for _, r := range replies {
		b.WriteString(r.Script)
	}
	return b.String()
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

// differential runs one scenario on both sides and judges the difference.
func differential(t *testing.T, ev *evidence.Case, o *pioracle.Oracle, ledger pioracle.Ledger, entry string, sc pidiffScenario) {
	ev.Fixture(sc.fixture, sc.raw)
	replies := sc.replies
	if replies == nil {
		replies = []provider.Reply{sc.reply}
	}
	ev.Record("response-script", replies)
	req := sc.req
	target := ai.Target{BindingID: "primary", ModelID: sc.model}

	// barness-ai through the public Client.
	patch := withModelPatch(sc.model, sc.modelCompat, sc.modelPatch)
	w := newWorldWith(t, func(c *ai.Config) {
		patch(c)
		c.Clock = sc.clock
		if sc.policy != nil {
			sc.policy(c.Policy)
		}
	}, tenantA)
	w.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = sc.retry })
	if sc.unreachable {
		w.updateBinding(tenantA, func(b *ai.Binding) { b.Endpoint = deadEndpoint() })
	} else {
		w.provider.Enqueue(replies...)
	}
	var s *ai.Stream
	if entry == "stream" {
		s = w.client.Stream(ctxFor(t), textScope("req-pidiff-"+entry), target, req, sc.full)
	} else {
		s = w.client.StreamSimple(ctxFor(t), textScope("req-pidiff-"+entry), target, req, sc.simple)
	}
	barness, barnessReqs := barnessObservation(t, ev, w, s, sc.abortAfter)

	// Frozen pi-ai against its own local controlled Provider with the same script.
	srv := provider.New(map[string]string{tenantA.secret: tenantA.alias})
	defer srv.Close()
	baseURL := srv.URL() + "/v1"
	if sc.unreachable {
		baseURL = deadEndpoint()
	} else {
		srv.Enqueue(replies...)
	}
	var compat json.RawMessage
	if sc.modelCompat != nil {
		compat = mustMarshal(t, sc.modelCompat)
	}
	options := piOptions(t, sc.piOptions)
	if sc.piOptionsRaw != nil {
		options = sc.piOptionsRaw
	}
	piRun, err := o.Run(ctxFor(t), pioracle.Case{
		ModelCompat:      compat,
		ModelPatch:       sc.modelPatch,
		API:              string(ai.APIOpenAIResponses),
		Provider:         string(ai.ProviderOpenAI),
		Model:            sc.model,
		BaseURL:          baseURL,
		APIKey:           tenantA.secret,
		Entry:            entry,
		Context:          piContext(t, req),
		Options:          options,
		AbortAfterEvents: sc.abortAfter,
	})
	if !ev.Check("pi runner completed", err == nil, "%v", err) {
		return
	}
	pi, piReqs := piObservation(ev, srv, piRun)
	if !ev.Check("each side sent the expected inference requests", piReqs == sc.requests && barnessReqs == sc.requests,
		"pi=%d barness=%d want %d", piReqs, barnessReqs, sc.requests) {
		return
	}
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", barness)

	diffs, err := pioracle.Compare(pi, barness)
	if !ev.Check("observations compare", err == nil, "%v", err) {
		return
	}
	protocol := string(ai.APIOpenAIResponses)
	verdict := ledger.Classify(protocol, ev.ID(), diffs)
	rec := o.Record(ev.ID(), protocol, verdict, pioracle.Sides{
		BarnessVersions: map[string]string{
			"go":        runtime.Version(),
			"openai-go": evidence.ModuleVersion("github.com/openai/openai-go/v3"),
		},
		NodeVersion:      piRun.Node,
		ModelCatalogHash: hashJSON(ai.BuiltinCatalog()),
		PiRequest:        mustMarshal(t, pi.Request),
		BarnessRequest:   mustMarshal(t, barness.Request),
		Frames:           []byte(replyScripts(replies)),
	})
	ev.Record("pidiff", rec)

	var pending []string
	for _, finding := range verdict.Findings {
		if finding.Classification == pioracle.Pending {
			pending = append(pending, fmt.Sprintf("%s %s (%s)", finding.Kind, finding.Path, finding.Decision))
		}
	}
	ev.Check("differential gate: no pending differences", verdict.Pass, "%d pending:\n  %s", len(pending), strings.Join(pending, "\n  "))
}

// barnessObservation consumes s and projects barness-ai's public results onto
// pi-ai's JSON shape (see piEvent/piMessage). A failed call is still an
// observation: its error message and stop reason are compared like any result.
// With abortAfter > 0 another goroutine Closes s once that many events were
// read, as the pi runner aborts its signal. It returns the observation and the
// number of inference requests the provider received.
func barnessObservation(t *testing.T, ev *evidence.Case, w *world, s *ai.Stream, abortAfter int) (pioracle.Observation, int) {
	t.Helper()
	defer s.Close()
	var obs pioracle.Observation
	var events []ai.Event
	closed := make(chan struct{})
	for s.Next() {
		events = append(events, s.Event())
		obs.Events = append(obs.Events, mustMarshal(t, piEvent(t, s.Event())))
		if len(events) == abortAfter {
			go func() { _ = s.Close(); close(closed) }()
		}
	}
	if abortAfter > 0 && len(events) >= abortAfter {
		<-closed
	}
	res, err := s.Result()
	ev.Record("barness-events", eventsJSON(events))
	ev.Record("barness-result", map[string]any{"result": res, "error": errString(err)})
	obs.Result = mustMarshal(t, piMessage(t, res.Message))
	return obs, observeRequests(ev, "barness-requests", w.provider, &obs)
}

func piObservation(ev *evidence.Case, srv *provider.Server, r pioracle.Run) (pioracle.Observation, int) {
	ev.Record("pi-run", r)
	obs := pioracle.Observation{Events: r.Events, Result: r.Result}
	return obs, observeRequests(ev, "pi-requests", srv, &obs)
}

// observeRequests records what srv received and sets the last inference
// request, the one whose response the result came from, as obs's request.
// The attempts before it are compared by count; TestExplicitRetry checks
// they send the same request. It returns the request count.
func observeRequests(ev *evidence.Case, name string, srv *provider.Server, obs *pioracle.Observation) int {
	reqs := srv.Requests()
	ev.Record(name, reqs)
	if len(reqs) > 0 {
		r := reqs[len(reqs)-1]
		obs.Request = pioracle.Request{Method: r.Method, Path: r.Path, Query: r.Query, Auth: r.KeyAlias, Headers: r.Header, Body: r.Body}
	}
	return len(reqs)
}

// piOptions encodes pi stream options, or nothing when there are none.
func piOptions(t *testing.T, opts map[string]any) json.RawMessage {
	if opts == nil {
		return nil
	}
	return mustMarshal(t, opts)
}

// piEvent is a barness event in pi-ai's event shape, taken when the consumer
// receives it. Like the pi runner, which serializes each event as received, the
// live partial becomes a snapshot of the view at that moment. Barness-only
// extensions (ErrorEvent.Err with Code/Phase) are not part of the pi golden
// (spec §5) and are asserted by the offline E2E instead. An event kind without
// a mapping shows up as a difference rather than being dropped.
func piEvent(t *testing.T, e ai.Event) map[string]any {
	block := func(typ string, index int, partial *ai.PartialView, fields ...any) map[string]any {
		out := map[string]any{"type": typ, "contentIndex": index, "partial": piMessage(t, partial.Snapshot())}
		for i := 0; i < len(fields); i += 2 {
			out[fields[i].(string)] = fields[i+1]
		}
		return out
	}
	switch e := e.(type) {
	case ai.StartEvent:
		return map[string]any{"type": "start", "partial": piMessage(t, e.Partial.Snapshot())}
	case ai.TextStartEvent:
		return block("text_start", e.ContentIndex, e.Partial)
	case ai.TextDeltaEvent:
		return block("text_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.TextEndEvent:
		return block("text_end", e.ContentIndex, e.Partial, "content", e.Content)
	case ai.ThinkingStartEvent:
		return block("thinking_start", e.ContentIndex, e.Partial)
	case ai.ThinkingDeltaEvent:
		return block("thinking_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.ThinkingEndEvent:
		return block("thinking_end", e.ContentIndex, e.Partial, "content", e.Content)
	case ai.ToolCallStartEvent:
		return block("toolcall_start", e.ContentIndex, e.Partial)
	case ai.ToolCallDeltaEvent:
		return block("toolcall_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.ToolCallEndEvent:
		return block("toolcall_end", e.ContentIndex, e.Partial, "toolCall", piBlock(t, e.ToolCall))
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
		content = append(content, piBlock(t, c))
	}
	out["content"] = content
	return out
}

// piBlock is a content block in pi's shape: barness's fields plus pi's type
// discriminator.
func piBlock(t *testing.T, c ai.AssistantContent) map[string]any {
	block := jsonObject(t, c)
	switch c.(type) {
	case ai.Text:
		block["type"] = "text"
	case ai.Thinking:
		block["type"] = "thinking"
	case ai.ToolCall:
		block["type"] = "toolCall"
	default:
		block["type"] = fmt.Sprintf("barness-unmapped:%T", c)
	}
	return block
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
