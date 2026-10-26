package e2e

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// Blend one deidentified retained response into an explicitly synthetic native
// bundle. Its current Client replay supplies normalized values; no original
// paid bundle, source pin or claim is rewritten or promoted.
func useXAIReviewRetainedEntry(t *testing.T, dir string) {
	t.Helper()
	_, captures, _ := loadXAIRealCaptures(t)
	var capture xaiRealCapture
	for _, c := range captures {
		if c.Scenario == "entry-complete" {
			capture = c
			break
		}
	}
	if capture.ResponseBody == "" {
		t.Fatal("retained public fixture missing")
	}
	w := newXAIWorld(t, nil)
	w.provider.Enqueue(capture.reply())
	result, err := w.client.Complete(ctxFor(t), textScope("synthetic-retained-native"), xaiTarget(), capture.initial(t), capture.options(t))
	if err != nil {
		t.Fatal(err)
	}
	prefix := "E17-xai-entry-complete/text"
	record := readXAINativeObject(t, filepath.Join(dir, prefix+"-result.json"))
	oldMessage := record["message"].(map[string]any)
	oldUsage := oldMessage["usage"].(map[string]any)
	oldMax := record["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["costEstimate"].(map[string]any)["maximum"].(map[string]any)["total"].(float64)
	raw, _ := json.Marshal(result.Message)
	var message map[string]any
	mustUnmarshal(t, raw, &message)
	message["timestamp"], message["durationMs"] = oldMessage["timestamp"], oldMessage["durationMs"]
	record["message"] = message
	attempt := record["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)
	raw, _ = json.Marshal(result.Metadata.Attempts[0].CostEstimate)
	var cost map[string]any
	mustUnmarshal(t, raw, &cost)
	attempt["usage"], attempt["costEstimate"] = message["usage"], cost
	writeXAINativeJSON(t, filepath.Join(dir, prefix+"-result.json"), record)
	editXAINativeArray(t, dir, prefix+"-observer.json", func(v []any) {
		v[2].(map[string]any)["attempt"] = attempt
		last := v[3].(map[string]any)
		last["usage"] = message["usage"]
		last["call"] = record["metadata"]
	})
	editXAINativeArray(t, dir, prefix+"-http.json", func(v []any) {
		e := v[0].(map[string]any)
		e["requestBody"] = string(capture.Request)
		e["responseBody"] = capture.ResponseBody
	})
	usage := message["usage"].(map[string]any)
	editXAINativeJSON(t, dir, "xai-formal-report.json", func(v map[string]any) {
		b := v["budget"].(map[string]any)
		b["reportedTokens"] = b["reportedTokens"].(float64) + usage["totalTokens"].(float64) - oldUsage["totalTokens"].(float64)
		b["observedMaximumUSD"] = b["observedMaximumUSD"].(float64) + cost["maximum"].(map[string]any)["total"].(float64) - oldMax
	})
	editXAINativeJSON(t, dir, "live-report.json", func(v map[string]any) {
		b := v["budget"].(map[string]any)
		b["inputTokensUsed"] = b["inputTokensUsed"].(float64) + usage["input"].(float64) + usage["cacheRead"].(float64) - oldUsage["input"].(float64) - oldUsage["cacheRead"].(float64)
		b["outputTokensUsed"] = b["outputTokensUsed"].(float64) + usage["output"].(float64) - oldUsage["output"].(float64)
	})
}
