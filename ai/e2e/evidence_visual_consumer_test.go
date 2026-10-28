package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// A consumer must reach the independent verifier with the original catalog
// identity. A changed global hash alone cannot preempt the selected-model gate.
func TestEvidenceVisualConsumerKeepsOriginalCatalog(t *testing.T) {
	ev := run.Case(t, "VS55-visual-original-catalog")
	base := t.TempDir()
	marker := filepath.Join(base, "verifier-build-invoked")
	goCommand := filepath.Join(base, "go")
	if err := os.WriteFile(goCommand, []byte("#!/bin/sh\ntouch '"+marker+"'\nexit 77\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", base+string(os.PathListSeparator)+os.Getenv("PATH"))
	writeReleaseJSON(t, base, "report.json", map[string]any{"schema": 1, "route": "jev", "model": "jev-1.14.0", "catalogKind": "formal", "catalogVersion": "TEST_ONLY_original-version", "catalogHash": "sha256:TEST_ONLY_original"})
	checks := release.VisualEffectsEvidence(base, "")
	_, err := os.Stat(marker)
	ev.Check("independent verifier owns catalog applicability", err == nil && checks[0].Status == release.Fail && checks[0].Detail == "visual verifier build failed", "%v %+v", err, checks[0])
	ev.Record("process-boundary", map[string]any{"verifierBuildInvoked": err == nil, "vendorCalls": 0, "qualification": "FAIL"})
}
