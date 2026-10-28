package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Failure inventory: the full native declaration must pass both the artifact
// and report gates in the public process. Requiring only rejection would hide
// old all-PASS/positive-call checks. A mandatory NOT_RUN must still fail report
// qualification. Changed assertion reasons, hidden captures,
// and changed inventory bytes must fail before consumer-receipt validation.
// TEST_ONLY authority is statically seeded in an isolated Git namespace; no
// producer or Claim is called, and this cannot qualify outside that namespace.
func TestGenericDeclaredUnsupportedCLI(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build public consumer: %v %s", err, output)
	}
	for _, mutation := range []string{"declared", "mandatory-not-run", "assertion-reason", "hidden-capture", "inventory-bytes", "missing-namespace"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "GROK-GENERIC-UNSUPPORTED-CLI-"+mutation)
			dir, record, gitDir := genericUnsupportedNativeFixture(t, root, ev)
			unsupported := "synthetic-reasoning-history/assertions.json"
			switch mutation {
			case "mandatory-not-run":
				editXAINativeJSON(t, dir, "live-report.json", func(v map[string]any) {
					for _, raw := range v["scenarios"].([]any) {
						s := raw.(map[string]any)
						if s["id"] == grokhistory.Scenario {
							s["outcome"] = "NOT_RUN"
						}
					}
				})
				editXAINativeJSON(t, dir, "manifest.json", func(v map[string]any) {
					for _, raw := range v["cases"].([]any) {
						c := raw.(map[string]any)
						if c["id"] == "synthetic-"+grokhistory.Scenario {
							c["status"] = "NOT_RUN"
						}
					}
				})
				editXAINativeJSON(t, dir, "synthetic-"+grokhistory.Scenario+"/assertions.json", func(v map[string]any) { v["status"] = "NOT_RUN" })
			case "assertion-reason":
				editXAINativeJSON(t, dir, unsupported, func(v map[string]any) { v["unsupported"] = "TEST_ONLY arbitrary unsupported reason" })
			case "hidden-capture":
				writeReleaseJSON(t, filepath.Join(dir, "synthetic-reasoning-history"), "hidden-http.json", map[string]any{"synthetic": true})
			case "inventory-bytes":
				raw, err := os.ReadFile(filepath.Join(dir, unsupported))
				if err != nil {
					t.Fatal(err)
				}
				mustWriteXAINative(t, filepath.Join(dir, unsupported), append(raw, ' '))
			}
			if mutation != "inventory-bytes" {
				resealXAINativeManifest(t, dir)
			}
			genericUnsupportedNativeRecord(t, dir, record)
			audited := release.AuditBundle(dir, nil)
			ev.Record("synthetic-audit", map[string]any{"error": audited.Err, "findings": audited.Findings, "vendorCalls": 0})
			out := filepath.Join(filepath.Dir(record), "public-consumer")
			cmd := exec.CommandContext(ctxFor(t), binary, "-record", record, "-out", out)
			cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
			if mutation == "missing-namespace" {
				gitDir = filepath.Join(t.TempDir(), "absent.git")
			}
			cmd.Env = append(cmd.Env, "GIT_DIR="+gitDir)
			output, processErr := cmd.CombinedOutput()
			exit := -1
			if cmd.ProcessState != nil {
				exit = cmd.ProcessState.ExitCode()
			}
			raw, err := os.ReadFile(filepath.Join(out, "release-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report release.RecordReport
			mustUnmarshal(t, raw, &report)
			problem := ""
			for _, check := range report.Checks {
				if check.ID == "live" {
					problem = check.Problem
				}
			}
			ev.Fixture("public-consumer-report.json", raw)
			binaryRaw, err := os.ReadFile(binary)
			if err != nil {
				t.Fatal(err)
			}
			ev.Record("public-consumer-process", map[string]any{"argv": cmd.Args, "exitCode": exit, "output": string(output), "gitDir": gitDir, "binarySHA256": evidence.Digest(binaryRaw), "notActualLive": true, "vendorCalls": 0})
			ev.Check("synthetic evidence cannot pass total release", processErr != nil && exit == 1 && !report.Pass, "exit=%d pass=%t", exit, report.Pass)
			want := "native evidence audit failed"
			if mutation == "declared" {
				want = "live qualification stale or consumer receipt missing"
			} else if mutation == "mandatory-not-run" {
				want = "generic native capability did not pass"
			}
			ev.Check("public CLI reaches exactly the intended qualification boundary", problem == want, "problem=%s want=%s", problem, want)
			// The persisted fixture retains UNSUPPORTED; acceptance may not rewrite
			// it into support or create a last-passed timestamp.
			live := readXAINativeObject(t, filepath.Join(dir, "live-report.json"))
			for _, raw := range live["scenarios"].([]any) {
				s := raw.(map[string]any)
				if s["id"] == "reasoning-history" {
					ev.Check("native unsupported status and zero activity retained", s["outcome"] == "UNSUPPORTED" && s["calls"] == float64(0) && s["attempts"] == float64(0) && !strings.Contains(string(mustMarshal(t, s)), "lastPassed"), "%+v", s)
				}
			}
		})
	}
}
