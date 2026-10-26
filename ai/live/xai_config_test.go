//go:build live

package live

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
)

const xaiEntry = "^TestXAICandidate$"
const xaiFormalEntry = "^TestXAIFormal$"
const xaiCombo = "xai-responses"
const xaiEndpoint = "https://api.x.ai/v1"
const xaiCallOutput = 1024

// Set with -ldflags for the frozen candidate or formal binary used in paid runs.
var xaiBuildSource string
var xaiState *xaiProcess

type xaiProcess struct {
	proxy                            *url.URL
	report                           xaiReport
	cat                              ai.Catalog
	grant                            xaiGrant
	grantRaw                         []byte
	key, problem, notRun, controlled string
	totalInput, totalOutput          int64
}

type xaiReport struct {
	Schema            int                            `json:"schema"`
	Kind              string                         `json:"kind"`
	Mode              string                         `json:"mode"`
	Stage             string                         `json:"stage"`
	Combo             string                         `json:"combo"`
	Provider          string                         `json:"provider"`
	API               string                         `json:"api"`
	Endpoint          string                         `json:"endpoint"`
	Model             string                         `json:"model"`
	Outcome           string                         `json:"outcome"`
	Failure           string                         `json:"failure,omitempty"`
	Complete          bool                           `json:"complete"`
	FormalQualified   bool                           `json:"formalQualified"`
	AuditPassed       bool                           `json:"auditPassed"`
	StartedAt         time.Time                      `json:"startedAt"`
	FinishedAt        time.Time                      `json:"finishedAt"`
	GitCommit         string                         `json:"gitCommit"`
	SourceSHA256      string                         `json:"sourceSha256"`
	BuildSourceSHA256 string                         `json:"buildSourceSha256"`
	BinarySHA256      string                         `json:"binarySha256"`
	CatalogVersion    string                         `json:"catalogVersion"`
	CatalogHash       string                         `json:"catalogHash"`
	SDK               string                         `json:"sdk"`
	AccountAlias      string                         `json:"accountAlias,omitempty"`
	RegionAlias       string                         `json:"regionAlias,omitempty"`
	GrantID           string                         `json:"grantId,omitempty"`
	GrantSHA256       string                         `json:"grantSha256,omitempty"`
	ClaimID           string                         `json:"claimId,omitempty"`
	Expected          []string                       `json:"expected"`
	Scenarios         []supportmatrix.ScenarioResult `json:"scenarios"`
	Budget            xaiBudget                      `json:"budget"`
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

func xaiSelected() bool {
	return os.Getenv(envCombo) == xaiCombo || (flag.Lookup("test.run").Value.String() == xaiEntry || flag.Lookup("test.run").Value.String() == xaiFormalEntry) || os.Getenv(keyVar(xaiCombo)) != ""
}
func xaiSourceFiles() (map[string]string, error) {
	// This complete runtime/harness inventory grants no regional identity.
	files, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		return nil, err
	}
	root, err := minimaxlive.Root()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(root, "ai/internal/testkit/xai/candidate.json"))
	if err != nil {
		return nil, err
	}
	files["ai/internal/testkit/xai/candidate.json"] = sha256Text(raw)
	return files, nil
}
func xaiSource() (string, error) {
	files, err := xaiSourceFiles()
	if err != nil {
		return "", err
	}
	return supportmatrix.RegionalSourceHash(files), nil
}
func (p *xaiProcess) entry() string {
	if p.report.Stage == "formal" {
		return xaiFormalEntry
	}
	return xaiEntry
}
func (p *xaiProcess) reportName() string {
	if p.report.Stage == "formal" {
		return "xai-formal-report"
	}
	return "xai-candidate-report"
}
func initializeXAI() {
	p := &xaiProcess{cat: xai.CandidateCatalog(), controlled: os.Getenv("BARNESS_XAI_CONTROLLED")}
	stage, kind := "candidate", "xai-candidate-only"
	if flag.Lookup("test.run").Value.String() == xaiFormalEntry {
		stage, kind, p.cat = "formal", "xai-formal", ai.BuiltinCatalog()
	}
	xaiState = p
	hash, hashErr := p.cat.Hash()
	source, sourceErr := xaiSource()
	binary, binaryErr := regionalBinaryHash()
	mode := "live"
	if p.controlled != "" {
		mode = "controlled"
	}
	p.report = xaiReport{Schema: 1, Kind: kind, Mode: mode, Stage: stage, Combo: xaiCombo, Provider: "xai", API: "openai-responses", Endpoint: xaiEndpoint, Model: "grok-4.7", Outcome: "NOT_RUN", StartedAt: started, GitCommit: evidence.GitCommit(), SourceSHA256: source, BuildSourceSHA256: xaiBuildSource, BinarySHA256: binary, CatalogVersion: p.cat.Version, CatalogHash: hash, SDK: openAISDK(), Expected: xaiExpected(), Budget: xaiBudget{MaxCalls: 24, MaxAttempts: 24, MaxRequestedOutputTokens: 24576, MaxReservedInputTokens: 196608, ProcessSeconds: 1800, CallSeconds: 65, FinalizationReserveSeconds: 30}}
	for k, v := range map[string]string{"model_catalog_version": p.cat.Version, "model_catalog_hash": hash, "live_combo": xaiCombo, "live_sdk": openAISDK(), "xai_execution_source": source} {
		run.SetVersion(k, v)
	}
	for _, id := range p.report.Expected {
		p.report.Scenarios = append(p.report.Scenarios, supportmatrix.ScenarioResult{ID: id, Model: "grok-4.7", CaseID: "E17-xai-" + id, Outcome: supportmatrix.NotRun})
	}
	if stage == "formal" {
		if _, ok := p.cat.Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, "grok-4.7"); !ok {
			p.problem = "formal Builtin model unavailable"
			return
		}
	}
	if stage == "formal" && sourceErr == nil {
		files, e := xaiSourceFiles()
		if e != nil || run.Record("xai-execution-source", files) != nil {
			p.problem = "formal source inventory unavailable"
			return
		}
	}
	if os.Getenv(envCombo) != xaiCombo {
		if os.Getenv(envCombo) == "" {
			p.notRun = "explicit xAI route absent"
		} else {
			p.problem = "selected route differs from xAI process"
		}
		return
	}
	if hashErr != nil || sourceErr != nil || binaryErr != nil {
		p.problem = "execution identity unavailable"
		return
	}
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value == "" {
			continue
		}
		if name != keyVar(xaiCombo) && (strings.HasSuffix(name, "_KEY") || strings.HasSuffix(name, "_TOKEN") || strings.HasSuffix(name, "_SECRET") || strings.HasPrefix(name, envKeyBase)) || slices.Contains([]string{"BARNESS_REGIONAL_ROUTE", "BARNESS_GLM_ROUTE", "BARNESS_MINIMAX_ROUTE", "BARNESS_MM_ROUTE", "BARNESS_AI_LIVE_ENDPOINT", "OPENAI_API_BASE"}, name) || strings.HasSuffix(name, "_BASE_URL") {
			p.problem = "mixed routes, credentials or endpoint overrides forbidden"
			return
		}
	}
	if p.controlled != "" && !slices.Contains([]string{"fixture", "cancel", "final-write", "final-sync", "final-audit", "projection-write", "final-manifest"}, p.controlled) {
		p.problem = "unknown controlled mode"
		return
	}
	if slices.Contains([]string{"projection-write", "final-manifest"}, p.controlled) && stage != "formal" {
		p.problem = "formal finalization mode requires formal entry"
		return
	}
	if p.controlled != "" && (os.Getenv(envLive) != "" || os.Getenv(keyVar(xaiCombo)) != "") {
		p.problem = "controlled process forbids live switch and credentials"
		return
	}
	if raw := os.Getenv("BARNESS_AI_LIVE_PROXY"); raw != "" {
		var err error
		p.proxy, err = smokeProxy(raw)
		if err != nil {
			p.problem = "explicit candidate proxy invalid"
			return
		}
	}
	if p.controlled == "" && os.Getenv(envLive) != "1" {
		p.notRun = "explicit live switch absent"
		return
	}
	if p.controlled == "" && strings.TrimSpace(os.Getenv(keyVar(xaiCombo))) == "" {
		p.notRun = "selected credential absent"
		return
	}
	path := os.Getenv("BARNESS_AI_LIVE_PERMISSIONS_FILE")
	if path == "" {
		p.notRun = "explicit one-process grant absent"
		return
	}
	raw, err := readDecisionsFile(path, 64<<10)
	if err != nil || xaiDecodeGrant(raw, &p.grant) != nil || !p.validGrant() {
		p.problem = "xAI identity, stage, aliases, expiry, source or finite budget grant invalid"
		return
	}
	p.grantRaw = raw
	p.report.GrantID = p.grant.ID
	p.report.GrantSHA256 = sha256Text(raw)
	p.report.AccountAlias = p.grant.AccountAlias
	p.report.RegionAlias = p.grant.RegionAlias
	if p.controlled == "" && xaiBuildSource != source {
		p.problem = "live binary requires explicit frozen source injection"
		return
	}
	if p.controlled == "" && flag.Lookup("test.timeout").Value.String() != "30m0s" {
		p.problem = "live xAI requires exact 1800s process limit"
		return
	}
	p.key = os.Getenv(keyVar(xaiCombo))
	if p.controlled != "" {
		p.key = "synthetic-xai-controlled"
	}
	run.RedactSecret(p.key, "[XAI-CREDENTIAL]")
	for k, v := range map[string]string{"model_catalog_version": p.cat.Version, "model_catalog_hash": hash, "live_combo": xaiCombo, "live_sdk": openAISDK(), "xai_execution_source": source} {
		run.SetVersion(k, v)
	}
}
func (p *xaiProcess) validGrant() bool {
	g, r := p.grant, p.report
	b := r.Budget
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	now := time.Now().UTC()
	return g.Schema == 1 && alias.MatchString(g.ID) && g.Mode == r.Mode && g.Stage == r.Stage && g.Combo == r.Combo && g.Provider == r.Provider && g.API == r.API && g.Endpoint == r.Endpoint && g.Model == r.Model && g.SourceSHA256 == r.SourceSHA256 && g.BinarySHA256 == r.BinarySHA256 && g.CatalogVersion == r.CatalogVersion && g.CatalogHash == r.CatalogHash && g.SDK == r.SDK && alias.MatchString(g.AccountAlias) && alias.MatchString(g.RegionAlias) && g.AccountAlias == os.Getenv(envAlias) && g.RegionAlias == os.Getenv("BARNESS_AI_LIVE_REGION_ALIAS") && g.ExpiresAt.After(now) && g.ExpiresAt.Sub(now) <= 24*time.Hour && slices.Equal(g.Expected, r.Expected) && g.MaxCalls == b.MaxCalls && g.MaxAttempts == b.MaxAttempts && g.MaxRequestedOutputTokens == b.MaxRequestedOutputTokens && g.MaxReservedInputTokens == b.MaxReservedInputTokens && g.MaxUSD > 0 && !math.IsNaN(g.MaxUSD) && !math.IsInf(g.MaxUSD, 0)
}
func xaiDecodeGrant(raw []byte, g *xaiGrant) error {
	// Authorization controls require exact keys once each, without nulls.
	d := json.NewDecoder(bytes.NewReader(raw))
	tok, err := d.Token()
	if err != nil || tok != json.Delim('{') {
		return fmt.Errorf("grant object required")
	}
	seen := map[string]bool{}
	canonicalRaw, _ := json.Marshal(xaiGrant{})
	var canonical map[string]json.RawMessage
	_ = json.Unmarshal(canonicalRaw, &canonical)
	for d.More() {
		tok, err = d.Token()
		if err != nil {
			return err
		}
		key, ok := tok.(string)
		if _, known := canonical[key]; !ok || seen[key] || !known {
			return fmt.Errorf("duplicate grant field")
		}
		seen[key] = true
		var value json.RawMessage
		if err = d.Decode(&value); err != nil {
			return err
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("null grant field")
		}
	}
	if _, err = d.Token(); err != nil {
		return err
	}
	if _, err = d.Token(); err != io.EOF {
		return fmt.Errorf("trailing grant data")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(g)
}
