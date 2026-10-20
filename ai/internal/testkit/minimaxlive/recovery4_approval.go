package minimaxlive

import (
	"fmt"
	"path/filepath"
	"reflect"
)

type recovery4Scope struct {
	recoveryScope
	ThirdCohortSealSHA256    string `json:"thirdCohortSealSHA256"`
	ThirdCohortHistorySHA256 string `json:"thirdCohortHistorySHA256"`
}

func (s recovery4Scope) matches4(g RenewalGrant) bool {
	return g.Recovery != nil && g.ThirdCohort != nil && s.Decision == "APPROVED" && s.RenewalID == Recovery4RenewalID && s.ValidationRound == Recovery4RenewalID && s.NativeRenewalID == Recovery4RenewalID && s.SpendingGuardPolicyID == FinalAcceptanceRenewalID && s.FirstCohortSealSHA256 == g.FirstCohortSealSHA256 && s.SecondCohortSealSHA256 == g.Recovery.SecondCohortSealSHA256 && s.HistorySHA256 == g.ClosedHistorySHA256 && s.RecoveryHistorySHA256 == recoveryHistoryPin(g.Recovery) && s.PlanSHA256 == g.PlanSHA256 && s.ThirdCohortSealSHA256 == g.ThirdCohort.CohortSealSHA256 && s.ThirdCohortHistorySHA256 == thirdHistoryPin(g.ThirdCohort)
}
func recovery4Approval(base string, g RenewalGrant) error {
	var human struct {
		recovery4Scope
		MaxCalls                  int      `json:"maxCalls"`
		MaxAttempts               int      `json:"maxAttempts"`
		CampaignMaxCalls          int      `json:"campaignMaxCalls"`
		CampaignMaxAttempts       int      `json:"campaignMaxAttempts"`
		NewReservedMicrousd       int64    `json:"newReservedMicrousd"`
		RetainedReservedMicrousd  int64    `json:"retainedReservedMicrousd"`
		CampaignMaximumMicrousd   int64    `json:"campaignMaximumMicrousd"`
		RequestReference          string   `json:"requestReference"`
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
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || decodeRecoveryApproval(raw, &human) != nil || !human.matches4(g) || human.MaxCalls != 506 || human.MaxAttempts != 518 || human.CampaignMaxCalls != 1070 || human.CampaignMaxAttempts != 1130 || human.NewReservedMicrousd != 17300000 || human.RetainedReservedMicrousd != 12600000 || human.CampaignMaximumMicrousd != 30000000 || human.RequestReference == "" || human.RequestText == "" || !regexpHex64(human.ActualHumanApprovalSHA256) || human.Provider != "minimax" || human.Model != M3 || human.Purpose != "standard" || !reflect.DeepEqual(human.Routes, []string{"chat", "responses", "messages"}) || human.RenewalMaxCalls != 60 || human.RenewalMaxAttempts != 72 {
		return fmt.Errorf("own recovery4 506/518,1070/1130 actual human approval required")
	}
	var review struct {
		recovery4Scope
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || decodeRecoveryApproval(raw, &review) != nil || !review.matches4(g) || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 {
		return fmt.Errorf("exact recovery4 source/history review required")
	}
	owner, e := evidenceOwner()
	if e != nil {
		return e
	}
	a, e := readActualRecoveryApproval(filepath.Join(owner, ".evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json"), human.ActualHumanApprovalSHA256, Recovery4RenewalID, 1070, 1130, 17300000, 12600000)
	if e != nil || a.SourceCommit != g.SourceCommit || a.SourceManifestSHA256 != g.SourceManifestSHA256 || a.BinariesManifestSHA256 != g.BinariesManifestSHA256 || a.PlanSHA256 != g.PlanSHA256 || a.RequestReference != human.RequestReference || a.RequestText != human.RequestText {
		return fmt.Errorf("canonical actual recovery4 human record differs: %v", e)
	}
	return nil
}

type actualRecoveryApproval struct {
	MaxUSD                   int    `json:"maxUSD"`
	Decision                 string `json:"decision"`
	Round                    string `json:"round"`
	ValidationRound          string `json:"validationRound"`
	NativeRenewalID          string `json:"nativeRenewalId"`
	SpendingGuardPolicyID    string `json:"spendingGuardPolicyId"`
	SourceCommit             string `json:"sourceCommit"`
	SourceManifestSHA256     string `json:"sourceManifestSHA256"`
	BinariesManifestSHA256   string `json:"binariesManifestSHA256"`
	PlanSHA256               string `json:"planSHA256"`
	MaxCalls                 int    `json:"maxCalls"`
	MaxAttempts              int    `json:"maxAttempts"`
	CampaignMaxCalls         int    `json:"campaignMaxCalls"`
	CampaignMaxAttempts      int    `json:"campaignMaxAttempts"`
	NewReservedMicrousd      int64  `json:"newReservedMicrousd"`
	RetainedReservedMicrousd int64  `json:"retainedReservedMicrousd"`
	RequestReference         string `json:"requestReference"`
	RequestText              string `json:"requestText"`
}

func readActualRecoveryApproval(path, pin, id string, calls, attempts int, newUSD, oldUSD int64) (actualRecoveryApproval, error) {
	var a actualRecoveryApproval
	raw, e := ReadFile(path, 1<<20)
	if e != nil || Hash(raw) != pin || decodeRecoveryApproval(raw, &a) != nil || a.MaxUSD != 30 || a.Decision != "APPROVED" || a.Round != id || a.ValidationRound != id || a.NativeRenewalID != id || a.SpendingGuardPolicyID != FinalAcceptanceRenewalID || a.MaxCalls != 506 || a.MaxAttempts != 518 || a.CampaignMaxCalls != calls || a.CampaignMaxAttempts != attempts || a.NewReservedMicrousd != newUSD || a.RetainedReservedMicrousd != oldUSD || a.RequestReference == "" || a.RequestText == "" {
		return a, fmt.Errorf("actual direct recovery human scope differs")
	}
	return a, nil
}
