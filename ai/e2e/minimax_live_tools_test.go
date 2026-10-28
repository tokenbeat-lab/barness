package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Public executable proof is part of the ordinary/race E2E bundle, so release
// tracing cannot accidentally depend on live-tag cases that were never loaded.
func TestMiniMaxLiveEvidenceTools(t *testing.T) {
	ev := run.Case(t, "MM10-live-evidence-tools")
	base := filepath.Dir(run.Dir())
	before, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, entry := range before {
		seen[entry.Name()] = true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, miniMaxNativeBinary(t), "-test.count=1", "-test.timeout=180s", "-test.run=^TestMiniMax")
	cmd.Dir = "../live"
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "BARNESS_AI_EVIDENCE_DIR="+base)
	output, e := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"success": e == nil, "credentialsInherited": false, "output": string(output)})
	if !ev.Check("complete isolated tools executable passes", e == nil, "%v %s", e, output) {
		return
	}
	var paths []string
	entries, e := os.ReadDir(base)
	for _, entry := range entries {
		if !seen[entry.Name()] && entry.IsDir() {
			paths = append(paths, filepath.Join(base, entry.Name(), "manifest.json"))
		}
	}
	if !ev.Check("one tools source bundle", e == nil && len(paths) == 1, "%d", len(paths)) {
		return
	}
	raw, e := os.ReadFile(paths[0])
	if e != nil {
		t.Fatal(e)
	}
	ev.Fixture("live-tools-manifest.json", raw)
	var manifest struct{ Cases []struct{ ID, Status string } }
	if json.Unmarshal(raw, &manifest) != nil {
		t.Fatal("invalid live tools manifest")
	}
	ev.Check("finite complete tools inventory", len(manifest.Cases) >= 45, "%d", len(manifest.Cases))
	for _, c := range manifest.Cases {
		ev.Check("controlled proof "+c.ID, c.Status == "PASS", "%s", c.Status)
	}
	audit, e := os.ReadFile(filepath.Join(filepath.Dir(paths[0]), "audit.json"))
	if e != nil {
		t.Fatal(e)
	}
	ev.Fixture("live-tools-audit.json", audit)
	var findings struct{ Findings []json.RawMessage }
	ev.Check("child redaction audit clean", json.Unmarshal(audit, &findings) == nil && len(findings.Findings) == 0, "audit failed")
}

// Shared by distinct native process contracts. Build once per test, then launch
// the exact controlled entry directly; the outer collector owns the Go test.
func miniMaxNativeBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "minimax-native.test")
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live", "-o", binary, "./ai/live")
	cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native build: %v %s", err, regionalSafeOutput(raw))
	}
	return binary
}
