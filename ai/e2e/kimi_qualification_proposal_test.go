package e2e

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// Full-ten captures remain valid for their own qualification contract. Kimi
// promotion now consumes exactly six independently proven nine-plus-one pairs.
func kimiQualificationPromotion(t *testing.T, dirs []string) {
	t.Helper()
	ev := run.Case(t, "KIMI-K16-standalone-promotion-refused")
	binary := kimiPromotionCLI(t)
	path := filepath.Join(t.TempDir(), "refused.json")
	cmd := exec.CommandContext(ctxFor(t), binary, append([]string{"-kimi", "-controlled", "-proposal", path}, dirs...)...)
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
	ev.Check("standalone captures cannot replace fixed paired promotion", err != nil, "%s", raw)
}
