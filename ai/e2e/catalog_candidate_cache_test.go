package e2e

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func catalogChatCache(t *testing.T, ev *evidence.Case, w *world, cat ai.Catalog, i catalogContractIdentity, m ai.Model, req ai.Request) {
	write := int64(1)
	forbiddenWrite := i.Provider == ai.ProviderXAI || ((i.Provider == ai.ProviderMiniMax || i.Provider == ai.ProviderMiniMaxCN || i.Provider == ai.ProviderZAI) && m.Cost.CacheWrite == 0)
	if forbiddenWrite || i.API == ai.APIGoogleGenerativeAI {
		write = 0
	}
	prompt := int64(10) + write
	if i.API == ai.APIAnthropicMessages {
		prompt = 8
	}
	if i.API == ai.APIGoogleGenerativeAI {
		prompt = 10
	}
	reply := catalogChatReply(t, m, prompt, false)
	lines := strings.Split(reply.Script, "\n")
	for n, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimPrefix(line, "data: ")
		if raw == "[DONE]" {
			continue
		}
		var event map[string]any
		if json.Unmarshal([]byte(raw), &event) != nil {
			t.Fatal("invalid scripted event")
		}
		switch i.API {
		case ai.APIOpenAICompletions:
			if usage, ok := event["usage"].(map[string]any); ok {
				usage["prompt_tokens_details"] = map[string]any{"cached_tokens": 2, "cache_write_tokens": write}
			}
		case ai.APIOpenAIResponses:
			if response, ok := event["response"].(map[string]any); ok {
				usage := response["usage"].(map[string]any)
				usage["input_tokens_details"] = map[string]any{"cached_tokens": 2, "cache_write_tokens": write}
				if i.Provider == ai.ProviderMoonshot || i.Provider == ai.ProviderMoonshotCN {
					response["prompt_cache_options"] = map[string]any{"mode": "implicit", "ttl": "5m"}
				}
			}
		case ai.APIAnthropicMessages:
			patch := func(usage map[string]any) {
				usage["cache_read_input_tokens"] = 2
				usage["cache_creation_input_tokens"] = write
				usage["cache_creation"] = map[string]any{"ephemeral_5m_input_tokens": write, "ephemeral_1h_input_tokens": 0}
			}
			if usage, ok := event["usage"].(map[string]any); ok {
				patch(usage)
			}
			if message, ok := event["message"].(map[string]any); ok {
				patch(message["usage"].(map[string]any))
			}
		case ai.APIGoogleGenerativeAI:
			if usage, ok := event["usageMetadata"].(map[string]any); ok {
				usage["cachedContentTokenCount"] = 2
			}
		}
		lines[n] = "data: " + string(mustMarshal(t, event))
	}
	reply.Script = strings.Join(lines, "\n")
	reply.Chunks = [][]byte{[]byte(reply.Script)}
	enqueue(ev, w, reply)
	result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-cache"), catalogContractTarget(i), req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
	ev.Record("result", result)
	ev.Record("requests", w.provider.Requests())
	ev.Check("controlled cache usage accepted", err == nil, "%v", err)
	if err != nil {
		return
	}
	catalogContractMetadata(ev, cat, i, result.Metadata)

	u := result.Message.Usage
	ev.Check("explicit cache counters and distinct ordinary input", u.Input == 8 && u.Output == 4 && u.CacheRead == 2 && u.CacheWrite == write && u.TotalTokens == 14+write, "%+v", u)
	rates := m.Cost.CostRates
	highest := int64(-1)
	for _, tier := range m.Cost.Tiers {
		if 10+write > tier.InputTokensAbove && tier.InputTokensAbove >= highest {
			rates = tier.CostRates
			highest = tier.InputTokensAbove
		}
	}
	writeRate := rates.CacheWrite
	kimiChat := (i.Provider == ai.ProviderMoonshot || i.Provider == ai.ProviderMoonshotCN) && i.API == ai.APIOpenAICompletions
	if kimiChat {
		writeRate = min(writeRate, 2*rates.Input)
	}
	expected := ai.UsageCost{Input: 8 * rates.Input / 1e6, Output: 4 * rates.Output / 1e6, CacheRead: 2 * rates.CacheRead / 1e6, CacheWrite: float64(write) * writeRate / 1e6}
	expected.Total = expected.Input + expected.Output + expected.CacheRead + expected.CacheWrite
	ev.Check("independent nonzero cache component prices", math.Abs(u.Cost.Input-expected.Input) < 1e-12 && math.Abs(u.Cost.Output-expected.Output) < 1e-12 && math.Abs(u.Cost.CacheRead-expected.CacheRead) < 1e-12 && math.Abs(u.Cost.CacheWrite-expected.CacheWrite) < 1e-12 && math.Abs(u.Cost.Total-expected.Total) < 1e-12, "got=%+v expected=%+v", u.Cost, expected)
	ev.Check("explicit cache report complete", result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", result.Metadata.Attempts[0])
	if kimiChat {
		lo, lok := result.Metadata.Attempts[0].CostEstimate.Minimum.Get()
		hi, hok := result.Metadata.Attempts[0].CostEstimate.Maximum.Get()
		wantHigh := expected.Total - float64(write)*writeRate/1e6 + float64(write)*max(rates.CacheWrite, 2*rates.Input)/1e6
		ev.Check("unreported Chat TTL stays an independently bounded estimate", lok && hok && math.Abs(lo.Total-expected.Total) < 1e-12 && math.Abs(hi.Total-wantHigh) < 1e-12, "%+v", result.Metadata.Attempts[0].CostEstimate)
	}
	if forbiddenWrite {
		denied := catalogChatReply(t, m, 11, false)
		denied.Script = strings.NewReplacer(`"cache_write_tokens":0`, `"cache_write_tokens":1`, `"cache_creation_input_tokens":0`, `"cache_creation_input_tokens":1`).Replace(denied.Script)
		denied.Chunks = [][]byte{[]byte(denied.Script)}
		enqueue(ev, w, denied)
		failed, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-unpriced-write"), catalogContractTarget(i), req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
		ev.Record("unpriced-write-result", failed)
		ev.Check("unsupported positive write remains a protocol refusal", err != nil, "%v", err)
		enqueue(ev, w, catalogChatReply(t, m, 8, false))
		after, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-after-runtime-refusal"), catalogContractTarget(i), req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
		ev.Check("runtime refusal never removes loaded model", err == nil, "%v", err)
		if err == nil {
			catalogChatResult(ev, cat, i, m, after, 8)
		}
	}
}
