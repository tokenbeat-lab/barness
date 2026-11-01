package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These are explicitly labelled derived consumer fixtures. Native originals
// remain unchanged. Updating the fixture inventories makes the assertions
// sensitive to actual missing child/identity coverage beyond a hash mismatch.
func TestCatalogCandidateEvidenceRefusals(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-native-evidence-refusals")
	_, _, _, check := frozenCheckedCatalog(t, ev)
	root := catalogSnapshotRoot(t)
	original := filepath.Dir(filepath.Join(root, check.Receipt))
	for _, mutation := range []string{"intact", "parent-only", "missing-plan-identity", "missing-contract", "bad-exit", "missing-artifact"} {
		dir, err := os.MkdirTemp(filepath.Join(root, ".evidence", "barness-ai-dynamic-catalog"), "consumer-fixture-"+mutation+"-")
		if err != nil {
			t.Fatal(err)
		}
		copied := filepath.Join(dir, "process")
		catalogCopyTree(t, original, copied)
		catalogWriteFixtureJSON(t, filepath.Join(dir, "fixture-origin.json"), map[string]any{"purpose": "derived negative consumer fixture; not an actual stage", "mutation": mutation, "sourceReceipt": check.Receipt, "sourceSHA256": check.ReceiptSHA256})
		var receipt map[string]any
		catalogReadFixtureJSON(t, filepath.Join(copied, "receipt.json"), &receipt)
		bundles := receipt["bundles"].([]any)
		bundleRef := bundles[0].(map[string]any)
		bundle := filepath.Join(copied, bundleRef["path"].(string))
		var manifest map[string]any
		catalogReadFixtureJSON(t, filepath.Join(bundle, "manifest.json"), &manifest)
		switch mutation {
		case "parent-only":
			event := map[string]any{"Action": "pass", "Package": "github.com/tokenbeat-lab/barness/ai/e2e", "Test": "TestCatalogCandidateContracts"}
			catalogWriteFixtureJSON(t, filepath.Join(copied, "stdout.jsonl"), event)
			receipt["files"].(map[string]any)["stdout.jsonl"] = catalogFileSHA(t, filepath.Join(copied, "stdout.jsonl"))
		case "missing-plan-identity":
			for _, item := range manifest["cases"].([]any) {
				c := item.(map[string]any)
				if c["id"] != "CATALOG-CANDIDATE-PLAN" {
					continue
				}
				planPath := filepath.Join(bundle, c["dir"].(string), "plan.json")
				var plan map[string]any
				catalogReadFixtureJSON(t, planPath, &plan)
				cases := plan["cases"].([]any)
				plan["cases"] = cases[1:]
				catalogWriteFixtureJSON(t, planPath, plan)
				c["artifacts"].(map[string]any)["plan.json"] = catalogFileSHA(t, planPath)
			}
			catalogWriteFixtureJSON(t, filepath.Join(bundle, "manifest.json"), manifest)
			bundleRef["manifestSHA256"] = catalogFileSHA(t, filepath.Join(bundle, "manifest.json"))
		case "missing-contract":
			cases := manifest["cases"].([]any)
			for n, item := range cases {
				if strings.HasPrefix(item.(map[string]any)["id"].(string), "CC-") {
					manifest["cases"] = append(cases[:n], cases[n+1:]...)
					break
				}
			}
			catalogWriteFixtureJSON(t, filepath.Join(bundle, "manifest.json"), manifest)
			bundleRef["manifestSHA256"] = catalogFileSHA(t, filepath.Join(bundle, "manifest.json"))
		case "bad-exit":
			if err := os.WriteFile(filepath.Join(copied, "exit-code"), []byte("1\n"), 0600); err != nil {
				t.Fatal(err)
			}
			receipt["files"].(map[string]any)["exit-code"] = catalogFileSHA(t, filepath.Join(copied, "exit-code"))
		case "missing-artifact":
			for _, item := range manifest["cases"].([]any) {
				c := item.(map[string]any)
				if !strings.HasPrefix(c["id"].(string), "CC-") {
					continue
				}
				if err := os.Remove(filepath.Join(bundle, c["dir"].(string), "identity.json")); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		catalogWriteFixtureJSON(t, filepath.Join(copied, "receipt.json"), receipt)
		relative, _ := filepath.Rel(root, filepath.Join(copied, "receipt.json"))
		args := []string{"tools/verification/catalog_check.py", "inspect", "--prepared", check.Prepared, "--receipt", filepath.ToSlash(relative)}
		command := exec.CommandContext(ctxFor(t), "python3", args...)
		command.Dir = root
		var stderr bytes.Buffer
		command.Stderr = &stderr
		out, err := command.Output()
		exit := commandExit(err)
		ev.Record("process-"+mutation, map[string]any{"argv": command.Args, "exitCode": exit, "stdout": string(out), "stderr": stderr.String()})
		if mutation == "intact" {
			ev.Check("intact actual receipt remains consumable", exit == 0, "%s", out)
		} else {
			ev.Check("incomplete native facts refused "+mutation, exit != 0, "exit=%d %s", exit, out)
		}
	}
}
func catalogReadFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatal(err)
	}
}
func catalogWriteFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw := mustMarshal(t, value)
	if err := os.WriteFile(path, append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
func catalogFileSHA(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
