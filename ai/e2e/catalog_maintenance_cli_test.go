package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var catalogSubprocessRace = flag.Bool("catalog-subprocess-race", false, "build catalog CLI with the race detector")

// The compiled public CLI is the seam. Its read-only commands must leave even
// a nonexistent store absent; no implementation file layout is asserted.
func TestCatalogMaintenanceCLIReadOnly(t *testing.T) {
	ev := run.Case(t, "CATALOG-T04-cli-read-only")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "absent")
	for _, args := range [][]string{
		{"help"}, {"versions", "--store", store, "--json"},
		{"history", "--store", store, "--json"},
		{"withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "fixture", "--dry-run", "--json"},
	} {
		out, exit := invokeCatalogCLI(t, ev, binary, args...)
		ev.Check("read-only command succeeds", exit == 0, "args=%v exit=%d output=%s", args, exit, out)
		_, err := os.Stat(store)
		ev.Check("read-only command creates no store", os.IsNotExist(err), "args=%v err=%v", args, err)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "fixture", "--json")
	var response map[string]any
	_ = json.Unmarshal(out, &response)
	ev.Check("missing explicit confirmation is structured refusal", exit != 0 && response["category"] == "confirmation_required", "exit=%d output=%s", exit, out)
	_, err := os.Stat(store)
	ev.Check("refusal creates no store", os.IsNotExist(err), "err=%v", err)
}

func buildCatalogCLI(t *testing.T, ev *evidence.Case) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "catalogctl")
	cmd := exec.CommandContext(ctxFor(t), "go", catalogBuildArgs(binary, "./cmd/catalogctl")...)
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	exit := commandExit(err)
	ev.Record("build", map[string]any{"argv": cmd.Args, "exitCode": exit, "output": string(out)})
	if !ev.Check("compiled public CLI is available", err == nil, "exit=%d output=%s", exit, out) {
		return ""
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	info, infoErr := buildinfo.ReadFile(binary)
	ev.Check("binary build information available", infoErr == nil, "%v", infoErr)
	if infoErr == nil {
		ev.Record("binary-buildinfo", info)
	}
	ev.Record("binary", map[string]any{"sha256": hex.EncodeToString(sum[:]), "bytes": len(raw)})
	return binary
}

func catalogBuildArgs(binary, pkg string) []string {
	args := []string{"build"}
	if *catalogSubprocessRace {
		args = append(args, "-race")
	}
	return append(args, "-o", binary, pkg)
}

var catalogProcessSequence atomic.Uint64

func invokeCatalogCLI(t *testing.T, ev *evidence.Case, binary string, args ...string) ([]byte, int) {
	t.Helper()
	// Candidate checks can run nested Go builds; keep their test-process budget
	// bounded without changing the product check timeout or cancellation tests.
	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = "../.."
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	started := time.Now()
	out, err := cmd.Output()
	elapsed := time.Since(started)
	exit := commandExit(err)
	// Artifact names use a digest of the actual argv; repeated calls preserve
	// both observations through separate cases/scenarios in the caller.
	sum := sha256.Sum256([]byte(cmd.String()))
	name := fmt.Sprintf("process-%03d-%s", catalogProcessSequence.Add(1), hex.EncodeToString(sum[:8]))
	category := "none"
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			category = "exit"
		} else {
			category = "start_or_wait"
		}
	}
	var nativeExit any
	signal := ""
	if cmd.ProcessState != nil {
		if code := cmd.ProcessState.ExitCode(); code >= 0 {
			nativeExit = code
		}
		if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			signal = status.Signal().String()
		}
	}
	ev.Record(name, map[string]any{"argv": cmd.Args, "commandExit": exit, "nativeExitCode": nativeExit, "started": cmd.Process != nil, "processState": cmd.ProcessState != nil, "signal": signal, "elapsedSeconds": elapsed.Seconds(), "timeoutSeconds": timeout.Seconds(), "contextError": errString(ctx.Err()), "execError": errString(err), "execErrorCategory": category, "stdout": string(out), "stderr": stderr.String()})
	return out, exit
}

func commandExit(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(*exec.ExitError); ok {
		return e.ExitCode()
	}
	return -1
}

