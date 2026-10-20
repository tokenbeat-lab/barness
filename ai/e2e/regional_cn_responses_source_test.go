package e2e

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// These own-derived controlled responses are independent: full-stream has native
// reasoning, whereas the valid tool turn has none. Never select by success.
func regionalCNResponsesSource(t *testing.T, kind string) provider.Reply {
	t.Helper()
	const base = "testdata/minimax-cn-responses-native-source"
	raw, err := os.ReadFile(base + "/index.json")
	if err != nil {
		t.Fatal(err)
	}
	var index struct {
		Cases []struct {
			Kind  string
			Files map[string]string
		}
	}
	mustUnmarshal(t, raw, &index)
	found := false
	for _, row := range index.Cases {
		for name, pin := range row.Files {
			data, e := os.ReadFile(base + "/" + name)
			if e != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != pin {
				t.Fatal("own CN Responses capture changed")
			}
		}
		found = found || row.Kind == kind
	}
	if !found {
		t.Fatal("only fixed own full-stream/tool-call captures allowed")
	}
	response, err := os.ReadFile(base + "/" + kind + "-response.sse")
	if err != nil {
		t.Fatal(err)
	}
	return provider.SSE([][]byte{response})
}

func regionalCNResponsesNativeOriginals(t *testing.T, dir string) {
	t.Helper()
	const combo = "minimax-cn-responses"
	ev := run.Case(t, "RP18-minimax-cn-responses-fixed-native-originals")
	read := func(caseID, name string, v any) {
		raw, err := os.ReadFile(filepath.Join(dir, caseID, name))
		if err != nil {
			t.Fatal(err)
		}
		mustUnmarshal(t, raw, v)
	}
	first := supportmatrix.RegionalCaseID(combo, "MiniMax-M3/full-stream")
	tool := supportmatrix.RegionalCaseID(combo, "MiniMax-M3/tool-call")
	next := supportmatrix.RegionalCaseID(combo, "MiniMax-M3/native-history")
	type saved struct {
		Message struct {
			Content    []struct{ Thinking, Name string }
			StopReason ai.StopReason
		}
		Metadata ai.CallMetadata
	}
	var full, call, continued saved
	read(first, "result.json", &full)
	read(tool, "result.json", &call)
	read(next, "result.json", &continued)
	thinking := ""
	tools, toolThinking := 0, 0
	for _, b := range full.Message.Content {
		thinking += b.Thinking
	}
	for _, b := range call.Message.Content {
		if b.Name == "inspect_regional" {
			tools++
		}
		if b.Thinking != "" {
			toolThinking++
		}
	}
	ev.Check("actual own fixed text source preserves native reasoning", len(thinking) == 150 && fmt.Sprintf("%x", sha256.Sum256([]byte(thinking))) == "96ae55c00845a526e2462facc9097e62ab8a7d9a0a4530823334f19d82260c61", "thinking=%d", len(thinking))
	ev.Check("original legal tool without thinking remains intact", toolThinking == 0 && tools == 1 && call.Message.StopReason == ai.StopReasonToolUse, "thinking=%d tools=%d stop=%s", toolThinking, tools, call.Message.StopReason)
	ev.Check("same own continuation has no native downgrade", continued.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "downgrades=%+v", continued.Metadata.NativeStateDowngrades)
	var request, firstRequest struct{ Messages []json.RawMessage }
	var firstResult, toolResult struct{ Message json.RawMessage }
	read(next, "request.json", &request)
	read(first, "request.json", &firstRequest)
	read(first, "result.json", &firstResult)
	read(tool, "result.json", &toolResult)
	prefix := len(request.Messages) == 6 && len(firstRequest.Messages) == 1
	if prefix {
		prefix = jsonEqual(request.Messages[0], firstRequest.Messages[0]) && jsonEqual(request.Messages[1], firstResult.Message) && jsonEqual(request.Messages[3], toolResult.Message)
	}
	ev.Check("exact predeclared user assistant and independent tool chain", prefix, "messages=%d", len(request.Messages))
	var exchanges []struct{ RequestBody string }
	read(next, "exchanges.json", &exchanges)
	if len(exchanges) != 1 {
		t.Fatal("one continuation wire required")
	}
	var wire struct {
		Input []struct {
			Type    string
			CallID  string `json:"call_id"`
			Output  string
			Summary json.RawMessage
		}
	}
	mustUnmarshal(t, []byte(exchanges[0].RequestBody), &wire)
	reasons, calls, outputs := 0, 0, 0
	for _, item := range wire.Input {
		switch item.Type {
		case "reasoning":
			if jsonEqual(item.Summary, mustMarshal(t, []map[string]string{{"type": "summary_text", "text": thinking}})) {
				reasons++
			}
		case "function_call":
			if item.CallID == "controlled-call-2-00000000000" {
				calls++
			}
		case "function_call_output":
			if item.CallID == "controlled-call-2-00000000000" && item.Output == "REGIONAL42" {
				outputs++
			}
		}
	}
	ev.Check("actual reasoning native container and complete Host tool pair", reasons == 1 && calls == 1 && outputs == 1, "reasoning=%d calls=%d outputs=%d", reasons, calls, outputs)
}
