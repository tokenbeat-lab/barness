package minimaxlive

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StageLedgerPath is common to every worktree of this repository. A new
// checkout or process cannot reset the six/three-identity paid stage budget.
func StageLedgerPath(stage string) (string, error) {
	if id := os.Getenv("BARNESS_MINIMAX_FORMAL_RENEWAL"); id != "" {
		if stage != "formal" {
			return "", fmt.Errorf("renewal requires formal stage")
		}
		if _, _, e := approvedRenewal(id, true); e != nil {
			return "", e
		}
		return RenewalLedgerPath(id)
	}
	return originalStageLedgerPath(stage)
}
func originalStageLedgerPath(stage string) (string, error) {
	if stage != "candidate" && stage != "formal" {
		return "", fmt.Errorf("invalid MiniMax stage")
	}
	root, e := Root()
	if e != nil {
		return "", e
	}
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = root
	cmd.Env = []string{}
	raw, e := cmd.Output()
	if e != nil {
		return "", fmt.Errorf("common MiniMax stage root unavailable")
	}
	common := strings.TrimSpace(string(raw))
	if !filepath.IsAbs(common) {
		return "", fmt.Errorf("common MiniMax stage root must be absolute")
	}
	common, e = filepath.EvalSymlinks(common)
	if e != nil {
		return "", e
	}
	return filepath.Join(filepath.Dir(common), ".evidence", "minimax-stages", stage+".jsonl"), nil
}
