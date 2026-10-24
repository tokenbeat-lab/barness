//go:build live

package live

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The existing public producer uses only its explicit controlled transport.
// This process boundary checks the owned verifier and its native source pin.
func TestJevOwnedProtocolVerifierProcess(t *testing.T) {
	cs := run.Case(t, "UC14-owned-protocol-verifier")
	protocol := formalProtocolBundle(t, cs)
	if protocol == "" {
		return
	}
	raw, err := os.ReadFile(filepath.Join(protocol, "multimodal-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report multimodalExecution
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	const source = "../internal/testkit/jevformal/verify_protocol.py"
	verifier := "../internal/testkit/jevformal/verify_protocol.py"
	verifierRaw, readErr := os.ReadFile(verifier)
	cs.Check("owned Python verifier is part of complete native source", readErr == nil && report.SourceHashes[source] == sha256Text(verifierRaw), "owned source unavailable or not sealed")
	catalog := filepath.Join(run.Dir(), "owned-verifier-catalog.json")
	raw, err = json.Marshal(ai.BuiltinCatalog())
	if err == nil {
		err = os.WriteFile(catalog, raw, 0600)
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, requireLive := range []bool{false, true} {
		args := []string{verifier, "--formal-catalog", catalog}
		if requireLive {
			args = append(args, "--require-live")
		}
		args = append(args, protocol)
		cmd := exec.Command("python3", args...)
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "PYTHONDONTWRITEBYTECODE=1"}
		output, err := cmd.CombinedOutput()
		cs.Record("owned-verifier-process", map[string]any{"requireLive": requireLive, "failed": err != nil, "output": string(output)})
		cs.Check("owned verifier accepts controlled structure and refuses real qualification", (err != nil) == requireLive, "requireLive=%v, %s", requireLive, output)
	}
}
