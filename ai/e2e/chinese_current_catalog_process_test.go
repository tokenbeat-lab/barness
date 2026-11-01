package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

// Current route defaults follow the generated default snapshot. Historical
// reports keep their recorded catalogs and cannot acquire current applicability.
func TestChineseEvaluationCurrentRoutePins(t *testing.T) {
	for _, tc := range []struct{ route, model string }{{"jev", "jev-1.13.0"}, {"luna", "gpt-6-luna"}} {
		t.Run(tc.route, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-current-"+tc.route)
			out := t.TempDir()
			output, err := evaluationCommand(t, "-route", tc.route, "-out", out)
			ev.Record("disabled-current-route-process", map[string]any{"exitZero": err == nil, "output": string(output), "actualVendorRequests": 0})
			if !ev.Check("current default config loads before disabled execution", err == nil, "%v", err) {
				return
			}
			dir, _ := evaluationBundle(t, out, ev)
			raw, err := os.ReadFile(filepath.Join(dir, "evaluation-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report chineseeval.Report
			mustUnmarshal(t, raw, &report)
			raw, err = os.ReadFile("../catalog_builtin.json")
			if err != nil {
				t.Fatal(err)
			}
			var pin struct {
				Hash    string
				Catalog struct{ Version string }
			}
			mustUnmarshal(t, raw, &pin)
			ev.Check("disabled workflow keeps current catalog and exact route", report.Status == "NOT_RUN" && report.Budget.Calls == 0 && report.Config.Model == tc.model && report.Config.CatalogVersion == pin.Catalog.Version && report.Config.CatalogHash == pin.Hash, "current configuration or zero-call outcome differs")
		})
	}
}

// Prior defaults are immutable historical inputs. Current defaults name a new
// document revision when a root catalog pin changes; no loader repairs them.
func TestChineseEvaluationConfigurationRevisions(t *testing.T) {
	data, err := os.ReadFile("../examples/chineseeval/testdata/zh-support-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ base, oldHash, previousHash string }{
		{"jev-1.13.0", "sha256:001556ea3be2a33e5b9f06a5530733fc9088a8942d073b51b8e07d33f1665ed3", "sha256:d97d7f6baa5d71814285610a80afed7fc6ba01e1826b5ca67a822d4d012f258c"},
		{"jev-route", "sha256:b4501a5274b2092c09169395d5a870b6049f809f34c42510bdcabdf960b9c297", "sha256:383321f05b3dc8825c228ba568cea7594a2f142fd616c3faa22a247c9fb66ff1"},
		{"luna-route", "sha256:7679d0bcdda0698900623353eafe3203541e1bc15725a1c362ed66407f7b4af7", "sha256:1cf4e178b2f31226ce0156c9194621df34b3057778878df980e8136d95ff2583"},
	} {
		t.Run(tc.base, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-config-revision-"+tc.base)
			old, err := os.ReadFile("../examples/chineseeval/testdata/" + tc.base + "-v1.json")
			if err != nil {
				t.Fatal(err)
			}
			previous, err := os.ReadFile("../examples/chineseeval/testdata/" + tc.base + "-v2.json")
			if err != nil {
				t.Fatal(err)
			}
			ev.Check("prior default v2 bytes remain immutable", chineseeval.Hash(previous) == tc.previousHash, "v2 input changed")
			_, err = chineseeval.Load(data, previous, ai.BuiltinCatalog())
			ev.Check("prior default v2 is not silently rebound to current", err != nil, "v2 accepted current composition")
			snapshot, err := os.ReadFile("../release/catalog-snapshot.json")
			if err != nil {
				t.Fatal(err)
			}
			var recorded struct {
				Catalog json.RawMessage `json:"catalog"`
			}
			if err := json.Unmarshal(snapshot, &recorded); err != nil {
				t.Fatal(err)
			}
			_, err = chineseeval.LoadRecorded(data, previous, recorded.Catalog)
			ev.Check("prior default remains valid with its original snapshot", err == nil, "err=%v", err)
			current, err := os.ReadFile("../examples/chineseeval/testdata/" + tc.base + "-v3.json")
			if err != nil {
				t.Fatal(err)
			}
			_, err = chineseeval.LoadRecorded(data, current, recorded.Catalog)
			ev.Check("new composition config rejects the historical snapshot", err != nil, "new config accepted old composition")
			ev.Check("historical v1 bytes match Git b8080341", chineseeval.Hash(old) == tc.oldHash, "historical input changed")
			_, err = chineseeval.Load(data, old, ai.BuiltinCatalog())
			ev.Check("old root pin is never silently upgraded", err != nil, "old config accepted current catalog")
			plan, err := chineseeval.Load(data, current, ai.BuiltinCatalog())
			ev.Check("new current config loads with strict pins", err == nil, "err=%v", err)
			var before, after chineseeval.Config
			if json.Unmarshal(previous, &before) != nil || json.Unmarshal(current, &after) != nil {
				t.Fatal("invalid stored config")
			}
			ev.Check("config revision changes explicitly", before.Version != after.Version, "old=%s new=%s", before.Version, after.Version)
			before.Version, before.CatalogVersion, before.CatalogHash = after.Version, after.CatalogVersion, after.CatalogHash
			ev.Check("task model dataset and finite budget remain unchanged", reflect.DeepEqual(before, after), "unexpected config changes")
			if plan != nil {
				ev.Record("current-config", plan.Config())
			}
		})
	}
}
