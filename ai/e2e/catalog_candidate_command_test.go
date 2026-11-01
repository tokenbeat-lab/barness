package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Coherent derived fixtures keep the actual native PASS artifacts. Every
// enclosing argv and changed inventory is updated, so a wrong entry cannot
// be refused merely because command/prepared/receipt disagree with each other.
func TestCatalogCandidateFixedEntryReceipt(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-fixed-entry-receipt")
	binary := buildSingleModelCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	candidate := freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "fixed-entry consumer fixture")
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var check maintenance.CheckResult
	mustUnmarshal(t, out, &check)
	if !ev.Check("actual fixed empty contract is consumable", exit == 0 && check.Status == "applicable", "exit=%d %s", exit, out) {
		return
	}
	for _, mutation := range []string{"intact", "wrong-regex", "wrong-command", "unbounded-timeout", "unsupported-schema"} {
		prepared, receipt := catalogCommandFixture(t, check, mutation)
		cmd := exec.CommandContext(ctxFor(t), "python3", "tools/verification/catalog_check.py", "inspect", "--prepared", prepared, "--receipt", receipt)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		exit := commandExit(err)
		ev.Record("consumer-"+mutation, map[string]any{"argv": cmd.Args, "exitCode": exit, "output": string(out)})
		if mutation == "intact" {
			if !ev.Check("coherent relocated fixture passes before negative mutations", exit == 0, "exit=%d %s", exit, out) {
				return
			}
		} else {
			ev.Check("coherent wrong fixed entry refused "+mutation, exit != 0, "exit=%d %s", exit, out)
		}
	}
}

func catalogCommandFixture(t *testing.T, check maintenance.CheckResult, mutation string) (string, string) {
	t.Helper()
	root := catalogSnapshotRoot(t)
	directory, err := os.MkdirTemp(filepath.Join(root, ".evidence", "barness-ai-dynamic-catalog"), "consumer-command-fixture-"+mutation+"-")
	if err != nil {
		t.Fatal(err)
	}
	inputs := filepath.Join(directory, "inputs")
	process := filepath.Join(directory, "process")
	catalogCopyTree(t, filepath.Dir(filepath.Join(root, check.Prepared)), inputs)
	catalogCopyTree(t, filepath.Dir(filepath.Join(root, check.Receipt)), process)
	catalogWriteFixtureJSON(t, filepath.Join(directory, "fixture-origin.json"), map[string]any{"purpose": "coherent derived command consumer fixture; not an actual stage", "mutation": mutation, "sourcePrepared": check.Prepared, "sourceReceipt": check.Receipt, "sourceReceiptSHA256": check.ReceiptSHA256})
	var prepared, receipt, command map[string]any
	catalogReadFixtureJSON(t, filepath.Join(inputs, "prepared.json"), &prepared)
	catalogReadFixtureJSON(t, filepath.Join(process, "receipt.json"), &receipt)
	catalogReadFixtureJSON(t, filepath.Join(process, "command.json"), &command)
	relative, _ := filepath.Rel(root, inputs)
	relative = filepath.ToSlash(relative)
	prepared["directory"] = relative
	argv := prepared["argv"].([]any)
	argv[len(argv)-1] = relative + "/snapshot.json"
	switch mutation {
	case "wrong-regex":
		argv[5] = ".*"
	case "wrong-command":
		argv[0] = "echo"
	case "unbounded-timeout":
		prepared["timeout"] = 1201
		argv[7] = "-timeout=1201s"
		command["timeoutSeconds"] = 1201
	case "unsupported-schema":
		prepared["schema"] = 2
	}
	prepared["argv"], receipt["argv"], command["argv"] = argv, argv, argv
	catalogWriteFixtureJSON(t, filepath.Join(inputs, "prepared.json"), prepared)
	selected := map[string]any{}
	for name, digest := range prepared["extra"].(map[string]any) {
		selected[name] = digest
	}
	for name, digest := range prepared["frozen"].(map[string]any) {
		selected[relative+"/"+name] = digest
	}
	selected[relative+"/prepared.json"] = catalogFileSHA(t, filepath.Join(inputs, "prepared.json"))
	receipt["selectedInputs"] = selected
	catalogWriteFixtureJSON(t, filepath.Join(process, "command.json"), command)
	commandSHA := catalogFileSHA(t, filepath.Join(process, "command.json"))
	receipt["commandSHA256"] = commandSHA
	receipt["files"].(map[string]any)["command.json"] = commandSHA
	catalogWriteFixtureJSON(t, filepath.Join(process, "receipt.json"), receipt)
	receiptName, _ := filepath.Rel(root, filepath.Join(process, "receipt.json"))
	return relative + "/prepared.json", filepath.ToSlash(receiptName)
}
