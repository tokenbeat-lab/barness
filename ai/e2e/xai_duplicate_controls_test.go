package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIDuplicateProtocolControls(t *testing.T) {
	for _, source := range []string{"call", "model", "profile", "hook"} {
		for _, sc := range []struct{ name, field, raw string }{
			{"reasoning-effort", "reasoning", `{"effort":"off","effort":"high"}`},
			{"reasoning-escaped-key", "reasoning", `{"effort":"off","eff\u006frt":"high"}`},
			{"tool-type", "tools", `[{"type":"custom","type":"function","name":"probe","parameters":{"type":"object"}}]`},
			{"input-role", "input", `[{"role":"developer","role":"user","content":"hello"}]`},
			{"input-content", "input", `[{"role":"user","content":[{"type":"input_text","text":"one","text":"two"}]}]`},
			{"text-format", "text", `{"format":{"type":"json_object","type":"text"}}`},
		} {
			// Catalog sampling never permits tools/input overrides. Those public
			// fields reach the wire through call options or Hooks instead.
			if (source == "model" || source == "profile") && (sc.field == "tools" || sc.field == "input") {
				continue
			}
			t.Run(source+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, "XAI-review-duplicate-"+source+"-"+sc.name)
				sampling := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
				w := newXAIWorld(t, func(c *ai.Catalog) {
					if source == "model" {
						c.Models[0].SamplingParams = sampling
					}
					if source == "profile" {
						c.Models[0].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingHigh: sampling}
					}
				})
				enqueue(ev, w, xaiReply(t))
				opts, hooks := ai.ResponsesOptions{}, ai.Hooks{}
				if source == "call" {
					opts.SamplingParams = sampling
				}
				if source == "hook" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						p.Body[sc.field] = json.RawMessage(sc.raw)
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), opts))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				if source == "hook" {
					code, phase = ai.CodeCallbackFailed, ai.PhaseRequest
				}
				checkRejected(ev, o, code, phase)
				ev.Check("ambiguous controls never reach HTTP", len(w.provider.Requests()) == 0, "requests=%d", len(w.provider.Requests()))
				if source != "hook" {
					ev.Check("preflight reads no credentials", w.host.ReadsFor(t.Name()).Credentials == 0, "reads=%+v", w.host.ReadsFor(t.Name()))
				}
			})
		}
	}
}

func TestXAIDuplicateControlsAuthorityAndBusiness(t *testing.T) {
	for _, mode := range []string{"widen-model", "widen-native", "business"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-review-duplicate-scope-"+mode)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				if mode == "business" {
					p.Body["tools"] = json.RawMessage(`[{"type":"function","name":"probe","parameters":{"type":"object","properties":{"model":{"type":"string"},"store":{"type":"boolean"},"file_id":{"type":"string"}},"examples":[{"model":"first","model":"second","store":true,"file_id":"business"}]}}]`)
					p.Body["input"] = json.RawMessage(`[{"type":"function_call","name":"probe","call_id":"c","arguments":"{\"model\":\"first\",\"model\":\"second\",\"store\":true,\"file_id\":\"business\"}"},{"type":"function_call_output","call_id":"c","output":"{\"store\":true,\"file_id\":\"business\"}"}]`)
				} else {
					p.Body["reasoning"] = json.RawMessage(`{"effort":"off","effort":"high"}`)
					if mode == "widen-model" {
						p.Body["model"] = "foreign"
					}
					if mode == "widen-native" {
						p.Body["input"] = json.RawMessage(`[{"role":"assistant","id":"foreign","content":"history"}]`)
					}
				}
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if mode == "business" {
				ev.Check("business schema and string JSON keep independent names", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
			} else {
				checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
				ev.Check("authority widening wins before HTTP", len(w.provider.Requests()) == 0, "requests=%d", len(w.provider.Requests()))
			}
		})
	}
}
