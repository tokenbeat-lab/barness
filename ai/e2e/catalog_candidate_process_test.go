package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// The external Go executable seam supplies a bounded blocked process. The
// fixed production command is unchanged; these failures never qualify models.
func TestCatalogCandidateProcessCancellation(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CATALOG-CHECK-process-"+mode)
			binary := buildCatalogCLI(t, ev)
			if binary == "" {
				return
			}
			root := catalogSnapshotRoot(t)
			store := filepath.Join(t.TempDir(), "store")
			out, exit := invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "controlled cancellation", "--confirm")
			var frozen struct {
				Candidate maintenance.Candidate `json:"candidate"`
			}
			mustUnmarshal(t, out, &frozen)
			if !ev.Check("candidate frozen", exit == 0, "%s", out) {
				return
			}
			actualGo, err := exec.LookPath("go")
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			gate := filepath.Join(directory, "ready")
			fifo := filepath.Join(directory, "blocked")
			if err := syscall.Mkfifo(fifo, 0600); err != nil {
				t.Fatal(err)
			}
			quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
			script := []byte("#!/bin/sh\nif [ \"$1\" = env ]; then exec " + quote(actualGo) + " \"$@\"; fi\nprintf ready > " + quote(gate) + "\nexec /bin/cat " + quote(fifo) + "\n")
			if err := os.WriteFile(filepath.Join(directory, "go"), script, 0700); err != nil {
				t.Fatal(err)
			}
			ev.Fixture("controlled-blocking-go.sh", script)
			budget := "30s"
			if mode == "timeout" {
				budget = "1s"
			}
			command := exec.Command(binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--timeout", budget, "--confirm")
			command.Dir = root
			command.Env = append(os.Environ(), "PATH="+directory+string(os.PathListSeparator)+os.Getenv("PATH"))
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- command.Wait() }()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			deadline := time.NewTimer(15 * time.Second)
			defer deadline.Stop()
		ready:
			for {
				select {
				case err := <-done:
					t.Fatalf("process ended before controlled gate: %v %s %s", err, stdout.String(), stderr.String())
				case <-deadline.C:
					_ = command.Process.Signal(syscall.SIGTERM)
					<-done
					t.Fatal("controlled gate did not open")
				case <-ticker.C:
					if _, err := os.Stat(gate); err == nil {
						break ready
					}
				}
			}
			started := time.Now()
			forced := false
			if mode == "cancel" {
				if err := command.Process.Signal(syscall.SIGTERM); err != nil {
					t.Fatal(err)
				}
			}
			var processErr error
			select {
			case processErr = <-done:
			case <-time.After(8 * time.Second):
				forced = true
				_ = command.Process.Signal(syscall.SIGTERM)
				select {
				case processErr = <-done:
				case <-time.After(10 * time.Second):
					_ = command.Process.Kill()
					processErr = <-done
				}
			}
			exit = commandExit(processErr)
			ev.Record("process", map[string]any{"argv": command.Args, "exitCode": exit, "stdout": stdout.String(), "stderr": stderr.String(), "forcedAfterDeadline": forced, "elapsedAfterGateSeconds": time.Since(started).Seconds()})
			ev.Check("bounded selected cancellation/timeout", !forced && exit != 0, "forced=%t exit=%d", forced, exit)
			var response struct {
				Category string                  `json:"category"`
				Result   maintenance.CheckResult `json:"result"`
			}
			_ = json.Unmarshal(stdout.Bytes(), &response)
			ev.Check("failure has stable execution state and retained reference", response.Category == "execution_failed" && response.Result.Reference != "" && response.Result.Receipt != "", "%s", stdout.String())
			_, err = os.Stat(filepath.Join(root, response.Result.Reference))
			ev.Check("failure check original retained", err == nil, "%v", err)
			_, err = os.Stat(filepath.Join(root, response.Result.Receipt))
			ev.Check("actual cancelled process receipt retained", err == nil, "%v", err)
			_, err = os.Stat(filepath.Join(store, "current.json"))
			ev.Check("cancelled check never advances effective revision", os.IsNotExist(err), "%v", err)
		})
	}
}

func TestCatalogCandidateCurrentPathFailure(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-current-path-failure")
	binary, store, candidate, check := frozenCheckedCatalog(t, ev)
	root := catalogSnapshotRoot(t)
	output := filepath.Join(t.TempDir(), "runtime.json")
	old := []byte("unchanged output\n")
	if err := os.WriteFile(output, old, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(store, "current.json"), 0700); err != nil {
		t.Fatal(err)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--out", output, "--confirm")
	ev.Check("invalid current path refused before output installation", exit != 0, "%s", out)
	var failed struct {
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	_ = json.Unmarshal(out, &failed)
	ev.Check("failed apply never reports an applied result", failed.Result.Status != "applied", "%s", out)
	raw, err := os.ReadFile(output)
	ev.Check("failed current path leaves prior output unchanged", err == nil && bytes.Equal(raw, old), "%v %s", err, raw)
	info, err := os.Stat(filepath.Join(store, "current.json"))
	ev.Check("failed path never becomes a current pointer", err == nil && info.IsDir(), "%v", err)
}
