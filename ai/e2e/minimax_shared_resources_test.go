package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"reflect"
	"strings"
	"testing"
)

func miniMaxSharedBeforeTerminal(t *testing.T, r miniMaxSharedRoute, models ...string) provider.Reply {
	t.Helper()
	var events []string
	switch r.name {
	case "chat":
		events = []string{`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":3}}}`, `{"choices":[{"delta":{"content":"known-content"}}]}`}
	case "messages":
		events = []string{`{"type":"message_start","message":{"id":"mm05-partial","model":"MiniMax-M3","role":"assistant","content":[],"usage":{"input_tokens":8,"cache_read_input_tokens":4,"cache_creation_input_tokens":0,"output_tokens":0}}}`, `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`, `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"known-content"}}`, `{"type":"content_block_stop","index":0}`, `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":7}}`}
	default:
		events = []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"msg05","role":"assistant","content":[]}}`, `{"type":"response.content_part.added","output_index":0,"content_index":0,"part":{"type":"output_text","text":""}}`, `{"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"known-content"}`}
	}
	raw := make([]json.RawMessage, len(events))
	for i, e := range events {
		if len(models) > 0 {
			e = strings.ReplaceAll(e, "MiniMax-M3", models[0])
		}
		raw[i] = json.RawMessage(e)
	}
	if r.name == "chat" {
		return chatEvents(t, raw, "")
	}
	return sseEvents(t, raw, provider.FramingLF)
}

func TestMiniMaxSharedResources(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
			for _, mode := range []string{"request", "frame", "error-body", "tool-json", "output", "queue-count", "queue-bytes", "admission", "connect-timeout", "header-timeout", "read-timeout", "call-timeout", "host-deadline"} {
				t.Run(r.name+"/"+model+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "MM05-resource-"+t.Name())
					w := newMiniMaxSharedWorld(t, func(c *ai.Config) {
						switch mode {
						case "request":
							setRequestBytes(c.Policy, 128)
						case "frame":
							setFrameBytes(c.Policy, 128)
						case "error-body":
							setErrorBodyBytes(c.Policy, 1024)
						case "tool-json":
							setToolJSONBytes(c.Policy, 4)
						case "output":
							setOutputBytes(c.Policy, 128)
						case "queue-count":
							c.Policy.MaxQueuedEvents = 2
						case "queue-bytes":
							c.Policy.MaxQueuedEventBytes = 4
						case "connect-timeout":
							c.Policy.ConnectTimeout = short
							c.Transport = provider.StalledDialTransport()
						case "header-timeout":
							c.Policy.ResponseHeaderTimeout = short
						case "read-timeout":
							c.Policy.ReadIdleTimeout = short
						case "call-timeout":
							fitCallTimeout(c.Policy, short)
						}
					})
					req, reply, _, raw := r.script(t, model)
					ev.Fixture("synthetic-source.json", raw)
					if mode == "admission" {
						w.adm.Deny(errors.New("synthetic-admission-denied"))
					}
					switch mode {
					case "error-body":
						reply = heldReply(500, "application/json", strings.Repeat("e", 1100))
					case "header-timeout":
						reply = provider.Reply{End: provider.EndSilent}
					case "read-timeout", "call-timeout", "host-deadline":
						reply = miniMaxSharedBeforeTerminal(t, r, model)
						reply.End = provider.EndHold
					}
					ev.Record("response-script", reply)
					w.provider.EnqueueAt(r.prefix(miniMaxSharedA), reply)
					ctx := ctxFor(t)
					if mode == "host-deadline" {
						var cancel context.CancelFunc
						ctx, cancel = context.WithTimeout(ctx, short)
						defer cancel()
					}
					s := w.client.Stream(ctx, scopeFor(miniMaxSharedA, "mm05-resource-"+r.name+model+mode), r.target(model), req, r.options())
					defer s.Close()
					settled := resultWithin(ev, "Result settles without a consumer", s)
					o := drain(s)
					miniMaxSharedRecord(ev, w, o)
					ev.Record("requests", w.provider.Requests())
					ev.Check("Result remains immutable", errors.Is(settled.err, o.err) && reflect.DeepEqual(settled.res, o.result), "%v / %v", settled.err, o.err)
					code, phase := ai.CodeResourceLimit, ai.PhaseStream
					switch mode {
					case "request", "error-body":
						phase = ai.PhaseRequest
					case "queue-count", "queue-bytes":
						phase = ai.PhaseEventQueue
					case "admission":
						code, phase = ai.CodeAdmissionDenied, ai.PhaseAdmission
					case "connect-timeout", "header-timeout":
						code, phase = ai.CodeDeadlineExceeded, ai.PhaseRequest
					case "read-timeout", "call-timeout", "host-deadline":
						code, phase = ai.CodeDeadlineExceeded, ai.PhaseStream
					}
					ev.Check("Host bound applies to own protocol", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}), "%v", o.err)
					want := 1
					if mode == "request" || mode == "admission" {
						want = 0
					}
					wantRequests := want
					if mode == "connect-timeout" {
						wantRequests = 0
					}
					ev.Check("only actual send becomes Attempt", len(o.result.Metadata.Attempts) == want && len(w.provider.Requests()) == wantRequests, "%+v", o.result.Metadata)
					miniMaxSharedOneTerminal(ev, o)
					w.released(ev)
				})
			}
		}
	}
}

