package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

func TestMiniMaxChatTerminal(t *testing.T) {
	for _, sc := range []struct {
		name, finish       string
		tools, done, valid bool
		stop               ai.StopReason
	}{
		{"eof", "stop", false, false, true, ai.StopReasonStop}, {"done", "stop", false, true, true, ai.StopReasonStop},
		{"length", "length", false, false, true, ai.StopReasonLength}, {"missing", "", false, false, false, ""},
		{"filtered", "content_filter", false, true, false, ""}, {"unknown", "unexpected", false, false, false, ""},
		{"tool-stop-conflict", "stop", true, false, false, ""}, {"empty-tools", "tool_calls", false, false, false, ""},
		{"tool-finish", "tool_calls", true, true, true, ai.StopReasonToolUse},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM02-terminal-"+sc.name)
			w := newMiniMaxChatWorld(t, nil)
			frames := miniMaxChatFrames("MiniMax-M3")[:2]
			if sc.tools {
				frames = append(frames, json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call-a","function":{"name":"lookup","arguments":"{}"}}]}}]}`))
			}
			if sc.finish != "" {
				frames = append(frames, json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"`+sc.finish+`"}],"usage":{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2},"completion_tokens_details":{"reasoning_tokens":3}}}`))
			}
			if sc.done {
				frames = append(frames, json.RawMessage(`"[DONE]"`))
			}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("terminal agrees with provider", (o.err == nil) == sc.valid && (!sc.valid || o.result.Message.StopReason == sc.stop), "%v %+v", o.err, o.result.Message)
			ev.Check("no stream replay", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata.Attempts)
			ev.Check("raw stop retained", o.result.Message.RawStopReason == sc.finish, "%+v", o.result.Metadata)
			if sc.finish != "" {
				ev.Check("finish usage retained even on errors", o.result.Metadata.Attempts[0].Usage.Input == 10, "%+v", o.result.Metadata.Attempts)
			}
		})
	}
	for _, sc := range []struct{ name, raw string }{{"bad-json", `{"choices":`}, {"error", `{"error":{"type":"server_error","message":"synthetic failure"}}`}} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM02-terminal-"+sc.name)
			w := newMiniMaxChatWorld(t, nil)
			enqueue(ev, w, provider.SSE([][]byte{[]byte("data: " + sc.raw + "\n\n")}))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("stream fails once", o.err != nil && len(w.provider.Requests()) == 1 && o.result.Message.StopReason == ai.StopReasonError, "%v", o.err)
		})
	}
}

func TestMiniMaxChatToolIdentityConflicts(t *testing.T) {
	for _, raw := range []string{
		`[{"index":0,"id":"call-a","function":{"name":"lookup","arguments":"{}"}},{"index":1,"id":"call-a","function":{"name":"lookup","arguments":"{}"}}]`,
		`[{"index":0,"id":"call-a","function":{"name":"lookup","arguments":"{}"}},{"index":0,"id":"call-b","function":{"name":"lookup","arguments":"{}"}}]`,
	} {
		t.Run(raw, func(t *testing.T) {
			ev := run.Case(t, "MM02-tool-identity-"+t.Name())
			w := newMiniMaxChatWorld(t, nil)
			frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"tool_calls":` + raw + `},"finish_reason":"tool_calls"}]}`)}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("conflicting calls are never merged", o.err != nil && o.result.Message.StopReason == ai.StopReasonError, "%v %+v", o.err, o.result.Message)
		})
	}
}
