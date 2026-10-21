package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
)

func ownedRenewalDirectory(base string) error {
	for p := base; p != filepath.Dir(filepath.Dir(filepath.Dir(base))); p = filepath.Dir(p) {
		s, e := os.Lstat(p)
		if e != nil || !s.IsDir() {
			return fmt.Errorf("canonical renewal authority unavailable")
		}
	}
	return nil
}

// The predecessor is selected here, never by a submitted manifest or a path in
// a report. Its entire fixed authority directory and all six formal launcher
// receipts are pinned, including the original formal's consumption.
func closedRenewalHistory(f RenewalSourceFacts) (map[string]string, []Receipt, error) {
	base, e := renewalBase(FormalRenewalID)
	if e != nil {
		return nil, nil, e
	}
	if e = ownedRenewalDirectory(base); e != nil {
		return nil, nil, e
	}
	names := []string{"chat.json", "finalize-intent.json", "formal.jsonl", "grant.json", "human-approval.json", "messages.json", "responses.json", "source-review.json"}
	entries, e := os.ReadDir(base)
	if e != nil || len(entries) != len(names) {
		return nil, nil, fmt.Errorf("closed renewal authority file set differs")
	}
	actual := make([]string, len(entries))
	for i, d := range entries {
		actual[i] = d.Name()
	}
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, names) {
		return nil, nil, fmt.Errorf("closed renewal authority file set differs")
	}
	evidence := filepath.Dir(filepath.Dir(base))
	if info, e := os.Lstat(filepath.Join(evidence, "minimax-stages")); e != nil || !info.IsDir() {
		return nil, nil, fmt.Errorf("original formal canonical directory unavailable")
	}
	history := map[string]string{}
	read := func(name string, max int64) ([]byte, error) {
		raw, e := ReadFile(filepath.Join(evidence, name), max)
		if e == nil {
			history[name] = Hash(raw)
		}
		return raw, e
	}
	for _, n := range names {
		if _, e = read(filepath.Join("minimax-renewals", FormalRenewalID, n), 4<<20); e != nil {
			return nil, nil, e
		}
	}
	original, e := read("minimax-stages/formal.jsonl", 4<<20)
	if e != nil || Hash(original) != f.OriginalLedgerSHA256 {
		return nil, nil, fmt.Errorf("closed original formal differs")
	}
	grantRaw, e := ReadFile(filepath.Join(base, "grant.json"), 1<<20)
	var g RenewalGrant
	if e != nil || StrictJSON(grantRaw, &g) != nil || g.Schema != 1 || g.RenewalID != FormalRenewalID || g.Decision != "APPROVED" || g.CommonRootSHA256 != f.CommonRootSHA256 || g.OriginalLedgerSHA256 != f.OriginalLedgerSHA256 || g.OriginalLedgerTailSHA256 != f.OriginalLedgerTailSHA256 || g.OriginalCalls != 39 || g.OriginalAttempts != 36 || !fixedRenewalCaps(g) || !closedGrantOnly(g) {
		return nil, nil, fmt.Errorf("closed renewal grant/history differs")
	}
	if history[filepath.Join("minimax-renewals", FormalRenewalID, "human-approval.json")] != g.HumanApprovalSHA256 || history[filepath.Join("minimax-renewals", FormalRenewalID, "source-review.json")] != g.SourceReviewSHA256 {
		return nil, nil, fmt.Errorf("closed renewal approval pins differ")
	}
	var closed struct {
		State             string `json:"state"`
		NoAutomaticReruns bool   `json:"noAutomaticReruns"`
	}
	raw, _ := ReadFile(filepath.Join(base, "finalize-intent.json"), 1<<20)
	if json.Unmarshal(raw, &closed) != nil || closed.State != "COMPLETED" || !closed.NoAutomaticReruns {
		return nil, nil, fmt.Errorf("previous renewal is not closed")
	}
	rows, e := ReadLedger(filepath.Join(base, "formal.jsonl"))
	if e != nil {
		return nil, nil, e
	}
	sessions, e := closedFormalSessions(rows, FormalRenewalID, Hash(grantRaw))
	if e != nil {
		return nil, nil, e
	}
	originalRows, e := ReadLedger(filepath.Join(evidence, "minimax-stages/formal.jsonl"))
	if e != nil {
		return nil, nil, e
	}
	originalSessions, e := closedFormalSessions(originalRows, "", "")
	if e != nil {
		return nil, nil, e
	}
	launches := filepath.Join(evidence, "minimax-launches")
	if info, err := os.Lstat(launches); err != nil || !info.IsDir() {
		return nil, nil, fmt.Errorf("closed launcher canonical directory unavailable")
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := Select(route, M3, "standard")
		raw, _ := ReadFile(filepath.Join(base, route+".json"), 16384)
		var claim renewalClaim
		if StrictJSON(raw, &claim) != nil || claim != (renewalClaim{Schema: 1, RenewalID: FormalRenewalID, GrantSHA256: Hash(grantRaw), Session: sessions[route], Identity: id}) {
			return nil, nil, fmt.Errorf("closed renewal route claim differs")
		}
		for _, previous := range []bool{false, true} {
			name, session := "mm-global-formal-m3-"+route, originalSessions[route]
			if previous {
				name, session = FormalRenewalID+"-"+name, sessions[route]
			}
			raw, e = read("minimax-launches/"+name+".json", 2<<20)
			if e != nil || closedFormalLauncher(raw, name, session, id, previous, g, Hash(grantRaw)) != nil {
				return nil, nil, fmt.Errorf("closed formal launcher differs")
			}
		}
	}
	return history, rows, nil
}

