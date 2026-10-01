package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// nativeFixture is testdata/responses/native-state.json.
type nativeFixture struct {
	Model string `json:"model"`
	// SimpleMaxOutputTokens is pi's simple-entry max_output_tokens.
	SimpleMaxOutputTokens int         `json:"simpleMaxOutputTokens"`
	OtherModel            string      `json:"otherModel"`
	SystemPrompt          string      `json:"systemPrompt"`
	User                  string      `json:"user"`
	Tools                 []ai.Tool   `json:"tools"`
	ToolResult            string      `json:"toolResult"`
	Redacted              ai.Thinking `json:"redacted"`
	Round1                struct {
		Events []json.RawMessage `json:"events"`
		Expect struct {
			StopReason ai.StopReason `json:"stopReason"`
			Content    []string      `json:"content"`
		} `json:"expect"`
	} `json:"round1"`
	Round2 struct {
		Events []json.RawMessage `json:"events"`
		Expect struct {
			StopReason     ai.StopReason   `json:"stopReason"`
			Content        []string        `json:"content"`
			NativeBody     json.RawMessage `json:"nativeBody"`
			DowngradedBody json.RawMessage `json:"downgradedBody"`
		} `json:"expect"`
	} `json:"round2"`
}

func loadNativeFixture(t *testing.T) (nativeFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "native-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f nativeFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("native-state.json: %v", err)
	}
	return f, raw
}

func (f nativeFixture) target(model string) ai.Target {
	return ai.Target{BindingID: "primary", ModelID: model}
}

func (f nativeFixture) round1Request() ai.Request {
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User)}, Tools: f.Tools}
}

// round2Request replays round 1's message with the host-held redacted block
// inserted before its tool call, then the tool's result.
func (f nativeFixture) round2Request(assistant ai.AssistantMessage) ai.Request {
	call := 0
	for i, c := range assistant.Content {
		if _, ok := c.(ai.ToolCall); ok {
			call = i
			break
		}
	}
	assistant.Content = slices.Insert(slices.Clone(assistant.Content), call, ai.AssistantContent(f.Redacted))
	result := ai.ToolResultText(assistant.Content[call+1].(ai.ToolCall), f.ToolResult, false)
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User), assistant, result}, Tools: f.Tools}
}

func (f nativeFixture) round1Reply(t *testing.T) provider.Reply {
	return sseEvents(t, f.Round1.Events, provider.FramingLF)
}

func (f nativeFixture) round2Reply(t *testing.T) provider.Reply {
	return sseEvents(t, f.Round2.Events, provider.FramingLF)
}

// round1 runs round 1 on model in w for tenant k and returns its result.
func (f nativeFixture) round1(t *testing.T, w *world, k tenantKey, model string) ai.Result {
	t.Helper()
	w.provider.Enqueue(f.round1Reply(t))
	scope := ai.CallScope{TenantID: k.tenant, RequestID: "req-native-round1-" + model}
	res, err := w.client.Complete(ctxFor(t), scope, f.target(model), f.round1Request(), nil)
	if err != nil || res.Message.StopReason != ai.StopReasonToolUse {
		t.Fatalf("round 1: %v (stop reason %q)", err, res.Message.StopReason)
	}
	return res
}

// storedTurn is the test host's persistence record for an assistant message:
// the host's own model, not barness-ai's (spec I1: a host converts its
// persistence model at its boundary). The envelope is stored beside the
// message; whatever else a client managed to put into the record is kept as
// raw JSON so a scenario can try to smuggle trust through it.
type storedTurn struct {
	Message  storedMessage           `json:"message"`
	Envelope *ai.NativeStateEnvelope `json:"envelope,omitempty"`
	// NativeState and Trusted are what a forged record claims.
	NativeState json.RawMessage `json:"nativeState,omitempty"`
	Trusted     bool            `json:"trusted,omitempty"`
}

type storedMessage struct {
	Content    []storedBlock `json:"content"`
	API        ai.API        `json:"api"`
	Provider   ai.ProviderID `json:"provider"`
	Model      string        `json:"model"`
	ResponseID string        `json:"responseId,omitempty"`
	Usage      ai.Usage      `json:"usage"`
	StopReason ai.StopReason `json:"stopReason"`
	Timestamp  int64         `json:"timestamp"`
}

