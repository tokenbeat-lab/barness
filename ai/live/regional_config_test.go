//go:build live

package live

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type regionalProcess struct {
	report                                      supportmatrix.Report
	cat                                         ai.Catalog
	grant                                       supportmatrix.RegionalGrant
	grantRaw                                    []byte
	grantPath, key, problem, notRun, controlled string
	source                                      map[string]string
}

var regionalState *regionalProcess

// Enabled only by the separate controlled fixture build tag; never live.
var regionalBaselineFixture bool

func initializeRegional() {
	p := &regionalProcess{}
	regionalState = p
	combo := os.Getenv("BARNESS_REGIONAL_ROUTE")
	route, ok := supportmatrix.RegionalRouteOf(combo)
	if !ok {
		p.problem = "unknown explicit regional route"
		return
	}
	stage := os.Getenv("BARNESS_REGIONAL_STAGE")
	if stage == "" {
		stage = "candidate"
	}
	if stage != "candidate" && stage != "formal" {
		p.problem = "regional stage must be explicit candidate or formal"
		return
	}
	p.controlled = os.Getenv("BARNESS_REGIONAL_CONTROLLED")
	mode := "live"
	if p.controlled != "" {
		mode = "controlled"
	}
	p.cat, _ = supportmatrix.RegionalCandidateCatalog(combo)
	if stage == "formal" {
		var err error
		p.cat, err = supportmatrix.RegionalFormalCatalog(combo)
		if err != nil {
			p.problem = "regional formal model remains unpromoted"
			return
		}
	}
	hash, err := p.cat.Hash()
	if err != nil {
		p.problem = "regional fixed catalog invalid"
		return
	}
	for _, model := range supportmatrix.RegionalModels(combo) {
		if _, ok := p.cat.Lookup(ai.ProviderID(route.Provider), ai.API(route.API), model); !ok {
			p.problem = "regional formal model remains unpromoted"
			return
		}
	}
	sdk, err := supportmatrix.RegionalSDK(combo)
	if err != nil {
		p.problem = "regional SDK unavailable"
		return
	}
	p.source, err = supportmatrix.RegionalExecutionSource()
	if err != nil {
		p.problem = "regional execution source unavailable"
		return
	}
	p.report = supportmatrix.Report{Schema: supportmatrix.SchemaVersion, Combo: combo, Operation: "chat", Provider: route.Provider, API: route.API, Endpoint: route.Endpoint, Model: supportmatrix.RegionalModels(combo)[0], SDK: sdk, GitCommit: evidence.GitCommit(), CatalogVersion: p.cat.Version, CatalogHash: hash, StartedAt: started, Expected: supportmatrix.RegionalExpected(combo), Regional: &supportmatrix.RegionalEvidence{Schema: 1, Mode: mode, Stage: stage, Purpose: "standard", CredentialProvider: route.Provider, Contract: supportmatrix.RegionalAuthorization(combo), SourceSHA256: supportmatrix.RegionalSourceHash(p.source)}, Budget: supportmatrix.RegionalBudgetFor(combo)}
	p.report.Regional.NativeSourceCases = map[string]string{}
	for _, model := range supportmatrix.RegionalModels(combo) {
		p.report.Regional.NativeSourceCases[model] = supportmatrix.RegionalNativeSourceCase(combo, model)
	}
	binaryHash, e := regionalBinaryHash()
	if e != nil {
		p.problem = "regional executable identity unavailable"
		return
	}
	p.report.Regional.BinarySHA256 = binaryHash
	p.report.Regional.BuildSourceSHA256 = regionalBuildSource
	if mode == "live" && os.Getenv(envLive) == "1" && regionalBuildSource != p.report.Regional.SourceSHA256 {
		p.problem = "regional binary was not built from current explicitly frozen source"
		return
	}
	for _, id := range p.report.Expected {
		model, _, _ := strings.Cut(id, "/")
		p.report.Scenarios = append(p.report.Scenarios, supportmatrix.ScenarioResult{ID: id, Model: model, CaseID: supportmatrix.RegionalCaseID(combo, id), Outcome: supportmatrix.NotRun})
	}
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value == "" {
			continue
		}
		if name != "BARNESS_REGIONAL_KEY" && (strings.HasSuffix(name, "_KEY") || strings.HasSuffix(name, "_TOKEN") || strings.HasSuffix(name, "_SECRET")) || name == envCombo || name == "BARNESS_GLM_ROUTE" || name == "BARNESS_MINIMAX_ROUTE" || name == "BARNESS_MM_ROUTE" || strings.HasSuffix(name, "_BASE_URL") || name == "BARNESS_AI_LIVE_ENDPOINT" || name == "OPENAI_API_BASE" || name == "BARNESS_AI_LIVE_PROXY" {
			p.problem = "regional process forbids mixed routes, ambient credentials and endpoint overrides"
			return
		}
	}
	if p.controlled == "" && os.Getenv(envLive) != "1" {
		p.notRun = "new regional live switch absent"
		return
	}
	if p.controlled != "" && (os.Getenv(envLive) == "1" || os.Getenv("BARNESS_REGIONAL_KEY") != "") {
		p.problem = "controlled regional mode forbids live and credentials"
		return
	}
	p.grantPath = os.Getenv("BARNESS_REGIONAL_GRANT")
	if p.grantPath == "" {
		p.notRun = "new own regional authorization absent"
		return
	}
	p.grantRaw, err = readDecisionsFile(p.grantPath, 64<<10)
	if err == nil {
		err = json.Unmarshal(p.grantRaw, &p.grant)
	}
	if err == nil {
		if regionalBaselineFixture && mode == "controlled" && p.grant.MaxCalls == 9 {
			err = p.grant.CheckKimiBaselineFixture(combo, time.Now().UTC())
		} else {
			err = p.grant.Check(combo, time.Now().UTC())
		}
	}
	if err != nil || p.grant.Stage != stage || p.grant.Mode != mode {
		p.problem = "new own regional authorization invalid"
		return
	}
	if p.grant.Purpose == supportmatrix.KimiHistoryPurpose {
		p.report.Expected = []string{"kimi-k3/xai-origin-history"}
		p.report.Scenarios = []supportmatrix.ScenarioResult{{ID: p.report.Expected[0], Model: "kimi-k3", CaseID: supportmatrix.RegionalCaseID(combo, p.report.Expected[0]), Outcome: supportmatrix.NotRun}}
		p.report.Budget = supportmatrix.KimiHistoryBudget()
		p.report.Regional.Purpose, p.report.Regional.Contract = p.grant.Purpose, p.grant.Contract
		p.report.Regional.NativeSourceCases = map[string]string{}
	} else if regionalBaselineFixture && mode == "controlled" && p.grant.MaxCalls == 9 {
		p.report.Expected = p.report.Expected[:9]
		p.report.Scenarios = p.report.Scenarios[:9]
		p.report.Budget.MaxCalls, p.report.Budget.MaxAttempts, p.report.Budget.MaxRequestedOutputTokens = 9, 9, 36864
	}
	if os.Getenv("BARNESS_REGIONAL_SERVICE") != route.Provider {
		p.problem = "regional credential source Provider differs"
		return
	}
	if supportmatrix.RegionalIsKimi(combo) {
		if p.grant.BinarySHA256 != binaryHash || regionalBuildSource != p.report.Regional.SourceSHA256 {
			p.problem = "Kimi source/binary grant binding differs"
			return
		}
		p.report.Regional.MaxUSD = p.grant.MaxUSD
		p.report.Regional.CallIntervalMs = p.grant.CallIntervalMs
	}
	p.report.AccountAlias = p.grant.AccountAlias
	p.report.Regional.RegionAlias, p.report.Regional.GrantID, p.report.Regional.GrantSHA256 = p.grant.RegionAlias, p.grant.ID, sha256Text(p.grantRaw)
	if p.controlled != "" {
		p.key = "synthetic-regional-controlled"
	} else {
		// Read only this new exact service's explicitly authorized secret.
		p.key = os.Getenv("BARNESS_REGIONAL_KEY")
		if strings.TrimSpace(p.key) == "" {
			p.notRun = "selected regional credential absent"
			return
		}
	}
	run.RedactSecret(p.key, "[REGIONAL-CREDENTIAL]")
	for key, value := range map[string]string{"regional_mode": mode, "regional_stage": stage, "live_combo": combo, "live_sdk": sdk, "model_catalog_version": p.cat.Version, "model_catalog_hash": hash, "regional_execution_source": p.report.Regional.SourceSHA256} {
		run.SetVersion(key, value)
	}
}

func regionalEntry() string {
	if regionalState != nil {
		if regionalState.grant.Purpose == supportmatrix.KimiHistoryPurpose {
			return "^TestKimiHistorySupplement$"
		}
		if regionalBaselineFixture && regionalState.controlled != "" && regionalState.grant.MaxCalls == 9 {
			return "^TestKimiBaselineControlled$"
		}
	}
	if regionalState != nil && regionalState.report.Regional != nil && regionalState.report.Regional.Stage == "formal" {
		return "^TestRegionalFormal$"
	}
	return "^TestRegionalCandidate$"
}

func (p *regionalProcess) consumeGrant() error {
	receipt, err := supportmatrix.ReserveRegionalGrant(p.grant, p.grantRaw, filepath.Base(run.Dir()), time.Now().UTC())
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(run.Dir(), "regional-grant.json"), p.grantRaw, 0600); err != nil {
		return err
	}
	return run.Record("regional-authorization-receipt", receipt)
}
