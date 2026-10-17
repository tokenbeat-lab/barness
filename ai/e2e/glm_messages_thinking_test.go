package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type glmMessagesCapture struct {
	Name, Source, SSE, SHA256, Request, ResponseID string
	Content                                        json.RawMessage
	Input, Output                                  int64
	StopReason                                     ai.StopReason
}

func glmMessagesCaptureByName(t *testing.T, name string) glmMessagesCapture {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-messages/tools-thinking.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct{ Cases []glmMessagesCapture }
	mustUnmarshal(t, raw, &fixtures)
	for _, c := range fixtures.Cases {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no Messages capture %s", name)
	return glmMessagesCapture{}
}

func (c glmMessagesCapture) reply(t *testing.T) provider.Reply {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-messages/" + c.SSE)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	if hex.EncodeToString(hash[:]) != c.SHA256 {
		t.Fatal("immutable Messages tool/thinking SSE changed")
	}
	return provider.SSE([][]byte{raw})
}

func (c glmMessagesCapture) request(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-messages/" + c.Request)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func glmMessagesEnabled() ai.AnthropicOptions {
	return ai.AnthropicOptions{MaxTokens: ai.Value(4096), ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 1024}
}

// Only this exact 5.2 wire shape was tested against the actual Messages API;
// it supplies no canonical effort mapping or evidence of budget enforcement.
func TestGLMMessagesExactThinking(t *testing.T) {
	c := glmMessagesCaptureByName(t, "glm-5.2-thinking-enabled")
	for _, entry := range []string{"stream", "complete", "payload"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "GLM07-exact-thinking-"+entry)
			w := newGLMMessagesWorld(t, nil)
			enqueue(ev, w, c.reply(t))
			wire := c.request(t)
			ev.Fixture(c.Request, wire)
			var body struct {
				Messages []struct{ Content []struct{ Text string } }
			}
			mustUnmarshal(t, wire, &body)
			req := ai.Request{Messages: []ai.Message{ai.UserText(body.Messages[0].Content[0].Text)}}
			target := ai.Target{BindingID: "glm-messages", ModelID: "glm-5.2"}
			var o outcome
			switch entry {
			case "stream":
				o = drain(w.client.Stream(ctxFor(t), textScope("glm07-thinking"), target, req, glmMessagesEnabled()))
			case "complete":
				o.result, o.err = w.client.Complete(ctxFor(t), textScope("glm07-thinking"), target, req, glmMessagesEnabled())
			case "payload":
				hooks := glmMessagesPayload(func(p *ai.Payload) {
					p.Body["max_tokens"] = 4096
					p.Body["thinking"] = map[string]any{"type": "enabled", "budget_tokens": 1024}
				})
				o = drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm07-thinking"), target, req, nil))
			}
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("proven exact thinking request succeeds", o.err == nil, "%v", o.err)
			ev.Check("all thinking text and signature fragments preserved", jsonEqual(mustMarshal(t, o.result.Message.Content), c.Content), "%+v", o.result.Message.Content)
			ev.Check("actual stop and usage preserved", o.result.Message.StopReason == c.StopReason && o.result.Message.Usage.Input == c.Input && o.result.Message.Usage.Output == c.Output, "%+v", o.result.Message)
			if ev.Check("one actual Messages attempt", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata) {
				ev.Check("exact observed wire without Claude extensions", jsonEqual(w.provider.Requests()[0].Body, wire), "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}
