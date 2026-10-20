package e2e

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Public Client failure matrix: a new identity must not bypass shared queue
// bounds, stream validation, cancellation or retry termination; every path
// settles once and returns bodies/permits without affecting a sibling service.
func TestRegionalProviderResources(t *testing.T) {
	for i, p := range regionalRuntimeProviders {
		for _, r := range miniMaxSharedRoutes {
			for _, mode := range []string{"queue-count", "queue-bytes", "malformed", "truncated", "cancel-before-headers", "cancel-after-content", "cancel-retry"} {
				t.Run(string(p)+"/"+r.name+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "RP16-resource-"+string(p)+"-"+r.name+"-"+mode)
					fake := clock.NewFake(time.Unix(1790000000, 0), 0)
					retryWait := fake.HoldSleeps()
					w := newRegionalRuntimeWorld(t, func(c *ai.Config) {
						c.Clock = fake
						if mode == "queue-count" {
							c.Policy.MaxQueuedEvents = 2
						}
						if mode == "queue-bytes" {
							c.Policy.MaxQueuedEventBytes = 4
						}
					})
					tenant := "regional-a"
					target := regionalRuntimeTarget(p, r.api)
					b := w.host.Binding(tenant, target.BindingID)
					b.Retry = ai.RetryPolicy{MaxRetries: 2}
					w.host.PutBinding(b)
					reply := regionalRuntimeReply(t, r.api)
					ready := make(chan struct{})
					switch mode {
					case "queue-count", "queue-bytes":
						reply = miniMaxSharedBeforeTerminal(t, r, mixedModel)
					case "malformed":
						reply = provider.SSE([][]byte{[]byte("event: message_start\ndata: {malformed-json}\n\n")})
					case "truncated":
						reply = miniMaxSharedBeforeTerminal(t, r, mixedModel)
					case "cancel-before-headers":
						reply = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
					case "cancel-after-content":
						reply = miniMaxSharedBeforeTerminal(t, r, mixedModel)
						reply.End = provider.EndHold
					case "cancel-retry":
						reply = provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic temporary failure"}}`)}, Script: `{"error":{"message":"synthetic temporary failure"}}`}
					}
					ev.Record("response-script", reply)
					w.provider.EnqueueAt(regionalRuntimePrefix(tenant, p, r.api), reply)
					ctx, cancel := context.WithCancel(ctxFor(t))
					defer cancel()
					id := "regional-resource-" + string(p) + "-" + r.name + "-" + mode
					s := w.client.Stream(ctx, ai.CallScope{TenantID: tenant, RequestID: id}, target, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic resource input.")}}, nil)
					defer s.Close()
					var prefix []ai.EventEnvelope
					switch mode {
					case "cancel-before-headers":
						waitFor(ev, "request reaches Provider before headers", ready)
						cancel()
					case "cancel-after-content":
						prefix = within(ev, "parsed visible content precedes cancellation", func() []ai.EventEnvelope {
							var events []ai.EventEnvelope
							for s.Next() {
								events = append(events, s.Envelope())
								if s.Event().Type() == "text_delta" {
									return events
								}
							}
							return events
						})
						ev.Check("visible content reached consumer", len(prefix) > 0 && prefix[len(prefix)-1].Event.Type() == "text_delta", "content absent")
						cancel()
					case "cancel-retry":
						d := waitValue(ev, "first attempt reaches retry backoff", retryWait)
						ev.Record("retry-wait", d.String())
						ev.Check("attempt releases permits before backoff", w.gauges.Permits() == 0 && w.adm.Held() == 0 && w.bodies.Open() == 0, "resources retained during retry")
						cancel()
					}
					settled := resultWithin(ev, "Result settles independently of consumer", s)
					envs, result, err := drainEnvelopes(ev, s)
					envs = append(prefix, envs...)
					o := outcome{result: result, err: err, streamed: true}
					for _, env := range envs {
						o.events = append(o.events, env.Event)
					}
					regionalRuntimeRecord(ev, w, o)
					ev.Record(id+"-envelopes", envelopesJSON(envs))
					ev.Check("Result remains immutable and settles once", settled.err == o.err && reflect.DeepEqual(settled.res, o.result), "Result changed")
					switch mode {
					case "queue-count", "queue-bytes":
						ev.Check("event budget bounds every identity", errors.Is(o.err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseEventQueue}), "%v", o.err)
					case "malformed", "truncated":
						ev.Check("invalid stream is terminal", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
					default:
						ev.Check("cancellation terminates attempt or backoff", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}), "%v", o.err)
					}
					ev.Check("one attempt and no replay after stream or cancel", len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%+v", o.result.Metadata)
					miniMaxSharedOneTerminal(ev, o)
					regionalRuntimeAdmitted(ev, w, tenant, p, id, 1)
					for _, env := range envs {
						ev.Check("terminal and partial events retain real identity", env.Call == o.result.Metadata.CallAttribution, "%+v", env.Call)
					}
					// Complete does not queue events. A distinct tenant/service can keep
					// operating even after the subject stream exceeds its queue budget.
					sibling := regionalRuntimeProviders[(i+1)%len(regionalRuntimeProviders)]
					siblingTenant := "regional-b"
					good := regionalRuntimeReply(t, r.api)
					ev.Record("sibling-script", good)
					w.provider.EnqueueAt(regionalRuntimePrefix(siblingTenant, sibling, r.api), good)
					res, e := w.client.Complete(ctxFor(t), ai.CallScope{TenantID: siblingTenant, RequestID: id + "-sibling"}, regionalRuntimeTarget(sibling, r.api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic unaffected sibling.")}}, nil)
					regionalRuntimeRecord(ev, w, outcome{result: res, err: e})
					ev.Check("sibling tenant and service remain usable", e == nil && res.Metadata.ProviderID == sibling && res.Metadata.TenantID == siblingTenant, "%v %+v", e, res.Metadata)
					w.released(ev)
				})
			}
		}
	}
}
