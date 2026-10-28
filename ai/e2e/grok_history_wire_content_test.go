package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Mutate only a real public Client capture; source Host input and pairing IDs
// stay intact, so the independent native proof must inspect function content.
func grokV6MutateCalls(t *testing.T, api ai.API, raw []byte, mutation string) []byte {
	t.Helper()
	var body map[string]any
	mustUnmarshal(t, raw, &body)
	var calls []map[string]any
	var names, args string
	switch api {
	case ai.APIOpenAIResponses:
		names, args = "name", "arguments"
		for _, v := range body["input"].([]any) {
			if m := v.(map[string]any); m["type"] == "function_call" {
				calls = append(calls, m)
			}
		}
	case ai.APIOpenAICompletions:
		names, args = "name", "arguments"
		for _, v := range body["messages"].([]any) {
			if list, ok := v.(map[string]any)["tool_calls"].([]any); ok {
				for _, v := range list {
					calls = append(calls, v.(map[string]any)["function"].(map[string]any))
				}
			}
		}
	case ai.APIAnthropicMessages:
		names, args = "name", "input"
		for _, v := range body["messages"].([]any) {
			if list, ok := v.(map[string]any)["content"].([]any); ok {
				for _, v := range list {
					if m := v.(map[string]any); m["type"] == "tool_use" {
						calls = append(calls, m)
					}
				}
			}
		}
	case ai.APIGoogleGenerativeAI:
		names, args = "name", "args"
		for _, v := range body["contents"].([]any) {
			for _, v := range v.(map[string]any)["parts"].([]any) {
				if m, ok := v.(map[string]any)["functionCall"].(map[string]any); ok {
					calls = append(calls, m)
				}
			}
		}
	}
	if len(calls) != 3 {
		t.Fatalf("captured calls=%d", len(calls))
	}
	value := func(label string, extra bool) any {
		m := map[string]any{"label": label}
		if extra {
			m["unexpected"] = true
		}
		if args == "arguments" {
			return string(mustMarshal(t, m))
		}
		return m
	}
	switch mutation {
	case "swapped-arguments":
		calls[0][args], calls[1][args] = calls[1][args], calls[0][args]
	case "function-name":
		calls[1][names] = "wrong_history_function"
	case "extra-argument":
		calls[0][args] = value("first", true)
	case "reused-argument":
		calls[2][args] = value("second", false)
	case "result-function-name":
		for _, v := range body["contents"].([]any) {
			for _, v := range v.(map[string]any)["parts"].([]any) {
				if m, ok := v.(map[string]any)["functionResponse"].(map[string]any); ok {
					m["name"] = "wrong_history_function"
				}
			}
		}
	case "semantic-json":
		if args == "arguments" {
			calls[0][args] = " { \"label\" : \"first\" } "
		}
	}
	return mustMarshal(t, body)
}

func TestGrokHistoryWireFunctionContent(t *testing.T) {
	for _, route := range []string{"openai-responses", "openai-chat", "anthropic-messages", "google-gemini"} {
		t.Run(route, func(t *testing.T) {
			w, b, model := grokHistoryWorld(t, route)
			reply, _ := regionalHostScript(t, b, model)
			if b.API == ai.APIGoogleGenerativeAI {
				reply = dataEvents(t, []json.RawMessage{json.RawMessage(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Ready."}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":8,"candidatesTokenCount":4,"totalTokenCount":12},"modelVersion":"gemini-3.8-flash"}`)}, provider.FramingLF)
			}
			w.provider.Enqueue(reply)
			scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: t.Name()}
			req, err := grokhistory.Request(scope)
			if err != nil {
				t.Fatal(err)
			}
			o := drain(w.client.StreamSimple(ctxFor(t), scope, ai.Target{BindingID: b.BindingID, ModelID: model}, req, ai.SimpleOptions{MaxTokens: ai.Value(64)}))
			if o.err != nil || len(w.provider.Requests()) != 1 {
				t.Fatalf("capture: %v", o.err)
			}
			original := w.provider.Requests()[0].Body
			mutations := []string{"swapped-arguments", "function-name", "extra-argument", "reused-argument", "semantic-json"}
			if b.API == ai.APIGoogleGenerativeAI {
				mutations = append(mutations, "result-function-name")
			}
			for _, mutation := range mutations {
				t.Run(mutation, func(t *testing.T) {
					ev := run.Case(t, "GROK-HISTORY-V6-CONTENT-"+route+"-"+mutation)
					raw := grokV6MutateCalls(t, b.API, original, mutation)
					err := grokhistory.Verify(b.API, raw)
					ev.Record("public-client-original", json.RawMessage(original))
					ev.Record("changed-wire", json.RawMessage(raw))
					ev.Check("proof matches function name and full arguments", (err == nil) == (mutation == "semantic-json"), "mutation=%s error=%v", mutation, err)
				})
			}
		})
	}
}
