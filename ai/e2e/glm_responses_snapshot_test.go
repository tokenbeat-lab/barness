package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMResponsesFrozenCallbacks(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM05-frozen-callbacks-"+model)
			w := newGLMResponsesWorld(t, nil)
			w.updateBindingOf(tenantA, "glm-responses", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			format := map[string]any{"type": "json_object"}
			failure := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "Retry-After-Ms": "1"}, Script: `{"error":{"code":"server_error","message":"Synthetic retry."}}`, Chunks: [][]byte{[]byte(`{"error":{"code":"server_error","message":"Synthetic retry."}}`)}, OnReceive: func() { format["type"] = "json_schema" }}
			scenarios, raw := glmResponsesToolFixtures(t)
			ev.Fixture("tools.json", raw)
			var script glmResponsesToolScript
			for _, sc := range scenarios {
				if sc.Model == model && sc.Name == "schema-native-array" {
					script = sc
				}
			}
			enqueue(ev, w, failure, glmResponsesReply(t, script.Tools))
			req := glmResponsesToolRequest()
			opts := ai.ResponsesOptions{SamplingParams: map[string]json.RawMessage{"tool_choice": json.RawMessage(`"auto"`)}}
			entered, resume := make(chan struct{}), make(chan struct{})
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
					p.Body["text"] = map[string]any{"format": format}
					return ai.KeepPayload(), nil
				},
				OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					order = append(order, "response")
					return nil
				},
			}
			s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-responses-frozen"), glmResponsesTarget(model), req, opts)
			defer s.Close()
			select {
			case <-entered:
			case <-ctxFor(t).Done():
				t.Fatal("header callback not reached")
			}
			req.Tools[0].Parameters[0] = '['
			req.Messages[0].(ai.UserMessage).Content[0] = ai.Text{Text: "Changed caller history."}
			opts.SamplingParams["tool_choice"][1] = 'x'
			close(resume)
			for s.Next() {
				if delta, ok := s.Event().(ai.ThinkingDeltaEvent); ok {
					snapshot := delta.Partial.Snapshot()
					if len(snapshot.Content) > 0 {
						snapshot.Content[0] = ai.Text{Text: "Changed event snapshot."}
					}
				}
			}
			res, err := s.Result()
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Record("callback-order", order)
			ev.Check("retry uses frozen legal request", err == nil && len(res.Metadata.Attempts) == 2, "%v %+v", err, res.Metadata)
			ev.Check("callbacks occur once in order", slices.Equal(order, []string{"header", "payload", "response"}), "%v", order)
			reqs := w.provider.Requests()
			if len(reqs) == 2 {
				ev.Check("identical bytes across attempts", bytes.Equal(reqs[0].Body, reqs[1].Body), "%s / %s", reqs[0].Body, reqs[1].Body)
				ev.Check("caller mutation cannot change frozen history or callback format", strings.Contains(string(reqs[0].Body), "Look up alpha") && !strings.Contains(string(reqs[0].Body), "Changed caller") && strings.Contains(string(reqs[0].Body), `"type":"json_object"`), "%s", reqs[0].Body)
			}
			if len(res.Message.Content) > 0 {
				thought, ok := res.Message.Content[0].(ai.Thinking)
				if ok {
					before := thought.Signature
					res.Message.Content[0] = ai.Thinking{Thinking: "Changed Result.", Signature: "{}"}
					again, _ := s.Result()
					value, ok := again.Message.Content[0].(ai.Thinking)
					ev.Check("Result reread retains original native state", ok && value.Signature == before && value.Thinking == "Check both.", "%+v", again.Message)
				}
			}
		})
	}
}
