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
	base := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-tags", "live", "-count=1", "-timeout=180s", "./ai/live", "-run", "^TestMiniMax")
	cmd.Dir = "../.."
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
	paths, e := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
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
