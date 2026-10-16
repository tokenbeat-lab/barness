package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestCurrentReleaseVisualEffects(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(route+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC14-visual-formal-"+route+"-"+entry)
				out := filepath.Join(run.Dir(), "current-visual-"+route+"-"+entry)
				output, err := visualCommand(t, "-formal", "-route", route, "-entry", entry, "-out", out)
				ev.Check("formal fixed visual task crosses public Classify", err == nil, "formal producer unavailable: %s", output)
				if err != nil {
					return
				}
				r := visualRead(t, out, "report.json")
				ev.Record("report", r)
				ev.Check("own current catalog and no invented live quality", r["catalogKind"] == "formal" && r["catalogHash"] == builtinCatalogHash && r["realQuality"] == "NOT_RUN" && r["status"] == "COMPLETE", "wrong task identity")
				_, err = visualCommand(t, "-verify", out)
				ev.Check("independent formal replay verifies", err == nil, "formal verification failed")
				report := releaseProcess(t, ev, "-visual-evaluation-"+route, out)
				e := releaseEvidence(t, report, "visual-effect-"+route)
				ev.Check("controlled visual evidence never passes real gate", e.Status == release.NotRun, "got %s %s", e.Status, e.Detail)
			})
		}
	}
}

func TestCurrentReleaseVisualCorruption(t *testing.T) {
	source := filepath.Join(run.Dir(), "visual-corruption-source")
	if output, err := visualCommand(t, "-formal", "-route", "jev", "-out", source); err != nil {
		t.Fatalf("visual producer: %s", output)
	}
	for _, mode := range []string{"wrong-route", "candidate", "model", "budget", "source", "missing-capture", "audit", "forged-live"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "UC14-visual-import-"+mode)
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS(source)); err != nil {
				t.Fatal(err)
			}
			raw, _ := os.ReadFile(filepath.Join(dir, "report.json"))
			var r map[string]any
			mustUnmarshal(t, raw, &r)
			route := "jev"
			switch mode {
			case "wrong-route":
				route = "decisions"
			case "candidate":
				delete(r, "catalogKind")
			case "model":
				r["model"] = "jev-1.13.0"
			case "budget":
				r["budget"].(map[string]any)["calls"] = 13
			case "source":
				r["codeHash"] = "sha256:untrusted"
			case "forged-live":
				r["mode"] = "live"
				r["realQuality"] = "COMPLETE"
			}
			writeReleaseJSON(t, dir, "report.json", r)
			raw, _ = os.ReadFile(filepath.Join(dir, "manifest.json"))
			var m map[string]any
			mustUnmarshal(t, raw, &m)
			files := m["artifacts"].(map[string]any)
			stored, _ := os.ReadFile(filepath.Join(dir, "report.json"))
			files["report.json"] = pioracle.SHA256(stored)
			file := map[string]string{"missing-capture": "samples/v01/response.bin", "audit": "audit.json"}[mode]
			if file != "" {
				os.Remove(filepath.Join(dir, file))
				delete(files, file)
			}
			writeReleaseJSON(t, dir, "manifest.json", m)
			report := releaseProcess(t, ev, "-visual-evaluation-"+route, dir)
			e := releaseEvidence(t, report, "visual-effect-"+route)
			ev.Check("invalid visual package fails its own contract", e.Status == release.Fail, "got %s %s", e.Status, e.Detail)
		})
	}
}
