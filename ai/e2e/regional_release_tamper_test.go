package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Mutations operate on explicitly synthetic controlled copies. Original
// captures and their grant receipts are retained untouched beside the run.
func regionalCaptureRefusals(t *testing.T, original, combo string) {
	model := supportmatrix.RegionalModels(combo)[0]
	for _, mutation := range []string{"inventory", "source", "catalog", "endpoint", "model", "tool-pair", "response", "events", "origin", "extra", "raw-usage-missing", "raw-usage-zero", "raw-usage-partial", "raw-usage-count", "raw-text", "raw-thinking", "raw-tool", "raw-terminal"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "RP18-native-refusal-"+combo+"-"+mutation)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			caseID := supportmatrix.RegionalCaseID(combo, model+"/tool-result")
			if strings.HasPrefix(mutation, "raw-") {
				kind := "full-stream"
				if mutation == "raw-thinking" || mutation == "raw-tool" {
					kind = "tool-call"
				}
				caseID = supportmatrix.RegionalCaseID(combo, model+"/"+kind)
				if mutation == "raw-thinking" {
					caseID = supportmatrix.RegionalNativeSourceCase(combo, model)
				}
			}
			path := caseID + "/exchanges.json"
			if mutation == "inventory" {
				path = caseID + "/result.json"
			}
			if mutation == "source" {
				path = "regional-execution-source.json"
			}
			if mutation == "catalog" {
				path = "regional-catalog.json"
			}
			if mutation == "model" {
				path = caseID + "/result.json"
			}
			if mutation == "events" {
				path = caseID + "/events.json"
			}
			if mutation == "origin" {
				path = "regional-authorization-receipt.json"
			}
			if mutation == "extra" {
				path = "unowned-case.json"
			}
			raw, err := os.ReadFile(filepath.Join(dir, path))
			if err != nil && mutation != "extra" {
				t.Fatal(err)
			}
			var value any
			if mutation != "extra" && json.Unmarshal(raw, &value) != nil {
				t.Fatal("invalid seed")
			}
			if strings.HasPrefix(mutation, "raw-") {
				exchange := value.([]any)[0].(map[string]any)
				exchange["responseBody"] = regionalMutateRaw(t, exchange["responseBody"].(string), mutation)
			}
			switch mutation {
			case "inventory":
				value.(map[string]any)["metadata"].(map[string]any)["modelId"] = "synthetic-sibling"
			case "source":
				value.(map[string]any)["ai/cache_key.go"] = strings.Repeat("0", 64)
			case "catalog":
				value.(map[string]any)["version"] = "synthetic-edited"
			case "model":
				value.(map[string]any)["metadata"].(map[string]any)["modelId"] = "synthetic-sibling"
			case "endpoint":
				value.([]any)[0].(map[string]any)["url"] = "https://synthetic.invalid/sibling"
			case "response":
				value.([]any)[0].(map[string]any)["responseBody"] = "synthetic unparseable stream"
			case "events":
				value.([]any)[len(value.([]any))-1].(map[string]any)["type"] = "start"
			case "origin":
				value.(map[string]any)["mode"] = "live"
			case "tool-pair":
				x := value.([]any)[0].(map[string]any)
				var wire any
				if json.Unmarshal([]byte(x["requestBody"].(string)), &wire) != nil {
					t.Fatal("invalid wire")
				}
				regionalChangeToolPair(wire)
				body, _ := json.Marshal(wire)
				x["requestBody"] = string(body)
			case "extra":
				value = map[string]string{"synthetic": "extra"}
			}
			raw = mustMarshal(t, value)
			if err := os.WriteFile(filepath.Join(dir, path), raw, 0600); err != nil {
				t.Fatal(err)
			}
			if mutation != "inventory" {
				if strings.Contains(path, "/") {
					var manifest map[string]any
					mraw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
					if json.Unmarshal(mraw, &manifest) != nil {
						t.Fatal("invalid manifest")
					}
					for _, v := range manifest["cases"].([]any) {
						c := v.(map[string]any)
						if c["id"] == caseID {
							c["artifacts"].(map[string]any)[filepath.Base(path)] = pioracle.SHA256(raw)
						}
					}
					if err := os.WriteFile(filepath.Join(dir, "manifest.json"), mustMarshal(t, manifest), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := supportmatrix.SealRegionalBundle(dir); err != nil {
					t.Fatal(err)
				}
			}
			_, err = supportmatrix.VerifyRegionalControlledCapture(dir)
			ev.Record("synthetic-mutation", map[string]any{"name": mutation, "path": path, "refused": err != nil, "error": err, "originalRetained": true})
			ev.Check("independent native consumer rejects mutation", err != nil, "accepted %s", mutation)
		})
	}
}

