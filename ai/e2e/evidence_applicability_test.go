package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure inventory precedes implementation in verification-simplification/spec.md.
// These isolated evidence contracts also feed the compiled release CLI.
func TestEvidenceRouteSourceApplicability(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		want       bool
	}{
		{"consumer", "ai/internal/testkit/supportmatrix/xai.go", true},
		{"ordinary-test", "ai/e2e/xai_stream_test.go", true},
		{"runner", "tools/verification/run_process.py", true},
		{"other-protocol", "ai/chat_stream.go", true},
		{"other-provider", "ai/responses_xai_history.go", true},
		{"own-protocol", "ai/responses_request.go", false},
		{"shared-runtime", "ai/call.go", false},
		{"unknown-runtime", "ai/new_runtime.go", false},
		{"unknown-subpackage", "ai/transport/runtime.go", false},
		{"dependency", "go.mod", false},
		{"catalog-runtime", "ai/catalog_sampling.go", false},
		{"unknown-catalog-runtime", "ai/catalog_new_runtime.go", false},
		{"catalog-data", "ai/catalog_builtin.go", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "VS55-source-"+tc.name)
			old := map[string]string{tc.file: strings.Repeat("a", 64)}
			current := map[string]string{tc.file: strings.Repeat("b", 64)}
			err := evidence.VerifyRouteSource(old, current, "deepseek", "openai-responses")
			ev.Check("only product changes of this route invalidate applicability", (err == nil) == tc.want, "%v", err)
			ev.Record("source-change", map[string]any{"file": tc.file, "applicable": err == nil})
		})
	}
}

func TestEvidenceRouteCatalogApplicability(t *testing.T) {
	for _, mode := range []string{"unrelated-model", "display-name", "own-capability", "missing-model"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "VS55-catalog-"+mode)
			old, current := ai.BuiltinCatalog(), ai.BuiltinCatalog()
			index := -1
			for i, m := range current.Models {
				if m.Provider == ai.ProviderDeepSeek && m.API == ai.APIOpenAIResponses {
					index = i
				}
			}
			if index < 0 {
				t.Fatal("fixture identity missing")
			}
			id := current.Models[index].ID
			switch mode {
			case "unrelated-model":
				current.Models = append(current.Models, ai.Model{Provider: "unrelated", API: ai.APIOpenAIResponses, ID: "new"})
			case "display-name":
				current.Models[index].Name = "Display only"
			case "own-capability":
				current.Models[index].MaxTokens--
			case "missing-model":
				current.Models = append(current.Models[:index], current.Models[index+1:]...)
			}
			err := evidence.VerifyRouteCatalog(old, current, ai.OperationChat, ai.ProviderDeepSeek, ai.APIOpenAIResponses, []string{id})
			want := mode == "unrelated-model" || mode == "display-name"
			ev.Check("catalog applicability uses exact route identities", (err == nil) == want, "%v", err)
			ev.Record("comparison", map[string]any{"mode": mode, "applicable": err == nil})
		})
	}
}

func TestEvidenceProviderProtocolScope(t *testing.T) {
	for _, tc := range []struct {
		name, file, provider, api string
		want                      bool
	}{
		{"minimax-other-api", "ai/responses_minimax_history.go", "minimax", "openai-completions", true},
		{"minimax-own-api", "ai/responses_minimax_history.go", "minimax-cn", "openai-responses", false},
		{"kimi-other-api", "ai/anthropic_kimi_history.go", "moonshot-cn", "openai-responses", true},
		{"kimi-own-api", "ai/anthropic_kimi_history.go", "moonshot", "anthropic-messages", false},
		{"xai-guard-own", "ai/catalog_xai.go", "xai", "openai-responses", false},
		{"xai-guard-other", "ai/catalog_xai.go", "deepseek", "openai-responses", true},
		{"minimax-data-loader-own", "ai/catalog_minimax.go", "minimax", "openai-responses", false},
		{"minimax-data-loader-other", "ai/catalog_minimax.go", "xai", "openai-responses", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "VS55-provider-"+tc.name)
			err := evidence.VerifyRouteSource(map[string]string{tc.file: strings.Repeat("a", 64)}, map[string]string{tc.file: strings.Repeat("b", 64)}, tc.provider, tc.api)
			ev.Check("runtime ownership includes provider and protocol", (err == nil) == tc.want, "%v", err)
			ev.Record("scope", map[string]any{"file": tc.file, "provider": tc.provider, "api": tc.api, "applicable": err == nil})
		})
	}
}

func TestEvidenceDecisionsCatalogScope(t *testing.T) {
	for _, mode := range []string{"unrelated-model", "own-capability"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "VS55-decisions-"+mode)
			catalog := ai.BuiltinCatalog()
			catalog.Version = "TEST_ONLY_measured-catalog"
			if mode == "unrelated-model" {
				other := catalog.Models[0]
				other.ID += "-historical-fixture"
				catalog.Models = append(catalog.Models, other)
			} else {
				for i := range catalog.ClassifierModels {
					if catalog.ClassifierModels[i].API == ai.APIOpenAIDecisions {
						catalog.ClassifierModels[i].ContextWindow--
					}
				}
			}
			hash, err := catalog.Hash()
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			raw := formalDecisionsFixture(t)
			raw["catalogVersion"], raw["catalogHash"] = catalog.Version, hash
			writeReleaseJSON(t, dir, "model-catalog.json", catalog)
			writeReleaseJSON(t, dir, "live-report.json", raw)
			r, err := supportmatrix.LoadReport(filepath.Join(dir, "live-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			matrix := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{{Combo: r.Combo, Operation: r.Operation, Provider: r.Provider, API: r.API}}}
			_, err = supportmatrix.Merge(matrix, r)
			ev.Check("Decisions keeps only its measured model dependency", (err == nil) == (mode == "unrelated-model"), "%v", err)
			ev.Record("synthetic-measurement", map[string]any{"mode": mode, "vendorCalls": 0, "error": errString(err)})
		})
	}
}
