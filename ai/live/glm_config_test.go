//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const glmSourcePin = "837a83c4711dc4914b8110341403c0d6e43cac254b3101f601d6d538b209c907"

var glmReportName = "candidate-glm-report"

const glmEntry = "^TestGLMCandidate$"

type glmCasePlan struct {
	ID, Model, Kind, Baseline, OnlyChangedField string
	Value                                       any
}
type glmPlan struct {
	Schema  int
	Prompts map[string]string
	Routes  map[string][]glmCasePlan
}
type glmSources struct {
	Schema                      int
	Sources                     []candidateSource
	CandidateSHA256, PlanSHA256 string
	Provenance                  map[string]any
}
type glmBudget struct {
	ReplayAttempts             int `json:"replayAttempts"`
	PlannedCalls               int `json:"plannedCalls"`
	MaxCalls                   int `json:"maxCalls"`
	MaxAttempts                int `json:"maxAttempts"`
	MaxOutputTokens            int `json:"maxOutputTokens"`
	CallSeconds                int `json:"callSeconds"`
	ProcessSeconds             int `json:"processSeconds"`
	FinalizationReserveSeconds int `json:"finalizationReserveSeconds"`
	CallsUsed                  int `json:"callsUsed"`
	AttemptsReserved           int `json:"attemptsReserved"`
	HTTPAttempts               int `json:"httpAttempts"`
	Retries                    int `json:"retries"`
}
type glmOutcome struct {
	ReplayAttempts     int                 `json:"replayAttempts"`
	ID                 string              `json:"id"`
	Model              string              `json:"model"`
	Outcome            string              `json:"outcome"`
	Failure            string              `json:"failure"`
	Calls              int                 `json:"calls"`
	HTTPAttempts       int                 `json:"httpAttempts"`
	DurationMS         int64               `json:"durationMs"`
	ProviderRequestIDs []string            `json:"providerRequestIds"`
	RequestIDSources   []string            `json:"requestIdSources"`
	RequestID          string              `json:"requestId"`
	Usage              ai.Usage            `json:"usage"`
	UsageReporting     []ai.UsageReporting `json:"usageReporting"`
	RawUsagePresence   []json.RawMessage   `json:"rawUsagePresence"`
	Truncated          bool                `json:"truncated"`
}
type glmReport struct {
	OriginalLiveGitCommit    string            `json:"originalLiveGitCommit,omitempty"`
	OriginalLiveReportSHA256 string            `json:"originalLiveReportSHA256,omitempty"`
	CaptureArchiveSHA256     string            `json:"captureArchiveSHA256,omitempty"`
	Schema                   int               `json:"schema"`
	Kind                     string            `json:"kind"`
	Outcome                  string            `json:"outcome"`
	Complete                 bool              `json:"complete"`
	AuditPassed              bool              `json:"auditPassed"`
	Route                    string            `json:"route"`
	API                      string            `json:"api"`
	Endpoint                 string            `json:"endpoint"`
	SDK                      string            `json:"sdk"`
	CatalogVersion           string            `json:"catalogVersion"`
	CatalogHash              string            `json:"catalogHash"`
	SourceManifestHash       string            `json:"sourceManifestHash"`
	GitCommit                string            `json:"gitCommit"`
	Mode                     string            `json:"mode"`
	AccountAlias             string            `json:"accountAlias"`
	RegionAlias              string            `json:"regionAlias"`
	AccountVerification      string            `json:"accountVerification"`
	RegionVerification       string            `json:"regionVerification"`
	PrivilegeVerification    string            `json:"privilegeVerification"`
	StartedAt                time.Time         `json:"startedAt"`
	FinishedAt               time.Time         `json:"finishedAt"`
	ProcessDeadline          time.Time         `json:"processDeadline"`
	ExecutionDeadline        time.Time         `json:"executionDeadline"`
	Budget                   glmBudget         `json:"budget"`
	Expected                 []string          `json:"expected"`
	Scenarios                []glmOutcome      `json:"scenarios"`
	Models                   map[string]string `json:"models"`
	Failure                  string            `json:"failure"`
	Unknown                  []string          `json:"unknown"`
}

var glmReportState *glmReport
var glmSelected *combo
var glmCandidateCatalog ai.Catalog
var glmFixedPlan glmPlan
var glmKey string
var glmConfigError string

