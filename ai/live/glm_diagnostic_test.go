//go:build live

package live

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

const glmDiagnosticEntry = "^TestGLMMessagesCancelDiagnostic$"
const glmDiagnosticKind = "glm-single-case-diagnostic-only"
const glmDiagnosticCase = "glm-5.3/cancel"

func glmDiagnosticSelected() bool { return os.Getenv("BARNESS_GLM_DIAGNOSTIC") != "" }

// This explicit diagnostic is separate from formal22 selection. Only the fixed
// own cancellation case is reused; no model/route qualification is published.
func selectGLMDiagnostic() error {
	r := glmReportState
	r.Kind = glmDiagnosticKind
	glmReportName = "diagnostic-glm-messages-cancel-report"
	r.Budget = glmBudget{PlannedCalls: 1, MaxCalls: 1, MaxAttempts: 1, MaxOutputTokens: 4096, CallSeconds: 120, ProcessSeconds: 180, FinalizationReserveSeconds: 20}
	if os.Getenv("BARNESS_GLM_DIAGNOSTIC") != glmDiagnosticCase || r.Route != "messages" {
		return fmt.Errorf("diagnostic requires only Messages glm-5.3/cancel")
	}
	for _, name := range []string{"BARNESS_GLM_FORMAL", "BARNESS_GLM_REPLAY_ARCHIVE", "BARNESS_MINIMAX_ROUTE", "BARNESS_MM_ROUTE"} {
		if os.Getenv(name) != "" {
			return fmt.Errorf("diagnostic forbids mixed formal/replay/provider mode")
		}
	}
	fixed := glmFixedPlan.Routes["messages"]
	var own []glmCasePlan
	for _, sc := range fixed {
		if sc.ID == glmDiagnosticCase && sc.Model == "glm-5.3" && sc.Kind == "cancel" {
			own = append(own, sc)
		}
	}
	if len(own) != 1 {
		return fmt.Errorf("fixed cancellation plan missing or ambiguous")
	}
	glmCandidateCatalog = ai.BuiltinCatalog()
	if _, ok := glmCandidateCatalog.Lookup(ai.ProviderZhipu, ai.APIAnthropicMessages, "glm-5.3"); !ok {
		return fmt.Errorf("builtin diagnostic identity unavailable")
	}
	glmSelected.name = "zhipu-messages-diagnostic"
	glmFixedPlan.Routes["messages"] = own
	return nil
}

func TestGLMMessagesCancelDiagnostic(t *testing.T) {
	if glmReportState == nil || glmReportState.Kind != glmDiagnosticKind {
		t.Skip("one-case diagnostic not selected")
	}
	r := glmReportState
	if glmConfigError != "" {
		glmFail(t, glmConfigError)
		return
	}
	if os.Getenv("BARNESS_AI_LIVE") != "1" || glmKey == "" {
		t.Skip("NOT_RUN: explicit live switch or selected key absent")
	}
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	if name := os.Getenv("BARNESS_GLM_CONTROLLED"); name != "" {
		r.Mode = "controlled"
		run.SetVersion("glm_mode", r.Mode)
		installGLMControlled(rec, name)
	}
	rec.next = &glmHTTPGate{next: rec.next}
	client, err := newGLMCandidateClient(glmSelected, glmKey, rec, glmCandidateCatalog)
	if err != nil {
		glmFail(t, "public diagnostic Client assembly failed")
		return
	}
	r.ProcessDeadline = started.Add(180 * time.Second)
	if d, ok := t.Deadline(); ok && d.Before(r.ProcessDeadline) {
		r.ProcessDeadline = d
	}
	r.ExecutionDeadline = r.ProcessDeadline.Add(-20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), r.ExecutionDeadline)
	defer cancel()
	session := &glmSession{ctx: ctx, client: client, rec: rec, results: map[string]ai.Result{}, requests: map[string]ai.Request{}, bodies: map[string]map[string]any{}}
	// Reuse the actual fixed prompt, options,120s call and observable-content
	// cancellation assertions, without copying or weakening any formal case.
	session.execute(t, glmFixedPlan.Routes["messages"][0], &r.Scenarios[0])
	r.Outcome = r.Scenarios[0].Outcome
	r.Models["glm-5.3"] = r.Outcome
}

func setGLMDiagnosticReplay(cs *evidence.Case) {
	if os.Getenv("BARNESS_GLM_CONTROLLED") != "" {
		cs.ReplayEnv("BARNESS_AI_LIVE=0 BARNESS_GLM_ROUTE= BARNESS_GLM_DIAGNOSTIC= BARNESS_GLM_CONTROLLED= BARNESS_GLM_KEY=")
		cs.ReplayRun("^TestGLMMessagesCancelDiagnosticProcess$/^success$")
		return
	}
	cs.ReplayEnv("BARNESS_AI_LIVE=1 BARNESS_GLM_ROUTE=messages BARNESS_GLM_DIAGNOSTIC=glm-5.3/cancel")
	cs.ReplayRun(glmDiagnosticEntry)
}
