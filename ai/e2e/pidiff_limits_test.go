package e2e

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type limitPidiffCase struct {
	id string
	sc pidiffScenario
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
		{"request-body-at-limit", textScenario(setRequestBytes, int(sentBodySize(t, outcomeEntries[0], text, textRaw)))},
		{"image-at-limit", imageScenario},
		{"frame-at-limit", textScenario(setFrameBytes, largestFrame)},
		{"frame-over-limit", textScenario(setFrameBytes, largestFrame-1)},
		{"output-at-limit", textScenario(setOutputBytes, streamBytes)},
		{"output-over-limit", textScenario(setOutputBytes, streamBytes-1)},
		{"tool-json-at-limit", toolScenario(arguments)},
		{"tool-json-over-limit", toolScenario(arguments - 1)},
		{"error-body-at-limit", errorBodyScenario(len(rateLimited.Reply.Body))},
		{"error-body-over-limit", errorBodyScenario(len(rateLimited.Reply.Body) - 1)},
	}
}
