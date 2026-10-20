package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func regionalTextReply(t *testing.T, p ai.ProviderID, api ai.API, model string) provider.Reply {
	t.Helper()
	switch api {
	case ai.APIOpenAICompletions:
		return miniMaxChatReply(t, model)
	case ai.APIOpenAIResponses:
		content := `[{"type":"output_text","text":"Regional text."}]`
		if p == "zai" {
			content = `{"type":"output_text","text":"Regional text."}`
		}
		return minimaxResponsesReply(t, `{"type":"response.completed","response":{"id":"regional-response","model":"`+model+`","status":"completed","output":[{"type":"message","id":"msg-regional","role":"assistant","status":"completed","content":`+content+`}],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`)
	default:
		r := miniMaxMessagesReply(t, model)
		r.Header["X-Request-Id"] = "regional-request"
		return r
	}
}

func TestRegionalProviderText(t *testing.T) {
	for _, p := range []ai.ProviderID{"zai", "minimax-cn"} {
		for _, api := range regionalAPIs {
			for _, model := range regionalModels(p) {
				for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
					t.Run(string(p)+"/"+string(api)+"/"+model+"/"+entry, func(t *testing.T) {
						ev := run.Case(t, "RP05-text-"+t.Name())
						w, target := regionalWorld(t, p, api)
						target.ModelID = model
						enqueue(ev, w, regionalTextReply(t, p, api, target.ModelID))
						var o outcome
						scope, req := textScope("regional-text-"+entry), ai.Request{Messages: []ai.Message{ai.UserText("hello")}}
						switch entry {
						case "stream":
							o = drain(w.client.Stream(ctxFor(t), scope, target, req, nil))
						case "complete":
							o.result, o.err = w.client.Complete(ctxFor(t), scope, target, req, nil)
						case "stream-simple":
							o = drain(w.client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{}))
						default:
							o.result, o.err = w.client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{})
						}

						o.record(ev)
						ev.Record("requests", w.provider.Requests())
						if !ev.Check("one successful attempt", o.err == nil && len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%v", o.err) {
							return
						}
						ev.Check("real region attribution", o.result.Metadata.ProviderID == p && o.result.Message.Provider == p, "%+v", o.result.Metadata)
						ev.Check("content and usage survive", len(o.result.Message.Content) > 0 && o.result.Message.Usage.Output > 0, "%+v", o.result.Message)
						fields := bodyFields(t, w.provider.Requests()[0].Body)
						ev.Check("wire model unchanged", string(fields["model"]) == string(mustMarshal(t, target.ModelID)), "%s", fields["model"])
						for _, key := range []string{"prompt_cache_retention", "include"} {
							ev.Check("no original vendor "+key, len(fields[key]) == 0, "%s", w.provider.Requests()[0].Body)
						}
						if api == ai.APIOpenAICompletions {
							ev.Check("no original vendor store", len(fields["store"]) == 0, "%s", w.provider.Requests()[0].Body)
							if p == "minimax-cn" {
								ev.Check("MiniMax split enabled", string(fields["reasoning_split"]) == "true", "%s", fields["reasoning_split"])
							}
						}
						if api == ai.APIAnthropicMessages {
							ev.Check("default thinking omitted", len(fields["thinking"]) == 0, "%s", fields["thinking"])
							ev.Check("no native browser header", w.provider.Requests()[0].Header["Anthropic-Dangerous-Direct-Browser-Access"] == "", "%v", w.provider.Requests()[0].Header)
						}
						if api == ai.APIOpenAIResponses {
							ev.Check("terminal-only text preserved", json.Valid(fields["input"]) && len(o.result.Message.Content) > 0 && o.result.Message.Content[0].(ai.Text).Text == "Regional text.", "%+v", o.result.Message.Content)
						}
					})
				}
			}
		}
	}
}
