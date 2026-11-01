package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

const catalogContractPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jBMsAAAAASUVORK5CYII="

func catalogChatParameters(t *testing.T, ev *evidence.Case, w *world, cat ai.Catalog, i catalogContractIdentity, m ai.Model, req ai.Request) {
	enqueue(ev, w, catalogChatReply(t, m, 8, false))
	result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-parameters"), catalogContractTarget(i), req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
	ev.Record("result", result)
	ev.Check("bounded options accepted", err == nil, "%v", err)
	if err == nil {
		catalogChatResult(ev, cat, i, m, result, 8)
		raw := catalogChatWire(ev, w, i)
		var body map[string]json.RawMessage
		mustUnmarshal(t, raw, &body)
		if i.API == ai.APIOpenAIResponses || i.API == ai.APIOpenAICompletions {
			defaults := map[string]json.RawMessage{}
			for key, value := range m.SamplingParams {
				defaults[key] = value
			}
			for key, value := range m.SamplingParamsByThinkingLevel[m.DefaultThinkingLevel] {
				defaults[key] = value
			}
			for key, value := range defaults {
				ev.Check("frozen default sampling "+key, jsonEqual(body[key], value), "got=%s want=%s", body[key], value)
			}
		} else {
			ev.Check("no ignored native sampling declarations", len(m.SamplingParams) == 0 && len(m.SamplingParamsByThinkingLevel) == 0, "code_gap: native sampling declarations are not implemented")
		}
	}
	before := len(w.provider.Requests())
	_, err = w.client.CompleteSimple(ctxFor(t), textScope("catalog-invalid-level"), catalogContractTarget(i), req, ai.SimpleOptions{Reasoning: "invalid"})
	ev.Check("invalid reasoning refused before request", err != nil && len(w.provider.Requests()) == before, "%v", err)
	_, err = w.client.Complete(ctxFor(t), textScope("catalog-wrong-operation"), catalogContractTarget(i), req, ai.TypeSafeOptions{})
	ev.Check("foreign protocol options refused", err != nil && len(w.provider.Requests()) == before, "%v", err)
}

