//go:build live

package live

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Candidate evidence deliberately has no supportmatrix.Report schema. It can
// only qualify this fixed candidate; inclusion and formal Hash live are issue 10.
type candidateDecisionsReport struct {
	FailureCategory    string                         `json:"failureCategory,omitempty"`
	FailureNote        string                         `json:"failureNote,omitempty"`
	ProcessDeadline    time.Time                      `json:"processDeadline,omitzero"`
	ExecutionDeadline  time.Time                      `json:"executionDeadline,omitzero"`
	Schema             int                            `json:"schema"`
	Kind               string                         `json:"kind"`
	Mode               string                         `json:"mode"`
	Outcome            supportmatrix.Outcome          `json:"outcome"`
	Complete           bool                           `json:"complete"`
	AuditPassed        bool                           `json:"auditPassed"`
	Combo              string                         `json:"combo"`
	Operation          ai.Operation                   `json:"operation"`
	Provider           ai.ProviderID                  `json:"provider"`
	API                ai.API                         `json:"api"`
	Endpoint           string                         `json:"endpoint"`
	Model              string                         `json:"model"`
	SDK                string                         `json:"sdk"`
	CatalogVersion     string                         `json:"candidateVersion"`
	CatalogHash        string                         `json:"candidateHash"`
	SourceManifestHash string                         `json:"sourceManifestHash"`
	Sources            candidateSources               `json:"sources"`
	Permissions        candidatePermissions           `json:"hostPermissionVerification"`
	PermissionsHash    string                         `json:"hostPermissionVerificationHash,omitempty"`
	GitCommit          string                         `json:"gitCommit"`
	StartedAt          time.Time                      `json:"startedAt"`
	FinishedAt         time.Time                      `json:"finishedAt"`
	Budget             candidateBudget                `json:"budget"`
	Expected           []string                       `json:"expected"`
	Scenarios          []supportmatrix.ScenarioResult `json:"scenarios"`
	Unknown            []string                       `json:"unknown"`
}
type candidateBudget struct {
	FinalizationReserveSeconds int `json:"finalizationReserveSeconds"`
	MaxCalls                   int `json:"maxCalls"`
	MaxAttempts                int `json:"maxAttempts"`
	MaxQuestions               int `json:"maxQuestions"`
	MaxStateBytes              int `json:"maxStateBytesTotal"`
	MaxImages                  int `json:"maxImages"`
	MaxImageBytes              int `json:"maxDecodedImageBytesTotal"`
	ProcessSeconds             int `json:"processSeconds"`
	CallSeconds                int `json:"callSeconds"`
	CallsUsed                  int `json:"callsUsed"`
	AttemptsReserved           int `json:"attemptsReserved"`
	HTTPAttempts               int `json:"httpAttempts"`
	QuestionsUsed              int `json:"questionsUsed"`
	StateBytesSubmitted        int `json:"stateBytesSubmitted"`
	ImagesUsed                 int `json:"imagesUsed"`
	ImageBytesSubmitted        int `json:"decodedImageBytesSubmitted"`
	Retries                    int `json:"retries"`
}

var candidateReport *candidateDecisionsReport

