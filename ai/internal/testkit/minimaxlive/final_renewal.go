package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
)

// RenewalFacts exports the current fixed round and the complete canonical
// predecessor selection. It creates nothing and never changes prior budgets.
func RenewalFacts() (RenewalSourceFacts, error) {
	f, e := closedRenewalFacts()
	f.RenewalID = FinalAcceptanceRenewalID
	if e != nil {
		return f, e
	}
	h, rows, e := closedRenewalHistory(f)
	if e != nil {
		return f, e
	}
	f.ClosedHistory = h
	raw, e := json.Marshal(h)
	if e != nil {
		return f, e
	}
	f.ClosedHistorySHA256 = Hash(raw)
	f.PreviousRenewalID = FormalRenewalID
	base, e := renewalBase(FormalRenewalID)
	if e != nil {
		return f, e
	}
	raw, e = ReadFile(filepath.Join(base, "formal.jsonl"), 4<<20)
	if e != nil {
		return f, e
	}
	f.PreviousLedgerSHA256 = Hash(raw)
	f.PreviousLedgerTailSHA256 = rows[len(rows)-1].Hash
	for _, r := range rows {
		if r.Kind == "reserve" {
			f.PreviousCalls++
		} else {
			f.PreviousAttempts++
		}
	}
	e = closedFirstCohort(&f)
	return f, e
}

func approvedFinalRenewal(active bool) (RenewalGrant, string, error) {
	var g RenewalGrant
	base, e := renewalBase(FinalAcceptanceRenewalID)
	if e != nil {
		return g, "", e
	}
	if e = ownedRenewalDirectory(base); e != nil {
		return g, "", e
	}
	raw, e := ReadFile(filepath.Join(base, "grant.json"), 1<<20)
	if e != nil {
		return g, "", fmt.Errorf("canonical final renewal grant unavailable")
	}
	pin := Hash(raw)
	if e = StrictJSON(raw, &g); e != nil {
		return g, "", e
	}
	if !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) {
		return g, "", fmt.Errorf("final renewal source commit invalid")
	}
	f, e := RenewalFacts()
	if e != nil {
		return g, "", e
	}
	if g.FifthCohort != nil || g.FourthCohort != nil || g.ThirdCohort != nil || g.Recovery != nil || g.Schema != 3 || g.RenewalID != f.RenewalID || g.Decision != "APPROVED" || g.CommonRootSHA256 != f.CommonRootSHA256 || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || g.OriginalLedgerSHA256 != f.OriginalLedgerSHA256 || g.OriginalLedgerTailSHA256 != f.OriginalLedgerTailSHA256 || g.OriginalCalls != f.OriginalCalls || g.OriginalAttempts != f.OriginalAttempts || g.PreviousRenewalID != f.PreviousRenewalID || g.PreviousLedgerSHA256 != f.PreviousLedgerSHA256 || g.PreviousLedgerTailSHA256 != f.PreviousLedgerTailSHA256 || g.PreviousCalls != f.PreviousCalls || g.PreviousAttempts != f.PreviousAttempts || g.ClosedHistorySHA256 != f.ClosedHistorySHA256 || !fixedRenewalCaps(g) || !sameFirstCohort(g, f) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("final renewal source/closed history/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	hex := regexp.MustCompile(`^[a-f0-9]{64}$`)
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !hex.MatchString(p) {
			return g, "", fmt.Errorf("final renewal authority pin missing")
		}
	}
	if e = finalRenewalApproval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

func fixedRenewalCaps(g RenewalGrant) bool {
	return g.MaxCalls == 60 && g.MaxAttempts == 72 && g.PerIdentityCalls == 20 && g.PerIdentityAttempts == 24 && g.MaxOutputTokens == 4096 && g.MaxRequestedOutput == 40000 && g.MaxReservedInput == 120000
}

func finalRenewalApproval(base string, g RenewalGrant) error {
	var human struct {
		FirstCohortSealSHA256 string   `json:"firstCohortSealSHA256"`
		CampaignMaxCalls      int      `json:"campaignMaxCalls"`
		CampaignMaxAttempts   int      `json:"campaignMaxAttempts"`
		NewReservedMicrousd   int64    `json:"newReservedMicrousd"`
		Decision              string   `json:"decision"`
		RenewalID             string   `json:"renewalId"`
		MaxCalls              int      `json:"maxCalls"`
		MaxAttempts           int      `json:"maxAttempts"`
		RequestReference      string   `json:"requestReference"`
		RequestText           string   `json:"requestText"`
		HistorySHA256         string   `json:"historySHA256"`
		PlanSHA256            string   `json:"planSHA256"`
		Provider              string   `json:"provider"`
		Model                 string   `json:"model"`
		Purpose               string   `json:"purpose"`
		Routes                []string `json:"routes"`
		RenewalMaxCalls       int      `json:"renewalMaxCalls"`
		RenewalMaxAttempts    int      `json:"renewalMaxAttempts"`
	}
	raw, e := ReadFile(filepath.Join(base, "human-approval.json"), 1<<20)
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || json.Unmarshal(raw, &human) != nil || human.Decision != "APPROVED" || human.RenewalID != g.RenewalID || human.MaxCalls != 722 || human.MaxAttempts != 758 || human.CampaignMaxCalls != 774 || human.CampaignMaxAttempts != 810 || human.NewReservedMicrousd != 28400000 || human.FirstCohortSealSHA256 != g.FirstCohortSealSHA256 || human.RequestReference == "" || human.RequestText == "" || human.HistorySHA256 != g.ClosedHistorySHA256 || human.PlanSHA256 != g.PlanSHA256 || human.Provider != "minimax" || human.Model != M3 || human.Purpose != "standard" || !reflect.DeepEqual(human.Routes, []string{"chat", "responses", "messages"}) || human.RenewalMaxCalls != 60 || human.RenewalMaxAttempts != 72 {
		return fmt.Errorf("own final round human approval and closed history required")
	}
	var review struct {
		FirstCohortSealSHA256  string `json:"firstCohortSealSHA256"`
		Decision               string `json:"decision"`
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
		PlanSHA256             string `json:"planSHA256"`
		HistorySHA256          string `json:"historySHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || json.Unmarshal(raw, &review) != nil || review.Decision != "APPROVED" || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 || review.PlanSHA256 != g.PlanSHA256 || review.HistorySHA256 != g.ClosedHistorySHA256 || review.FirstCohortSealSHA256 != g.FirstCohortSealSHA256 {
		return fmt.Errorf("exact final source/closed history review required")
	}
	return nil
}

func sameFirstCohort(g RenewalGrant, f RenewalSourceFacts) bool {
	return g.FirstCohortSealSHA256 == f.FirstCohortSealSHA256 && g.FirstCohortActualCalls == 40 && g.FirstCohortActualAttempts == 40 && g.FirstCohortClaimedCalls == 52 && g.FirstCohortClaimedAttempts == 52 && g.FirstCohortReservedMicrousd == 1500000 && g.NativeRound1NeverIssued
}
