//go:build live

package live

import (
	"fmt"
	"os"
	"path/filepath"
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
	// The final report replaces the previous non-PASS state only after a
	// complete pending write. The controlled fault creates an actual EISDIR
	// failure on that same write path; it cannot manufacture a passing file.
	pending := filepath.Join(run.Dir(), "xai-candidate-report.pending.json")
	if p.controlled == "final-write" {
		if err := os.Mkdir(pending, 0700); err != nil {
			return 1
		}
	}
	if err := run.Record("xai-candidate-report.pending", r); err != nil {
		fmt.Println("xai candidate final report write failed")
		return 1
	}
	file, err := os.OpenFile(pending, os.O_RDWR, 0600)
	if err != nil {
		return 1
	}
	syncErr := file.Sync()
	closeErr := file.Close()
	if syncErr != nil || closeErr != nil {
		return 1
	}
	if err := os.Rename(pending, filepath.Join(run.Dir(), "xai-candidate-report.json")); err != nil {
		return 1
	}
	dir, err := os.Open(run.Dir())
	if err != nil {
		return 1
	}
	syncErr = dir.Sync()
	closeErr = dir.Close()
	if syncErr != nil || closeErr != nil {
		return 1
	}

	if findings, err := run.Audit(); err != nil || len(findings) > 0 {
		return 1
	}
	return code
}
