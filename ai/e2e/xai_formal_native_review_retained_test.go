package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"os"
	"path/filepath"
	"testing"
)

// Blend one deidentified retained response into an explicitly synthetic native
// bundle. Its current Client replay supplies normalized values; no original
// paid bundle, source pin or claim is rewritten or promoted.
func useXAIReviewRetainedEntry(t *testing.T, dir string) {
	t.Helper()
	useXAIReviewRetainedScenario(t, dir, "entry-complete")
}
func useXAIReviewRetainedScenario(t *testing.T, dir, scenario string) {
	t.Helper()
	_, captures, _ := loadXAIRealCaptures(t)
	count := 0
	for _, capture := range captures {
		if capture.Scenario == scenario {
			replaceXAIReviewRetainedCall(t, dir, capture)
			count++
		}
	}
	if count == 0 {
		t.Fatal("retained public scenario missing")
	}
}
func replaceXAIReviewRetainedCall(t *testing.T, dir string, capture xaiRealCapture) {
	t.Helper()
	prefix := "E17-xai-" + capture.Scenario + "/" + capture.Label
	w := newXAIWorld(t, nil)
	w.provider.Enqueue(capture.reply())
	var request ai.Request
	if capture.Label == "results" || capture.Label == "continued" {
		request = ai.Request{Messages: []ai.Message{ai.UserText("Synthetic isolated response replay.")}, Tools: capture.body(t).Tools}
	} else {
		request = capture.initial(t)
	}
	var replay outcome
	if _, err := os.Stat(filepath.Join(dir, prefix+"-events.json")); err == nil {
		replay = drain(w.client.Stream(ctxFor(t), textScope("synthetic-retained-native"), xaiTarget(), request, capture.options(t)))
	} else {
		replay.result, replay.err = w.client.Complete(ctxFor(t), textScope("synthetic-retained-native"), xaiTarget(), request, capture.options(t))
	}
	if replay.err != nil {
		t.Fatal(replay.err)
	}
	result := replay.result
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
	if len(replay.events) > 0 {
		var events []map[string]any
		for _, event := range replay.events {
			raw, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			var e map[string]any
			mustUnmarshal(t, raw, &e)
			e["type"] = string(event.Type())
			for _, key := range []string{"Partial", "Message"} {
				if m, ok := e[key].(map[string]any); ok {
					m["timestamp"], m["durationMs"] = oldMessage["timestamp"], oldMessage["durationMs"]
				}
			}
			events = append(events, e)
		}
		writeXAINativeJSON(t, filepath.Join(dir, prefix+"-events.json"), events)
	}
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
