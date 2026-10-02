package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// deepseekResponsesProtocol is DeepSeek × Responses (P05): scenarios in
// testdata/deepseek-responses, the "deepseek" binding on tenant A's DeepSeek
// account and key, and replies as Responses SSE events. It shares the
// Responses adapter with OpenAI but nothing of its configuration.
var deepseekResponsesProtocol = &fixtureProtocol{
	dir: "deepseek-responses", binding: "deepseek", api: ai.APIOpenAIResponses, provider: ai.ProviderDeepSeek,
	key: deepseekKey, account: "acct-tenant-a-deepseek",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.ResponsesOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: sseEvents, textMarker: `"delta":"Hello`,
}

// deepseekResponsesFixtureFiles are the P05 fixtures. None runs in the pi
// differential: frozen pi has no DeepSeek Responses route (ADR-0014).
var deepseekResponsesFixtureFiles = []string{"text.json", "failures.json", "history.json", "unsupported.json", "usage.json", "retry.json"}

// TestDeepSeekResponsesText is P05/E01 through the public Client: text,
// reasoning_text and a reasoning tool call; the text scenarios also run
// through Result without Next and through Complete.
func TestDeepSeekResponsesText(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "text.json", "E01", func(sc fixtureScenario) bool { return sc.ID == "text" || sc.ID == "text-simple" })
}

// TestDeepSeekResponsesFailures is P05/E02: incomplete, failed, missing
// terminal, HTTP refusals and cancellation.
func TestDeepSeekResponsesFailures(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "failures.json", "E02", func(sc fixtureScenario) bool { return sc.CancelAfterEvents == 0 })
}

// TestDeepSeekResponsesHistory is P05/E03: stateless full-history replay,
// reasoning and tool round trips, other providers' state and images.
func TestDeepSeekResponsesHistory(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "history.json", "E03", nil)
}

// TestDeepSeekResponsesUsage is P05/E11: cache reads, reasoning, partial
// and unreported usage, priced at DeepSeek's rates.
func TestDeepSeekResponsesUsage(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "usage.json", "E11", nil)
}

// TestDeepSeekResponsesRetry is P05/E05 under the binding's retry policy.
func TestDeepSeekResponsesRetry(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "retry.json", "E05", nil)
}

// runProtocolScenarios streams every scenario of proto's file as case
// "<p>-<e>-<scenario>-<mode>"; those more selects also run through Result
// without Next and through Complete.
func runProtocolScenarios(t *testing.T, proto *fixtureProtocol, p, file, e string, more func(fixtureScenario) bool) {
	f, raw := loadFixture(t, proto, file)
	for _, sc := range f.Scenarios {
		modes := []callMode{modeStream}
		if more != nil && more(sc) {
			modes = append(modes, modeResult, modeComplete)
		}
		for _, mode := range modes {
			t.Run(sc.ID+"/"+string(mode), func(t *testing.T) {
				ev := run.Case(t, p+"-"+e+"-"+sc.ID+"-"+string(mode))
				runScenario(t, ev, sc, raw, mode)
			})
		}
	}
}

