//go:build live

package live

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Candidate and formal runs share the already accepted protocol pipeline;
// catalog authority and report identities remain independent at both boundaries.
var decisionsFormal = combo{name: "openai-decisions", spec: "P10", operation: ai.OperationClassifier, provider: ai.ProviderOpenAI, api: ai.APIOpenAIDecisions, endpoint: "https://api.openai.com/v1", model: "gpt-6-luna", sdk: func() string { return "direct HTTP (OpenAI Decisions; ADR-0026)" }}

func isDecisionsProcess() bool {
	return cfg.combo == &decisionsCandidate || cfg.combo == &decisionsFormal
}
func decisionsBindingVersion() string {
	if cfg.combo == &decisionsFormal {
		return "live"
	}
	return "candidate"
}
func decisionsCasePrefix() string {
	if cfg.combo == &decisionsFormal {
		return "LIVE-P10-"
	}
	return "P10-candidate-"
}
func decisionsDetailName() string {
	if cfg.combo == &decisionsFormal {
		return "decisions-execution-report"
	}
	return candidateReportName
}
func decisionsIntegrityName() string {
	if cfg.combo == &decisionsFormal {
		return "decisions-integrity"
	}
	return "candidate-integrity"
}
func TestFormalDecisions(t *testing.T) {
	if cfg.combo != &decisionsFormal {
		t.Skip("formal Decisions not selected")
	}
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	driveDecisions(t, rec, decisionsScenarios, "live")
}
func formalDecisionsReport() supportmatrix.Report {
	r := decisionsReport
	b := r.Budget
	return supportmatrix.Report{
		Schema: supportmatrix.SchemaVersion, Combo: decisionsFormal.name, Operation: string(r.Operation), Provider: string(r.Provider), API: string(r.API), Endpoint: r.Endpoint, Model: r.Model, SDK: r.SDK,
		AccountAlias: r.AccountScopeID, GitCommit: r.GitCommit, CatalogVersion: r.CatalogVersion, CatalogHash: r.CatalogHash, StartedAt: r.StartedAt, FinishedAt: r.FinishedAt, Expected: r.Expected, Scenarios: r.Scenarios,
		Budget:    supportmatrix.Budget{MaxCalls: b.MaxCalls, MaxAttempts: b.MaxAttempts, AttemptsReserved: b.AttemptsReserved, MaxQuestions: b.MaxQuestions, MaxStateBytesTotal: b.MaxStateBytes, MaxImages: b.MaxImages, MaxDecodedImageBytesTotal: b.MaxImageBytes, ProcessSeconds: b.ProcessSeconds, CallSeconds: b.CallSeconds, FinalizationReserveSeconds: b.FinalizationReserveSeconds, CallsUsed: b.CallsUsed, HTTPAttempts: b.HTTPAttempts, QuestionsUsed: b.QuestionsUsed, StateBytesSubmitted: b.StateBytesSubmitted, ImagesUsed: b.ImagesUsed, DecodedImageBytesSubmitted: b.ImageBytesSubmitted, EnvironmentRetries: b.Retries, InputTokensUsed: b.ReportedInputTokens, OutputTokensUsed: b.ReportedOutputTokens},
		Decisions: &supportmatrix.DecisionsEvidence{Mode: r.Mode, Complete: r.Complete, AuditPassed: r.AuditPassed, SourceManifestHash: r.SourceManifestHash, AccountVerificationStatus: r.AccountVerificationStatus, RegionVerificationStatus: r.RegionVerificationStatus, PermissionVerificationStatus: r.PermissionVerificationStatus, AccountVerificationPolicy: r.AccountVerificationPolicy},
	}
}

// Both statuses are nonaccepting until finalization has passed every fallible
// inventory read and audit. They share a sealed execution time, not filenames.
func recordDecisionsStatus() error {
	if err := run.Record(decisionsDetailName(), decisionsReport); err != nil {
		return err
	}
	if cfg.combo == &decisionsFormal {
		provisional := formalDecisionsReport()
		provisional.Decisions.Complete = false
		return run.Record(supportmatrix.ReportName, provisional)
	}
	return nil
}

const decisionsPendingReport = "decisions-pending-report"

// Sync both the final bytes and the inventory before their atomic publication.
// An abrupt exit before Rename leaves the stable report nonaccepting; afterwards
// it names exactly the final bytes already sealed by the inventory and audit.
func publishFormalDecisionsReport() error {
	for _, name := range []string{decisionsPendingReport + ".json", decisionsIntegrityName() + ".json"} {
		f, err := os.Open(filepath.Join(run.Dir(), name))
		if err != nil {
			return err
		}
		err = f.Sync()
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return os.Rename(filepath.Join(run.Dir(), decisionsPendingReport+".json"), filepath.Join(run.Dir(), supportmatrix.ReportFile))
}
