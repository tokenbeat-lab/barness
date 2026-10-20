package minimaxlive

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
)

const ClosedFinalAcceptanceRenewalID = "provider-sites-final-acceptance-1"
const firstCohort = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-1/"

type cohortNativeHistory struct {
	Facts               string            `json:"facts"`
	FactsSHA256         string            `json:"factsSHA256"`
	Command             string            `json:"command"`
	CommandSHA256       string            `json:"commandSHA256"`
	ClosedHistory       map[string]string `json:"closedHistory"`
	ClosedHistorySHA256 string            `json:"closedHistorySHA256"`
	OriginalCalls       int               `json:"originalCalls"`
	OriginalAttempts    int               `json:"originalAttempts"`
	PreviousCalls       int               `json:"previousCalls"`
	PreviousAttempts    int               `json:"previousAttempts"`
	Historical          bool              `json:"countsAreHistoricalNotAdditionalCampaignConsumption"`
}

type cohortSeal struct {
	Schema                        int                 `json:"schema"`
	Kind                          string              `json:"kind"`
	CohortIndexSHA256             string              `json:"cohortIndexSHA256"`
	CohortIndex                   string              `json:"cohortIndex"`
	NativeHistory                 cohortNativeHistory `json:"nativeHistory"`
	NativeRound1NeverIssued       bool                `json:"nativeRound1NeverIssued"`
	CampaignActualCalls           int                 `json:"campaignActualCalls"`
	CampaignActualHTTPAttempts    int                 `json:"campaignActualHTTPAttempts"`
	CampaignClaimedCalls          int                 `json:"campaignClaimedCalls"`
	CampaignClaimedHTTPAttempts   int                 `json:"campaignClaimedHTTPAttempts"`
	CampaignWholeReservedMicrousd int64               `json:"campaignWholeReservedMicrousd"`
	NewApproval                   bool                `json:"remainingPaidScopeRequiresNewApproval"`
}

type cohortIndex struct {
	Schema                           int               `json:"schema"`
	Kind                             string            `json:"kind"`
	RoundID                          string            `json:"roundId"`
	SourceCommit                     string            `json:"sourceCommit"`
	CreatedAt                        string            `json:"createdAt"`
	State                            string            `json:"state"`
	Reason                           string            `json:"reason"`
	Research                         json.RawMessage   `json:"research"`
	Processes                        []json.RawMessage `json:"processes"`
	ActualCalls                      int               `json:"actualCalls"`
	ActualHTTPAttempts               int               `json:"actualHTTPAttempts"`
	ClaimedCalls                     int               `json:"claimedCalls"`
	ClaimedHTTPAttempts              int               `json:"claimedHTTPAttempts"`
	WholeReservedMicrousd            int64             `json:"wholeReservedMicrousd"`
	RefundMicrousd                   int64             `json:"refundMicrousd"`
	NoProcessRestarts                bool              `json:"noProcessRestarts"`
	UnstartedCanonicalClaims         []string          `json:"unstartedCanonicalClaims"`
	NativeMiniMaxRound1Absent        []string          `json:"nativeMiniMaxRound1Absent"`
	OriginalNativeM3HistoryUnchanged bool              `json:"originalNativeM3HistoryUnchanged"`
	SourceArchive                    string            `json:"sourceArchive"`
	Files                            map[string]string `json:"files"`
	PaidContinuationAuthorized       bool              `json:"paidContinuationAuthorized"`
}

