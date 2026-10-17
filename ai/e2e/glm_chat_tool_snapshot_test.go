package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMChatToolSnapshots(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM03-tool-snapshots-"+model)
			w := newGLMChatWorld(t, nil)
			sc, raw := glmToolFixture(t, model)
			ev.Fixture("tools.json", raw)
			enqueue(ev, w, chatEvents(t, sc.Tools, ""))
			req := glmToolRequest(model)
			opts := ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"tool_choice": json.RawMessage(`"auto"`)}}
			entered, resume := make(chan struct{}), make(chan struct{})
			thinking := map[string]any{"type": "enabled", "clear_thinking": true}
			var order []string
			hooks := ai.Hooks{
				TransformHeaders: func(ctx context.Context, _ ai.CallScope, _ http.Header) error {
					order = append(order, "header")
					close(entered)
					select {
					case <-resume:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				},
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					order = append(order, "payload")
					p.Body["thinking"] = thinking
					return ai.KeepPayload(), nil
				},
				OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					order = append(order, "response")
					// After serialization, caller-owned hook maps cannot change wire.
					thinking["type"], thinking["clear_thinking"] = "disabled", false
					return nil
				},
			}
			s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-snapshots"), ai.Target{BindingID: "glm-chat", ModelID: model}, req, opts)
			defer s.Close()
			select {
			case <-entered:
			case <-ctxFor(t).Done():
				t.Fatal("header hook was not entered")
			}
			// Stream synchronously snapshots input. These writes are sequenced
			// before the header hook resumes, never concurrent with reads.
			req.Tools[0].Parameters[0] = '['
			req.Messages[1].(ai.UserMessage).Content[0] = ai.Text{Text: "Caller changed prompt."}
			opts.SamplingParams["tool_choice"][1] = 'x'
			close(resume)
			var beforeMutation []json.RawMessage
			for s.Next() {
				e := s.Event()
				beforeMutation = append(beforeMutation, mustMarshal(t, eventsJSON([]ai.Event{e})[0]))
				if delta, ok := e.(ai.ToolCallDeltaEvent); ok {
					snapshot := delta.Partial.Snapshot()
					if len(snapshot.Content) > 0 {
						snapshot.Content[0] = ai.Text{Text: "Caller changed event snapshot."}
					}
				}
			}
			res, err := s.Result()
			ev.Record("events-before-caller-mutation", beforeMutation)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("call remains valid after input mutations", err == nil && len(res.Message.Content) == 3 && res.Message.Content[0].(ai.Thinking).Thinking == "Check both.", "%v %+v", err, res.Message)
			res.Message.Content[0] = ai.Text{Text: "Caller changed result."}
			again, _ := s.Result()
			ev.Check("Result reread owns its content", again.Message.Content[0].(ai.Thinking).Thinking == "Check both.", "%+v", again.Message)
			ev.Check("legal hooks preserve order", slices.Equal(order, []string{"header", "payload", "response"}), "%v", order)
			reqs := w.provider.Requests()
			if !ev.Check("one frozen request", len(reqs) == 1, "%d", len(reqs)) {
				return
			}
			fields := bodyFields(t, reqs[0].Body)
			ev.Check("tool choice and prompt preserve the call snapshot", string(fields["tool_choice"]) == `"auto"` && strings.Contains(string(fields["messages"]), "Look up x twice") && !strings.Contains(string(fields["messages"]), "Caller changed"), "%s", reqs[0].Body)
			ev.Check("callback thinking frozen once", jsonEqual(fields["thinking"], []byte(`{"type":"enabled","clear_thinking":true}`)) && string(fields["tool_stream"]) == "true", "%s", reqs[0].Body)
		})
	}
}

func TestGLMChatSimpleTools(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, entry := range []string{"stream-simple", "complete-simple", "complete"} {
			t.Run(model+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "GLM03-simple-tools-"+model+"-"+entry)
				w := newGLMChatWorld(t, nil)
				sc, raw := glmToolFixture(t, model)
				ev.Fixture("tools.json", raw)
				enqueue(ev, w, chatEvents(t, sc.Tools, ""))
				req, target := glmToolRequest(model), ai.Target{BindingID: "glm-chat", ModelID: model}
				var o outcome
				opts := ai.SimpleOptions{ToolChoice: ai.Value(ai.ToolChoiceAuto)}
				switch entry {
				case "stream-simple":
					o = drain(w.client.StreamSimple(ctxFor(t), textScope("glm-simple-tools"), target, req, opts))
				case "complete-simple":
					o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope("glm-simple-tools"), target, req, opts)
				default:
					o.result, o.err = w.client.Complete(ctxFor(t), textScope("glm-complete-tools"), target, req, nil)
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				_, first := ai.ValidateToolCall(o.result.Message, 1, req.Tools)
				_, second := ai.ValidateToolCall(o.result.Message, 2, req.Tools)
				ev.Check("Host receives two executable calls", o.err == nil && first == nil && second == nil, "%v / %v / %v", o.err, first, second)
				if len(w.provider.Requests()) == 1 {
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					ev.Check("auto tools stream without disabling default thinking", string(fields["tool_stream"]) == "true" && fields["thinking"] == nil, "%s", w.provider.Requests()[0].Body)
				}
			})
		}
	}
}
