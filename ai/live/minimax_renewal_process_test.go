//go:build live

package live

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Fixed renewal is an authorization boundary, never a controlled formal mode.
func TestMiniMaxFormalRenewalControlledRefusal(t *testing.T) {
	cs := run.Case(t, "MM19-renewal-controlled-refusal")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	// Normalize only this case's diagnostics; keep the original argv for exec
	// and leave the shared environment audit/redactor unchanged.
	aliases := strings.NewReplacer(run.Dir(), "$MINIMAX_RENEWAL_ROOT", exe, "$MINIMAX_RENEWAL_TEST_BINARY")
	for _, mode := range []string{"controlled-candidate", "controlled-formal", "unknown-renewal", "final-controlled-candidate", "final-controlled-formal", "recovery-controlled-candidate", "recovery-controlled-formal"} {
		out := filepath.Join(run.Dir(), "renewal-"+mode)
		id := "current-source-eleven-renewal-1"
		if mode == "unknown-renewal" {
			id = "invented-round"
		}
		if strings.HasPrefix(mode, "recovery-") {
			id = "provider-sites-final-acceptance-3"
		}
		if strings.HasPrefix(mode, "final-") {
			id = "provider-sites-final-acceptance-1"
		}
		entry := "^TestMiniMaxCandidate$"
		formal := "0"
		if strings.HasSuffix(mode, "controlled-formal") {
			entry = "^TestMiniMaxFormal$"
			formal = "1"
		}
		cmd := exec.Command(exe, "-test.run="+entry, "-test.timeout=30s")
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-renewal-key", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic", "BARNESS_MINIMAX_REGION_ALIAS=synthetic", "BARNESS_MINIMAX_CONTROLLED=success-chat", "BARNESS_MINIMAX_FORMAL=" + formal, "BARNESS_MINIMAX_FORMAL_RENEWAL=" + id, "BARNESS_AI_EVIDENCE_DIR=" + out, "BARNESS_MINIMAX_LEDGER=" + filepath.Join(out, "private.jsonl")}
		raw, e := cmd.CombinedOutput()
		output := aliases.Replace(string(raw))
		diagnosticArgv := append([]string(nil), cmd.Args...)
		for i := range diagnosticArgv {
			diagnosticArgv[i] = aliases.Replace(diagnosticArgv[i])
		}
		cs.Record(mode+"-command", map[string]any{"argv": diagnosticArgv, "output": output, "failed": e != nil, "paid": false})
		cs.Check(mode+" rejects before selected calls", e != nil, "output=%s", output)
		cs.Check(mode+" diagnostic retains owned suffix through alias", strings.Contains(output, "$MINIMAX_RENEWAL_ROOT/renewal-"+mode+"/"), "owned alias and path suffix required")
		cs.Check(mode+" recorded argv preserves exact selection", len(diagnosticArgv) == 3 && diagnosticArgv[0] == "$MINIMAX_RENEWAL_TEST_BINARY" && diagnosticArgv[1] == "-test.run="+entry && diagnosticArgv[2] == "-test.timeout=30s", "aliased executable and exact flags required")
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
		cs.Check(mode+" no formal publication", os.IsNotExist(e), "%s", aliases.Replace(errText(e)))
	}
}
