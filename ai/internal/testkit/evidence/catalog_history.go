package evidence

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// Older native bundles did not copy the full catalog. Its immutable measured
// Git snapshot remains original evidence under ADR-0055; reading it never
// changes the bundle, the measured catalog hash or execution authority.
func measuredGitCatalog(dir, version, hash string) (ai.Catalog, error) {
	bad := fmt.Errorf("original catalog snapshot unavailable or changed")
	commit := ""
	for _, name := range []string{"live-report.json", "regional-report.json", "formal-glm-report.json", "multimodal-report.json", "report.json"} {
		path := filepath.Join(dir, name)
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 4<<20 {
			continue
		}
		raw, err := os.ReadFile(path)
		var report struct {
			GitCommit string `json:"gitCommit"`
		}
		if err == nil && json.Unmarshal(raw, &report) == nil && report.GitCommit != "" {
			commit = report.GitCommit
			break
		}
	}
	decoded, err := hex.DecodeString(commit)
	if err != nil || len(decoded) != 20 {
		return ai.Catalog{}, bad
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "cat-file", "blob", commit+":ai/release/catalog-snapshot.json")
	cmd.Stderr = io.Discard
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "GIT_") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		return ai.Catalog{}, bad
	}
	raw, readErr := io.ReadAll(io.LimitReader(pipe, 4<<20+1))
	if readErr != nil || len(raw) > 4<<20 {
		cancel()
	}
	waitErr := cmd.Wait()
	var snapshot struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}
	if readErr != nil || len(raw) > 4<<20 || waitErr != nil || json.Unmarshal(raw, &snapshot) != nil || snapshot.Hash != hash || snapshot.Catalog.Version != version {
		return ai.Catalog{}, bad
	}
	if got, err := snapshot.Catalog.Hash(); err != nil || got != hash {
		return ai.Catalog{}, bad
	}
	return snapshot.Catalog, nil
}