func catalogChatReasoning(t *testing.T, ev *evidence.Case, w *world, cat ai.Catalog, i catalogContractIdentity, m ai.Model, req ai.Request) {
	entries := []struct {
		level ai.ThinkingLevel
		value ai.Nullable[string]
	}{{ai.ThinkingMinimal, m.ThinkingLevelMap.Minimal}, {ai.ThinkingLow, m.ThinkingLevelMap.Low}, {ai.ThinkingMedium, m.ThinkingLevelMap.Medium}, {ai.ThinkingHigh, m.ThinkingLevelMap.High}, {ai.ThinkingXHigh, m.ThinkingLevelMap.XHigh}, {ai.ThinkingMax, m.ThinkingLevelMap.Max}}
	exercised := 0
	for _, entry := range entries {
		if entry.value.IsNull() || (entry.level == ai.ThinkingXHigh || entry.level == ai.ThinkingMax) && entry.value.IsZero() {
			continue
		}
		exercised++
		enqueue(ev, w, catalogChatReply(t, m, 8, false))
		result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-reasoning-"+string(entry.level)), catalogContractTarget(i), req, ai.SimpleOptions{Reasoning: entry.level, MaxTokens: ai.Value(64)})
		ev.Record("result-"+string(entry.level), result)
		ev.Check("declared effort accepted "+string(entry.level), err == nil, "%v", err)
		if err != nil {
			continue
		}
		catalogChatResult(ev, cat, i, m, result, 8)
		raw := catalogChatWire(ev, w, i)
		var body map[string]json.RawMessage
		mustUnmarshal(t, raw, &body)
		wire := string(entry.level)
		if mapped, ok := entry.value.Get(); ok {
			wire = mapped
		}
		switch i.API {
		case ai.APIOpenAIResponses:
			var reasoning map[string]json.RawMessage
			_ = json.Unmarshal(body["reasoning"], &reasoning)
			var effort string
			_ = json.Unmarshal(reasoning["effort"], &effort)
			ev.Check("exact declared Responses effort "+wire, effort == wire, "%s", raw)
		case ai.APIOpenAICompletions:
			if i.Provider == ai.ProviderMiniMax || i.Provider == ai.ProviderMiniMaxCN {
				ev.Check("MiniMax native thinking retained", strings.Contains(string(raw), `"reasoning_split"`), "%s", raw)
			} else if i.Provider == ai.ProviderZhipu || i.Provider == ai.ProviderZAI {
				ev.Check("GLM native thinking retained", strings.Contains(string(raw), `"thinking"`), "%s", raw)
			} else if i.Provider == ai.ProviderDeepSeek {
				ev.Check("DeepSeek native thinking retained", strings.Contains(string(raw), `"thinking"`), "%s", raw)
			} else {
				var effort string
				_ = json.Unmarshal(body["reasoning_effort"], &effort)
				ev.Check("exact declared Chat effort "+wire, effort == wire, "%s", raw)
			}
		case ai.APIAnthropicMessages:
			ev.Check("Messages thinking enabled or required native effort", strings.Contains(string(raw), `"thinking"`) || (i.Provider == ai.ProviderMoonshot || i.Provider == ai.ProviderMoonshotCN) && strings.Contains(string(raw), `"output_config"`), "%s", raw)
			if m.Compat.ForceAdaptiveThinking || m.Compat.SupportsMidConvoEffort || i.Provider == ai.ProviderMoonshot || i.Provider == ai.ProviderMoonshotCN {
				ev.Check("Messages effort wired", strings.Contains(string(raw), `"effort":`), "%s", raw)
			}
		case ai.APIGoogleGenerativeAI:
			ev.Check("Google thinking config wired", strings.Contains(string(raw), `"thinkingConfig"`), "%s", raw)
		}
	}
	if exercised == 0 {
		if m.Compat.MessagesThinkingBudgetTokens > 0 && i.API == ai.APIAnthropicMessages {
			enqueue(ev, w, catalogChatReply(t, m, 8, false))
			result, err := w.client.Complete(ctxFor(t), textScope("catalog-native-budget"), catalogContractTarget(i), req, ai.AnthropicOptions{ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: m.Compat.MessagesThinkingBudgetTokens, MaxTokens: ai.Value(4096)})
			ev.Record("native-budget-result", result)
			ev.Check("declared native budget accepted", err == nil, "%v", err)
			if err == nil {
				catalogChatResult(ev, cat, i, m, result, 8)
				raw := catalogChatWire(ev, w, i)
				ev.Check("exact declared native budget", strings.Contains(string(raw), `"budget_tokens":`+strconv.Itoa(m.Compat.MessagesThinkingBudgetTokens)), "%s", raw)
			}
		} else {
			enqueue(ev, w, catalogChatReply(t, m, 8, false))
			result, err := w.client.CompleteSimple(ctxFor(t), textScope("catalog-reasoning-default"), catalogContractTarget(i), req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
			ev.Record("default-only-result", result)
			ev.Check("native reasoning default remains callable", err == nil, "%v", err)
			if err == nil {
				catalogChatResult(ev, cat, i, m, result, 8)
				raw := catalogChatWire(ev, w, i)
				ev.Check("no unsupported canonical effort synthesized", !strings.Contains(string(raw), `"effort":`), "%s", raw)
			}
			before := len(w.provider.Requests())
			_, err = w.client.CompleteSimple(ctxFor(t), textScope("catalog-unsupported-canonical"), catalogContractTarget(i), req, ai.SimpleOptions{Reasoning: ai.ThinkingHigh, MaxTokens: ai.Value(64)})
			ev.Check("unsupported canonical control refused", err != nil && len(w.provider.Requests()) == before, "%v", err)
		}
	}
	ev.Record("requests", w.provider.Requests())
}
