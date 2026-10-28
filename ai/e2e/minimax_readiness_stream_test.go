package e2e

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

// The retained stream is a legal prefix, not a complete success fixture. The
// injected cause is synthetic: the original recorder did not retain that cause.
func TestMiniMaxReadinessStreamBoundary(t *testing.T) {
	prefix, err := os.ReadFile("testdata/minimax-readiness/thinking-on-prefix.sse")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("testdata/minimax-readiness/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Thinking       string
		ResponseSHA256 string
		Request        struct {
			Messages []struct{ Content []struct{ Text string } }
		}
	}
	mustUnmarshal(t, raw, &fixture)
	wire, err := os.ReadFile("testdata/minimax-readiness/request.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"read-error", "eof", "complete", "request-eof"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "MMREADY-stream-"+kind)
			ev.Fixture("thinking-on-prefix.sse", prefix)
			ev.Fixture("provenance.json", raw)
			ev.Fixture("original-wire.json", wire)
			ev.Check("retained prefix matches its original hash", evidence.Digest(prefix) == fixture.ResponseSHA256, "changed fixture")
			transport := &miniMaxReadinessTransport{prefix: string(prefix), kind: kind}
			w := newMiniMaxSharedWorld(t, func(c *ai.Config) { c.Transport = transport })
			route := miniMaxSharedRoutes[2]
			binding := minimax.MessagesBinding(miniMaxSharedA.tenant, "acct-"+miniMaxSharedA.alias, "same-credential", w.provider.URL()+route.prefix(miniMaxSharedA))
			// Prove the SDK cannot replay an already-started stream even when
			// initial-request retries are explicitly configured.
			binding.Retry.MaxRetries = 1
			if kind == "request-eof" {
				binding.Retry.MaxRetries = 0
			}
			w.host.PutBinding(binding)
			req := ai.Request{Messages: []ai.Message{ai.UserText(fixture.Request.Messages[0].Content[0].Text)}}
			o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm-ready-"+kind), route.target("MiniMax-M3"), req, ai.AnthropicOptions{MaxTokens: ai.Value(2048), ThinkingEnabled: ai.Value(true)}))
			o.record(ev)
			ev.Record("observations", w.rec.all())
			ev.Record("controlled-network", map[string]any{"kind": kind, "requests": transport.calls.Load(), "bodyClosed": transport.closed.Load(), "readCauseSynthetic": kind == "read-error", "wire": transport.wire, "vendorCalls": 0})
			ev.Check("exactly one attempt and no replay", transport.calls.Load() == 1 && len(o.result.Metadata.Attempts) == 1, "attempts=%+v", o.result.Metadata.Attempts)
			ev.Check("current encoder reproduces original wire", jsonEqual([]byte(transport.wire), wire), "wire differs")
			thought, text := "", ""
			for _, p := range o.result.Message.Content {
				if v, ok := p.(ai.Thinking); ok {
					thought += v.Thinking
				}
				if v, ok := p.(ai.Text); ok {
					text += v.Text
				}
			}
			code, phase := ai.CodeTransport, ai.PhaseStream
			if kind == "eof" {
				code = ai.CodeProtocol
			}
			if kind == "request-eof" {
				phase = ai.PhaseRequest
			}
			if kind == "complete" {
				ev.Check("same legal prefix can finish normally", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop && text == "synthetic completion", "%v %+v", o.err, o.result.Message)
			} else {
				ev.Check("precise caller-visible failure", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}) && o.result.Message.StopReason == ai.StopReasonError, "%v", o.err)
				ev.Check("no invented successful text", text == "", "%q", text)
			}
			if kind != "request-eof" {
				ev.Check("all original partial thinking retained", thought == fixture.Thinking && thought != "", "%q", thought)
				ev.Check("body closed", transport.closed.Load(), "not closed")
			} else {
				ev.Check("request failure has no thinking", thought == "", "%q", thought)
			}
			miniMaxSharedOneTerminal(ev, o)

			w.record(ev, mixedOutcome{metadata: o.result.Metadata, usage: o.result.Message.Usage, chat: o.result, err: o.err})
			w.released(ev)
		})
	}
}

type miniMaxReadinessTransport struct {
	prefix, kind, wire string
	calls              atomic.Int64
	closed             atomic.Bool
}

func (r *miniMaxReadinessTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls.Add(1)
	b, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	r.wire = string(b)
	if r.kind == "request-eof" {
		return nil, io.ErrUnexpectedEOF
	}
	body := r.prefix
	if r.kind == "complete" {
		body += miniMaxReadinessEnding
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: &miniMaxReadinessBody{Reader: strings.NewReader(body), owner: r}, Request: req}, nil
}

type miniMaxReadinessBody struct {
	*strings.Reader
	owner *miniMaxReadinessTransport
}

func (b *miniMaxReadinessBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF && b.owner.kind == "read-error" {
		return n, errors.New("synthetic-read-cut")
	}
	return n, err
}
func (b *miniMaxReadinessBody) Close() error { b.owner.closed.Store(true); return nil }

const miniMaxReadinessEnding = `event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"synthetic completion"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}

`
