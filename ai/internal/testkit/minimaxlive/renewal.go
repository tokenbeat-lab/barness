package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// FormalRenewalID identifies the closed first renewal. Its evidence remains
// readable, but it can never acquire new execution authority.
const FormalRenewalID = "current-source-eleven-renewal-1"

// FinalAcceptanceRenewalID is the only additional, explicitly bounded round.
// Neither ID is a caller-defined round namespace.
const FinalAcceptanceRenewalID = "provider-sites-final-acceptance-2"

type RenewalGrant struct {
	Schema                      int    `json:"schema"`
	RenewalID                   string `json:"renewalId"`
	Decision                    string `json:"decision"`
	CommonRootSHA256            string `json:"commonRootSha256"`
	SourceCommit                string `json:"sourceCommit"`
	ExecutionSourceSHA256       string `json:"executionSourceSha256"`
	OriginalLedgerSHA256        string `json:"originalLedgerSha256"`
	OriginalLedgerTailSHA256    string `json:"originalLedgerTailSha256"`
	OriginalCalls               int    `json:"originalCalls"`
	OriginalAttempts            int    `json:"originalAttempts"`
	MaxCalls                    int    `json:"maxCalls"`
	MaxAttempts                 int    `json:"maxAttempts"`
	PerIdentityCalls            int    `json:"perIdentityCalls"`
	PerIdentityAttempts         int    `json:"perIdentityAttempts"`
	MaxOutputTokens             int    `json:"maxOutputTokens"`
	MaxRequestedOutput          int    `json:"maxRequestedOutput"`
	MaxReservedInput            int    `json:"maxReservedInput"`
	HumanApprovalSHA256         string `json:"humanApprovalSha256"`
	SourceReviewSHA256          string `json:"sourceReviewSha256"`
	SourceManifestSHA256        string `json:"sourceManifestSha256"`
	BinariesManifestSHA256      string `json:"binariesManifestSha256"`
	PlanSHA256                  string `json:"planSha256"`
	PreviousRenewalID           string `json:"previousRenewalId,omitempty"`
	PreviousLedgerSHA256        string `json:"previousLedgerSha256,omitempty"`
	PreviousLedgerTailSHA256    string `json:"previousLedgerTailSha256,omitempty"`
	PreviousCalls               int    `json:"previousCalls,omitempty"`
	PreviousAttempts            int    `json:"previousAttempts,omitempty"`
	ClosedHistorySHA256         string `json:"closedHistorySha256,omitempty"`
	FirstCohortSealSHA256       string `json:"firstCohortSealSha256,omitempty"`
	FirstCohortActualCalls      int    `json:"firstCohortActualCalls,omitempty"`
	FirstCohortActualAttempts   int    `json:"firstCohortActualAttempts,omitempty"`
	FirstCohortClaimedCalls     int    `json:"firstCohortClaimedCalls,omitempty"`
	FirstCohortClaimedAttempts  int    `json:"firstCohortClaimedAttempts,omitempty"`
	FirstCohortReservedMicrousd int64  `json:"firstCohortReservedMicrousd,omitempty"`
	NativeRound1NeverIssued     bool   `json:"nativeRound1NeverIssued,omitempty"`
}
type RenewalSourceFacts struct {
	RenewalID                   string            `json:"renewalId"`
	CommonRootSHA256            string            `json:"commonRootSha256"`
	SourceCommit                string            `json:"sourceCommit"`
	ExecutionSourceSHA256       string            `json:"executionSourceSha256"`
	OriginalLedgerSHA256        string            `json:"originalLedgerSha256"`
	OriginalLedgerTailSHA256    string            `json:"originalLedgerTailSha256"`
	OriginalCalls               int               `json:"originalCalls"`
	OriginalAttempts            int               `json:"originalAttempts"`
	PreviousRenewalID           string            `json:"previousRenewalId,omitempty"`
	PreviousLedgerSHA256        string            `json:"previousLedgerSha256,omitempty"`
	PreviousLedgerTailSHA256    string            `json:"previousLedgerTailSha256,omitempty"`
	PreviousCalls               int               `json:"previousCalls,omitempty"`
	PreviousAttempts            int               `json:"previousAttempts,omitempty"`
	ClosedHistorySHA256         string            `json:"closedHistorySha256,omitempty"`
	ClosedHistory               map[string]string `json:"closedHistory,omitempty"`
	FirstCohortSealSHA256       string            `json:"firstCohortSealSha256,omitempty"`
	FirstCohortActualCalls      int               `json:"firstCohortActualCalls,omitempty"`
	FirstCohortActualAttempts   int               `json:"firstCohortActualAttempts,omitempty"`
	FirstCohortClaimedCalls     int               `json:"firstCohortClaimedCalls,omitempty"`
	FirstCohortClaimedAttempts  int               `json:"firstCohortClaimedAttempts,omitempty"`
	FirstCohortReservedMicrousd int64             `json:"firstCohortReservedMicrousd,omitempty"`
	NativeRound1NeverIssued     bool              `json:"nativeRound1NeverIssued,omitempty"`
}

