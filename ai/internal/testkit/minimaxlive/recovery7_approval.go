package minimaxlive

import (
	"fmt"
	"path/filepath"
	"reflect"
)

type recovery7Scope struct {
	recovery6Scope
	SixthCohortSealSHA256    string `json:"sixthCohortSealSHA256"`
	SixthCohortHistorySHA256 string `json:"sixthCohortHistorySHA256"`
}

func (s recovery7Scope) matches7(g RenewalGrant) bool {
	return g.Recovery != nil && g.ThirdCohort != nil && g.FourthCohort != nil && g.FifthCohort != nil && g.SixthCohort != nil && s.Decision == "APPROVED" && s.RenewalID == Recovery7RenewalID && s.ValidationRound == Recovery7RenewalID && s.NativeRenewalID == Recovery7RenewalID && s.SpendingGuardPolicyID == FinalAcceptanceRenewalID && s.FirstCohortSealSHA256 == g.FirstCohortSealSHA256 && s.SecondCohortSealSHA256 == g.Recovery.SecondCohortSealSHA256 && s.HistorySHA256 == g.ClosedHistorySHA256 && s.RecoveryHistorySHA256 == recoveryHistoryPin(g.Recovery) && s.PlanSHA256 == g.PlanSHA256 && s.ThirdCohortSealSHA256 == g.ThirdCohort.CohortSealSHA256 && s.ThirdCohortHistorySHA256 == thirdHistoryPin(g.ThirdCohort) && s.FourthCohortSealSHA256 == g.FourthCohort.CohortSealSHA256 && s.FourthCohortHistorySHA256 == fourthHistoryPin(g.FourthCohort) && reflect.DeepEqual(s.SpendingGuardSourceSHA256, g.SpendingGuardSourceSHA256) && s.SixthCohortSealSHA256 == g.SixthCohort.CohortSealSHA256 && s.SixthCohortHistorySHA256 == sixthHistoryPin(g.SixthCohort) && s.FifthCohortSealSHA256 == g.FifthCohort.CohortSealSHA256 && s.FifthCohortHistorySHA256 == fifthHistoryPin(g.FifthCohort) && s.RequestReference != "" && s.RequestReference == g.RequestReference
}
func recovery7Approval(base string, g RenewalGrant) error {
	var human struct {
		recovery7Scope
		MaxCalls                  int      `json:"maxCalls"`
		MaxAttempts               int      `json:"maxAttempts"`
		CampaignMaxCalls          int      `json:"campaignMaxCalls"`
		CampaignMaxAttempts       int      `json:"campaignMaxAttempts"`
		NewReservedMicrousd       int64    `json:"newReservedMicrousd"`
		RetainedReservedMicrousd  int64    `json:"retainedReservedMicrousd"`
		CampaignMaximumMicrousd   int64    `json:"campaignMaximumMicrousd"`
		RequestText               string   `json:"requestText"`
		ActualHumanApprovalSHA256 string   `json:"actualHumanApprovalSHA256"`
		Provider                  string   `json:"provider"`
		Model                     string   `json:"model"`
		Purpose                   string   `json:"purpose"`
		Routes                    []string `json:"routes"`
		RenewalMaxCalls           int      `json:"renewalMaxCalls"`
		RenewalMaxAttempts        int      `json:"renewalMaxAttempts"`
	}
	raw, e := ReadFile(filepath.Join(base, "human-approval.json"), 1<<20)
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || decodeRecoveryApproval(raw, &human) != nil || !human.matches7(g) || human.MaxCalls != 378 || human.MaxAttempts != 414 || human.CampaignMaxCalls != 1582 || human.CampaignMaxAttempts != 1694 || human.NewReservedMicrousd != 6900000 || human.RetainedReservedMicrousd != 32900000 || human.CampaignMaximumMicrousd != 40000000 || human.RequestText == "" || !regexpHex64(human.ActualHumanApprovalSHA256) || human.Provider != "minimax" || human.Model != M3 || human.Purpose != "standard" || !reflect.DeepEqual(human.Routes, []string{"chat", "responses", "messages"}) || human.RenewalMaxCalls != 60 || human.RenewalMaxAttempts != 72 {
		return fmt.Errorf("own recovery7 378/414,1582/1694,40USD actual human approval required")
	}
	var review struct {
		recovery7Scope
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || decodeRecoveryApproval(raw, &review) != nil || !review.matches7(g) || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 {
		return fmt.Errorf("exact recovery7 source/history/guard review required")
	}
	owner, e := evidenceOwner()
	if e != nil {
		return e
	}
	var actual struct {
		actualRecoveryApproval
		SpendingGuardSourceSHA256 map[string]string `json:"spendingGuardSourceSHA256"`
	}
	raw, e = ReadFile(filepath.Join(owner, ".evidence/provider-sites/final-acceptance-20261020/recovery7-authorization.json"), 1<<20)
	if e != nil || Hash(raw) != human.ActualHumanApprovalSHA256 || decodeRecoveryApproval(raw, &actual) != nil {
		return fmt.Errorf("canonical actual recovery7 human record missing")
	}
	a := actual.actualRecoveryApproval
	if a.Decision != "APPROVED" || a.MaxUSD != 40 || a.Round != Recovery7RenewalID || a.ValidationRound != Recovery7RenewalID || a.NativeRenewalID != Recovery7RenewalID || a.SpendingGuardPolicyID != FinalAcceptanceRenewalID || a.SourceCommit != g.SourceCommit || a.SourceManifestSHA256 != g.SourceManifestSHA256 || a.BinariesManifestSHA256 != g.BinariesManifestSHA256 || a.PlanSHA256 != g.PlanSHA256 || a.MaxCalls != 378 || a.MaxAttempts != 414 || a.CampaignMaxCalls != 1582 || a.CampaignMaxAttempts != 1694 || a.NewReservedMicrousd != 6900000 || a.RetainedReservedMicrousd != 32900000 || a.RequestReference != human.RequestReference || a.RequestText != human.RequestText || !reflect.DeepEqual(actual.SpendingGuardSourceSHA256, g.SpendingGuardSourceSHA256) {
		return fmt.Errorf("canonical actual recovery7 scope/source/guard differs")
	}
	var previous actualRecoveryApproval
	raw, e = ReadFile(filepath.Join(owner, ".evidence/provider-sites/final-acceptance-20261020/recovery6-authorization.json"), 1<<20)
	if e != nil || Hash(raw) != g.SixthCohort.ActualHumanApprovalSHA256 || decodeRecoveryApproval(raw, &previous) != nil || previous.RequestReference == "" || a.RequestReference == previous.RequestReference {
		return fmt.Errorf("recovery7 requires a new direct human response reference")
	}
	return recoveryFirstRoute(g, human.ActualHumanApprovalSHA256, Recovery7RenewalID, seventhCollection)
}
