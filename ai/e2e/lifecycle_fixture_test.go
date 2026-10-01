package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// lifecycleFixture is testdata/responses/interleaved-blocks.json.
type lifecycleFixture struct {
	Model        string            `json:"model"`
	SystemPrompt string            `json:"systemPrompt"`
	User         string            `json:"user"`
	Events       []json.RawMessage `json:"events"`
	Expect       struct {
		Request struct {
			Method string          `json:"method"`
			Path   string          `json:"path"`
			Body   json.RawMessage `json:"body"`
		} `json:"request"`
		Events  []string `json:"events"`
		Content []struct {
			Type      string `json:"type"`
			Text      string `json:"text"`
			Signature string `json:"signature"`
		} `json:"content"`
		ResponseID string   `json:"responseId"`
		StopReason string   `json:"stopReason"`
		Usage      ai.Usage `json:"usage"`
	} `json:"expect"`
}

func loadLifecycleFixture(t *testing.T) (lifecycleFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "interleaved-blocks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f lifecycleFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return f, raw
}

func startLifecycleCase(t *testing.T, caseID string, framing provider.Framing) (*evidence.Case, *world, lifecycleFixture) {
	t.Helper()
	ev := run.Case(t, caseID)
	f, raw := loadLifecycleFixture(t)
	ev.Fixture("fixture.json", raw)
	w := newWorld(t, tenantA)
	enqueue(ev, w, sseEvents(t, f.Events, framing))
	return ev, w, f
}

func lifecycleTarget(f lifecycleFixture) ai.Target {
	return ai.Target{BindingID: "primary", ModelID: f.Model}
}

func lifecycleRequest(f lifecycleFixture) ai.Request {
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User)}}
}

// streamEntry is one of the two Stream entry points.
type streamEntry struct {
	name  string
	start func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) *ai.Stream
}

var streamEntries = []streamEntry{
	{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) *ai.Stream {
		return w.client.Stream(ctx, scope, target, req, ai.ResponsesOptions{})
	}},
	{"stream-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) *ai.Stream {
		return w.client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{})
	}},
}

// lifecycleEntry is one of the four public entry points, run to completion and
// checked against the interleaved fixture.
type lifecycleEntry struct {
	name       string
	invokeWith func(ev *evidence.Case, w *world, f lifecycleFixture, scope ai.CallScope, req ai.Request) ai.Result
}

func (e lifecycleEntry) invoke(ev *evidence.Case, w *world, f lifecycleFixture, scope ai.CallScope) ai.Result {
	return e.invokeWith(ev, w, f, scope, lifecycleRequest(f))
}

var lifecycleEntries = func() []lifecycleEntry {
	var out []lifecycleEntry
	for _, se := range streamEntries {
		out = append(out, lifecycleEntry{se.name, func(ev *evidence.Case, w *world, f lifecycleFixture, scope ai.CallScope, req ai.Request) ai.Result {
			s := se.start(ctxFor(ev.T()), w, scope, lifecycleTarget(f), req)
			o := drain(s)
			o.record(ev)
			ev.Check("Result error is nil", o.err == nil, "got %v", o.err)
			ev.Check("Err is nil after a successful terminal", o.streamErr == nil, "got %v", o.streamErr)
			checkLifecycleEvents(ev, f, o.events, o.result)
			return o.result
		}})
	}
	out = append(out,
		lifecycleEntry{"complete-full", func(ev *evidence.Case, w *world, f lifecycleFixture, scope ai.CallScope, req ai.Request) ai.Result {
			res, err := w.client.Complete(ctxFor(ev.T()), scope, lifecycleTarget(f), req, ai.ResponsesOptions{})
			ev.Record("result", res)
			ev.Check("Complete error is nil", err == nil, "got %v", err)
			return res
		}},
		lifecycleEntry{"complete-simple", func(ev *evidence.Case, w *world, f lifecycleFixture, scope ai.CallScope, req ai.Request) ai.Result {
			res, err := w.client.CompleteSimple(ctxFor(ev.T()), scope, lifecycleTarget(f), req, ai.SimpleOptions{})
			ev.Record("result", res)
			ev.Check("Complete error is nil", err == nil, "got %v", err)
			return res
		}},
	)
	return out
}()

// checkLifecycleEvents asserts the full event contract of one drained stream:
// order and block indexes, start first, exactly one terminal and it last, one
// shared live view on every non-terminal event, and the terminal's message
// equal to both the Result and the view's final state.
func checkLifecycleEvents(ev *evidence.Case, f lifecycleFixture, events []ai.Event, res ai.Result) {
	got := eventSummary(events)
	ev.Check("event sequence and block indexes", reflect.DeepEqual(got, f.Expect.Events), "got %q\nwant %q", got, f.Expect.Events)
	ev.Check("start is the first event", len(events) > 0 && events[0].Type() == ai.EventStart, "got %q", got)
	terminals := 0
	for _, e := range events {
		if e.Type() == ai.EventDone || e.Type() == ai.EventError {
			terminals++
		}
	}
	ev.Check("exactly one terminal event, last", terminals == 1 && len(events) > 0 &&
		(events[len(events)-1].Type() == ai.EventDone || events[len(events)-1].Type() == ai.EventError), "got %q", got)
	if len(events) == 0 {
		return
	}
	view := partialOf(events[0])
	ev.Check("start carries a PartialView", view != nil, "event=%#v", events[0])
	for i, e := range events[:len(events)-1] {
		ev.Check("non-terminal events share one live view", partialOf(e) == view, "event %d (%s)", i, e.Type())
	}
	if done, ok := events[len(events)-1].(ai.DoneEvent); ok {
		ev.Check("done and Result carry the same final message", reflect.DeepEqual(done.Message, res.Message),
			"done=%+v\nresult=%+v", done.Message, res.Message)
	}
	if view != nil {
		ev.Check("view settles on the final message", reflect.DeepEqual(view.Snapshot(), res.Message),
			"view=%+v\nresult=%+v", view.Snapshot(), res.Message)
	}
}

