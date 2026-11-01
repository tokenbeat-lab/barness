package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCatalogCandidateCheckReadOnly(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-read-only-cli")
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "catalogctl")
	build := exec.Command("go", "build", "-o", binary, "./cmd/catalogctl")
	build.Dir = root
	output, err := build.CombinedOutput()
	ev.Record("build", map[string]any{"output": string(output), "error": errString(err)})
	ev.Check("compiled CLI available", err == nil, "build: %s %v", output, err)
	if err != nil {
		return
	}
	store := filepath.Join(t.TempDir(), "absent-store")
	evidence := "catalog-check-read-only-no-output"
	command := exec.Command(binary, "check", "--help", "--store", store, "--evidence-feature", evidence)
	command.Dir = root
	output, err = command.Output()
	ev.Record("help", map[string]any{"output": string(output), "error": errString(err)})
	var result map[string]any
	ev.Check("help succeeds", err == nil, "%s %v", output, err)
	ev.Check("help JSON", json.Unmarshal(output, &result) == nil && result["status"] == "help", "%s", output)
	_, err = os.Stat(store)
	ev.Check("help does not create store", os.IsNotExist(err), "%v", err)
	_, err = os.Stat(filepath.Join(root, ".evidence", evidence))
	ev.Check("help does not create evidence", os.IsNotExist(err), "%v", err)
}
