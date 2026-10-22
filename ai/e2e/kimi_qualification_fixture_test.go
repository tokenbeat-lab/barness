package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func kimiQualificationBinary(t *testing.T) (string, string) {
	t.Helper()
	source, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "kimi-live.test")
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource="+supportmatrix.RegionalSourceHash(source), "-o", binary, "./ai/live")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native build: %v %s", err, regionalSafeOutput(raw))
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	return binary, evidence.Digest(raw)
}

// These are synthetic Kimi-native response schemas, with complete billed
// counters and deliberate unknown actual TTL. They establish no live support.
func kimiQualificationReply(t *testing.T, api, kind string) provider.Reply {
	t.Helper()
	tool := kind == "tool-call"
	native := kind == "full-stream"
	var frames []json.RawMessage
	switch api {
	case "anthropic-messages":
		frames = append(frames, json.RawMessage(`{"type":"message_start","message":{"id":"kimi-q-msg","model":"kimi-k3","usage":{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}}}`))
		index := "0"
		if native {
			frames = append(frames, json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}`), json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Kimi synthetic reasoning."}}`), json.RawMessage(`{"type":"content_block_stop","index":0}`))
			index = "1"
		}
		stop := "end_turn"
		if tool {
			frames = append(frames, json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"opaque|/kimi","name":"inspect_regional","input":{}}}`), json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"label\":\"alpha\"}"}}`), json.RawMessage(`{"type":"content_block_stop","index":0}`))
			stop = "tool_use"
		} else {
			frames = append(frames, json.RawMessage(`{"type":"content_block_start","index":`+index+`,"content_block":{"type":"text","text":""}}`), json.RawMessage(`{"type":"content_block_delta","index":`+index+`,"delta":{"type":"text_delta","text":"REGIONAL42 Kimi ready."}}`), json.RawMessage(`{"type":"content_block_stop","index":`+index+`}`))
		}
		if kind != "cancel" {
			frames = append(frames, json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"`+stop+`"},"usage":{"output_tokens":50}}`), json.RawMessage(`{"type":"message_stop"}`))
		}
	case "openai-completions":
		delta := `{"content":"REGIONAL42 Kimi ready."}`
		stop := "stop"
		if native {
			delta = `{"content":"REGIONAL42 Kimi ready.","reasoning_content":"Kimi synthetic reasoning."}`
		}
		if tool {
			delta = `{"tool_calls":[{"index":0,"id":"opaque|/kimi","type":"function","function":{"name":"inspect_regional","arguments":"{\"label\":\"alpha\"}"}}]}`
			stop = "tool_calls"
		}
		frames = append(frames, json.RawMessage(`{"id":"kimi-q-chat","model":"kimi-k3","choices":[{"index":0,"delta":`+delta+`}]}`))
		if kind != "cancel" {
			frames = append(frames, json.RawMessage(`{"id":"kimi-q-chat","model":"kimi-k3","choices":[{"index":0,"delta":{},"finish_reason":"`+stop+`"}]}`), json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":200,"completion_tokens":50,"total_tokens":250,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"completion_tokens_details":{"reasoning_tokens":10}}}`))
		}
		reply := chatEvents(t, frames, "")
		if kind == "cancel" {
			reply.End = provider.EndHold
		}
		return reply
	case "openai-responses":
		output := `[{"type":"message","id":"kimi-q-message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"REGIONAL42 Kimi ready.","annotations":[]}]}]`
		if native {
			output = `[{"type":"reasoning","id":"kimi-q-reasoning","summary":[{"type":"summary_text","text":"Kimi synthetic reasoning."}]},{"type":"message","id":"kimi-q-message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"REGIONAL42 Kimi ready.","annotations":[]}]}]`
		}
		if tool {
			output = `[{"type":"function_call","id":"item|/kimi","call_id":"opaque|/kimi","name":"inspect_regional","arguments":"{\"label\":\"alpha\"}","status":"completed"}]`
		}
		frames = append(frames, json.RawMessage(`{"type":"response.created","response":{"id":"kimi-q-response","model":"kimi-k3","status":"in_progress","output":[]}}`))
		if kind == "cancel" {
			frames = append(frames, json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"kimi-q-message","role":"assistant","content":[]}}`), json.RawMessage(`{"type":"response.content_part.added","output_index":0,"content_index":0,"part":{"type":"output_text","text":""}}`), json.RawMessage(`{"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"Kimi partial."}`))
		} else {
			frames = append(frames, json.RawMessage(`{"type":"response.completed","response":{"id":"kimi-q-response","model":"kimi-k3","status":"completed","output":`+output+`,"usage":{"input_tokens":200,"output_tokens":50,"total_tokens":250,"input_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"output_tokens_details":{"reasoning_tokens":10}}}}`))
		}
	}
	reply := sseEvents(t, frames, provider.FramingLF)
	if kind == "cancel" {
		reply.End = provider.EndHold
	}
	return reply
}

func kimiQualificationKinds() []string {
	return strings.Fields("full-stream full-complete simple-stream simple-complete native-history tool-call tool-result cancel recovery")
}
