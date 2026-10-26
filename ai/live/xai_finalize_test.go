//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type xaiFinalizationError struct {
	step string
	err  error
}

func (e *xaiFinalizationError) Error() string { return "xAI finalization failed at " + e.step }

func finalizeXAI(p *xaiProcess) *xaiFinalizationError {
	fail := func(step string, err error) *xaiFinalizationError { return &xaiFinalizationError{step, err} }
	final := filepath.Join(run.Dir(), p.reportName()+".json")
	// Preserve the previously published non-PASS report and its actual budget.
	raw, err := os.ReadFile(final)
	if err != nil {
		return fail("preserve-report", err)
	}
	if err := run.Record(p.reportName()+".unfinalized", json.RawMessage(raw)); err != nil {
		return fail("preserve-report", err)
	}
	pending := filepath.Join(run.Dir(), p.reportName()+".pending.json")
	if p.controlled == "final-write" {
		if err := os.Mkdir(pending, 0700); err != nil {
			return fail("pending-write", err)
		}
	}
	if err := run.Record(p.reportName()+".pending", &p.report); err != nil {
		return fail("pending-write", err)
	}
	if err := syncXAIFile(pending); err != nil {
		return fail("pending-sync", err)
	}
	projected := p.report.Stage == "formal" && p.report.Complete
	if projected {
		report, err := p.formalProjection()
		if err != nil {
			return fail("standard-projection", err)
		}
		if p.controlled == "projection-write" {
			if err := os.Mkdir(filepath.Join(run.Dir(), "live-report.pending.json"), 0700); err != nil {
				return fail("projection-pending-write", err)
			}
		}
		if err := run.Record("live-report.pending", report); err != nil {
			return fail("projection-pending-write", err)
		}
		if err := syncXAIFile(filepath.Join(run.Dir(), "live-report.pending.json")); err != nil {
			return fail("projection-pending-sync", err)
		}
	}
	// Audit the complete staged bytes before publishing success claims.
	if err := auditXAI(); err != nil {
		return fail("staged-audit", err)
	}
	if err := os.Rename(pending, final); err != nil {
		return fail("report-promotion", err)
	}
	if projected {
		if err := os.Rename(filepath.Join(run.Dir(), "live-report.pending.json"), filepath.Join(run.Dir(), "live-report.json")); err != nil {
			return fail("projection-promotion", err)
		}
	}
	dir, err := os.Open(run.Dir())
	if err != nil {
		return fail("directory-open", err)
	}
	if p.controlled == "final-sync" {
		_ = dir.Close() // Exercise the real filesystem operation on a closed fd.
	}
	syncErr, closeErr := dir.Sync(), dir.Close()
	if err := errors.Join(syncErr, closeErr); err != nil {
		return fail("directory-sync", err)
	}
	if p.controlled == "final-audit" {
		// An actual EISDIR on the final audit write, after report promotion.
		if err := os.Remove(filepath.Join(run.Dir(), "audit.json")); err != nil {
			return fail("audit-fault-setup", err)
		}
		if err := os.Mkdir(filepath.Join(run.Dir(), "audit.json"), 0700); err != nil {
			return fail("audit-fault-setup", err)
		}
	}
	if err := auditXAI(); err != nil {
		return fail("final-audit", err)
	}
	return nil
}

func auditXAI() error {
	findings, err := run.Audit()
	if err != nil {
		return err
	}
	if len(findings) > 0 {
		return fmt.Errorf("xAI redaction audit failed")
	}
	return nil
}

func syncXAIFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}

// Late I/O/audit failure invalidates both the canonical report and any staged
// success. If a non-PASS rewrite fails, unpublish the file rather than retaining
// a readable success claim. A broken filesystem can still prevent cleanup;
// the process remains failed and the original report/budget is retained above.
func failXAIFinalization(p *xaiProcess, failure *xaiFinalizationError) int {
	r := &p.report
	r.Outcome, r.Failure = "FAIL", failure.Error()
	r.Complete, r.AuditPassed, r.FormalQualified = false, false, false
	projectionInvalidated := p.unpublishProjection()
	for _, name := range []string{p.reportName(), p.reportName() + ".pending"} {
		path := filepath.Join(run.Dir(), name+".json")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			continue // Retain the controlled EISDIR target as evidence.
		}
		if err := run.Record(name, r); err != nil {
			_ = os.Remove(path)
			continue
		}
		_ = syncXAIFile(path)
	}
	// Keep the failed audit target, then reseal the bundle's non-PASS state.
	auditPath := filepath.Join(run.Dir(), "audit.json")
	if info, err := os.Stat(auditPath); err == nil && info.IsDir() {
		_ = os.Rename(auditPath, filepath.Join(run.Dir(), "xai-final-audit-fault"))
	}
	detail := map[string]any{"step": failure.step, "projectionInvalidated": projectionInvalidated, "closedFile": errors.Is(failure.err, os.ErrClosed), "isDirectory": errors.Is(failure.err, syscall.EISDIR)}
	var pathErr *os.PathError
	if errors.As(failure.err, &pathErr) {
		detail["operation"] = pathErr.Op
	}
	_ = run.Record("xai-finalization-failure", detail)
	_ = auditXAI()
	return 1
}

// Invalidate the standard projection first: while the own report is being
// rewritten, no readable generic report may advertise a completed formal run.
func (p *xaiProcess) unpublishProjection() bool {
	if p.report.Stage != "formal" {
		return true
	}
	ok := true
	for _, name := range []string{"live-report.json", "live-report.pending.json"} {
		path := filepath.Join(run.Dir(), name)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			ok = false
			continue
		}
		if info.IsDir() {
			continue
		} // Keep the controlled EISDIR fault target.
		if err = os.Remove(path); err != nil {
			ok = false
		}
	}
	dir, err := os.Open(run.Dir())
	if err != nil {
		return false
	}
	return errors.Join(dir.Sync(), dir.Close()) == nil && ok
}
