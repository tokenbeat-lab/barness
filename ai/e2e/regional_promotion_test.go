package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Only independently captured live candidate acceptance may change production.
// Missing evidence, a changed target, an existing receipt or a held activation
// lock must leave both the target and lineage history untouched.
func TestRegionalPromotionBoundary(t *testing.T) {
	ev := run.Case(t, "RP18-promotion-boundary")
	binary := filepath.Join(t.TempDir(), "regionalpromote")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalpromote")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	if !ev.Check("public promotion command builds", err == nil, "%v %s", err, raw) {
		return
	}
	original, err := os.ReadFile("testdata/regional/inactive-catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ai"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module isolated-regional-promotion\n"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "ai/catalog_regional.json")
	evidence := filepath.Join(root, ".evidence/provider-sites/final-acceptance-20261020")
	if err := os.MkdirAll(evidence, 0700); err != nil {
		t.Fatal(err)
	}
	receipt := filepath.Join(evidence, "lineage.json")
	inputs := []string{"absent-zai-chat", "absent-zai-responses", "absent-zai-messages", "absent-cn-chat", "absent-cn-responses", "absent-cn-messages"}
	invoke := func(name, want string, args ...string) {
		child := exec.CommandContext(ctxFor(t), binary, append(args, inputs...)...)
		child.Dir, child.Env = root, []string{}
		raw, err := child.CombinedOutput()
		ev.Record(name, map[string]any{"argv": child.Args[1:], "failed": err != nil, "output": string(raw), "vendorCalls": 0})
		ev.Check("refused "+name, err != nil && strings.Contains(string(raw), want), "%v %s", err, raw)
	}
	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}
	invoke("invalid-receipt-location", "current acceptance .evidence", "-receipt", filepath.Join(root, "outside.json"))
	invoke("missing-own-candidates", "regional", "-receipt", receipt)
	if _, err := os.Lstat(receipt); !os.IsNotExist(err) {
		t.Fatal("refused activation created receipt")
	}
	invoke("controlled-cannot-activate", "controlled", "-controlled", "-receipt", receipt)
	changed := []byte(`{"schema":1,"active":true,"version":"edited","models":[]}`)
	if err := os.WriteFile(target, changed, 0600); err != nil {
		t.Fatal(err)
	}
	invoke("changed-target", "inactive catalog target changed", "-receipt", receipt)
	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target+".activation.lock", 0700); err != nil {
		t.Fatal(err)
	}
	invoke("held-activation-lock", "activation locked", "-receipt", receipt)
	after, err := os.ReadFile(target)
	ev.Check("rejections preserve target bytes", err == nil && string(after) == string(original), "%v", err)
	ev.Fixture("inactive-catalog.json", original)
}
