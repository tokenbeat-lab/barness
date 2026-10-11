package e2e

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// Configuration seam (S18/S19/S25): inactive profiles must be validated on
// ignored native protocols too. Hash and NewClient accept exactly the same
// structure, without echoing the parameter value in an error.
func TestCatalogThinkingSampling(t *testing.T) {
	for _, proto := range []*fixtureProtocol{responsesProtocol, chatProtocol, deepseekChatProtocol, deepseekResponsesProtocol, anthropicProtocol, geminiProtocol} {
		for _, sc := range []struct {
			name, profile string
			valid         bool
		}{
			{"unknown", `{"none":{"temperature":0}}`, false},
			{"uppercase", `{"MINIMAL":{}}`, false},
			{"empty-key", `{"":{}}`, false},
			{"nil-profile", `{"max":null}`, false},
			{"empty-profile", `{"max":{}}`, true},
			{"null-value", `{"max":{"temperature":null}}`, true},
			{"empty-map", `{}`, true},
		} {
			t.Run(proto.dir+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, "E03-thinking-sampling-config-"+proto.dir+"-"+sc.name)
				model := ai.BuiltinCatalog().ModelsOf(proto.provider, proto.api)[0]
				model = patchModel(model, json.RawMessage(`{"samplingParamsByThinkingLevel":`+sc.profile+`}`))
				catalog := ai.Catalog{Version: "synthetic-thinking-sampling", Models: []ai.Model{model}}
				ev.Record("input", json.RawMessage(`{"samplingParamsByThinkingLevel":`+sc.profile+`}`))
				h := host.New()
				client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &catalog})
				hash, hashErr := catalog.Hash()
				ev.Record("outcome", map[string]any{"error": errString(err), "hash_error": errString(hashErr), "hash": hash})
				ev.Check("both boundaries agree", (err == nil) == sc.valid && (hashErr == nil) == sc.valid, "client=%v hash=%v", err, hashErr)
				if !sc.valid {
					ev.Check("safe configuration location", errors.Is(err, ai.ErrInvalidConfig) && strings.Contains(errString(err), "samplingParamsByThinkingLevel"), "err=%v", err)
					ev.Check("same error", errString(err) == errString(hashErr), "client=%v hash=%v", err, hashErr)
				} else {
					ev.Check("client exists", client != nil, "err=%v", err)
				}
			})
		}
	}
}
