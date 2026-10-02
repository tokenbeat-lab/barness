package e2e

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestDeepSeekChatToolRoundTrip is P06's tool round trip with thinking on:
// the host validates DeepSeek's call, executes it and passes the first
// call's message straight into the second call, which sends the reasoning
// back as reasoning_content beside the call, as DeepSeek's thinking mode
// requires, and the result as a tool message.
func TestDeepSeekChatToolRoundTrip(t *testing.T) {
	ev := run.Case(t, "P06-E03-live-tool-round-trip")
	tf, traw := loadFixture(t, deepseekChatProtocol, "text.json")
	hf, hraw := loadFixture(t, deepseekChatProtocol, "history.json")
	first := scenarioByID(t, tf, "reasoning-tool-auto")
	answer := scenarioByID(t, hf, "tool-round-trip")
	ev.Fixture("text.json", traw)
	ev.Fixture("history.json", hraw)
	w := scenarioWorld(t, first)
	enqueue(ev, w, first.replies(t)...)
	enqueue(ev, w, answer.replies(t)...)

	req := first.request(t)
	opts := ai.ChatOptions{ReasoningEffort: ai.ThinkingHigh}
	res, err := w.client.Complete(ctxFor(t), textScope("req-round-1"), first.target(), req, opts)
	ev.Record("round-1", res)
	if !ev.Check("first round asks for a tool", err == nil && res.Message.StopReason == ai.StopReasonToolUse, "err=%v stop=%q", err, res.Message.StopReason) {
		return
	}
	call, err := ai.ValidateToolCall(res.Message, 1, req.Tools)
	if !ev.Check("the host validates the call", err == nil && call.ID() == "call_00_ds1", "err=%v", err) {
		return
	}
	var args struct{ A, B int }
	ev.Check("arguments decode", call.Decode(&args) == nil && args.A == 17 && args.B == 19, "got %+v", args)

	req.Messages = append(req.Messages, res.Message, ai.ToolResultText(res.Message.Content[1].(ai.ToolCall), "36", false))
	res2, err := w.client.Complete(ctxFor(t), textScope("req-round-2"), first.target(), req, opts)
	ev.Record("round-2", res2)
	ev.Check("second round completes", err == nil && res2.Message.StopReason == ai.StopReasonStop, "err=%v", err)
	ev.Check("native state replayed, nothing downgraded", res2.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
		"got %+v", res2.Metadata.NativeStateDowngrades)

	reqs := w.provider.Requests()
	if !ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
		return
	}
	ev.Check("the second request is the stored round trip's", jsonEqual(reqs[1].Body, answer.Expect.Request.Body),
		"got %s\nwant %s", reqs[1].Body, answer.Expect.Request.Body)
}

