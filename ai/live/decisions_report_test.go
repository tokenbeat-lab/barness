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

// Audit success is independent of a vendor FAIL. Complete requires both the
// complete protocol pipeline and the audit, never missing token counters.
func finishCandidateReport(code int, auditOK bool) int {
	r := candidateReport
	r.FinishedAt = time.Now().UTC()
	r.AuditPassed = auditOK
	r.Complete = r.Outcome == supportmatrix.Pass && auditOK && code == 0
	if !auditOK {
		r.Outcome = supportmatrix.Fail
		code = 1
	}
	if err := run.Record(candidateReportName, r); err != nil {
		fmt.Fprintln(os.Stderr, "candidate report write failed")
		return 1
	}
	findings, err := run.Audit()
	if err != nil || len(findings) > 0 {
		r.AuditPassed = false
		r.Complete = false
		r.Outcome = supportmatrix.Fail
		code = 1
		if err := run.Record(candidateReportName, r); err != nil {
			fmt.Fprintln(os.Stderr, "candidate final status write failed")
		}
	}
	// The ordinary manifest inventories case artifacts. This adds root report,
	// source metadata, manifest and audit to a complete checksum inventory.
	files := map[string]string{}
	err = filepath.WalkDir(run.Dir(), func(path string, e os.DirEntry, err error) error {
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
		fmt.Fprintln(os.Stderr, "candidate inventory failed")
		return 1
	}
	if err := run.Record("candidate-integrity", map[string]any{"schema": 1, "files": files}); err != nil {
		fmt.Fprintln(os.Stderr, "candidate inventory write failed")
		return 1
	}
	return code
}
