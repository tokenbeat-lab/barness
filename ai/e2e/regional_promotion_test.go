package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The old proof consumer can emit controlled nonproduction proposals only.
// Runtime activation belongs to the maintenance candidate/check/apply flow.
func TestRegionalPromotionBoundary(t *testing.T) {
	ev := run.Case(t, "RP18-promotion-boundary")
	binary := filepath.Join(t.TempDir(), "regionalpromote")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalpromote")
	build.Dir, build.Env = "../..", regionalChildEnvironment()
	raw, err := build.CombinedOutput()
	if !ev.Check("controlled proposal command builds", err == nil, "%v %s", err, raw) {
		return
	}
	before, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	help := exec.CommandContext(ctxFor(t), binary, "-help")
	help.Env = regionalChildEnvironment()
	raw, _ = help.CombinedOutput()
	ev.Check("help exposes no activation receipt", !strings.Contains(string(raw), "-receipt"), "%s", raw)
	receipt := filepath.Join(t.TempDir(), "lineage.json")
	args := []string{"-receipt", receipt, "absent-zai-chat", "absent-zai-responses", "absent-zai-messages", "absent-cn-chat", "absent-cn-responses", "absent-cn-messages"}
	cmd := exec.CommandContext(ctxFor(t), binary, args...)
	cmd.Env = regionalChildEnvironment()
	raw, err = cmd.CombinedOutput()
	ev.Record("retired-activation", map[string]any{"argv": cmd.Args, "failed": err != nil, "output": string(raw), "vendorCalls": 0})
	ev.Check("activation option is removed", err != nil && strings.Contains(string(raw), "flag provided but not defined: -receipt"), "%v %s", err, raw)
	_, err = os.Lstat(receipt)
	ev.Check("no activation intent created", os.IsNotExist(err), "created receipt")
	after, err := os.ReadFile("../catalog_builtin.json")
	ev.Check("proposal command cannot change generated catalog", err == nil && bytes.Equal(before, after), "catalog changed")
	ev.Fixture("unchanged-generated-catalog.json", after)
}
