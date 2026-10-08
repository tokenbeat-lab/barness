package release

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func verifyLiveManifest(dir string, r supportmatrix.Report) error {
	var m struct {
		Package  string
		Commit   string    `json:"git_commit"`
		Started  time.Time `json:"started_at"`
		Finished time.Time `json:"finished_at"`
		Versions map[string]string
		Cases    []struct {
			ID, Dir, Status string
			Fixtures        map[string]string
			Artifacts       map[string]string
		}
	}
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m); err != nil {
		return err
	}
	if m.Package != "-tags live ./ai/live" || m.Commit != r.GitCommit || m.Started.IsZero() || m.Finished.Before(r.FinishedAt) || m.Started.Before(r.StartedAt) || m.Started.After(r.FinishedAt) || m.Versions["live_combo"] != r.Combo || m.Versions["live_sdk"] != r.SDK || m.Versions["model_catalog_version"] != r.CatalogVersion || m.Versions["model_catalog_hash"] != r.CatalogHash {
		return fmt.Errorf("live manifest provenance disagrees with report")
	}
	for _, s := range r.Scenarios {
		matches := 0
		for _, c := range m.Cases {
			if c.ID != s.CaseID {
				continue
			}
			matches++
			if c.Dir != c.ID || filepath.Base(c.Dir) != c.Dir || c.Status != string(s.Outcome) {
				return fmt.Errorf("live case identity/outcome disagrees with report")
			}
			if len(c.Artifacts) == 0 {
				return fmt.Errorf("live case lacks its complete artifact integrity inventory")
			}
			for name, hash := range c.Artifacts {
				if filepath.Base(name) != name {
					return fmt.Errorf("invalid live artifact path")
				}
				raw, err := os.ReadFile(filepath.Join(dir, c.Dir, name))
				if err != nil || pioracle.SHA256(raw) != hash {
					return fmt.Errorf("live artifact missing or changed")
				}
			}
			if s.Attempts > 0 {
				for _, pattern := range []string{"attempt-*-exchanges.json", "attempt-*-result.json", "observations-*.json"} {
					files, err := filepath.Glob(filepath.Join(dir, c.Dir, pattern))
					if err != nil || len(files) == 0 {
						return fmt.Errorf("live capture, consumption or Observer missing")
					}
				}
				if r.Operation == "image" && s.Outcome == supportmatrix.Pass {
					files, err := filepath.Glob(filepath.Join(dir, c.Dir, "attempt-*-image-check.json"))
					if err != nil || len(files) == 0 {
						return fmt.Errorf("live raster verification missing")
					}
				}
			}
			var assertions struct {
				ID         string `json:"case_id"`
				Status     string
				Assertions []struct{ Pass bool }
			}
			if err := readJSON(filepath.Join(dir, c.Dir, "assertions.json"), &assertions); err != nil {
				return err
			}
			if assertions.ID != c.ID || assertions.Status != c.Status || (s.Outcome == supportmatrix.Pass && len(assertions.Assertions) == 0) {
				return fmt.Errorf("live assertions missing or inconsistent")
			}
			for _, a := range assertions.Assertions {
				if !a.Pass && s.Outcome == supportmatrix.Pass {
					return fmt.Errorf("live PASS has a failed assertion")
				}
			}
			for name, hash := range c.Fixtures {
				if filepath.Base(name) != name {
					return fmt.Errorf("invalid live fixture path")
				}
				raw, err := os.ReadFile(filepath.Join(dir, c.Dir, name))
				if err != nil || pioracle.SHA256(raw) != hash {
					return fmt.Errorf("live fixture missing or changed")
				}
			}
		}
		if matches != 1 {
			return fmt.Errorf("live scenario lacks a unique captured case")
		}
	}
	return nil
}
