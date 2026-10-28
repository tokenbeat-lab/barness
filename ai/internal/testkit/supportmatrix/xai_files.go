package supportmatrix

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// The bounded pinned filesystem reader is shared with regional evidence; the
// xAI report, inventory and semantic contracts remain independently owned.
func xaiReadJSON(root *os.Root, name string, v any) error {
	raw, err := regionalRead(root, name)
	if err != nil {
		return err
	}
	if err := regionalDecode(raw, v); err != nil {
		return err
	}
	normalized, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var original, canonical any
	if json.Unmarshal(raw, &original) != nil || json.Unmarshal(normalized, &canonical) != nil || !xaiCanonicalKeys(original, canonical) {
		return refuse("xAI evidence contains unknown or noncanonical fields")
	}
	return nil
}
func xaiInventory(dir string) (map[string]string, error) {
	st, err := os.Lstat(filepath.Clean(dir))
	if err != nil || !st.IsDir() {
		return nil, refuse("xAI bundle root unavailable")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	out := map[string]string{}
	total := 0
	err = fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() || len(out) >= 512 {
			return refuse("xAI nonregular or excessive inventory")
		}
		raw, err := regionalRead(root, name)
		if err != nil {
			return err
		}
		total += len(raw)
		if total > 64<<20 {
			return refuse("xAI inventory bytes exceed bound")
		}
		out[name] = glmHash(raw)
		return nil
	})
	return out, err
}
func xaiSourceFiles() (map[string]string, error) {
	files, err := RegionalExecutionSource()
	if err != nil {
		return nil, err
	}
	root, err := minimaxlive.Root()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(root, "ai/internal/testkit/xai/candidate.json"))
	if err != nil {
		return nil, err
	}
	files["ai/internal/testkit/xai/candidate.json"] = glmHash(raw)
	return files, nil
}

// Manifest fields are decoded independently from the evidence writer's DTO.
type xaiManifestCase struct {
	ID        string            `json:"id"`
	Dir       string            `json:"dir"`
	Status    string            `json:"status"`
	Test      string            `json:"test"`
	Replay    string            `json:"replay"`
	Artifacts map[string]string `json:"artifacts"`
	Fixtures  map[string]string `json:"fixtures"`
}

