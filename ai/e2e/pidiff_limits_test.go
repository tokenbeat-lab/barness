package e2e

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type limitPidiffCase struct {
	id string
	sc pidiffScenario
	// entry is the entry point both sides use; empty is "stream".
	entry string
}

// limitPidiffScenarios are TestByteLimits' scenarios on the full stream
// entry, each at its boundary and, where barness still sends the request,
// at limit+1. Request and image refusals send nothing, so they have no pi
// counterpart to compare and are asserted offline only.
func limitPidiffScenarios(t *testing.T) []limitPidiffCase {
	t.Helper()
	text, textRaw := loadTextFixture(t, "text-basic.json")
	chunks := lfChunks(t, text.Events)
	largestFrame, streamBytes := largestAndTotal(chunks)
	textScenario := func(set func(*ai.ResourcePolicy, int64), n int) pidiffScenario {
		return pidiffScenario{fixture: "text-basic.json", raw: textRaw, model: text.Model, req: textRequest(text),
			reply: provider.SSE(chunks), requests: 1, policy: func(p *ai.ResourcePolicy) { set(p, int64(n)) }}
	}

	tool, toolRaw := loadToolFixture(t)
	arguments := longestArguments(t, tool.Round1.Events)
	toolScenario := func(n int) pidiffScenario {
		sc := toolPidiffScenario(t, "tool-call", "stream")
		sc.raw = toolRaw
		sc.policy = func(p *ai.ResourcePolicy) { setToolJSONBytes(p, int64(n)) }
		return sc
	}

	failures, failuresRaw := loadFailureFixture(t)
	rateLimited := failureScenarioByID(t, failures, "http-429-retry-after")
	errorBodyScenario := func(n int) pidiffScenario {
		sc := failurePidiffScenario(t, failures, failuresRaw, rateLimited)
		sc.policy = func(p *ai.ResourcePolicy) { setErrorBodyBytes(p, int64(n)) }
		return sc
	}

	image := ai.Image{Data: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xAB}, 47)), MimeType: "image/png"}
	imageScenario := textScenario(setImageBytes, 47)
	imageScenario.req = ai.Request{SystemPrompt: text.SystemPrompt, Messages: []ai.Message{
		ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: text.User}, image}},
	}}

	return []limitPidiffCase{
		{id: "request-body-at-limit", sc: textScenario(setRequestBytes, int(sentBodySize(t, outcomeEntries[0], text, textRaw)))},
		{id: "image-at-limit", sc: imageScenario},
		{id: "frame-at-limit", sc: textScenario(setFrameBytes, largestFrame)},
		{id: "frame-over-limit", sc: textScenario(setFrameBytes, largestFrame-1)},
		{id: "output-at-limit", sc: textScenario(setOutputBytes, streamBytes)},
		{id: "output-over-limit", sc: textScenario(setOutputBytes, streamBytes-1)},
		{id: "tool-json-at-limit", sc: toolScenario(arguments)},
		{id: "tool-json-over-limit", sc: toolScenario(arguments - 1)},
		{id: "error-body-at-limit", sc: errorBodyScenario(len(rateLimited.Reply.Body))},
		{id: "error-body-over-limit", sc: errorBodyScenario(len(rateLimited.Reply.Body) - 1)},
	}
}

// timeoutPidiffScenarios are TestTimeouts' protocol timeout scenarios: an
// upstream that never answers ends at timeoutMs with pi's error terminal on
// both entries, and a timed-out attempt is retried as pi retries it. The
// policy's own limits are far later, so only timeoutMs acts.
func timeoutPidiffScenarios(t *testing.T) []limitPidiffCase {
	t.Helper()
	text, textRaw := loadTextFixture(t, "text-basic.json")
	const ms = 200
	silent := provider.Reply{End: provider.EndSilent}
	timedOut := func(entry string) limitPidiffCase {
		return limitPidiffCase{id: "protocol-timeout", entry: entry, sc: pidiffScenario{fixture: "text-basic.json", raw: textRaw,
			model: text.Model, req: textRequest(text), reply: silent, requests: 1,
			full: ai.ResponsesOptions{TimeoutMs: ms}, simple: ai.SimpleOptions{TimeoutMs: ms}, piOptions: map[string]any{"timeoutMs": ms}}}
	}
	retried := pidiffScenario{fixture: "text-basic.json", raw: textRaw, model: text.Model, req: textRequest(text),
		replies: []provider.Reply{silent, sseReply(t, text, provider.FramingLF)}, requests: 2,
		retry: ai.RetryPolicy{MaxRetries: 1}, clock: clock.NewFake(time.Unix(1790000000, 0), 0),
		full: ai.ResponsesOptions{TimeoutMs: ms}, piOptions: map[string]any{"timeoutMs": ms, "maxRetries": 1}}
	return []limitPidiffCase{timedOut("stream"), timedOut("streamSimple"), {id: "protocol-timeout-retried", sc: retried}}
}
