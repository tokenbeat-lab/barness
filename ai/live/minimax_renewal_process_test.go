//go:build live

package live

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Fixed renewal is an authorization boundary, never a controlled formal mode.
func TestMiniMaxFormalRenewalControlledRefusal(t *testing.T) {
	cs := run.Case(t, "MM19-renewal-controlled-refusal")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"controlled-candidate", "controlled-formal", "unknown-renewal"} {
		out := filepath.Join(run.Dir(), "renewal-"+mode)
		id := "current-source-eleven-renewal-1"
		if mode == "unknown-renewal" {
			id = "invented-round"
		}
		entry := "^TestMiniMaxCandidate$"
		formal := "0"
		if mode == "controlled-formal" {
			entry = "^TestMiniMaxFormal$"
			formal = "1"
		}
		cmd := exec.Command(exe, "-test.run="+entry, "-test.timeout=30s")
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-renewal-key", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic", "BARNESS_MINIMAX_REGION_ALIAS=synthetic", "BARNESS_MINIMAX_CONTROLLED=success-chat", "BARNESS_MINIMAX_FORMAL=" + formal, "BARNESS_MINIMAX_FORMAL_RENEWAL=" + id, "BARNESS_AI_EVIDENCE_DIR=" + out, "BARNESS_MINIMAX_LEDGER=" + filepath.Join(out, "private.jsonl")}
		raw, e := cmd.CombinedOutput()
		cs.Record(mode+"-command", map[string]any{"argv": cmd.Args, "output": string(raw), "failed": e != nil, "paid": false})
		cs.Check(mode+" rejects before selected calls", e != nil, "output=%s", raw)
		paths, _ := filepath.Glob(filepath.Join(out, "*", "minimax-report.json"))
		if !cs.Check(mode+" own report", len(paths) == 1, "reports=%d", len(paths)) {
			continue
		}
		b, _ := os.ReadFile(paths[0])
		var r struct {
			Calls    []any `json:"calls"`
			Complete bool  `json:"complete"`
		}
		json.Unmarshal(b, &r)
		cs.Check(mode+" zero calls and no formal completeness", len(r.Calls) == 0 && !r.Complete, "calls=%d complete=%v", len(r.Calls), r.Complete)
		_, e = os.Stat(filepath.Join(filepath.Dir(paths[0]), "live-report.json"))
		cs.Check(mode+" no formal publication", os.IsNotExist(e), "%v", e)
	}
}
