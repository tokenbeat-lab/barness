package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxMessagesSupportedEfforts(t *testing.T) {
	for _, effort := range []ai.AnthropicEffort{ai.AnthropicEffortLow, ai.AnthropicEffortMedium, ai.AnthropicEffortHigh, ai.AnthropicEffortXHigh, ai.AnthropicEffortMax} {
		t.Run(string(effort), func(t *testing.T) {
			ev := run.Case(t, "MM04-effort-"+string(effort))
			w := newMiniMaxMessagesWorld(t, nil)
			enqueue(ev, w, miniMaxMessagesReply(t, "MiniMax-M3.1-Flash-Preview"))
			o := drain(w.client.Stream(ctxFor(t), textScope("mm04-effort"), miniMaxMessagesTarget("MiniMax-M3.1-Flash-Preview"), miniMaxMessagesRequest(), ai.AnthropicOptions{Effort: effort}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if ev.Check("M3.1 declared effort accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				f := bodyFields(t, w.provider.Requests()[0].Body)
				ev.Check("effort-only survives with no synthetic switch", jsonEqual(f["output_config"], []byte(`{"effort":"`+string(effort)+`"}`)) && len(f["thinking"]) == 0, "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}

func TestMiniMaxMessagesSupportedBoundaries(t *testing.T) {
	for _, c := range []struct {
		name   string
		opts   ai.AnthropicOptions
		simple *ai.SimpleOptions
		edit   func(*ai.Payload)
		budget string
	}{
		{name: "one-full", opts: ai.AnthropicOptions{MaxTokens: ai.Value(1)}, budget: "1"},
		{name: "one-simple", simple: &ai.SimpleOptions{MaxTokens: ai.Value(1), Reasoning: ai.ThinkingHigh}, budget: "1"},
		{name: "one-hook", edit: func(p *ai.Payload) { p.Body["max_tokens"] = 1 }, budget: "1"},
		{name: "max-full", opts: ai.AnthropicOptions{MaxTokens: ai.Value(131072)}, budget: "131072"},
		{name: "null-budget", opts: ai.AnthropicOptions{MaxTokens: ai.Null[int]()}, budget: "131072"},
		{name: "temperature-zero", opts: ai.AnthropicOptions{Temperature: ai.Value(0.0)}},
		{name: "temperature-max", opts: ai.AnthropicOptions{ThinkingEnabled: ai.Value(true), Temperature: ai.Value(2.0)}},
		{name: "temperature-null", opts: ai.AnthropicOptions{Temperature: ai.Null[float64]()}},
		{name: "top-p-zero", edit: func(p *ai.Payload) { p.Body["top_p"] = 0 }},
		{name: "top-p-max", edit: func(p *ai.Payload) { p.Body["top_p"] = 1 }},
		{name: "standard", edit: func(p *ai.Payload) { p.Body["service_tier"] = "standard" }},
		{name: "auto", opts: ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "auto"})}},
		{name: "none", opts: ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "none"})}},
		{name: "cache-short", opts: ai.AnthropicOptions{CacheRetention: ai.CacheRetentionShort}},
		{name: "cache-long", opts: ai.AnthropicOptions{CacheRetention: ai.CacheRetentionLong}},
		{name: "cache-none", opts: ai.AnthropicOptions{CacheRetention: ai.CacheRetentionNone}},
		{name: "metadata", opts: ai.AnthropicOptions{Metadata: map[string]json.RawMessage{"user_id": json.RawMessage(`"synthetic-user"`), "tenant": json.RawMessage(`"private-tenant"`)}}},
		{name: "sampling-ignored", simple: &ai.SimpleOptions{SamplingParams: map[string]json.RawMessage{"thinking": json.RawMessage(`{"type":"disabled"}`), "max_tokens": json.RawMessage(`0`)}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-boundary-"+c.name)
			w := newMiniMaxMessagesWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					m := &cat.Models[i]
					m.Compat.ForceAdaptiveThinking = true
					m.Compat.SupportsMidConvoEffort = true
					m.Compat.SupportsLongCacheRetention = ai.Value(true)
					m.SamplingParams = map[string]json.RawMessage{"thinking": json.RawMessage(`{"type":"disabled"}`), "max_tokens": json.RawMessage(`0`)}
					m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingHigh: m.SamplingParams}
				}
			})
			enqueue(ev, w, miniMaxMessagesReply(t, "MiniMax-M3"))
			req := miniMaxToolRequest()
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				if c.edit != nil {
					c.edit(p)
				}
				return ai.KeepPayload(), nil
			}}
			var o outcome
			if c.simple != nil {
				o = drain(w.client.WithHooks(hooks).StreamSimple(ctxFor(t), textScope("mm04-boundary"), miniMaxMessagesTarget("MiniMax-M3"), req, *c.simple))
			} else {
				o = drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm04-boundary"), miniMaxMessagesTarget("MiniMax-M3"), req, c.opts))
			}
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("supported boundary remains usable", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			r := w.provider.Requests()[0]
			f := bodyFields(t, r.Body)
			if c.budget != "" {
				ev.Check("exact output limit retained", string(f["max_tokens"]) == c.budget, "%s", r.Body)
			}
			if !c.opts.Temperature.IsZero() {
				ev.Check("temperature preserves native value or null", jsonEqual(f["temperature"], mustMarshal(t, c.opts.Temperature)), "%s", r.Body)
			}
			ev.Check("catalog cannot enable Claude extensions", !strings.Contains(string(r.Body), "cache_control") && !strings.Contains(string(r.Body), "eager_input_streaming") && !strings.Contains(string(r.Body), "block_binding") && r.Header["Anthropic-Beta"] == "" && o.result.Message.ProviderThinkingLevel == "", "%s", r.Body)
			ev.Check("tenant metadata is never inferred or forwarded", !strings.Contains(string(r.Body), tenantA.tenant) && !strings.Contains(string(r.Body), "private-tenant"), "%s", r.Body)
		})
	}
}

