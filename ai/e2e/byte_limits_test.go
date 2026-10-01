package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestByteLimits covers E08's byte limits (spec I9, ADR-0002, User Story 33):
// each of the request body, an image, an SSE frame, the streamed output, a
// tool call's argument JSON and a provider error body at its boundary and at
// limit+1 on every entry point. Request and image limits refuse before
// anything is sent; the others are enforced while reading, so a provider
// that never finishes an oversized frame or body cannot hold the call.
func TestByteLimits(t *testing.T) {
	text, textRaw := loadTextFixture(t, "text-basic.json")
	textChunks := lfChunks(t, text.Events)
	largestFrame, streamBytes := largestAndTotal(textChunks)
	scriptText := func(_ *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
		ev.Fixture("text-basic.json", textRaw)
		enqueue(ev, lw.world, provider.SSE(textChunks))
	}
	textWithin := func(ev *evidence.Case, _ limitWorld, o outcome) { checkTextSucceeded(ev, text, o) }
	// keepsReceived: what arrived before the limit stays in the message.
	keepsReceived := func(ev *evidence.Case, _ limitWorld, o outcome) {
		got := textOf(o.result.Message)
		ev.Check("message keeps the text received before the limit", strings.HasPrefix(text.Expect.Text, got), "got %q", got)
	}
	nothingSent := func(ev *evidence.Case, lw limitWorld, o outcome) {
		ev.Check("no inference request sent", len(lw.provider.Requests()) == 0, "got %d", len(lw.provider.Requests()))
		ev.Check("no attempt recorded", len(o.result.Metadata.Attempts) == 0, "got %+v", o.result.Metadata.Attempts)
	}

	t.Run("request-body", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "request-body", field: "MaxRequestBytes", phase: ai.PhaseRequest, set: setRequestBytes,
			size:   func(t *testing.T, e outcomeEntry) int64 { return sentBodySize(t, e, text, textRaw) },
			script: scriptText, target: textTarget(text), request: textRequest(text),
			within: textWithin, over: nothingSent,
		})
	})

	t.Run("request-body-grown-by-payload-callback", func(t *testing.T) {
		// The limit applies to the body actually sent, after the trusted
		// payload callback, which here makes a body that fit too large.
		ev := run.Case(t, "E08-request-body-grown-by-payload-callback")
		ev.Fixture("text-basic.json", textRaw)
		size := sentBodySize(t, outcomeEntries[2], text, textRaw)
		lw := newLimitWorld(t, func(p *ai.ResourcePolicy) { setRequestBytes(p, size) })
		enqueue(ev, lw.world, provider.SSE(textChunks))
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["metadata"] = map[string]any{"pad": "x"}
			return ai.KeepPayload(), nil
		}}
		res, err := within(ev, "call ends", func() callResult {
			res, err := lw.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-e08-grown"), textTarget(text), textRequest(text), ai.ResponsesOptions{})
			return callResult{res, err}
		}).unpack()
		o := outcome{result: res, err: err}
		o.record(ev)
		checkLimited(ev, o, ai.PhaseRequest, "MaxRequestBytes")
		nothingSent(ev, lw, o)
		checkReleased(ev, lw)
	})

	image := ai.Image{Data: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xAB}, 47)), MimeType: "image/png"}
	imageRequest := ai.Request{SystemPrompt: text.SystemPrompt, Messages: []ai.Message{
		ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: text.User}, image}},
	}}
	t.Run("image", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "image", field: "MaxImageBytes", phase: ai.PhaseScope, set: setImageBytes,
			// The image's own bytes, not its base64 text.
			size: fixedSize(47), script: scriptText, target: textTarget(text), request: imageRequest,
			within: func(ev *evidence.Case, lw limitWorld, o outcome) {
				checkTextSucceeded(ev, text, o)
				reqs := lw.provider.Requests()
				ev.Check("the image was sent", len(reqs) == 1 && strings.Contains(string(reqs[0].Body), image.Data), "requests=%d", len(reqs))
			},
			over: nothingSent,
		})
	})

	t.Run("image-in-tool-result", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-image-in-tool-result-over-limit-"+e.name)
				lw := newLimitWorld(t, func(p *ai.ResourcePolicy) { setImageBytes(p, 46) })
				req := ai.Request{Messages: []ai.Message{
					ai.UserText(text.User),
					ai.ToolResultMessage{ToolCallID: "call_1", ToolName: "snapshot", Content: []ai.ToolResultContent{image}},
				}}
				o := within(ev, "call ends", func() outcome {
					return e.invoke(ctxFor(t), lw.world, textScope("req-e08-tool-image-"+e.name), textTarget(text), req)
				})
				o.record(ev)
				checkLimited(ev, o, ai.PhaseScope, "MaxImageBytes")
				nothingSent(ev, lw, o)
			})
		}
	})

	t.Run("frame", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "frame", field: "MaxFrameBytes", phase: ai.PhaseStream, set: setFrameBytes,
			size: fixedSize(largestFrame), script: scriptText, target: textTarget(text), request: textRequest(text),
			within: textWithin, over: keepsReceived,
		})
	})

	// CRLF line endings: a blank "\r\n" line ends a frame too.
	crlfChunks, err := provider.EncodeSSE(text.Events, provider.FramingCRLF)
	if err != nil {
		t.Fatal(err)
	}
	largestCRLF, _ := largestAndTotal(crlfChunks)
	t.Run("frame-crlf", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "frame-crlf", field: "MaxFrameBytes", phase: ai.PhaseStream, set: setFrameBytes,
			size: fixedSize(largestCRLF), target: textTarget(text), request: textRequest(text),
			script: func(_ *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
				ev.Fixture("text-basic.json", textRaw)
				enqueue(ev, lw.world, provider.SSE(crlfChunks))
			},
			within: textWithin, over: keepsReceived,
		})
	})

	t.Run("output", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "output", field: "MaxOutputBytes", phase: ai.PhaseStream, set: setOutputBytes,
			size: fixedSize(streamBytes), script: scriptText, target: textTarget(text), request: textRequest(text),
			within: textWithin, over: keepsReceived,
		})
	})

	tool, toolRaw := loadToolFixture(t)
	t.Run("tool-json", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "tool-json", field: "MaxToolJSONBytes", phase: ai.PhaseStream, set: setToolJSONBytes,
			size: fixedSize(longestArguments(t, tool.Round1.Events)),
			script: func(t *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
				ev.Fixture("tool-round-trip.json", toolRaw)
				enqueue(ev, lw.world, tool.reply(t, tool.Round1.Events, provider.FramingLF, provider.EndClose))
			},
			target: tool.target(), request: tool.request(),
			within: func(ev *evidence.Case, _ limitWorld, o outcome) {
				m := o.result.Message
				ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
				ev.Check("stop reason", m.StopReason == tool.Round1.Expect.StopReason, "got %q", m.StopReason)
				ev.Check("tool calls", reflect.DeepEqual(contentSummary(m.Content), tool.Round1.Expect.Content), "got %q", contentSummary(m.Content))
			},
			over: func(ev *evidence.Case, _ limitWorld, o outcome) {
				for _, c := range o.result.Message.Content {
					if call, ok := c.(ai.ToolCall); ok {
						ev.Check("no call holds more argument JSON than the limit", int64(len(call.RawArguments)) < int64(longestArguments(t, tool.Round1.Events)),
							"call %s holds %d bytes", call.ID, len(call.RawArguments))
					}
				}
			},
		})
	})

	failures, failuresRaw := loadFailureFixture(t)
	rateLimited := failureScenarioByID(t, failures, "http-429-retry-after")
	t.Run("error-body", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "error-body", field: "MaxErrorBodyBytes", phase: ai.PhaseRequest, set: setErrorBodyBytes,
			size: fixedSize(len(rateLimited.Reply.Body)),
			script: func(t *testing.T, ev *evidence.Case, lw limitWorld, over bool) {
				ev.Fixture("failures.json", failuresRaw)
				if over {
					// A limited error body ends the call; it is never retried.
					lw.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 2} })
				}
				enqueue(ev, lw.world, failures.reply(t, rateLimited), failures.reply(t, rateLimited), failures.reply(t, rateLimited))
			},
			target: failures.target(), request: failures.request(),
			within: func(ev *evidence.Case, lw limitWorld, o outcome) {
				checkFailureOutcome(ev, failures, rateLimited, o)
				checkProviderSaw(ev, lw.world, rateLimited)
			},
			over: func(ev *evidence.Case, lw limitWorld, o outcome) {
				var ae *ai.Error
				if errors.As(o.err, &ae) {
					ev.Check("HTTP status, request id and retry-after are kept", ae.HTTPStatus == 429 &&
						ae.ProviderRequestID == rateLimited.Expect.ProviderRequestID && ae.RetryAfter.Milliseconds() == rateLimited.Expect.RetryAfterMs,
						"got %d %q %s", ae.HTTPStatus, ae.ProviderRequestID, ae.RetryAfter)
				}
				checkProviderSaw(ev, lw.world, rateLimited)
				ev.Check("the attempt is recorded as limited", len(o.result.Metadata.Attempts) == 1 &&
					o.result.Metadata.Attempts[0].Code == ai.CodeResourceLimit, "got %+v", o.result.Metadata.Attempts)
			},
		})
	})

	// A redirect's body is read by barness itself, not by the SDK.
	redirectBody := `{"moved":"` + strings.Repeat("r", 40) + `"}`
	t.Run("error-body-redirect", func(t *testing.T) {
		runBoundary(t, boundary{
			id: "error-body-redirect", field: "MaxErrorBodyBytes", phase: ai.PhaseRequest, set: setErrorBodyBytes,
			size: fixedSize(len(redirectBody)),
			script: func(_ *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
				enqueue(ev, lw.world, provider.Reply{Status: 307, Header: map[string]string{"Location": "https://elsewhere.invalid/v1/responses", "Content-Type": "application/json"},
					Chunks: [][]byte{[]byte(redirectBody)}, Script: redirectBody})
			},
			target: textTarget(text), request: textRequest(text),
			within: func(ev *evidence.Case, _ limitWorld, o outcome) {
				ev.Check("a redirect is a protocol failure", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseRequest}), "got %v", o.err)
			},
			over: func(ev *evidence.Case, _ limitWorld, o outcome) {
				var ae *ai.Error
				ev.Check("HTTP status is kept", errors.As(o.err, &ae) && ae.HTTPStatus == 307, "got %v", o.err)
			},
		})
	})

	t.Run("held", func(t *testing.T) { testHeldOversize(t) })
}

