package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMSharedLimitsAndTimeouts(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, mode := range []string{"request", "frame", "error-body", "tool-json", "output", "queue-count", "queue-bytes", "admission", "connect-timeout", "header-timeout", "read-timeout", "call-timeout", "host-deadline"} {
				t.Run(r.name+"/"+model+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "GLM09-resource-"+t.Name())
					w := newGLMSharedWorld(t, func(c *ai.Config) {
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
					ev.Fixture(r.name+".json", raw)
					if mode == "admission" {
						w.adm.Deny(errors.New("synthetic-admission-denied"))
					}
					switch mode {
					case "error-body":
						reply = heldReply(500, "application/json", strings.Repeat("e", 1100))
					case "header-timeout":
						reply = provider.Reply{End: provider.EndSilent}
					case "read-timeout", "call-timeout", "host-deadline":
						reply = glmOwnReply(t, r.name, glmSharedBeforeTerminal(t, r.name))
						reply.End = provider.EndHold
					}
					ev.Record("response-script", reply)
					w.provider.EnqueueAt(r.prefix(glmA), reply)
					ctx := ctxFor(t)
					if mode == "host-deadline" {
						var cancel context.CancelFunc
						ctx, cancel = context.WithTimeout(ctx, short)
						defer cancel()
					}
					s := w.client.Stream(ctx, scopeFor(glmA, "glm09-resource-"+r.name+model+mode), r.target(model), req, r.options())
					defer s.Close()
					// Consumer does not drain until Result independently settles.
					settled := resultWithin(ev, "Result independently ends bounded call", s)
					o := drain(s)
					glmSharedRecord(ev, w, o)
					ev.Record("requests", w.provider.Requests())
					ev.Check("Result and terminal agree", errors.Is(settled.err, o.err) && reflect.DeepEqual(settled.res, o.result), "%v / %v", settled.err, o.err)
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
					ev.Check("own protocol respects Host resource bound", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}), "%v", o.err)
					want := 1
					if mode == "request" || mode == "admission" {
						want = 0
					}
					wantRequests := want
					if mode == "connect-timeout" {
						wantRequests = 0
					}
					ev.Check("only actual send becomes Attempt", len(o.result.Metadata.Attempts) == want && len(w.provider.Requests()) == wantRequests, "%+v", o.result.Metadata)
					glmSharedOneTerminal(ev, o)
					w.released(ev)
				})
			}
		}
	}
}

func glmSharedBeforeTerminal(t *testing.T, protocol string) []json.RawMessage {
	t.Helper()
	var events []string
	switch protocol {
	case "chat":
		events = []string{`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":3}}}`, `{"choices":[{"delta":{"content":"known-content"}}]}`}
	case "messages":
		events = []string{`{"type":"message_start","message":{"id":"glm09-message","usage":{"input_tokens":0,"output_tokens":0}}}`, `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`, `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"known-content"}}`, `{"type":"content_block_stop","index":0}`, `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":8,"output_tokens":7,"cache_read_input_tokens":4,"cache_creation_input_tokens":0}}`}
	case "responses":
		events = []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"msg09","role":"assistant","content":[]}}`, `{"type":"response.content_part.added","output_index":0,"content_index":0,"part":{"type":"output_text","text":""}}`, `{"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"known-content"}`}
	}
	out := make([]json.RawMessage, len(events))
	for i, e := range events {
		out[i] = json.RawMessage(e)
	}
	return out
}
func glmSharedOneTerminal(ev *evidence.Case, o outcome) {
	n := 0
	for _, e := range o.events {
		if e.Type() == "done" || e.Type() == "error" {
			n++
		}
	}
	ev.Check("exactly one terminal", n == 1, "%v", eventSummary(o.events))
}

