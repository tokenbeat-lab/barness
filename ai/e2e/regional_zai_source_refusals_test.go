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

// These are resealed synthetic controlled copies. Their lineage statements,
// domain request prefix and native wire must agree independently.
func regionalNativeSourceRefusals(t *testing.T, original, combo string) {
	modes := []string{"case", "model", "tenant", "request", "wire"}
	if combo == "zai-chat" {
		modes = append(modes, "wire-visible")
	}
	for _, mode := range modes {
		t.Run("native-source-"+mode, func(t *testing.T) {
			ev := run.Case(t, "RP18-"+combo+"-native-source-"+mode)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			read := func(name string) any {
				raw, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				var v any
				mustUnmarshal(t, raw, &v)
				return v
			}
			write := func(name string, value any) {
				raw := mustMarshal(t, value)
				if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
					t.Fatal(err)
				}
				caseID := filepath.Dir(name)
				if caseID == "." {
					return
				}
				manifest := read("manifest.json").(map[string]any)
				for _, v := range manifest["cases"].([]any) {
					c := v.(map[string]any)
					if c["id"] == caseID {
						c["artifacts"].(map[string]any)[filepath.Base(name)] = pioracle.SHA256(raw)
					}
				}
				if err := os.WriteFile(filepath.Join(dir, "manifest.json"), mustMarshal(t, manifest), 0600); err != nil {
					t.Fatal(err)
				}
			}
			source := supportmatrix.RegionalCaseID(combo, "glm-5.3/full-stream")
			next := supportmatrix.RegionalCaseID(combo, "glm-5.3/native-history")
			switch mode {
			case "case", "model":
				value := read(supportmatrix.RegionalReportFile).(map[string]any)
				wrong := supportmatrix.RegionalCaseID(combo, "glm-5.3/tool-call")
				if mode == "model" {
					wrong = supportmatrix.RegionalCaseID(combo, "glm-5.2/full-stream")
				}
				value["regional"].(map[string]any)["nativeSourceCases"].(map[string]any)["glm-5.3"] = wrong
				write(supportmatrix.RegionalReportFile, value)
			case "tenant":
				value := read(source + "/result.json").(map[string]any)
				value["metadata"].(map[string]any)["tenantId"] = "other-synthetic-tenant"
				write(source+"/result.json", value)
				observed := read(source + "/observations.json").([]any)
				for _, v := range observed {
					if call, ok := v.(map[string]any)["call"].(map[string]any); ok {
						call["tenantId"] = "other-synthetic-tenant"
					}
				}
				write(source+"/observations.json", observed)
			case "request":
				value := read(next + "/request.json").(map[string]any)
				value["Messages"].([]any)[1].(map[string]any)["responseId"] = "another-synthetic-response"
				write(next+"/request.json", value)
			default:
				value := read(next + "/exchanges.json").([]any)
				x := value[0].(map[string]any)
				var wire map[string]any
				mustUnmarshal(t, []byte(x["requestBody"].(string)), &wire)
				var assistant map[string]any
				if combo == "zai-chat" {
					assistant = wire["messages"].([]any)[1].(map[string]any)
				} else {
					for _, raw := range wire["input"].([]any) {
						item := raw.(map[string]any)
						if item["type"] == "reasoning" {
							assistant = item
							break
						}
					}
					if assistant == nil {
						t.Fatal("native wire reasoning item absent")
					}
				}
				changeLetter := func(value string) string {
					raw := []byte(value)
					for i, b := range raw {
						if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' {
							raw[i] ^= 32
							return string(raw)
						}
					}
					t.Fatal("synthetic wire mutation needs one ASCII letter")
					return ""
				}
				if combo == "zai-responses" {
					content := assistant["content"].(map[string]any)
					content["text"] = changeLetter(content["text"].(string))
				} else if mode == "wire" {
					assistant["reasoning_content"] = changeLetter(assistant["reasoning_content"].(string))
				} else {
					assistant["content"] = changeLetter(assistant["content"].(string))
				}
				raw, err := json.Marshal(wire)
				if err != nil {
					t.Fatal(err)
				}
				if len(raw) != len(x["requestBody"].(string)) {
					t.Fatal("wire mutation must preserve the independent input reservation")
				}
				x["requestBody"] = string(raw)
				write(next+"/exchanges.json", value)
			}
			if err := supportmatrix.SealRegionalBundle(dir); err != nil {
				t.Fatal(err)
			}
			_, err := supportmatrix.VerifyRegionalControlledCapture(dir)
			want := map[string]string{"case": "native source case selection differs", "model": "native source case selection differs", "tenant": "native source scope differs", "request": "does not extend fixed source", "wire": "was not replayed in a native wire container", "wire-visible": "was not replayed in a native wire container"}[mode]
			ev.Check("consumer refuses inconsistent fixed native source at its boundary", err != nil && strings.Contains(err.Error(), want), "mutation=%s err=%v", mode, err)
			ev.Record("mutation", mode)
		})
	}
}
