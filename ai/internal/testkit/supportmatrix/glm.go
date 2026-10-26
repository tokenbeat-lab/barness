package supportmatrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
)

const GLMSourcePin = "837a83c4711dc4914b8110341403c0d6e43cac254b3101f601d6d538b209c907"

type GLMEvidence struct {
	Mode                 string         `json:"mode"`
	SourceManifestHash   string         `json:"sourceManifestHash"`
	UsageBasis           string         `json:"usageBasis"`
	UnknownUsageAttempts int            `json:"unknownUsageAttempts"`
	Derivation           *GLMDerivation `json:"derivation,omitempty"`
}

func GLMExpected(combo string) []string {
	route := strings.TrimPrefix(combo, "zhipu-")
	if route != "chat" && route != "responses" && route != "messages" {
		return nil
	}
	var out []string
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		kinds := []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "tool-call", "tool-result", "native-history", "cancel", "recovery"}
		switch route {
		case "chat":
			kinds = append(kinds, "control-temperature", "control-usage")
		case "responses":
			kinds = append(kinds, "control-low", "control-max")
		case "messages":
			kinds = append(kinds, "control-unsigned")
			if model == "glm-5.3" {
				kinds = append(kinds, "control-auto")
			} else {
				kinds = append(kinds, "control-default")
			}
		}
		for _, kind := range kinds {
			out = append(out, model+"/"+kind)
		}
		if grokhistory.Applies(combo) {
			out = append(out, model+"/"+grokhistory.Scenario)
		}
	}
	return out
}

type glmDetail struct {
	Schema                                                                                                                                int
	Kind, Outcome, Route, API, Endpoint, SDK, CatalogVersion, CatalogHash, SourceManifestHash, GitCommit, Mode, AccountAlias, RegionAlias string
	Complete, AuditPassed                                                                                                                 bool
	StartedAt, FinishedAt                                                                                                                 time.Time
	Expected                                                                                                                              []string
	Models                                                                                                                                map[string]string
	Budget                                                                                                                                struct{ PlannedCalls, MaxCalls, MaxAttempts, MaxOutputTokens, CallSeconds, ProcessSeconds, FinalizationReserveSeconds, CallsUsed, AttemptsReserved, HTTPAttempts, Retries, ReplayAttempts int }
	Scenarios                                                                                                                             []struct {
		ID, Model, Outcome, Failure, RequestID string
		Calls, HTTPAttempts, ReplayAttempts    int
		DurationMS                             int64
		ProviderRequestIDs, RequestIDSources   []string
		Truncated                              bool
	}
}

// NewGLMReport derives the formal report from its sealed own route capture.
// Candidate, controlled and replay documents cannot enter this boundary.
func NewGLMReport(dir string) (Report, error) { return glmCaptureReport(dir, "live") }

// VerifyGLMControlledCapture checks the complete sealed synthetic capture.
// It returns no support report; NewGLMReport still requires actual live authority.
func VerifyGLMControlledCapture(dir string) error {
	r, e := glmCaptureReport(dir, "controlled")
	if e != nil {
		return e
	}
	for _, s := range r.Scenarios {
		if s.Outcome != Pass {
			return glmBad("controlled full capture contains a failed scene")
		}
	}
	return nil
}