func TestCatalogMaintenanceCLIFreeze(t *testing.T) {
	ev := run.Case(t, "CATALOG-T06-cli-freeze-immutable")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	seedVersions := catalogSeedVersions(t, ev, binary, store)
	args := []string{"withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "withdraw fixture", "--confirm", "--json"}
	out, exit := invokeCatalogCLI(t, ev, binary, args...)
	var frozen struct {
		Changes []struct {
			Field         string `json:"field"`
			Before, After json.RawMessage
		} `json:"changes"`
		Status    string `json:"status"`
		Candidate struct {
			ID       string `json:"id"`
			Snapshot []byte `json:"snapshot"`
			State    struct {
				Providers []struct {
					ID      string `json:"id"`
					Version uint64 `json:"version"`
				} `json:"providers"`
			} `json:"state"`
		} `json:"candidate"`
	}
	_ = json.Unmarshal(out, &frozen)
	if !ev.Check("confirmed mutation freezes a candidate", exit == 0 && frozen.Status == "awaiting_check" && frozen.Candidate.ID != "", "exit=%d output=%s", exit, out) {
		return
	}
	ev.Check("preview includes the explicit lifecycle diff", len(frozen.Changes) == 1 && frozen.Changes[0].Field == "/lifecycle" && string(frozen.Changes[0].Before) == `"active"` && string(frozen.Changes[0].After) == `"withdrawn"`, "changes=%+v", frozen.Changes)
	var snapshot struct {
		Catalog struct {
			Models []struct{ Provider, API, ID string } `json:"models"`
		} `json:"catalog"`
	}
	_ = json.Unmarshal(frozen.Candidate.Snapshot, &snapshot)
	found := false
	for _, m := range snapshot.Catalog.Models {
		found = found || (m.Provider == "openai" && m.API == "openai-responses" && m.ID == "gpt-6.1-sol")
	}
	ev.Check("candidate alone omits explicitly withdrawn identity", !found && len(snapshot.Catalog.Models) == len(ai.BuiltinCatalog().Models)-1, "count=%d found=%v", len(snapshot.Catalog.Models), found)
	for _, p := range frozen.Candidate.State.Providers {
		expected := seedVersions[p.ID]
		if p.ID == "openai" {
			expected++
		}
		ev.Check("only changed Provider proposes next version", p.Version == expected, "%s=%d", p.ID, p.Version)
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "versions", "--store", store)
	var current struct {
		Versions map[string]uint64 `json:"versions"`
	}
	_ = json.Unmarshal(out, &current)
	ev.Check("freezing does not advance effective versions", exit == 0 && current.Versions["openai"] == seedVersions["openai"], "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", frozen.Candidate.ID)
	ev.Check("candidate can be queried without applying", exit == 0 && bytes.Contains(out, []byte(frozen.Candidate.ID)), "exit=%d output=%s", exit, out)
	badStore := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(badStore, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	args[2] = badStore
	out, exit = invokeCatalogCLI(t, ev, binary, args...)
	after, err := os.ReadFile(badStore)
	ev.Check("save failure leaves original path intact", exit != 0 && err == nil && string(after) == "untouched", "exit=%d err=%v output=%s", exit, err, out)
}

func TestCatalogMaintenanceCLICandidateTamper(t *testing.T) {
	ev := run.Case(t, "CATALOG-T06-cli-candidate-path-binding")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	ids := []string{}
	for _, model := range []string{"gpt-6.1-sol", "gpt-6-sol"} {
		out, exit := invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", model, "--reason", "tamper fixture", "--confirm")
		var result struct {
			Candidate struct {
				ID string `json:"id"`
			} `json:"candidate"`
		}
		_ = json.Unmarshal(out, &result)
		if !ev.Check("valid public candidate is prepared", exit == 0 && result.Candidate.ID != "", "exit=%d output=%s", exit, out) {
			return
		}
		ids = append(ids, result.Candidate.ID)
	}
	objects := map[string][]byte{}
	paths := map[string]string{}
	filepath.WalkDir(store, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var object struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &object) == nil {
			objects[object.ID] = raw
			paths[object.ID] = path
		}
		return nil
	})
	if err := os.WriteFile(paths[ids[0]], objects[ids[1]], 0600); err != nil {
		t.Fatal(err)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", ids[0])
	ev.Check("requested identity cannot resolve another valid object", exit != 0, "exit=%d output=%s", exit, out)
}
