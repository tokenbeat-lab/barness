package e2e

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
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
// catalog are listed, and are served the way pi serves them with that
// feature off. Each protocol's catalog.json runs its scenarios — no tools,
// tools, mid-conversation system and tool changes, level maps and tiered
// prices — through the public Client; TestPiDifferential runs the same
// scenarios through frozen pi, where pi's optional features are ledger
// extensions limited to these cases.
//
// Ways it could fail:
//
//	L1 a listed model is missing, or listed on a protocol that cannot
//	   serve it (gpt-5.4-pro on Chat Completions)
//	L2 (managed mid-conversation effort, the hard constraint ADR-0018 kept
//	   claude-fable-5-1, claude-opus-5 and claude-opus-5-5 out for, is
//	   TestManagedEffort's)
//	L3 a field differs from pi's frozen data: name, limits, level map,
//	   carried compat or prices (checked against pi with BARNESS_AI_PIDIFF=1)
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
		for _, id := range []string{"claude-opus-4-8", "claude-fable-5"} {
			ev.Check(id+" on Messages", slices.Contains(anthropic, id), "listed %q", anthropic)
		}
		responses := builtinModelIDs(ai.APIOpenAIResponses, ai.ProviderOpenAI)
		chat := builtinModelIDs(ai.APIOpenAICompletions, ai.ProviderOpenAI)
		for _, id := range []string{"gpt-5.4", "gpt-5.4-mini", "gpt-5.4-pro", "gpt-5.5", "gpt-5.6-luna", "gpt-5.6-sol",
			"gpt-5.6-terra", "gpt-6-astra", "gpt-6-luna", "gpt-6-sol"} {
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

	t.Run("builtin-models-are-pi's", func(t *testing.T) {
		ev := run.Case(t, "E03-catalog-builtin-models-are-pis")
		ev.ReplayEnv(pioracle.EnableEnv + "=1")
		o := openOracle(t)
		catalog := ai.BuiltinCatalog()
		keys := make([]pioracle.ModelKey, len(catalog.Models))
		for i, m := range catalog.Models {
			keys[i] = piModelKey(m)
		}
		ctx, cancel := context.WithTimeout(ctxFor(t), 30*time.Second)
		defer cancel()
		pi, err := o.Models(ctx, keys)
		if !ev.Check("pi reports its model data", err == nil && len(pi) == len(keys), "err=%v", err) {
			return
		}
		ev.Record("pi-models", pi)
		for i, m := range catalog.Models {
			var want ai.Model
			if !ev.Check(m.ID+" ("+string(m.API)+") is in pi's data", string(pi[i]) != "null" && json.Unmarshal(pi[i], &want) == nil,
				"key %+v", keys[i]) {
				continue
			}
			// The catalog lists pi's entry on the API barness routes it
			// on; compat keeps the flags ModelCompat carries, which
			// decoding pi's entry into Model already does.
			want.Provider, want.API = m.Provider, m.API
			ev.Check(m.ID+" ("+string(m.API)+") fields are pi's", reflect.DeepEqual(m, want),
				"barness %s\npi      %s", mustMarshal(t, m), mustMarshal(t, want))
		}
	})
}

// piModelKey is where pi's data holds m: OpenAI's Chat models are its
// Responses entries (ADR-0013) and DeepSeek's Responses model is its Chat
// entry (ADR-0014), which the oracle routes on those APIs.
func piModelKey(m ai.Model) pioracle.ModelKey {
	api := m.API
	if m.Provider == ai.ProviderDeepSeek && m.API == ai.APIOpenAIResponses {
		api = ai.APIOpenAICompletions
	}
	return pioracle.ModelKey{Provider: string(m.Provider), API: string(api), ID: m.ID}
}