// storedBlock is a content block with its kind; Data is the block's own JSON.
type storedBlock struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// persist is what the test host writes after a call: the message and, when the
// call resolved, the envelope its native state came with.
func persist(t *testing.T, m ai.AssistantMessage) []byte {
	t.Helper()
	rec := storedTurn{Message: storedMessage{API: m.API, Provider: m.Provider, Model: m.Model, ResponseID: m.ResponseID,
		Usage: m.Usage, StopReason: m.StopReason, Timestamp: m.Timestamp}}
	for _, c := range m.Content {
		var kind string
		switch c.(type) {
		case ai.Text:
			kind = "text"
		case ai.Thinking:
			kind = "thinking"
		case ai.ToolCall:
			kind = "toolCall"
		default:
			t.Fatalf("persist: unsupported block %T", c)
		}
		rec.Message.Content = append(rec.Message.Content, storedBlock{Type: kind, Data: mustMarshal(t, c)})
	}
	if env, ok := m.NativeState.Envelope(); ok {
		rec.Envelope = &env
	}
	return mustMarshal(t, rec)
}

// restore decodes a stored record into a barness-ai message and the envelope
// stored beside it. It never makes the state trusted: that is the host's
// separate, explicit decision through ai.TrustNativeState. The forged fields
// are decoded the only way JSON can reach barness-ai's types.
func restore(t *testing.T, data []byte) (ai.AssistantMessage, *ai.NativeStateEnvelope, storedTurn) {
	t.Helper()
	var rec storedTurn
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	s := rec.Message
	m := ai.AssistantMessage{API: s.API, Provider: s.Provider, Model: s.Model, ResponseID: s.ResponseID,
		Usage: s.Usage, StopReason: s.StopReason, Timestamp: s.Timestamp}
	for _, b := range s.Content {
		var block ai.AssistantContent
		var err error
		switch b.Type {
		case "text":
			var v ai.Text
			err = json.Unmarshal(b.Data, &v)
			block = v
		case "thinking":
			var v ai.Thinking
			err = json.Unmarshal(b.Data, &v)
			block = v
		case "toolCall":
			var v ai.ToolCall
			err = json.Unmarshal(b.Data, &v)
			block = v
		default:
			t.Fatalf("restore: unknown block %q", b.Type)
		}
		if err != nil {
			t.Fatal(err)
		}
		m.Content = append(m.Content, block)
	}
	if len(rec.NativeState) > 0 {
		// A forged record's "native state" decoded straight into the
		// library's trust type.
		if err := json.Unmarshal(rec.NativeState, &m.NativeState); err != nil {
			t.Fatal(err)
		}
	}
	return m, rec.Envelope, rec
}

// forge rewrites a stored record the way a malicious client would: it claims
// trust, names the victim tenant and keeps every signature.
func forge(t *testing.T, data []byte, env ai.NativeStateEnvelope) []byte {
	t.Helper()
	var rec map[string]any
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	rec["trusted"] = true
	rec["nativeState"] = map[string]any{"trusted": true, "env": env, "envelope": env, "tenantId": env.TenantID,
		"TenantID": env.TenantID, "AccountScopeID": env.AccountScopeID}
	return mustMarshal(t, rec)
}

// nativePidiffScenario is TestNativeStateReplay's round 1 ("reasoning-call"),
// or its round 2 replaying a trusted same-model message ("reasoning-replay")
// or another model's message ("reasoning-cross-model"), for the differential.
// pi has no trust concept, so only these two replays have a pi counterpart; a
// downgrade of same-model state is asserted offline to equal the cross-model
// body.
func nativePidiffScenario(t *testing.T, name string) pidiffScenario {
	t.Helper()
	f, raw := loadNativeFixture(t)
	sc := pidiffScenario{fixture: "native-state.json", raw: raw, model: f.Model, requests: 1}
	switch name {
	case "reasoning-call":
		sc.req, sc.reply = f.round1Request(), f.round1Reply(t)
	case "reasoning-replay", "reasoning-cross-model":
		source := f.Model
		if name == "reasoning-cross-model" {
			source = f.OtherModel
		}
		round1 := f.round1(t, newWorld(t, tenantA), tenantA, source)
		sc.req, sc.reply = f.round2Request(round1.Message), f.round2Reply(t)
	default:
		t.Fatalf("unknown native pidiff scenario %q", name)
	}
	return sc
}
