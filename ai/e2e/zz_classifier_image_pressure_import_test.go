package e2e

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Failure scenarios are mutations of actual public design-load artifacts,
// including a fully missing/new report, wrong route/load/limits and leaks.
// The complete public pipeline is required to replay this filesystem seam.
func TestClassifierImagesPressureImport(t *testing.T) {
	modes := []string{"intact", "old-report", "case-fail", "missing-report", "missing-identity", "wrong-operation", "wrong-provider", "wrong-api", "wrong-model", "wrong-catalog", "small-load", "wrong-count", "wrong-image-bytes", "wrong-total-bytes", "wrong-base64", "wrong-data-url", "wrong-json-bytes", "changed-policy", "changed-limit", "wrong-ratio", "missing-headroom", "expanded-budget", "missing-batch", "concurrency", "missing-EOF", "missing-result-heap", "missing-alloc", "excessive-alloc", "missing-copies", "wrong-copies", "missing-throughput", "wrong-throughput", "cancel-not-proven", "recovery-not-proven", "resource-leak", "missing-resources", "tampered-capture", "null-batch-resources", "null-final-resources", "null-old-resources", "tiny-memory", "tiny-allocation", "duplicated-results", "empty-attempt", "wrong-answer-name"}
	for _, field := range []string{"schema", "identity", "load", "policy", "headroom", "heapBudget", "batches", "imageBase64Bytes", "imageDataURLBytes", "requestJSONBytes", "providerSharedResponseScriptBytes",
		"identity.operation", "identity.provider", "identity.api", "identity.model", "identity.catalogVersion", "identity.catalogHash",
		"load.calls", "load.tenants", "load.callsPerTenant", "load.stateBytes", "load.questions", "load.instructionBytes", "load.imageCount", "load.imageBytes", "load.totalImageBytes",
		"batch.kind", "batch.calls", "batch.successful", "batch.canceled", "batch.peakPermits", "batch.permitsAtEOF", "batch.clientUnaryBytesAtEOF", "batch.responseBytesPerCall", "batch.providerCapturedRequestBytes", "batch.heap", "batch.heapGrowth", "batch.heapAtEOF", "batch.heapWithResults", "batch.totalAllocDuringCalls", "batch.duration", "batch.callsPerSecond", "batch.resources"} {
		modes = append(modes, "delete-"+field)
	}
	for _, name := range []string{"local", "cloud"} {
		for _, mode := range modes {
			t.Run(name+"/"+mode, func(t *testing.T) {
				ev := pressureCase(t, "E08-classifier-images-import-"+name+"-"+mode)
				ev.ReplayRun("^Test(MixedPressure|ClassifierImagesPressure|ClassifierImagesPressureImport)$")
				dir := t.TempDir()
				oldID := "E08-mixed-pressure-" + name + "-design-load"
				newID := "E08-mixed-pressure-" + name + "-classifier-images-design-load"
				manifestCases := []map[string]any{}
				for _, id := range []string{oldID, newID} {
					dest := filepath.Join(dir, id)
					if err := os.MkdirAll(dest, 0755); err != nil {
						t.Fatal(err)
					}
					entries, err := os.ReadDir(filepath.Join(run.Dir(), id))
					if err != nil {
						t.Fatalf("run full public pipeline first: %v", err)
					}
					for _, entry := range entries {
						raw, err := os.ReadFile(filepath.Join(run.Dir(), id, entry.Name()))
						if err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(dest, entry.Name()), raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
					status := "PASS"
					if mode == "case-fail" && id == newID {
						status = "FAIL"
					}
					manifestCases = append(manifestCases, map[string]any{"id": id, "dir": id, "status": status})
				}
				path := filepath.Join(dir, newID, "classifier-images-pressure.json")
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var report map[string]any
				if err := json.Unmarshal(raw, &report); err != nil {
					t.Fatal(err)
				}
				identity := report["identity"].(map[string]any)
				load := report["load"].(map[string]any)
				rooms := report["headroom"].(map[string]any)
				batch := report["batches"].([]any)[0].(map[string]any)
				if field, ok := strings.CutPrefix(mode, "delete-"); ok {
					parent := report
					parts := strings.Split(field, ".")
					if parts[0] == "batch" {
						parent = batch
						parts = parts[1:]
					}
					for _, key := range parts[:len(parts)-1] {
						parent = parent[key].(map[string]any)
					}
					delete(parent, parts[len(parts)-1])
				}
				switch mode {
				case "old-report":
					report = map[string]any{"load": load}
				case "missing-identity":
					delete(report, "identity")
				case "wrong-operation":
					identity["operation"] = "chat"
				case "wrong-provider":
					identity["provider"] = "typesafe"
				case "wrong-api":
					identity["api"] = "typesafe-system-one"
				case "wrong-model":
					identity["model"] = "jev-1.13.0"
				case "wrong-catalog":
					identity["catalogHash"] = "sha256:wrong"
				case "small-load":
					load["stateBytes"] = 1024
				case "wrong-count":
					load["imageCount"] = 1
				case "wrong-image-bytes":
					load["imageBytes"] = []int{1, 1}
				case "wrong-total-bytes":
					load["totalImageBytes"] = 1
				case "wrong-base64":
					report["imageBase64Bytes"] = []int{1, 1}
				case "wrong-data-url":
					report["imageDataURLBytes"] = []int{1, 1}
				case "wrong-json-bytes":
					report["requestJSONBytes"] = 1
				case "changed-policy":
					report["policy"].(map[string]any)["MaxRequestBytes"] = 8 << 20
				case "changed-limit":
					rooms["MaxRequestBytes"].(map[string]any)["limit"] = 8 << 20
				case "wrong-ratio":
					rooms["MaxImageBytes"].(map[string]any)["ratio"] = .1
				case "missing-headroom":
					delete(rooms, "Classifier.Images.MaxImages")
				case "expanded-budget":
					report["heapBudget"] = 1 << 30
				case "missing-batch":
					report["batches"] = report["batches"].([]any)[:1]
				case "concurrency":
					batch["peakPermits"] = 1
				case "missing-EOF":
					delete(batch, "heapAtEOF")
				case "missing-result-heap":
					delete(batch, "heapWithResults")
				case "missing-alloc":
					delete(batch, "totalAllocDuringCalls")
				case "excessive-alloc":
					batch["totalAllocDuringCalls"] = 1 << 30
				case "missing-copies":
					delete(batch, "providerCapturedRequestBytes")
				case "wrong-copies":
					batch["providerCapturedRequestBytes"] = 1
				case "missing-throughput":
					delete(batch, "callsPerSecond")
				case "wrong-throughput":
					batch["callsPerSecond"] = 1
				case "cancel-not-proven":
					report["batches"].([]any)[1].(map[string]any)["canceled"] = 0
				case "recovery-not-proven":
					report["batches"].([]any)[2].(map[string]any)["successful"] = 0
				case "resource-leak":
					batch["resources"].(map[string]any)["bodies"] = 1
				case "null-batch-resources":
					batch["resources"] = map[string]any{"calls": nil, "permits": nil, "waiters": nil, "events": nil, "bodies": nil, "hostPermits": nil}
				case "tiny-memory":
					batch["heap"] = map[string]any{"baselineLive": 1, "peakLive": 1}
					batch["heapGrowth"] = 0
					batch["heapAtEOF"] = 1
					batch["heapWithResults"] = 1
					batch["totalAllocDuringCalls"] = 1
				case "tiny-allocation":
					batch["totalAllocDuringCalls"] = 1
				case "missing-resources":
					delete(batch, "resources")
				}
				raw, err = json.Marshal(report)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
				if mode == "missing-report" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "tampered-capture" {
					capture := filepath.Join(dir, newID, "design-requests.json")
					raw, err := os.ReadFile(capture)
					if err != nil {
						t.Fatal(err)
					}
					raw = []byte(strings.Replace(string(raw), "data:image/png;base64,", "data:image/jpeg;base64,", 1))
					if err := os.WriteFile(capture, raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "null-final-resources" || mode == "null-old-resources" {
					id := newID
					if mode == "null-old-resources" {
						id = oldID
					}
					if err := os.WriteFile(filepath.Join(dir, id, "resources.json"), []byte(`{"calls":null,"permits":null,"waiters":null,"events":null,"bodies":null,"hostPermits":null}`), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "duplicated-results" || mode == "empty-attempt" || mode == "wrong-answer-name" {
					for _, kind := range []string{"design", "cancel", "recovery"} {
						first, err := os.ReadFile(filepath.Join(dir, newID, kind+"-result-0.json"))
						if err != nil {
							t.Fatal(err)
						}
						for i := range int(load["calls"].(float64)) {
							path := filepath.Join(dir, newID, fmt.Sprintf("%s-result-%d.json", kind, i))
							data, err := os.ReadFile(path)
							if err != nil {
								t.Fatal(err)
							}
							var result map[string]any
							if err := json.Unmarshal(data, &result); err != nil {
								t.Fatal(err)
							}
							switch mode {
							case "duplicated-results":
								data = first
							case "empty-attempt":
								result["metadata"].(map[string]any)["attempts"] = []any{map[string]any{}}
								data, err = json.Marshal(result)
							case "wrong-answer-name":
								if kind != "cancel" {
									answers := result["answers"].(map[string]any)
									answers["unrelated"] = answers["q0"]
									delete(answers, "q0")
									data, err = json.Marshal(result)
								}
							}
							if err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(path, data, 0600); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				// Rehash the mutated bundle: semantic validation must reject a
				// forged report even when its inventory hashes are consistent.
				for _, c := range manifestCases {
					folder := filepath.Join(dir, c["dir"].(string))
					entries, err := os.ReadDir(folder)
					if err != nil {
						t.Fatal(err)
					}
					hashes := map[string]string{}
					for _, entry := range entries {
						data, err := os.ReadFile(filepath.Join(folder, entry.Name()))
						if err != nil {
							t.Fatal(err)
						}
						hashes[entry.Name()] = fmt.Sprintf("%x", sha256.Sum256(data))
					}
					c["artifacts"] = hashes
				}
				raw, err = json.Marshal(map[string]any{"package": release.OfflinePackage, "cases": manifestCases})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
				b, err := release.LoadBundle(dir)
				if err != nil {
					t.Fatal(err)
				}
				checks := release.DesignLoadEvidence(b)
				index := 0
				if name == "cloud" {
					index = 1
				}
				ev.Record("import", checks[index])
				if mode == "intact" {
					ev.Check("actual classifier image capacity imported", checks[index].Status == release.Pass, "%+v", checks[index])
				} else {
					ev.Check("missing or forged capacity rejected", checks[index].Status != release.Pass, "%+v", checks[index])
				}
			})
		}
	}
}
