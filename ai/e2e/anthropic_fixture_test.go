package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// anthropicFixture is a testdata/anthropic/*.json file: P02 scenarios for
// Anthropic × Messages. Each scenario is one call — logical input, options
// in pi's shape, the response script — with what the Provider must receive
// and what the call must produce. The offline E2E runs them through the
// public Client; the pi differential runs the same scenarios through frozen
// pi-ai.
type anthropicFixture struct {
	Case      string              `json:"case"`
	Source    string              `json:"source"`
	Scenarios []anthropicScenario `json:"scenarios"`
}

type anthropicScenario struct {
	ID    string `json:"id"`
	About string `json:"about"`
	Model string `json:"model"`
	// Entry is "stream" (the default) or "streamSimple"; the offline E2E
	// also runs a stream scenario through Complete.
	Entry string `json:"entry"`
	// ModelPatch replaces top-level fields of the catalog model on both
	// sides, as a custom model does.
	ModelPatch json.RawMessage `json:"modelPatch"`
	Context    struct {
		SystemPrompt string            `json:"systemPrompt"`
		Tools        []ai.Tool         `json:"tools"`
		Messages     []json.RawMessage `json:"messages"`
	} `json:"context"`
	// Options are pi's options for the entry; barness decodes the same JSON
	// into AnthropicOptions or SimpleOptions.
	Options json.RawMessage `json:"options"`
	// Events script a single 200 SSE reply; Replies script several
	// attempts instead.
	Events  []json.RawMessage `json:"events"`
	Replies []anthropicReply  `json:"replies"`
	// End is how the single SSE reply ends ("abort", "hold").
	End provider.End `json:"end"`
	// CancelAfterEvents > 0 cancels the call once that many events were
	// received (stream entries only).
	CancelAfterEvents int `json:"cancelAfterEvents"`
	// MaxRetries is the binding's retry policy (pi's maxRetries option).
	MaxRetries int `json:"maxRetries"`
	// PidiffSkip says why the differential does not run the scenario.
	PidiffSkip string          `json:"pidiffSkip"`
	Expect     anthropicExpect `json:"expect"`
	// framing lays out SSE replies; empty is LF.
	framing provider.Framing
}

// anthropicReply is one scripted response: SSE events, or a status with a
// raw body.
type anthropicReply struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
	Events []json.RawMessage `json:"events"`
	// SSE is a raw event-stream body, for framings EncodeSSE cannot express.
	SSE string       `json:"sse"`
	End provider.End `json:"end"`
}

type anthropicExpect struct {
	// Requests is how many inference requests are sent; 0 means 1.
	Requests int `json:"requests"`
	// Request is the last inference request; headers are a subset, and
	// AbsentHeaders must not be sent.
	Request *struct {
		Path          string            `json:"path"`
		Query         string            `json:"query"`
		Headers       map[string]string `json:"headers"`
		AbsentHeaders []string          `json:"absentHeaders"`
		Body          json.RawMessage   `json:"body"`
	} `json:"request"`
	// Events is the event summary (see eventSummary); empty skips it.
	Events       []string `json:"events"`
	StopReason   string   `json:"stopReason"`
	ErrorMessage string   `json:"errorMessage"`
	Code         ai.Code  `json:"code"`
	Phase        ai.Phase `json:"phase"`
	// HTTPStatus and ProviderRequestID are the classified error's.
	HTTPStatus        int    `json:"httpStatus"`
	ProviderRequestID string `json:"providerRequestId"`
	RawStopReason     string `json:"rawStopReason"`
	ResponseID        string `json:"responseId"`
	ResponseModel     string `json:"responseModel"`
	// Content is the final content in pi's block shape (with barness's
	// rawArguments); null skips it.
	Content        json.RawMessage           `json:"content"`
	Usage          *ai.Usage                 `json:"usage"`
	UsageReporting ai.UsageReporting         `json:"usageReporting"`
	Downgrades     *ai.NativeStateDowngrades `json:"downgrades"`
}

func loadAnthropicFixture(t *testing.T, name string) (anthropicFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "anthropic", name))
	if err != nil {
		t.Fatal(err)
	}
	var f anthropicFixture
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return f, raw
}

func (sc anthropicScenario) entry() string {
	if sc.Entry == "" {
		return "stream"
	}
	return sc.Entry
}

func (sc anthropicScenario) target() ai.Target {
	return ai.Target{BindingID: "claude", ModelID: sc.Model}
}

func (sc anthropicScenario) request(t *testing.T) ai.Request {
	t.Helper()
	req := ai.Request{SystemPrompt: sc.Context.SystemPrompt, Tools: sc.Context.Tools}
	for _, raw := range sc.Context.Messages {
		req.Messages = append(req.Messages, decodeStoredMessage(t, raw, tenantA))
	}
	return req
}

func (sc anthropicScenario) fullOptions(t *testing.T) ai.AnthropicOptions {
	t.Helper()
	var o ai.AnthropicOptions
	if sc.Options != nil {
		mustUnmarshal(t, sc.Options, &o)
	}
	return o
}

