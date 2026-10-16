package supportmatrix

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/jevformal"
)

// JevFormalCombo deliberately cannot match the historical 1.13 validator/row.
const JevFormalCombo = "typesafe-classifier-jev-1.14.0"

// JevEvidence binds two immutable packages. Schema 3 already has per-route
// optional evidence; this adds a new route, without migrating historical bytes.
type JevEvidence struct {
	Mode         string `json:"mode"`
	ProtocolHash string `json:"protocolHash"`
	EffectsHash  string `json:"effectsHash"`
}

func JevExpected() []string {
	var out []string
	for _, entry := range []string{"plain", "hooked"} {
		for _, name := range []string{"default-context", "explicit-context", "ordered-reuse", "local-state-instructions", "visual-criteria", "box-existence"} {
			out = append(out, name+"-"+entry)
		}
	}
	return out
}

// NewJevReport derives every field from independently verified formal captures.
// Controlled pipelines can be inspected but never merged as live support.
func NewJevReport(bundle string) (Report, error) {
	var r Report
	proof, err := jevformal.CheckInputs(filepath.Join(bundle, "protocol"), filepath.Join(bundle, "effects"), false)
	if err != nil {
		return r, err
	}
	p := proof.Protocol
	b := p.Budget.Actual
	r = Report{Schema: SchemaVersion, Combo: JevFormalCombo, Operation: "classifier", Provider: "typesafe", API: "typesafe-system-one", Endpoint: "https://api.typesafe.ai/v1", Model: "jev-1.14.0", SDK: "direct HTTP (fixed Jev formal)", AccountAlias: p.Authorization.AccountAlias + "@" + p.Authorization.RegionAlias, GitCommit: p.GitCommit, CatalogVersion: p.CatalogVersion, CatalogHash: p.CatalogHash, StartedAt: p.StartedAt, FinishedAt: p.FinishedAt, Expected: JevExpected(), Jev: &JevEvidence{p.Mode, proof.ProtocolHash, proof.EffectsHash}, evidenceDir: bundle,
		Budget: Budget{MaxCalls: 12, MaxAttempts: 12, AttemptsReserved: b.Calls, MaxQuestions: 32, MaxImages: 24, MaxDecodedImageBytesTotal: 8 << 20, MaxStateBytesTotal: 4 << 20, ProcessSeconds: 180, CallSeconds: 30, FinalizationReserveSeconds: 10, CallsUsed: b.Calls, HTTPAttempts: b.Attempts, QuestionsUsed: b.Questions, ImagesUsed: b.Sources, DecodedImageBytesSubmitted: b.SourceBytes, StateBytesSubmitted: b.StateBytes, InputTokensUsed: p.Budget.ReportedInputTokens, OutputTokensUsed: p.Budget.ReportedOutputTokens}}
	if !slices.Equal(p.Expected, r.Expected) {
		return Report{}, refuse("formal Jev lacks independent complete capability set")
	}
	for _, id := range r.Expected {
		res, questions, err := jevformal.Result(filepath.Join(bundle, "protocol"), id)
		if err != nil {
			return Report{}, err
		}
		s := ScenarioResult{ID: id, CaseID: "MM11-" + id, Model: r.Model, Outcome: Pass, Calls: 1, Attempts: len(res.Metadata.Attempts), Questions: questions, Images: len(res.ImageInputs), StateBytes: res.StateBytes, DecodedImageBytes: res.ImageBytes, DurationMS: res.DurationMS}
		for _, attempt := range res.Metadata.Attempts {
			s.ProviderRequestIDs = append(s.ProviderRequestIDs, attempt.ProviderRequestID)
		}
		r.Scenarios = append(r.Scenarios, s)
	}
	return r, nil
}

func VerifyJevReport(r Report, requireLive bool) error {
	if r.Combo != JevFormalCombo || r.Jev == nil || r.evidenceDir == "" || (requireLive && r.Jev.Mode != "live") {
		return refuse("Jev report lacks independent model-specific file evidence")
	}
	if err := jevformal.CheckInventory(r.evidenceDir); err != nil {
		return refuse("Jev formal inventory/audit rejected")
	}
	want, err := NewJevReport(r.evidenceDir)
	if err != nil {
		return refuse("Jev independent captures rejected: %v", err)
	}
	a, err := json.Marshal(r)
	if err != nil {
		return err
	}
	b, err := json.Marshal(want)
	if err != nil {
		return err
	}
	if !bytes.Equal(a, b) {
		return refuse("Jev formal report differs from its verified captures, identity or budget")
	}
	return nil
}

func checkJev(r Report) error {
	if r.Combo != JevFormalCombo && r.Jev == nil {
		return nil
	}
	return VerifyJevReport(r, true)
}
