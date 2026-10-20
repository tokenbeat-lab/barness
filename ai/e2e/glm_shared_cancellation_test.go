package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// At the next HTTP Read after the complete scripted prefix has been handed
// back, the synchronous SSE parser has processed those complete frames. This
// external barrier removes the race between a server flush and cancellation.
// It retains no body copy and wraps only the subject's public HTTP response.
type glmReadBarrier struct {
	next   http.RoundTripper
	prefix string
	bytes  int
	ready  chan struct{}
}

func (b *glmReadBarrier) RoundTrip(q *http.Request) (*http.Response, error) {
	res, err := b.next.RoundTrip(q)
	if err == nil && res.Body != nil && strings.HasPrefix(q.URL.Path, b.prefix) {
		res.Body = &glmReadBarrierBody{ReadCloser: res.Body, barrier: b}
	}
	return res, err
}

type glmReadBarrierBody struct {
	io.ReadCloser
	barrier *glmReadBarrier
	read    int
	once    sync.Once
}

func (b *glmReadBarrierBody) Read(p []byte) (int, error) {
	if b.read >= b.barrier.bytes {
		b.once.Do(func() { close(b.barrier.ready) })
	}
	n, err := b.ReadCloser.Read(p)
	b.read += n
	return n, err
}

func TestGLMSharedCancellation(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, mode := range []string{"before-send", "before-headers", "after-content"} {
				t.Run(r.name+"/"+model+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "GLM09-cancel-"+t.Name())
					ready := make(chan struct{})
					reply := glmOwnReply(t, r.name, glmSharedBeforeTerminal(t, r.name))
					reply.End = provider.EndHold
					_, bytes := largestAndTotal(reply.Chunks)
					w := newGLMSharedWorld(t, func(c *ai.Config) {
						if mode == "after-content" {
							c.Transport = &glmReadBarrier{next: c.Transport, prefix: r.prefix(glmA), bytes: bytes, ready: ready}
						}
					})
					if mode == "before-headers" {
						reply = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
					}
					ev.Record("subject-response-script", reply)
					w.provider.EnqueueAt(r.prefix(glmA), reply)
					ctx, cancel := context.WithCancel(ctxFor(t))
					defer cancel()
					if mode == "before-send" {
						cancel()
					}
					subject := w.client.Stream(ctx, scopeFor(glmA, "glm09-cancel-"+r.name+model+mode), r.target(model), ai.Request{Messages: []ai.Message{ai.UserText("synthetic-cancel-subject")}}, r.options())
					defer subject.Close()
					if mode != "before-send" {
						waitFor(ev, "subject reaches exact cancellation boundary", ready)
						ev.Check("subject owns its HTTP lease", w.adm.Held() == 1, "%d", w.adm.Held())
					}
					// A different tenant, with identical names, completes its own real tool
					// round trip while the subject still owns its connection/permit.
					req, first, last, raw := r.script(t, model)
					ev.Fixture(r.name+".json", raw)
					ev.Record("sibling-responses", []any{first, last})
					w.provider.EnqueueAt(r.prefix(glmB), first, last)
					siblingScope := scopeFor(glmB, "glm09-cancel-sibling-"+r.name+model+mode)
					a, err := w.client.Complete(ctxFor(t), siblingScope, r.target(model), req, r.options())
					ev.Record("sibling-tool-result", a)
					ev.Check("sibling actual tool generation succeeds", err == nil && a.Message.StopReason == ai.StopReasonToolUse, "%v", err)
					siblingScope.RequestID += "-continue"
					b, err := w.client.Complete(ctxFor(t), siblingScope, r.target(model), glmSharedContinue(req, a.Message), r.options())
					ev.Record("sibling-final", b)
					ev.Check("sibling independent continuation succeeds", err == nil && b.Metadata.TenantID == glmB.tenant, "%v", err)
					cancel()
					settled := resultWithin(ev, "canceled Result does not require a consumer", subject)
					o := drain(subject)
					glmSharedRecord(ev, w, o)
					ev.Record("requests", w.provider.Requests())
					ev.Check("cancellation remains classified and stable", errors.Is(settled.err, &ai.Error{Code: ai.CodeCanceled}) && settled.err == o.err, "%v", o.err)
					want := 1
					if mode == "before-send" {
						want = 0
					}
					ev.Check("only actual subject attempt is retained", len(o.result.Metadata.Attempts) == want, "%+v", o.result.Metadata)
					if mode == "after-content" {
						ev.Check("partial visible content survives cancellation", textOf(o.result.Message) == "known-content", "%+v", o.result.Message)
						u := o.result.Message.Usage
						if r.name == "responses" {
							ev.Check("Responses has no usage before legal terminal", o.result.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported && u.TotalTokens == 0, "%+v", u)
						} else {
							ev.Check("legally reported known usage survives local cancellation", u.Input == 8 && u.Output == 7 && u.CacheRead == 4 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", u)
						}
					}
					glmSharedOneTerminal(ev, o)
					w.released(ev)
				})
			}
		}
	}
}

func TestGLMSharedResponsesSettlement(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, mode := range []string{"terminal-awaiting-eof", "already-settled"} {
			t.Run(model+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "GLM09-settlement-"+t.Name())
				r := glmSharedRoutes[1]
				ready := make(chan struct{})
				reply := glmOwnReply(t, r.name, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[],"usage":{"input_tokens":12,"output_tokens":7,"total_tokens":19,"input_tokens_details":{"cached_tokens":4},"output_tokens_details":{"reasoning_tokens":3}}}}`)})
				_, bytes := largestAndTotal(reply.Chunks)
				w := newGLMSharedWorld(t, func(c *ai.Config) {
					if mode == "terminal-awaiting-eof" {
						c.Transport = &glmReadBarrier{next: c.Transport, prefix: r.prefix(glmA), bytes: bytes, ready: ready}
					}
				})
				if mode == "terminal-awaiting-eof" {
					reply.End = provider.EndHold
				}
				ev.Record("subject-response-script", reply)
				w.provider.EnqueueAt(r.prefix(glmA), reply)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				s := w.client.Stream(ctx, scopeFor(glmA, "glm09-settlement-"+model+mode), r.target(model), glmResponsesRequest(), r.options())
				defer s.Close()
				if mode == "terminal-awaiting-eof" {
					waitFor(ev, "legal terminal parsed while EOF pending", ready)
					cancel()
				}
				settled := resultWithin(ev, "Result settles once", s)
				cancel()
				again := resultWithin(ev, "Result stays settled after context cancellation", s)
				o := drain(s)
				glmSharedRecord(ev, w, o)
				if mode == "terminal-awaiting-eof" {
					ev.Check("interrupted transport retains known terminal usage", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}), "%v", o.err)
				} else {
					ev.Check("already settled success stays successful", o.err == nil && again.err == nil, "%v", o.err)
				}
				ev.Check("single immutable settlement", reflect.DeepEqual(settled.res, again.res) && settled.err == again.err && o.result.Message.Usage.Input == 8 && o.result.Message.Usage.Output == 7 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", o.result)
				ev.Record("requests", w.provider.Requests())
				glmSharedOneTerminal(ev, o)
				w.released(ev)
			})
		}
	}
}
