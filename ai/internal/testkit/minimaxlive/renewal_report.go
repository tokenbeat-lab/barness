package minimaxlive

import (
	"fmt"
)

func verifyRenewalReport(r Report) error {
	if r.FormalRenewalID == "" {
		if r.FormalRenewalGrantSHA256 != "" {
			return fmt.Errorf("renewal grant without fixed identity")
		}
		return nil
	}
	if (r.FormalRenewalID != FormalRenewalID && r.FormalRenewalID != FinalAcceptanceRenewalID) || r.Mode != "live" || r.Stage != "formal" || r.Kind != "minimax-formal" || r.Identity.Model != M3 {
		return fmt.Errorf("renewal requires own actual M3 formal report")
	}
	g, pin, e := approvedRenewal(r.FormalRenewalID, false)
	if e != nil {
		return fmt.Errorf("renewal canonical authority: %w", e)
	}
	if r.FormalRenewalGrantSHA256 != pin || r.GitCommit != g.SourceCommit {
		return fmt.Errorf("renewal report source/grant pins differ")
	}
	return nil
}
func verifyRenewalOrigin(r Report, o Origin) error {
	if o.FormalRenewalID != r.FormalRenewalID || o.FormalRenewalGrantSHA256 != r.FormalRenewalGrantSHA256 {
		return fmt.Errorf("renewal origin identity differs")
	}
	if r.FormalRenewalID == "" {
		return nil
	}
	g, pin, e := approvedRenewal(r.FormalRenewalID, false)
	if e != nil {
		return e
	}
	if pin != o.FormalRenewalGrantSHA256 || g.SourceCommit != o.SourceCommit || g.SourceManifestSHA256 != o.SourceManifestSHA256 || g.BinariesManifestSHA256 != o.BinariesManifestSHA256 || g.PlanSHA256 != o.PlanSHA256 {
		return fmt.Errorf("renewal origin approval pins differ")
	}
	return nil
}