func initializeGLM() {
	route := os.Getenv("BARNESS_GLM_ROUTE")
	glmReportState = &glmReport{Schema: 1, Kind: "glm-candidate-only", Outcome: "NOT_RUN", Route: route, StartedAt: started, GitCommit: evidence.GitCommit(), SourceManifestHash: glmSourcePin, Mode: "live", AccountAlias: "local-glm", RegionAlias: "cn-standard-endpoint", AccountVerification: "unknown", RegionVerification: "unknown", PrivilegeVerification: "unknown", Models: map[string]string{}, Budget: glmBudget{PlannedCalls: 22, MaxCalls: 24, MaxAttempts: 32, MaxOutputTokens: 4096, CallSeconds: 120, ProcessSeconds: 1800, FinalizationReserveSeconds: 20}, Unknown: []string{"vendor account/region/permissions/billing", "capacity load not tested", "storage/time charges", "acceptance does not establish intensity"}}
	if os.Getenv("BARNESS_GLM_REPLAY_ARCHIVE") != "" {
		glmReportState.Mode = "replay"
		glmReportState.Kind = "glm-capture-replay-only"
		glmReportName = "capture-glm-replay-report"
	}
	glmKey = strings.TrimSpace(os.Getenv("BARNESS_GLM_KEY"))
	if glmReportState.Mode == "replay" && (glmKey != "" || os.Getenv("BARNESS_AI_LIVE") == "1") {
		glmConfigError = "replay forbids live switch and vendor credentials"
	}
	if glmKey != "" {
		run.RedactSecret(glmKey, "[GLM-CREDENTIAL]")
	}
	glmSelected = &combo{name: "zhipu-" + route + "-candidate", operation: ai.OperationChat, provider: ai.ProviderZhipu, sdk: openAISDK}
	switch route {
	case "chat":
		glmSelected.api = ai.APIOpenAICompletions
		glmSelected.endpoint = glm.ChatEndpoint
	case "responses":
		glmSelected.api = ai.APIOpenAIResponses
		glmSelected.endpoint = glm.ResponsesEndpoint
	case "messages":
		glmSelected.api = ai.APIAnthropicMessages
		glmSelected.endpoint = glm.MessagesEndpoint
		glmSelected.sdk = anthropicSDK
	default:
		glmConfigError = "unknown GLM candidate route"
	}
	r := glmReportState
	r.API = string(glmSelected.api)
	r.Endpoint = glmSelected.endpoint
	r.SDK = glmSelected.sdk()
	for _, kv := range os.Environ() {
		n, v, _ := strings.Cut(kv, "=")
		if v == "" {
			continue
		}
		if strings.HasPrefix(n, envKeyBase) || n == envCombo || n == "BARNESS_MM_ROUTE" || strings.HasSuffix(n, "_BASE_URL") || n == "BARNESS_AI_LIVE_PROXY" || n == "BARNESS_AI_LIVE_ENDPOINT" || n == "OPENAI_API_BASE" {
			glmConfigError = "mixed route, credential or endpoint override forbidden"
		}
		if n != "BARNESS_GLM_KEY" && (strings.HasSuffix(n, "_KEY") || strings.HasSuffix(n, "_TOKEN") || strings.HasSuffix(n, "_SECRET")) {
			glmConfigError = "ambient vendor credential forbidden"
		}
	}
	// Independent fixed source files; never load dotenv or alternate catalog paths.
	read := func(name string) ([]byte, error) {
		return readDecisionsFile(filepath.Join("testdata", "glm-candidate", name), 2<<20)
	}
	raw, err := read("sources.json")
	var src glmSources
	if err != nil || sha256Text(raw) != glmSourcePin || json.Unmarshal(raw, &src) != nil {
		glmConfigError = "fixed source manifest unavailable or changed"
		return
	}
	for _, s := range src.Sources {
		raw, err = read(s.File)
		if err != nil || sha256Text(raw) != s.SHA256 {
			glmConfigError = "fixed source snapshot unavailable or changed"
			return
		}
	}
	raw, err = read("candidate.json")
	if err != nil || sha256Text(raw) != src.CandidateSHA256 || json.Unmarshal(raw, &glmCandidateCatalog) != nil {
		glmConfigError = "fixed candidate snapshot unavailable or changed"
		return
	}
	raw, err = read("plan.json")
	if err != nil || sha256Text(raw) != src.PlanSHA256 || json.Unmarshal(raw, &glmFixedPlan) != nil {
		glmConfigError = "fixed scenario plan unavailable or changed"
		return
	}
	if len(glmCandidateCatalog.Models) != 6 || len(glmFixedPlan.Routes[route]) != 22 {
		glmConfigError = "fixed six-candidate/22-call identity set required"
		return
	}
	if glmDiagnosticSelected() {
		if err := selectGLMDiagnostic(); err != nil {
			glmConfigError = err.Error()
			return
		}
	} else if os.Getenv("BARNESS_GLM_FORMAL") == "1" {
		if err := selectGLMFormalCatalog(); err != nil {
			glmConfigError = err.Error()
			return
		}
	}
	r.CatalogVersion = glmCandidateCatalog.Version
	r.CatalogHash, _ = glmCandidateCatalog.Hash()
	run.SetVersion("model_catalog_version", r.CatalogVersion)
	run.SetVersion("model_catalog_hash", r.CatalogHash)
	run.SetVersion("glm_source_manifest_hash", glmSourcePin)
	run.SetVersion("glm_kind", r.Kind)
	run.SetVersion("glm_mode", r.Mode)
	run.SetVersion("live_combo", glmSelected.name)
	run.SetVersion("live_sdk", r.SDK)
	for _, sc := range glmFixedPlan.Routes[route] {
		r.Expected = append(r.Expected, sc.ID)
		r.Scenarios = append(r.Scenarios, glmOutcome{ID: sc.ID, Model: sc.Model, Outcome: "NOT_RUN"})
		r.Models[sc.Model] = "NOT_RUN"
	}
}

func glmFail(t interface{ Errorf(string, ...any) }, why string) {
	glmReportState.Outcome = "FAIL"
	glmReportState.Failure = why
	t.Errorf("GLM candidate: %s", why)
}
func glmReserve() error {
	b := &glmReportState.Budget
	if b.CallsUsed >= b.MaxCalls || b.AttemptsReserved >= b.MaxAttempts || b.HTTPAttempts >= b.MaxAttempts {
		return fmt.Errorf("fixed cumulative reservation exhausted")
	}
	b.CallsUsed++
	b.AttemptsReserved++
	return nil
}
