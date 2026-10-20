package minimaxlive

import (
	"fmt"
	"path/filepath"
	"reflect"
)

type recovery5Scope struct {
	recovery4Scope
	FourthCohortSealSHA256    string            `json:"fourthCohortSealSHA256"`
	FourthCohortHistorySHA256 string            `json:"fourthCohortHistorySHA256"`
	SpendingGuardSourceSHA256 map[string]string `json:"spendingGuardSourceSHA256"`
	RequestReference          string            `json:"requestReference"`
}

func (s recovery5Scope) matches5(g RenewalGrant) bool {
	return g.Recovery != nil && g.ThirdCohort != nil && g.FourthCohort != nil && s.Decision == "APPROVED" && s.RenewalID == Recovery5RenewalID && s.ValidationRound == Recovery5RenewalID && s.NativeRenewalID == Recovery5RenewalID && s.SpendingGuardPolicyID == FinalAcceptanceRenewalID && s.FirstCohortSealSHA256 == g.FirstCohortSealSHA256 && s.SecondCohortSealSHA256 == g.Recovery.SecondCohortSealSHA256 && s.HistorySHA256 == g.ClosedHistorySHA256 && s.RecoveryHistorySHA256 == recoveryHistoryPin(g.Recovery) && s.PlanSHA256 == g.PlanSHA256 && s.ThirdCohortSealSHA256 == g.ThirdCohort.CohortSealSHA256 && s.ThirdCohortHistorySHA256 == thirdHistoryPin(g.ThirdCohort) && s.FourthCohortSealSHA256 == g.FourthCohort.CohortSealSHA256 && s.FourthCohortHistorySHA256 == fourthHistoryPin(g.FourthCohort) && reflect.DeepEqual(s.SpendingGuardSourceSHA256, g.SpendingGuardSourceSHA256) && s.RequestReference != "" && s.RequestReference == g.RequestReference
}
func recovery5Approval(base string, g RenewalGrant) error {
	var human struct {
		recovery5Scope
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
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || decodeRecoveryApproval(raw, &human) != nil || !human.matches5(g) || human.MaxCalls != 378 || human.MaxAttempts != 414 || human.CampaignMaxCalls != 1358 || human.CampaignMaxAttempts != 1454 || human.NewReservedMicrousd != 6900000 || human.RetainedReservedMicrousd != 28800000 || human.CampaignMaximumMicrousd != 36000000 || human.RequestText == "" || !regexpHex64(human.ActualHumanApprovalSHA256) || human.Provider != "minimax" || human.Model != M3 || human.Purpose != "standard" || !reflect.DeepEqual(human.Routes, []string{"chat", "responses", "messages"}) || human.RenewalMaxCalls != 60 || human.RenewalMaxAttempts != 72 {
		return fmt.Errorf("own recovery5 378/414,1358/1454,36USD actual human approval required")
	}
	var review struct {
		recovery5Scope
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || decodeRecoveryApproval(raw, &review) != nil || !review.matches5(g) || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 {
		return fmt.Errorf("exact recovery5 source/history/guard review required")
	}
	owner, e := evidenceOwner()
	if e != nil {
		return e
	}
	var actual struct {
		actualRecoveryApproval
		SpendingGuardSourceSHA256 map[string]string `json:"spendingGuardSourceSHA256"`
	}
	raw, e = ReadFile(filepath.Join(owner, ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json"), 1<<20)
	if e != nil || Hash(raw) != human.ActualHumanApprovalSHA256 || decodeRecoveryApproval(raw, &actual) != nil {
		return fmt.Errorf("canonical actual recovery5 human record missing")
	}
	a := actual.actualRecoveryApproval
	if a.Decision != "APPROVED" || a.MaxUSD != 36 || a.Round != Recovery5RenewalID || a.ValidationRound != Recovery5RenewalID || a.NativeRenewalID != Recovery5RenewalID || a.SpendingGuardPolicyID != FinalAcceptanceRenewalID || a.SourceCommit != g.SourceCommit || a.SourceManifestSHA256 != g.SourceManifestSHA256 || a.BinariesManifestSHA256 != g.BinariesManifestSHA256 || a.PlanSHA256 != g.PlanSHA256 || a.MaxCalls != 378 || a.MaxAttempts != 414 || a.CampaignMaxCalls != 1358 || a.CampaignMaxAttempts != 1454 || a.NewReservedMicrousd != 6900000 || a.RetainedReservedMicrousd != 28800000 || a.RequestReference != human.RequestReference || a.RequestText != human.RequestText || !reflect.DeepEqual(actual.SpendingGuardSourceSHA256, g.SpendingGuardSourceSHA256) {
		return fmt.Errorf("canonical actual recovery5 scope/source/guard differs")
	}
	previous, e := readActualRecoveryApproval(filepath.Join(owner, fourthActualApproval), g.FourthCohort.ActualHumanApprovalSHA256, Recovery4RenewalID, 1070, 1130, 17300000, 12600000)
	if e != nil || a.RequestReference == previous.RequestReference {
		return fmt.Errorf("recovery5 requires a new direct human response reference")
	}
	return recovery5FirstRoute(g, human.ActualHumanApprovalSHA256)
}
