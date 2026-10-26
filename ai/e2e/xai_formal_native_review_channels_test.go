package e2e

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// Obtain every normalized block/event by actually replaying the changed raw
// stream through the public Client. No claimed Thinking text is hand-written.
func xaiV6DualChannelReplay(t *testing.T, dir string, empty bool) {
	t.Helper()
	prefix := "E17-xai-entry-stream/text"
	editXAINativeTerminalAt(t, dir, prefix+"-http.json", func(r map[string]any) {
		reason := r["output"].([]any)[0].(map[string]any)
		reason["summary"] = []any{map[string]any{"type": "summary_text", "text": "Summary."}}
		reason["content"] = []any{map[string]any{"type": "reasoning_text", "text": "Detail."}}
	})
	prependXAIReviewFramesAt(t, dir, prefix, func(r map[string]any) []map[string]any {
		id := r["output"].([]any)[0].(map[string]any)["id"]
		summary := "Summary."
		if empty {
			summary = ""
		}
		return []map[string]any{
			{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"type": "reasoning", "id": id, "summary": []any{}}},
			{"type": "response.reasoning_summary_part.added", "output_index": 0, "item_id": id, "summary_index": 0, "part": map[string]any{"type": "summary_text", "text": summary}},
			{"type": "response.reasoning_text.delta", "output_index": 0, "item_id": id, "content_index": 0, "delta": "Detail."},
		}
	})
	exchanges := readXAIReviewArray(t, filepath.Join(dir, prefix+"-http.json"))
	ex := exchanges[0].(map[string]any)
	capture := xaiRealCapture{Scenario: "entry-stream", Label: "text", Request: json.RawMessage(ex["requestBody"].(string)), ResponseBody: ex["responseBody"].(string)}
	replaceXAIReviewRetainedCall(t, dir, capture)
}

func TestXAIFormalNativeReviewPrimaryChannel(t *testing.T) {
	fixture := newXAIFormalNativeFixture(t)
	for _, sc := range []struct {
		name, want string
		empty      bool
	}{
		{"empty-summary-then-detail", "Detail.", true},
		{"nonempty-summary-then-detail", "Summary.", false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E17-v6-channel-"+sc.name)
			dir := fixture.copy(t, sc.name)
			xaiV6DualChannelReplay(t, dir, sc.empty)
			result := readXAINativeObject(t, filepath.Join(dir, "E17-xai-entry-stream/text-result.json"))
			ev.Record("public-client-replayed-result", result)
			blocks := result["message"].(map[string]any)["content"].([]any)
			ev.Check("actual public Client selects first valid reasoning channel", blocks[0].(map[string]any)["thinking"] == sc.want, "blocks=%v", blocks)
			resealXAINativeManifest(t, dir)
			code, raw := fixture.consume(t, ev, dir)
			var got struct {
				Pass  bool
				Error string
			}
			mustUnmarshal(t, raw, &got)
			ev.Check("native CLI accepts actual dual-channel Client evidence", code == 0 && got.Pass, "exit=%d error=%s", code, got.Error)
		})
	}
}
