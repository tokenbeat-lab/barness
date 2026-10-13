package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// Both offline and live imports require the exact case file inventory. Looking
// only at listed hashes would let a deleted manifest entry bypass verification.
func verifyArtifactInventory(dir string, artifacts map[string]string) error {
	if len(artifacts) == 0 {
		return fmt.Errorf("case artifact integrity inventory missing")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != len(artifacts) {
		return fmt.Errorf("case artifact inventory is incomplete")
	}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || artifacts[entry.Name()] == "" {
			return fmt.Errorf("case contains an unlisted or non-regular artifact")
		}
	}
	for name, hash := range artifacts {
		if len(audit.File("identifiers.txt", []byte(name), audit.Environment())) != 0 || filepath.Base(name) != name {
			return fmt.Errorf("artifact identifier or path invalid")
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || pioracle.SHA256(raw) != hash {
			return fmt.Errorf("case artifact missing or changed")
		}
	}
	return nil
}