func renewalBase(id string) (string, error) {
	old, e := originalStageLedgerPath("formal")
	if e != nil {
		return "", e
	}
	return filepath.Join(filepath.Dir(filepath.Dir(old)), "minimax-renewals", id), nil
}
func currentCommit() (string, error) {
	root, e := Root()
	if e != nil {
		return "", e
	}
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = root
	c.Env = []string{}
	b, e := c.Output()
	return strings.TrimSpace(string(b)), e
}

func closedRenewalFacts() (RenewalSourceFacts, error) {
	f := RenewalSourceFacts{RenewalID: FormalRenewalID}
	old, e := originalStageLedgerPath("formal")
	if e != nil {
		return f, e
	}
	owner := filepath.Dir(filepath.Dir(filepath.Dir(old)))
	f.CommonRootSHA256 = Hash([]byte(owner))
	f.SourceCommit, e = currentCommit()
	if e != nil {
		return f, e
	}
	source, e := ExecutionSource()
	if e != nil {
		return f, e
	}
	raw, e := json.Marshal(source)
	if e != nil {
		return f, e
	}
	f.ExecutionSourceSHA256 = Hash(raw)
	raw, e = ReadFile(old, 4<<20)
	if e != nil {
		return f, e
	}
	f.OriginalLedgerSHA256 = Hash(raw)
	rows, e := ReadLedger(old)
	if e != nil {
		return f, e
	}
	for _, r := range rows {
		id, e := Select(r.Identity.Route, M3, "standard")
		if e != nil || id != r.Identity || r.Stage != "formal" || r.RenewalID != "" {
			return f, fmt.Errorf("original formal identity changed")
		}
		if r.Kind == "reserve" {
			f.OriginalCalls++
		} else {
			f.OriginalAttempts++
		}
		f.OriginalLedgerTailSHA256 = r.Hash
	}
	if f.OriginalCalls != 39 || f.OriginalAttempts != 36 {
		return f, fmt.Errorf("original39/36 formal history required")
	}
	return f, nil
}
func approvedRenewal(id string, active bool) (RenewalGrant, string, error) {
	if id == FinalAcceptanceRenewalID {
		return approvedFinalRenewal(active)
	}
	if id == ClosedFinalAcceptanceRenewalID || active && id == FormalRenewalID {
		return RenewalGrant{}, "", fmt.Errorf("closed MiniMax renewal cannot execute again")
	}
	return approvedClosedRenewal(id)
}

