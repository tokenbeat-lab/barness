package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLIExport(t *testing.T) {
	ev := run.Case(t, "CATALOG-T01-cli-approved-export")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "absent")
	out, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	expected, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("approved snapshot export retains exact bytes", exit == 0 && bytes.Equal(out, expected), "exit=%d bytes=%d", exit, len(out))
	state, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--state")
	ev.Check("approved source state can be exported", exit == 0 && bytes.Contains(state, []byte(`"rulesRevision"`)), "exit=%d output=%s", exit, state)
	_, err = os.Stat(store)
	ev.Check("stdout export creates no store", os.IsNotExist(err), "err=%v", err)
	statePath := filepath.Join(dir, "state.json")
	if err := os.WriteFile(statePath, state, 0600); err != nil {
		t.Fatal(err)
	}
	generator := filepath.Join(dir, "cataloggen")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", generator, "./cmd/cataloggen")
	cmd.Dir = "../.."
	build, err := cmd.CombinedOutput()
	ev.Record("generator-build", map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(build)})
	if !ev.Check("generator compiles", err == nil, "err=%v output=%s", err, build) {
		return
	}
	output := filepath.Join(dir, "generated.json")
	if err := os.WriteFile(output, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	out, exit = invokeCatalogCLI(t, ev, generator, "--state", statePath, "--out", output)
	after, _ := os.ReadFile(output)
	ev.Check("generator refuses unchecked overwrite", exit != 0 && string(after) == "original", "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, generator, "--state", statePath, "--out", output, "--dry-run")
	after, _ = os.ReadFile(output)
	ev.Check("generator dry-run is read-only", exit == 0 && string(after) == "original", "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, generator, "--state", statePath, "--out", output, "--confirm")
	after, _ = os.ReadFile(output)
	ev.Check("selected exported state reproduces builtin bytes", exit == 0 && bytes.Equal(after, expected), "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, generator, "--check", "--state", statePath, "--out", output)
	ev.Check("check verifies without rewriting", exit == 0, "exit=%d output=%s", exit, out)
	bad := filepath.Join(dir, "parent-file")
	os.WriteFile(bad, []byte("keep"), 0600)
	out, exit = invokeCatalogCLI(t, ev, generator, "--state", statePath, "--out", filepath.Join(bad, "catalog.json"), "--confirm")
	after, _ = os.ReadFile(bad)
	ev.Check("generator failure does not truncate existing paths", exit != 0 && string(after) == "keep", "exit=%d output=%s", exit, out)
}
