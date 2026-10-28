package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// This current assessment is separate from original wire and Usage. Automatic
// cache writes have no independent fee for this one M3 contract; missing tokens
// are still unreported and UsagePartial is never normalized to complete.
type ThinkingBilling struct {
	IndependentCacheWriteFee string            `json:"independentCacheWriteFee"`
	CacheCreationCounter     string            `json:"cacheCreationCounter"`
	UsageReporting           ai.UsageReporting `json:"usageReporting"`
	BasisSourceSHA256        string            `json:"basisSourceSha256"`
}

type ThinkingInspection struct {
	Report
	Billing ThinkingBilling `json:"billing"`
}

// AssessThinkingBilling is not a qualification or an execution permission. The
// native verifier additionally checks sealed originals, authority and replay.
func AssessThinkingBilling(dir string, r Report) (ThinkingBilling, error) {
	var out ThinkingBilling
	want, err := Select("messages", M3, "standard")
	if err != nil || r.Identity != want || r.Kind != ThinkingKind || len(r.Calls) != 1 || r.Calls[0].ID != "thinking-on" {
		return out, fmt.Errorf("MiniMax automatic-write assessment requires fixed international Messages/M3/standard single")
	}
	if _, err := Sources(); err != nil {
		return out, err
	}
	raw, err := ReadFile(filepath.Join(dir, r.Calls[0].CaseID, "request.json"), 2048)
	if err != nil {
		return out, err
	}
	var wire any
	if json.Unmarshal(raw, &wire) != nil {
		return out, fmt.Errorf("MiniMax automatic-cache wire invalid")
	}
	if thinkingCacheControl(wire) {
		return out, fmt.Errorf("MiniMax automatic-write assessment forbids cache_control")
	}
	seen := map[string]bool{}
	for _, raw := range r.Calls[0].UsagePresence {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return out, fmt.Errorf("MiniMax single native usage invalid")
		}
		for k, v := range fields {
			if strings.HasPrefix(k, "cache_creation") && k != "cache_creation_input_tokens" {
				return out, fmt.Errorf("MiniMax explicit creation/TTL outside automatic-cache contract")
			}
			switch k {
			case "input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens":
				var n int64
				if string(v) == "null" || json.Unmarshal(v, &n) != nil || n < 0 {
					return out, fmt.Errorf("MiniMax single invalid billable counter: %s", k)
				}
				if k == "cache_creation_input_tokens" && n != 0 {
					return out, fmt.Errorf("MiniMax positive cache creation requires separate review")
				}
				seen[k] = true
			}
		}
	}
	for _, k := range []string{"input_tokens", "output_tokens", "cache_read_input_tokens"} {
		if !seen[k] {
			return out, fmt.Errorf("MiniMax single billable counter unreported: %s", k)
		}
	}
	var captured Capture
	raw, err = ReadFile(filepath.Join(dir, r.Calls[0].CaseID, "result.json"), 2<<20)
	if err != nil || json.Unmarshal(raw, &captured) != nil || len(captured.Result.Metadata.Attempts) != 1 {
		return out, fmt.Errorf("MiniMax single usage assessment unavailable")
	}
	counter := "UNREPORTED"
	if seen["cache_creation_input_tokens"] {
		counter = "REPORTED_ZERO"
	}
	out = ThinkingBilling{IndependentCacheWriteFee: "NOT_APPLICABLE", CacheCreationCounter: counter, UsageReporting: captured.Result.Metadata.Attempts[0].UsageReporting, BasisSourceSHA256: SourcePin}
	return out, nil
}

func thinkingCacheControl(v any) bool {
	switch x := v.(type) {
	case map[string]any:
		for k, v := range x {
			if k == "cache_control" || thinkingCacheControl(v) {
				return true
			}
		}
	case []any:
		for _, v := range x {
			if thinkingCacheControl(v) {
				return true
			}
		}
	}
	return false
}
