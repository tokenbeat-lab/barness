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

	"github.com/tokenbeat-lab/barness/ai"
)

// The same fixed22 public executable runs without sockets or real credentials.
// Only the low-effort scene is large, matching the actual capture failure shape.
func TestGLMCorrectiveCaptureProcess(t *testing.T) {
	for _, name := range []string{"large", "length", "overcap"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "GLM-corrective-1-capture-"+name)
			base := t.TempDir()
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMFormal$", "-test.timeout=80s")
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=responses", "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_GLM_CONTROLLED=corrective-" + name, "BARNESS_AI_EVIDENCE_DIR=" + base}
			output, err := cmd.CombinedOutput()
			cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
			cs.Check("full fixed process truthful exit", (err == nil) == (name == "large"), "%v %s", err, output)
			matches, _ := filepath.Glob(filepath.Join(base, "*", "formal-glm-report.json"))
			if !cs.Check("own report retained", len(matches) == 1, "reports=%d", len(matches)) {
				return
			}
			dir := filepath.Dir(matches[0])
			cs.Check("full controlled capture retained", glmRetainCorrectiveChild(cs, dir, "controlled-capture") == nil, "capture copy failed")
			raw, _ := os.ReadFile(matches[0])
			cs.Fixture("formal-glm-report.json", raw)
			var report glmReport
			json.Unmarshal(raw, &report)
			cs.Check("all22 attempts retained", report.Budget.CallsUsed == 22 && report.Budget.HTTPAttempts == 22 && len(report.Scenarios) == 22, "budget=%+v", report.Budget)
			own := filepath.Join(dir, "GLM11-glm-5.3-control-low")
			var ex []exchange
			raw, _ = os.ReadFile(filepath.Join(own, "attempt-1-exchanges.json"))
			json.Unmarshal(raw, &ex)
			if !cs.Check("single observed HTTP attempt", len(ex) == 1, "attempts=%d", len(ex)) {
				return
			}
			cs.Record("wire-metadata", map[string]any{"bytes": len(ex[0].ResponseBody), "truncated": ex[0].Truncated, "sha256": sha256Text([]byte(ex[0].ResponseBody))})
			if name == "overcap" {
				cs.Check("overcap truthfully fails without publication", ex[0].Truncated && report.Outcome == "FAIL" && !report.Complete, "outcome=%s", report.Outcome)
				return
			}
			cs.Check("entire legal large SSE including late usage retained", !ex[0].Truncated && len(ex[0].ResponseBody) > 512<<10 && strings.Contains(ex[0].ResponseBody, `"output_tokens":4096`), "truncated=%v bytes=%d", ex[0].Truncated, len(ex[0].ResponseBody))
			var result struct {
				Result ai.Result
				Error  string
			}
			raw, _ = os.ReadFile(filepath.Join(own, "attempt-1-result.json"))
			json.Unmarshal(raw, &result)
			cs.Fixture("result.json", raw)
			cs.Check("late known usage and native state retained", result.Result.Message.Usage.Output == 4096 && result.Result.Message.Usage.Reasoning == ai.Value[int64](4094) && len(result.Result.Message.Content) > 0, "usage=%+v", result.Result.Message.Usage)
			if name == "length" {
				cs.Check("length thinking-only is never support PASS", report.Outcome == "FAIL" && !report.Complete && result.Result.Message.StopReason == ai.StopReasonLength, "outcome=%s stop=%s", report.Outcome, result.Result.Message.StopReason)
			}
			var events struct {
				Schema int
				Events []struct {
					Type  ai.EventType
					Call  ai.CallAttribution
					Event struct {
						Delta   string
						Partial json.RawMessage
					}
				}
				FinalPartial struct {
					StopReason ai.StopReason `json:"stopReason"`
				}
			}
			raw, _ = os.ReadFile(filepath.Join(own, "events.json"))
			cs.Fixture("events.json", raw)
			decode := json.Unmarshal(raw, &events)
			cs.Check("bounded explicit event-owned schema", decode == nil && events.Schema == 1 && len(raw) < 4<<20, "decode=%v bytes=%d", decode, len(raw))
			deltas := 0
			text := ""
			for _, event := range events.Events {
				cs.Check("own attribution on every envelope", event.Call.RequestID == result.Result.Metadata.RequestID && event.Call.ModelID == "glm-5.3", "foreign event")
				if event.Type == ai.EventThinkingDelta {
					deltas++
					text += event.Event.Delta
					cs.Check("live Partial not serialized as historical snapshot", string(event.Event.Partial) == "null", "duplicate Partial")
				}
			}
			cs.Check("every fragmented delta in order retained", deltas == 4094 && text == strings.Repeat("x", 4094), "deltas=%d bytes=%d", deltas, len(text))
			cs.Check("terminal and final Partial retained", len(events.Events) > 0 && events.Events[len(events.Events)-1].Type == ai.EventDone && events.FinalPartial.StopReason == result.Result.Message.StopReason, "missing terminal")
			files, _ := filepath.Glob(filepath.Join(dir, "live-report.json"))
			cs.Check("controlled never actual publication", len(files) == 0, "controlled publication")
			if name == "large" {
				glmCorrectiveReplay(t, cs, dir)
			}
		})
	}
}
