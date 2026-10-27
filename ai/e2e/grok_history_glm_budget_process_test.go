package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Test the original public wrapper through a private Go proxy, rather than
// asserting timeout constants. Native assertions and inventory stay unchanged.
func TestGrokHistoryGLMProcessBudget(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "e2e.test")
	build := exec.Command(realGo, "test", "-c", "-o", binary, "./ai/e2e")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("compile public test entry: %v %s", e, output)
	}
	tool := exec.Command(realGo, "tool", "-n", "test2json")
	tool.Dir, tool.Env = root, xaiProcessEnv(t)
	output, err := tool.Output()
	if err != nil {
		t.Fatalf("locate test2json: %v", err)
	}
	test2json := strings.TrimSpace(string(output))
	for _, mode := range []string{"slow-start", "cancel-tree"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-GLM-BUDGET-"+strings.ToUpper(mode))
			private := t.TempDir()
			proxy, err := os.ReadFile("testdata/glm-process/go_proxy.py")
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(private, "go"), proxy, 0700); err != nil {
				t.Fatal(err)
			}
			config := map[string]string{"mode": mode, "go": realGo, "started": filepath.Join(private, "started"), "heartbeat": filepath.Join(private, "heartbeat"), "pids": filepath.Join(private, "pids.json")}
			if err = os.WriteFile(filepath.Join(private, "config.json"), mustMarshal(t, config), 0600); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				raw, _ := os.ReadFile(config["pids"])
				var pids []int
				_ = json.Unmarshal(raw, &pids)
				for _, pid := range pids {
					if process, e := os.FindProcess(pid); e == nil {
						_ = process.Kill()
					}
				}
			})
			env := xaiProcessEnv(t)
			for i, value := range env {
				if strings.HasPrefix(value, "PATH=") {
					env[i] = "PATH=" + private + string(os.PathListSeparator) + strings.TrimPrefix(value, "PATH=")
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			args := []string{"-p", "github.com/tokenbeat-lab/barness/ai/e2e", "-t", binary, "-test.v=test2json", "-test.run=^TestGrokHistoryGLMControlledFullCapture$", "-test.count=1", "-test.timeout=15s"}
			program, directory := test2json, filepath.Join(root, "ai/e2e")
			if mode == "slow-start" {
				program, directory = "python3", root
				args[len(args)-1] = "-test.timeout=80s"
				// The collector remains at the repo root; only the public Go
				// test executable needs the package working directory.
				args = append([]string{"tools/verification/run.py", "--feature", "barness-ai-grok-glm-budget", "--stage", "e2e", "--timeout", "100", "--", "python3", "-c", "import os,sys; os.chdir('ai/e2e'); os.execv(sys.argv[1], sys.argv[1:])", test2json}, args...)
			} else {
				env = append(env, evidence.DirEnv+"="+filepath.Join(private, "bundles"))
			}
			cmd := exec.CommandContext(ctx, program, args...)
			cmd.Dir, cmd.Env = directory, env
			stdout, err := os.Create(filepath.Join(private, "stdout"))
			if err != nil {
				t.Fatal(err)
			}
			defer stdout.Close()
			stderr, err := os.Create(filepath.Join(private, "stderr"))
			if err != nil {
				t.Fatal(err)
			}
			defer stderr.Close()
			// Files avoid inherited-pipe hangs in this regression's own observer.
			cmd.Stdout, cmd.Stderr = stdout, stderr
			start := time.Now()
			runErr := cmd.Run()
			elapsed := time.Since(start)
			out, _ := os.ReadFile(stdout.Name())
			errOut, _ := os.ReadFile(stderr.Name())
			ev.Record("public-wrapper", map[string]any{"argv": cmd.Args, "exitCode": cmd.ProcessState.ExitCode(), "elapsedSeconds": elapsed.Seconds(), "stdout": string(out), "stderr": string(errOut), "vendorCalls": 0})
			_, startedErr := os.Stat(config["started"])
			ev.Check("proxy intercepted the exact controlled native command", startedErr == nil, "%v", startedErr)
			if mode == "slow-start" {
				if !ev.Check("startup plus actual full controlled suite seals after old ten-second limit", runErr == nil && elapsed > 11*time.Second, "exit=%v elapsed=%s", runErr, elapsed) {
					return
				}
				var result struct{ Attempt string }
				mustUnmarshal(t, out, &result)
				raw, e := os.ReadFile(filepath.Join(root, result.Attempt, "receipt.json"))
				if e != nil {
					t.Fatal(e)
				}
				var receipt release.StageReceipt
				mustUnmarshal(t, raw, &receipt)
				ev.Record("collector", map[string]any{"receipt": result.Attempt + "/receipt.json", "sha256": evidence.Digest(raw), "bundles": receipt.Bundles})
				nativeCases := 0
				for _, bundle := range receipt.Bundles {
					if bundle.Package != "-tags live ./ai/live" {
						continue
					}
					manifest, readErr := os.ReadFile(filepath.Join(root, result.Attempt, bundle.Path, "manifest.json"))
					if readErr != nil {
						t.Fatal(readErr)
					}
					var native struct{ Cases []struct{ ID, Status string } }
					mustUnmarshal(t, manifest, &native)
					nativeCases = len(native.Cases)
					for _, c := range native.Cases {
						ev.Check("native case "+c.ID+" passed", c.Status == "PASS", "%s", c.Status)
					}
				}
				ev.Check("collector validates independent main and full32 auxiliary inventories and audits", receipt.Status == "PASS" && len(receipt.Bundles) == 2 && nativeCases == 32, "bundles=%d nativeCases=%d", len(receipt.Bundles), nativeCases)
				return
			}
			// Preserve the failed public wrapper's own process observation.
			// No native bundle is invented for this cancellation countercase.
			processFiles, _ := filepath.Glob(filepath.Join(private, "bundles", "*", "GROK-HISTORY-GLM-CONTROLLED-FULL", "process.json"))
			var process glmProcessResult
			if len(processFiles) == 1 {
				raw, readErr := os.ReadFile(processFiles[0])
				if readErr == nil {
					_ = json.Unmarshal(raw, &process)
				}
				ev.Record("cancelled-capture", process)
			}
			ev.Check("parent context interrupts the capture owner before its fallback timeout", process.Status == "CANCELLED" && process.ContextError == context.DeadlineExceeded.Error() && process.ExitCode != nil, "%+v", process)
			before, e := os.ReadFile(config["heartbeat"])
			time.Sleep(200 * time.Millisecond)
			after, e2 := os.ReadFile(config["heartbeat"])
			ev.Record("descendant", map[string]any{"beforeBytes": len(before), "afterBytes": len(after), "unchanged": string(before) == string(after)})
			ev.Check("parent deadline rejects incomplete native capture", runErr != nil, "exit=%v", runErr)
			ev.Check("wrapper returns before Go's emergency test timeout", elapsed < 14*time.Second && !strings.Contains(string(out), "test timed out"), "elapsed=%s", elapsed)
			ev.Check("all descendant writers stop before wrapper returns without inherited-pipe hang", e == nil && e2 == nil && len(before) > 0 && string(before) == string(after), "before=%d after=%d errors=%v,%v", len(before), len(after), e, e2)
		})
	}
}
