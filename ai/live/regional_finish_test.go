//go:build live

package live

import (
	"fmt"
	"os"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func finishRegional(code int, auditOK bool) int {
	p := regionalState
	if p.report.Regional == nil {
		return code
	}
	e := p.report.Regional
	e.AuditPassed = auditOK
	e.Complete = code == 0 && auditOK && p.problem == "" && p.notRun == ""
	for _, s := range p.report.Scenarios {
		e.Complete = e.Complete && s.Outcome == supportmatrix.Pass
	}
	current, err := supportmatrix.RegionalExecutionSource()
	if err != nil || !reflect.DeepEqual(current, p.source) {
		e.Complete = false
		code = 1
	}
	selectedComplete := e.Complete
	if p.grant.Purpose == supportmatrix.KimiHistoryPurpose {
		e.Complete = false
	}
	if err := run.Record("regional-report", p.report); err != nil {
		return 1
	}
	if findings, err := run.Audit(); err != nil || len(findings) > 0 {
		return 1
	}
	if err := supportmatrix.SealRegionalBundle(run.Dir()); err != nil {
		return 1
	}
	if p.grant.Purpose == supportmatrix.KimiHistoryPurpose {
		if !selectedComplete {
			return 1
		}
		if _, err := supportmatrix.VerifyKimiHistoryCapture(run.Dir(), e.Mode); err != nil {
			fmt.Fprintln(os.Stderr, "Kimi selected capture:", err)
			return 1
		}
		return code
	}
	if regionalBaselineFixture && p.controlled != "" && p.grant.MaxCalls == 9 {
		if !e.Complete {
			return 1
		}
		if _, err := supportmatrix.VerifyKimiBaselineCapture(run.Dir(), "controlled"); err != nil {
			return 1
		}
		return code
	}
	if e.Complete && e.Mode == "live" {
		own, err := supportmatrix.VerifyRegionalQualification(run.Dir())
		if err != nil {
			return 1
		}
		if e.Stage == "formal" {
			if err := run.Record(supportmatrix.ReportName, own); err != nil {
				return 1
			}
		}
	}
	return code
}
