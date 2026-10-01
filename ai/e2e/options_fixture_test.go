package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// derivedCacheKey stands, in options.json, for the cache key barness-ai
// derives from tenant, account and session; the scenario checks the observed
// key's properties and then compares the rest of the body with it in place.
const derivedCacheKey = "$derivedCacheKey"

// optionsFixture is testdata/responses/options.json.
type optionsFixture struct {
	Events  []json.RawMessage `json:"events"`
	Context struct {
		SystemPrompt string            `json:"systemPrompt"`
		Tools        []ai.Tool         `json:"tools"`
		Messages     []json.RawMessage `json:"messages"`
	} `json:"context"`
	Scenarios []optionsScenario `json:"scenarios"`
}

type optionsScenario struct {
	ID    string `json:"id"`
	About string `json:"about"`
	// Entry is the pi entry the options belong to: "stream" (full options)
	// or "streamSimple" (simple options).
	Entry string `json:"entry"`
	Model string `json:"model"`
	// ModelPatch replaces top-level fields of the catalog model and
	// ModelCompat its compat flags, on both sides, as a custom model does.
	ModelPatch  json.RawMessage `json:"modelPatch"`
	ModelCompat *ai.ModelCompat `json:"modelCompat"`
	// Options are in pi-ai's option shape, which is also barness-ai's.
	Options json.RawMessage `json:"options"`
	// Messages replace the shared context's messages when set.
	Messages []json.RawMessage `json:"messages"`
	Expect   struct {
		Body json.RawMessage `json:"body"`
		// Affinity: the session affinity headers carry the derived key.
		Affinity bool `json:"affinity"`
	} `json:"expect"`
}

func loadOptionsFixture(t *testing.T) (optionsFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "options.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f optionsFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("options.json: %v", err)
	}
	return f, raw
}

func (f optionsFixture) reply(t *testing.T) provider.Reply {
	return sseEvents(t, f.Events, provider.FramingLF)
}

func (f optionsFixture) request(t *testing.T, sc optionsScenario) ai.Request {
	t.Helper()
	req := ai.Request{SystemPrompt: f.Context.SystemPrompt, Tools: f.Context.Tools}
	messages := f.Context.Messages
	if sc.Messages != nil {
		messages = sc.Messages
	}
	for _, raw := range messages {
		req.Messages = append(req.Messages, decodeStoredMessage(t, raw, tenantA))
	}
	return req
}

func (sc optionsScenario) simple() bool { return sc.Entry == "streamSimple" }

func (sc optionsScenario) configure(cfg *ai.Config) {
	withModelPatch(sc.Model, sc.ModelCompat, sc.ModelPatch)(cfg)
}

// fullOptions decodes the scenario's options as the full entry's.
func (sc optionsScenario) fullOptions(t *testing.T) ai.ResponsesOptions {
	t.Helper()
	var o ai.ResponsesOptions
	mustUnmarshal(t, sc.Options, &o)
	return o
}

// simpleOptions decodes the scenario's options as the simple entry's.
func (sc optionsScenario) simpleOptions(t *testing.T) ai.SimpleOptions {
	t.Helper()
	var o ai.SimpleOptions
	mustUnmarshal(t, sc.Options, &o)
	return o
}

// optionsEntry is one public entry point an options scenario runs through:
// Stream and Complete for full options, StreamSimple and CompleteSimple for
// simple options.
type optionsEntry struct {
	name   string
	stream bool
}

var optionsEntries = []optionsEntry{{"stream", true}, {"complete", false}}

// invoke runs the scenario's call. It returns the outcome and the options
// value, marshaled before and after the call, so a mutation shows up.
func (e optionsEntry) invoke(ctx context.Context, t *testing.T, w *world, scope ai.CallScope, sc optionsScenario, req ai.Request) (outcome, string, string) {
	target := ai.Target{BindingID: "primary", ModelID: sc.Model}
	if sc.simple() {
		opts := sc.simpleOptions(t)
		before := string(mustMarshal(t, opts))
		var o outcome
		if e.stream {
			o = drain(w.client.StreamSimple(ctx, scope, target, req, opts))
		} else {
			res, err := w.client.CompleteSimple(ctx, scope, target, req, opts)
			o = outcome{result: res, err: err}
		}
		return o, before, string(mustMarshal(t, opts))
	}
	opts := sc.fullOptions(t)
	before := string(mustMarshal(t, opts))
	var o outcome
	if e.stream {
		o = drain(w.client.Stream(ctx, scope, target, req, opts))
	} else {
		res, err := w.client.Complete(ctx, scope, target, req, opts)
		o = outcome{result: res, err: err}
	}
	return o, before, string(mustMarshal(t, opts))
}

// withModelPatch gives model the patched fields and compat flags, as a host's
// own catalog would. Like pi's object spread, each top-level key of patch
// replaces the model's field entirely.
func withModelPatch(model string, compat *ai.ModelCompat, patch json.RawMessage) func(*ai.Config) {
	return func(cfg *ai.Config) {
		if compat == nil && patch == nil {
			return
		}
		catalog := ai.BuiltinCatalog()
		for i, m := range catalog.Models {
			if m.ID != model {
				continue
			}
			if patch != nil {
				catalog.Models[i] = patchModel(m, patch)
			}
			if compat != nil {
				catalog.Models[i].Compat = *compat
			}
		}
		cfg.Catalog = &catalog
	}
}

func patchModel(m ai.Model, patch json.RawMessage) ai.Model {
	var fields, changes map[string]json.RawMessage
	raw, err := json.Marshal(m)
	if err == nil {
		err = json.Unmarshal(raw, &fields)
	}
	if err == nil {
		err = json.Unmarshal(patch, &changes)
	}
	for k, v := range changes {
		fields[k] = v
	}
	if err == nil {
		raw, err = json.Marshal(fields)
	}
	var out ai.Model
	if err == nil {
		err = json.Unmarshal(raw, &out)
	}
	if err != nil {
		panic("patchModel: " + err.Error())
	}
	return out
}

// substituteCacheKey puts key in place of the derived key marker.
func substituteCacheKey(body json.RawMessage, key string) []byte {
	return []byte(strings.ReplaceAll(string(body), `"`+derivedCacheKey+`"`, mustQuote(key)))
}

func mustQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// sentCacheKey is the prompt_cache_key of a request body, "" when absent.
func sentCacheKey(t *testing.T, body []byte) string {
	t.Helper()
	var b struct {
		Key string `json:"prompt_cache_key"`
	}
	mustUnmarshal(t, body, &b)
	return b.Key
}

// pidiff is the scenario of TestOptions for the differential.
func (f optionsFixture) pidiff(t *testing.T, raw []byte, sc optionsScenario) pidiffScenario {
	t.Helper()
	p := pidiffScenario{fixture: "options.json", raw: raw, model: sc.Model, req: f.request(t, sc), reply: f.reply(t),
		requests: 1, modelCompat: sc.ModelCompat, modelPatch: sc.ModelPatch, piOptionsRaw: sc.Options}
	if sc.simple() {
		p.simple = sc.simpleOptions(t)
	} else {
		p.full = sc.fullOptions(t)
	}
	return p
}
