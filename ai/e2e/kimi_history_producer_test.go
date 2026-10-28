package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Keep the original controlled receipt and restore it after each semantic
// mutation. Re-pinning a bad receipt must not bypass actual producer checks.
func kimiProducerFailures(t *testing.T, ev *evidence.Case, descriptor map[string]any, path, baseline string, check func(string, string) ([]byte, error)) {
	t.Helper()
	receipt := filepath.Join(filepath.Dir(baseline), "producer-receipt.json")
	prior, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatal(err)
	}
	defer os.WriteFile(receipt, prior, 0600)
	for _, kind := range []string{"entry", "count", "timeout", "duplicate-entry", "unknown-arg", "exit", "missing-exit", "binary", "bundle", "source"} {
		var bad, composed map[string]any
		mustUnmarshal(t, prior, &bad)
		mustUnmarshal(t, mustMarshal(t, descriptor), &composed)
		switch kind {
		case "entry":
			bad["argv"].([]any)[1] = "-test.run=^TestKimiHistorySupplement$"
		case "count":
			bad["argv"] = append(bad["argv"].([]any), "-test.count=2")
		case "timeout":
			bad["argv"].([]any)[2] = "-test.timeout=1s"
		case "duplicate-entry":
			bad["argv"] = append(bad["argv"].([]any), "-test.run=^TestKimiBaselineControlled$")
		case "unknown-arg":
			bad["argv"] = append(bad["argv"].([]any), "-test.bench=.")
		case "exit":
			bad["exitCode"] = 1
		case "missing-exit":
			delete(bad, "exitCode")
		case "binary":
			bad["argv"].([]any)[0] = repoRelative(t, receipt)
		case "bundle":
			bad["bundles"] = []any{}
		case "source":
			bad["source"].(map[string]any)["commit"] = "other"
		}
		raw := mustMarshal(t, bad)
		if err := os.WriteFile(receipt, raw, 0600); err != nil {
			t.Fatal(err)
		}
		composed["baseline"].(map[string]any)["producer"].(map[string]any)["sha256"] = evidence.Digest(raw)
		if err := os.WriteFile(path, mustMarshal(t, composed), 0600); err != nil {
			t.Fatal(err)
		}
		out, err := check("controlled", "candidate")
		ev.Check("reject re-pinned producer "+kind, err != nil, "%s", regionalSafeOutput(out))
	}
	if err := os.WriteFile(receipt, prior, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, mustMarshal(t, descriptor), 0600); err != nil {
		t.Fatal(err)
	}
}