// testHeldOversize scripts providers that exceed a limit and then keep the
// connection open without finishing: a limit checked only after reading a
// whole frame or body would wait forever, so ending in time proves the check
// happens while reading.
func testHeldOversize(t *testing.T) {
	created := `{"type":"response.created","sequence_number":0,"response":{"id":"resp_held","object":"response","created_at":1790000000,"status":"in_progress","model":"gpt-4.1-mini-2025-04-14","output":[],"usage":null}}`
	createdFrame := "event: response.created\ndata: " + created + "\n\n"
	heartbeats := strings.Repeat(": "+strings.Repeat("h", 150)+"\n\n", 20)
	cases := []struct {
		id, field string
		phase     ai.Phase
		set       func(*ai.ResourcePolicy)
		reply     provider.Reply
	}{
		{"frame", "MaxFrameBytes", ai.PhaseStream, func(p *ai.ResourcePolicy) { setFrameBytes(p, 1024) },
			heldReply(200, "text/event-stream", createdFrame, "event: response.output_text.delta\ndata: "+strings.Repeat("x", 1100))},
		{"output", "MaxOutputBytes", ai.PhaseStream, func(p *ai.ResourcePolicy) { setOutputBytes(p, 2048) },
			heldReply(200, "text/event-stream", createdFrame, heartbeats)},
		{"error-body", "MaxErrorBodyBytes", ai.PhaseRequest, func(p *ai.ResourcePolicy) { setErrorBodyBytes(p, 1024) },
			heldReply(500, "application/json", `{"error":{"message":"`+strings.Repeat("e", 1100))},
	}
	for _, c := range cases {
		for _, e := range outcomeEntries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-held-"+c.id+"-"+e.name)
				lw := newLimitWorld(t, c.set)
				enqueue(ev, lw.world, c.reply)
				target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
				o := within(ev, "call ends while the provider still holds the connection", func() outcome {
					return e.invoke(ctxFor(t), lw.world, textScope("req-e08-held-"+c.id+"-"+e.name), target, ai.Request{Messages: []ai.Message{ai.UserText("hi")}})
				})
				o.record(ev)
				checkLimited(ev, o, c.phase, c.field)
				checkReleased(ev, lw)
			})
		}
	}
}

