package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// S19/S20/S25: every canonical profile on every consumer rejects each reserved
// key, even though no call selects it. Native APIs validate JSON only. Invalid
// bytes are recorded as bytes (base64), never passed to JSON evidence encoding.
func TestThinkingSamplingValidation(t *testing.T) {
	levels := []ai.ThinkingLevel{"off", ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax}
	for _, proto := range []*fixtureProtocol{responsesProtocol, chatProtocol, deepseekResponsesProtocol, deepseekChatProtocol, anthropicProtocol, geminiProtocol} {
		keys := []string{"model", "input", "stream", "tools", "previous_response_id", "conversation", "prompt", "prompt_cache_key", "store", "background"}
		consumer := proto.api == ai.APIOpenAIResponses || proto.api == ai.APIOpenAICompletions
		if proto.api == ai.APIOpenAICompletions {
			keys = []string{"model", "messages", "stream", "tools", "functions", "prompt_cache_key", "store", "web_search_options"}
		}
		for _, level := range levels {
			for _, key := range append(keys, "invalid-json", "empty-json") {
				name := proto.dir + "-" + string(level) + "-" + key
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "E03-thinking-sampling-validation-"+name)
					value := json.RawMessage(`"synthetic-private-parameter"`)
					if key == "invalid-json" {
						value = json.RawMessage(`{"broken":`)
					}
					if key == "empty-json" {
						value = nil
					}
					m := ai.BuiltinCatalog().ModelsOf(proto.provider, proto.api)[0]
					m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{level: {key: value}}
					catalog := ai.Catalog{Version: "synthetic", Models: []ai.Model{m}}
					ev.Record("input", map[string]any{"model_id": m.ID, "level": level, "key": key, "raw": []byte(value)})
					h := host.New()
					_, err := ai.NewClient(ai.Config{Catalog: &catalog, Policy: validPolicy(), Bindings: h, Credentials: h})
					_, hashErr := catalog.Hash()
					invalid := consumer || key == "invalid-json" || key == "empty-json"
					ev.Check("construct/hash share validation", (err != nil) == invalid && (hashErr != nil) == invalid, "client=%v hash=%v", err, hashErr)
					if invalid {
						ev.Check("same error and canonical location", errString(err) == errString(hashErr) && strings.Contains(errString(err), "samplingParamsByThinkingLevel."+string(level)), "client=%v hash=%v", err, hashErr)
						ev.Check("error omits parameter value", !strings.Contains(errString(err), "synthetic-private-parameter"), "value echoed")
					}
					ev.Record("outcome", map[string]any{"client_error": errString(err), "hash_error": errString(hashErr)})
				})
			}
		}
	}
	// Non-object profile values fail while decoding the public metadata type.
	for _, raw := range []string{`{"high":1}`, `{"high":[]}`, `{"high":"value"}`} {
		ev := run.Case(t, "E03-thinking-sampling-non-object-"+string(raw[8]))
		var model ai.Model
		err := json.Unmarshal([]byte(`{"samplingParamsByThinkingLevel":`+raw+`}`), &model)
		ev.Check("profile must decode as object", err != nil, "err=%v", err)
	}
}

func TestNativeThinkingSamplingIgnored(t *testing.T) {
	for _, proto := range []*fixtureProtocol{anthropicProtocol, geminiProtocol} {
		f, raw := loadFixture(t, proto, "text.json")
		for _, sc := range f.Scenarios {
			if sc.ID != "text" && sc.ID != "text-simple" {
				continue
			}
			sc.ModelPatch = json.RawMessage(`{"samplingParams":{"model":"unused"},"samplingParamsByThinkingLevel":{"off":{"model":"unused"},"high":{"temperature":false}}}`)
			for _, mode := range []callMode{modeStream, modeComplete} {
				t.Run(proto.dir+"/"+sc.ID+"/"+string(mode), func(t *testing.T) {
					ev := run.Case(t, "E04-thinking-sampling-native-"+proto.dir+"-"+sc.ID+"-"+string(mode))
					ev.Record("profile-input", sc.ModelPatch)
					runScenario(t, ev, sc, raw, mode)
				})
			}
		}
	}
}
