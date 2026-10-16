package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/tidwall/sjson"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func visualHash(raw []byte) string { s := sha256.Sum256(raw); return hex.EncodeToString(s[:]) }
func visualReseal(t *testing.T, dir string) {
	t.Helper()
	m := visualRead(t, dir, "manifest.json")
	files := m["artifacts"].(map[string]any)
	for name := range files {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			delete(files, name)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		files[name] = visualHash(raw)
	}
	visualWrite(t, filepath.Join(dir, "manifest.json"), m)
}

// Patch only changed leaves so unchanged RawMessage answer ordering stays exact.
func visualPatch(t *testing.T, path string, value any) {
	t.Helper()
	raw, _ := os.ReadFile(path)
	var before any
	if json.Unmarshal(raw, &before) != nil {
		visualWrite(t, path, value)
		return
	}
	var patch func(string, any, any)
	patch = func(pointer string, old, new any) {
		if reflect.DeepEqual(old, new) {
			return
		}
		if om, ok := old.(map[string]any); ok {
			if nm, ok := new.(map[string]any); ok {
				for k := range om {
					if _, ok := nm[k]; !ok {
						var err error
						raw, err = sjson.DeleteBytes(raw, pointer+"."+k)
						if err != nil {
							t.Fatal(err)
						}
					}
				}
				for k, v := range nm {
					next := k
					if pointer != "" {
						next = pointer + "." + k
					}
					patch(next, om[k], v)
				}
				return
			}
		}
		if oa, ok := old.([]any); ok {
			if na, ok := new.([]any); ok && len(oa) == len(na) {
				for i, v := range na {
					next := strconv.Itoa(i)
					if pointer != "" {
						next = pointer + "." + next
					}
					patch(next, oa[i], v)
				}
				return
			}
		}
		v, _ := json.Marshal(new)
		if pointer == "" {
			raw = v
			return
		}
		var err error
		raw, err = sjson.SetRawBytes(raw, pointer, v)
		if err != nil {
			t.Fatal(err)
		}
	}
	patch("", before, value)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestVisualEvaluationIntegrity(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		t.Run(route, func(t *testing.T) {
			out := filepath.Join(run.Dir(), "visual-integrity-"+route)
			if _, err := visualCommand(t, "-route", route, "-out", out); err != nil {
				t.Fatal("controlled workflow failed")
			}
			for _, name := range []string{"score", "prediction", "response", "request", "observer", "inventory", "source", "protocol", "promotion", "time", "symlink", "hidden_calls", "assertion_keys"} {
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "VQ12-integrity-"+route+"-"+name)
					if output, err := visualCommand(t, "-verify", out); err != nil {
						t.Logf("baseline verifier: %s", output)
						t.Fatal("baseline package failed before tamper")
					}
					backups := map[string][]byte{}
					edit := func(path string, v any) {
						t.Helper()
						full := filepath.Join(out, path)
						if _, ok := backups[path]; !ok {
							raw, err := os.ReadFile(full)
							if err != nil {
								t.Fatal(err)
							}
							backups[path] = raw
						}
						if path == "manifest.json" {
							visualWrite(t, full, v)
						} else {
							visualPatch(t, full, v)
						}
					}
					r := visualRead(t, out, "report.json")
					s := r["samples"].([]any)[0].(map[string]any)
					switch name {
					case "score":
						r["metrics"].(map[string]any)["common"].(map[string]any)["classCorrect"] = float64(8)
						edit("report.json", r)
					case "prediction":
						s["view"].(map[string]any)["kind"].(map[string]any)["choice"] = "blue"
						edit("report.json", r)
						edit("samples/v01/result.json", s)
					case "response", "request":
						w := visualRead(t, out, "samples/v01/wire.json")
						field, file := "response", "response.bin"
						if name == "request" {
							field, file = "request", "request.json"
						}
						raw, _ := base64.StdEncoding.DecodeString(w[field].(string))
						var body map[string]any
						json.Unmarshal(raw, &body)
						if name == "response" {
							body["model"] = "changed-model"
						} else {
							body["state"] = "changed task"
							if route == "decisions" {
								body["input"] = "changed task"
							}
						}
						raw, _ = json.Marshal(body)
						w[field] = base64.StdEncoding.EncodeToString(raw)
						edit("samples/v01/wire.json", w)
						path := "samples/v01/" + file
						backup, _ := os.ReadFile(filepath.Join(out, path))
						backups[path] = backup
						os.WriteFile(filepath.Join(out, path), raw, 0600)
					case "observer":
						v := []any{}
						edit("samples/v01/observations.json", v)
					case "inventory":
						edit("manifest.json", map[string]any{"schema": 1, "goVersion": "unknown", "artifacts": map[string]any{}})
					case "source":
						edit("source/go.mod", map[string]any{"changed": true})
					case "protocol":
						v := visualRead(t, out, "protocol-evidence.json")
						for _, live := range v["live"].([]any) {
							l := live.(map[string]any)
							if l["route"] == route {
								l["catalogHash"] = "sha256:changed"
							}
						}
						edit("protocol-evidence.json", v)
						raw, _ := os.ReadFile(filepath.Join(out, "protocol-evidence.json"))
						r["protocolEvidenceHash"] = visualHash(raw)
						edit("report.json", r)
					case "promotion":
						r["mode"] = "live"
						r["realQuality"] = "COMPLETE"
						edit("report.json", r)
						config := visualRead(t, out, "run-config.json")
						config["mode"] = "live"
						edit("run-config.json", config)
					case "time":
						end, _ := time.Parse(time.RFC3339Nano, r["finishedAt"].(string))
						s["finishedAt"] = end.Add(time.Millisecond).Format(time.RFC3339Nano)
						edit("report.json", r)
						edit("samples/v01/result.json", s)
					case "hidden_calls":
						notrun := filepath.Join(run.Dir(), "visual-hidden-notrun-"+route)
						if _, e := visualCommand(t, "-route", route, "-mode", "live", "-out", notrun); e != nil {
							t.Fatal("NOT_RUN setup failed")
						}
						v := visualRead(t, notrun, "report.json")
						v["mode"] = "controlled"
						edit("report.json", v)
					case "assertion_keys":
						assertions := s["assertions"].(map[string]any)
						delete(assertions, "observerTerminal")
						assertions["unrelated"] = true
						edit("report.json", r)
						edit("samples/v01/result.json", s)
					case "symlink":
						path := filepath.Join(out, "audit.json")
						raw, _ := os.ReadFile(path)
						backups["audit.json"] = raw
						target := filepath.Join(t.TempDir(), "audit.json")
						os.WriteFile(target, raw, 0600)
						os.Remove(path)
						os.Symlink(target, path)
					}
					manifestRaw, _ := os.ReadFile(filepath.Join(out, "manifest.json"))
					if _, ok := backups["manifest.json"]; !ok {
						backups["manifest.json"] = manifestRaw
					}
					visualReseal(t, out)
					output, err := visualCommand(t, "-verify", out, "-require-live")
					if name == "promotion" || name == "time" || name == "hidden_calls" || name == "assertion_keys" {
						t.Logf("tamper verifier: %s", output)
					}
					// Promotion specifically checks live provenance; all other corruptions must
					// fail ordinary verification too, independently of the controlled/live guard.
					if name != "promotion" {
						_, err = visualCommand(t, "-verify", out)
					}
					ev.Check("resealed semantic corruption rejected", err != nil, "accepted tampered package")
					for path, raw := range backups {
						full := filepath.Join(out, path)
						os.Remove(full)
						if err := os.WriteFile(full, raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
				})
			}
			if _, err := visualCommand(t, "-verify", out); err != nil {
				t.Fatal("original package did not restore")
			}
		})
	}
}
