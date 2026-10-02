package e2e

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestDeepSeekResponsesRouting is P05's routing (E10) and authorization
// (E07): OpenAI and DeepSeek share the Responses adapter, yet a call
// reaches exactly its binding's provider, endpoint, key, model catalog and
// capabilities, and DeepSeek Responses is registered as an extension of
// the frozen pi routes rather than a differential pass.
func TestDeepSeekResponsesRouting(t *testing.T) {
	t.Run("shared-adapter-independent-configuration", func(t *testing.T) {
		ev := run.Case(t, "P05-E10-shared-adapter-independent-configuration")
		of, oraw := loadTextFixture(t, "text-basic.json")
		df, draw := loadFixture(t, deepseekResponsesProtocol, "text.json")
		ds := scenarioByID(t, df, "text")
		ev.Fixture("text-basic.json", oraw)
		ev.Fixture("text.json", draw)
		w := newWorld(t, tenantA)
		w.provider.EnqueueAt("/v1/", sseReply(t, of, provider.FramingLF))
		w.provider.EnqueueAt("/responses", ds.replies(t)...)

		// The same session and retention on both calls, run concurrently on
		// one Client: only OpenAI serves a prompt cache.
		opts := ai.ResponsesOptions{SessionID: "session-shared", CacheRetention: ai.CacheRetentionLong}
		calls := []struct {
			name   string
			target ai.Target
			req    ai.Request
		}{
			{"openai", textTarget(of), textRequest(of)},
			{"deepseek", ds.target(), ds.request(t)},
		}
		type routed struct {
			envs []ai.EventEnvelope
			res  ai.Result
			err  error
		}
		results := make([]routed, len(calls))
		var wg sync.WaitGroup
		for i, c := range calls {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s := w.client.Stream(ctxFor(t), textScope("req-route-"+c.name), c.target, c.req, opts)
				defer s.Close()
				for s.Next() {
					results[i].envs = append(results[i].envs, s.Envelope())
				}
				results[i].res, results[i].err = s.Result()
			}()
		}
		wg.Wait()
		for i, r := range results {
			ev.Record("result-"+calls[i].name, r.res)
			ev.Record("envelopes-"+calls[i].name, envelopesJSON(r.envs))
		}

		want := []struct {
			provider ai.ProviderID
			model    string
			binding  string
			account  string
		}{
			{ai.ProviderOpenAI, of.Model, "primary", "acct-tenant-a"},
			{ai.ProviderDeepSeek, ds.Model, "deepseek", "acct-tenant-a-deepseek"},
		}
		for i, o := range results {
			x, meta, m := want[i], o.res.Metadata, o.res.Message
			ev.Check(calls[i].name+" call succeeds", o.err == nil && m.StopReason == ai.StopReasonStop, "err=%v", o.err)
			ev.Check(calls[i].name+" result names its provider, API, model, binding and account",
				meta.ProviderID == x.provider && meta.API == ai.APIOpenAIResponses && meta.ModelID == x.model &&
					meta.BindingID == x.binding && meta.AccountScopeID == x.account, "metadata %+v", meta)
			ev.Check(calls[i].name+" message names its provider", m.Provider == x.provider && m.Model == x.model,
				"provider %q model %q", m.Provider, m.Model)
			envs := o.envs
			ev.Check(calls[i].name+" every event carries the call's attribution", len(envs) > 0 &&
				!slices.ContainsFunc(envs, func(e ai.EventEnvelope) bool { return e.Call.ProviderID != x.provider || e.Call.BindingID != x.binding }),
				"got %d envelopes", len(envs))
		}

		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if !ev.Check("one request per call", len(reqs) == 2, "got %d", len(reqs)) {
			return
		}
		byPath := map[string]provider.Request{}
		for _, r := range reqs {
			byPath[r.Path] = r
		}
		oa, okA := byPath["/v1/responses"]
		dr, okD := byPath["/responses"]
		ev.Check("each endpoint received one request", okA && okD, "paths %v", reqs)
		ev.Check("OpenAI's endpoint received tenant A's OpenAI key", oa.KeyAlias == tenantA.alias, "got %q", oa.KeyAlias)
		ev.Check("DeepSeek's endpoint received tenant A's DeepSeek key", dr.KeyAlias == deepseekA.alias, "got %q", dr.KeyAlias)
		ofields, dfields := bodyFields(t, oa.Body), bodyFields(t, dr.Body)
		ev.Check("OpenAI keeps its stateless storage and prompt cache",
			string(ofields["store"]) == "false" && ofields["prompt_cache_key"] != nil && string(ofields["prompt_cache_retention"]) == `"24h"` &&
				oa.Header["Session_id"] != "", "body %s", oa.Body)
		for _, f := range []string{"store", "prompt_cache_key", "prompt_cache_retention", "prompt_cache_options", "include",
			"previous_response_id", "conversation", "service_tier"} {
			_, present := dfields[f]
			ev.Check("DeepSeek's request carries no "+f, !present, "body %s", dr.Body)
		}
		_, affinity := dr.Header["Session_id"]
		ev.Check("DeepSeek's request carries no session affinity", !affinity && dr.Header["X-Client-Request-Id"] == "",
			"headers %v", dr.Header)
	})

	// Each binding resolves models in its own provider's catalog: a model
	// the binding allows is still refused when that provider does not
	// serve it on the protocol, before any request.
	for _, c := range []struct {
		name, binding, model string
	}{
		{"deepseek-model-on-openai-binding", "primary", "deepseek-flash"},
		{"openai-model-on-deepseek-binding", "deepseek", "gpt-4.1-mini"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "P05-E07-"+c.name)
			w := newWorld(t, tenantA)
			w.updateBindingOf(tenantA, c.binding, func(b *ai.Binding) {
				if !slices.Contains(b.AllowedModels, c.model) {
					b.AllowedModels = append(b.AllowedModels, c.model)
				}
			})
			res, err := w.client.Complete(ctxFor(t), textScope("req-"+c.name), ai.Target{BindingID: c.binding, ModelID: c.model},
				ai.Request{Messages: []ai.Message{ai.UserText("hi")}}, nil)
			ev.Record("result", res)
			ev.Check("refused as outside the provider's catalog", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseCapability}),
				"got %v", err)
			ev.Check("no request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		})
	}

	// spec Testing Decisions §5: DeepSeek Responses is proven by its own
	// fixtures, registered as an extension route in the differential
	// ledger, and none of its scenarios runs as a pi differential case.
	t.Run("registered-as-extension", func(t *testing.T) {
		ev := run.Case(t, "P05-E10-registered-as-extension")
		route, ok := loadLedger(t).Route(string(ai.ProviderDeepSeek), string(ai.APIOpenAIResponses))
		ev.Record("route", route)
		ev.Check("the ledger registers DeepSeek × Responses as an extension route", ok, "no route entry")
		for _, file := range deepseekResponsesFixtureFiles {
			f, _ := loadFixture(t, deepseekResponsesProtocol, file)
			for _, sc := range f.Scenarios {
				ev.Check(file+" "+sc.ID+" stays out of the differential", sc.PidiffSkip != "", "no pidiffSkip")
			}
		}
	})
}

func bodyFields(t *testing.T, body []byte) map[string]json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	mustUnmarshal(t, body, &fields)
	return fields
}
