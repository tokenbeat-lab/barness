package e2e

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure inventory: source supplements authenticate an old measurement only.
// Missing, changed, linked or unrelated sources cannot replace the grant pin;
// coherent raw/grant tampering and changed inputs after consumption still fail.
func TestGenericHistoricalSourceSupplement(t *testing.T) {
	root, _ := filepath.Abs("../..")
	binary := filepath.Join(t.TempDir(), "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, raw)
	}
	for _, mode := range []string{"valid", "none", "wrong-source", "malformed", "missing", "symlink", "conflicting-inner", "product-drift", "raw-tamper", "grant-tamper", "input-drift"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "VS55-source-supplement-"+mode)
			dir, record, gitDir, witness := genericSourceSupplementFixture(t, root, ev, mode == "product-drift")
			if mode == "wrong-source" {
				mustWriteXAINative(t, witness, []byte(`{"ai/call.go":"unrelated"}`))
			} else if mode == "malformed" {
				mustWriteXAINative(t, witness, []byte("{"))
			} else if mode == "missing" {
				witness += ".absent"
			} else if mode == "symlink" {
				link := witness + ".link"
				if err := os.Symlink(witness, link); err != nil {
					t.Fatal(err)
				}
				witness = link
			} else if mode == "conflicting-inner" {
				source, err := supportmatrix.RegionalExecutionSource()
				if err != nil {
					t.Fatal(err)
				}
				writeReleaseJSON(t, dir, "execution-source.json", map[string]string{"go.mod": source["go.mod"]})
			} else if mode == "raw-tamper" {
				r, err := supportmatrix.LoadReport(filepath.Join(dir, "live-report.json"))
				if err != nil {
					t.Fatal(err)
				}
				mustWriteXAINative(t, filepath.Join(dir, r.Scenarios[0].CaseID, "attempt-1-result.json"), []byte(`{}`))
			} else if mode == "grant-tamper" {
				editXAINativeJSON(t, dir, "process-grant.json", func(v map[string]any) { v["accountAlias"] = "different-account" })
			}
			before, err := liveSupplementInventory(t, dir)
			if err != nil {
				t.Fatal(err)
			}
			// The launcher gives only this synthetic child its own authority store.
			// The ordinary collector itself receives no Git/credential overrides.
			launcher := filepath.Join(t.TempDir(), "releasegate")
			script := fmt.Sprintf("#!/usr/bin/env python3\nimport os,sys\nenv=dict(os.environ,GIT_DIR=%q,GIT_WORK_TREE=%q)\nos.execve(%q,[%q]+sys.argv[1:],env)\n", gitDir, root, binary, binary)
			if err := os.WriteFile(launcher, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(filepath.Dir(record), "source-consumer")
			rel, _ := filepath.Rel(root, dir)
			outRel, _ := filepath.Rel(root, out)
			args := []string{"tools/verification/consume_live.py", "--binary", launcher, "--bundle", rel, "--out", outRel}
			if mode != "none" {
				sourceRel, _ := filepath.Rel(root, witness)
				args = append(args, "--execution-source", sourceRel)
			}
			cmd := exec.CommandContext(ctxFor(t), "python3", args...)
			cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
			raw, processErr := cmd.CombinedOutput()
			exit := cmd.ProcessState.ExitCode()
			want := mode == "valid" || mode == "input-drift"
			ev.Record("consumer-process", map[string]any{"argv": cmd.Args, "exitCode": exit, "output": regionalSafeOutput(raw), "notActualLive": true, "vendorCalls": 0})
			if !ev.Check("explicit source can qualify only the exact intact old scope", (processErr == nil) == want && exit == map[bool]int{true: 0, false: 1}[want], "exit=%d %s", exit, raw) {
				return
			}
			after, err := liveSupplementInventory(t, dir)
			ev.Check("consumer leaves original bytes intact", err == nil && maps.Equal(before, after), "%v", err)
			if !want {
				return
			}
			proofPath := filepath.Join(out, "live-consumer.json")
			var proof release.LiveConsumerReceipt
			proofRaw, err := os.ReadFile(proofPath)
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, proofRaw, &proof)
			sourceRel, _ := filepath.Rel(root, witness)
			sourceRaw, _ := os.ReadFile(witness)
			ev.Check("external original source is an authenticated consumer input", proof.Inputs["@execution-source/"+filepath.ToSlash(sourceRel)] == evidence.Digest(sourceRaw), "%v", proof.Inputs)
			var index release.Record
			indexRaw, _ := os.ReadFile(record)
			mustUnmarshal(t, indexRaw, &index)
			index.Source = proof.Source
			stage := index.Live.Routes["openai-chat"]
			proofRel, _ := filepath.Rel(root, proofPath)
			stage.Consumers = []release.FileRef{{Path: filepath.ToSlash(proofRel), SHA256: evidence.Digest(proofRaw)}}
			index.Live.Routes["openai-chat"] = stage
			writeReleaseJSON(t, filepath.Dir(record), filepath.Base(record), index)
			if mode == "input-drift" {
				mustWriteXAINative(t, witness, append(sourceRaw, ' '))
			}
			gate := exec.CommandContext(ctxFor(t), binary, "-record", record, "-out", filepath.Join(filepath.Dir(record), "release-consumer"))
			gate.Dir, gate.Env = root, append(xaiProcessEnv(t), "GIT_DIR="+gitDir, "GIT_WORK_TREE="+root)
			gateRaw, gateErr := gate.CombinedOutput()
			var report release.RecordReport
			bytes, err := os.ReadFile(filepath.Join(filepath.Dir(record), "release-consumer", "release-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, bytes, &report)
			for _, check := range report.Checks {
				if check.ID == "live" {
					ev.Check("release independently retains the source supplement binding", check.Pass == (mode == "valid"), "%s", check.Problem)
				}
			}
			ev.Check("synthetic source proof cannot publish a release", gateErr != nil && !report.Pass, "%s", gateRaw)
			ev.Fixture("release-report.json", bytes)
		})
	}
}

