package e2e

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// The public catalog boundary owns default labels; these are sampling
// metadata, not a user request to turn on reasoning.
func TestMiniMaxDefaultThinkingMetadata(t *testing.T) {
	for _, sc := range []struct {
		name, patch string
		valid       bool
	}{
		{"zero", `{}`, true},
		{"canonical", `{"defaultThinkingLevel":"high"}`, true},
		{"unknown", `{"defaultThinkingLevel":"none"}`, false},
		{"unsupported", `{"defaultThinkingLevel":"high","thinkingLevelMap":{"high":null}}`, false},
		{"mandatory-off", `{"defaultThinkingLevel":"off","compat":{"requiresThinking":true}}`, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM01-default-metadata-"+sc.name)
			base, _ := ai.BuiltinCatalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-5")
			model := patchModel(base, json.RawMessage(sc.patch))
			catalog := ai.Catalog{Version: "synthetic-default-label", Models: []ai.Model{model}}
			h := host.New()
			client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &catalog})
			hash, hashErr := catalog.Hash()
			ev.Record("catalog", catalog)
			ev.Record("outcome", map[string]any{"error": errString(err), "hash_error": errString(hashErr), "hash": hash})
			ev.Check("configuration boundaries agree", (err == nil) == sc.valid && (hashErr == nil) == sc.valid, "client=%v hash=%v", err, hashErr)
			if !sc.valid {
				ev.Check("safe default-label error", errors.Is(err, ai.ErrInvalidConfig) && strings.Contains(errString(err), "defaultThinkingLevel"), "%v", err)
				return
			}
			if client == nil {
				return
			}
			snapshot := client.Catalog()
			before := mustMarshal(t, snapshot)
			catalog.Models[0] = base
			ev.Check("construction owns label", jsonEqual(before, mustMarshal(t, client.Catalog())), "label changed")
			fields := bodyFields(t, mustMarshal(t, snapshot.Models[0]))
			if sc.name == "canonical" {
				ev.Check("JSON retains canonical label", string(fields["defaultThinkingLevel"]) == `"high"`, "%s", before)
				ev.Check("label enters price snapshot hash", catalogHash(t, catalog) != hash, "hash did not change")
			} else {
				_, exists := fields["defaultThinkingLevel"]
				ev.Check("zero remains omitted", !exists, "%s", before)
			}
		})
	}
}
