//go:build live

package live

import (
	"fmt"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func selectGLMFormalCatalog() error {
	formal := ai.BuiltinCatalog()
	if glmReportState.Mode == "replay" {
		var err error
		formal, err = glmOriginalReplayCatalog()
		if err != nil {
			return err
		}
	}
	count := 0
	for _, m := range formal.Models {
		if m.Provider == ai.ProviderZhipu {
			count++
		}
	}
	if count != 6 {
		return fmt.Errorf("formal catalog lacks independent six identities")
	}
	for _, want := range glmCandidateCatalog.Models {
		got, ok := formal.Lookup(want.Provider, want.API, want.ID)
		want.Name = got.Name
		if !ok || !reflect.DeepEqual(got, want) {
			return fmt.Errorf("formal catalog differs from accepted source facts")
		}
	}
	glmCandidateCatalog = formal
	glmSelected.name = "zhipu-" + glmReportState.Route
	if glmReportState.Mode != "replay" {
		glmReportState.Kind = "glm-formal"
	}
	if glmReportState.Mode != "replay" {
		glmReportName = "formal-glm-report"
	}
	source, err := supportmatrix.GLMExecutionSource()
	if err != nil {
		return err
	}
	if err = run.Record("glm-execution-source", source); err != nil {
		return err
	}
	return nil
}

// Only a sealed own live detail can produce the public support report. This is
// after finishGLM's complete inventory/audit, never a controlled pipeline PASS.
func publishGLMFormal(code int) int {
	if glmReportState.Kind != "glm-formal" || glmReportState.Mode != "live" || glmReportState.Outcome == "NOT_RUN" {
		return code
	}
	r, err := supportmatrix.NewGLMReport(run.Dir())
	if err != nil {
		return 1
	}
	if err = run.Record(supportmatrix.ReportName, r); err != nil {
		return 1
	}
	return code
}
