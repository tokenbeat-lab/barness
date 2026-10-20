package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func regionalMetadataUsageScript(t *testing.T, b ai.Binding, usage json.RawMessage) (provider.Reply, []byte) {
	t.Helper()
	model := regionalMetadataModel(b.ProviderID)
	frame := map[string]any{}
	switch b.API {
	case ai.APIOpenAICompletions:
		frame = map[string]any{"id": "synthetic-usage", "model": model, "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"content": "Synthetic usage."}, "finish_reason": "stop"}}}
		if usage != nil {
			frame["usage"] = usage
		}
		raw := mustMarshal(t, frame)
		return chatEvents(t, []json.RawMessage{raw}, ""), raw
	case ai.APIOpenAIResponses:
		response := map[string]any{"id": "synthetic-usage", "model": model, "status": "completed", "output": []any{}}
		if usage != nil {
			response["usage"] = usage
		}
		frame = map[string]any{"type": "response.completed", "response": response}
		raw := mustMarshal(t, frame)
		return sseEvents(t, []json.RawMessage{raw}, provider.FramingLF), raw
	default:
		message := map[string]any{"id": "synthetic-usage", "model": model, "role": "assistant", "type": "message", "content": []any{}}
		delta := map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": "end_turn"}}
		if usage != nil {
			message["usage"] = usage
			delta["usage"] = usage
		}
		frames := []json.RawMessage{mustMarshal(t, map[string]any{"type": "message_start", "message": message}), mustMarshal(t, delta), json.RawMessage(`{"type":"message_stop"}`)}
		return sseEvents(t, frames, provider.FramingLF), mustMarshal(t, frames)
	}
}

func TestRegionalProviderCatalogUsagePresence(t *testing.T) {
	cat := regionalMetadataCatalog()
	w := regionalMetadataWorld(t, &cat)
	for pi, p := range regionalRuntimeProviders {
		for ri, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			for _, state := range []string{"zero", "partial", "missing"} {
				t.Run(string(p)+"/"+string(api)+"/"+state, func(t *testing.T) {
					ev := run.Case(t, "RP17-usage-presence-"+string(p)+"-"+string(api)+"-"+state)
					b := regionalMetadataBinding(w, p, api)
					var usage json.RawMessage
					if state == "partial" {
						usage = json.RawMessage(`{"input_tokens":8}`)
						if api == ai.APIOpenAICompletions {
							usage = json.RawMessage(`{"prompt_tokens":8}`)
						}
					}
					if state == "zero" {
						switch api {
						case ai.APIOpenAICompletions:
							usage = json.RawMessage(`{"prompt_tokens":0,"completion_tokens":0,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}`)
						case ai.APIOpenAIResponses:
							usage = json.RawMessage(`{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`)
						default:
							usage = json.RawMessage(`{"input_tokens":0,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}`)
						}
					}
					reply, raw := regionalMetadataUsageScript(t, b, usage)
					ev.Fixture("synthetic-script.json", raw)
					w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
					out := regionalHostCall(t, w.client, b, regionalMetadataModel(p), "complete", t.Name())
					out.record(ev)
					report := ai.UsageComplete
					if state == "partial" {
						report = ai.UsagePartial
					}
					if state == "missing" {
						report = ai.UsageUnreported
					}
					u := out.result.Message.Usage
					want := 0.0
					wantInput := int64(0)
					if state == "partial" {
						want = 8 * float64((pi+1)*10+ri+1) / 1e6
						wantInput = 8
					}
					ev.Check("known zero, partial and absent usage stay distinct beside fixed prices", out.err == nil && len(out.result.Metadata.Attempts) == 1 && out.result.Metadata.Attempts[0].UsageReporting == report && u.Input == wantInput && math.Abs(u.Cost.Total-want) < 1e-12, "err=%v attempts=%+v usage=%+v", out.err, out.result.Metadata.Attempts, u)
					ev.Record("interpretation", "Synthetic positive rates; zero estimated cost with absent usage is not a free-call or billing statement. Partial usage prices only reported known tokens.")
				})
			}
		}
	}
}

