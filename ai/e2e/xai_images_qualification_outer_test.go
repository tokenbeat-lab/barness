package e2e

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func imageQualificationOuter(t *testing.T, f imageQualificationFixture, first map[string]any) {
	t.Helper()
	ev := run.Case(t, "XAIIMG-qualification-outer")
	before, err := liveSupplementInventory(t, f.bundle)
	if err != nil {
		t.Fatal(err)
	}
	inputs := first["inputs"].(map[string]any)
	authority, producer := 0, 0
	for name := range inputs {
		if strings.HasPrefix(name, "@xai-authority/") {
			authority++
		}
		if strings.HasPrefix(name, "@xai-producer/") {
			producer++
		}
	}
	ev.Check("all 45 native, 10 producer and 2 authority originals are bound", len(inputs) == 57 && authority == 2 && producer == 10, "inputs=%d authority=%d producer=%d", len(inputs), authority, producer)
	source := first["source"].(map[string]any)["files"].(map[string]any)
	ev.Check("independent capture facts are a current consumer input", source["ai/e2e/testdata/xai-images/live-contract.json"] == "1b76056460aa2c9711d00860814274b3b156bbcc5c2055bde398c623923520c6", "capture fact input differs")
	second, code := f.consume(t, "repeat")
	ev.Check("qualification projection is deterministic", code == 0 && first["reportSHA256"] == second["reportSHA256"], "repeat exit=%d hash=%v", code, second["reportSHA256"])
	// Only the child receives the private authority store. The outer collector
	// retains its ordinary environment and cannot access a real approval.
	launcher := filepath.Join(t.TempDir(), "releasegate")
	script := fmt.Sprintf("#!/usr/bin/env python3\nimport os,sys\nenv=dict(os.environ,GIT_DIR=%q,GIT_WORK_TREE=%q)\nos.execve(%q,[%q]+sys.argv[1:],env)\n", os.Getenv("GIT_DIR"), f.root, f.binary, f.binary)
	if err := os.WriteFile(launcher, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Dir(filepath.Dir(run.Dir()))
	out := filepath.Join(outside, fmt.Sprintf("xai-image-consumer-%d", time.Now().UnixNano()))
	bundleRel, _ := filepath.Rel(f.root, f.bundle)
	outRel, _ := filepath.Rel(f.root, out)
	cmd := exec.CommandContext(f.ctx, "python3", "tools/verification/consume_live.py", "--binary", launcher, "--bundle", filepath.ToSlash(bundleRel), "--out", filepath.ToSlash(outRel))
	cmd.Dir, cmd.Env = f.root, xaiProcessEnv(t)
	raw, err := cmd.CombinedOutput()
	code = imagineProcessExit(err)
	ev.Record("outer-process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(raw), "vendorCalls": 0, "notActualLive": true})
	proof := readXAINativeObject(t, filepath.Join(out, "live-consumer.json"))
	ev.Check("outer captures one actual successful native consumer", code == 0 && proof["schema"] == float64(2) && proof["pass"] == true && proof["status"] == "PASS" && proof["exitCode"] == float64(0), "exit=%d status=%v pass=%v", code, proof["status"], proof["pass"])
	ev.Check("outer binds same facts and deterministic projection", proof["reportSHA256"] == first["reportSHA256"] && len(proof["inputs"].(map[string]any)) == 57, "outer report or inputs differ")
	after, err := liveSupplementInventory(t, f.bundle)
	native := readXAINativeObject(t, filepath.Join(f.bundle, "xai-images-candidate-report.json"))
	ev.Check("original candidate and every native byte remain intact", err == nil && maps.Equal(before, after) && native["formalQualified"] == false && native["stage"] == "candidate" && native["mode"] == "live", "original native evidence changed: %v", err)
}
