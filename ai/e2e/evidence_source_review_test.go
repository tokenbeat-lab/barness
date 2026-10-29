package e2e

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Historical positive cases bind the last pre-Imagine product inventory from
// Git commit 1e244ee26f024ce3534eed0600c43f6317d77adb, not this checkout and
// not the consumer's allowlist. This prevents real new product changes from
// rewriting the meaning of a previously reviewed source transition.
// Failure inventory: exact reviewed transitions may pass; changed old bytes,
// later edits, additions, removals, dependency changes and another route must
// still fail. These are hash-only original source inventories, not live grants.
type reviewedHistoricalInventory struct {
	SourceCommit string            `json:"sourceCommit"`
	Files        map[string]string `json:"files"`
}

func reviewedHistoricalSource(t *testing.T) reviewedHistoricalInventory {
	t.Helper()
	var historical reviewedHistoricalInventory
	frozen, err := os.ReadFile("testdata/source-review/current-1e244ee.json")
	if err != nil || json.Unmarshal(frozen, &historical) != nil || historical.SourceCommit != "1e244ee26f024ce3534eed0600c43f6317d77adb" || len(historical.Files) == 0 {
		t.Fatal("frozen historical current source unavailable", err)
	}
	return historical
}

func TestEvidenceReviewedSourceTransition(t *testing.T) {
	historical := reviewedHistoricalSource(t)
	for _, route := range []string{"typesafe-classifier-jev-1.14.0", "zhipu-responses", "zai-responses"} {
		modes := []string{"reviewed", "current-unreviewed", "changed-origin", "later-edit", "added-runtime", "added-subpackage", "removed-runtime", "dependency", "wrong-provider", "wrong-api", "later-clock"}
		if route == "zhipu-responses" {
			modes = append(modes, "changed-inventory")
		}
		for _, mode := range modes {
			t.Run(route+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "SOURCE55-"+route+"-"+mode)
				var fixture struct {
					Provider, API string
					Measured      map[string]string
				}
				raw, err := os.ReadFile(filepath.Join("testdata/source-review", route+".json"))
				if err != nil || json.Unmarshal(raw, &fixture) != nil {
					t.Fatal("original source fixture unavailable", err)
				}
				current := maps.Clone(historical.Files)
				if mode == "current-unreviewed" {
					current, err = supportmatrix.RegionalExecutionSource()
					if err != nil {
						t.Fatal(err)
					}
					// Capturing helpers are independently checked by the native consumer.
					for name := range current {
						if strings.HasPrefix(name, "ai/internal/testkit/") || strings.HasPrefix(name, "ai/live/") {
							delete(current, name)
						}
					}
				}
				switch mode {
				case "changed-origin":
					fixture.Measured["ai/history.go"] = strings.Repeat("1", 64)
				case "later-edit":
					current["ai/history.go"] = strings.Repeat("2", 64)
				case "later-clock":
					current["ai/internal/clock/clock.go"] = strings.Repeat("2", 64)
				case "changed-inventory":
					fixture.Measured["ai/live/glm_harness_test.go"] = strings.Repeat("2", 64)
				case "added-runtime":
					current["ai/unreviewed.go"] = strings.Repeat("3", 64)
				case "added-subpackage":
					current["ai/unreviewed/runtime.go"] = strings.Repeat("3", 64)
				case "removed-runtime":
					delete(current, "ai/history.go")
				case "dependency":
					current["go.mod"] = strings.Repeat("4", 64)
				case "wrong-provider":
					fixture.Provider = "deepseek"
				case "wrong-api":
					fixture.API = "openai-completions"
				}
				err = evidence.VerifyRouteSource(fixture.Measured, current, fixture.Provider, fixture.API)
				ev.Check("only exact reviewed product delta is reusable", (err == nil) == (mode == "reviewed"), "%v", err)
				ev.Record("source-review", map[string]any{"route": route, "mode": mode, "error": errString(err), "vendorCalls": 0, "historicalCurrentCommit": historical.SourceCommit})
			})
		}
	}
}

func TestEvidenceProtocolEntrypointScope(t *testing.T) {
	for _, file := range []string{"chat.go", "anthropic.go", "responses.go"} {
		for _, api := range []string{"typesafe-system-one", "openai-responses", "openai-completions", "anthropic-messages"} {
			t.Run(file+"/"+api, func(t *testing.T) {
				ev := run.Case(t, "SOURCE55-entry-"+file+"-"+api)
				own := map[string]string{"chat.go": "openai-completions", "anthropic.go": "anthropic-messages", "responses.go": "openai-responses"}[file]
				name := "ai/" + file
				err := evidence.VerifyRouteSource(map[string]string{name: strings.Repeat("a", 64)}, map[string]string{name: strings.Repeat("b", 64)}, "zhipu", api)
				ev.Check("entrypoint belongs to its actual protocol", (err == nil) == (api != own), "%v", err)
				ev.Record("ownership", map[string]any{"file": file, "api": api, "applicable": err == nil})
			})
		}
	}
}

func TestEvidenceVisualExecutionScope(t *testing.T) {
	for _, mode := range []string{"reviewed-budget-hook", "later-host-edit", "scorer-edit", "new-generator", "consumer-edit"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "SOURCE55-visual-"+mode)
			old := map[string]string{"ai/examples/visualeval/host.go": "2bb077fac84996cc38134f4eb33718573c18ae9350938346a630c10c1aee0e92"}
			current := map[string]string{"ai/examples/visualeval/host.go": "21202f7857524fcb9317ca6119dbadc53d00016be16e8632d8fb0e3ac14b671a"}
			switch mode {
			case "later-host-edit":
				current["ai/examples/visualeval/host.go"] = strings.Repeat("c", 64)
			case "scorer-edit":
				current["ai/examples/visualeval/score.go"] = strings.Repeat("d", 64)
			case "new-generator":
				current["ai/examples/visualeval/new_generator.go"] = strings.Repeat("e", 64)
			case "consumer-edit":
				current["ai/examples/visualeval/verify.go"] = strings.Repeat("f", 64)
			}
			err := evidence.VerifyVisualSource(old, current, "typesafe", "typesafe-system-one")
			want := mode == "reviewed-budget-hook" || mode == "consumer-edit"
			ev.Check("measurement execution remains protected", (err == nil) == want, "%v", err)
			ev.Record("source-review", map[string]any{"mode": mode, "error": errString(err), "vendorCalls": 0})
		})
	}
}

func TestEvidenceSourceCollectorsIncludeProductSubpackages(t *testing.T) {
	for name, collect := range map[string]func() (map[string]string, error){"glm": supportmatrix.GLMExecutionSource, "regional": supportmatrix.RegionalExecutionSource} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "SOURCE55-collector-"+name)
			root := t.TempDir()
			for path, raw := range map[string]string{"go.mod": "module fixture\n", "go.sum": "", "ai/newproduct/runtime.go": "package newproduct\n"} {
				full := filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Chdir(root)
			files, err := collect()
			ev.Check("unknown product subpackage reaches applicability guard", err == nil && files["ai/newproduct/runtime.go"] != "", "%v", err)
			ev.Record("collected", files)
		})
	}
}
