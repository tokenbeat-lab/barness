package e2e

import (
	"encoding/json"

	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// responsesProtocol is OpenAI × Responses in the shared scenario format;
// only testdata/responses/catalog.json uses it, the older P01 fixtures
// keep their own shapes.
var responsesProtocol = &fixtureProtocol{
	dir: "responses", binding: "primary", api: ai.APIOpenAIResponses, provider: ai.ProviderOpenAI,
	key: func(k tenantKey) tenantKey { return k }, account: "acct-tenant-a",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.ResponsesOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: sseEvents, textMarker: `"type":"response.output_text.delta"`,
}

// catalogInclusionProtocols are the protocols whose catalog.json holds the
// scenarios of the models ADR-0018 lists, with their spec prefix.
var catalogInclusionProtocols = []struct {
	spec  string
	proto *fixtureProtocol
}{{"P01", responsesProtocol}, {"P02", anthropicProtocol}, {"P04", chatProtocol}}

// TestCatalogInclusion is issue 34 (ADR-0018): the models that only an
// optional protocol feature barness-ai does not implement kept out of the
// catalog are listed and served with explicit supported options. Protocol
// catalog fixtures exercise tools, history, level maps and tiered prices
// through the public Client with independent literal expectations.
//
// Ways it could fail:
//
//	L1 a listed model is missing, or listed on a protocol that cannot
//	   serve it (gpt-5.4-pro on Chat Completions)
//	L2 (managed mid-conversation effort, the hard constraint ADR-0018 kept
//	   claude-fable-5-1, claude-opus-5 and claude-opus-5-5 out for, is
//	   TestManagedEffort's)
//	L3 a field differs from pi's frozen data: name, limits, level map,
//	   carried capabilities or prices (checked against independent fixtures)
//	L4 a hard constraint is not kept: a temperature reaches
//	   claude-opus-4-8, thinking is switched off on a model whose level
//	   map has off null, a level outside the map is sent unclamped
//	L5 a later system message folds into the leading one, or the request
//	   declares other than the tools current after the last change
//	L6 an optional feature leaks in: a placeholder tool, deferred tools,
//	   tool_addition/tool_removal blocks, fallbacks, a beta header,
//	   additional_tools or tool_search items
func TestCatalogInclusion(t *testing.T) {
	t.Run("listed", func(t *testing.T) {
		ev := run.Case(t, "E03-catalog-inclusion-listed")
		ev.Record("catalog", ai.BuiltinCatalog())
		anthropic := builtinModelIDs(ai.APIAnthropicMessages, ai.ProviderAnthropic)
		for _, id := range []string{"claude-opus-4-8", "claude-fable-5", "claude-sonnet-5-5"} {
			ev.Check(id+" on Messages", slices.Contains(anthropic, id), "listed %q", anthropic)
		}
		responses := builtinModelIDs(ai.APIOpenAIResponses, ai.ProviderOpenAI)
		chat := builtinModelIDs(ai.APIOpenAICompletions, ai.ProviderOpenAI)
		for _, id := range []string{"gpt-5.4", "gpt-5.4-mini", "gpt-5.4-pro", "gpt-5.5", "gpt-5.6-luna", "gpt-5.6-sol",
			"gpt-5.6-terra", "gpt-6-astra", "gpt-6-luna", "gpt-6-sol", "gpt-6.1-sol", "gpt-daybreak-blue-latest", "gpt-daybreak-red-latest"} {
			ev.Check(id+" on Responses", slices.Contains(responses, id), "listed %q", responses)
			pro := id == "gpt-5.4-pro"
			ev.Check(id+" on Chat unless a pro model", slices.Contains(chat, id) != pro, "listed %q", chat)
		}
		for _, id := range responses {
			if !slices.Contains(chat, id) {
				// Only pro models answer on Responses alone (ADR-0013 决策二).
				ev.Check(id+" left off Chat is a pro model", strings.HasSuffix(id, "-pro"), "chat %q", chat)
			}
		}
	})

	for _, p := range catalogInclusionProtocols {
		f, raw := loadFixture(t, p.proto, "catalog.json")
		for _, sc := range f.Scenarios {
			t.Run(p.spec+"/"+sc.ID, func(t *testing.T) {
				ev := run.Case(t, p.spec+"-E03-catalog-"+sc.ID)
				runScenario(t, ev, sc, raw, modeStream)
			})
		}
	}

}
