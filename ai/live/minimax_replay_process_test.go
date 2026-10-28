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

func TestMiniMaxReplayProcess(t *testing.T) {
	for _, route := range []string{"chat", "responses", "messages", "rerun-tool", "transient-request", "transient-transport"} {
		t.Run(route, func(t *testing.T) {
			cs := run.Case(t, "MM07-replay-"+route)
			control := "success-" + route
			if route == "rerun-tool" || strings.HasPrefix(route, "transient-") {
				control = "success-" + route
				route = "chat"
			}
			base := t.TempDir()
			exe, _ := os.Executable()
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=" + route, "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=global-test", "BARNESS_MINIMAX_CONTROLLED=" + control, "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "source-stage.jsonl"), "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "source")}
			execute := func(entry string, environment []string) []byte {
				ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=35s")
				cmd.Env = environment
				output, e := cmd.CombinedOutput()
				if !cs.Check("bounded child passes "+entry, e == nil, "err=%v output=%s", e, output) {
					t.FailNow()
				}
				return output
			}
			execute("^TestMiniMaxCandidate$", env)
			paths, _ := filepath.Glob(filepath.Join(base, "source", "*", minimaxlive.ReportFile))
			if len(paths) != 1 {
				t.Fatal("missing source")
			}
			env[0] = "BARNESS_AI_LIVE=0"
			env[4] = "BARNESS_MINIMAX_KEY="
			env[8] = "BARNESS_MINIMAX_CONTROLLED="
			env[9] = "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "replay-stage.jsonl")
			env[10] = "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "replay")
			env = append(env, "BARNESS_MINIMAX_REPLAY="+filepath.Dir(paths[0]))
			execute("^TestMiniMaxReplay$", env)
			replay, _ := filepath.Glob(filepath.Join(base, "replay", "*", minimaxlive.ReportFile))
			if !cs.Check("replay source distinct", len(replay) == 1, "count=%d", len(replay)) {
				return
			}
			r, e := minimaxlive.Verify(filepath.Dir(replay[0]), false)
			cs.Check("own complete replay preserves mode", e == nil && r.Mode == "replay" && r.Kind == "minimax-capture-replay-only", "err=%v", e)
			archive, archiveErr := archiveMiniMaxProcess(filepath.Dir(replay[0]))
			if cs.Check("replay captures retained", archiveErr == nil, "%v", archiveErr) {
				cs.Fixture("replay-capture.tar.gz", archive)
			}
			origin, e := os.ReadFile(filepath.Join(filepath.Dir(replay[0]), "minimax-replay-origin.json"))
			cs.Check("immutable replay source retained", e == nil && len(origin) > 0, "origin absent")
			_, e = minimaxlive.Verify(filepath.Dir(replay[0]), true)
			cs.Check("replay cannot become live", e != nil, "formal accepted replay")
			replayDir := filepath.Dir(replay[0])
			r.Mode = "live"
			r.Kind = "minimax-candidate-only"
			r.ReplaySource = ""
			raw, _ := json.Marshal(r)
			os.WriteFile(filepath.Join(replayDir, minimaxlive.ReportFile), raw, 0600)
			manifestRaw, _ := os.ReadFile(filepath.Join(replayDir, "manifest.json"))
			var m map[string]any
			json.Unmarshal(manifestRaw, &m)
			m["versions"].(map[string]any)["minimax_mode"] = "live"
			m["versions"].(map[string]any)["minimax_kind"] = "minimax-candidate-only"
			manifestRaw, _ = json.Marshal(m)
			os.WriteFile(filepath.Join(replayDir, "manifest.json"), manifestRaw, 0600)
			os.Remove(filepath.Join(replayDir, "minimax-replay-origin.json"))
			files, _ := minimaxlive.Inventory(replayDir)
			seal, _ := json.Marshal(minimaxlive.Seal{Schema: 1, Files: files})
			os.WriteFile(filepath.Join(replayDir, minimaxlive.SealFile), seal, 0600)
			_, e = minimaxlive.Verify(replayDir, false)
			cs.Check("removed replay provenance cannot forge actual origin", e != nil && strings.Contains(e.Error(), "fresh process approval missing"), "%v", e)

		})
	}
}
