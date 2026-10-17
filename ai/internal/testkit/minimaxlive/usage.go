package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
)

// This independent evidence calculation intentionally does not call adapter
// parsers: changing runtime counts cannot silently bless a changed report.
func verifyUsage(dir string, r Report, result Capture, raw []json.RawMessage) error {
	actual := result.Result.Message.Usage
	want := ai.Usage{}
	count := func(obj map[string]json.RawMessage, name string) (int64, bool) {
		var n int64
		v, ok := obj[name]
		if !ok || string(v) == "null" || json.Unmarshal(v, &n) != nil {
			return 0, false
		}
		return n, true
	}
	nested := func(obj map[string]json.RawMessage, name string) map[string]json.RawMessage {
		var out map[string]json.RawMessage
		json.Unmarshal(obj[name], &out)
		return out
	}
	hasUsage := false
	for _, u := range raw {
		var obj map[string]json.RawMessage
		if json.Unmarshal(u, &obj) != nil || obj == nil {
			continue
		}
		hasUsage = true
		switch r.Identity.Route {
		case "chat":
			prompt, _ := count(obj, "prompt_tokens")
			want.Output, _ = count(obj, "completion_tokens")
			details := nested(obj, "prompt_tokens_details")
			want.CacheRead, _ = count(details, "cached_tokens")
			want.CacheWrite, _ = count(details, "cache_write_tokens")
			want.Input = max(0, prompt-want.CacheRead-want.CacheWrite)
			want.TotalTokens = want.Input + want.Output + want.CacheRead + want.CacheWrite
			n, _ := count(nested(obj, "completion_tokens_details"), "reasoning_tokens")
			want.Reasoning = ai.Value(n)
		case "responses":
			input, _ := count(obj, "input_tokens")
			want.Output, _ = count(obj, "output_tokens")
			want.CacheRead, _ = count(nested(obj, "input_tokens_details"), "cached_tokens")
			want.CacheWrite, _ = count(nested(obj, "input_tokens_details"), "cache_write_tokens")
			want.Input = max(0, input-want.CacheRead-want.CacheWrite)
			want.TotalTokens, _ = count(obj, "total_tokens")
			n, _ := count(nested(obj, "output_tokens_details"), "reasoning_tokens")
			want.Reasoning = ai.Value(n)
		default:
			for name, target := range map[string]*int64{"input_tokens": &want.Input, "output_tokens": &want.Output, "cache_read_input_tokens": &want.CacheRead, "cache_creation_input_tokens": &want.CacheWrite} {
				if n, ok := count(obj, name); ok {
					*target = n
				}
			}
			if n, ok := count(nested(obj, "cache_creation"), "ephemeral_1h_input_tokens"); ok {
				want.CacheWrite1h = ai.Value(n)
			}
			if n, ok := count(nested(obj, "output_tokens_details"), "thinking_tokens"); ok {
				want.Reasoning = ai.Value(n)
			}
			want.TotalTokens = want.Input + want.Output + want.CacheRead + want.CacheWrite
		}
	}
	if hasUsage {
		rawCatalog, e := ReadFile(filepath.Join(dir, "minimax-catalog.json"), 2<<20)
		if e != nil {
			return e
		}
		var catalog ai.Catalog
		if json.Unmarshal(rawCatalog, &catalog) != nil {
			return fmt.Errorf("MiniMax captured catalog invalid")
		}
		hash, e := catalog.Hash()
		if e != nil || hash != r.CatalogHash || catalog.Version != r.CatalogVersion {
			return fmt.Errorf("MiniMax catalog hash differs")
		}
		model, ok := catalog.Lookup(ai.ProviderMiniMaxCN, r.Identity.API, r.Identity.Model)
		if !ok {
			return fmt.Errorf("MiniMax captured model absent")
		}
		rates := model.Cost.CostRates
		matched := int64(-1)
		for _, tier := range model.Cost.Tiers {
			if want.Input+want.CacheRead+want.CacheWrite > tier.InputTokensAbove && tier.InputTokensAbove > matched {
				rates = tier.CostRates
				matched = tier.InputTokensAbove
			}
		}
		want.Cost.Input = float64(float64(rates.Input/1e6) * float64(want.Input))
		want.Cost.Output = float64(float64(rates.Output/1e6) * float64(want.Output))
		want.Cost.CacheRead = float64(float64(rates.CacheRead/1e6) * float64(want.CacheRead))
		want.Cost.CacheWrite = float64(rates.CacheWrite*float64(want.CacheWrite)) / 1e6
		want.Cost.Total = want.Cost.Input + want.Cost.Output + want.Cost.CacheRead + want.Cost.CacheWrite
	}
	if !reflect.DeepEqual(want, actual) {
		return fmt.Errorf("MiniMax usage/cost differs from captured raw counters")
	}
	return nil
}
