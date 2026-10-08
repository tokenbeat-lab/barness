//go:build live

package live

import (
	"context"
	"errors"
	"fmt"
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
	if run, err = evidence.NewRun("-tags live ./ai/live"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg = loadConfig()
	if cfg.combo != nil && cfg.combo.operation == ai.OperationClassifier {
		budget = classifierBudget()
	}
	if cfg.key != "" {
		run.RedactSecret(cfg.key, "[LIVE-KEY:"+cfg.combo.name+"]")
	}
	catalogHash, err := catalog.Hash()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	run.SetVersion("model_catalog_version", catalog.Version)
	run.SetVersion("model_catalog_hash", catalogHash)
	if cfg.combo != nil {
		run.SetVersion("live_combo", cfg.combo.name)
		run.SetVersion("live_sdk", cfg.combo.sdk())
	}
	code := m.Run()
	if cfg.combo != nil {
		if err := run.Record(supportmatrix.ReportName, report(catalogHash)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	if err := run.Finish(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	// The redaction audit forbids the key (registered above) and every
	// credential-looking variable of this process, the key's included.
	if findings, err := run.Audit(); err != nil || len(findings) > 0 {
		if err != nil {
			fmt.Fprintln(os.Stderr, "live: redaction audit:", err)
		}
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "live: redaction audit: %s [%s] %s\n", f.File, f.Rule, f.Detail)
		}
		code = 1
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
					cs.ReplayEnv(envLive + "=1 " + envCombo + "=" + c.name)
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
	rec := newRecorder()
	client, err := newClient(c, key, rec, catalog)
	if err != nil {
		t.Fatalf("live: assembling the Client: %v", err)
	}
	return &liveEnv{combo: c, client: client, rec: rec}
}

// newClient assembles the Client on cat, allowing every model cat lists for
// the combination.
func newClient(c *combo, key string, rec *recorder, cat ai.Catalog) (*ai.Client, error) {
	var models []string
	for _, m := range cat.Models {
		if m.Provider == c.provider && m.API == c.api {
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
	return ai.NewClient(ai.Config{Policy: livePolicy(), Bindings: res, Credentials: res, Transport: rec, Observer: rec, Catalog: &cat})
}

// livePolicy bounds a smoke process: short turns, one call at a time.
// Test values, not deployment guidance.
func livePolicy() *ai.ResourcePolicy {
	return &ai.ResourcePolicy{
		Classifier:            &ai.ClassifierPolicy{MaxQuestions: 3, MaxStateBytes: 131072, MaxQuestionBytes: 16384},
		MaxQueuedObservations: 128,
		MaxRequestBytes:       1 << 20, MaxImageBytes: 1 << 19, MaxFrameBytes: 1 << 20, MaxToolJSONBytes: 1 << 16,
		MaxErrorBodyBytes: 1 << 16, MaxOutputBytes: 8 << 20, MaxQueuedEvents: 1 << 14, MaxQueuedEventBytes: 8 << 20,
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