func TestMiniMaxSharedCancellation(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, mode := range []string{"before-send", "before-headers", "after-content"} {
			t.Run(r.name+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM05-cancel-"+r.name+mode)
				ready := make(chan struct{})
				reply := miniMaxSharedBeforeTerminal(t, r)
				reply.End = provider.EndHold
				_, bytes := largestAndTotal(reply.Chunks)
				// This existing external HTTP read barrier is vendor-neutral; its explicit
				// prefix/byte count proves parsed-prefix cancellation for our own fixture.
				w := newMiniMaxSharedWorld(t, func(c *ai.Config) {
					if mode == "after-content" {
						c.Transport = &glmReadBarrier{next: c.Transport, prefix: r.prefix(miniMaxSharedA), bytes: bytes, ready: ready}
					}
				})
				if mode == "before-headers" {
					reply = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
				}
				ev.Record("response-script", reply)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedA), reply)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				if mode == "before-send" {
					cancel()
				}
				s := w.client.Stream(ctx, scopeFor(miniMaxSharedA, "mm05-cancel-"+r.name+mode), r.target("MiniMax-M3"), miniMaxMessagesRequest(), r.options())
				defer s.Close()
				if mode != "before-send" {
					waitFor(ev, "subject reaches cancellation boundary", ready)
				}
				req, first, last, raw := r.script(t, "MiniMax-M3")
				ev.Fixture("sibling-source.json", raw)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedB), first, last)
				scope := scopeFor(miniMaxSharedB, "mm05-sibling-"+r.name+mode)
				a, err := w.client.Complete(ctxFor(t), scope, r.target("MiniMax-M3"), req, r.options())
				ev.Record("sibling-tools", a)
				if ev.Check("sibling tool generation survives", err == nil, "%v", err) {
					scope.RequestID += "-continue"
					b, e := w.client.Complete(ctxFor(t), scope, r.target("MiniMax-M3"), miniMaxSharedContinue(req, a.Message), r.options())
					ev.Record("sibling-final", b)
					ev.Check("sibling round trip completes", e == nil, "%v", e)
				}
				cancel()
				settled := resultWithin(ev, "canceled Result independent of consumer", s)
				o := drain(s)
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				ev.Check("cancellation classified once", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}) && settled.err == o.err, "%v", o.err)
				if mode == "after-content" {
					ev.Check("parsed content retained", textOf(o.result.Message) == "known-content", "%+v", o.result.Message)
					if r.name != "responses" {
						ev.Check("known interrupted usage retained", o.result.Message.Usage.Input == 8 && o.result.Message.Usage.Output == 7 && o.result.Message.Usage.CacheRead == 4, "%+v", o.result.Message.Usage)
					}
				}
				miniMaxSharedOneTerminal(ev, o)
				w.released(ev)
			})
		}
	}
}