func initialCandidateReport() *candidateDecisionsReport {
	expected := []string{"text-choice", "text-score", "text-bool", "mixed-questions", "image-png", "image-jpeg", "image-webp", "cancel-in-flight", "recovery"}
	out := &candidateDecisionsReport{Schema: 1, Kind: "candidate-only", Mode: "live", Outcome: supportmatrix.NotRun, Combo: decisionsCandidate.name, Operation: ai.OperationClassifier, Provider: ai.ProviderOpenAI, API: ai.APIOpenAIDecisions, Endpoint: decisionsCandidate.endpoint, Model: decisionsCandidate.model, SDK: decisionsCandidate.sdk(), StartedAt: started, GitCommit: evidence.GitCommit(), SourceManifestHash: candidateSourceHash, Expected: expected,
		Budget:  candidateBudget{FinalizationReserveSeconds: 10, MaxCalls: 12, MaxAttempts: 12, MaxQuestions: 24, MaxStateBytes: 4 << 20, MaxImages: 8, MaxImageBytes: 4 << 20, ProcessSeconds: 180, CallSeconds: 30},
		Unknown: []string{"vendor billing", "cache billing", "regional processing premium", "922000 token service capacity", "128 image service capacity", "service score maximum", "alias weights are not frozen by catalog hash"}}
	for _, id := range expected {
		out.Scenarios = append(out.Scenarios, supportmatrix.ScenarioResult{ID: id, CaseID: "P10-candidate-" + id, Model: out.Model, Outcome: supportmatrix.NotRun, Note: "scenario not executed in this process"})
	}
	return out
}
func TestCandidateDecisions(t *testing.T) {
	if cfg.combo != &decisionsCandidate {
		t.Skip("candidate combination is not selected")
	}
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	driveCandidateDecisions(t, rec, decisionsScenarios, "live")
}

type decisionsScenario struct {
	id      string
	request ai.ClassifierRequest
	cancel  bool
}
type candidateSession struct {
	ctx    context.Context
	client *ai.Client
	rec    *recorder
	report *candidateDecisionsReport
}

