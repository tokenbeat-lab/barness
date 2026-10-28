//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// A grant is consumed in common Git storage, before requests, across worktrees
// and reruns. An incomplete write consumes it too; failure never restores it.
func (p *xaiProcess) claim() error {
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	raw, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("claim storage unavailable")
	}
	parent := strings.TrimSpace(string(raw))
	store := filepath.Join(parent, "barness-xai-grants")
	receipt := map[string]any{"schema": 1, "grantId": p.grant.ID, "grantSha256": p.report.GrantSHA256, "runId": filepath.Base(run.Dir()), "claimedAt": time.Now().UTC(), "sourceSha256": p.report.SourceSHA256, "binarySha256": p.report.BinarySHA256}
	data, _ := json.Marshal(receipt)
	if err := livegrant.Consume(store, []string{p.grant.ID + ".json"}, data); err != nil {
		return err
	}
	p.report.ClaimID = p.grant.ID
	if err = run.Record("xai-authorization-claim", receipt); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(run.Dir(), "xai-authorization-grant.json"), p.grantRaw, 0600)
}
