//go:build live

package live

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func setCandidateReplay(cs *evidence.Case, mode string) {
	if mode == "controlled" {
		cs.ReplayRun("^TestDecisionsCandidateProcess$/^" + regexp.QuoteMeta(os.Getenv("BARNESS_DECISIONS_CONTROLLED")) + "$")
		return
	}
	cs.ReplayEnv(envLive + "=1 " + envCombo + "=" + decisionsCandidate.name)
	cs.ReplayRun("^TestCandidateDecisions$")
}

// Audit success is independent of a vendor FAIL. Complete requires the whole
// protocol pipeline, clean audit and successfully written integrity inventory.
func finishCandidateReport(code int, auditOK bool) int {
	r := candidateReport
	r.FinishedAt = time.Now().UTC()
	r.AuditPassed = auditOK
	complete := r.Outcome == supportmatrix.Pass && auditOK && code == 0
	// Keep a non-accepting report throughout the fallible inventory read phase.
	r.Complete = false
	if !auditOK {
		r.Outcome = supportmatrix.Fail
		r.FailureCategory = "audit"
		code = 1
	}
	if err := run.Record(candidateReportName, r); err != nil {
		return candidateEvidenceFailure("report-write")
	}
	files := map[string]string{}
	err := filepath.WalkDir(run.Dir(), func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		if e.Name() == "candidate-integrity.json" {
			return nil
		}
		raw, err := readCandidateFile(path, 16<<20)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(run.Dir(), path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = sha256Text(raw)
		return nil
	})
	if err != nil {
		return candidateEvidenceFailure("integrity-read")
	}
	r.Complete = complete
	if err := run.Record(candidateReportName, r); err != nil {
		return candidateEvidenceFailure("final-report-write")
	}
	findings, err := run.Audit()
	if err != nil || len(findings) > 0 {
		r.AuditPassed = false
		return candidateEvidenceFailure("final-audit")
	}
	// Refresh the two files written since inventory collection. The inventory
	// hashes the final status and audit, never a transient provisional report.
	for _, name := range []string{candidateReportName + ".json", "audit.json"} {
		raw, err := readCandidateFile(filepath.Join(run.Dir(), name), 1<<20)
		if err != nil {
			return candidateEvidenceFailure("final-status-read")
		}
		files[name] = sha256Text(raw)
	}
	if err := run.Record("candidate-integrity", map[string]any{"schema": 1, "files": files}); err != nil {
		return candidateEvidenceFailure("integrity-write")
	}
	return code
}
func candidateEvidenceFailure(note string) int {
	candidateReport.Complete = false
	candidateReport.Outcome = supportmatrix.Fail
	candidateReport.FailureCategory = "evidence"
	candidateReport.FailureNote = note
	if err := run.Record(candidateReportName, candidateReport); err != nil {
		fmt.Fprintln(os.Stderr, "candidate failure status could not be written")
	}
	fmt.Fprintln(os.Stderr, "candidate evidence finalization failed:", note)
	return 1
}
