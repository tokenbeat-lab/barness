package e2e

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func grokHistoryRegionalRefusals(t *testing.T, original, combo string) {
	if !grokhistory.Applies(combo) {
		return
	}
	for _, mutation := range []string{"missing-scene", "substituted-scene", "wire-marker", "immutable-input", "native-envelope"} {
		t.Run("grok-"+mutation, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-NATIVE-"+combo+"-"+mutation)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if e := os.CopyFS(dir, os.DirFS(original)); e != nil {
				t.Fatal(e)
			}
			model := supportmatrix.RegionalModels(combo)[0]
			caseID := supportmatrix.RegionalCaseID(combo, model+"/"+grokhistory.Scenario)
			name := supportmatrix.RegionalReportFile
			switch mutation {
			case "wire-marker":
				name = caseID + "/exchanges.json"
			case "immutable-input":
				name = caseID + "/immutable-input.json"
			case "native-envelope":
				name = caseID + "/history-provenance.json"
			}
			raw, e := os.ReadFile(filepath.Join(dir, name))
			if e != nil {
				t.Fatal(e)
			}
			if mutation == "native-envelope" {
				raw = []byte(strings.ReplaceAll(string(raw), "synthetic-xai-source-account", "synthetic-xai-forged-account"))
			} else if mutation == "wire-marker" || mutation == "immutable-input" {
				raw = []byte(strings.ReplaceAll(string(raw), "SYNTHETIC_XAI_RESULT_FIRST", "SYNTHETIC_XAI_RESULT_OTHER"))
			} else {
				var report map[string]any
				mustUnmarshal(t, raw, &report)
				rows := report["scenarios"].([]any)
				for i, v := range rows {
					row := v.(map[string]any)
					if row["id"] != model+"/"+grokhistory.Scenario {
						continue
					}
					if mutation == "missing-scene" {
						rows = append(rows[:i], rows[i+1:]...)
					} else {
						row["id"] = model + "/native-history"
					}
					break
				}
				report["scenarios"] = rows
				raw = mustMarshal(t, report)
			}
			if e = os.WriteFile(filepath.Join(dir, name), raw, 0600); e != nil {
				t.Fatal(e)
			}
			if strings.Contains(name, "/") {
				kimiQualificationManifest(t, dir, caseID, filepath.Base(name), raw)
			}
			if e = supportmatrix.SealRegionalBundle(dir); e != nil {
				t.Fatal(e)
			}
			_, e = supportmatrix.VerifyRegionalControlledCapture(dir)
			ev.Record("synthetic-mutation", map[string]any{"name": mutation, "refused": e != nil, "error": e, "originalRetained": true, "vendorCalls": 0})
			ev.Check("complete native consumer refuses missing/substituted/changed xAI proof", e != nil, "accepted mutation")
		})
	}
}
