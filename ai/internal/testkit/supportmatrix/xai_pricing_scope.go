package supportmatrix

import "encoding/json"

// The frozen catalog prices default-tier model tokens and Host functions.
// Invoice ticks and context annotations are metadata, not extra token buckets.
func verifyXAIPricingScope(terminal, usage, inputs, outputs map[string]json.RawMessage) error {
	if tier, ok := terminal["service_tier"]; ok && string(tier) != "null" && xaiWireString(tier) != "default" {
		return refuse("xAI service tier is outside resolved default pricing")
	}
	for _, detail := range []struct {
		fields map[string]json.RawMessage
		known  string
	}{{inputs, "cached_tokens"}, {outputs, "reasoning_tokens"}} {
		for key, raw := range detail.fields {
			if key == detail.known || string(raw) == "null" {
				continue
			}
			n, err := xaiCount(raw)
			if err != nil || n != 0 {
				return refuse("xAI extra usage category has no resolved catalog price")
			}
		}
	}
	for _, key := range []string{"cache_write_tokens", "num_sources_used", "num_server_side_tools_used"} {
		if raw, ok := usage[key]; ok && string(raw) != "null" {
			n, err := xaiCount(raw)
			if err != nil || n != 0 {
				return refuse("xAI cache write or hosted tool quantity is not integrated")
			}
		}
	}
	if raw, ok := usage["server_side_tool_usage"]; ok && string(raw) != "null" {
		var v any
		if regionalDecode(raw, &v) != nil {
			return refuse("xAI hosted tool usage malformed")
		}
		empty := false
		switch v := v.(type) {
		case map[string]any:
			empty = len(v) == 0
		case []any:
			empty = len(v) == 0
		}
		if !empty {
			return refuse("xAI hosted tool usage has no resolved Host-function price")
		}
	}
	return nil
}
