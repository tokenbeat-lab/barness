package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIRequestContract(t *testing.T) {
	for _, entry := range []string{"complete", "stream", "complete-simple", "stream-simple"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "XAI-E02-request-"+entry)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			var o outcome
			switch entry {
			case "complete":
				o.result, o.err = w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), ai.ResponsesOptions{MaxTokens: ai.Value(1)})
			case "stream":
				o = drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			case "complete-simple":
				o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), ai.SimpleOptions{})
			default:
				o = drain(w.client.StreamSimple(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), ai.SimpleOptions{}))
			}
			o.record(ev)
			reqs := w.provider.Requests()
			ev.Record("requests", reqs)
			if !ev.Check("one successful request", o.err == nil && len(reqs) == 1, "err=%v requests=%d", o.err, len(reqs)) {
				return
			}
			b := bodyFields(t, reqs[0].Body)
			ev.Check("unconditional stateless encrypted request", string(b["store"]) == "false" && string(b["stream"]) == "true" && string(b["include"]) == `["reasoning.encrypted_content"]`, "%s", reqs[0].Body)
			_, reasoning := b["reasoning"]
			ev.Check("omission keeps native effort", !reasoning, "%s", reqs[0].Body)
			var input []map[string]any
			mustUnmarshal(t, b["input"], &input)
			ev.Check("system remains system", len(input) == 2 && input[0]["role"] == "system", "%v", input)
			want := "128000"
			if entry == "complete" {
				want = "1"
			}
			ev.Check("exact output budget", string(b["max_output_tokens"]) == want, "%s", reqs[0].Body)
		})
	}
}

func TestXAIFinalGuard(t *testing.T) {
	for _, sc := range []struct {
		name   string
		change func(map[string]any)
		code   ai.Code
	}{
		{"missing-store", func(b map[string]any) { delete(b, "store") }, ai.CodeCallbackFailed},
		{"missing-include", func(b map[string]any) { delete(b, "include") }, ai.CodeCallbackFailed},
		{"stream-false", func(b map[string]any) { b["stream"] = false }, ai.CodeCallbackFailed},
		{"unsupported-field", func(b map[string]any) { b["metadata"] = map[string]any{} }, ai.CodeCallbackFailed},
		{"model", func(b map[string]any) { b["model"] = "other" }, ai.CodeTenantDenied},
		{"store-true", func(b map[string]any) { b["store"] = true }, ai.CodeTenantDenied},
		{"cache", func(b map[string]any) { b["prompt_cache_key"] = "foreign" }, ai.CodeTenantDenied},
		{"resource", func(b map[string]any) { b["input"] = []any{map[string]any{"type": "item_reference", "id": "foreign"}} }, ai.CodeTenantDenied},
		{"budget-precise", func(b map[string]any) { b["max_output_tokens"] = json.Number("1.00000000000000001") }, ai.CodeCallbackFailed},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E03-final-guard-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				sc.change(p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			checkRejected(ev, o, sc.code, ai.PhaseRequest)
			ev.Check("no HTTP or attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}
