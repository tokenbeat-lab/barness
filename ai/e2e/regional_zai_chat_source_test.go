package e2e

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func zaiChatSource(t *testing.T, model, kind string) (provider.Reply, ai.Request, []byte, string) {
	t.Helper()
	var index struct {
		Cases []struct {
			Model, Kind, ReasoningSHA256 string
			Files                        map[string]string
		}
	}
	raw, err := os.ReadFile("testdata/zai-chat-native-source/index.json")
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &index)
	for _, row := range index.Cases {
		if row.Model != model || row.Kind != kind {
			continue
		}
		for name, pin := range row.Files {
			raw, err := os.ReadFile("testdata/zai-chat-native-source/" + name)
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != pin {
				t.Fatal("immutable own Chat capture changed")
			}
		}
		request, err := os.ReadFile("testdata/zai-chat-native-source/" + model + "-" + kind + "-request.json")
		if err != nil {
			t.Fatal(err)
		}
		var body struct {
			Messages []struct{ Content json.RawMessage }
			Tools    []struct {
				Function struct {
					Name, Description string
					Parameters        json.RawMessage
				}
			}
		}
		mustUnmarshal(t, request, &body)
		var text string
		if json.Unmarshal(body.Messages[0].Content, &text) != nil {
			var parts []struct{ Text string }
			mustUnmarshal(t, body.Messages[0].Content, &parts)
			for _, part := range parts {
				text += part.Text
			}
		}
		req := ai.Request{Messages: []ai.Message{ai.UserText(text)}}
		for _, tool := range body.Tools {
			req.Tools = append(req.Tools, ai.Tool{Name: tool.Function.Name, Description: tool.Function.Description, Parameters: tool.Function.Parameters})
		}
		response, err := os.ReadFile("testdata/zai-chat-native-source/" + model + "-" + kind + "-response.sse")
		if err != nil {
			t.Fatal(err)
		}
		return provider.SSE([][]byte{response}), req, request, row.ReasoningSHA256
	}
	t.Fatal("own Chat source not found")
	return provider.Reply{}, ai.Request{}, nil, ""
}

// The captured tool response can legally contain no reasoning. Client must
// preserve the actual wire without inventing a Thinking block or native state.
func TestRegionalZAIChatOwnNativePresence(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, kind := range []string{"full-stream", "tool-call"} {
			t.Run(model+"/"+kind, func(t *testing.T) {
				ev := run.Case(t, "RP07-zai-chat-own-native-"+model+"-"+kind)
				cat, err := regional.CandidateCatalog()
				if err != nil {
					t.Fatal(err)
				}
				w, target := regionalWorld(t, ai.ProviderZAI, ai.APIOpenAICompletions, func(c *ai.Config) { c.Catalog = &cat })
				target.ModelID = model
				reply, req, wire, nativeHash := zaiChatSource(t, model, kind)
				enqueue(ev, w, reply)
				o := drain(w.client.Stream(ctxFor(t), textScope("zai-chat-presence"), target, req, ai.ChatOptions{MaxTokens: ai.Value(4096), ReasoningEffort: ai.ThinkingHigh}))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("exact own request succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
					return
				}
				ev.Check("own logical request encodes unchanged", jsonEqual(w.provider.Requests()[0].Body, wire), "request differs")
				thinking, blocks := "", 0
				for _, b := range o.result.Message.Content {
					if v, ok := b.(ai.Thinking); ok {
						thinking += v.Thinking
						blocks++
					}
				}
				ev.Check("exact raw native text retained without synthesis", fmt.Sprintf("%x", sha256.Sum256([]byte(thinking))) == nativeHash && (model != "glm-5.3" || kind != "tool-call" || blocks == 0), "thinking blocks=%d", blocks)
			})
		}
	}
}
