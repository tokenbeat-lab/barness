package finalbudget

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func estimate(req *http.Request, raw []byte) (reservation, error) {
	var body map[string]any
	if json.Unmarshal(raw, &body) != nil {
		return reservation{}, fmt.Errorf("final USD requires native JSON body")
	}
	provider := map[string]ai.ProviderID{"api.openai.com": ai.ProviderOpenAI, "api.anthropic.com": ai.ProviderAnthropic,
		"generativelanguage.googleapis.com": ai.ProviderGoogle, "api.deepseek.com": ai.ProviderDeepSeek,
		"api.typesafe.ai": ai.ProviderTypeSafe, "open.bigmodel.cn": ai.ProviderZhipu, "api.z.ai": ai.ProviderZAI,
		"api.minimax.io": ai.ProviderMiniMax, "api.minimax.cn": ai.ProviderMiniMaxCN}[req.URL.Hostname()]
	if provider == "" || req.URL.Scheme != "https" {
		return reservation{}, fmt.Errorf("final USD requires exact official provider service")
	}
	model, _ := body["model"].(string)
	if model == "" && provider == ai.ProviderGoogle {
		_, tail, ok := strings.Cut(req.URL.Path, "/models/")
		if ok {
			model, _, _ = strings.Cut(tail, ":")
		}
	}
	cat := ai.BuiltinCatalog()
	if provider == ai.ProviderZAI || provider == ai.ProviderMiniMaxCN {
		candidates, err := regional.CandidateCatalog()
		if err != nil {
			return reservation{}, fmt.Errorf("final USD regional price source unavailable")
		}
		cat.Models = append(cat.Models, candidates.Models...)
	}
	output := int64(0)
	for _, name := range []string{"max_tokens", "max_output_tokens", "max_completion_tokens"} {
		output = max(output, positive(body[name]))
	}
	if config, ok := body["generationConfig"].(map[string]any); ok {
		output = max(output, positive(config["maxOutputTokens"]))
	}
	estimate := reservation{Model: model, InputReserve: int64(len(raw) + 1024), OutputReserve: output}
	found := false
	// Model IDs can be shared across operations with different price contracts.
	// Select the native operation's directory before accumulating highest tiers.
	if strings.Contains(req.URL.Path, "/decisions") || provider == ai.ProviderTypeSafe {
		estimate.OutputReserve = 0
		for _, m := range cat.ClassifierModels {
			if m.Provider == provider && m.ID == model {
				found = true
				rates(&estimate, m.Cost, hasLongCache(body))
			}
		}
	} else if strings.Contains(req.URL.Path, "/images/") || req.URL.Path == "/v1beta/interactions" {
		for _, m := range cat.ImageModels {
			if m.Provider == provider && m.ID == model {
				found = true
				inputText, _ := m.Pricing.InputText.Get()
				inputImage, _ := m.Pricing.InputImage.Get()
				outputText, _ := m.Pricing.OutputText.Get()
				outputImage, _ := m.Pricing.OutputImage.Get()
				estimate.InputRate = max(inputText, inputImage)
				estimate.OutputRate = max(outputText, outputImage)
				if provider == ai.ProviderGoogle {
					if m.ID != "gemini-nano-banana-2.1" {
						return reservation{}, fmt.Errorf("final USD image source does not cover this model")
					}
					// Official model: total output32768. Keep all modalities
					// at the highest sourced rate rather than guessing split.
					estimate.OutputReserve = 32768
				} else {
					if m.ID != "gpt-image-2.5-sunburst-2026-09-08" {
						return reservation{}, fmt.Errorf("final USD image calculator does not cover this model")
					}
					// Official GptImageTokenCalculator.react.COKX_Fgn.js,
					// SHA256 3bd869c65b0ec7291a7f6900d4efa1e92c54e76aad2de866e3d918a5161fd18b:
					// ceil(96²*(2000000+1024²)/4000000)=7024. Existing
					// low/medium use16/24; max96 reserves above both.
					estimate.OutputReserve = 7024
				}
			}
		}
		// New-round engineering reserve: each inline reference/mask adds65536
		// input units even when a highly compressible PNG has few wire bytes.
		// This is a conservative project allowance, not a tokenization claim.
	} else {
		for _, m := range cat.Models {
			if m.Provider == provider && m.ID == model {
				found = true
				rates(&estimate, m.Cost, hasLongCache(body))
			}
		}
	}
	estimate.InputReserve += int64(imageOccurrences(body)) * 65536
	if !found || estimate.InputRate <= 0 || estimate.OutputRate < 0 || estimate.OutputRate > 0 && estimate.OutputReserve <= 0 {
		return reservation{}, fmt.Errorf("final USD exact model/output price estimate unavailable")
	}
	// Dollars/M tokens is numerically microdollars/token. Ceiling makes the
	// reservation conservative across floating point and sub-microdollar calls.
	value := float64(estimate.InputReserve)*estimate.InputRate + float64(estimate.OutputReserve)*estimate.OutputRate
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 || value > 30000000 {
		return reservation{}, fmt.Errorf("final USD estimate invalid or exceeds whole round")
	}
	estimate.Microusd = int64(math.Ceil(value))
	return estimate, nil
}

func rates(out *reservation, cost ai.ModelCost, long bool) {
	update := func(r ai.CostRates) {
		out.InputRate = max(out.InputRate, r.Input, r.CacheRead, r.CacheWrite)
		if long {
			out.InputRate = max(out.InputRate, 2*r.Input)
		}
		out.OutputRate = max(out.OutputRate, r.Output)
	}
	update(cost.CostRates)
	for _, tier := range cost.Tiers {
		update(tier.CostRates)
	}
}

func imageOccurrences(value any) int {
	total := 0
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "data:image/") {
			total++
		}
	case []any:
		for _, item := range v {
			total += imageOccurrences(item)
		}
	case map[string]any:
		for _, key := range []string{"mime_type", "media_type", "mimeType"} {
			if mime, _ := v[key].(string); strings.HasPrefix(mime, "image/") {
				for _, field := range []string{"data", "base64", "data_base64"} {
					if data, _ := v[field].(string); data != "" && !strings.HasPrefix(data, "data:image/") {
						total++
						break
					}
				}
				break
			}
		}
		for _, item := range v {
			total += imageOccurrences(item)
		}
	}
	return total
}

// The SDK's existing one-hour cache write estimate is twice input cost,
// independently of the short-write CostRates field. Inspect final cache
// markers so only an actually encoded long request selects that premium.
func hasLongCache(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		if control, ok := v["cache_control"].(map[string]any); ok && control["ttl"] == "1h" {
			return true
		}
		for _, item := range v {
			if hasLongCache(item) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if hasLongCache(item) {
				return true
			}
		}
	}
	return false
}
