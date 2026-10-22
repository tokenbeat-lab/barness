package release

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func verifyLiveManifest(dir string, r supportmatrix.Report) error {
	if _, ok := supportmatrix.RegionalRouteOf(r.Combo); ok {
		return supportmatrix.VerifyRegionalReport(r)
	}
	if strings.HasPrefix(r.Combo, "minimax-") {
		return supportmatrix.VerifyMiniMaxReport(r)
	}
	if len(supportmatrix.GLMExpected(r.Combo)) > 0 {
		return supportmatrix.VerifyGLMReport(r)
	}
	if r.Combo == supportmatrix.JevFormalCombo {
		return supportmatrix.VerifyJevReport(r, true)
	}
	source, root, err := openArtifactBundle(dir)
	if err != nil {
		return err
	}
	defer root.Close()
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
	if err := readBundleManifest(root, &m); err != nil {
		return err
	}
	packageArgs := "-tags live ./ai/live"
	if r.Combo == "openai-decisions" {
		packageArgs = "-tags live -timeout 180s ./ai/live"
	}
	if m.Package != packageArgs || m.Commit != r.GitCommit || m.Started.IsZero() || m.Finished.Before(r.FinishedAt) || m.Started.Before(r.StartedAt) || m.Started.After(r.FinishedAt) || m.Versions["live_combo"] != r.Combo || m.Versions["live_sdk"] != r.SDK || m.Versions["model_catalog_version"] != r.CatalogVersion || m.Versions["model_catalog_hash"] != r.CatalogHash {
		return fmt.Errorf("live manifest provenance disagrees with report")
	}
	seenCalls := map[string]bool{}
	var chatTotals chatUsage
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
			child, err := openInventoryCase(root, c.Dir)
			if err != nil {
				return err
			}
			err = verifyCaseArtifactInventory(child, c.Artifacts)
			child.Close()
			if err != nil {
				return err
			}
			files := verifiedCase{bundle: source, dir: c.Dir, artifacts: c.Artifacts, verified: true}
			if r.Operation == "chat" {
				usage, err := verifyChatCaptures(files, r, s, seenCalls)
				if err != nil {
					return err
				}
				chatTotals.input += usage.input
				chatTotals.output += usage.output
			}
			if s.Attempts > 0 {
				patterns := []string{"attempt-*-exchanges.json", "attempt-*-result.json", "observations-*.json"}
				if r.Combo == "openai-decisions" {
					patterns = []string{"exchanges.json", "result.json", "observations.json"}
				}
				for _, pattern := range patterns {
					if len(files.matches(pattern)) == 0 {
						return fmt.Errorf("live capture, consumption or Observer missing")
					}
				}
				if r.Operation == "image" && s.Outcome == supportmatrix.Pass {
					if len(files.matches("attempt-*-image-check.json")) == 0 {
						return fmt.Errorf("live raster verification missing")
					}
				}
			}
			var assertions struct {
				ID         string `json:"case_id"`
				Status     string
				Assertions []struct{ Pass bool }
			}
			if err := files.json("assertions.json", maxInventoryFile, &assertions); err != nil {
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
				raw, err := files.read(name, maxInventoryFile)
				if err != nil || evidence.Digest(raw) != hash {
					return fmt.Errorf("live fixture missing or changed")
				}
			}
		}
		if matches != 1 {
			return fmt.Errorf("live scenario lacks a unique captured case")
		}
	}
	if r.Operation == "chat" && (chatTotals.input != r.Budget.InputTokensUsed || chatTotals.output != r.Budget.OutputTokensUsed) {
		return fmt.Errorf("chat captured usage disagrees with budget")
	}
	return nil
}