func driveCandidateDecisions(t *testing.T, rec *recorder, scenarioFactory func(*testing.T) []decisionsScenario, mode string) {
	t.Helper()
	candidateReport.Mode = mode
	rec.requestLimit = 8 << 20
	sourcesCase := run.Case(t, "P10-candidate-sources")
	setCandidateReplay(sourcesCase, mode)
	cat, sources, sourceErr := loadDecisionsCandidate(sourcesCase)
	candidateReport.Sources = sources
	if sourceErr == nil {
		candidateReport.CatalogVersion = cat.Version
		candidateReport.CatalogHash, _ = cat.Hash()
		run.SetVersion("model_catalog_version", cat.Version)
		run.SetVersion("model_catalog_hash", candidateReport.CatalogHash)
		run.SetVersion("candidate_source_manifest_hash", candidateSourceHash)
	}
	reason := cfg.refused
	if err := candidateEnvironment(); err != nil {
		reason = err.Error()
	}
	if sourceErr != nil {
		reason = sourceErr.Error()
	}
	if reason != "" {
		candidateConfigFail(t, reason)
		return
	}
	if cfg.notRun != "" {
		for i := range candidateReport.Scenarios {
			candidateReport.Scenarios[i].Note = cfg.notRun
		}
		t.Skip("NOT_RUN: " + cfg.notRun)
	}
	p, permissionsRaw, err := candidateConfiguration(rec)
	candidateReport.Permissions = p
	candidateReport.PermissionsHash = sha256Text(permissionsRaw)
	if err != nil {
		candidateConfigFail(t, err.Error())
		return
	}
	sourcesCase.Fixture("host-permissions.json", permissionsRaw)
	policy := livePolicy()
	policy.Image = nil
	policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 24, MaxStateBytes: 4 << 20, MaxQuestionBytes: 16384, Images: &ai.ClassifierImagePolicy{MaxImages: 8, MaxTotalImageBytes: 4 << 20}}
	policy.MaxImageBytes = 4 << 20
	policy.MaxRequestBytes = 8 << 20
	policy.MaxOutputBytes = 1 << 20
	policy.CallTimeout = 30 * time.Second
	policy.ConnectTimeout = 10 * time.Second
	policy.ResponseHeaderTimeout = 25 * time.Second
	policy.ReadIdleTimeout = 25 * time.Second
	r := &resolver{binding: ai.Binding{TenantID: liveTenant, BindingID: decisionsCandidate.name, Version: "candidate", Enabled: true, Operation: ai.OperationClassifier, ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIDecisions, Endpoint: decisionsCandidate.endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: p.AccountAlias + "@" + p.RegionAlias, CredentialRef: "decisions-test-key", AllowedModels: []string{"gpt-6-luna"}, AllowClassifierImages: true},
		credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: "decisions-test-key", Version: "candidate", BindingVersion: "candidate", AccountScopeID: p.AccountAlias + "@" + p.RegionAlias, Active: true, APIKey: ai.NewSecret(cfg.key)}}
	// Binding.Retry zero: each reserved Logical Call has at most one Attempt.
	client, err := ai.NewClient(ai.Config{Policy: policy, Bindings: r, Credentials: r, Catalog: &cat, Transport: rec, Observer: rec})
	if err != nil {
		candidateConfigFail(t, "candidate Client assembly failed: "+err.Error())
		return
	}
	// Include configuration actually used, with no credential payload.
	sourcesCase.Record("host-configuration", map[string]any{"policy": policy, "binding": map[string]any{"operation": r.binding.Operation, "provider": r.binding.ProviderID, "api": r.binding.API, "endpoint": r.binding.Endpoint, "models": r.binding.AllowedModels, "allowClassifierImages": true, "retry": r.binding.Retry}, "permissions": p, "permissionHash": candidateReport.PermissionsHash})
	deadline := started.Add(180 * time.Second)
	if outer, ok := t.Deadline(); ok && outer.Before(deadline) {
		deadline = outer
	}
	candidateReport.ProcessDeadline = deadline
	candidateReport.ExecutionDeadline = deadline.Add(-time.Duration(candidateReport.Budget.FinalizationReserveSeconds) * time.Second)
	// Leave time for remaining failure summaries, Observer delivery and the
	// immutable report/manifest/audit/inventory before Go's hard timeout.
	ctx, cancel := context.WithDeadline(context.Background(), candidateReport.ExecutionDeadline)
	defer cancel()
	session := &candidateSession{ctx: ctx, client: client, rec: rec, report: candidateReport}
	for _, sc := range scenarioFactory(t) {
		t.Run(sc.id, func(t *testing.T) { session.execute(t, sc) })
	}
	candidateReport.FinishedAt = time.Now().UTC()
	candidateReport.Outcome = supportmatrix.Pass
	for _, s := range candidateReport.Scenarios {
		if s.Outcome == supportmatrix.Fail {
			candidateReport.Outcome = supportmatrix.Fail
			break
		}
		if s.Outcome != supportmatrix.Pass {
			candidateReport.Outcome = supportmatrix.NotRun
		}
	}
}
func candidateConfigFail(t *testing.T, reason string) {
	for i := range candidateReport.Scenarios {
		candidateReport.Scenarios[i].Outcome = supportmatrix.Fail
		candidateReport.Scenarios[i].ErrorCategory = "config"
		candidateReport.Scenarios[i].Note = reason
	}
	candidateReport.Outcome = supportmatrix.Fail
	t.Errorf("candidate config: %s", reason)
}
func (s *candidateSession) reserve(req ai.ClassifierRequest) bool {
	b := &s.report.Budget
	size := 0
	for _, img := range req.Images {
		raw, err := decodeCandidateImage(img)
		if err != nil {
			return false
		}
		size += len(raw)
	}
	if s.ctx.Err() != nil || b.CallsUsed >= b.MaxCalls || b.AttemptsReserved >= b.MaxAttempts || len(req.Questions) > b.MaxQuestions-b.QuestionsUsed || len(req.State) > b.MaxStateBytes-b.StateBytesSubmitted || len(req.Images) > b.MaxImages-b.ImagesUsed || size > b.MaxImageBytes-b.ImageBytesSubmitted {
		return false
	}
	b.CallsUsed++
	b.AttemptsReserved++
	b.QuestionsUsed += len(req.Questions)
	b.StateBytesSubmitted += len(req.State)
	b.ImagesUsed += len(req.Images)
	b.ImageBytesSubmitted += size
	return true
}
func (s *candidateSession) execute(t *testing.T, sc decisionsScenario) {
	cs := run.Case(t, "P10-candidate-"+sc.id)
	setCandidateReplay(cs, s.report.Mode)
	summary := supportmatrix.ScenarioResult{ID: sc.id, CaseID: cs.ID(), Model: s.report.Model, Outcome: supportmatrix.Pass}
	defer func() {
		if t.Failed() {
			summary.Outcome = supportmatrix.Fail
			if summary.ErrorCategory == "" {
				summary.ErrorCategory = "assertion"
			}
		}
		replaced := false
		for i := range s.report.Scenarios {
			if s.report.Scenarios[i].ID == sc.id {
				s.report.Scenarios[i] = summary
				replaced = true
				break
			}
		}
		if !replaced {
			s.report.Scenarios = append(s.report.Scenarios, summary)
		}
	}()
	if !s.reserve(sc.request) {
		summary.ErrorCategory = "budget"
		cs.Check("cumulative reservation", false, "candidate budget exhausted before network")
		return
	}
	summary.Calls = 1
	summary.Questions = len(sc.request.Questions)
	summary.Images = len(sc.request.Images)
	cs.Record("public-request", sc.request)
	recordCandidateImages(cs, sc.request.Images)
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()
	scope := newScope()
	scope.ActorID = "candidate-maintainer"
	scope.JobID = "decisions-candidate"
	began := time.Now()
	var res ai.ClassifierResult
	var err error
	if sc.cancel {
		res, err = s.client.WithHooks(ai.Hooks{OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error { cancel(); return nil }}).Classify(ctx, scope, ai.Target{BindingID: decisionsCandidate.name, ModelID: "gpt-6-luna"}, sc.request, ai.OpenAIDecisionsOptions{})
	} else {
		res, err = s.client.Classify(ctx, scope, ai.Target{BindingID: decisionsCandidate.name, ModelID: "gpt-6-luna"}, sc.request, ai.OpenAIDecisionsOptions{})
	}
	summary.DurationMS = time.Since(began).Milliseconds()
	ex := s.rec.take()
	s.report.Budget.HTTPAttempts += len(ex)
	summary.Attempts = len(ex)
	for _, a := range res.Metadata.Attempts {
		if a.ProviderRequestID != "" {
			summary.ProviderRequestIDs = append(summary.ProviderRequestIDs, a.ProviderRequestID)
		}
	}
	cs.Record("result", map[string]any{"result": res, "error": errText(err)})
	cs.Record("exchanges", ex)
	obs := s.rec.observationsOf(scope.RequestID)
	cs.Record("observations", obs)
	cs.Record("observer-stats", s.client.ObserverStats())
	cs.Record("budget", s.report.Budget)
	validateCandidateMetadata(cs, s.report, scope, res, ex, obs)
	if sc.cancel {
		var e *ai.Error
		cs.Check("in-flight cancellation is aborted", errors.As(err, &e) && e.Code == ai.CodeCanceled && res.StopReason == ai.StopReasonAborted && len(res.Answers) == 0, "cancel outcome %v", err)
		return
	}
	if err != nil {
		var e *ai.Error
		if errors.As(err, &e) {
			summary.ErrorCategory = string(e.Code)
			summary.Environment = e.Code == ai.CodeRateLimited || e.Code == ai.CodeUpstreamError || e.Code == ai.CodeTransport || e.Code == ai.CodeDeadlineExceeded
		}
		cs.Check("own live call succeeds", false, "%s", errText(err))
		return
	}
	validateCandidateAnswers(cs, sc.request, res)
	reporting := res.Metadata.Attempts[0].UsageReporting
	summary.Note = fmt.Sprintf("actual response model captured; usage reporting %s; billing/cache/regional premium unobserved", reporting)
	// Cost is an estimate from reported input; absent counters stay unreported.
	rate := 0.1
	if res.Usage.Input > 272000 {
		rate = 0.2
	}
	cs.Check("Decisions input-only price snapshot", costClose(res.Usage.Cost.Input, float64(res.Usage.Input)*rate/1e6) && res.Usage.Cost.Output == 0 && res.Usage.Cost.CacheRead == 0 && res.Usage.Cost.CacheWrite == 0, "wrong estimate")
}