func regionalChangeToolPair(v any) {
	switch v := v.(type) {
	case []any:
		for _, x := range v {
			regionalChangeToolPair(x)
		}
	case map[string]any:
		if v["role"] == "tool" {
			v["tool_call_id"] = "synthetic-wrong-call"
		}
		if v["type"] == "function_call_output" {
			v["call_id"] = "synthetic-wrong-call"
		}
		if v["type"] == "tool_result" {
			v["tool_use_id"] = "synthetic-wrong-call"
		}
		for _, x := range v {
			regionalChangeToolPair(x)
		}
	}
}

func regionalMutateRaw(t *testing.T, body, mutation string) string {
	t.Helper()
	lines := strings.Split(body, "\n")
	changes := 0
	for i, line := range lines {
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok || !json.Valid([]byte(data)) {
			continue
		}
		var frame any
		if json.Unmarshal([]byte(data), &frame) != nil {
			t.Fatal("invalid mutation frame")
		}
		var walk func(any)
		walk = func(v any) {
			switch v := v.(type) {
			case []any:
				for _, x := range v {
					walk(x)
				}
			case map[string]any:
				if usage, ok := v["usage"].(map[string]any); ok {
					switch mutation {
					case "raw-usage-missing":
						delete(v, "usage")
						changes++
					case "raw-usage-zero", "raw-usage-count":
						var counts func(any)
						counts = func(x any) {
							if m, ok := x.(map[string]any); ok {
								for k, n := range m {
									if _, ok := n.(float64); ok {
										if mutation == "raw-usage-zero" {
											m[k] = float64(0)
										} else {
											m[k] = n.(float64) + 1
										}
										changes++
									} else {
										counts(n)
									}
								}
							}
						}
						counts(usage)
					case "raw-usage-partial":
						for _, key := range []string{"prompt_tokens", "input_tokens", "output_tokens", "completion_tokens"} {
							if _, ok := usage[key]; ok {
								delete(usage, key)
								changes++
								break
							}
						}
					}
				}
				for key, x := range v {
					text, _ := x.(string)
					if mutation == "raw-text" && text != "" && (key == "content" || key == "text" || key == "delta" && v["type"] == "response.output_text.delta") {
						v[key] = "synthetic changed answer"
						changes++
					}
					if mutation == "raw-thinking" && text != "" && (key == "reasoning_content" || key == "thinking" || key == "signature" || key == "text" && strings.Contains(text, "reasoning")) {
						v[key] = "synthetic changed native"
						changes++
					}
					if mutation == "raw-tool" && key == "name" && text == "inspect_regional" {
						v[key] = "synthetic_other_tool"
						changes++
					}
					if mutation == "raw-terminal" && (key == "finish_reason" || key == "stop_reason") && text != "" {
						v[key] = "error"
						changes++
					}
					if mutation == "raw-terminal" && key == "type" && text == "response.completed" {
						v[key] = "response.failed"
						changes++
					}
					walk(v[key])
				}
			}
		}
		walk(frame)
		raw, _ := json.Marshal(frame)
		lines[i] = "data: " + string(raw)
	}
	if changes == 0 {
		t.Fatalf("mutation %s did not affect raw capture", mutation)
	}
	return strings.Join(lines, "\n")
}
