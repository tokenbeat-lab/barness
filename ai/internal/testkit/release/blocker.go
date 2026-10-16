package release

import (
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// HistoricalPressureBlocker preserves the original issue's causal-evidence
// requirement independently of rerun success. Only resolution in its original
// tracker can clear it; the integration gate cannot close that issue.
func HistoricalPressureBlocker() Snapshot {
	const issue = ".scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md"
	s := Snapshot{Name: "pressure-stream-integrity", Detail: issue + ": original delivery acceptance unresolved"}
	root, err := sourceRoot()
	if err != nil {
		return s
	}
	raw, err := readArtifact(filepath.Join(root, filepath.FromSlash(issue)), 64<<10)
	if err != nil {
		return s
	}
	statuses := []string{}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "Status:") {
			statuses = append(statuses, strings.TrimSpace(strings.TrimPrefix(line, "Status:")))
		}
	}
	s.Current = len(statuses) == 1 && statuses[0] == "resolved"
	s.Detail += "; sha256:" + pioracle.SHA256(raw)
	if s.Current {
		s.Detail = issue + ": resolved by original tracker; sha256:" + pioracle.SHA256(raw)
	}
	return s
}