func (sc anthropicScenario) simpleOptions(t *testing.T) ai.SimpleOptions {
	t.Helper()
	var o ai.SimpleOptions
	if sc.Options != nil {
		mustUnmarshal(t, sc.Options, &o)
	}
	return o
}

// replies scripts the scenario's responses.
func (sc anthropicScenario) replies(t *testing.T) []provider.Reply {
	t.Helper()
	framing := sc.framing
	if framing == "" {
		framing = provider.FramingLF
	}
	if sc.Replies == nil {
		return []provider.Reply{anthropicReply{Status: http.StatusOK, Events: sc.Events, End: sc.End}.script(t, framing)}
	}
	out := make([]provider.Reply, len(sc.Replies))
	for i, r := range sc.Replies {
		out[i] = r.script(t, framing)
	}
	return out
}

func (r anthropicReply) script(t *testing.T, framing provider.Framing) provider.Reply {
	t.Helper()
	var reply provider.Reply
	switch {
	case r.Events != nil:
		reply = sseEvents(t, r.Events, framing)
	case r.SSE != "":
		reply = provider.SSE([][]byte{[]byte(r.SSE)})
	default:
		reply = provider.Reply{Status: r.Status, Header: map[string]string{"Content-Type": "application/json"},
			Chunks: [][]byte{[]byte(r.Body)}, Script: r.Body}
	}
	if r.Status != 0 {
		reply.Status = r.Status
	}
	for k, v := range r.Header {
		if reply.Header == nil {
			reply.Header = map[string]string{}
		}
		reply.Header[k] = v
	}
	reply.End = r.End
	return reply
}

// anthropicWorld assembles tenant A's world for the scenario: its model
// patch and its binding retry policy.
func anthropicWorld(t *testing.T, sc anthropicScenario) *world {
	t.Helper()
	w := newWorldWith(t, withModelPatch(sc.Model, nil, sc.ModelPatch), tenantA)
	if sc.MaxRetries > 0 {
		b := w.host.Binding(tenantA.tenant, "claude")
		b.Retry = ai.RetryPolicy{MaxRetries: sc.MaxRetries}
		w.host.PutBindingAt(tenantA.tenant, "claude", b)
	}
	return w
}

// anthropicMode is how the offline E2E calls the scenario.
type anthropicMode string

const (
	// modeStream consumes every event, then awaits Result.
	modeStream anthropicMode = "stream"
	// modeResult awaits Result without ever calling Next.
	modeResult anthropicMode = "result"
	// modeComplete calls Complete.
	modeComplete anthropicMode = "complete"
)

// runAnthropic runs sc once through the public Client in mode and asserts
// everything the scenario expects. It returns the outcome for further checks.
func runAnthropic(t *testing.T, ev *evidence.Case, sc anthropicScenario, raw []byte, mode anthropicMode) (*world, outcome) {
	t.Helper()
	ev.Fixture("fixture.json", raw)
	w := anthropicWorld(t, sc)
	enqueue(ev, w, sc.replies(t)...)
	o := invokeAnthropic(t, w, sc, mode, "req-"+sc.ID+"-"+string(mode))
	o.record(ev)
	checkAnthropic(ev, w, sc, mode, o)
	return w, o
}

func invokeAnthropic(t *testing.T, w *world, sc anthropicScenario, mode anthropicMode, requestID string) outcome {
	t.Helper()
	ctx := ctxFor(t)
	scope := textScope(requestID)
	req := sc.request(t)
	simple := sc.entry() == "streamSimple"
	if mode == modeComplete {
		var res ai.Result
		var err error
		if simple {
			res, err = w.client.CompleteSimple(ctx, scope, sc.target(), req, sc.simpleOptions(t))
		} else {
			res, err = w.client.Complete(ctx, scope, sc.target(), req, sc.fullOptions(t))
		}
		return outcome{result: res, err: err}
	}
	var s *ai.Stream
	if simple {
		s = w.client.StreamSimple(ctx, scope, sc.target(), req, sc.simpleOptions(t))
	} else {
		s = w.client.Stream(ctx, scope, sc.target(), req, sc.fullOptions(t))
	}
	defer s.Close()
	o := outcome{streamed: mode == modeStream}
	if mode == modeStream {
		if sc.CancelAfterEvents > 0 {
			o = drainClosingAfter(s, sc.CancelAfterEvents)
		} else {
			for s.Next() {
				o.events = append(o.events, s.Event())
			}
			o.streamErr = s.Err()
		}
	}
	done := make(chan struct{})
	go func() {
		o.result, o.err = s.Result()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(waitDeadline):
		t.Fatalf("Result did not return within %s", waitDeadline)
	}
	return o
}

