//go:build live

package live

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

var xaiImagesState *xaiImagesProcess

type xaiImagesProcess struct {
	cat                              ai.Catalog
	report                           xaiImagesReport
	authorization                    *processAuthorization
	key, controlled, problem, notRun string
}

type xaiImagesReport struct {
	Schema             int                    `json:"schema"`
	Kind               string                 `json:"kind"`
	Mode               string                 `json:"mode"`
	Stage              string                 `json:"stage"`
	Outcome            string                 `json:"outcome"`
	Failure            string                 `json:"failure,omitempty"`
	Complete           bool                   `json:"complete"`
	FormalQualified    bool                   `json:"formalQualified"`
	AuditPassed        bool                   `json:"auditPassed"`
	StartedAt          time.Time              `json:"startedAt"`
	FinishedAt         time.Time              `json:"finishedAt"`
	GitCommit          string                 `json:"gitCommit"`
	Scope              livegrant.Scope        `json:"scope"`
	BuildSourceSHA256  string                 `json:"buildSourceSha256"`
	Plan               []xaiimages.Scene      `json:"plan"`
	Scenarios          []xaiImagesSceneResult `json:"scenarios"`
	CallsUsed          int                    `json:"callsUsed"`
	HTTPAttempts       int                    `json:"httpAttempts"`
	VendorHTTPAttempts int                    `json:"vendorHTTPAttempts"`
	// InputImages and OutputImages are reservations, including failed calls.
	InputImages  int     `json:"inputImages"`
	OutputImages int     `json:"outputImages"`
	ReservedUSD  float64 `json:"reservedUSD"`
	ReportedUSD  float64 `json:"reportedUSD"`
}

type xaiImagesSceneResult struct {
	ID      string `json:"id"`
	Outcome string `json:"outcome"`
	CaseID  string `json:"caseId"`
}

func xaiImagesSelected() bool {
	return os.Getenv(envCombo) == xaiimages.Route || flag.Lookup("test.run").Value.String() == xaiimages.Entry || os.Getenv("BARNESS_XAI_IMAGES_CONTROLLED") != ""
}

func initializeXAIImages() {
	p := &xaiImagesProcess{cat: xai.ImagesCandidateCatalog(), controlled: os.Getenv("BARNESS_XAI_IMAGES_CONTROLLED")}
	xaiImagesState = p
	mode := "live"
	if p.controlled != "" {
		mode = "controlled"
	}
	binary, err := regionalBinaryHash()
	scope, scopeErr := supportmatrix.ProcessScope(xaiimages.Route, mode, "candidate", binary)
	p.report = xaiImagesReport{Schema: 1, Kind: "xai-images-candidate-only", Mode: mode, Stage: "candidate", Outcome: "NOT_RUN", StartedAt: started, GitCommit: evidence.GitCommit(), Scope: scope, BuildSourceSHA256: legacyBuildSource, Plan: xaiimages.Plan()}
	for _, scene := range p.report.Plan {
		p.report.Scenarios = append(p.report.Scenarios, xaiImagesSceneResult{scene.ID, "NOT_RUN", "XAIIMG-live-" + scene.ID})
	}
	for k, v := range map[string]string{"model_catalog_version": scope.CatalogVersion, "model_catalog_hash": scope.CatalogHash, "live_combo": xaiimages.Route, "live_sdk": xaiimages.SDK, "xai_images_source": scope.SourceSHA256} {
		run.SetVersion(k, v)
	}
	if err != nil || scopeErr != nil {
		p.problem = "candidate execution identity unavailable"
		return
	}
	if os.Getenv(envCombo) != xaiimages.Route {
		p.notRun = "explicit Imagine route absent"
		return
	}
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value == "" {
			continue
		}
		if name != keyVar(xaiimages.Route) && (strings.HasSuffix(name, "_KEY") || strings.HasSuffix(name, "_TOKEN") || strings.HasSuffix(name, "_SECRET") || strings.HasPrefix(name, envKeyBase)) || strings.HasSuffix(name, "_BASE_URL") || name == "BARNESS_AI_LIVE_ENDPOINT" || name == "OPENAI_API_BASE" || name == "BARNESS_AI_LIVE_PROXY" || name == "BARNESS_REGIONAL_ROUTE" || name == "BARNESS_GLM_ROUTE" || name == "BARNESS_MINIMAX_ROUTE" || name == "BARNESS_MM_ROUTE" {
			p.problem = "mixed credentials, routes or target overrides forbidden"
			return
		}
	}
	if p.controlled != "" && p.controlled != "fixture" && p.controlled != "cancel" && p.controlled != "final-write" {
		p.problem = "unknown controlled mode"
		return
	}
	if p.controlled != "" && (os.Getenv(envLive) != "" || os.Getenv(keyVar(xaiimages.Route)) != "") {
		p.problem = "controlled mode forbids live permission and credentials"
		return
	}
	if p.controlled == "" && os.Getenv(envLive) != "1" {
		p.notRun = "explicit live permission absent"
		return
	}
	if p.controlled == "" && os.Getenv(keyVar(xaiimages.Route)) == "" {
		p.notRun = "selected credential absent"
		return
	}
	path := os.Getenv("BARNESS_AI_LIVE_PERMISSIONS_FILE")
	if path == "" {
		p.notRun = "fresh one-process grant absent"
		return
	}
	p.authorization, err = loadProcessAuthorization(xaiimages.Route, mode, "candidate", os.Getenv(envAlias), os.Getenv("BARNESS_AI_LIVE_REGION_ALIAS"), path)
	if err != nil {
		p.problem = "fixed identity, source, binary, aliases or finite grant differs"
		return
	}
	if flag.Lookup("test.count").Value.String() != "1" || p.controlled == "" && flag.Lookup("test.timeout").Value.String() != "20m0s" {
		p.problem = "exact one-shot count and 1200s live process limit required"
		return
	}
	p.key = os.Getenv(keyVar(xaiimages.Route))
	if p.controlled != "" {
		p.key = "synthetic-xai-images-controlled"
	}
	run.RedactSecret(p.key, "[XAI-IMAGES-CREDENTIAL]")
}