func checkLifecycleMessage(ev *evidence.Case, f lifecycleFixture, m ai.AssistantMessage) {
	ev.Check("stop reason", string(m.StopReason) == f.Expect.StopReason, "got %q", m.StopReason)
	ev.Check("no error message", m.ErrorMessage == "", "got %q", m.ErrorMessage)
	ev.Check("response id", m.ResponseID == f.Expect.ResponseID, "got %q", m.ResponseID)
	ev.Check("usage", m.Usage == f.Expect.Usage, "got %+v want %+v", m.Usage, f.Expect.Usage)
	ev.Check("message identity", m.API == ai.APIOpenAIResponses && m.Provider == ai.ProviderOpenAI && m.Model == f.Model,
		"api=%q provider=%q model=%q", m.API, m.Provider, m.Model)
	if !ev.Check("block count", len(m.Content) == len(f.Expect.Content), "got %+v", m.Content) {
		return
	}
	for i, want := range f.Expect.Content {
		var typ, text, sig string
		switch b := m.Content[i].(type) {
		case ai.Text:
			typ, text, sig = "text", b.Text, b.Signature
		case ai.Thinking:
			typ, text, sig = "thinking", b.Thinking, b.Signature
		}
		ev.Check(fmt.Sprintf("block %d", i), typ == want.Type && text == want.Text && sig == want.Signature,
			"got %s %q sig=%q\nwant %s %q sig=%q", typ, text, sig, want.Type, want.Text, want.Signature)
	}
}

func checkLifecycleRequest(ev *evidence.Case, w *world, f lifecycleFixture) {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if !ev.Check("exactly one inference request", len(reqs) == 1, "got %d", len(reqs)) {
		return
	}
	r := reqs[0]
	ev.Check("method and path", r.Method == f.Expect.Request.Method && r.Path == f.Expect.Request.Path, "got %s %s", r.Method, r.Path)
	ev.Check("tenant key used", r.KeyAlias == tenantA.alias, "got %q", r.KeyAlias)
	ev.Check("request body", jsonEqual(r.Body, f.Expect.Request.Body), "got %s\nwant %s", r.Body, f.Expect.Request.Body)
}

// checkGrowingSnapshot reports why snap is not a state the interleaved call
// passes through, or "": each block is a prefix of what it will become (or,
// for thinking replaced at its end, of the streamed deltas), in final order.
func checkGrowingSnapshot(f lifecycleFixture, snap ai.AssistantMessage) string {
	if len(snap.Content) > len(f.Expect.Content) {
		return fmt.Sprintf("%d blocks, want at most %d", len(snap.Content), len(f.Expect.Content))
	}
	streamed := streamedText(f)
	for i, b := range snap.Content {
		want := f.Expect.Content[i]
		got := blockText(b)
		typ := "text"
		if _, ok := b.(ai.Thinking); ok {
			typ = "thinking"
		}
		if typ != want.Type {
			return fmt.Sprintf("block %d is %s, want %s", i, typ, want.Type)
		}
		if !strings.HasPrefix(want.Text, got) && !strings.HasPrefix(streamed[i], got) {
			return fmt.Sprintf("block %d text %q is not a prefix of %q", i, got, want.Text)
		}
	}
	return ""
}

// streamedText is what each block accumulates from its deltas, before an end
// event may replace it with the provider's final text.
func streamedText(f lifecycleFixture) map[int]string {
	out := map[int]string{}
	for _, e := range f.Expect.Events {
		var idx int
		var delta string
		for _, kind := range []string{"text_delta#", "thinking_delta#"} {
			if rest, ok := strings.CutPrefix(e, kind); ok {
				n, _ := fmt.Sscanf(rest, "%d:", &idx)
				_, delta, _ = strings.Cut(rest, ":")
				if n == 1 {
					out[idx] += delta
				}
			}
		}
	}
	return out
}

func blockText(b ai.AssistantContent) string {
	switch b := b.(type) {
	case ai.Text:
		return b.Text
	case ai.Thinking:
		return b.Thinking
	}
	return ""
}

// partialOf returns the live view an event refers to, or nil for terminals.
func partialOf(e ai.Event) *ai.PartialView {
	switch e := e.(type) {
	case ai.StartEvent:
		return e.Partial
	case ai.TextStartEvent:
		return e.Partial
	case ai.TextDeltaEvent:
		return e.Partial
	case ai.TextEndEvent:
		return e.Partial
	case ai.ThinkingStartEvent:
		return e.Partial
	case ai.ThinkingDeltaEvent:
		return e.Partial
	case ai.ThinkingEndEvent:
		return e.Partial
	}
	return nil
}
