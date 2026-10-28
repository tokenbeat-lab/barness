package supportmatrix

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

const regionalSealFile = "regional-integrity.json"

type regionalSeal struct {
	Schema int               `json:"schema"`
	Files  map[string]string `json:"files"`
}
type regionalManifestCase struct {
	ID, Dir, Status string
	Artifacts       map[string]string
}

func regionalRead(root *os.Root, name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, refuse("unsafe regional artifact path")
	}
	before, err := root.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Size() > 16<<20 {
		return nil, refuse("regional artifact missing, nonregular or oversized")
	}
	f, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, e1 := f.Stat()
	after, e2 := root.Lstat(name)
	if e1 != nil || e2 != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) || !os.SameFile(before, after) {
		return nil, refuse("regional artifact changed during open")
	}
	raw, err := io.ReadAll(io.LimitReader(f, 16<<20+1))
	if err != nil || len(raw) > 16<<20 {
		return nil, refuse("regional artifact exceeds bound")
	}
	return raw, nil
}

func regionalInventory(dir string) (map[string]string, error) {
	st, err := os.Lstat(filepath.Clean(dir))
	if err != nil || !st.IsDir() {
		return nil, refuse("regional bundle root unavailable")
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
		if !d.Type().IsRegular() {
			return refuse("regional bundle contains nonregular member")
		}
		if name == regionalSealFile || name == ReportFile {
			return nil
		}
		if len(out) >= 512 {
			return refuse("regional bundle file count exceeds bound")
		}
		raw, err := regionalRead(root, name)
		if err != nil {
			return err
		}
		total += len(raw)
		if total > 64<<20 {
			return refuse("regional bundle byte count exceeds bound")
		}
		out[name] = glmHash(raw)
		return nil
	})
	return out, err
}

// SealRegionalBundle is called only after final reports and the native audit.
// It never rewrites an old origin, source pin or receipt to renew support.
func SealRegionalBundle(dir string) error {
	files, err := regionalInventory(dir)
	if err != nil {
		return err
	}
	raw, err := Encode(regionalSeal{Schema: 1, Files: files})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, regionalSealFile), raw, 0600)
}

