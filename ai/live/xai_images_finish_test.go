//go:build live

package live

import (
	"os"
	"path/filepath"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func recordXAIImages(code int, audited bool) int {
	p := xaiImagesState
	r := &p.report
	r.FinishedAt = time.Now().UTC()
	r.AuditPassed = audited
	r.Complete = code == 0 && audited && p.problem == "" && p.notRun == "" && r.CallsUsed == 6 && r.HTTPAttempts == 6
	for _, row := range r.Scenarios {
		r.Complete = r.Complete && row.Outcome == "PASS"
	}
	source, err := supportmatrix.XAIImagesExecutionSource()
	if err != nil || supportmatrix.RegionalSourceHash(source) != r.Scope.SourceSHA256 {
		p.problem = "execution source changed"
		r.Complete = false
		code = 1
	}
	if r.Complete {
		r.Outcome = "PASS"
	} else if code != 0 || p.problem != "" {
		r.Outcome = "FAIL"
		r.Failure = p.problem
		if r.Failure == "" {
			r.Failure = "selected candidate scenario failed"
		}
		code = 1
	} else if r.CallsUsed > 0 {
		r.Outcome = "RUNNING"
	} else {
		r.Outcome = "NOT_RUN"
	}
	if err := run.Record("xai-images-candidate-report", r); err != nil {
		return 1
	}
	return code
}

func finishXAIImages(code int, auditOK bool) int {
	p := xaiImagesState
	if p.controlled == "final-write" {
		// Exercise EISDIR at a new pending destination, retaining the original report.
		if err := os.Mkdir(filepath.Join(run.Dir(), "xai-images-final.pending.json"), 0700); err != nil {
			return 1
		}
		if err := run.Record("xai-images-final.pending", p.report); err != nil {
			p.problem = "candidate final report write failed"
			code = 1
		}
	}
	code = recordXAIImages(code, auditOK)
	if p.report.CallsUsed > 0 && !p.report.Complete {
		code = 1
	}
	if err := run.Finish(); err != nil {
		return 1
	}
	findings, err := run.Audit()
	if err != nil || len(findings) > 0 {
		p.problem = "candidate final audit failed"
		_ = recordXAIImages(1, false)
		_ = run.Finish()
		return 1
	}
	return code
}