// TestDeepSeekChatRouting is P06's routing (E10) and authorization (E07):
// one DeepSeek account reached through two bindings, one per protocol, that
// reference the same credential; each call reaches its own protocol's
// endpoint path, request shape and model configuration, and is attributed
// to its binding, provider, API, model and the shared account.
func TestDeepSeekChatRouting(t *testing.T) {
	t.Run("dual-protocol-one-account", func(t *testing.T) {
		ev := run.Case(t, "P06-E10-dual-protocol-one-account")
		rf, rraw := loadFixture(t, deepseekResponsesProtocol, "text.json")
		cf, craw := loadFixture(t, deepseekChatProtocol, "text.json")
		rs, cs := scenarioByID(t, rf, "reasoning"), scenarioByID(t, cf, "reasoning")
		ev.Fixture("deepseek-responses-text.json", rraw)
		ev.Fixture("deepseek-chat-text.json", craw)
		// A host catalog that changes the Chat entry's level map leaves the
		// Responses entry of the same model alone: high goes out as max on
		// Chat only.
		catalog := ai.BuiltinCatalog()
		for i, m := range catalog.Models {
			if m.Provider == ai.ProviderDeepSeek && m.API == ai.APIOpenAICompletions && m.ID == "deepseek-flash" {
				catalog.Models[i].ThinkingLevelMap.High = ai.Value("max")
			}
		}
		w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &catalog }, tenantA)
		w.provider.EnqueueAt("/responses", rs.replies(t)...)
		w.provider.EnqueueAt("/chat/completions", cs.replies(t)...)

		// The same simple options on both, run concurrently on one Client.
		opts := ai.SimpleOptions{Reasoning: ai.ThinkingHigh}
		calls := []struct {
			name, binding string
			api           ai.API
			sc            fixtureScenario
		}{
			{"responses", "deepseek", ai.APIOpenAIResponses, rs},
			{"chat", "deepseek-chat", ai.APIOpenAICompletions, cs},
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
				s := w.client.StreamSimple(ctxFor(t), textScope("req-dual-"+c.name), c.sc.target(), c.sc.request(t), opts)
				defer s.Close()
				for s.Next() {
					results[i].envs = append(results[i].envs, s.Envelope())
				}
				results[i].res, results[i].err = s.Result()
			}()
		}
		wg.Wait()
		for i, r := range results {
			c, meta, m := calls[i], r.res.Metadata, r.res.Message
			ev.Record("result-"+c.name, r.res)
			ev.Record("envelopes-"+c.name, envelopesJSON(r.envs))
			ev.Check(c.name+" call succeeds", r.err == nil && m.StopReason == ai.StopReasonStop, "err=%v", r.err)
			ev.Check(c.name+" result names DeepSeek, its API, model and binding, and the shared account",
				meta.ProviderID == ai.ProviderDeepSeek && meta.API == c.api && meta.ModelID == "deepseek-flash" &&
					meta.BindingID == c.binding && meta.AccountScopeID == "acct-tenant-a-deepseek" && meta.CredentialVersion == "v1",
				"metadata %+v", meta)
			ev.Check(c.name+" message names its provider and API", m.Provider == ai.ProviderDeepSeek && m.API == c.api,
				"provider %q api %q", m.Provider, m.API)
			ev.Check(c.name+" every event carries the call's attribution", len(r.envs) > 0 &&
				!slices.ContainsFunc(r.envs, func(e ai.EventEnvelope) bool {
					return e.Call.ProviderID != ai.ProviderDeepSeek || e.Call.API != c.api || e.Call.BindingID != c.binding
				}), "got %d envelopes", len(r.envs))
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
		rr, okR := byPath["/responses"]
		cr, okC := byPath["/chat/completions"]
		ev.Check("each protocol's path received one request", okR && okC, "paths %v", reqs)
		ev.Check("both carry the shared credential's key", rr.KeyAlias == deepseekA.alias && cr.KeyAlias == deepseekA.alias,
			"responses %q chat %q", rr.KeyAlias, cr.KeyAlias)
		rfields, cfields := bodyFields(t, rr.Body), bodyFields(t, cr.Body)
		ev.Check("Responses keeps its own level map", string(rfields["reasoning"]) == `{"effort":"high","summary":"auto"}`, "body %s", rr.Body)
		ev.Check("Chat uses its own entry's level map and thinking switch",
			string(cfields["reasoning_effort"]) == `"max"` && string(cfields["thinking"]) == `{"type":"enabled"}`, "body %s", cr.Body)
		for _, f := range []string{"input", "instructions", "reasoning", "include", "max_output_tokens", "store", "text"} {
			_, present := cfields[f]
			ev.Check("Chat's request carries no Responses field "+f, !present, "body %s", cr.Body)
		}
		for _, f := range []string{"messages", "thinking", "reasoning_effort", "max_tokens", "stream_options"} {
			_, present := rfields[f]
			ev.Check("Responses' request carries no Chat field "+f, !present, "body %s", rr.Body)
		}
	})

	// Each binding resolves models in its own provider × API catalog: a
	// model the binding allows is still refused when that protocol of that
	// provider does not serve it, before any request.
	for _, c := range []struct {
		name, binding, model string
	}{
		{"chat-only-model-on-responses-binding", "deepseek", "deepseek-v4-pro"},
		{"openai-model-on-deepseek-chat-binding", "deepseek-chat", "gpt-4.1-mini"},
		{"deepseek-model-on-openai-chat-binding", "chat", "deepseek-flash"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "P06-E07-"+c.name)
			w := newWorld(t, tenantA)
			w.updateBindingOf(tenantA, c.binding, func(b *ai.Binding) {
				if !slices.Contains(b.AllowedModels, c.model) {
					b.AllowedModels = append(b.AllowedModels, c.model)
				}
			})
			res, err := w.client.Complete(ctxFor(t), textScope("req-"+c.name), ai.Target{BindingID: c.binding, ModelID: c.model},
				ai.Request{Messages: []ai.Message{ai.UserText("hi")}}, nil)
			ev.Record("result", res)
			ev.Check("refused as outside the protocol's catalog", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseCapability}),
				"got %v", err)
			ev.Check("no request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		})
	}

	// DeepSeek × Chat is a frozen pi route (pi's deepseek.json serves
	// openai-completions): unlike P05 it is not registered as an extension
	// and every P06 scenario runs in the differential.
	t.Run("pi-route", func(t *testing.T) {
		ev := run.Case(t, "P06-E10-pi-route")
		_, registered := loadLedger(t).Route(string(ai.ProviderDeepSeek), string(ai.APIOpenAICompletions))
		ev.Check("not registered as an extension route", !registered, "the ledger lists it")
		for _, file := range deepseekChatFixtureFiles {
			f, _ := loadFixture(t, deepseekChatProtocol, file)
			for _, sc := range f.Scenarios {
				ev.Check(file+" "+sc.ID+" runs in the differential", sc.PidiffSkip == "", "pidiffSkip %q", sc.PidiffSkip)
			}
		}
	})
}
