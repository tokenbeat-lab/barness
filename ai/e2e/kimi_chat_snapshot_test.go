package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"net/http"
	"strings"
	"testing"
)

func TestKimiChatSnapshots(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, kimiChatID(market, "K14-snapshot"))
			w := newKimiChatWorld(t, market, nil)
			b := w.host.Binding(tenantA.tenant, "kimi-chat")
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			enqueue(ev, w, provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"type":"internal_server_error"}}`)}})
			enqueue(ev, w, kimiChatReply(t))
			entered, resume := make(chan struct{}), make(chan struct{})
			var retained map[string]any
			headers, payloads, responses := 0, 0, 0
			hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, h http.Header) error {
				headers++
				close(entered)
				select {
				case <-resume:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				retained = p.Body
				p.Body["max_completion_tokens"] = 4096
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				responses++
				retained["model"] = "after-serialization"
				return nil
			}}
			req := kimiChatRequest()
			params := map[string]json.RawMessage{"temperature": json.RawMessage(`1`)}
			s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("kimi-snapshot-"+string(market)), kimiChatTarget(), req, ai.ChatOptions{SamplingParams: params})
			defer s.Close()
			waitFor(ev, "header barrier entered", entered)
			req.Messages[0] = ai.UserText("caller-mutated")
			params["temperature"][0] = '9'
			b.Retry.MaxRetries = 0
			w.host.PutBinding(b)
			close(resume)
			o := drain(s)
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("frozen retry and hooks succeed", o.err == nil && headers == 1 && payloads == 1 && responses == 1 && len(w.provider.Requests()) == 2, "%v", o.err)
			if len(w.provider.Requests()) == 2 {
				a, b := w.provider.Requests()[0], w.provider.Requests()[1]
				ev.Check("identical request across attempts", string(a.Body) == string(b.Body) && !strings.Contains(string(a.Body), "caller-mutated") && !strings.Contains(string(a.Body), "after-serialization"), "snapshot escaped")
			}
			if len(o.result.Metadata.Attempts) > 0 {
				o.result.Metadata.Attempts[0].CostEstimate.Status = ai.CostPartial
			}
			o.result.Message.Content = nil
			again, err := s.Result()
			ev.Record("second-result", again)
			ev.Check("Result cost and content own values", err == nil && len(again.Message.Content) > 0 && again.Metadata.Attempts[0].CostEstimate.Status == ai.CostUnreported, "%v %+v", err, again)
		})
	}
}

func TestKimiChatHookAuthority(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, mode := range []string{"credential", "cache-key", "model", "native-reference"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K14-authority-")+mode)
				w := newKimiChatWorld(t, market, nil)
				hooks := ai.Hooks{}
				if mode == "credential" {
					hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
						h.Set("Authorization", "Bearer synthetic-replacement")
						return nil
					}
				} else {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						switch mode {
						case "cache-key":
							p.Body["prompt_cache_key"] = "raw-tenant-session"
						case "model":
							p.Body["model"] = "unauthorized"
						case "native-reference":
							p.Body["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "file", "file_id": "foreign-file"}}}}
						}
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), nil))
				o.record(ev)
				checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
				ev.Check("no HTTP or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
			})
		}
	}
}
