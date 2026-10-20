package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Independent synthetic ZAI Responses: text produces its own native reasoning,
// while the valid fragmented tool turn reports reasoning_tokens=0 and no native
// item. This fixture exposes source selection through the public CLI/Client.
func regionalZAIResponsesSource(t *testing.T, model, kind string) provider.Reply {
	t.Helper()
	if kind == "full-stream" {
		return minimaxResponsesReply(t, `{"type":"response.completed","response":{"id":"zai-text-source-`+model+`","model":"`+model+`","status":"completed","output":[{"type":"reasoning","id":"zai-source-reasoning","summary":[],"content":[{"type":"reasoning_text","text":"Regional source reasoning."}]},{"type":"message","id":"zai-source-message","role":"assistant","status":"completed","content":{"type":"output_text","text":"Regional text."}}],"usage":{"input_tokens":12,"output_tokens":8,"total_tokens":20,"output_tokens_details":{"reasoning_tokens":3}}}}`)
	}
	if kind != "tool-call" {
		t.Fatal("fixed Responses source fixture requires full-stream or tool-call")
	}
	return minimaxResponsesReply(t,
		`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"opaque-item","call_id":"regional-call","name":"inspect_regional","arguments":""}}`,
		`{"type":"response.function_call_arguments.delta","output_index":0,"delta":"{\"label\":"}`,
		`{"type":"response.function_call_arguments.delta","output_index":0,"delta":"\"alpha\"}"}`,
		`{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"opaque-item","call_id":"regional-call","name":"inspect_regional","arguments":"{\"label\":\"alpha\"}"}}`,
		`{"type":"response.completed","response":{"id":"zai-tool-source-`+model+`","model":"`+model+`","status":"completed","output":[],"usage":{"input_tokens":12,"output_tokens":7,"total_tokens":19,"output_tokens_details":{"reasoning_tokens":0}}}}`)
}

// Inspect originals produced by the full public process. The two independent
// source messages and Host result chain must be saved without fabricating a
// Thinking block on the tool turn or downgrading the fixed text source.
func regionalResponsesNativeOriginals(t *testing.T, dir string) {
	t.Helper()
	ev := run.Case(t, "RP18-zai-responses-fixed-native-originals")
	read := func(caseID, name string, v any) {
		raw, e := os.ReadFile(filepath.Join(dir, caseID, name))
		if e != nil {
			t.Fatal(e)
		}
		mustUnmarshal(t, raw, v)
	}
	type saved struct {
		Message struct {
			Content []struct {
				Name     string  `json:"name"`
				Thinking *string `json:"thinking"`
			}
			StopReason ai.StopReason `json:"stopReason"`
		}
		Metadata ai.CallMetadata `json:"metadata"`
	}
	for _, model := range supportmatrix.RegionalModels("zai-responses") {
		first := supportmatrix.RegionalCaseID("zai-responses", model+"/full-stream")
		tool := supportmatrix.RegionalCaseID("zai-responses", model+"/tool-call")
		next := supportmatrix.RegionalCaseID("zai-responses", model+"/native-history")
		var full, call, continued saved
		read(first, "result.json", &full)
		read(tool, "result.json", &call)
		read(next, "result.json", &continued)
		thinking, tools := 0, 0
		for _, b := range call.Message.Content {
			if b.Thinking != nil {
				thinking++
			}
			if b.Name == "inspect_regional" {
				tools++
			}
		}
		ev.Check("legal tool response has no invented Thinking "+model, thinking == 0 && tools == 1 && call.Message.StopReason == ai.StopReasonToolUse, "thinking=%d tools=%d", thinking, tools)
		native := false
		for _, b := range full.Message.Content {
			native = native || b.Thinking != nil && *b.Thinking == "Regional source reasoning."
		}
		ev.Check("fixed text turn owns actual reasoning "+model, native, "native=%t", native)
		ev.Check("own continuation has no native downgrade "+model, continued.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "downgrades=%+v", continued.Metadata.NativeStateDowngrades)
		var request struct{ Messages []json.RawMessage }
		read(next, "request.json", &request)
		var firstRequest struct{ Messages []json.RawMessage }
		read(first, "request.json", &firstRequest)
		var firstResult, toolResult struct{ Message json.RawMessage }
		read(first, "result.json", &firstResult)
		read(tool, "result.json", &toolResult)
		prefix := len(request.Messages) == 6 && len(firstRequest.Messages) == 1
		if prefix {
			prefix = jsonEqual(request.Messages[0], firstRequest.Messages[0]) && jsonEqual(request.Messages[1], firstResult.Message) && jsonEqual(request.Messages[3], toolResult.Message)
		}
		ev.Check("domain continuation keeps exact fixed source and independent tool chain "+model, prefix, "messages=%d", len(request.Messages))
		var exchanges []struct {
			RequestBody string `json:"requestBody"`
		}
		read(next, "exchanges.json", &exchanges)
		var wire struct {
			Input []struct {
				Type    string          `json:"type"`
				CallID  string          `json:"call_id"`
				Output  string          `json:"output"`
				Content json.RawMessage `json:"content"`
			}
		}
		if len(exchanges) != 1 {
			t.Fatal("native continuation requires one HTTP request")
		}
		mustUnmarshal(t, []byte(exchanges[0].RequestBody), &wire)
		reasoning, calls, outputs := 0, 0, 0
		for _, item := range wire.Input {
			switch item.Type {
			case "reasoning":
				if jsonEqual(item.Content, []byte(`{"type":"reasoning_text","text":"Regional source reasoning."}`)) {
					reasoning++
				}
			case "function_call":
				if item.CallID == "regional-call" {
					calls++
				}
			case "function_call_output":
				if item.CallID == "regional-call" && item.Output == "REGIONAL42" {
					outputs++
				}
			}
		}
		ev.Check("wire retains native reasoning and exact Host tool pair "+model, reasoning == 1 && calls == 1 && outputs == 1, "reasoning=%d calls=%d outputs=%d", reasoning, calls, outputs)
	}
}
