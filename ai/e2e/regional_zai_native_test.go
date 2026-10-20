package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func TestRegionalZAIDefaultThinkingResearchReplay(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			t.Run(model+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "RP06-zai-omitted-thinking-"+model+"-"+entry)
				cat, err := regional.CandidateCatalog()
				if err != nil {
					t.Fatal(err)
				}
				w, target := regionalWorld(t, ai.ProviderZAI, ai.APIAnthropicMessages, func(c *ai.Config) { c.Catalog = &cat })
				target.ModelID = model
				c := zaiResearchCase(t, model+"-omitted")
				enqueue(ev, w, c.reply(t))
				req, scope := glmMessagesCapturedRequest(t, c), textScope("zai-omitted-thinking")
				var o outcome
				switch entry {
				case "stream":
					o = drain(w.client.Stream(ctxFor(t), scope, target, req, ai.AnthropicOptions{MaxTokens: ai.Value(4096)}))
				case "complete":
					o.result, o.err = w.client.Complete(ctxFor(t), scope, target, req, ai.AnthropicOptions{MaxTokens: ai.Value(4096)})
				case "stream-simple":
					o = drain(w.client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(4096)}))
				default:
					o.result, o.err = w.client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(4096)})
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("own omitted control succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
					return
				}
				ev.Check("omission preserves exact research request", jsonEqual(w.provider.Requests()[0].Body, c.request(t)), "omitted request changed")
				zaiCheckResearchNative(t, o, model+"-omitted")
			})
		}
	}
}

// Only turn one is a real raw research response. Turn two is an explicitly
// controlled text response: this tests Client replay, never actual API acceptance.
func TestRegionalZAIResearchNativeContinuation(t *testing.T) {
	for _, origin := range []string{"same", "restored"} {
		t.Run(origin, func(t *testing.T) {
			ev := run.Case(t, "RP07-zai-research-native-controlled-"+origin)
			cat, err := regional.CandidateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			w, target := regionalWorld(t, ai.ProviderZAI, ai.APIAnthropicMessages, func(c *ai.Config) { c.Catalog = &cat })
			target.ModelID = "glm-5.2"
			c := zaiResearchCase(t, "glm-5.2-enabled-budget1024")
			enqueue(ev, w, c.reply(t), regionalTextReply(t, ai.ProviderZAI, ai.APIAnthropicMessages, target.ModelID))
			req := glmMessagesCapturedRequest(t, c)
			first := drain(w.client.Stream(ctxFor(t), textScope("zai-research-first"), target, req, glmMessagesEnabled()))
			first.record(ev)
			if !ev.Check("own exact research turn succeeds", first.err == nil, "%v", first.err) {
				return
			}
			zaiCheckResearchNative(t, first, "glm-5.2-enabled-budget1024")
			m := first.result.Message
			if origin == "restored" {
				var env *ai.NativeStateEnvelope
				m, env, _ = restore(t, persist(t, m))
				if env == nil {
					t.Fatal("missing native envelope")
				}
				m.NativeState, err = ai.TrustNativeState(textScope("zai-research-restore"), *env)
				if err != nil {
					t.Fatal(err)
				}
			}
			before := persist(t, m)
			req.Messages = append(req.Messages, m)
			calls := 0
			var native ai.Thinking
			for i, b := range m.Content {
				switch v := b.(type) {
				case ai.Thinking:
					native = v
				case ai.ToolCall:
					valid, err := ai.ValidateToolCall(m, i, req.Tools)
					var args struct{ Label string }
					if !ev.Check("Host validates own tool before execution", err == nil && valid.Decode(&args) == nil && args.Label == "alpha", "tool validation failed") {
						return
					}
					req.Messages = append(req.Messages, ai.ToolResultText(v, "REGIONAL42", false))
					calls++
				}
			}
			second := drain(w.client.Stream(ctxFor(t), textScope("zai-research-controlled-next"), target, req, glmMessagesEnabled()))
			ev.Record("controlled-continuation", second.result)
			ev.Record("requests", w.provider.Requests())
			ev.Record("scope", "own raw turn one; controlled synthetic turn two; no live qualification")
			if !ev.Check("controlled continuation succeeds without downgrade or mutation", second.err == nil && len(w.provider.Requests()) == 2 && calls == 1 && second.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}) && string(before) == string(persist(t, m)), "%v", second.err) {
				return
			}
			var body struct {
				Messages []struct{ Content []map[string]any }
			}
			mustUnmarshal(t, w.provider.Requests()[1].Body, &body)
			n := body.Messages[1].Content[0]
			ev.Check("same own native container preserves exact text and signature", n["type"] == "thinking" && n["thinking"] == native.Thinking && n["signature"] == native.Signature && native.Signature != "", "native replay mismatch")
		})
	}
}
