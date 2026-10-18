//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

var minimaxReport *minimaxlive.Report
var minimaxConfigError string
var minimaxSelected *combo
var minimaxCatalog ai.Catalog
var minimaxKey string
var minimaxLedger minimaxlive.Ledger

func initializeMiniMax() {
	id, e := minimaxlive.Select(os.Getenv("BARNESS_MINIMAX_ROUTE"), os.Getenv("BARNESS_MINIMAX_MODEL"), os.Getenv("BARNESS_MINIMAX_PURPOSE"))
	r := &minimaxlive.Report{Schema: 1, Kind: "minimax-candidate-only", Mode: "live", Stage: "candidate", Identity: id, Outcome: "NOT_RUN", StartedAt: started, Deadline: started.Add(1800 * time.Second), GitCommit: evidence.GitCommit(), SourceHash: minimaxlive.SourcePin, Unknown: []string{"known usage is not an invoice; cancellation may consume unreported tokens", "conservative input reservation uses complete synthetic request bytes +1024, minimum2048; no remote token estimator", "cache-hit and capacity/quality are not guaranteed"}}
	r.Session = filepath.Base(run.Dir())
	minimaxReport = r
	if e != nil {
		minimaxConfigError = e.Error()
	}
	if os.Getenv("BARNESS_MINIMAX_CONTROLLED") != "" {
		r.Mode = "controlled"
	}
	if os.Getenv("BARNESS_MINIMAX_REPLAY") != "" {
		r.Mode = "replay"
		r.Kind = "minimax-capture-replay-only"
	}
	if os.Getenv("BARNESS_MINIMAX_FORMAL") == "1" {
		r.Stage = "formal"
		r.Kind = "minimax-formal"
	}
	if r.Mode != "live" && r.Kind == "minimax-formal" {
		r.Kind = "minimax-nonlive-formal-rejected"
		minimaxConfigError = "only actual vendor live can be formal"
	}
	if os.Getenv("BARNESS_MINIMAX_RESEARCH") == "1" {
		if r.Stage != "candidate" || r.Mode == "replay" {
			minimaxConfigError = "research cannot combine formal or replay modes"
		}
		r.Kind = "minimax-research-only"
	}
	minimaxKey = strings.TrimSpace(os.Getenv("BARNESS_MINIMAX_KEY"))
	if minimaxKey != "" {
		run.RedactSecret(minimaxKey, "[MINIMAX-CREDENTIAL]")
	}
	r.AccountAlias = os.Getenv("BARNESS_MINIMAX_ACCOUNT_ALIAS")
	r.RegionAlias = os.Getenv("BARNESS_MINIMAX_REGION_ALIAS")
	r.Entitlement = os.Getenv("BARNESS_MINIMAX_ENTITLEMENT")
	alias := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	if minimaxKey != "" && (!alias.MatchString(r.AccountAlias) || !alias.MatchString(r.RegionAlias)) {
		minimaxConfigError = "nonsecret account and region aliases required"
	}
	for _, kv := range os.Environ() {
		n, v, _ := strings.Cut(kv, "=")
		if v == "" {
			continue
		}
		if n == "BARNESS_GLM_ROUTE" || n == "BARNESS_MM_ROUTE" || n == envCombo || strings.HasPrefix(n, envKeyBase) || strings.HasSuffix(n, "_BASE_URL") || n == "BARNESS_AI_LIVE_ENDPOINT" || n == "OPENAI_API_BASE" || n == "BARNESS_AI_LIVE_PROXY" {
			minimaxConfigError = "mixed route, credential or endpoint override forbidden"
		}
		if n != "BARNESS_MINIMAX_KEY" && (strings.HasSuffix(n, "_KEY") || strings.HasSuffix(n, "_TOKEN") || strings.HasSuffix(n, "_SECRET")) {
			minimaxConfigError = "ambient vendor credential forbidden"
		}
	}
	if r.Mode == "replay" && (minimaxKey != "" || os.Getenv(envLive) == "1") {
		minimaxConfigError = "replay forbids credentials and live switch"
	}
	minimaxSelected = &combo{name: "minimax-cn-" + id.Route, operation: ai.OperationChat, provider: ai.ProviderMiniMaxCN, api: id.API, endpoint: id.Endpoint, model: id.Model, sdk: openAISDK}
	if id.Route == "messages" {
		minimaxSelected.sdk = anthropicSDK
	}
	r.SDK = minimaxSelected.sdk()
	minimaxCatalog = minimaxlive.CandidateCatalog(id)
	if r.Stage == "formal" {
		minimaxCatalog = ai.BuiltinCatalog()
		if id.Model != minimaxlive.M3 {
			minimaxConfigError = "M3.1 remains candidate"
		}
	}
	// Pass only this exact selected model to the Host assembly. Formal metadata
	// still refers to the complete built-in snapshot, never synthetic candidate prices.
	if _, ok := minimaxCatalog.Lookup(ai.ProviderMiniMaxCN, id.API, id.Model); !ok {
		minimaxConfigError = "selected model absent from exact catalog"
	}
	r.CatalogVersion = minimaxCatalog.Version
	r.CatalogHash, _ = minimaxCatalog.Hash()
	r.Expected = minimaxlive.Plan(id)
	if r.Kind == "minimax-research-only" {
		r.Expected = minimaxlive.ResearchPlan()
	}
	raw, _ := json.Marshal(r.Expected)
	r.PlanHash = minimaxlive.Hash(raw)
	r.ExecutionSource, e = minimaxlive.ExecutionSource()
	if e != nil {
		minimaxConfigError = "execution source unavailable"
	}
	if _, e = minimaxlive.Sources(); e != nil {
		minimaxConfigError = e.Error()
	}
	path := os.Getenv("BARNESS_MINIMAX_LEDGER")
	if r.Mode == "live" {
		canonical, e := minimaxlive.StageLedgerPath(r.Stage)
		if e != nil {
			minimaxConfigError = e.Error()
		} else if path != "" && path != canonical {
			minimaxConfigError = "actual live requires common-Git durable stage ledger"
		} else {
			path = canonical
		}
	} else if path == "" {
		root, _ := minimaxlive.Root()
		path = filepath.Join(root, ".evidence", "minimax-controlled-stages", r.Stage+".jsonl")
	}

	minimaxLedger = minimaxlive.Ledger{Path: path, Stage: r.Stage, Session: r.Session, Identity: id}
	run.SetVersion("minimax_mode", r.Mode)
	run.SetVersion("minimax_kind", r.Kind)
	run.SetVersion("model_catalog_version", r.CatalogVersion)
	run.SetVersion("model_catalog_hash", r.CatalogHash)
	run.SetVersion("minimax_plan_hash", r.PlanHash)
	run.SetVersion("minimax_source_hash", r.SourceHash)
}
func minimaxEntry() string {
	if minimaxReport.Kind == "minimax-research-only" {
		return "^TestMiniMaxResearch$"
	}
	if minimaxReport.Mode == "replay" {
		return "^TestMiniMaxReplay$"
	}
	if minimaxReport.Stage == "formal" {
		return "^TestMiniMaxFormal$"
	}
	return "^TestMiniMaxCandidate$"
}
func minimaxFail(message string) { minimaxReport.Outcome = "FAIL"; minimaxReport.Failure = message }
func miniMaxBudgetError(e error) error {
	if e != nil {
		return fmt.Errorf("MiniMax durable reservation rejected: %w", e)
	}
	return nil
}