// TestDeepSeekResponsesUnsupported is P05's unsupported fields: what
// DeepSeek ignores is never derived or sent (testdata/deepseek-responses/
// unsupported.json), and what a caller explicitly asks for that DeepSeek
// does not serve, or that would point at server-side state, is refused
// before any request.
func TestDeepSeekResponsesUnsupported(t *testing.T) {
	runProtocolScenarios(t, deepseekResponsesProtocol, "P05", "unsupported.json", "E04", nil)

	f, raw := loadFixture(t, deepseekResponsesProtocol, "text.json")
	text := scenarioByID(t, f, "text")
	refused := []struct {
		name    string
		options ai.Options
		hooks   ai.Hooks
		code    ai.Code
	}{
		{"service-tier", ai.ResponsesOptions{ServiceTier: ai.Value("priority")}, ai.Hooks{}, ai.CodeInvalidRequest},
		{"sampling-previous-response-id", ai.ResponsesOptions{SamplingParams: map[string]json.RawMessage{
			"previous_response_id": json.RawMessage(`"resp_ds_text"`)}}, ai.Hooks{}, ai.CodeInvalidRequest},
		{"sampling-conversation", ai.ResponsesOptions{SamplingParams: map[string]json.RawMessage{
			"conversation": json.RawMessage(`"conv_1"`)}}, ai.Hooks{}, ai.CodeInvalidRequest},
		{"sampling-store", ai.ResponsesOptions{SamplingParams: map[string]json.RawMessage{
			"store": json.RawMessage(`false`)}}, ai.Hooks{}, ai.CodeInvalidRequest},
		{"payload-previous-response-id", nil, payloadSetting("previous_response_id", "resp_ds_text"), ai.CodeTenantDenied},
		{"payload-conversation", nil, payloadSetting("conversation", "conv_1"), ai.CodeTenantDenied},
		{"payload-store", nil, payloadSetting("store", true), ai.CodeTenantDenied},
		{"payload-prompt-cache-key", nil, payloadSetting("prompt_cache_key", "k"), ai.CodeTenantDenied},
	}
	for _, c := range refused {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "P05-E04-refused-"+c.name)
			ev.Fixture("text.json", raw)
			w := scenarioWorld(t, text)
			opts := c.options
			if opts == nil {
				opts = ai.ResponsesOptions{}
			}
			res, err := w.client.WithHooks(c.hooks).Complete(ctxFor(t), textScope("req-refused-"+c.name), text.target(), text.request(t), opts)
			ev.Record("result", res)
			var aiErr *ai.Error
			ev.Check("refused with "+string(c.code), errors.As(err, &aiErr) && aiErr.Code == c.code, "got %v", err)
			ev.Check("message ends as error", res.Message.StopReason == ai.StopReasonError && res.Message.ErrorMessage != "",
				"got %q %q", res.Message.StopReason, res.Message.ErrorMessage)
			ev.Check("no request reached DeepSeek", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
			ev.Check("no attempt recorded", len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata.Attempts)
		})
	}
}

// payloadSetting is a trusted payload callback that sets field on the
// request body.
func payloadSetting(field string, value any) ai.Hooks {
	return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		p.Body[field] = value
		return ai.KeepPayload(), nil
	}}
}

// TestDeepSeekResponsesToolRoundTrip is P05's tool round trip: the host
// validates DeepSeek's call, executes it and passes the first call's message
// straight into the second call, which replays the reasoning item, the call
// and the result as complete input with no server-side reference.
func TestDeepSeekResponsesToolRoundTrip(t *testing.T) {
	ev := run.Case(t, "P05-E03-live-tool-round-trip")
	tf, traw := loadFixture(t, deepseekResponsesProtocol, "text.json")
	hf, hraw := loadFixture(t, deepseekResponsesProtocol, "history.json")
	first := scenarioByID(t, tf, "reasoning-tool-call")
	answer := scenarioByID(t, hf, "tool-round-trip")
	ev.Fixture("text.json", traw)
	ev.Fixture("history.json", hraw)
	w := scenarioWorld(t, first)
	enqueue(ev, w, first.replies(t)...)
	enqueue(ev, w, answer.replies(t)...)

	req := first.request(t)
	res, err := w.client.Complete(ctxFor(t), textScope("req-round-1"), first.target(), req, ai.ResponsesOptions{})
	ev.Record("round-1", res)
	if !ev.Check("first round asks for a tool", err == nil && res.Message.StopReason == ai.StopReasonToolUse, "err=%v stop=%q", err, res.Message.StopReason) {
		return
	}
	call, err := ai.ValidateToolCall(res.Message, 1, req.Tools)
	if !ev.Check("the host validates the call", err == nil && call.ID() == "call_ds_1|fc_ds_1", "err=%v", err) {
		return
	}
	var args struct{ A, B int }
	ev.Check("arguments decode", call.Decode(&args) == nil && args.A == 17 && args.B == 19, "got %+v", args)

	req.Messages = append(req.Messages, res.Message, ai.ToolResultText(res.Message.Content[1].(ai.ToolCall), "36", false))
	res2, err := w.client.Complete(ctxFor(t), textScope("req-round-2"), first.target(), req, ai.ResponsesOptions{})
	ev.Record("round-2", res2)
	ev.Check("second round completes", err == nil && res2.Message.StopReason == ai.StopReasonStop, "err=%v", err)
	ev.Check("native state replayed, nothing downgraded", res2.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
		"got %+v", res2.Metadata.NativeStateDowngrades)

	reqs := w.provider.Requests()
	if !ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
		return
	}
	ev.Check("the second request is the stored round trip's", jsonEqual(reqs[1].Body, answer.Expect.Request.Body),
		"got %s\nwant %s", reqs[1].Body, answer.Expect.Request.Body)
}
