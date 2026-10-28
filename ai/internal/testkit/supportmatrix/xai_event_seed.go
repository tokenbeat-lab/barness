package supportmatrix

import (
	"encoding/json"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/tokenbeat-lab/barness/ai"
	"io"
	"net/http"
	"strings"
)

// A function's first native item can carry complete arguments before any delta.
// Public tool starts expose a shared view, so recover that seed from original SSE.
func xaiEventToolSeed(c xaiCall, call ai.ToolCall) string {
	ids, err := xaiOpaque(call.ID)
	if err != nil || len(ids) != 2 {
		return ""
	}
	d := ssestream.NewDecoder(&http.Response{Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(c.ex.ResponseBody))})
	defer d.Close()
	for d.Next() {
		var frame map[string]json.RawMessage
		if regionalDecode(d.Event().Data, &frame) != nil {
			continue
		}
		var items []map[string]json.RawMessage
		if item, ok := frame["item"]; ok {
			var v map[string]json.RawMessage
			if regionalDecode(item, &v) == nil {
				items = append(items, v)
			}
		}
		if response, ok := frame["response"]; ok {
			var r map[string]json.RawMessage
			if regionalDecode(response, &r) == nil {
				regionalDecode(r["output"], &items)
			}
		}
		for _, item := range items {
			var id, kind, args string
			json.Unmarshal(item["call_id"], &id)
			json.Unmarshal(item["type"], &kind)
			if kind == "function_call" && id == ids[0] {
				json.Unmarshal(item["arguments"], &args)
				return args
			}
		}
	}
	return ""
}
