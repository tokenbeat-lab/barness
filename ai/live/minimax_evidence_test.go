//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// The auxiliary import boundary is failure-inventoried before implementation.
// Re-sealing deliberately tests semantic proof rather than a checksum alone.
func TestMiniMaxEvidenceContract(t *testing.T) {
	base := t.TempDir()
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestMiniMaxCandidate$", "-test.timeout=40s")
	cmd.Env = []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=global-test", "BARNESS_MINIMAX_CONTROLLED=success-chat", "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "stage.jsonl"), "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("controlled source failed: %v %s", e, output)
	}
	reports, _ := filepath.Glob(filepath.Join(base, "*", minimaxlive.ReportFile))
	if len(reports) != 1 {
		t.Fatal("missing source")
	}
	dir := filepath.Dir(reports[0])
	if _, e = minimaxlive.Verify(dir, false); e != nil {
		t.Fatal(e)
	}
	original, _ := minimaxlive.Inventory(dir)
	saved := map[string][]byte{}
	for name := range original {
		saved[name], _ = os.ReadFile(filepath.Join(dir, name))
	}
	for _, name := range []string{"cn-identity", "cn-provider", "cn-source", "cn-prices", "candidate-mode-spoof", "formal-mode-spoof", "formal-controlled", "invented-case", "omitted-attempt", "altered-usage", "coordinated-usage", "altered-request", "altered-events", "source-addition", "source-deletion", "ledger-omission", "unknown-field", "duplicate-field", "noncanonical-field", "extra-file", "unsafe-file"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "MM07-validator-"+name)
			for path, raw := range saved {
				if e := os.WriteFile(filepath.Join(dir, path), raw, 0600); e != nil {
					t.Fatal(e)
				}
			}
			mutate := func(path string, fn func(map[string]any)) {
				var v map[string]any
				json.Unmarshal(saved[path], &v)
				fn(v)
				raw, _ := json.Marshal(v)
				os.WriteFile(filepath.Join(dir, path), raw, 0600)
			}
			switch name {

			case "cn-identity":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { v["identity"].(map[string]any)["endpoint"] = "https://api.minimax.cn/v1" })
			case "cn-provider":
				mutate("MM07-full-stream/result.json", func(v map[string]any) {
					v["result"].(map[string]any)["metadata"].(map[string]any)["providerId"] = "minimax-cn"
				})
			case "cn-source":
				mutate(minimaxlive.ReportFile, func(v map[string]any) {
					v["sourceHash"] = "cf78d95276518b97c551cf887f3655f02de0749a25ffc867ae8dba5f6a0f9b9a"
				})
			case "cn-prices":
				mutate(minimaxlive.ReportFile, func(v map[string]any) {
					v["catalogHash"] = "sha256:0ef9cf881b44453009017f52b120b44ad575f4ecd6c35cf79e956f2d2ba66659"
					v["catalogVersion"] = "minimax-m3-standard-priced-candidate-2026-10-18.1"
				})
			case "candidate-mode-spoof", "formal-mode-spoof":
				mutate(minimaxlive.ReportFile, func(v map[string]any) {
					v["mode"] = "live"
					if name == "formal-mode-spoof" {
						v["kind"] = "minimax-formal"
						v["stage"] = "formal"
					}
				})
				mutate("manifest.json", func(v map[string]any) {
					v["versions"].(map[string]any)["minimax_mode"] = "live"
					if name == "formal-mode-spoof" {
						v["versions"].(map[string]any)["minimax_kind"] = "minimax-formal"
					}
				})

			case "invented-case":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { v["calls"].([]any)[0].(map[string]any)["id"] = "invented" })
			case "omitted-attempt":
				mutate("MM07-full-stream/result.json", func(v map[string]any) {
					v["result"].(map[string]any)["metadata"].(map[string]any)["attempts"] = []any{}
				})
			case "altered-usage":
				mutate("MM07-full-stream/result.json", func(v map[string]any) {
					v["result"].(map[string]any)["message"].(map[string]any)["usage"].(map[string]any)["input"] = 999
				})
			case "coordinated-usage":
				for _, path := range []string{"MM07-full-stream/result.json", "MM07-full-stream/observations.json"} {
					var v any
					json.Unmarshal(saved[path], &v)
					var replace func(any)
					replace = func(x any) {
						switch x := x.(type) {
						case map[string]any:
							if _, ok := x["input"]; ok {
								if _, yes := x["cost"]; yes {
									x["input"] = 999
								}
							}
							for _, child := range x {
								replace(child)
							}
						case []any:
							for _, child := range x {
								replace(child)
							}
						}
					}
					replace(v)
					raw, _ := json.Marshal(v)
					os.WriteFile(filepath.Join(dir, path), raw, 0600)
				}
			case "altered-request":
				os.WriteFile(filepath.Join(dir, "MM07-full-stream/request.json"), []byte(`{"model":"different"}`), 0600)
			case "altered-events":
				os.WriteFile(filepath.Join(dir, "MM07-full-stream/events.json"), []byte(`[]`), 0600)
			case "source-addition":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { v["executionSource"].(map[string]any)["ai/invented.go"] = "invented" })
			case "source-deletion":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { delete(v["executionSource"].(map[string]any), "ai/client.go") })
			case "ledger-omission":
				mutate("minimax-ledger.json", func(v map[string]any) { v["receipts"] = v["receipts"].([]any)[1:] })
			case "unknown-field":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { v["invented"] = true })
			case "duplicate-field":
				raw := append([]byte(`{"schema":1,`), saved[minimaxlive.ReportFile][1:]...)
				os.WriteFile(filepath.Join(dir, minimaxlive.ReportFile), raw, 0600)
			case "noncanonical-field":
				mutate(minimaxlive.ReportFile, func(v map[string]any) { v["SCHEMA"] = v["schema"]; delete(v, "schema") })
			case "extra-file":
				os.WriteFile(filepath.Join(dir, "invented.json"), []byte(`{}`), 0600)
				defer os.Remove(filepath.Join(dir, "invented.json"))
			case "unsafe-file":
				os.Symlink(filepath.Join(dir, "manifest.json"), filepath.Join(dir, "link.json"))
				defer os.Remove(filepath.Join(dir, "link.json"))
			}
			if name != "unsafe-file" {
				files, _ := minimaxlive.Inventory(dir)
				raw, _ := json.Marshal(minimaxlive.Seal{Schema: 1, Files: files})
				os.WriteFile(filepath.Join(dir, minimaxlive.SealFile), raw, 0600)
			}
			_, err := minimaxlive.Verify(dir, name == "formal-controlled")
			cs.Check("tampered or ineligible report refused", err != nil, "accepted %s", name)
			if strings.HasSuffix(name, "mode-spoof") {
				cs.Check("actual-only external origin required", err != nil && strings.Contains(err.Error(), "origin"), "%v", err)
			}
		})
	}
}