func closedGrantOnly(g RenewalGrant) bool {
	return g.SixthCohort == nil && g.FifthCohort == nil && g.FourthCohort == nil && g.ThirdCohort == nil && g.Recovery == nil && g.PreviousRenewalID == "" && g.PreviousLedgerSHA256 == "" && g.PreviousLedgerTailSHA256 == "" && g.PreviousCalls == 0 && g.PreviousAttempts == 0 && g.ClosedHistorySHA256 == "" && g.FirstCohortSealSHA256 == "" && g.FirstCohortActualCalls == 0 && g.FirstCohortActualAttempts == 0 && g.FirstCohortClaimedCalls == 0 && g.FirstCohortClaimedAttempts == 0 && g.FirstCohortReservedMicrousd == 0 && !g.NativeRound1NeverIssued
}

func closedFormalSessions(rows []Receipt, renewal, pin string) (map[string]string, error) {
	sessions := map[string]string{}
	calls, attempts := 0, 0
	for _, r := range rows {
		if r.Stage != "formal" || r.RenewalID != renewal || r.RenewalGrantSHA256 != pin || r.Identity.Model != M3 || r.Identity.Purpose != "standard" || r.Session == "" || filepath.Base(r.Session) != r.Session {
			return nil, fmt.Errorf("closed formal ledger authority differs")
		}
		if prior := sessions[r.Identity.Route]; prior != "" && prior != r.Session {
			return nil, fmt.Errorf("closed formal identity changed session")
		}
		sessions[r.Identity.Route] = r.Session
		if r.Kind == "reserve" {
			calls++
		} else {
			attempts++
		}
	}
	if len(sessions) != 3 || calls != 39 || attempts != 36 {
		return nil, fmt.Errorf("closed formal complete39/36 required")
	}
	return sessions, nil
}

func closedFormalLauncher(raw []byte, name, session string, id Identity, previous bool, g RenewalGrant, pin string) error {
	var r struct {
		State           string `json:"state"`
		Route           string `json:"route"`
		SourceCommit    string `json:"sourceCommit"`
		ExitCode        int    `json:"exitCode"`
		CredentialsRead bool   `json:"credentialsRead"`
		Audit           struct {
			Passed bool `json:"passed"`
		} `json:"launcherAudit"`
		Origin Origin `json:"origin"`
	}
	if json.Unmarshal(raw, &r) != nil || r.State != "COMPLETED" || r.Route != name || r.ExitCode != 0 || !r.CredentialsRead || !r.Audit.Passed || r.Origin.Schema != 1 || r.Origin.Stage != "formal" || r.Origin.Identity != id || r.Origin.Session != session || r.Origin.SourceCommit != r.SourceCommit {
		return fmt.Errorf("closed formal launcher identity differs")
	}
	if previous {
		if r.SourceCommit != g.SourceCommit || r.Origin.FormalRenewalID != FormalRenewalID || r.Origin.FormalRenewalGrantSHA256 != pin || r.Origin.SourceManifestSHA256 != g.SourceManifestSHA256 || r.Origin.BinariesManifestSHA256 != g.BinariesManifestSHA256 || r.Origin.PlanSHA256 != g.PlanSHA256 {
			return fmt.Errorf("closed renewal launcher approval differs")
		}
	} else if r.Origin.FormalRenewalID != "" || r.Origin.FormalRenewalGrantSHA256 != "" {
		return fmt.Errorf("original launcher cannot borrow renewal")
	}
	return nil
}
