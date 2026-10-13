//go:build live

package live

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func setDecisionsReplay(cs *evidence.Case, mode string) {
	if mode == "controlled" {
		if cfg.combo == &decisionsFormal {
			cs.ReplayRun("^TestDecisionsFormalProcess$/^" + regexp.QuoteMeta(os.Getenv("BARNESS_DECISIONS_FORMAL_CONTROLLED")) + "$")
		} else {
			cs.ReplayRun("^TestDecisionsCandidateProcess$/^" + regexp.QuoteMeta(os.Getenv("BARNESS_DECISIONS_CONTROLLED")) + "$")
		}
		return
	}
	assignments := envLive + "=1 " + envCombo + "=" + cfg.combo.name
	if os.Getenv("BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT") == "1" {
		assignments += " BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1"
	}
	cs.ReplayEnv(assignments)
	if cfg.combo == &decisionsFormal {
		cs.ReplayRun("^TestFormalDecisions$")
	} else {
		cs.ReplayRun("^TestCandidateDecisions$")
	}
}

// Audit success is independent of a vendor FAIL. Complete requires the whole
// protocol pipeline, clean audit and successfully written integrity inventory.
func finishDecisionsReport(code int, auditOK bool) int {
	r := decisionsReport
	// Keep the execution timestamp sealed before run.Finish builds its manifest.
	r.AuditPassed = auditOK
	complete := r.Outcome == supportmatrix.Pass && auditOK && code == 0
	// Keep a non-accepting report throughout the fallible inventory read phase.
	r.Complete = false
	if !auditOK {
		r.Outcome = supportmatrix.Fail
		r.FailureCategory = "audit"
		code = 1
	}
	if err := recordDecisionsStatus(); err != nil {
		return decisionsEvidenceFailure("report-write")
	}
	files := map[string]string{}
	err := filepath.WalkDir(run.Dir(), func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		if e.Name() == decisionsIntegrityName()+".json" {
			return nil
		}
		raw, err := readDecisionsFile(path, 16<<20)
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
		return decisionsEvidenceFailure("integrity-read")
	}
	r.Complete = complete
	if err := recordDecisionsStatus(); err != nil {
		return decisionsEvidenceFailure("final-report-write")
	}
	findings, err := run.Audit()
	if err != nil || len(findings) > 0 {
		r.AuditPassed = false
		return decisionsEvidenceFailure("final-audit")
	}
	// Refresh the two files written since inventory collection. The inventory
	// hashes the final status and audit, never a transient provisional report.
	finalFiles := []string{decisionsDetailName() + ".json", "audit.json"}
	if cfg.combo == &decisionsFormal {
		finalFiles = append(finalFiles, supportmatrix.ReportFile)
	}
	for _, name := range finalFiles {
		raw, err := readDecisionsFile(filepath.Join(run.Dir(), name), 1<<20)
		if err != nil {
			return decisionsEvidenceFailure("final-status-read")
		}
		files[name] = sha256Text(raw)
	}
	if err := run.Record(decisionsIntegrityName(), map[string]any{"schema": 1, "files": files}); err != nil {
		return decisionsEvidenceFailure("integrity-write")
	}
	return code
}
func decisionsEvidenceFailure(note string) int {
	decisionsReport.Complete = false
	decisionsReport.Outcome = supportmatrix.Fail
	decisionsReport.FailureCategory = "evidence"
	decisionsReport.FailureNote = note
	if err := recordDecisionsStatus(); err != nil {
		fmt.Fprintln(os.Stderr, "candidate failure status could not be written")
	}
	fmt.Fprintln(os.Stderr, "candidate evidence finalization failed:", note)
	return 1
}
