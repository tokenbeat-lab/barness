//go:build live

package live

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Budget of one live process (spec §2, §5: bounded tokens and calls). A
// normal run of the largest combination (anthropic-messages) makes up to 20
// logical calls; the rest leaves room for a retry of the scenarios that
// meet a vendor fault.
const (
	maxCalls          = 32
	maxRetries        = 1
	maxOutputTokens   = 4096
	retryBackoff      = 5 * time.Second
	scenarioTimeLimit = 4 * time.Minute
)

var (
	run     *evidence.Run
	cfg     processConfig
	catalog = ai.BuiltinCatalog()
	started = time.Now().UTC()

	mu      sync.Mutex
	results []supportmatrix.ScenarioResult
	budget  = supportmatrix.Budget{MaxCalls: maxCalls, MaxRetries: maxRetries, MaxOutputTokens: maxOutputTokens}
)

func TestMain(m *testing.M) {
	var err error
	flag.Parse()
	// Select the isolated mode before legacy configuration can enable a route.
	// An isolated qualification process never executes a legacy live combination.
	if xaiImagesSelected() {
		cfg = processConfig{notRun: "isolated xAI Images candidate"}
	} else if xaiSelected() {
		cfg = processConfig{notRun: "isolated xAI qualification"}
	} else if os.Getenv("BARNESS_REGIONAL_ROUTE") != "" {
		cfg = processConfig{notRun: "isolated new regional authorization"}
	} else if glmDiagnosticSelected() {
		cfg = processConfig{notRun: "isolated one-case GLM diagnostic"}
	} else if os.Getenv("BARNESS_MINIMAX_ROUTE") != "" {
		cfg = processConfig{notRun: "isolated MiniMax identity"}
	} else if os.Getenv("BARNESS_GLM_ROUTE") != "" {
		cfg = processConfig{notRun: "isolated GLM candidate"}
	} else if os.Getenv("BARNESS_MM_ROUTE") != "" {
		cfg = processConfig{notRun: "isolated multimodal candidate process"}
	} else {
		cfg = loadConfig()
	}
	packageArgs := "-tags live ./ai/live"
	if isDecisionsProcess() {
		packageArgs = "-tags live -timeout 180s ./ai/live"
	}
	if os.Getenv("BARNESS_GLM_ROUTE") != "" {
		packageArgs = "-tags live -timeout 1800s ./ai/live"
	}
	if xaiImagesSelected() || xaiSelected() {
		packageArgs = "-tags live -timeout " + flag.Lookup("test.timeout").Value.String() + " ./ai/live"
	}
	if os.Getenv("BARNESS_MINIMAX_ROUTE") != "" {
		// Record the actual process limit, including controlled/replay and
		// rejected inputs. An environment round ID grants no extra lifetime.
		packageArgs = "-tags live -timeout " + flag.Lookup("test.timeout").Value.String() + " ./ai/live"
	}
	if os.Getenv("BARNESS_REGIONAL_ROUTE") != "" {
		timeout, err := time.ParseDuration(flag.Lookup("test.timeout").Value.String())
		if err == nil {
			packageArgs = fmt.Sprintf("-tags live -timeout %ds ./ai/live", int(timeout/time.Second))
		}
	}

	if glmDiagnosticSelected() {
		packageArgs = "-tags live -timeout 180s ./ai/live"
	}
	if run, err = evidence.NewRun(packageArgs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if xaiImagesSelected() {
		initializeXAIImages()
	} else if xaiSelected() {
		initializeXAI()
	} else if os.Getenv("BARNESS_REGIONAL_ROUTE") != "" {
		initializeRegional()
	} else if glmDiagnosticSelected() {
		initializeGLM()
	} else if os.Getenv("BARNESS_MINIMAX_ROUTE") != "" {
		initializeMiniMax()
	} else if os.Getenv("BARNESS_GLM_ROUTE") != "" {
		initializeGLM()
	}
	if isDecisionsProcess() {
		decisionsReport = initialDecisionsReport()
	}
	if os.Getenv("BARNESS_MM_ROUTE") != "" && glmReportState == nil && minimaxReport == nil {
		multimodalReport = initialMultimodalReport()
		if key := os.Getenv("BARNESS_MM_KEY"); key != "" {
			run.RedactSecret(key, "[MULTIMODAL-CREDENTIAL]")
		}
	}
	if cfg.combo != nil && cfg.combo.operation == ai.OperationClassifier {
		budget = classifierBudget()
	}
	if cfg.combo != nil && cfg.combo.operation == ai.OperationImage {
		budget = imagesBudget()
		// Both routes passed their candidate probes and are now published.
		// Release smoke must use the actual versioned catalog and prices.
	}
	if cfg.key != "" {
		run.RedactSecret(cfg.key, "[LIVE-KEY:"+cfg.combo.name+"]")
	}
	measuredCatalog := catalog
	if xaiImagesState != nil {
		measuredCatalog = xaiImagesState.cat
	}
	if minimaxReport != nil {
		measuredCatalog = minimaxCatalog
	}
	if err := run.Record("model-catalog", measuredCatalog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	catalogHash, err := catalog.Hash()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if xaiImagesState == nil && xaiState == nil && glmReportState == nil && minimaxReport == nil && regionalState == nil {
		run.SetVersion("model_catalog_version", catalog.Version)
		run.SetVersion("model_catalog_hash", catalogHash)
	}
	if cfg.combo != nil {
		run.SetVersion("live_combo", cfg.combo.name)
		run.SetVersion("live_sdk", cfg.combo.sdk())
	}
	code := 0
	mmEntry := "^TestMultimodalCandidate$"
	if os.Getenv("BARNESS_MM_FORMAL") == "1" {
		mmEntry = "^TestMultimodalFormal$"
	}
	if xaiImagesState != nil && flag.Lookup("test.run").Value.String() != xaiImagesEntry() {
		xaiImagesState.problem = "exact isolated Imagine entry required"
		code = 1
	} else if xaiState != nil && flag.Lookup("test.run").Value.String() != xaiState.entry() {
		xaiState.report.Outcome, xaiState.report.Failure = "FAIL", "exact isolated xAI entry required"
		code = 1
	} else if regionalState != nil && (flag.Lookup("test.run").Value.String() != regionalEntry() || regionalState.grant.Purpose == supportmatrix.KimiHistoryPurpose && (flag.Lookup("test.timeout").Value.String() != "4m0s" || flag.Lookup("test.count").Value.String() != "1")) {
		regionalState.problem = "exact isolated regional entry required"
		code = 1
	} else if minimaxReport != nil && flag.Lookup("test.run").Value.String() != minimaxEntry() {
		minimaxFail("exact isolated entry selection required")
		code = 1
	} else if glmReportState != nil && flag.Lookup("test.run").Value.String() != glmSelectedEntry() {
		glmReportState.Outcome = "FAIL"
		glmReportState.Failure = "exact isolated entry selection required"
		code = 1
	} else if multimodalReport != nil && flag.Lookup("test.run").Value.String() != mmEntry {
		multimodalReport.Outcome = "FAIL"
		multimodalReport.Failure = "isolated candidate entry selection is required before any tests execute"
		code = 1
	} else {
		code = m.Run()
	}
	if cfg.combo != nil && !isDecisionsProcess() && multimodalReport == nil {
		if err := run.Record(supportmatrix.ReportName, report(catalogHash)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	if decisionsReport != nil {
		decisionsReport.FinishedAt = time.Now().UTC()
		if err := run.Record(decisionsDetailName(), decisionsReport); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	if glmReportState != nil {
		glmReportState.FinishedAt = time.Now().UTC()
		if err := run.Record(glmReportName, glmReportState); err != nil {
			code = 1
		}
	}
	if minimaxReport != nil {
		minimaxReport.FinishedAt = time.Now().UTC()
		if err := run.Record("minimax-report", minimaxReport); err != nil {
			code = 1
		}
	}
	if regionalState != nil {
		regionalState.report.FinishedAt = time.Now().UTC()
		if err := run.Record("regional-report", regionalState.report); err != nil {
			code = 1
		}
	}
	if xaiState != nil {
		xaiState.report.FinishedAt = time.Now().UTC()
		if err := run.Record(xaiState.reportName(), xaiState.report); err != nil {
			code = 1
		}
	}
	if xaiImagesState != nil {
		code = recordXAIImages(code, false)
	}
	if err := run.Finish(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	// The redaction audit forbids the key (registered above) and every
	// credential-looking variable of this process, the key's included.
	auditOK := true
	if findings, err := run.Audit(); err != nil || len(findings) > 0 {
		auditOK = false
		if err != nil {
			fmt.Fprintln(os.Stderr, "live: redaction audit:", err)
		}
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "live: redaction audit: %s [%s] %s\n", f.File, f.Rule, f.Detail)
		}
		code = 1
	}
	if xaiImagesState != nil {
		code = finishXAIImages(code, auditOK)
	}
	if xaiState != nil {
		code = finishXAI(code, auditOK)
	}
	if decisionsReport != nil {
		code = finishDecisionsReport(code, auditOK)
	}
	if multimodalReport != nil {
		code = finishMultimodalReport(code, auditOK)
	}
	if glmReportState != nil {
		code = finishGLM(code, auditOK)
		code = publishGLMFormal(code)
	}
	if minimaxReport != nil {
		code = finishMiniMax(code, auditOK)
	}
	if regionalState != nil {
		code = finishRegional(code, auditOK)
	}
	os.Exit(code)
}

// report is this process's result for its combination.
func report(catalogHash string) supportmatrix.Report {
	mu.Lock()
	defer mu.Unlock()
	c := cfg.combo
	var expected []string
	for _, sc := range scenariosOf(c) {
		expected = append(expected, sc.id)
	}
	return supportmatrix.Report{
		Schema: supportmatrix.SchemaVersion, Combo: c.name, Operation: string(c.operation), Provider: string(c.provider), API: string(c.api),
		Endpoint: c.endpoint, Model: c.model, SDK: c.sdk(), AccountAlias: cfg.alias,
		GitCommit: evidence.GitCommit(), CatalogVersion: catalog.Version, CatalogHash: catalogHash,
		StartedAt: started, FinishedAt: time.Now().UTC(), Budget: budget, Expected: expected, Scenarios: results,
	}
}

// TestLive runs every combination's scenarios. Only the combination this
// process was given a key for executes; the others report NOT_RUN.
func TestLive(t *testing.T) {
	if cfg.combo == nil && cfg.refused != "" {
		t.Fatal(cfg.refused)
	}
	for i := range combos {
		c := &combos[i]
		t.Run(c.name, func(t *testing.T) {
			selected := cfg.combo == c
			var env *liveEnv
			if selected && cfg.notRun == "" && cfg.refused == "" {
				env = newLiveEnv(t, c, cfg.key)
			}
			for _, sc := range scenariosOf(c) {
				t.Run(sc.id, func(t *testing.T) {
					cs := run.Case(t, "LIVE-"+c.spec+"-"+sc.id)
					setLiveReplay(cs, c)
					runScenario(t, cs, c, selected, env, sc)
				})
			}
		})
	}
}

// liveEnv is a selected combination's Client, assembled as a host would:
// one tenant, one binding to the vendor's real endpoint, one credential.
type liveEnv struct {
	combo  *combo
	client *ai.Client
	rec    *recorder
}

const liveTenant = "live-smoke"

func newLiveEnv(t *testing.T, c *combo, key string) *liveEnv {
	t.Helper()
	if len(supportmatrix.GenericExpected(c.name)) > 0 {
		if err := claimGenericProcess(); err != nil {
			t.Fatalf("fresh process admission: %v", err)
		}
	}
	rec := newRecorder()
	if cfg.proxy != nil {
		rec.next.(*http.Transport).Proxy = http.ProxyURL(cfg.proxy)
	}
	t.Cleanup(rec.next.(*http.Transport).CloseIdleConnections)
	client, err := newClient(c, key, rec, catalog)
	if err != nil {
		t.Fatalf("live: assembling the Client: %v", err)
	}
	return &liveEnv{combo: c, client: client, rec: rec}
}

// newClient assembles the Client on cat, allowing every model cat lists for
// the combination.
func newClient(c *combo, key string, rec *recorder, cat ai.Catalog) (*ai.Client, error) {
	return newClientWithPolicy(c, key, rec, cat, livePolicy())
}

// newClientWithPolicy shares only Host assembly, not the legacy runner's
// scenario selection, call budget or time limits with candidate evidence.
func newClientWithPolicy(c *combo, key string, rec *recorder, cat ai.Catalog, policy *ai.ResourcePolicy) (*ai.Client, error) {
	var models []string
	for _, m := range cat.Models {
		if m.Provider == c.provider && m.API == c.api {
			if c.provider != ai.ProviderMiniMax || m.ID == c.model {
				models = append(models, m.ID)
			}
		}
	}
	for _, m := range cat.ImageModels {
		if c.operation == ai.OperationImage && m.Provider == c.provider && m.API == c.api {
			models = append(models, m.ID)
		}
	}
	for _, m := range cat.ClassifierModels {
		if c.operation == ai.OperationClassifier && m.Provider == c.provider && m.API == c.api {
			models = append(models, m.ID)
		}
	}
	res := &resolver{
		binding: ai.Binding{TenantID: liveTenant, BindingID: c.name, Version: "live", Enabled: true,
			Operation: c.operation, ProviderID: c.provider, API: c.api, Endpoint: c.endpoint, AuthKind: ai.AuthAPIKey,
			AccountScopeID: "live-" + c.name, CredentialRef: "live-key", AllowedModels: models},
		credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: "live-key", Version: "live",
			BindingVersion: "live", AccountScopeID: "live-" + c.name, Active: true, APIKey: ai.NewSecret(key)},
	}
	return ai.NewClient(ai.Config{Policy: policy, Bindings: res, Credentials: res, Transport: rec, Observer: rec, Catalog: &cat})
}

// livePolicy bounds a smoke process: short turns, one call at a time.
// Test values, not deployment guidance.
func livePolicy() *ai.ResourcePolicy {
	return &ai.ResourcePolicy{
		Image:                 &ai.ImagePolicy{MaxInputImages: 2, MaxOutputImages: 1, MaxOutputImageBytes: 8 << 20, MaxTotalOutputImageBytes: 8 << 20},
		Classifier:            &ai.ClassifierPolicy{MaxQuestions: 3, MaxStateBytes: 131072, MaxQuestionBytes: 16384},
		MaxQueuedObservations: 128,
		MaxRequestBytes:       1 << 20, MaxImageBytes: 1 << 19, MaxFrameBytes: 1 << 20, MaxToolJSONBytes: 1 << 16,
		MaxErrorBodyBytes: 1 << 16, MaxOutputBytes: 16 << 20, MaxQueuedEvents: 1 << 14, MaxQueuedEventBytes: 8 << 20,
		MaxConcurrentPerTenant: 2, MaxConcurrentProcess: 2, MaxAdmissionWaiters: 2, AdmissionWait: time.Second,
		CallTimeout: 3 * time.Minute, ConnectTimeout: 15 * time.Second,
		ResponseHeaderTimeout: 90 * time.Second, ReadIdleTimeout: 90 * time.Second,
	}
}

// resolver is the live tenant's one binding and credential.
type resolver struct {
	binding    ai.Binding
	credential ai.Credential
}

var errNotFound = errors.New("live: not found")

func (r *resolver) ResolveBinding(_ context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
	if scope.TenantID != liveTenant || id != r.binding.BindingID {
		return ai.Binding{}, errNotFound
	}
	return r.binding, nil
}

func (r *resolver) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	if scope.TenantID != liveTenant || b.CredentialRef != r.credential.CredentialID {
		return ai.Credential{}, errNotFound
	}
	return r.credential, nil
}

func newScope() ai.CallScope {
	return ai.CallScope{TenantID: liveTenant, RequestID: requestid.New(), JobID: "live-smoke"}
}

func setLiveReplay(cs *evidence.Case, c *combo) {
	cs.ReplayEnv(envLive + "=1 " + envCombo + "=" + c.name)
	if c.operation == ai.OperationImage {
		cs.ReplayRun("^TestLive$/^" + c.name + "$")
	}
}