// The design load is intentionally controlled, not a deployment default.
// 6 concurrent streams: two tenants × three protocols, 1 KiB prompts,
// 4 KiB maximum tool args, 32 KiB maximum wire output per stream. Each round
// uses an independent own-model protocol source plus synthetic text padding.
func TestGLMSharedDesignLoad(t *testing.T) {
	ev := pressureCase(t, "GLM09-design-load")
	w := newGLMSharedWorld(t, nil)
	baseline := runtime.NumGoroutine()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	var peakBytes, peakFrame, maxArgs, maxEvents, maxEventJSON int
	for round, model := range []string{"glm-5.3", "glm-5.2"} {
		held, release := make(chan struct{}, 6), make(chan struct{})
		var wg sync.WaitGroup
		out := make([]outcome, 6)
		for i, r := range glmSharedRoutes {
			for j, k := range []tenantKey{glmA, glmB} {
				req, first, _, raw := r.script(t, model)
				ev.Fixture(fmt.Sprintf("%s-%s.json", r.name, model), raw)
				ev.Record("response-"+r.name+model, first)
				req.Messages = append(req.Messages, ai.UserText(strings.Repeat("p", 1024)))
				frame, total := largestAndTotal(first.Chunks)
				peakFrame = max(peakFrame, frame)
				peakBytes = max(peakBytes, total)
				w.provider.EnqueueAt(r.prefix(k), first)
				wg.Go(func() {
					scope := scopeFor(k, fmt.Sprintf("glm09-load-%d-%s-%s", round, r.name, k.tenant))
					hooks := ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
						held <- struct{}{}
						select {
						case <-release:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}}
					out[i*2+j] = drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, r.target(model), req, r.options()))
				})
			}
		}
		for range 6 {
			waitValue(ev, "all six calls reach concurrent HTTP barrier", held)
		}
		ev.Record(fmt.Sprintf("peak-%d", round), map[string]int64{"permits": w.gauges.Permits(), "bodies": w.bodies.Open(), "hostPermits": int64(w.adm.Held())})
		ev.Check("six independent active leases", w.gauges.Permits() == 6 && w.bodies.Open() == 6 && w.adm.Held() == 6, "%d %d %d", w.gauges.Permits(), w.bodies.Open(), w.adm.Held())
		close(release)
		wg.Wait()
		for _, o := range out {
			glmSharedRecord(ev, w, o)
			ev.Check("design-load own protocol succeeds", o.err == nil, "%v", o.err)
			maxEvents = max(maxEvents, len(o.events))
			maxEventJSON = max(maxEventJSON, len(mustMarshal(t, eventsJSON(o.events))))
			for _, b := range o.result.Message.Content {
				if c, ok := b.(ai.ToolCall); ok {
					maxArgs = max(maxArgs, len(c.RawArguments))
				}
			}
		}
	}
	runtime.ReadMemStats(&after)
	ev.Record("design-load", map[string]any{"concurrent": 6, "rounds": 2, "models": []string{"glm-5.3", "glm-5.2"}, "promptBytes": 1024, "maxToolArgsBytes": 4096, "maxWireOutputBytes": 32768, "largestObservedWireBytes": peakBytes, "allocatedBytes": after.TotalAlloc - before.TotalAlloc, "heapAfter": after.HeapAlloc, "requestLimit": 1 << 20, "frameLimit": 1 << 16, "outputLimit": 1 << 20, "toolLimit": 1 << 16, "queueEvents": 1024, "queueBytes": 1 << 20, "processConcurrency": 32, "tenantConcurrency": 8, "observedWireOutputFraction": float64(peakBytes) / (1 << 20), "concurrencyFraction": 6.0 / 32})
	maxRequest := 0
	for _, q := range w.provider.Requests() {
		maxRequest = max(maxRequest, len(q.Body))
	}
	rooms := map[string]headroom{
		"MaxRequestBytes": room(int64(maxRequest), 1<<20), "MaxFrameBytes": room(int64(peakFrame), 1<<16),
		"MaxOutputBytes": room(int64(peakBytes), 1<<20), "MaxToolJSONBytes": room(int64(maxArgs), 1<<16),
		"MaxQueuedEventsUpperBound": room(int64(maxEvents), 1024), "MaxQueuedEventBytesSerializedUpperBound": room(int64(maxEventJSON), 1<<20),
		"MaxConcurrentPerTenant": room(3, 8), "MaxConcurrentProcess": room(6, 32),
		"MaxAdmissionWaiters": room(0, 64), "MaxImageBytes": room(0, 1<<19), "MaxErrorBodyBytes": room(0, 1<<14),
	}
	ev.Record("headroom", rooms)
	for name, h := range rooms {
		ev.Check("controlled load leaves headroom under "+name, h.Ratio <= 0.5, "%+v", h)
	}
	ev.Record("queue-byte-basis", "Serialized complete event snapshots are a conservative bound, not a measurement of internal queue allocation. The success load drains events; separate unread-queue cases enforce both queue limits.")
	ev.Record("requests", w.provider.Requests())
	w.released(ev)
	checkGoroutinesSettle(ev, baseline)
}