func heldReply(status int, contentType string, chunks ...string) provider.Reply {
	r := provider.Reply{Status: status, Header: map[string]string{"Content-Type": contentType}, End: provider.EndHold}
	for _, c := range chunks {
		r.Chunks = append(r.Chunks, []byte(c))
		r.Script += c
	}
	return r
}

// sentBodySize is the size of the request body entry e sends for the text
// fixture, measured on a Provider under the default test policy.
func sentBodySize(t *testing.T, e outcomeEntry, f textFixture, raw []byte) int64 {
	t.Helper()
	w := newWorld(t, tenantA)
	w.provider.Enqueue(sseReply(t, f, provider.FramingLF))
	if o := e.invoke(ctxFor(t), w, textScope("req-e08-measure-"+e.name), textTarget(f), textRequest(f)); o.err != nil {
		t.Fatalf("measuring call failed: %v", o.err)
	}
	reqs := w.provider.Requests()
	if len(reqs) != 1 {
		t.Fatalf("measuring call sent %d requests", len(reqs))
	}
	return int64(len(reqs[0].Body))
}

// longestArguments is the longest function call argument JSON in events.
func longestArguments(t *testing.T, events []json.RawMessage) int {
	t.Helper()
	longest := 0
	for _, raw := range events {
		var ev struct {
			Type      string `json:"type"`
			Arguments string `json:"arguments"`
		}
		mustUnmarshal(t, raw, &ev)
		if ev.Type == "response.function_call_arguments.done" {
			longest = max(longest, len(ev.Arguments))
		}
	}
	return longest
}

func failureScenarioByID(t *testing.T, f failureFixture, id string) failureScenario {
	t.Helper()
	for _, sc := range f.Scenarios {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("failures.json has no scenario %q", id)
	return failureScenario{}
}

func (r callResult) unpack() (ai.Result, error) { return r.res, r.err }