func TestRegionalProviderCatalogCNYTierBoundary(t *testing.T) {
	cat, err := regional.CandidateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	w := regionalMetadataWorld(t, &cat)
	for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
		for _, input := range []int{512000, 512001} {
			t.Run(fmt.Sprintf("%s/%d", api, input), func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("RP17-CNY-USD-tier-%s-%d", api, input))
				b := regionalMetadataBinding(w, ai.ProviderMiniMaxCN, api)
				uncached := input - 100
				usage := json.RawMessage(fmt.Sprintf(`{"input_tokens":%d,"output_tokens":20,"total_tokens":%d,"input_tokens_details":{"cached_tokens":100},"output_tokens_details":{"reasoning_tokens":0}}`, input, input+20))
				if api == ai.APIOpenAICompletions {
					usage = json.RawMessage(fmt.Sprintf(`{"prompt_tokens":%d,"completion_tokens":20,"prompt_tokens_details":{"cached_tokens":100},"completion_tokens_details":{"reasoning_tokens":0}}`, input))
				}
				if api == ai.APIAnthropicMessages {
					usage = json.RawMessage(fmt.Sprintf(`{"input_tokens":%d,"output_tokens":20,"cache_read_input_tokens":100,"cache_creation_input_tokens":0}`, uncached))
				}
				reply, raw := regionalMetadataUsageScript(t, b, usage)
				ev.Fixture("synthetic-source-price-script.json", raw)
				w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
				out := regionalHostCall(t, w.client, b, "MiniMax-M3", "complete", t.Name())
				out.record(ev)
				// Independent dated official CNY2.10/8.40/.42 divided by CFETS6.7220,
				// rounded to ten decimals; >512000 uses CNY4.20/16.80/.84.
				rates := []float64{.3124070217, 1.2496280869, .0624814043}
				if input == 512001 {
					rates = []float64{.6248140434, 2.4992561738, .1249628087}
				}
				want := float64(uncached)*rates[0]/1e6 + 20*rates[1]/1e6 + 100*rates[2]/1e6
				u := out.result.Message.Usage
				ev.Check("actual Result applies source FX and exact project tier boundary", out.err == nil && u.Input == int64(uncached) && u.CacheRead == 100 && u.Output == 20 && math.Abs(u.Cost.Total-want) < 1e-12, "err=%v usage=%+v want=%g", out.err, u, want)
				ev.Record("price-source-status", regional.Candidates())
				ev.Record("interpretation", "Offline synthetic token counts with own dated standard CNY price/FX; project512000 policy, not a billing or capacity measurement.")
			})
		}
	}
}

func TestRegionalProviderCatalogSamplingRepair(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses} {
			t.Run(string(p)+"/"+string(api), func(t *testing.T) {
				ev := run.Case(t, "RP13-hooks-supported-repair-"+string(p)+"-"+string(api))
				cat := regionalMetadataCatalog()
				w := regionalMetadataWorld(t, &cat)
				b := regionalMetadataBinding(w, p, api)
				var opts ai.Options = ai.ChatOptions{MaxTokens: ai.Value(64), SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage("9")}}
				if api == ai.APIOpenAIResponses {
					opts = ai.ResponsesOptions{MaxTokens: ai.Value(64), SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage("9")}}
				}
				target := ai.Target{BindingID: b.BindingID, ModelID: regionalMetadataModel(p)}
				order := []string{}
				payloadIdentity, responseIdentity := false, false
				intermediateInvalid := false
				hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
					order = append(order, "headers")
					h.Set("X-Metadata-Repair", "safe")
					return nil
				}, OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
					order = append(order, "payload")
					payloadIdentity = payload.ProviderID == p && payload.API == api && payload.ModelID == target.ModelID
					intermediateInvalid = fmt.Sprint(payload.Body["temperature"]) == "9"
					payload.Body["temperature"] = .5
					return ai.KeepPayload(), nil
				}, OnResponse: func(_ context.Context, _ ai.CallScope, response ai.ResponseInfo) error {
					order = append(order, "response")
					responseIdentity = response.ProviderID == p && response.API == api && response.ModelID == target.ModelID
					return nil
				}}
				reply, raw := regionalHostScript(t, b, target.ModelID)
				ev.Fixture("synthetic-script.json", raw)
				w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
				result, err := w.client.WithHooks(hooks).Complete(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: t.Name() + "-repaired"}, target, ai.Request{}, opts)
				ev.Record("result", result)
				ev.Record("hook-order", order)
				ev.Record("requests", w.provider.Requests())
				captures := w.provider.Requests()
				ev.Check("supported repair preserves identity and each hook runs once in order", err == nil && payloadIdentity && responseIdentity && intermediateInvalid && strings.Join(order, ",") == "headers,payload,response" && result.Metadata.ProviderID == p && result.Metadata.API == api && len(captures) == 1, "err=%v order=%v", err, order)
				if len(captures) == 1 {
					fields := bodyFields(t, captures[0].Body)
					ev.Check("only repaired legal sampling reaches the chosen target", jsonEqual(fields["temperature"], []byte("0.5")) && captures[0].Header["X-Metadata-Repair"] == "safe", "body=%s", captures[0].Body)
				}
			})
		}
	}
}