func verifyXAIOrigin(root *os.Root, dir string, r xaiEvidence, inventory map[string]string) error {
	var manifest struct {
		Package   string            `json:"package"`
		GoVersion string            `json:"go_version"`
		Module    string            `json:"module"`
		Platform  string            `json:"platform"`
		Commit    string            `json:"git_commit"`
		Started   time.Time         `json:"started_at"`
		Finished  time.Time         `json:"finished_at"`
		Versions  map[string]string `json:"versions"`
		Cases     []xaiManifestCase `json:"cases"`
	}
	if xaiReadJSON(root, "manifest.json", &manifest) != nil || manifest.Package != "-tags live -timeout 30m0s ./ai/live" || manifest.Commit != r.GitCommit || manifest.Started.Before(r.StartedAt) || manifest.Started.After(r.FinishedAt) || manifest.Finished.Before(r.FinishedAt) || manifest.Finished.Sub(r.StartedAt) > 1800*time.Second || manifest.Versions["live_combo"] != r.Combo || manifest.Versions["live_sdk"] != r.SDK || manifest.Versions["model_catalog_version"] != r.CatalogVersion || manifest.Versions["model_catalog_hash"] != r.CatalogHash || manifest.Versions["xai_execution_source"] != r.SourceSHA256 {
		return refuse("xAI manifest provenance differs")
	}
	var source map[string]string
	current, err := xaiSourceFiles()
	if xaiReadJSON(root, "xai-execution-source.json", &source) != nil || err != nil || evidence.VerifyRouteSource(source, current, r.Provider, r.API) != nil || RegionalSourceHash(source) != r.SourceSHA256 {
		return refuse("xAI native execution source drift")
	}
	var recordedAudit struct {
		Findings []json.RawMessage `json:"findings"`
	}
	if xaiReadJSON(root, "audit.json", &recordedAudit) != nil || len(recordedAudit.Findings) != 0 {
		return refuse("xAI native audit did not pass")
	}
	if findings, err := audit.Bundle(dir, nil); err != nil || len(findings) != 0 {
		return refuse("xAI independent redaction audit failed")
	}
	grantRaw, err := regionalRead(root, "xai-authorization-grant.json")
	if err != nil || glmHash(grantRaw) != r.GrantSHA256 {
		return refuse("xAI grant bytes differ")
	}
	var g xaiGrant
	if xaiReadJSON(root, "xai-authorization-grant.json", &g) != nil || !xaiGrantMatches(g, r) {
		return refuse("xAI formal grant differs")
	}
	var claim xaiClaim
	if xaiReadJSON(root, "xai-authorization-claim.json", &claim) != nil || claim.Schema != 1 || claim.GrantID != r.GrantID || claim.GrantSHA256 != r.GrantSHA256 || claim.RunID != filepath.Base(dir) || claim.SourceSHA256 != r.SourceSHA256 || claim.BinarySHA256 != r.BinarySHA256 || claim.ClaimedAt.Before(r.StartedAt) || claim.ClaimedAt.After(r.FinishedAt) || !g.ExpiresAt.After(claim.ClaimedAt) || g.ExpiresAt.Sub(claim.ClaimedAt) > 24*time.Hour {
		return refuse("xAI original consumed grant claim differs")
	}
	if err := xaiPersistentClaim(g, claim); err != nil {
		return err
	}
	if r.Budget.ObservedMaximumUSD > g.MaxUSD {
		return refuse("xAI observed estimate exceeds authorized threshold")
	}
	roots := []string{"manifest.json", "audit.json", XAIReportFile, ReportFile, "xai-execution-source.json", "xai-authorization-grant.json", "xai-authorization-claim.json", "xai-formal-report.unfinalized.json"}
	if inventory["model-catalog.json"] != "" {
		roots = append(roots, "model-catalog.json")
	}
	owned := map[string]bool{"E17-xai-process-gate": true}
	for _, s := range r.Scenarios {
		owned[s.CaseID] = true
	}
	if len(manifest.Cases) != 19 {
		return refuse("xAI native case set differs")
	}
	seen := map[string]bool{}
	for _, c := range manifest.Cases {
		if !owned[c.ID] || seen[c.ID] || c.Dir != c.ID || filepath.Base(c.Dir) != c.Dir || c.Status != "PASS" || len(c.Artifacts) == 0 {
			return refuse("xAI case identity or outcome differs")
		}
		seen[c.ID] = true
		if c.ID == "E17-xai-process-gate" && (len(c.Artifacts) != 1 || c.Artifacts["assertions.json"] == "") {
			return refuse("xAI process gate contains undeclared capture")
		}
		for name, hash := range c.Artifacts {
			if filepath.Base(name) != name || inventory[c.Dir+"/"+name] != hash {
				return refuse("xAI case inventory changed")
			}
		}
		var checks struct {
			ID         string `json:"case_id"`
			Status     string `json:"status"`
			Test       string `json:"test"`
			Replay     string `json:"replay"`
			Assertions []struct {
				Pass   bool   `json:"pass"`
				Name   string `json:"name"`
				Detail string `json:"detail"`
			} `json:"assertions"`
		}
		if xaiReadJSON(root, c.Dir+"/assertions.json", &checks) != nil || checks.ID != c.ID || checks.Status != "PASS" || len(checks.Assertions) == 0 {
			return refuse("xAI case assertions missing")
		}
		for _, a := range checks.Assertions {
			if !a.Pass {
				return refuse("xAI PASS contains failed assertion")
			}
		}
	}
	for name := range inventory {
		parts := strings.Split(name, "/")
		if len(parts) == 1 && slices.Contains(roots, name) {
			continue
		}
		if len(parts) != 2 || !owned[parts[0]] {
			return refuse("xAI hidden root case or artifact")
		}
		found := false
		for _, c := range manifest.Cases {
			found = found || c.Dir == parts[0] && c.Artifacts[parts[1]] == inventory[name]
		}
		if !found {
			return refuse("xAI unrecorded artifact")
		}
	}
	for _, name := range roots[:7] {
		if inventory[name] == "" {
			return refuse("xAI required root artifact missing")
		}
	}
	return nil
}
func xaiGrantMatches(g xaiGrant, r xaiEvidence) bool {
	return g.Schema == 1 && g.ID == r.GrantID && g.Mode == "live" && g.Stage == "formal" && g.Combo == r.Combo && g.Provider == r.Provider && g.API == r.API && g.Endpoint == r.Endpoint && g.Model == r.Model && g.SourceSHA256 == r.SourceSHA256 && g.BinarySHA256 == r.BinarySHA256 && g.CatalogVersion == r.CatalogVersion && g.CatalogHash == r.CatalogHash && g.SDK == r.SDK && g.AccountAlias == r.AccountAlias && g.RegionAlias == r.RegionAlias && slices.Equal(g.Expected, r.Expected) && g.MaxCalls == 24 && g.MaxAttempts == 24 && g.MaxRequestedOutputTokens == 24576 && g.MaxReservedInputTokens == 196608 && g.MaxUSD > 0 && xaiFinite(g.MaxUSD)
}
func xaiPersistentClaim(g xaiGrant, claim xaiClaim) error {
	repo, err := minimaxlive.Root()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = repo
	raw, err := cmd.Output()
	if err != nil {
		return refuse("xAI trusted claim store unavailable")
	}
	root, err := os.OpenRoot(filepath.Join(strings.TrimSpace(string(raw)), "barness-xai-grants"))
	if err != nil {
		return refuse("xAI authoritative claim unavailable")
	}
	defer root.Close()
	var actual xaiClaim
	if xaiReadJSON(root, g.ID+".json", &actual) != nil || !reflect.DeepEqual(actual, claim) {
		return refuse("xAI authoritative claim changed")
	}
	return nil
}

// Canonical producer field names are checked after typed decoding, including
// Nullable values. Raw protocol/message payloads remain independent maps.
func xaiCanonicalKeys(original, canonical any) bool {
	switch v := original.(type) {
	case map[string]any:
		w, ok := canonical.(map[string]any)
		if !ok {
			return false
		}
		for k, x := range v {
			y, found := w[k]
			if !found || !xaiCanonicalKeys(x, y) {
				return false
			}
		}
	case []any:
		w, ok := canonical.([]any)
		if !ok || len(v) != len(w) {
			return false
		}
		for i, x := range v {
			if !xaiCanonicalKeys(x, w[i]) {
				return false
			}
		}
	}
	return true
}
