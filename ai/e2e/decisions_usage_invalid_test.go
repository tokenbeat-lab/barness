package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsUsageInvalid(t *testing.T) {
	cases := []decisionsUsageCase{}
	add := func(name, fields string) {
		cases = append(cases, decisionsUsageCase{name: name, fields: fields, invalid: true, reporting: ai.UsageUnreported})
	}
	for _, value := range []struct{ name, raw string }{{"array", "[]"}, {"string", `"bad"`}, {"number", "1"}, {"bool", "false"}} {
		add("usage-"+value.name, `"usage":`+value.raw)
		for _, details := range []string{"input_tokens_details", "output_tokens_details"} {
			add(details+"-"+value.name, `"usage":{"`+details+`":`+value.raw+`}`)
		}
	}
	add("duplicate-usage", decisionsFullUsage+","+decisionsUsage)
	add("triple-usage", decisionsFullUsage+","+decisionsUsage+","+decisionsFullUsage)
	for _, field := range []struct{ key, value string }{{"input_tokens", "296"}, {"output_tokens", "20"}, {"total_tokens", "316"}, {"cached_tokens", "100"}, {"cache_write_tokens", "50"}, {"reasoning_tokens", "7"}} {
		old := `"` + field.key + `":` + field.value
		for _, value := range []struct{ name, raw string }{{"negative", "-1"}, {"fraction", "1.5"}, {"decimal", "1.0"}, {"exponent", "1e2"}, {"overflow", "9223372036854775808"}, {"nonfinite", "1e999"}, {"string", `"1"`}, {"bool", "false"}, {"object", "{}"}, {"array", "[]"}} {
			add(field.key+"-"+value.name, strings.Replace(decisionsFullUsage, old, `"`+field.key+`":`+value.raw, 1))
		}
		add(field.key+"-duplicate", strings.Replace(decisionsFullUsage, old, old+`,`+old, 1))
		add(field.key+"-alias", strings.Replace(decisionsFullUsage, old, `"`+strings.ToUpper(field.key)+`":`+field.value, 1))
		add(field.key+"-alias-pair", strings.Replace(decisionsFullUsage, old, old+`,"`+strings.ToUpper(field.key)+`":0`, 1))
	}
	for _, field := range []struct{ key, value string }{{"input_tokens_details", `{"cached_tokens":100,"cache_write_tokens":50}`}, {"output_tokens_details", `{"reasoning_tokens":7}`}} {
		old := `"` + field.key + `":` + field.value
		add(field.key+"-duplicate", strings.Replace(decisionsFullUsage, old, old+`,`+old, 1))
		add(field.key+"-alias", strings.Replace(decisionsFullUsage, old, `"`+strings.ToUpper(field.key)+`":`+field.value, 1))
	}
	add("cache-read-exceeds-input", strings.Replace(decisionsFullUsage, `"cached_tokens":100`, `"cached_tokens":297`, 1))
	add("cache-write-exceeds-input", strings.Replace(decisionsFullUsage, `"cache_write_tokens":50`, `"cache_write_tokens":297`, 1))
	add("cache-sum-exceeds-input", strings.Replace(decisionsFullUsage, `"cached_tokens":100`, `"cached_tokens":247`, 1))
	add("cache-sum-overflow", `"usage":{"input_tokens_details":{"cached_tokens":9223372036854775807,"cache_write_tokens":1}}`)
	add("input-output-sum-overflow", `"usage":{"input_tokens":9223372036854775807,"output_tokens":1}`)
	add("reported-total-sum-overflow", `"usage":{"input_tokens":9223372036854775807,"output_tokens":1,"total_tokens":9223372036854775807}`)
	add("total-mismatch", strings.Replace(decisionsFullUsage, `"total_tokens":316`, `"total_tokens":317`, 1))
	add("reasoning-exceeds-output", strings.Replace(decisionsFullUsage, `"reasoning_tokens":7`, `"reasoning_tokens":21`, 1))
	add("duplicate-vendor-detail", strings.TrimSuffix(decisionsFullUsage, "}")+`,"vendor_metadata":{"version":1,"version":2}}`)
	checkDecisionsUsage(t, cases)
}

func TestDecisionsUsageInputPricing(t *testing.T) {
	cases := []decisionsUsageCase{}
	for _, boundary := range []struct {
		input, total string
		cost         float64
	}{{"271999", "272019", .0271999}, {"272000", "272020", .0272}, {"272001", "272021", .0544002}} {
		for _, cache := range []struct{ name, read, write string }{{"no-cache", "0", "0"}, {"cached", "100000", "10000"}, {"all-cached", boundary.input, "0"}} {
			fields := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(decisionsFullUsage, `"input_tokens":296`, `"input_tokens":`+boundary.input), `"total_tokens":316`, `"total_tokens":`+boundary.total), `"cached_tokens":100`, `"cached_tokens":`+cache.read), `"cache_write_tokens":50`, `"cache_write_tokens":`+cache.write)
			var want ai.Usage
			mustUnmarshal(t, []byte(`{"input":`+boundary.input+`,"output":20,"totalTokens":`+boundary.total+`,"reasoning":7}`), &want)
			want.Cost = ai.UsageCost{Input: boundary.cost, Total: boundary.cost}
			cases = append(cases, decisionsUsageCase{name: "tier-" + boundary.input + "-" + cache.name, fields: fields, want: want, reporting: ai.UsageComplete})
		}
	}
	// Explicit total is authoritative when another top-level component is
	// unknown. The checked known sum also accepts the int64 limit exactly.
	cases = append(cases,
		decisionsUsageCase{name: "max-output-sum", fields: `"usage":{"output_tokens":9223372036854775807}`, want: ai.Usage{Output: 9223372036854775807, TotalTokens: 9223372036854775807}, reporting: ai.UsagePartial},
		decisionsUsageCase{name: "max-cache-count", fields: `"usage":{"input_tokens_details":{"cached_tokens":9223372036854775807,"cache_write_tokens":0}}`, reporting: ai.UsagePartial},
		decisionsUsageCase{name: "exact-large-total", fields: `"usage":{"total_tokens":9007199254740993}`, want: ai.Usage{TotalTokens: 9007199254740993}, reporting: ai.UsagePartial},
	)
	checkDecisionsUsage(t, cases)
}
