package supportmatrix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

type xaiImagesManifest struct {
	GoVersion string                  `json:"go_version,omitempty"`
	Platform  string                  `json:"platform,omitempty"`
	Module    string                  `json:"module"`
	Package   string                  `json:"package"`
	Commit    string                  `json:"git_commit"`
	Started   time.Time               `json:"started_at"`
	Finished  time.Time               `json:"finished_at"`
	Versions  map[string]string       `json:"versions"`
	Cases     []xaiImagesManifestCase `json:"cases"`
}
type xaiImagesManifestCase struct {
	Replay    string            `json:"replay,omitempty"`
	ID        string            `json:"id"`
	Dir       string            `json:"dir"`
	Status    string            `json:"status"`
	Test      string            `json:"test"`
	Artifacts map[string]string `json:"artifacts"`
	Fixtures  map[string]string `json:"fixtures"`
}

func verifyXAIImagesOrigin(root *os.Root, dir string, n xaiImagesNative, inventory map[string]string) (xaiImagesManifest, error) {
	var m xaiImagesManifest
	e := xaiReadJSON(root, "manifest.json", &m)
	if e != nil {
		return m, refuse("Imagine manifest unavailable")
	}
	if m.Module != "barness-ai" || m.Package != "-tags live -timeout 20m0s ./ai/live" || m.Commit != n.GitCommit || m.Started.Before(n.StartedAt) || m.Started.After(n.FinishedAt) || m.Finished.Before(n.FinishedAt) || m.Finished.IsZero() || len(m.Cases) != 7 {
		return m, refuse("Imagine manifest identity or interval differs")
	}
	for k, v := range map[string]string{"process_build_source": n.Scope.SourceSHA256, "xai_images_source": n.Scope.SourceSHA256, "live_combo": xaiimages.Route, "live_sdk": xaiimages.SDK, "model_catalog_version": n.Scope.CatalogVersion, "model_catalog_hash": n.Scope.CatalogHash} {
		if m.Versions[k] != v {
			return m, refuse("Imagine manifest source or catalog differs")
		}
	}
	expected := map[string]string{}
	for _, name := range []string{"audit.json", "manifest.json", "execution-source.json", "model-catalog.json", "process-grant.json", "process-authorization-receipt.json", XAIImagesReportFile} {
		if inventory[name] == "" {
			return m, refuse("Imagine root input missing")
		}
		expected[name] = inventory[name]
	}
	cases := map[string]bool{}
	for _, c := range m.Cases {
		if cases[c.ID] || c.ID != c.Dir || filepath.Base(c.ID) != c.ID || c.Status != "PASS" || (c.ID != "XAIIMG-live-gate" && !slices.Contains(xaiimages.Expected(), strings.TrimPrefix(c.ID, "XAIIMG-live-"))) {
			return m, refuse("Imagine manifest case identity differs")
		}
		cases[c.ID] = true
		gate := c.ID == "XAIIMG-live-gate"
		wantTest := "TestXAIImagesCandidate"
		if !gate {
			wantTest += "/" + strings.TrimPrefix(c.ID, "XAIIMG-live-")
		}
		if c.Test != wantTest {
			return m, refuse("Imagine case test entry differs")
		}
		for name, h := range c.Artifacts {
			if filepath.Base(name) != name || inventory[c.Dir+"/"+name] != h || !validManifestHash(h) {
				return m, refuse("Imagine captured case inventory changed")
			}
			expected[c.Dir+"/"+name] = h
		}
		for name, h := range c.Fixtures {
			if c.Artifacts[name] != h {
				return m, refuse("Imagine fixture inventory differs")
			}
		}
		var assertions struct {
			ID         string `json:"case_id"`
			Status     string `json:"status"`
			Test       string `json:"test"`
			Replay     string `json:"replay,omitempty"`
			Assertions []struct {
				Name   string `json:"name"`
				Detail string `json:"detail,omitempty"`
				Pass   bool   `json:"pass"`
			} `json:"assertions"`
		}
		e := xaiReadJSON(root, c.Dir+"/assertions.json", &assertions)
		if e != nil || assertions.ID != c.ID || assertions.Status != "PASS" || assertions.Test != c.Test || len(assertions.Assertions) == 0 {
			return m, refuse("Imagine case assertions absent")
		}
		for _, a := range assertions.Assertions {
			if !a.Pass || a.Name == "" {
				return m, refuse("Imagine PASS conceals failed assertion")
			}
		}
		if gate && (len(c.Artifacts) != 1 || len(c.Fixtures) != 0) {
			return m, refuse("Imagine gate hides extra artifacts")
		}
	}
	if !reflect.DeepEqual(inventory, expected) {
		return m, refuse("Imagine whole native inventory incomplete or excessive")
	}
	var recordedAudit struct {
		Findings []json.RawMessage `json:"findings"`
	}
	if xaiReadJSON(root, "audit.json", &recordedAudit) != nil || len(recordedAudit.Findings) != 0 {
		return m, refuse("Imagine original audit failed")
	}
	findings, e := audit.Bundle(dir, audit.Environment())
	if e != nil || len(findings) > 0 {
		return m, refuse("Imagine current native audit failed")
	}
	var original map[string]string
	if xaiReadJSON(root, "execution-source.json", &original) != nil || len(original) != 545 || RegionalSourceHash(original) != xaiImagesMeasuredSource {
		return m, refuse("Imagine original complete execution source differs")
	}
	current, e := XAIImagesExecutionSource()
	if e != nil {
		return m, e
	}
	if e = evidence.VerifyRouteSource(original, current, "xai", "xai-images"); e != nil {
		return m, e
	}
	if _, e = XAIImagesRouteCatalog(dir, n.Scope.CatalogVersion, n.Scope.CatalogHash); e != nil {
		return m, e
	}
	return m, nil
}

// XAIImagesRouteCatalog accepts only this route's measured narrowing.
// Shared catalog comparison
// remains exact for all other providers, protocols and measurement contracts.
func XAIImagesRouteCatalog(dir, version, hash string) (ai.Catalog, error) {
	var measured ai.Catalog
	root, e := os.OpenRoot(dir)
	if e != nil {
		return measured, e
	}
	defer root.Close()
	if xaiReadJSON(root, "model-catalog.json", &measured) != nil {
		return measured, refuse("Imagine original catalog absent")
	}
	candidate := xai.ImagesCandidateCatalog()
	h, e := measured.Hash()
	if e != nil || version != candidate.Version || h != hash || !reflect.DeepEqual(measured, candidate) {
		return measured, refuse("Imagine original candidate catalog changed")
	}
	current, ok := ai.BuiltinCatalog().LookupImage(ai.ProviderXAI, ai.APIXAIImages, xaiimages.Model)
	old := candidate.ImageModels[0]
	if !ok || current.Provider != old.Provider || current.API != old.API || current.ID != old.ID || !reflect.DeepEqual(current.Input, old.Input) || !reflect.DeepEqual(current.Output, old.Output) || !reflect.DeepEqual(current.Pricing, old.Pricing) {
		return measured, refuse("Imagine current identity or price is not measured")
	}
	expected := ai.ImageCapabilities{MaxReferenceImages: 5, MaxOutputImages: 2, ImageSizes: []string{"1k", "1.5k", "2k"}, Qualities: []string{"low", "medium", "auto"}, AspectRatios: []string{"1:1", "16:9"}}
	if !reflect.DeepEqual(current.Capabilities, expected) {
		return measured, refuse("Imagine current capabilities exceed measured scope")
	}
	return measured, nil
}