func genericSourceSupplementFixture(t *testing.T, root string, ev *evidence.Case, productDrift bool) (string, string, string, string) {
	t.Helper()
	dir, record, oldGit := genericUnsupportedNativeFixture(t, root, ev)
	source, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	source["ai/internal/testkit/retired-capture-note.go"] = evidence.Digest([]byte("TEST_ONLY_old_collector"))
	if productDrift {
		source["ai/call.go"] = evidence.Digest([]byte("TEST_ONLY_different_product"))
	}
	witness := filepath.Join(filepath.Dir(record), "old-execution-source.json")
	writeReleaseJSON(t, filepath.Dir(witness), filepath.Base(witness), source)
	var g livegrant.Grant
	raw, _ := os.ReadFile(filepath.Join(dir, "process-grant.json"))
	mustUnmarshal(t, raw, &g)
	g.Scope.SourceSHA256 = supportmatrix.RegionalSourceHash(source)
	raw = mustMarshal(t, g)
	mustWriteXAINative(t, filepath.Join(dir, "process-grant.json"), raw)
	var claim livegrant.Receipt
	claimRaw, _ := os.ReadFile(filepath.Join(dir, "process-authorization-receipt.json"))
	mustUnmarshal(t, claimRaw, &claim)
	claim.GrantSHA256 = evidence.Digest(raw)
	g.ID = ""
	claim.ScopeSHA256 = evidence.Digest(mustMarshal(t, g))
	writeReleaseJSON(t, dir, "process-authorization-receipt.json", claim)
	ns := miniMaxPrivateSourceNamespace(t)
	store := filepath.Join(ns.dir, "barness-live-process-grants")
	if err := os.CopyFS(store, os.DirFS(filepath.Join(oldGit, "barness-live-process-grants"))); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"id-" + evidence.Digest([]byte(claim.GrantID)) + ".json", "scope-" + claim.ScopeSHA256 + ".json"} {
		mustWriteXAINative(t, filepath.Join(store, name), mustMarshal(t, claim))
	}
	editXAINativeJSON(t, dir, "manifest.json", func(v map[string]any) { v["versions"].(map[string]any)["process_build_source"] = g.Scope.SourceSHA256 })
	resealXAINativeManifest(t, dir)
	genericUnsupportedNativeRecord(t, dir, record)
	return dir, record, ns.dir, witness
}

func liveSupplementInventory(t *testing.T, dir string) (map[string]string, error) {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		files[rel] = evidence.Digest(raw)
		return nil
	})
	return files, err
}
