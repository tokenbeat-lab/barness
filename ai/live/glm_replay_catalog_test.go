//go:build live

package live

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// Historical playback uses the catalog that priced the original requests.
// The repository's pinned capture index selects an immutable Git object, never
// a caller-supplied catalog. This authority is replay-only: actual formal
// selection and the support importer continue to require today's built-in.
func glmOriginalReplayCatalog() (ai.Catalog, error) {
	var empty ai.Catalog
	raw, err := glmCaptureIndexData()
	var index glmCaptureIndex
	if err != nil || json.Unmarshal(raw, &index) != nil || index.Schema != 1 ||
		!regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(index.OriginalLiveSourceCommit) {
		return empty, fmt.Errorf("historical replay catalog source invalid")
	}
	const name = "ai/release/catalog-snapshot.json"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "--no-replace-objects", "show", index.OriginalLiveSourceCommit+":"+name)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return empty, fmt.Errorf("historical replay catalog unavailable")
	}
	if err = cmd.Start(); err != nil {
		return empty, fmt.Errorf("historical replay catalog unavailable")
	}
	raw, err = io.ReadAll(io.LimitReader(stdout, (2<<20)+1))
	if err != nil || len(raw) > 2<<20 {
		cmd.Process.Kill()
		cmd.Wait()
		return empty, fmt.Errorf("historical replay catalog exceeds read bound")
	}
	if err = cmd.Wait(); err != nil {
		return empty, fmt.Errorf("historical replay catalog unavailable")
	}
	var snapshot struct {
		Hash    string
		Catalog ai.Catalog
	}
	if json.Unmarshal(raw, &snapshot) != nil || snapshot.Catalog.Version == "" {
		return empty, fmt.Errorf("historical replay catalog invalid")
	}
	hash, err := snapshot.Catalog.Hash()
	if err != nil || hash != snapshot.Hash {
		return empty, fmt.Errorf("historical replay catalog hash differs")
	}
	if err = run.Record("capture-replay-catalog", map[string]any{
		"sourceCommit": index.OriginalLiveSourceCommit, "path": name, "snapshotSHA256": sha256Text(raw),
		"catalogHash": hash, "catalogVersion": snapshot.Catalog.Version, "authority": "historical replay only; no current certification",
	}); err != nil {
		return empty, err
	}
	return snapshot.Catalog, nil
}