// checkAnthropic asserts the scenario's expectations on one outcome.
func checkAnthropic(ev *evidence.Case, w *world, sc anthropicScenario, mode anthropicMode, o outcome) {
	t := ev.T()
	x := sc.Expect
	m := o.result.Message
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	wantRequests := max(x.Requests, 1)
	ev.Check("inference requests sent", len(reqs) == wantRequests, "got %d want %d", len(reqs), wantRequests)
	if r := x.Request; r != nil && len(reqs) > 0 {
		got := reqs[len(reqs)-1]
		ev.Check("request path", got.Method == http.MethodPost && got.Path == r.Path && got.Query == r.Query,
			"got %s %s?%s", got.Method, got.Path, got.Query)
		ev.Check("tenant's Anthropic key used", got.KeyAlias == anthropicA.alias, "got %q", got.KeyAlias)
		for k, v := range r.Headers {
			ev.Check("header "+k, got.Header[k] == v, "got %q want %q", got.Header[k], v)
		}
		for _, k := range r.AbsentHeaders {
			_, present := got.Header[k]
			ev.Check("header "+k+" not sent", !present, "got %q", got.Header[k])
		}
		if r.Body != nil {
			ev.Check("request body", jsonEqual(got.Body, r.Body), "got %s\nwant %s", got.Body, r.Body)
		}
	}
	if x.Events != nil && mode == modeStream {
		got := eventSummary(o.events)
		ev.Check("event sequence", reflect.DeepEqual(got, x.Events), "got %q\nwant %q", got, x.Events)
	}
	ev.Check("stop reason", string(m.StopReason) == x.StopReason, "got %q want %q", m.StopReason, x.StopReason)
	ev.Check("error message", m.ErrorMessage == x.ErrorMessage, "got %q want %q", m.ErrorMessage, x.ErrorMessage)
	failed := x.StopReason == "error" || x.StopReason == "aborted"
	var aiErr *ai.Error
	if failed {
		ev.Check("classified error", errors.As(o.err, &aiErr) && aiErr.Code == x.Code && aiErr.Phase == x.Phase,
			"got %v want %s (%s)", o.err, x.Code, x.Phase)
		ev.Check("error message equals the message's", aiErr != nil && aiErr.Message == m.ErrorMessage, "got %v", o.err)
		ev.Check("error HTTP status and vendor request id", aiErr != nil && aiErr.HTTPStatus == x.HTTPStatus &&
			aiErr.ProviderRequestID == x.ProviderRequestID, "got %+v", aiErr)
	} else {
		ev.Check("no error", o.err == nil, "got %v", o.err)
	}
	if mode == modeStream {
		ev.Check("Err after the terminal is Result's error", errors.Is(o.streamErr, o.err) && (o.err == nil) == (o.streamErr == nil),
			"Err=%v Result=%v", o.streamErr, o.err)
		if n := len(o.events); n > 0 {
			switch e := o.events[n-1].(type) {
			case ai.DoneEvent:
				ev.Check("done carries the final message", !failed && reflect.DeepEqual(e.Message, m), "done=%+v", e.Message)
			case ai.ErrorEvent:
				ev.Check("error event carries the final message", failed && reflect.DeepEqual(e.Message, m), "error=%+v", e.Message)
			default:
				ev.Check("stream ends with a terminal event", false, "last %T", e)
			}
		}
	}
	ev.Check("message identity", m.API == ai.APIAnthropicMessages && m.Provider == ai.ProviderAnthropic && m.Model == sc.Model,
		"api=%q provider=%q model=%q", m.API, m.Provider, m.Model)
	ev.Check("raw stop reason", m.RawStopReason == x.RawStopReason, "got %q want %q", m.RawStopReason, x.RawStopReason)
	ev.Check("response id", m.ResponseID == x.ResponseID, "got %q want %q", m.ResponseID, x.ResponseID)
	ev.Check("response model", m.ResponseModel == x.ResponseModel, "got %q want %q", m.ResponseModel, x.ResponseModel)
	if x.Content != nil {
		got := []any{}
		for _, c := range m.Content {
			got = append(got, piBlock(t, c))
		}
		ev.Check("final content", jsonEqual(mustMarshal(t, got), x.Content), "got %s\nwant %s", mustMarshal(t, got), x.Content)
	}
	if x.Usage != nil {
		ev.Check("usage", m.Usage == *x.Usage, "got %+v\nwant %+v", m.Usage, *x.Usage)
	}
	meta := o.result.Metadata
	if x.UsageReporting != "" {
		var last ai.Attempt
		if n := len(meta.Attempts); n > 0 {
			last = meta.Attempts[n-1]
		}
		ev.Check("last attempt's usage reporting", last.UsageReporting == x.UsageReporting, "got %q want %q", last.UsageReporting, x.UsageReporting)
		ev.Check("last attempt's usage is the message's", last.Usage == m.Usage, "attempt %+v message %+v", last.Usage, m.Usage)
	}
	if x.Downgrades != nil {
		ev.Check("native state downgrades", meta.NativeStateDowngrades == *x.Downgrades, "got %+v want %+v", meta.NativeStateDowngrades, *x.Downgrades)
	}
	ev.Check("call resolved to the Anthropic binding", meta.Resolved && meta.ProviderID == ai.ProviderAnthropic &&
		meta.API == ai.APIAnthropicMessages && meta.AccountScopeID == "acct-tenant-a-anthropic", "metadata %+v", meta)
	ev.Check("attempts recorded", len(meta.Attempts) == wantRequests, "got %d want %d", len(meta.Attempts), wantRequests)
}