func verifyRegionalBundle(dir string, r Report, selection regionalSelection) error {
	var seal regionalSeal
	if err := readJSON(filepath.Join(dir, regionalSealFile), &seal); err != nil || seal.Schema != 1 {
		return refuse("regional sealed inventory unavailable")
	}
	actual, err := regionalInventory(dir)
	if err != nil || !reflect.DeepEqual(actual, seal.Files) {
		return refuse("regional captured inventory changed, missing or extra")
	}
	rootNames := []string{"manifest.json", "audit.json", RegionalReportFile, "regional-execution-source.json", "regional-catalog.json", "regional-candidate-facts.json", "regional-grant.json", "regional-authorization-receipt.json"}
	if actual["model-catalog.json"] != "" {
		rootNames = append(rootNames, "model-catalog.json")
	}
	caseNames := []string{"assertions.json", "request.json", "tools.json", "options.json", "result.json", "events.json", "exchanges.json", "observations.json", "resources.json", "immutable-input.json", "history-provenance.json"}
	owned := map[string]bool{}
	for _, s := range r.Scenarios {
		owned[s.CaseID] = true
	}
	for name := range actual {
		parts := strings.Split(name, "/")
		if len(parts) == 1 && slices.Contains(rootNames, name) {
			continue
		}
		if len(parts) != 2 || !owned[parts[0]] || !slices.Contains(caseNames, parts[1]) {
			return refuse("regional hidden case, attempt or artifact")
		}
	}
	for _, name := range rootNames {
		if actual[name] == "" {
			return refuse("regional required root artifact absent")
		}
	}
	if findings, err := audit.Bundle(dir, nil); err != nil || len(findings) > 0 {
		return refuse("regional independent redaction audit failed")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	read := func(name string, v any) error {
		raw, err := regionalRead(root, name)
		if err != nil {
			return err
		}
		return regionalDecode(raw, v)
	}
	var manifest struct {
		Package  string
		Commit   string    `json:"git_commit"`
		Started  time.Time `json:"started_at"`
		Finished time.Time `json:"finished_at"`
		Versions map[string]string
		Cases    []regionalManifestCase
	}
	if read("manifest.json", &manifest) != nil || manifest.Package != fmt.Sprintf("-tags live -timeout %ds ./ai/live", selectedRegionalBudget(r.Combo, selection).ProcessSeconds) || manifest.Commit != r.GitCommit || manifest.Started.Before(r.StartedAt) || manifest.Started.After(r.FinishedAt) || manifest.Finished.Before(r.FinishedAt) || manifest.Versions["regional_mode"] != r.Regional.Mode || manifest.Versions["regional_stage"] != r.Regional.Stage || manifest.Versions["live_combo"] != r.Combo || manifest.Versions["live_sdk"] != r.SDK || manifest.Versions["model_catalog_hash"] != r.CatalogHash || manifest.Versions["model_catalog_version"] != r.CatalogVersion {
		return refuse("regional native manifest provenance differs")
	}
	var recorded map[string]string
	current, err := RegionalExecutionSource()
	if read("regional-execution-source.json", &recorded) != nil || err != nil || evidence.VerifyRouteSource(recorded, current, r.Provider, r.API) != nil || r.Regional.SourceSHA256 != RegionalSourceHash(recorded) {
		return refuse("regional execution source drift")
	}
	var origin RegionalReservation
	if read("regional-authorization-receipt.json", &origin) != nil || origin.Schema != 2 || origin.Mode != r.Regional.Mode || origin.GrantID != r.Regional.GrantID || origin.GrantSHA256 != r.Regional.GrantSHA256 || origin.Combo != r.Combo || origin.Session != filepath.Base(dir) || origin.StartedAt.Before(r.StartedAt) || origin.StartedAt.After(r.FinishedAt) || origin.MaxCalls != r.Budget.MaxCalls || origin.MaxAttempts != r.Budget.MaxAttempts {
		return refuse("regional original authorization receipt differs")
	}
	grantRaw, err := regionalRead(root, "regional-grant.json")
	if err != nil || glmHash(grantRaw) != r.Regional.GrantSHA256 {
		return refuse("regional own authorization bytes differ")
	}
	var grant RegionalGrant
	if json.Unmarshal(grantRaw, &grant) != nil || grant.checkSelectedMeasurement(r, r.StartedAt, selection) != nil || grant.Stage != r.Regional.Stage || grant.Mode != r.Regional.Mode || grant.ID != r.Regional.GrantID || grant.AccountAlias != r.AccountAlias || grant.RegionAlias != r.Regional.RegionAlias {
		return refuse("regional own authorization target differs")
	}
	if RegionalIsKimi(r.Combo) && (grant.MaxUSD != r.Regional.MaxUSD || grant.BinarySHA256 != r.Regional.BinarySHA256 || origin.MaxUSD != grant.MaxUSD || origin.BinarySHA256 != grant.BinarySHA256) {
		return refuse("Kimi original USD/binary grant differs")
	}
	if err := verifyRegionalReservation(grant, origin); err != nil {
		return err
	}
	if err := verifyRegionalPacing(root, r, grant, origin); err != nil {
		return err
	}
	if err := verifyRegionalTotals(root, r); err != nil {
		return err
	}
	if len(manifest.Cases) != len(r.Scenarios) {
		return refuse("regional complete native case set differs")
	}
	for _, s := range r.Scenarios {
		matches := 0
		for _, c := range manifest.Cases {
			if c.ID != s.CaseID {
				continue
			}
			matches++
			if c.Dir != c.ID || filepath.Base(c.Dir) != c.Dir || c.Status != "PASS" {
				return refuse("regional case identity/outcome differs")
			}
			required := caseNames[:len(caseNames)-2]
			if strings.HasSuffix(s.ID, "/xai-origin-history") {
				required = caseNames
			}
			if len(c.Artifacts) != len(required) {
				return refuse("regional fixed case artifact set differs")
			}
			for _, name := range required {
				if c.Artifacts[name] == "" {
					return refuse("regional case artifact missing")
				}
			}
			for name, hash := range c.Artifacts {
				if filepath.Base(name) != name || actual[c.Dir+"/"+name] != hash {
					return refuse("regional case artifact inventory differs")
				}
			}
			if err := verifyRegionalCase(root, r, s); err != nil {
				return err
			}
		}
		if matches != 1 {
			return fmt.Errorf("%w: regional own unique case missing", ErrInvalidReport)
		}
	}
	return nil
}
