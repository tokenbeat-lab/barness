//go:build live

package live

import (
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func finishXAI(code int, auditOK bool) int {
	p := xaiState
	r := &p.report
	r.FinishedAt = time.Now().UTC()
	r.AuditPassed = auditOK
	r.Complete = code == 0 && auditOK && p.problem == "" && p.notRun == "" && r.ClaimID != "" && r.Budget.CallsUsed == 24 && r.Budget.HTTPAttempts == 24
	for _, row := range r.Scenarios {
		r.Complete = r.Complete && row.Outcome == supportmatrix.Pass
	}
	if source, err := xaiSource(); err != nil || source != r.SourceSHA256 {
		r.Complete = false
		code = 1
		p.problem = "execution source changed"
	}
	if r.ClaimID != "" && !r.Complete {
		code = 1
		if p.problem == "" {
			p.problem = "candidate scope failed or was interrupted"
		}
	}
	if r.Complete {
		r.Outcome = "PASS"
	} else if code != 0 || p.problem != "" {
		r.Outcome = "FAIL"
		if r.Failure == "" {
			r.Failure = p.problem
		}
	} else {
		r.Outcome = "NOT_RUN"
	}
	if failure := finalizeXAICandidate(p); failure != nil {
		return failXAIFinalization(p, failure)
	}
	return code
}
