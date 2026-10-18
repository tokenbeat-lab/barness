package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Approval/candidate/source failures must leave both compiled catalog data and
// prospective promotion receipt untouched. Positive activation needs three
// actual own reports; controlled process tests independently deny that bypass.
func TestMiniMaxPromotionBoundary(t *testing.T) {
	ev := run.Case(t, "MM10-promotion-boundary")
	path := "../catalog_minimax.json"
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	receipt := filepath.Join(t.TempDir(), "promotion.json")
	for _, withReceipt := range []bool{false, true} {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		args := []string{"run", "./ai/live/cmd/minimaxpromote"}
		if withReceipt {
			args = append(args, "-receipt", receipt)
		}
		args = append(args, "missing-chat", "missing-responses", "missing-messages")
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = "../.."
		for _, k := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
			if v, ok := os.LookupEnv(k); ok {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
		}
		output, e := cmd.CombinedOutput()
		cancel()
		ev.Record(fmt.Sprintf("process-%t", withReceipt), map[string]any{"success": e == nil, "output": string(output), "credentialsInherited": false})
		ev.Check("missing required proof refused", e != nil, "activation succeeded")
		_, e = os.Lstat(receipt)
		ev.Check("no unverified promotion receipt", os.IsNotExist(e), "receipt exists")
	}
	binary := filepath.Join(t.TempDir(), "minimaxpromote")
	build := exec.Command("go", "build", "-o", binary, "./ai/live/cmd/minimaxpromote")
	build.Dir = "../.."
	output, buildErr := build.CombinedOutput()
	if !ev.Check("prebuilt promotion command", buildErr == nil, "%v %s", buildErr, output) {
		return
	}
	targetRoot := t.TempDir()
	os.MkdirAll(filepath.Join(targetRoot, "ai"), 0700)
	os.WriteFile(filepath.Join(targetRoot, "go.mod"), []byte("module isolated-activation\n"), 0600)
	target := filepath.Join(targetRoot, "ai/catalog_minimax.json")
	changed := []byte(`{"version":"already-active-or-edited","models":[]}`)
	os.WriteFile(target, changed, 0600)
	invoke := func(name, want string) {
		cmd := exec.Command(binary, "-receipt", filepath.Join(targetRoot, name+".json"), "absent-chat", "absent-responses", "absent-messages")
		cmd.Dir = targetRoot
		cmd.Env = []string{}
		output, err := cmd.CombinedOutput()
		ev.Check("same binary refuses "+name, err != nil && strings.Contains(string(output), want), "%v %s", err, output)
		_, err = os.Lstat(filepath.Join(targetRoot, name+".json"))
		ev.Check("refused activation preserves receipt history "+name, os.IsNotExist(err), "receipt was created")
	}
	invoke("first-distinct-receipt", "inactive catalog target changed")
	invoke("second-distinct-receipt", "inactive catalog target changed")
	unchanged, _ := os.ReadFile(target)
	ev.Check("prebuilt binary cannot overwrite disk activation", string(unchanged) == string(changed), "activation overwritten")
	os.WriteFile(target, before, 0600)
	os.Mkdir(target+".activation.lock", 0700)
	var wg sync.WaitGroup
	for _, name := range []string{"concurrent-one", "concurrent-two"} {
		wg.Go(func() { invoke(name, "activation locked") })
	}
	wg.Wait()
	after, e := os.ReadFile(path)
	ev.Check("rejected proposal leaves catalog exact", e == nil && sha256.Sum256(before) == sha256.Sum256(after), "catalog mutated")
	ev.Fixture("unchanged-catalog.json", after)
}
