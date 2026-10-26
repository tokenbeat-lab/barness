//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The standard view is a projection, not a second qualification algorithm.
// NewXAIReport independently consumes the own report and capture triplets.
func (p *xaiProcess) formalProjection() (supportmatrix.Report, error) {
	r, b := p.report, p.report.Budget
	if r.Stage != "formal" || !r.Complete || r.Outcome != "PASS" || p.totalInput < 0 || p.totalOutput < 0 || p.totalInput > math.MaxInt || p.totalOutput > math.MaxInt {
		return supportmatrix.Report{}, fmt.Errorf("complete formal accounting cannot be projected")
	}
	scenarios := slices.Clone(r.Scenarios)
	for i := range scenarios {
		scenarios[i].ProviderRequestIDs = slices.Clone(scenarios[i].ProviderRequestIDs)
	}
	return supportmatrix.Report{
		Schema: supportmatrix.SchemaVersion, Combo: r.Combo, Operation: "chat", Provider: r.Provider, API: r.API,
		Endpoint: r.Endpoint, Model: r.Model, SDK: r.SDK, AccountAlias: r.AccountAlias,
		GitCommit: r.GitCommit, CatalogVersion: r.CatalogVersion, CatalogHash: r.CatalogHash,
		StartedAt: r.StartedAt, FinishedAt: r.FinishedAt, Expected: slices.Clone(r.Expected), Scenarios: scenarios,
		Budget: supportmatrix.Budget{
			MaxCalls: b.MaxCalls, MaxAttempts: b.MaxAttempts, MaxRetries: 0, MaxOutputTokens: xaiCallOutput,
			CallsUsed: b.CallsUsed, HTTPAttempts: b.HTTPAttempts, AttemptsReserved: b.CallsUsed,
			MaxRequestedOutputTokens: b.MaxRequestedOutputTokens, RequestedOutputTokens: b.RequestedOutputTokens,
			MaxReservedInputTokens: b.MaxReservedInputTokens, ReservedInputTokens: b.ReservedInputTokens,
			InputTokensUsed: int(p.totalInput), OutputTokensUsed: int(p.totalOutput),
			ProcessSeconds: b.ProcessSeconds, CallSeconds: b.CallSeconds, FinalizationReserveSeconds: b.FinalizationReserveSeconds,
		},
	}, nil
}

// Formal evidence records the public event discriminator as well as its
// concrete fields. The candidate artifact layout remains independent.
func (s *xaiSession) recordEvents(label string, events []ai.Event) {
	if s.p.report.Stage != "formal" {
		s.cs.Record(label+"-events", events)
		return
	}
	typed := make([]map[string]json.RawMessage, 0, len(events))
	for _, event := range events {
		raw, err := json.Marshal(event)
		if !s.cs.Check("formal event artifact encodes", err == nil, "event artifact unavailable") {
			return
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(raw, &fields); !s.cs.Check("formal event fields decode", err == nil && fields != nil, "event artifact unavailable") {
			return
		}
		kind, err := json.Marshal(event.Type())
		if !s.cs.Check("formal event kind encodes", err == nil, "event artifact unavailable") {
			return
		}
		fields["type"] = kind
		typed = append(typed, fields)
	}
	s.cs.Record(label+"-events", typed)
}
