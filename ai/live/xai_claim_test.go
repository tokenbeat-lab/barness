//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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
	if err = os.MkdirAll(store, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(store)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("claim storage must be a directory")
	}
	root, err := os.OpenRoot(store)
	if err != nil {
		return err
	}
	defer root.Close()
	receipt := map[string]any{"schema": 1, "grantId": p.grant.ID, "grantSha256": p.report.GrantSHA256, "runId": filepath.Base(run.Dir()), "claimedAt": time.Now().UTC(), "sourceSha256": p.report.SourceSHA256, "binarySha256": p.report.BinarySHA256}
	data, _ := json.Marshal(receipt)
	f, err := root.OpenFile(p.grant.ID+".json", os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return fmt.Errorf("grant already consumed or claim unavailable")
	}
	n, writeErr := f.Write(data)
	if writeErr == nil && n != len(data) {
		writeErr = fmt.Errorf("incomplete claim")
	}
	if writeErr == nil {
		writeErr = f.Sync()
	}
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	for _, dir := range []string{store, parent} {
		d, e := os.Open(dir)
		if e != nil {
			return e
		}
		e = d.Sync()
		d.Close()
		if e != nil {
			return e
		}
	}
	p.report.ClaimID = p.grant.ID
	if err = run.Record("xai-authorization-claim", receipt); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(run.Dir(), "xai-authorization-grant.json"), p.grantRaw, 0600)
}
