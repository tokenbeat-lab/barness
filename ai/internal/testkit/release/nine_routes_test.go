package release

import (
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Enumerated in issue 17 FAILURES.md before implementation. The existing
// pure decision seam must reject a matrix even when its history says PASS.
func TestLiveCannotDeleteCapabilities(t *testing.T) {
	in := goodInputs(t)
	in.Matrix.Rows[0].Capabilities = nil
	expectFail(t, Evaluate(in), GateLive, "capabilit")
}

func TestLiveEvidenceMustMatchCurrentRun(t *testing.T) {
	for _, name := range []string{"no-report", "old-run", "catalog-version", "catalog-hash", "wrong-route", "missing-scenario", "not-run", "different-model"} {
		t.Run(name, func(t *testing.T) {
			in := goodInputs(t)
			r := in.Audits[1].Live
			switch name {
			case "no-report":
				in.Audits[1].Live = nil
			case "old-run":
				r.FinishedAt = r.FinishedAt.Add(-time.Hour)
			case "catalog-version":
				r.CatalogVersion = "old"
			case "catalog-hash":
				r.CatalogHash = "old"
			case "wrong-route":
				r.Operation = "image"
			case "missing-scenario":
				r.Scenarios = r.Scenarios[:1]
			case "not-run":
				r.Scenarios[0].Outcome = supportmatrix.NotRun
			case "different-model":
				r.Model = "substitute"
			}
			expectFail(t, Evaluate(in), GateLive, "openai-responses")
		})
	}
}

func TestEveryEvidencePackageMustBeAudited(t *testing.T) {
	in := goodInputs(t)
	in.RequiredAudits = []string{"/b", "/race"}
	expectFail(t, Evaluate(in), GateAudit, "/race")
}

func TestCannotOmitPressureOrBusinessEvidence(t *testing.T) {
	in := goodInputs(t)
	in.Evidence = nil
	if r := Evaluate(in); r.Pass {
		t.Fatal("missing pressure/effect reports passed release")
	}
}

func TestLiveReportCannotReplaceMissingCaptureManifest(t *testing.T) {
	in := goodInputs(t)
	dir := t.TempDir()
	writeArtifact(t, dir, "live-report.json", in.Audits[1].Live)
	writeArtifact(t, dir, "audit.json", map[string]any{"findings": []any{}})
	writeArtifact(t, dir, "manifest.json", map[string]any{"cases": []any{}})
	if a := AuditBundle(dir, nil); a.Err == "" {
		t.Fatal("claimed live PASS without captured case manifest accepted")
	}
}

func TestLiveRequiredUnaryCannotBeUnsupported(t *testing.T) {
	for _, row := range []supportmatrix.Row{
		{Combo: "typesafe-classifier", Operation: "classifier", Provider: "typesafe", API: "typesafe-system-one", Model: "jev-1.13.0"},
		{Combo: "openai-images", Operation: "image", Provider: "openai", API: "openai-images", Model: "gpt-image-2.5-sunburst-2026-09-08"},
		{Combo: "google-interactions-image", Operation: "image", Provider: "google", API: "google-interactions", Model: "gemini-nano-banana-2.1"},
	} {
		t.Run(row.Combo, func(t *testing.T) {
			in := goodInputs(t)
			row.SDK, row.LastRunAt, row.AllPassedAt = "direct HTTP", passedAt(), passedAt()
			row.Capabilities = []supportmatrix.Capability{{ID: "generation", Model: row.Model, Outcome: supportmatrix.Unsupported, Note: "unavailable"}}
			in.Matrix.Rows = append(in.Matrix.Rows, row)
			in.Trace.LiveCombos = append(in.Trace.LiveCombos, row.Combo)
			in.Audits = append(in.Audits, AuditResult{Bundle: "/unary", Combo: row.Combo})
			expectFail(t, Evaluate(in), GateLive, row.Combo)
		})
	}
}
