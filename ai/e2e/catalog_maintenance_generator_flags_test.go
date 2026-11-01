package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceGeneratorFlagErrors(t *testing.T) {
	ev := run.Case(t, "CATALOG-T04-generator-structured-parse-errors")
	dir := t.TempDir()
	binary := filepath.Join(dir, "cataloggen")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./cmd/cataloggen")
	cmd.Dir = "../.."
	build, err := cmd.CombinedOutput()
	ev.Record("build", map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(build)})
	if !ev.Check("generator compiles", err == nil, "%v %s", err, build) {
		return
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	ev.Record("binary", map[string]any{"sha256": hex.EncodeToString(sum[:]), "bytes": len(raw)})
	output := filepath.Join(dir, "output.json")
	os.WriteFile(output, []byte("keep"), 0600)
	for _, args := range [][]string{{"--unknown"}, {"--out"}, {"--state"}, {"--out", output, "--confirm", "--builtin", "--unknown"}} {
		out, exit := invokeCatalogCLI(t, ev, binary, args...)
		var result struct{ Status, Category string }
		parseErr := json.Unmarshal(out, &result)
		ev.Check("parse errors use structured generation failure", exit != 0 && parseErr == nil && result.Status == "error" && result.Category == "generation_failed", "args=%v exit=%d output=%s", args, exit, out)
		after, _ := os.ReadFile(output)
		ev.Check("parse failure does not write selected output", bytes.Equal(after, []byte("keep")), "args=%v", args)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "--out", output, "--confirm", "--builtin", "--help")
	after, _ := os.ReadFile(output)
	ev.Check("help succeeds without changing output", exit == 0 && bytes.Equal(after, []byte("keep")), "exit=%d output=%s", exit, out)
}
