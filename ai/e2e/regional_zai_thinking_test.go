package e2e

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

// This is a replay of own research, not a candidate or live-history claim.
// The root acceptance round retains immutable source receipts separately.
func TestRegionalZAIExactThinking(t *testing.T) {
	for _, entry := range []string{"stream", "complete", "payload"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "RP06-zai-exact-thinking-"+entry)
			cat, err := regional.CandidateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			w, target := regionalWorld(t, ai.ProviderZAI, ai.APIAnthropicMessages, func(c *ai.Config) { c.Catalog = &cat })
			target.ModelID = "glm-5.2"
			c := zaiResearchCase(t, "glm-5.2-enabled-budget1024")
			enqueue(ev, w, c.reply(t))
			req := glmMessagesCapturedRequest(t, c)
			var o outcome
			switch entry {
			case "stream":
				o = drain(w.client.Stream(ctxFor(t), textScope("zai-explicit-thinking"), target, req, glmMessagesEnabled()))
			case "complete":
				o.result, o.err = w.client.Complete(ctxFor(t), textScope("zai-explicit-thinking"), target, req, glmMessagesEnabled())
			default:
				o = drain(w.client.WithHooks(glmMessagesPayload(func(p *ai.Payload) {
					p.Body["max_tokens"] = 4096
					p.Body["thinking"] = map[string]any{"type": "enabled", "budget_tokens": 1024}
				})).Stream(ctxFor(t), textScope("zai-explicit-thinking"), target, req, nil))
			}
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("own exact full thinking shape succeeds", o.err == nil && len(w.provider.Requests()) == 1, "err=%v", o.err) {
				return
			}
			ev.Check("exact research request retained without extra controls", jsonEqual(w.provider.Requests()[0].Body, c.request(t)), "wire differs from own request")
			zaiCheckResearchNative(t, o, "glm-5.2-enabled-budget1024")
		})
	}
}

func zaiResearchCase(t *testing.T, id string) glmMessagesCapture {
	t.Helper()
	var index struct {
		Cases []struct {
			ID, Model string
			Files     map[string]string
		}
	}
	raw, err := os.ReadFile("testdata/zai-messages-research/index.json")
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &index)
	for _, row := range index.Cases {
		if row.ID != id {
			continue
		}
		for file, pin := range row.Files {
			raw, err := os.ReadFile("testdata/zai-messages-research/" + file)
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != pin {
				t.Fatal("own research original changed")
			}
		}
		return glmMessagesCapture{Request: "../zai-messages-research/" + id + "-request.json", SSE: "../zai-messages-research/" + id + "-response.sse", SHA256: row.Files[id+"-response.sse"]}
	}
	t.Fatal("missing own research case")
	return glmMessagesCapture{}
}

func zaiCheckResearchNative(t *testing.T, o outcome, id string) {
	t.Helper()
	var index struct {
		Cases []struct {
			ID, ThinkingSHA256, SignatureSHA256 string
			ThinkingChars, SignatureChars       int
		}
	}
	raw, err := os.ReadFile("testdata/zai-messages-research/index.json")
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &index)
	for _, row := range index.Cases {
		if row.ID != id {
			continue
		}
		thinking, signature := "", ""
		for _, b := range o.result.Message.Content {
			if v, ok := b.(ai.Thinking); ok {
				thinking += v.Thinking
				signature += v.Signature
			}
		}
		if len([]rune(thinking)) != row.ThinkingChars || len(signature) != row.SignatureChars || fmt.Sprintf("%x", sha256.Sum256([]byte(thinking))) != row.ThinkingSHA256 || fmt.Sprintf("%x", sha256.Sum256([]byte(signature))) != row.SignatureSHA256 {
			t.Fatal("native thinking/signature differ from own raw research deltas")
		}
		return
	}
	t.Fatal("missing native expectation")
}
