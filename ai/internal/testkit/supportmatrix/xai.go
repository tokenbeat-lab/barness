package supportmatrix

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"regexp"
	"slices"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

const XAIReportFile = "xai-formal-report.json"
const xaiCombo = "xai-responses"
const xaiModel = "grok-4.7"
const xaiEndpoint = "https://api.x.ai/v1"

// XAIExpected is the consumer's fixed own capability set, independent of the
// producer's claimed Expected list and of every other Responses provider.
func XAIExpected() []string {
	return []string{"entry-stream", "entry-complete", "entry-stream-simple", "entry-complete-simple", "effort-default", "effort-low", "effort-medium", "effort-high", "effort-xhigh", "encrypted-continuation", "tools-auto", "tools-required", "tools-none", "tools-parallel", "user-png", "user-jpeg", "tool-result-image", "cache-usage"}
}

type xaiEvidence struct {
	Schema            int              `json:"schema"`
	Kind              string           `json:"kind"`
	Mode              string           `json:"mode"`
	Stage             string           `json:"stage"`
	Combo             string           `json:"combo"`
	Provider          string           `json:"provider"`
	API               string           `json:"api"`
	Endpoint          string           `json:"endpoint"`
	Model             string           `json:"model"`
	Outcome           string           `json:"outcome"`
	Failure           string           `json:"failure,omitempty"`
	Complete          bool             `json:"complete"`
	FormalQualified   bool             `json:"formalQualified"`
	AuditPassed       bool             `json:"auditPassed"`
	StartedAt         time.Time        `json:"startedAt"`
	FinishedAt        time.Time        `json:"finishedAt"`
	GitCommit         string           `json:"gitCommit"`
	SourceSHA256      string           `json:"sourceSha256"`
	BuildSourceSHA256 string           `json:"buildSourceSha256"`
	BinarySHA256      string           `json:"binarySha256"`
	CatalogVersion    string           `json:"catalogVersion"`
	CatalogHash       string           `json:"catalogHash"`
	SDK               string           `json:"sdk"`
	AccountAlias      string           `json:"accountAlias"`
	RegionAlias       string           `json:"regionAlias"`
	GrantID           string           `json:"grantId"`
	GrantSHA256       string           `json:"grantSha256"`
	ClaimID           string           `json:"claimId"`
	Expected          []string         `json:"expected"`
	Scenarios         []ScenarioResult `json:"scenarios"`
	Budget            xaiBudget        `json:"budget"`
}
type xaiBudget struct {
	MaxCalls                   int     `json:"maxCalls"`
	MaxAttempts                int     `json:"maxAttempts"`
	MaxRequestedOutputTokens   int     `json:"maxRequestedOutputTokens"`
	MaxReservedInputTokens     int     `json:"maxReservedInputTokens"`
	CallsUsed                  int     `json:"callsUsed"`
	HTTPAttempts               int     `json:"httpAttempts"`
	VendorHTTPAttempts         int     `json:"vendorHTTPAttempts"`
	RequestedOutputTokens      int     `json:"requestedOutputTokens"`
	ReservedInputTokens        int     `json:"reservedInputTokens"`
	ReportedTokens             int64   `json:"reportedTokens"`
	ObservedMaximumUSD         float64 `json:"observedMaximumUSD"`
	ProcessSeconds             int     `json:"processSeconds"`
	CallSeconds                int     `json:"callSeconds"`
	FinalizationReserveSeconds int     `json:"finalizationReserveSeconds"`
}
type xaiGrant struct {
	Schema                   int       `json:"schema"`
	ID                       string    `json:"id"`
	Mode                     string    `json:"mode"`
	Stage                    string    `json:"stage"`
	Combo                    string    `json:"combo"`
	Provider                 string    `json:"provider"`
	API                      string    `json:"api"`
	Endpoint                 string    `json:"endpoint"`
	Model                    string    `json:"model"`
	SourceSHA256             string    `json:"sourceSha256"`
	BinarySHA256             string    `json:"binarySha256"`
	CatalogVersion           string    `json:"catalogVersion"`
	CatalogHash              string    `json:"catalogHash"`
	SDK                      string    `json:"sdk"`
	AccountAlias             string    `json:"accountAlias"`
	RegionAlias              string    `json:"regionAlias"`
	ExpiresAt                time.Time `json:"expiresAt"`
	Expected                 []string  `json:"expected"`
	MaxCalls                 int       `json:"maxCalls"`
	MaxAttempts              int       `json:"maxAttempts"`
	MaxRequestedOutputTokens int       `json:"maxRequestedOutputTokens"`
	MaxReservedInputTokens   int       `json:"maxReservedInputTokens"`
	MaxUSD                   float64   `json:"maxUSD"`
}
type xaiClaim struct {
	Schema       int       `json:"schema"`
	GrantID      string    `json:"grantId"`
	GrantSHA256  string    `json:"grantSha256"`
	RunID        string    `json:"runId"`
	SourceSHA256 string    `json:"sourceSha256"`
	BinarySHA256 string    `json:"binarySha256"`
	ClaimedAt    time.Time `json:"claimedAt"`
}
type xaiTotals struct {
	input, output, tokens int64
	usd                   float64
	calls                 int
}