// Historical verification retains the original schema and pins. It never
// reopens the closed round or transfers any of its unused budget.
// Keep its approval checks independent: final-round fields and caps must not
// silently broaden the already closed schema-1 evidence contract.
func approvedClosedRenewal(id string) (RenewalGrant, string, error) {
	var g RenewalGrant
	if id != FormalRenewalID {
		return g, "", fmt.Errorf("unknown fixed MiniMax formal renewal")
	}
	base, e := renewalBase(id)
	if e != nil {
		return g, "", e
	}
	// Authority directories must be owned canonical entries, never bundle links.
	for p := base; p != filepath.Dir(filepath.Dir(filepath.Dir(base))); p = filepath.Dir(p) {
		s, e := os.Lstat(p)
		if e != nil || !s.IsDir() {
			return g, "", fmt.Errorf("canonical renewal authority unavailable")
		}
	}
	raw, e := ReadFile(filepath.Join(base, "grant.json"), 1<<20)
	if e != nil {
		return g, "", fmt.Errorf("canonical approved renewal grant unavailable")
	}
	pin := Hash(raw)
	if e = StrictJSON(raw, &g); e != nil {
		return g, "", e
	}
	if !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) {
		return g, "", fmt.Errorf("renewal source commit invalid")
	}
	f, e := closedRenewalFacts()
	if e != nil {
		return g, "", e
	}
	if g.Schema != 1 || g.RenewalID != id || g.Decision != "APPROVED" || g.CommonRootSHA256 != f.CommonRootSHA256 || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || g.OriginalLedgerSHA256 != f.OriginalLedgerSHA256 || g.OriginalLedgerTailSHA256 != f.OriginalLedgerTailSHA256 || g.OriginalCalls != 39 || g.OriginalAttempts != 36 || g.MaxCalls != 60 || g.MaxAttempts != 72 || g.PerIdentityCalls != 20 || g.PerIdentityAttempts != 24 || g.MaxOutputTokens != 4096 || g.MaxRequestedOutput != 40000 || g.MaxReservedInput != 120000 || !closedGrantOnly(g) {
		return g, "", fmt.Errorf("fixed renewal source/history/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	hex := regexp.MustCompile(`^[a-f0-9]{64}$`)
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256} {
		if !hex.MatchString(p) {
			return g, "", fmt.Errorf("renewal authority pin missing")
		}
	}
	var human struct {
		Decision         string `json:"decision"`
		RenewalID        string `json:"renewalId"`
		MaxCalls         int    `json:"maxCalls"`
		MaxAttempts      int    `json:"maxAttempts"`
		RequestReference string `json:"requestReference"`
	}
	raw, e = ReadFile(filepath.Join(base, "human-approval.json"), 1<<20)
	if e != nil || Hash(raw) != g.HumanApprovalSHA256 || json.Unmarshal(raw, &human) != nil || human.Decision != "APPROVED" || human.RenewalID != id || human.MaxCalls != 216 || human.MaxAttempts != 252 || human.RequestReference == "" {
		return g, "", fmt.Errorf("new human216/252 approval required")
	}
	var review struct {
		Decision               string `json:"decision"`
		SourceCommit           string `json:"sourceCommit"`
		SourceManifestSHA256   string `json:"sourceManifestSHA256"`
		BinariesManifestSHA256 string `json:"binariesManifestSHA256"`
		PlanSHA256             string `json:"planSHA256"`
	}
	raw, e = ReadFile(filepath.Join(base, "source-review.json"), 1<<20)
	if e != nil || Hash(raw) != g.SourceReviewSHA256 || json.Unmarshal(raw, &review) != nil || review.Decision != "APPROVED" || review.SourceCommit != g.SourceCommit || review.SourceManifestSHA256 != g.SourceManifestSHA256 || review.BinariesManifestSHA256 != g.BinariesManifestSHA256 || review.PlanSHA256 != g.PlanSHA256 {
		return g, "", fmt.Errorf("exact source renewal review required")
	}
	return g, pin, nil
}

// RenewalAuthority is read-only. Active execution additionally binds HEAD;
// later publication may follow unselected data commits, never selected drift.
func RenewalAuthority(id string, active bool) (RenewalGrant, string, error) {
	return approvedRenewal(id, active)
}
func RenewalLedgerPath(id string) (string, error) {
	if id != FormalRenewalID && id != FinalAcceptanceRenewalID {
		return "", fmt.Errorf("unknown renewal")
	}
	base, e := renewalBase(id)
	return filepath.Join(base, "formal.jsonl"), e
}

func renewalCommittedSource(commit string) error {
	root, e := Root()
	if e != nil {
		return e
	}
	c := exec.Command("git", "diff", "--name-only", commit, "--", "ai", "go.mod", "go.sum")
	c.Dir = root
	c.Env = []string{}
	raw, e := c.Output()
	if e != nil {
		return fmt.Errorf("renewal source commit unavailable")
	}
	for _, name := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if executionSourceName(name) {
			return fmt.Errorf("renewal selected source is uncommitted or stale")
		}
	}
	source, e := ExecutionSource()
	if e != nil {
		return e
	}
	args := []string{"ls-files", "--error-unmatch", "--"}
	for name := range source {
		args = append(args, name)
	}
	c = exec.Command("git", args...)
	c.Dir = root
	c.Env = []string{}
	if _, e = c.Output(); e != nil {
		return fmt.Errorf("renewal selected source is not committed")
	}
	return nil
}
