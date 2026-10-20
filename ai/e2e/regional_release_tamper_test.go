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
	for _, mutation := range []string{"inventory", "source", "catalog", "endpoint", "model", "tool-pair", "response", "events", "origin", "extra"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "RP18-native-refusal-"+combo+"-"+mutation)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			caseID := supportmatrix.RegionalCaseID(combo, model+"/tool-result")
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