func glmCaptureReport(dir, mode string) (Report, error) {
	var d glmDetail
	if err := readGLMJSON(filepath.Join(dir, "formal-glm-report.json"), &d); err != nil {
		return Report{}, err
	}
	combo := "zhipu-" + d.Route
	expected := GLMExpected(combo)
	endpoint, api := "", ""
	switch d.Route {
	case "chat":
		endpoint, api = "https://open.bigmodel.cn/api/paas/v4", "openai-completions"
	case "responses":
		endpoint, api = "https://open.bigmodel.cn/api/v1", "openai-responses"
	case "messages":
		endpoint, api = "https://open.bigmodel.cn/api/anthropic", "anthropic-messages"
	}
	cat := ai.BuiltinCatalog()
	hash, _ := cat.Hash()
	if d.Schema != 1 || d.Kind != "glm-formal" || d.Mode != mode || !d.AuditPassed || d.SourceManifestHash != GLMSourcePin || len(expected) == 0 || d.API != api || d.Endpoint != endpoint || d.CatalogVersion != cat.Version || d.CatalogHash != hash || d.GitCommit == "" || d.SDK == "" || d.AccountAlias == "" || d.StartedAt.IsZero() || d.FinishedAt.Before(d.StartedAt) || d.FinishedAt.Sub(d.StartedAt) > 1800*time.Second {
		return Report{}, refuse("GLM formal identity, source, mode, catalog or execution provenance invalid")
	}
	if mode == "live" && (combo == "zhipu-chat" || combo == "zhipu-messages") {
		if e := VerifyProcessAuthorization(dir, combo, d.AccountAlias, d.RegionAlias, d.StartedAt, d.FinishedAt); e != nil {
			return Report{}, e
		}
	}
	b := d.Budget
	if b.PlannedCalls != len(expected) || b.MaxCalls != 24 || b.MaxAttempts != 32 || b.MaxOutputTokens != 4096 || b.CallSeconds != 120 || b.ProcessSeconds != 1800 || b.FinalizationReserveSeconds != 20 || b.CallsUsed < 0 || b.CallsUsed > 24 || b.AttemptsReserved != b.CallsUsed || b.HTTPAttempts < 0 || b.HTTPAttempts > b.AttemptsReserved || b.Retries != 0 || b.ReplayAttempts != 0 {
		return Report{}, refuse("GLM fixed budget invalid")
	}
	if !slices.Equal(d.Expected, expected) || len(d.Scenarios) != len(expected) {
		return Report{}, refuse("GLM complete model-qualified scenario set required")
	}
	r := Report{Schema: SchemaVersion, Combo: combo, Operation: "chat", Provider: "zhipu", API: api, Endpoint: endpoint, Model: "glm-5.3", SDK: d.SDK, AccountAlias: d.AccountAlias + "@" + d.RegionAlias, GitCommit: d.GitCommit, CatalogVersion: cat.Version, CatalogHash: hash, StartedAt: d.StartedAt, FinishedAt: d.FinishedAt, Expected: expected, GLM: &GLMEvidence{Mode: d.Mode, SourceManifestHash: d.SourceManifestHash, UsageBasis: "reported-known-lower-bound"}, evidenceDir: dir,
		Budget: Budget{MaxCalls: 24, MaxAttempts: 32, MaxOutputTokens: 4096, ProcessSeconds: 1800, CallSeconds: 120, FinalizationReserveSeconds: 20, CallsUsed: b.CallsUsed, AttemptsReserved: b.AttemptsReserved, HTTPAttempts: b.HTTPAttempts}}
	allPassed := true
	calls, attempts := 0, 0
	for i, s := range d.Scenarios {
		model, _, _ := strings.Cut(s.ID, "/")
		if s.ID != expected[i] || s.Model != model || s.Outcome == string(Unsupported) || s.Calls < 0 || s.Calls > 1 || s.HTTPAttempts < 0 || s.HTTPAttempts > s.Calls || s.DurationMS < 0 || s.DurationMS > 120000 || s.Truncated || s.ReplayAttempts != 0 {
			return Report{}, refuse("GLM scenario identity or counters invalid")
		}
		outcome := Outcome(s.Outcome)
		if !outcome.valid() {
			return Report{}, refuse("GLM scenario outcome invalid")
		}
		if outcome == Pass && (s.Calls != 1 || s.HTTPAttempts != 1 || len(s.ProviderRequestIDs) != 1 || len(s.RequestIDSources) != 1 || (s.ProviderRequestIDs[0] == "" && s.RequestIDSources[0] != "unreported") || (s.ProviderRequestIDs[0] != "" && !strings.HasPrefix(s.RequestIDSources[0], "response header:"))) {
			return Report{}, refuse("GLM PASS lacks actual provider attempt")
		}
		if outcome == NotRun && (s.Calls != 0 || s.HTTPAttempts != 0) {
			return Report{}, refuse("GLM NOT_RUN spent resources")
		}
		sc := ScenarioResult{ID: s.ID, CaseID: "GLM11-" + strings.ReplaceAll(s.ID, "/", "-"), Model: model, Outcome: outcome, Calls: s.Calls, Attempts: s.HTTPAttempts, DurationMS: s.DurationMS, ProviderRequestIDs: s.ProviderRequestIDs}
		if outcome == Fail {
			sc.ErrorCategory = CategoryAssertion
		}
		r.Scenarios = append(r.Scenarios, sc)
		calls += s.Calls
		attempts += s.HTTPAttempts
		allPassed = allPassed && outcome == Pass
	}
	if calls != b.CallsUsed || attempts != b.HTTPAttempts || d.Complete != allPassed || (allPassed && d.Outcome != "PASS") {
		return Report{}, refuse("GLM actual counters or completeness disagree")
	}
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		pass := true
		for _, s := range d.Scenarios {
			if s.Model == model && s.Outcome != "PASS" {
				pass = false
			}
		}
		if (d.Models[model] == "PASS") != pass {
			return Report{}, refuse("GLM own model summary disagrees")
		}
	}
	if err := verifyGLMBundle(dir, r, d); err != nil {
		return Report{}, err
	}
	if err := glmKnownUsage(dir, &r); err != nil {
		return Report{}, err
	}
	return r, nil
}

func VerifyGLMReport(r Report) error {
	if r.GLM == nil || r.evidenceDir == "" {
		return refuse("GLM formal report requires owned file evidence")
	}
	var want Report
	var err error
	if r.GLM.Derivation != nil {
		want, err = NewGLMSupportViewReport(r.evidenceDir)
	} else {
		want, err = NewGLMReport(r.evidenceDir)
	}
	if err != nil {
		return err
	}
	a, _ := json.Marshal(r)
	b, _ := json.Marshal(want)
	if !bytes.Equal(a, b) {
		return refuse("GLM report differs from its independent sealed captures")
	}
	return nil
}
func checkGLM(r Report) error {
	if len(GLMExpected(r.Combo)) == 0 && r.GLM == nil {
		return nil
	}
	return VerifyGLMReport(r)
}

func glmBad(note string) error { return fmt.Errorf("%w: GLM %s", ErrInvalidReport, note) }