// NewXAIReport consumes original formal evidence. It never relabels candidate
// or controlled captures, writes a report, or starts a provider request.
func NewXAIReport(dir string) (Report, error) {
	var zero Report
	before, err := xaiInventory(dir)
	if err != nil {
		return zero, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return zero, err
	}
	defer root.Close()
	var detail xaiEvidence
	if err := xaiReadJSON(root, XAIReportFile, &detail); err != nil {
		return zero, err
	}
	if err := xaiFacts(detail); err != nil {
		return zero, err
	}
	var common Report
	err = xaiReadJSON(root, ReportFile, &common)
	common.evidenceDir = dir
	if err != nil {
		return zero, err
	}
	if err := verifyXAIOrigin(root, dir, detail, before); err != nil {
		return zero, err
	}
	totals, err := verifyXAICases(root, detail, before)
	if err != nil {
		return zero, err
	}
	b := detail.Budget
	if totals.input > int64(^uint(0)>>1) || totals.output > int64(^uint(0)>>1) || totals.calls != 24 || totals.tokens != b.ReportedTokens || !xaiAmountEqual(totals.usd, b.ObservedMaximumUSD) {
		return zero, refuse("xAI captured totals disagree with native budget")
	}
	want := xaiProjection(detail, totals)
	if _, err := RouteCatalog(dir, want); err != nil {
		return zero, err
	}
	want.evidenceDir = dir
	if !reflect.DeepEqual(common, want) {
		return zero, refuse("xAI generic projection differs from own original proof")
	}
	after, err := xaiInventory(dir)
	if err != nil || !reflect.DeepEqual(before, after) {
		return zero, refuse("xAI evidence changed during qualification")
	}
	return common, nil
}

