package supportmatrix

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func VerifyProcessAuthorization(dir, route, account, region string, started, finished time.Time, supplement ...string) error {
	g, raw, r, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return e
	}
	want, e := ProcessScope(route, "live", "formal", g.Scope.BinarySHA256)
	if e != nil {
		return e
	}

	current, err := RegionalExecutionSource()
	names := []string{"execution-source.json"}
	if strings.HasPrefix(route, "zhipu-") {
		current, err = GLMExecutionSource()
		names = append(names, "glm-execution-source.json")
	}
	if err != nil {
		return err
	}
	if len(supplement) > 1 {
		return refuse("one original source supplement required")
	}
	var original map[string]string
	if len(supplement) == 1 && supplement[0] != "" {
		rawSource, err := glmRead(supplement[0])
		if err != nil || json.Unmarshal(rawSource, &original) != nil || RegionalSourceHash(original) != g.Scope.SourceSHA256 {
			return refuse("original source supplement differs")
		}
	}
	for _, name := range names {
		path := filepath.Join(dir, name)
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		}
		rawSource, err := glmRead(path)
		var captured map[string]string
		if err != nil || json.Unmarshal(rawSource, &captured) != nil || RegionalSourceHash(captured) != g.Scope.SourceSHA256 {
			return refuse("original process source differs")
		}
		original = captured
		break
	}
	if original == nil {
		if RegionalSourceHash(current) != g.Scope.SourceSHA256 {
			return refuse("original process source unavailable")
		}
		original = current
	}
	if err := evidence.VerifyRouteSource(original, current, g.Scope.Provider, g.Scope.API); err != nil {
		return err
	}
	report, err := LoadReport(filepath.Join(dir, ReportFile))
	if err != nil {
		return err
	}
	if _, err := RouteCatalog(dir, report); err != nil {
		return err
	}
	want.SourceSHA256, want.CatalogVersion, want.CatalogHash = g.Scope.SourceSHA256, report.CatalogVersion, report.CatalogHash
	if region == "" {
		region = g.RegionAlias
	}
	return livegrant.VerifyEvidence(g, raw, r, want, account, region, started, finished)
}
