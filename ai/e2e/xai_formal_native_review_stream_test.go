package e2e

import "testing"

func TestXAIFormalNativeReviewStream(t *testing.T) {
	// Filtering rewrites directly because the ordinary edit helper retains length.
	filter := func(dir string, omit func(map[string]any) bool) {
		p := dir + "/E17-xai-entry-stream/text-events.json"
		v := readXAIReviewArray(t, p)
		var kept []any
		for _, raw := range v {
			if !omit(raw.(map[string]any)) {
				kept = append(kept, raw)
			}
		}
		writeXAINativeJSON(t, p, kept)
	}
	runXAINativeReview(t, "stream", []xaiNativeReviewCase{
		{"final-only-source", true, nil},
		{"split-deltas", true, func(d string) {
			p := d + "/E17-xai-entry-stream/text-events.json"
			v := readXAIReviewArray(t, p)
			var next []any
			for _, raw := range v {
				e := raw.(map[string]any)
				if delta, ok := e["Delta"].(string); ok && len(delta) > 1 {
					a, b := map[string]any{}, map[string]any{}
					for k, v := range e {
						a[k], b[k] = v, v
					}
					a["Delta"], b["Delta"] = delta[:len(delta)/2], delta[len(delta)/2:]
					next = append(next, a, b)
				} else {
					next = append(next, e)
				}
			}
			writeXAINativeJSON(t, p, next)
		}},
		{"block-start-order", false, func(d string) {
			p := d + "/E17-xai-entry-stream/text-events.json"
			v := readXAIReviewArray(t, p)
			var textStart any
			var kept []any
			for _, raw := range v {
				if raw.(map[string]any)["type"] == "text_start" {
					textStart = raw
				} else {
					kept = append(kept, raw)
				}
			}
			if textStart == nil {
				t.Fatal("synthetic text start missing")
			}
			next := append([]any{kept[0], textStart}, kept[1:]...)
			writeXAINativeJSON(t, p, next)
		}},

		{"blocks-missing", false, func(d string) {
			filter(d, func(e map[string]any) bool { return e["type"] != "start" && e["type"] != "done" })
		}},
		{"delta-dropped", false, func(d string) { filter(d, func(e map[string]any) bool { return e["type"] == "text_delta" }) }},
		{"delta-changed", false, func(d string) {
			editXAINativeArray(t, d, "E17-xai-entry-stream/text-events.json", func(v []any) {
				for _, raw := range v {
					e := raw.(map[string]any)
					if e["type"] == "text_delta" {
						e["Delta"] = "SYNTHETIC_wrong"
					}
				}
			})
		}},
		{"snapshot-changed", false, func(d string) {
			editXAINativeArray(t, d, "E17-xai-entry-stream/text-events.json", func(v []any) {
				v[0].(map[string]any)["Partial"].(map[string]any)["content"].([]any)[1].(map[string]any)["text"] = "SYNTHETIC_wrong"
			})
		}},
	})
}
