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

// Current route defaults follow today's stored release snapshot. Historical
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
			raw, err = os.ReadFile("../release/catalog-snapshot.json")
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
	for _, tc := range []struct{ base, oldHash string }{
		{"jev-1.13.0", "sha256:001556ea3be2a33e5b9f06a5530733fc9088a8942d073b51b8e07d33f1665ed3"},
		{"jev-route", "sha256:b4501a5274b2092c09169395d5a870b6049f809f34c42510bdcabdf960b9c297"},
		{"luna-route", "sha256:7679d0bcdda0698900623353eafe3203541e1bc15725a1c362ed66407f7b4af7"},
	} {
		t.Run(tc.base, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-config-revision-"+tc.base)
			old, err := os.ReadFile("../examples/chineseeval/testdata/" + tc.base + "-v1.json")
			if err != nil {
				t.Fatal(err)
			}
			current, err := os.ReadFile("../examples/chineseeval/testdata/" + tc.base + "-v2.json")
			if err != nil {
				t.Fatal(err)
			}
			ev.Check("historical v1 bytes match Git b8080341", chineseeval.Hash(old) == tc.oldHash, "historical input changed")
			_, err = chineseeval.Load(data, old, ai.BuiltinCatalog())
			ev.Check("old root pin is never silently upgraded", err != nil, "old config accepted current catalog")
			plan, err := chineseeval.Load(data, current, ai.BuiltinCatalog())
			ev.Check("new current config loads with strict pins", err == nil, "err=%v", err)
			var before, after chineseeval.Config
			if json.Unmarshal(old, &before) != nil || json.Unmarshal(current, &after) != nil {
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