// The only predecessor is the failed first campaign. Its immutable own facts
// are independent of the original15 M3 consumption facts, and grant/human/review
// bind this seal separately. No caller chooses another round or closure path.
func closedFirstCohort(f *RenewalSourceFacts) error {
	old, err := originalStageLedgerPath("formal")
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Dir(filepath.Dir(filepath.Dir(old))))
	if err != nil {
		return err
	}
	defer root.Close()
	var seal cohortSeal
	raw, err := cohortJSON(root, firstCohort+"seal.json", &seal)
	if err != nil || seal.Schema != 1 || seal.Kind != "CLOSED_FIRST_COHORT_HISTORY_SEAL" || seal.CohortIndex != "index.json" || !seal.NativeRound1NeverIssued || !seal.NewApproval || seal.CampaignActualCalls != 40 || seal.CampaignActualHTTPAttempts != 40 || seal.CampaignClaimedCalls != 52 || seal.CampaignClaimedHTTPAttempts != 52 || seal.CampaignWholeReservedMicrousd != 1500000 {
		return fmt.Errorf("fixed closed first-cohort seal40/40,52/52,1.5USD required: %v", err)
	}
	f.FirstCohortSealSHA256 = Hash(raw)
	var index cohortIndex
	raw, err = cohortJSON(root, firstCohort+"index.json", &index)
	if err != nil || Hash(raw) != seal.CohortIndexSHA256 || index.Schema != 1 || index.Kind != "CLOSED_FAILED_REGIONAL_CANDIDATE_COHORT" || index.RoundID != ClosedFinalAcceptanceRenewalID || index.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || !index.NoProcessRestarts || index.PaidContinuationAuthorized || !index.OriginalNativeM3HistoryUnchanged || index.ActualCalls != 40 || index.ActualHTTPAttempts != 40 || index.ClaimedCalls != 52 || index.ClaimedHTTPAttempts != 52 || index.WholeReservedMicrousd != 1500000 || index.RefundMicrousd != 0 || index.SourceArchive != "source" || len(index.Files) == 0 || len(index.Files) > 4096 || len(index.Processes) != 2 {
		return fmt.Errorf("fixed closed failed cohort index required: %v", err)
	}
	total := int64(0)
	for name, want := range index.Files {
		got, size, e := cohortFileHash(root, name, maxCohortBytes-total)
		if e != nil || got != want {
			return fmt.Errorf("closed cohort pinned file differs: %s: %v", name, e)
		}
		total += size
	}
	if err = cohortArchive(root, firstCohort+"source", index.Files); err != nil {
		return err
	}
	if err = unusedFirstNative(root, index); err != nil {
		return err
	}
	h := seal.NativeHistory
	if h.OriginalCalls != 39 || h.OriginalAttempts != 36 || h.PreviousCalls != 39 || h.PreviousAttempts != 36 || !h.Historical || len(h.ClosedHistory) != 15 || !reflect.DeepEqual(h.ClosedHistory, f.ClosedHistory) || h.ClosedHistorySHA256 != f.ClosedHistorySHA256 {
		return fmt.Errorf("separate original15 native histories39/36 twice differ")
	}
	var facts RenewalSourceFacts
	raw, err = cohortJSON(root, firstCohort+h.Facts, &facts)
	if err != nil || Hash(raw) != h.FactsSHA256 || facts.RenewalID != ClosedFinalAcceptanceRenewalID || facts.ClosedHistorySHA256 != f.ClosedHistorySHA256 || !reflect.DeepEqual(facts.ClosedHistory, f.ClosedHistory) || facts.OriginalCalls != 39 || facts.OriginalAttempts != 36 || facts.PreviousCalls != 39 || facts.PreviousAttempts != 36 {
		return fmt.Errorf("closed native facts changed: %v", err)
	}
	var command map[string]json.RawMessage
	raw, err = cohortJSON(root, firstCohort+h.Command, &command)
	var code int
	var secret bool
	if err != nil || Hash(raw) != h.CommandSHA256 || json.Unmarshal(command["exitCode"], &code) != nil || code != 0 || json.Unmarshal(command["credentialsRead"], &secret) != nil || secret {
		return fmt.Errorf("closed native read-only command receipt differs")
	}
	f.FirstCohortActualCalls, f.FirstCohortActualAttempts = 40, 40
	f.FirstCohortClaimedCalls, f.FirstCohortClaimedAttempts = 52, 52
	f.FirstCohortReservedMicrousd, f.NativeRound1NeverIssued = 1500000, true
	return nil
}

func unusedFirstNative(root *os.Root, index cohortIndex) error {
	base := ".evidence/minimax-renewals/" + ClosedFinalAcceptanceRenewalID + "/"
	directory, err := root.Open(strings.TrimSuffix(base, "/"))
	if err != nil {
		return err
	}
	entries, err := directory.ReadDir(4)
	directory.Close()
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if err != nil && err != io.EOF || !slices.Equal(names, []string{"launch.lock", "operator-authority.json"}) || index.Files[base+"operator-authority.json"] == "" {
		return fmt.Errorf("native1 grant/ledger/nonce/intent must never exist")
	}
	_, size, err := cohortFileHash(root, base+"launch.lock", maxCohortFile)
	if err != nil || size != 0 {
		return fmt.Errorf("closed native1 lock changed")
	}
	want := []string{}
	for _, name := range []string{"grant.json", "human-approval.json", "source-review.json", "formal.jsonl", "chat.json", "responses.json", "messages.json"} {
		want = append(want, base+name)
	}
	sort.Strings(want)
	given := slices.Clone(index.NativeMiniMaxRound1Absent)
	sort.Strings(given)
	if !slices.Equal(given, want) {
		return fmt.Errorf("native1 exact absence declaration required")
	}
	// The parent campaign also owns these fixed31 never-started processes.
	jobs := strings.Fields("openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 glm-chat glm-responses glm-messages mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages text-jev text-luna visual-jev visual-decisions candidate-zai-responses candidate-minimax-cn-chat candidate-minimax-cn-responses candidate-minimax-cn-messages formal-zai-chat formal-zai-responses formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages")
	wanted := []string{}
	for _, name := range jobs {
		wanted = append(wanted, ".evidence/minimax-launches/"+ClosedFinalAcceptanceRenewalID+"-"+name+".json")
	}
	given = slices.Clone(index.UnstartedCanonicalClaims)
	sort.Strings(given)
	sort.Strings(wanted)
	if !slices.Equal(given, wanted) {
		return fmt.Errorf("closed first campaign exact31 unstarted parent claims required")
	}
	for _, name := range wanted {
		if _, e := root.Lstat(name); !os.IsNotExist(e) {
			return fmt.Errorf("closed first campaign acquired a new claim")
		}
	}
	return nil
}
