//go:build live

package live

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
)

func finishMiniMax(code int, auditOK bool) int {
	r := minimaxReport
	r.AuditPassed = auditOK
	r.Complete = false
	if code != 0 || !auditOK {
		r.Outcome = "FAIL"
	}
	if e := captureMiniMaxLedger(); e != nil {
		r.Outcome = "FAIL"
		r.Failure = "durable ledger evidence unavailable"
		code = 1
	}
	if r.Outcome == "PASS" && r.Kind != "minimax-research-only" && r.Kind != minimaxlive.ThinkingKind && code == 0 && auditOK {
		r.Complete = true
	}
	if e := run.Record("minimax-report", r); e != nil {
		return 1
	}
	if findings, e := run.Audit(); e != nil || len(findings) > 0 {
		r.AuditPassed = false
		r.Complete = false
		r.Outcome = "FAIL"
		_ = run.Record("minimax-report", r)
		return 1
	}
	files, e := minimaxlive.Inventory(run.Dir())
	if e != nil || os.Getenv("BARNESS_MINIMAX_CONTROLLED") == "inventory-failure" {
		r.Complete = false
		r.Outcome = "FAIL"
		r.Failure = "inventory read failed"
		_ = run.Record("minimax-report", r)
		return 1
	}
	if e = run.Record("minimax-integrity", minimaxlive.Seal{Schema: 1, Files: files}); e != nil {
		return 1
	}
	if r.Kind == minimaxlive.ThinkingKind && r.Outcome == "PASS" {
		if _, e = minimaxlive.VerifyThinkingSupplement(run.Dir(), false); e != nil {
			return rejectMiniMaxVerification(e)
		}
	}
	if r.Complete {
		if _, e = minimaxlive.VerifyCapture(run.Dir(), false); e != nil {
			return rejectMiniMaxVerification(e)
		}
	}
	// Actual formal support projection is published by the keyless post-exit
	// consumer after the parent launcher has pinned this report and seal.
	return code
}

// The native verifier needs the provisional seal above. A rejected attempt is
// still being finalized: persist its final FAIL and audit before its last seal.
// After finishMiniMax returns, consumers are read-only and never repair history.
func rejectMiniMaxVerification(cause error) int {
	r := minimaxReport
	r.Outcome, r.Complete, r.Failure = "FAIL", false, cause.Error()
	if e := run.Record("minimax-report", r); e != nil {
		return 1
	}
	if findings, e := run.Audit(); e != nil || len(findings) > 0 {
		r.AuditPassed = false
		if e := run.Record("minimax-report", r); e != nil {
			return 1
		}
	}
	files, e := minimaxlive.Inventory(run.Dir())
	if e != nil {
		return 1
	}
	if e := run.Record("minimax-integrity", minimaxlive.Seal{Schema: 1, Files: files}); e != nil {
		return 1
	}
	return 1
}
