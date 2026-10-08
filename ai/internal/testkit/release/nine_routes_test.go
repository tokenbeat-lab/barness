package release

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Enumerated in issue 17 FAILURES.md before implementation. The existing
// pure decision seam must reject a matrix even when its history says PASS.
func TestLiveCannotDeleteCapabilities(t *testing.T) {
	in := goodInputs(t)
	in.Matrix.Rows[0].Capabilities = nil
	expectFail(t, Evaluate(in), GateLive, "capabilit")
}

func TestRejectedLiveReportIsNeverRepublished(t *testing.T) {
	in := goodInputs(t)
	in.Audits[1].Live.Endpoint = "FORBIDDEN-RAW-CREDENTIAL"
	in.Audits[1].Err = "audit failed"
	raw, err := json.Marshal(Evaluate(in))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("FORBIDDEN-RAW-CREDENTIAL")) {
		t.Fatal("rejected live report secret republished")
	}
}

func TestLiveCaptureAndObserverCannotBeDeleted(t *testing.T) {
	source := "../../../../.scratch/barness-ai-pi-1.0/google-images-live-evidence/live-pass"
	for _, suffix := range []string{"exchanges.json", "result.json", "image-check.json", "observations"} {
		t.Run(suffix, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS(source)); err != nil {
				t.Fatal(err)
			}
			var manifest map[string]any
			raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &manifest); err != nil {
				t.Fatal(err)
			}
			for _, v := range manifest["cases"].([]any) {
				c := v.(map[string]any)
				inventory := map[string]string{}
				files, err := os.ReadDir(filepath.Join(dir, c["dir"].(string)))
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range files {
					raw, err := os.ReadFile(filepath.Join(dir, c["dir"].(string), f.Name()))
					if err != nil {
						t.Fatal(err)
					}
					inventory[f.Name()] = pioracle.SHA256(raw)
				}
				c["artifacts"] = inventory
			}
			writeArtifact(t, dir, "manifest.json", manifest)
			if a := AuditBundle(dir, nil); a.Err != "" {
				t.Fatalf("complete captured baseline rejected: %s", a.Err)
			}
			err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				if bytes.HasSuffix([]byte(path), []byte(suffix)) || bytes.HasPrefix([]byte(d.Name()), []byte(suffix)) {
					return os.Remove(path)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if a := AuditBundle(dir, nil); a.Err == "" {
				t.Fatal("incomplete live captures accepted")
			}
		})
	}
}

func TestOfflineArtifactsCannotBeDeleted(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]any{{"id": "E01-known", "dir": "E01-known", "status": "PASS", "artifacts": map[string]string{"result.json": "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e"}}}})
	b, err := LoadBundle(dir)
	if err != nil {
		t.Fatal(err)
	}
	in := goodInputs(t)
	in.Offline = b
	expectFail(t, Evaluate(in), GateP0, "artifact")
}

func TestBusinessEffectTraceUsesIndependentArtifact(t *testing.T) {
	in := goodInputs(t)
	in.Trace.Items = append(in.Trace.Items, TraceItem{ID: "C11-ChineseEffect", Source: "US87", Requirement: "independent real Chinese evaluation", Scenarios: []string{"ARTIFACTS"}, Cases: "^chinese-effect$"})
	raw, err := json.Marshal(in.Trace)
	if err != nil {
		t.Fatal(err)
	}
	in.Trace = mustTrace(t, string(raw))
	if got := traceStatus(t, Evaluate(in), "C11-ChineseEffect"); got != Pass {
		t.Fatalf("complete business evidence = %s", got)
	}
	in.Evidence = nil
	if got := traceStatus(t, Evaluate(in), "C11-ChineseEffect"); got != NoEvidence {
		t.Fatalf("protocol PASS substituted for business effect = %s", got)
	}
}

func TestReleasePlanCannotBeReduced(t *testing.T) {
	for _, name := range []string{"combo", "p0", "differential", "route", "requirement", "scenario-mapping", "requirement-mapping"} {
		t.Run(name, func(t *testing.T) {
			trace, err := LoadTraceMap("../../../release/traceability.json")
			if err != nil {
				t.Fatal(err)
			}
			if !CheckReleasePlan(trace).Current {
				t.Fatal("approved release mapping rejected")
			}
			trace.Items[0].Requirement = "editorial wording is not part of the evidence contract"
			switch name {
			case "combo":
				trace.LiveCombos = trace.LiveCombos[:8]
			case "p0":
				trace.P0 = trace.P0[:2]
			case "differential":
				trace.Differential.Required = trace.Differential.Required[:5]
			case "route":
				trace.Differential.Routes = trace.Differential.Routes[:2]
			case "requirement":
				trace.Items = trace.Items[1:]
			case "scenario-mapping":
				trace.Scenarios["P07"] = ".*"
			case "requirement-mapping":
				trace.Items[0].Scenarios = []string{"P01"}
				trace.Items[0].Cases = ".*"
			}
			if check := CheckReleasePlan(trace); check.Current {
				t.Fatal("reduced release contract accepted")
			}
		})
	}
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
