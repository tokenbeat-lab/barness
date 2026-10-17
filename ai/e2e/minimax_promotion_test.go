package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	after, e := os.ReadFile(path)
	ev.Check("rejected proposal leaves catalog exact", e == nil && sha256.Sum256(before) == sha256.Sum256(after), "catalog mutated")
	ev.Fixture("unchanged-catalog.json", after)
}