func TestMiniMaxMessagesModelNarrowsSampling(t *testing.T) {
	for _, entry := range []string{"named", "payload"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "MM04-model-temperature-"+entry)
			w := newMiniMaxMessagesWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					cat.Models[i].Compat.SupportsTemperature = ai.Value(false)
				}
			})
			opts := ai.AnthropicOptions{}
			hooks := ai.Hooks{}
			if entry == "named" {
				opts.Temperature = ai.Value(1.0)
			} else {
				hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					p.Body["temperature"] = 1.0
					return ai.KeepPayload(), nil
				}
			}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm04-model-temperature"), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), opts))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			phase := ai.PhaseRequest
			if entry == "named" {
				phase = ai.PhaseCapability
			}
			checkRejected(ev, o, ai.CodeInvalidRequest, phase)
			ev.Check("model can narrow supported sampling without a send", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

func TestMiniMaxMessagesRequiredBinaryIsNotDepth(t *testing.T) {
	ev := run.Case(t, "MM04-required-binary")
	w := newMiniMaxMessagesWorld(t, func(cat *ai.Catalog) {
		for i := range cat.Models {
			if cat.Models[i].ID == "MiniMax-M3" {
				m := &cat.Models[i]
				m.Compat.RequiresThinking = true
				m.ThinkingLevelMap.Off = ai.Null[string]()
				m.DefaultThinkingLevel = ai.ThinkingHigh
			}
		}
	})
	o := drain(w.client.Stream(ctxFor(t), textScope("mm04-required-binary"), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
	ev.Check("required high-only declaration stays a binary switch", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
}

func TestMiniMaxMessagesNarrowedDepth(t *testing.T) {
	ev := run.Case(t, "MM04-narrowed-depth")
	w := newMiniMaxMessagesWorld(t, func(cat *ai.Catalog) {
		for i := range cat.Models {
			if cat.Models[i].ID == "MiniMax-M3.1-Flash-Preview" {
				m := &cat.Models[i]
				m.DefaultThinkingLevel = ai.ThinkingHigh
				m.ThinkingLevelMap.Low = ai.Null[string]()
				m.ThinkingLevelMap.Medium = ai.Null[string]()
				m.ThinkingLevelMap.XHigh = ai.Null[string]()
				m.ThinkingLevelMap.Max = ai.Null[string]()
			}
		}
	})
	enqueue(ev, w, miniMaxMessagesReply(t, "MiniMax-M3.1-Flash-Preview"))
	o := drain(w.client.Stream(ctxFor(t), textScope("mm04-narrowed-depth"), miniMaxMessagesTarget("MiniMax-M3.1-Flash-Preview"), miniMaxMessagesRequest(), ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if ev.Check("narrowed depth model retains effort semantics", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		f := bodyFields(t, w.provider.Requests()[0].Body)
		ev.Check("explicit high effort retained", jsonEqual(f["output_config"], []byte(`{"effort":"high"}`)), "%s", w.provider.Requests()[0].Body)
	}
}
