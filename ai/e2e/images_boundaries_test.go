package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestOpenAIImagesKnownCallbackSize(t *testing.T) {
	large := strings.Repeat("x", 8<<20)
	raw := json.RawMessage(`"` + large + `"`)
	for _, name := range []string{"string", "raw", "raw-pointer", "nested"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E08-callback-allocation-"+name)
			sc := firstImagesScenario(t)
			w := scenarioWorld(t, sc)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "string":
					p.Body["prompt"] = large
				case "raw":
					p.Body["prompt"] = raw
				case "raw-pointer":
					p.Body["prompt"] = &raw
				case "nested":
					p.Body["prompt"] = []any{large}
				}
				return ai.KeepPayload(), nil
			}}
			req := imagesInput(t, sc)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-image-allocation-"+name), sc.target(), req, nil)
			runtime.ReadMemStats(&after)
			allocated := after.TotalAlloc - before.TotalAlloc
			ev.Record("allocated-bytes", allocated)
			ev.Record("result", res)
			ev.Check("known input bounded before serialization", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && allocated < 4<<20 && len(w.provider.Requests()) == 0, "got %v allocated %d", err, allocated)
		})
	}
}

func TestOpenAIImagesCachePricingRefused(t *testing.T) {
	for _, name := range []string{"text", "image", "null"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E11-cache-price-"+name)
			w := newWorld(t, tenantA)
			cfg := ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host}
			configureImages(&cfg)
			switch name {
			case "text":
				cfg.Catalog.ImageModels[0].Pricing.CacheReadText = ai.Value(1.0)
			case "image":
				cfg.Catalog.ImageModels[0].Pricing.CacheReadImage = ai.Value(1.0)
			case "null":
				cfg.Catalog.ImageModels[0].Pricing.CacheReadText = ai.Null[float64]()
			}
			_, err := ai.NewClient(cfg)
			ev.Record("error", errString(err))
			ev.Check("direct Images has no cache pricing", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
		})
	}
}

func TestOpenAIImagesInvalidEntriesKeepUsage(t *testing.T) {
	for _, name := range []string{"numeric", "null", "wrong-data", "wrong-mime", "null-mime", "empty-mime", "duplicate-b64", "many-empty"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E02-entry-"+name)
			sc := firstImagesScenario(t)
			w := scenarioWorld(t, sc)
			var body map[string]json.RawMessage
			mustUnmarshal(t, []byte(sc.Replies[0].Body), &body)
			switch name {
			case "numeric":
				body["data"] = json.RawMessage(`[42]`)
			case "null":
				body["data"] = json.RawMessage(`[null]`)
			case "wrong-data":
				body["data"] = json.RawMessage(`{"nested":[]}`)
			case "wrong-mime":
				body["data"] = json.RawMessage(`[{"b64_json":"/9j/2Q==","mime_type":"image/png"}]`)
			case "null-mime", "empty-mime":
				var images []map[string]json.RawMessage
				mustUnmarshal(t, body["data"], &images)
				images[0]["mime_type"] = json.RawMessage(`null`)
				if name == "empty-mime" {
					images[0]["mime_type"] = json.RawMessage(`""`)
				}
				body["data"] = mustMarshal(t, images)
			case "duplicate-b64":
				body["data"] = json.RawMessage(`[{"b64_json":"/9j/2Q==","b64_json":"/9j/2Q=="}]`)
			case "many-empty":
				body["data"] = json.RawMessage(`[` + strings.Repeat(`{},`, 10000) + `{}]`)
			}
			enqueue(ev, w, fixtureReply{Status: 200, Body: string(mustMarshal(t, body))}.script(t, imagesProtocol, ""))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-entry-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			want := ai.CodeProtocol
			if name == "many-empty" {
				want = ai.CodeResourceLimit
			}
			ev.Check("usage survives invalid entries", errors.Is(err, &ai.Error{Code: want, Phase: ai.PhaseResponse}) && len(res.Content) == 0 && res.Usage.TotalTokens == 30 && res.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "got %v %+v", err, res)
		})
	}
}

func TestOpenAIImagesEntryOptionsSnapshot(t *testing.T) {
	ev := run.Case(t, "P08-E07-entry-options-snapshot")
	sc := firstImagesScenario(t)
	opts := ai.OpenAIImagesOptions{N: ai.Value(1)}
	w := newWorldWith(t, func(c *ai.Config) {
		configureImages(c)
		bindings := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
			// The resolver runs after entry; subsequent caller-owned mutations
			// must not alter the options already received by the call.
			opts.N = ai.Value(2)
			return bindings.ResolveBinding(ctx, scope, id)
		})
	}, tenantA)
	installImagesBinding(w, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.GenerateImages(ctxFor(t), textScope("req-entry-options-snapshot"), sc.target(), imagesInput(t, sc), &opts)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	ev.Record("error", errString(err))
	var request map[string]json.RawMessage
	if requests := w.provider.Requests(); len(requests) == 1 {
		mustUnmarshal(t, requests[0].Body, &request)
	}
	ev.Check("entry options independently frozen", err == nil && string(request["n"]) == "1" && opts.N == ai.Value(2), "got %v request=%s", err, request["n"])
}

func TestCatalogNativeImageExemptions(t *testing.T) {
	ev := run.Case(t, "P08-E11-native-catalog-exemption")
	ledger := loadLedger(t)
	ev.Record("image-models", ai.BuiltinCatalog().ImageModels)
	// Native image routes are never silently included in the chat-model pi
	// comparison. Future issue 11 additions must retain this explicit exemption.
	for _, m := range ai.BuiltinCatalog().ImageModels {
		r, ok := ledger.Route(string(m.Provider), string(m.API))
		ev.Check("native image exemption "+m.ID, ok && r.Ref != "" && m.Pricing.Source != "", "missing native evidence exemption")
	}
	r, ok := ledger.Route(string(ai.ProviderOpenAI), string(ai.APIOpenAIImages))
	ev.Check("explicit native Images exemption exists", ok && r.Ref != "", "missing exemption")
}
