package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

// Explicit execution IDs prevent the unchanged ID2 USD-policy selector from
// being interpreted as old-round execution authority. Old approval is history.
type recoveryScope struct {
	RenewalID              string `json:"renewalId"`
	ValidationRound        string `json:"validationRound"`
	NativeRenewalID        string `json:"nativeRenewalId"`
	SpendingGuardPolicyID  string `json:"spendingGuardPolicyId"`
	FirstCohortSealSHA256  string `json:"firstCohortSealSHA256"`
	SecondCohortSealSHA256 string `json:"secondCohortSealSHA256"`
	HistorySHA256          string `json:"historySHA256"`
	RecoveryHistorySHA256  string `json:"recoveryHistorySHA256"`
	PlanSHA256             string `json:"planSHA256"`
	Decision               string `json:"decision"`
}

func (s recoveryScope) matches(g RenewalGrant) bool {
	return s.Decision == "APPROVED" && s.RenewalID == RecoveryRenewalID && s.ValidationRound == RecoveryRenewalID && s.NativeRenewalID == RecoveryRenewalID && s.SpendingGuardPolicyID == FinalAcceptanceRenewalID && s.FirstCohortSealSHA256 == g.FirstCohortSealSHA256 && s.SecondCohortSealSHA256 == g.Recovery.SecondCohortSealSHA256 && s.HistorySHA256 == g.ClosedHistorySHA256 && s.RecoveryHistorySHA256 == recoveryHistoryPin(g.Recovery) && s.PlanSHA256 == g.PlanSHA256
}
func recoveryApproval(base string, g RenewalGrant) error {
	var human struct {
		recoveryScope
		MaxCalls                 int      `json:"maxCalls"`
		MaxAttempts              int      `json:"maxAttempts"`
		CampaignMaxCalls         int      `json:"campaignMaxCalls"`
		CampaignMaxAttempts      int      `json:"campaignMaxAttempts"`
		NewReservedMicrousd      int64    `json:"newReservedMicrousd"`
		RetainedReservedMicrousd int64    `json:"retainedReservedMicrousd"`
		CampaignMaximumMicrousd  int64    `json:"campaignMaximumMicrousd"`
		RequestReference         string   `json:"requestReference"`
		RequestText              string   `json:"requestText"`
		Provider                 string   `json:"provider"`
		Model                    string   `json:"model"`
		Purpose                  string   `json:"purpose"`
		Routes                   []string `json:"routes"`
		RenewalMaxCalls          int      `json:"renewalMaxCalls"`
		RenewalMaxAttempts       int      `json:"renewalMaxAttempts"`
	}
	raw, e := ReadFile(filepath.Join(base, "human-approval.json"), 1<<20)
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || decodeRecoveryApproval(raw, &human) != nil || !human.matches(g) || human.MaxCalls != 506 || human.MaxAttempts != 518 || human.CampaignMaxCalls != 962 || human.CampaignMaxAttempts != 1010 || human.NewReservedMicrousd != 19200000 || human.RetainedReservedMicrousd != 10700000 || human.CampaignMaximumMicrousd != 30000000 || human.RequestReference == "" || human.RequestText == "" || human.Provider != "minimax" || human.Model != M3 || human.Purpose != "standard" || !reflect.DeepEqual(human.Routes, []string{"chat", "responses", "messages"}) || human.RenewalMaxCalls != 60 || human.RenewalMaxAttempts != 72 {
		return fmt.Errorf("own recovery506/518,962/1010 actual human approval required")
	}
	var review struct {
		recoveryScope
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || decodeRecoveryApproval(raw, &review) != nil || !review.matches(g) || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 {
		return fmt.Errorf("exact recovery source/history review required")
	}
	return nil
}

// Human/review envelopes allow descriptive fields, but their interpretation
// must be unambiguous before projecting the native authorization fields.
func decodeRecoveryApproval(raw []byte, target any) error {
	var envelope map[string]json.RawMessage
	if err := StrictJSON(raw, &envelope); err != nil {
		return err
	}
	shape, err := json.Marshal(target)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(shape, &fields); err != nil {
		return err
	}
	// encoding/json accepts case-insensitive struct aliases. Descriptions may
	// extend the envelope, but may not supply a second spelling of authority.
	for name := range envelope {
		for field := range fields {
			if name != field && strings.EqualFold(name, field) {
				return fmt.Errorf("noncanonical recovery approval field")
			}
		}
	}
	return json.Unmarshal(raw, target)
}