func xaiFacts(r xaiEvidence) error {
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	if r.Schema != 1 || r.Kind != "xai-formal" || r.Stage != "formal" || r.Mode != "live" || r.Combo != xaiCombo || r.Provider != "xai" || r.API != "openai-responses" || r.Endpoint != xaiEndpoint || r.Model != xaiModel || r.Outcome != "PASS" || r.Failure != "" || !r.Complete || !r.FormalQualified || !r.AuditPassed || !alias.MatchString(r.AccountAlias) || !alias.MatchString(r.RegionAlias) || !alias.MatchString(r.GrantID) || r.ClaimID != r.GrantID || !xaiHash(r.SourceSHA256) || r.BuildSourceSHA256 != r.SourceSHA256 || !xaiHash(r.BinarySHA256) || !xaiHash(r.GrantSHA256) || r.GitCommit == "" {
		return refuse("xAI own formal identity, source, grant or outcome invalid")
	}
	if r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.FinishedAt.Sub(r.StartedAt) > 1800*time.Second {
		return refuse("xAI execution time invalid")
	}
	sdk, err := minimaxlive.SDKVersion(minimaxlive.Identity{Route: "chat"})
	if err != nil || sdk != r.SDK {
		return refuse("xAI SDK differs from selected dependency")
	}
	b := r.Budget
	if b.MaxCalls != 24 || b.MaxAttempts != 24 || b.MaxRequestedOutputTokens != 24576 || b.MaxReservedInputTokens != 196608 || b.CallsUsed != 24 || b.HTTPAttempts != 24 || b.VendorHTTPAttempts != 24 || b.RequestedOutputTokens != 24576 || b.ReservedInputTokens != 196608 || b.ReportedTokens < 0 || !xaiFinite(b.ObservedMaximumUSD) || b.ProcessSeconds != 1800 || b.CallSeconds != 65 || b.FinalizationReserveSeconds != 30 {
		return refuse("xAI fixed actual budget invalid")
	}
	if !slices.Equal(r.Expected, XAIExpected()) || len(r.Scenarios) != 18 {
		return refuse("xAI full own scenario set required")
	}
	for i, s := range r.Scenarios {
		calls := len(xaiLabels(s.ID))
		if s.ID != r.Expected[i] || s.CaseID != "E17-xai-"+s.ID || s.Model != xaiModel || s.Outcome != Pass || s.ErrorCategory != "" || s.Environment || s.Calls != calls || s.Attempts != calls || s.Retries != 0 || s.DurationMS < 0 || s.DurationMS > int64(calls*65000) {
			return refuse("xAI scenario substituted or incomplete")
		}
	}
	return nil
}
func xaiProjection(d xaiEvidence, t xaiTotals) Report {
	return Report{Schema: SchemaVersion, Combo: d.Combo, Operation: "chat", Provider: d.Provider, API: d.API, Endpoint: d.Endpoint, Model: d.Model, SDK: d.SDK, AccountAlias: d.AccountAlias, GitCommit: d.GitCommit, CatalogVersion: d.CatalogVersion, CatalogHash: d.CatalogHash, StartedAt: d.StartedAt, FinishedAt: d.FinishedAt, Expected: d.Expected, Scenarios: d.Scenarios, Budget: Budget{MaxCalls: 24, MaxAttempts: 24, MaxRetries: 0, MaxOutputTokens: 1024, AttemptsReserved: 24, MaxRequestedOutputTokens: 24576, RequestedOutputTokens: 24576, MaxReservedInputTokens: 196608, ReservedInputTokens: 196608, CallsUsed: 24, HTTPAttempts: 24, InputTokensUsed: int(t.input), OutputTokensUsed: int(t.output), ProcessSeconds: 1800, CallSeconds: 65, FinalizationReserveSeconds: 30}}
}
func VerifyXAIReport(r Report) error {
	if r.Combo != xaiCombo {
		return nil
	}
	if r.evidenceDir == "" {
		return refuse("xAI support needs its own captured original")
	}
	actual, err := NewXAIReport(r.evidenceDir)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, r) {
		return refuse("xAI projection changed after qualification")
	}
	return nil
}
func checkXAI(r Report) error {
	if r.Combo == xaiCombo {
		return VerifyXAIReport(r)
	}
	return nil
}
func xaiLabels(id string) []string {
	switch id {
	case "encrypted-continuation":
		return []string{"first", "continued"}
	case "tools-auto", "tools-required", "tools-parallel", "tool-result-image":
		return []string{"calls", "results"}
	case "tools-none":
		return []string{"calls"}
	case "user-png", "user-jpeg":
		return []string{"image"}
	case "cache-usage":
		return []string{"first", "repeat"}
	default:
		return []string{"text"}
	}
}
func xaiHash(s string) bool    { return regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(s) }
func xaiFinite(v float64) bool { return v >= 0 && !math.IsNaN(v) && !math.IsInf(v, 0) }
func xaiAmountEqual(a, b float64) bool {
	if !xaiFinite(a) || !xaiFinite(b) {
		return false
	}
	ulp := math.Abs(b - math.Nextafter(b, 0))
	return a == b || math.Abs(a-b) <= math.Max(1e-15, 8*ulp)
}
func xaiEqual(a, b any) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return xaiJSONEqual(left, right)
}
