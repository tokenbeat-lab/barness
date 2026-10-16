package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func visualWrite(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestVisualEvaluationFailureRetention(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, entry := range []string{"plain", "hooked"} {
			for _, name := range []string{"http", "empty", "refused", "partial", "unreported", "drift", "budget", "deadline", "iou", "redaction", "oversize"} {
				t.Run(route+"/"+entry+"/"+name, func(t *testing.T) {
					ev := run.Case(t, "VQ12-failure-"+route+"-"+entry+"-"+name)
					raw, err := os.ReadFile("../examples/visualeval/testdata/" + route + "-responses.json")
					if err != nil {
						t.Fatal(err)
					}
					var replies []map[string]any
					json.Unmarshal(raw, &replies)
					body := replies[0]["body"].(map[string]any)
					args := []string{"-route", route, "-entry", entry}
					wantStatus := "FAIL"
					wantFirst := "failed"
					switch name {
					case "http":
						replies[0]["status"] = float64(529)
						replies[0]["body"] = map[string]any{"error": "synthetic outage"}
					case "redaction":
						replies[0]["status"] = float64(529)
						replies[0]["body"] = map[string]any{"error": "synthetic-" + "visual-evaluation-credential"}
					case "oversize":
						body["padding"] = strings.Repeat("x", 70000)
					case "empty":
						replies[0]["body"] = nil
					case "refused":
						wantStatus = "COMPLETE"
						wantFirst = "refused"
						if route == "jev" {
							wantStatus = "FAIL"
							wantFirst = "failed"
						}
						if route == "jev" {
							body["answers"] = map[string]any{"kind": map[string]any{"type": "refusal"}, "exists": map[string]any{"type": "refusal"}}
						} else {
							body["answers"] = []any{map[string]any{"type": "refusal", "name": "kind"}, map[string]any{"type": "refusal", "name": "exists"}}
						}
					case "partial":
						wantFirst = "answered"
						body["usage"] = map[string]any{"input_tokens": 100}
					case "unreported":
						wantFirst = "answered"
						delete(body, "usage")
					case "drift":
						wantFirst = "version_drift"
						body["model"] = "unexpected-model"
					case "budget":
						wantFirst = "answered"
						args = append(args, "-calls", "1")
					case "deadline":
						wantFirst = "cancelled"
						replies[0]["delayMS"] = float64(1000)
						args = append(args, "-work-ms", "200", "-call-ms", "100")
					case "iou":
						wantStatus = "COMPLETE"
						wantFirst = "answered"
						if route == "jev" {
							b := replies[8]["body"].(map[string]any)["answers"].(map[string]any)["location"].(map[string]any)["box"].(map[string]any)
							b["left"] = float64(.5)
							b["right"] = float64(1)
						}
					}
					script := filepath.Join(t.TempDir(), "script.json")
					visualWrite(t, script, replies)
					out := filepath.Join(run.Dir(), "visual-failure-"+route+"-"+entry+"-"+name)
					_, err = visualCommand(t, append(args, "-responses", script, "-out", out)...)
					ev.Check("terminal exit reflects completeness", (err == nil) == (wantStatus == "COMPLETE"), "unexpected exit")
					r := visualRead(t, out, "report.json")
					ev.Record("report", r)
					samples := r["samples"].([]any)
					ev.Check("failure never drops annotations", r["status"] == wantStatus && len(samples) == 12 && samples[0].(map[string]any)["status"] == wantFirst, "wrong terminal/sample")
					common := r["metrics"].(map[string]any)["common"].(map[string]any)
					ev.Check("full denominator survives failure", common["total"] == float64(8), "lost sample")
					b := r["budget"].(map[string]any)
					ev.Check("budget is finite zero retry", b["calls"].(float64) <= 12 && b["attempts"].(float64) <= 12 && b["retries"] == float64(0) && b["questions"].(float64) <= 48 && b["occurrences"].(float64) <= 24 && b["sourceBytes"].(float64) <= 8<<20 && b["expandedBytes"].(float64) <= 16<<20, "budget overflow")
					if name == "redaction" {
						w := visualRead(t, out, "samples/v01/wire.json")
						captured, _ := base64.StdEncoding.DecodeString(w["response"].(string))
						ev.Check("binary capture is redacted before encoding", !bytes.Contains(captured, []byte("synthetic-"+"visual-evaluation-credential")), "credential survived encoded capture")
					}
					if name == "budget" {
						ev.Check("remaining samples preserved unsent", b["calls"] == float64(1) && samples[11].(map[string]any)["status"] == "budget_exhausted", "resumed or dropped samples")
					}
					if name == "iou" && route == "jev" {
						box := r["metrics"].(map[string]any)["box"].(map[string]any)
						ev.Check("intersection over union hand worked", box["meanIoU"] == 2.0/3.0, "wrong IoU")
					}
					if name == "partial" {
						ev.Check("partial spend is not free", b["partial"] == float64(1) && b["inputTokens"] == float64(1200), "lost partial usage")
					}
					if name == "unreported" || name == "http" || name == "empty" {
						ev.Check("unknown usage retained", b["unreported"] == float64(1), "invented complete usage")
					}
					_, err = visualCommand(t, "-verify", out)
					ev.Check("honest failure package replays", err == nil, "failure package rejected")
				})
			}
		}
	}
}
func TestVisualEvaluationNotRun(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "VQ12-not-run-"+route)
			out := filepath.Join(run.Dir(), "visual-not-run-"+route)
			_, err := visualCommand(t, "-route", route, "-mode", "live", "-out", out)
			ev.Check("missing authorization records NOT_RUN", err == nil, "missing configuration failed to retain receipt")
			r := visualRead(t, out, "report.json")
			ev.Record("report", r)
			b := r["budget"].(map[string]any)
			common := r["metrics"].(map[string]any)["common"].(map[string]any)
			ev.Check("no invented real result or spend", r["status"] == "NOT_RUN" && r["realQuality"] == "NOT_RUN" && b["calls"] == float64(0) && b["attempts"] == float64(0) && len(r["samples"].([]any)) == 12 && common["classAccuracy"] == nil, "misleading NOT_RUN")
			_, err = visualCommand(t, "-verify", out)
			ev.Check("NOT_RUN inventory verifies", err == nil, "invalid NOT_RUN")
			_, err = visualCommand(t, "-verify", out, "-require-live")
			ev.Check("NOT_RUN cannot satisfy quality", err != nil, "promoted NOT_RUN")
		})
	}
}
