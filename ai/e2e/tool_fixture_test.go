package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// toolFixture is testdata/responses/tool-round-trip.json.
type toolFixture struct {
	Model string `json:"model"`
	// SimpleMaxOutputTokens is pi's simple-entry max_output_tokens.
	SimpleMaxOutputTokens int       `json:"simpleMaxOutputTokens"`
	SystemPrompt          string    `json:"systemPrompt"`
	User                  string    `json:"user"`
	Tools                 []ai.Tool `json:"tools"`
	ToolChoice            struct {
		Full   string `json:"full"`
		Simple string `json:"simple"`
	} `json:"toolChoice"`
	Round1      toolRound         `json:"round1"`
	ToolResults map[string]string `json:"toolResults"`
	Round2      toolRound         `json:"round2"`
	Truncated   []toolTruncation  `json:"truncated"`
}

type toolRound struct {
	Events []json.RawMessage `json:"events"`
	Expect struct {
		Body       json.RawMessage `json:"body"`
		Events     []string        `json:"events"`
		Content    []string        `json:"content"`
		ResponseID string          `json:"responseId"`
		StopReason ai.StopReason   `json:"stopReason"`
		Usage      ai.Usage        `json:"usage"`
	} `json:"expect"`
}

type toolTruncation struct {
	ID     string            `json:"id"`
	End    provider.End      `json:"end"`
	Events []json.RawMessage `json:"events"`
	Expect struct {
		StopReason ai.StopReason `json:"stopReason"`
		Code       ai.Code       `json:"code"`
		Events     []string      `json:"events"`
		Content    []string      `json:"content"`
	} `json:"expect"`
}

func loadToolFixture(t *testing.T) (toolFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "tool-round-trip.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f toolFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("tool-round-trip.json: %v", err)
	}
	return f, raw
}

func (f toolFixture) target() ai.Target { return ai.Target{BindingID: "primary", ModelID: f.Model} }

// request is the first round's input: the user's question and the tools.
func (f toolFixture) request() ai.Request {
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User)}, Tools: f.Tools}
}

// reply scripts a round's events, ending as end says.
func (f toolFixture) reply(t *testing.T, events []json.RawMessage, framing provider.Framing, end provider.End) provider.Reply {
	t.Helper()
	r := sseEvents(t, events, framing)
	r.End = end
	return r
}

// toolEntry is one public entry point. With force set, it asks for a tool
// call through the entry's own tool-choice option: the full Responses
// tool_choice, or the protocol-neutral simple one.
type toolEntry struct {
	name   string
	simple bool
	stream bool
}

var toolEntries = []toolEntry{
	{"stream-full", false, true},
	{"stream-simple", true, true},
	{"complete-full", false, false},
	{"complete-simple", true, false},
}

// toolChoice is the wire tool_choice the entry sends when forcing.
func (e toolEntry) toolChoice(f toolFixture) string {
	if e.simple {
		return f.ToolChoice.Simple
	}
	return f.ToolChoice.Full
}

func (e toolEntry) invoke(ctx context.Context, w *world, f toolFixture, scope ai.CallScope, req ai.Request, force bool) outcome {
	full := ai.ResponsesOptions{}
	simple := ai.SimpleOptions{}
	if force {
		full.ToolChoice = ai.Value(ai.ResponsesToolChoice{Mode: f.ToolChoice.Full})
		simple.ToolChoice = ai.Value(ai.ToolChoice(f.ToolChoice.Simple))
	}
	switch {
	case e.stream && e.simple:
		return drain(w.client.StreamSimple(ctx, scope, f.target(), req, simple))
	case e.stream:
		return drain(w.client.Stream(ctx, scope, f.target(), req, full))
	case e.simple:
		res, err := w.client.CompleteSimple(ctx, scope, f.target(), req, simple)
		return outcome{result: res, err: err}
	default:
		res, err := w.client.Complete(ctx, scope, f.target(), req, full)
		return outcome{result: res, err: err}
	}
}

// withToolChoice is body with tool_choice set, or body itself for "".
func withToolChoice(t *testing.T, body json.RawMessage, choice string) []byte {
	t.Helper()
	if choice == "" {
		return body
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	m["tool_choice"] = choice
	return mustMarshal(t, m)
}

// round2Request is the host's second logical call: the question, round 1's
// message and one result per call in it.
func (f toolFixture) round2Request(assistant ai.AssistantMessage) ai.Request {
	history := []ai.Message{ai.UserText(f.User), assistant}
	for _, block := range assistant.Content {
		if call, ok := block.(ai.ToolCall); ok {
			history = append(history, ai.ToolResultText(call, f.ToolResults[call.Name], false))
		}
	}
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: history, Tools: f.Tools}
}

// round1Message is the message barness produces for round 1.
func (f toolFixture) round1Message(t *testing.T) ai.AssistantMessage {
	t.Helper()
	w := newWorld(t, tenantA)
	w.provider.Enqueue(f.reply(t, f.Round1.Events, provider.FramingLF, ""))
	res, err := w.client.Complete(ctxFor(t), textScope("req-tools-round1"), f.target(), f.request(), nil)
	if err != nil || res.Message.StopReason != ai.StopReasonToolUse {
		t.Fatalf("round 1: %v (stop reason %q)", err, res.Message.StopReason)
	}
	return res.Message
}

// toolPidiffScenario is TestToolRoundTrip's round 1 ("tool-call", with the
// entry's own tool choice; "tool-call-function" forcing get_weather) or round
// 2 ("tool-results") for the differential.
func toolPidiffScenario(t *testing.T, name, entry string) pidiffScenario {
	t.Helper()
	f, raw := loadToolFixture(t)
	sc := pidiffScenario{fixture: "tool-round-trip.json", raw: raw, model: f.Model, requests: 1}
	if name == "tool-results" {
		sc.req = f.round2Request(f.round1Message(t))
		sc.reply = f.reply(t, f.Round2.Events, provider.FramingLF, "")
		return sc
	}
	sc.req = f.request()
	sc.reply = f.reply(t, f.Round1.Events, provider.FramingLF, "")
	switch {
	case name == "tool-call-function":
		sc.full = ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Function: "get_weather"})}
		sc.piOptions = map[string]any{"toolChoice": map[string]any{"type": "function", "name": "get_weather"}}
	case entry == "stream":
		sc.full = ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: f.ToolChoice.Full})}
		sc.piOptions = map[string]any{"toolChoice": f.ToolChoice.Full}
	default:
		sc.simple = ai.SimpleOptions{ToolChoice: ai.Value(ai.ToolChoice(f.ToolChoice.Simple))}
		sc.piOptions = map[string]any{"toolChoice": f.ToolChoice.Simple}
	}
	return sc
}
