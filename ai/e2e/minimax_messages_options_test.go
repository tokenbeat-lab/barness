package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxMessagesThinkingWire(t *testing.T) {
	for _, c := range []struct {
		name, model      string
		opts             ai.Options
		simple           *ai.SimpleOptions
		thinking, effort string
	}{
		{name: "m3-nil", model: "MiniMax-M3"},
		{name: "m3-full", model: "MiniMax-M3", opts: ai.AnthropicOptions{}},
		{name: "m3-simple", model: "MiniMax-M3", simple: &ai.SimpleOptions{}},
		{name: "m31-nil", model: "MiniMax-M3.1-Flash-Preview"},
		{name: "m31-full", model: "MiniMax-M3.1-Flash-Preview", opts: ai.AnthropicOptions{}},
		{name: "m31-simple", model: "MiniMax-M3.1-Flash-Preview", simple: &ai.SimpleOptions{}},
		{name: "m3-on", model: "MiniMax-M3", opts: ai.AnthropicOptions{ThinkingEnabled: ai.Value(true), Temperature: ai.Value(1.2)}, thinking: "adaptive"},
		{name: "m3-off", model: "MiniMax-M3", opts: ai.AnthropicOptions{ThinkingEnabled: ai.Value(false)}, thinking: "disabled"},
		{name: "m3-simple-clamp", model: "MiniMax-M3", simple: &ai.SimpleOptions{Reasoning: ai.ThinkingLow}, thinking: "adaptive"},
		{name: "m31-effort-only", model: "MiniMax-M3.1-Flash-Preview", opts: ai.AnthropicOptions{Effort: ai.AnthropicEffortLow}, effort: "low"},
		{name: "m31-on-effort", model: "MiniMax-M3.1-Flash-Preview", opts: ai.AnthropicOptions{ThinkingEnabled: ai.Value(true), Effort: ai.AnthropicEffortMax, Temperature: ai.Value(1.2)}, thinking: "adaptive", effort: "max"},
		{name: "m31-simple-clamp", model: "MiniMax-M3.1-Flash-Preview", simple: &ai.SimpleOptions{Reasoning: ai.ThinkingMinimal}, thinking: "adaptive", effort: "low"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-thinking-"+c.name)
			w := newMiniMaxMessagesWorld(t, nil)
			enqueue(ev, w, miniMaxMessagesReply(t, c.model))
			var o outcome
			if c.simple != nil {
				o = drain(w.client.StreamSimple(ctxFor(t), textScope("mm04-thinking"), miniMaxMessagesTarget(c.model), miniMaxMessagesRequest(), *c.simple))
			} else {
				o = drain(w.client.Stream(ctxFor(t), textScope("mm04-thinking"), miniMaxMessagesTarget(c.model), miniMaxMessagesRequest(), c.opts))
			}
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("successful own protocol request", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			r := w.provider.Requests()[0]
			f := bodyFields(t, r.Body)
			wantThinking := []byte(nil)
			if c.thinking != "" {
				wantThinking = []byte(`{"type":"` + c.thinking + `"}`)
			}
			ev.Check("adaptive selection preserves omitted service default", len(f["thinking"]) == 0 && c.thinking == "" || jsonEqual(f["thinking"], wantThinking), "%s", r.Body)
			wantEffort := []byte(nil)
			if c.effort != "" {
				wantEffort = []byte(`{"effort":"` + c.effort + `"}`)
			}
			ev.Check("only supported explicit depth sent", len(f["output_config"]) == 0 && c.effort == "" || jsonEqual(f["output_config"], wantEffort), "%s", r.Body)
			if opts, ok := c.opts.(ai.AnthropicOptions); ok && !opts.Temperature.IsZero() {
				ev.Check("adaptive preserves explicit temperature", string(f["temperature"]) == "1.2", "%s", r.Body)
			}
			ev.Check("core Messages path and one authentication", r.Path == "/anthropic/v1/messages" && r.Header["X-Api-Key"] != "" && r.Header["Authorization"] == "", "%+v", r)
			ev.Check("native vendor extensions absent", r.Header["Anthropic-Beta"] == "" && r.Header["Anthropic-Dangerous-Direct-Browser-Access"] == "" && !json.Valid(f["cache_control"]), "%+v", r)
		})
	}
}
