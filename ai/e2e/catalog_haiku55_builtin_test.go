package e2e

import (
	"math"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Failure cases: missing/incorrect identity, changed prior definitions, unsafe
// thinking/temperature wiring, lost modalities/history/authorization, or a
// 100k threshold that ignores cached tokens or misprices one-hour writes.
// The generic public catalog contract owns the shared protocol scenarios;
// these literal checks independently pin this offering and its price boundary.
func TestHaiku55BuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "CATALOG-HAIKU55-builtin")
	cat := ai.BuiltinCatalog()
	i := catalogContractIdentity{Operation: ai.OperationChat, Provider: ai.ProviderAnthropic, API: ai.APIAnthropicMessages, ID: "claude-haiku-5-5"}
	m, ok := cat.Lookup(i.Provider, i.API, i.ID)
	if !ev.Check("Haiku 5.5 is discoverable", ok, "identity=%+v", i) {
		return
	}
	ev.Check("official identity limits and supported modalities", m.Name == "Claude Haiku 5.5" && m.Reasoning && m.ContextWindow == 1000000 && m.MaxTokens == 128000 && slices.Equal(m.Input, []ai.Modality{ai.ModalityText, ai.ModalityImage}), "model=%+v", m)
	temperature, explicit := m.Compat.SupportsTemperature.Get()
	ev.Check("adaptive managed effort and service default are declared", m.Compat.ForceAdaptiveThinking && m.Compat.SupportsMidConvoEffort && explicit && !temperature && m.DefaultThinkingLevel == ai.ThinkingMedium && m.ThinkingLevelMap.Off.IsNull() && m.ThinkingLevelMap.Minimal.IsNull(), "model=%+v", m)
	for level, value := range map[string]ai.Nullable[string]{"low": m.ThinkingLevelMap.Low, "medium": m.ThinkingLevelMap.Medium, "high": m.ThinkingLevelMap.High, "xhigh": m.ThinkingLevelMap.XHigh, "max": m.ThinkingLevelMap.Max} {
		got, present := value.Get()
		ev.Check("official effort "+level, present && got == level, "got=%s", got)
	}
	ev.Check("complete base and strictly-above-100k rates", m.Cost.CostRates == (ai.CostRates{Input: .1, Output: .5, CacheRead: .01, CacheWrite: .125}) && len(m.Cost.Tiers) == 1 && m.Cost.Tiers[0] == (ai.CostTier{InputTokensAbove: 100000, CostRates: ai.CostRates{Input: .5, Output: 2.5, CacheRead: .05, CacheWrite: .625}}), "cost=%+v", m.Cost)
	raw, err := os.ReadFile("testdata/catalog/pre-haiku55-builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	var prior struct {
		Hash    string
		Catalog ai.Catalog
	}
	mustUnmarshal(t, raw, &prior)
	ev.Check("prior snapshot is valid", catalogHash(t, prior.Catalog) == prior.Hash, "prior snapshot invalid")
	without := cat
	without.Version = prior.Catalog.Version
	without.Models = slices.DeleteFunc(slices.Clone(cat.Models), func(model ai.Model) bool {
		return model.Provider == i.Provider && model.API == i.API && model.ID == i.ID
	})
	ev.Check("original retained definitions preserve their approved order", reflect.DeepEqual(historicalCatalogInCurrentOrder(catalogWithHistoricalNames(without, prior.Catalog), prior.Catalog), prior.Catalog), "prior catalog changed")
	for _, scene := range []string{"load-query", "basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization", "tier-0-below", "tier-0-equal", "tier-0-above"} {
		t.Run(scene, func(t *testing.T) { catalogChatContract(t, run.Case(t, "CATALOG-HAIKU55-"+scene), cat, i, scene) })
	}
	t.Run("managed-wire", func(t *testing.T) {
		e := run.Case(t, "CATALOG-HAIKU55-managed-wire")
		w, _ := newCatalogContractWorld(t, cat, i)
		enqueue(e, w, catalogChatReply(t, m, 8, false))
		result, err := w.client.CompleteSimple(ctxFor(t), textScope("haiku55-managed"), catalogContractTarget(i), ai.Request{Messages: []ai.Message{ai.UserText("Controlled input.")}}, ai.SimpleOptions{Reasoning: ai.ThinkingMedium, MaxTokens: ai.Value(64), Temperature: ai.Value(.9)})
		e.Check("medium effort completes", err == nil && result.Message.ProviderThinkingLevel == "medium", "err=%v result=%+v", err, result)
		body := string(catalogChatWire(e, w, i))
		e.Check("adaptive binding controls and per-turn effort survive", strings.Contains(body, `"type":"adaptive"`) && strings.Contains(body, `"prefix_mismatch_behavior":"drop_block"`) && strings.Contains(body, `"effort":"medium"`) && !strings.Contains(body, `"temperature"`) && !strings.Contains(body, `"budget_tokens"`), "%s", body)
		e.Record("requests", w.provider.Requests())
	})
	for _, tc := range []struct {
		name                     string
		input, read, short, long int64
		want                     ai.UsageCost
	}{
		{"equal-cache-read", 90000, 10000, 0, 0, ai.UsageCost{Input: .009, Output: .000002, CacheRead: .0001}},
		{"above-cache-read", 90000, 10001, 0, 0, ai.UsageCost{Input: .045, Output: .00001, CacheRead: .00050005}},
		{"above-short-write", 99999, 0, 2, 0, ai.UsageCost{Input: .0499995, Output: .00001, CacheWrite: .00000125}},
		{"equal-mixed-ttl", 90000, 9998, 1, 1, ai.UsageCost{Input: .009, Output: .000002, CacheRead: .00009998, CacheWrite: .000000325}},
		{"above-mixed-ttl", 90000, 9999, 1, 1, ai.UsageCost{Input: .045, Output: .00001, CacheRead: .00049995, CacheWrite: .000001625}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := run.Case(t, "CATALOG-HAIKU55-"+tc.name)
			w, _ := newCatalogContractWorld(t, cat, i)
			reply := catalogChatReply(t, m, tc.input, false)
			lines := strings.Split(reply.Script, "\n")
			for n, line := range lines {
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				var event map[string]any
				mustUnmarshal(t, []byte(strings.TrimPrefix(line, "data: ")), &event)
				patch := func(usage map[string]any) {
					usage["input_tokens"] = tc.input
					usage["output_tokens"] = 4
					usage["cache_read_input_tokens"] = tc.read
					usage["cache_creation_input_tokens"] = tc.short + tc.long
					usage["cache_creation"] = map[string]any{"ephemeral_5m_input_tokens": tc.short, "ephemeral_1h_input_tokens": tc.long}
				}
				if usage, ok := event["usage"].(map[string]any); ok {
					patch(usage)
				}
				if message, ok := event["message"].(map[string]any); ok {
					patch(message["usage"].(map[string]any))
				}
				lines[n] = "data: " + string(mustMarshal(t, event))
			}
			reply.Script = strings.Join(lines, "\n")
			reply.Chunks = [][]byte{[]byte(reply.Script)}
			enqueue(e, w, reply)
			result, err := w.client.CompleteSimple(ctxFor(t), textScope("haiku55-"+tc.name), catalogContractTarget(i), ai.Request{Messages: []ai.Message{ai.UserText("Controlled cache usage.")}}, ai.SimpleOptions{Reasoning: ai.ThinkingMedium, MaxTokens: ai.Value(64), CacheRetention: ai.CacheRetentionLong})
			e.Record("result", result)
			e.Record("requests", w.provider.Requests())
			if !e.Check("controlled mixed usage completes", err == nil, "%v", err) {
				return
			}
			u := result.Message.Usage
			long, known := u.CacheWrite1h.Get()
			e.Check("distinct ordinary cache and TTL counters survive", u.Input == tc.input && u.Output == 4 && u.CacheRead == tc.read && u.CacheWrite == tc.short+tc.long && known && long == tc.long, "%+v", u)
			want := tc.want
			want.Total = want.Input + want.Output + want.CacheRead + want.CacheWrite
			e.Check("whole request uses the independently priced prompt tier", math.Abs(u.Cost.Input-want.Input) < 1e-12 && math.Abs(u.Cost.Output-want.Output) < 1e-12 && math.Abs(u.Cost.CacheRead-want.CacheRead) < 1e-12 && math.Abs(u.Cost.CacheWrite-want.CacheWrite) < 1e-12 && math.Abs(u.Cost.Total-want.Total) < 1e-12, "got=%+v want=%+v", u.Cost, want)
		})
	}
}
