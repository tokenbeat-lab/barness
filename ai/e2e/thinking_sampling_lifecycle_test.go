package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// S26/S27 exercise final payload boundaries with active profiles on all four
// routes and entries. A hook's reasoning rewrite cannot select a second profile.
func TestThinkingSamplingCallbacks(t *testing.T) {
	for _, p := range thinkingSamplingProtocols {
		f, raw := loadFixture(t, p.proto, "thinking-sampling.json")
		for _, sc := range f.Scenarios[:2] {
			for _, mode := range []callMode{modeStream, modeComplete} {
				for _, variant := range []string{"in-place", "replace", "model", "hosted", "limit"} {
					if variant == "hosted" && p.proto.api != ai.APIOpenAIResponses {
						continue
					}
					t.Run(p.proto.dir+"/"+sc.entry()+"/"+string(mode)+"/"+variant, func(t *testing.T) {
						ev := run.Case(t, p.prefix+"-E04-thinking-sampling-callback-"+sc.entry()+"-"+string(mode)+"-"+variant)
						ev.Fixture("fixture.json", raw)
						w := scenarioWorld(t, sc)
						enqueue(ev, w, sc.replies(t)...)
						calls := 0
						hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
							calls++
							ev.Check("hook sees merged payload", jsonEqual(mustMarshal(t, payload.Body), sc.Expect.Request.Body), "got %s", mustMarshal(t, payload.Body))
							body := payload.Body
							if variant == "replace" {
								body = nil
								mustUnmarshal(t, sc.Expect.Request.Body, &body)
							}
							body["temperature"] = json.Number("0.125")
							if p.proto.api == ai.APIOpenAIResponses {
								body["reasoning"] = map[string]any{"effort": "low"}
							} else {
								body["reasoning_effort"] = "low"
							}
							switch variant {
							case "model":
								body["model"] = "other-tenant-model"
							case "hosted":
								body["previous_response_id"] = "other-tenant-response"
							case "limit":
								body["extra"] = strings.Repeat("x", int(validPolicy().MaxRequestBytes)+1)
							case "replace":
								payload.Body["temperature"] = 2
								return ai.ReplacePayload(body), nil
							}
							return ai.KeepPayload(), nil
						}}
						o := invokeThinkingSamplingHooks(t, w, sc, mode, hooks)
						o.record(ev)
						ev.Check("hook runs once", calls == 1, "count=%d", calls)
						if variant == "model" || variant == "hosted" || variant == "limit" {
							code := ai.CodeTenantDenied
							if variant == "limit" {
								code = ai.CodeResourceLimit
							}
							checkRejected(ev, o, code, ai.PhaseRequest)
							ev.Check("zero sends/attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "attempts=%+v", o.result.Metadata.Attempts)
						} else {
							ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
							fields := map[string]any{"temperature": 0.125, "reasoning_effort": "low"}
							if p.proto.api == ai.APIOpenAIResponses {
								fields = map[string]any{"temperature": 0.125, "reasoning": map[string]any{"effort": "low"}}
							}
							checkSentBody(ev, w, withFields(t, sc.Expect.Request.Body, fields))
						}
						checkNoSecrets(ev, o)
					})
				}
			}
		}
	}
}

// S28/S30 pin merged bytes once. Mutations happen after Provider receipt,
// never concurrently with constructor reads. Cancellation ends a held backoff.
func TestThinkingSamplingRetry(t *testing.T) {
	for _, p := range thinkingSamplingProtocols {
		f, raw := loadFixture(t, p.proto, "thinking-sampling.json")
		for _, sc := range f.Scenarios[:2] {
			for _, mode := range []callMode{modeStream, modeComplete} {
				for _, variant := range []string{"retry", "cancel-backoff", "cancel-before-send", "send-failure"} {
					t.Run(p.proto.dir+"/"+sc.entry()+"/"+string(mode)+"/"+variant, func(t *testing.T) {
						ev := run.Case(t, p.prefix+"-E05-thinking-sampling-"+sc.entry()+"-"+string(mode)+"-"+variant)
						ev.Fixture("fixture.json", raw)
						catalog := ai.BuiltinCatalog()
						patch := withModelPatch(sc.Model, nil, sc.ModelPatch)
						clk := clock.NewFake(time.Unix(0, 0), 0)
						bodies := provider.TrackBodies(provider.LoopbackTransport())
						w := newWorldWith(t, func(c *ai.Config) {
							c.Catalog = &catalog
							patch(c)
							catalog = *c.Catalog
							c.Transport, c.Clock = bodies, clk
						}, tenantA)
						w.updateBindingOf(tenantA, p.proto.binding, func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
						ctx, cancel := context.WithCancel(ctxFor(t))
						defer cancel()
						var held <-chan time.Duration
						if variant == "cancel-backoff" {
							held = clk.HoldSleeps()
						}
						if variant == "cancel-before-send" {
							cancel()
						}
						failure := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{}`)}, Script: `{}`}
						failure.OnReceive = func() {
							for i := range catalog.Models {
								if catalog.Models[i].ID == sc.Model && catalog.Models[i].API == p.proto.api {
									catalog.Models[i].SamplingParamsByThinkingLevel[ai.ThinkingHigh]["top_k"] = json.RawMessage(`999`)
								}
							}
						}
						if variant == "send-failure" {
							failure.End = provider.EndDrop
							enqueue(ev, w, failure, failure)
						} else {
							enqueue(ev, w, append([]provider.Reply{failure}, sc.replies(t)...)...)
						}
						calls := 0
						hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
							calls++
							return ai.KeepPayload(), nil
						}}
						done := make(chan outcome, 1)
						go func() { done <- invokeThinkingSamplingAt(t, ctx, w, sc, mode, hooks, textScope("req-sampling-retry")) }()
						if held != nil {
							waitValue(ev, "backoff started", held)
							cancel()
						}
						o := waitValue(ev, "bounded terminal", done)
						o.record(ev)
						requests := w.provider.Requests()
						ev.Record("requests", requests)
						want := 2
						if variant == "cancel-backoff" {
							want = 1
						}
						if variant == "cancel-before-send" {
							want = 0
						}
						ev.Check("bounded attempts", len(requests) == want && len(o.result.Metadata.Attempts) == want, "requests=%d attempts=%d want=%d", len(requests), len(o.result.Metadata.Attempts), want)
						if want > 0 {
							ev.Check("hook once", calls == 1, "count=%d", calls)
						}
						for i, req := range requests {
							ev.Check("profile frozen", jsonEqual(req.Body, sc.Expect.Request.Body), "request %d: %s", i, req.Body)
							if i > 0 {
								ev.Check("exact retry bytes", bytes.Equal(req.Body, requests[0].Body), "bytes changed")
							}
						}
						if variant == "retry" {
							ev.Check("success", o.err == nil, "err=%v", o.err)
						} else {
							ev.Check("failure", o.err != nil, "expected failure")
						}
						ev.Check("closed response bodies", bodies.Open() == 0, "open=%d", bodies.Open())
					})
				}
			}
		}
	}
}

// S29 synchronizes calls at OnPayload, so both tenants' options are in flight
// together. Same-ID clients with distinct profiles additionally prove isolation.
func TestThinkingSamplingConcurrent(t *testing.T) {
	for _, p := range thinkingSamplingProtocols {
		f, raw := loadFixture(t, p.proto, "thinking-sampling.json")
		for _, separate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/separate-%t", p.proto.dir, separate), func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("%s-E06-thinking-sampling-concurrent-%t", p.prefix, separate))
				ev.Fixture("fixture.json", raw)
				worlds := make([]*world, 2)
				for i := range 2 {
					sc := f.Scenarios[0]
					if separate {
						sc.ModelPatch = withFields(t, sc.ModelPatch, map[string]any{"samplingParamsByThinkingLevel": map[string]any{"high": map[string]any{"top_k": 20 + i}}})
					}
					worlds[i] = newWorldWith(t, withModelPatch(sc.Model, nil, sc.ModelPatch), tenantA, tenantB)
				}
				if !separate {
					worlds[1] = worlds[0]
				}
				for _, w := range worlds {
					enqueue(ev, w, f.Scenarios[0].replies(t)...)
				}
				ready, release, done := make(chan struct{}, 2), make(chan struct{}), make(chan outcome, 2)
				for i, tenant := range []tenantKey{tenantA, tenantB} {
					sc := f.Scenarios[i]
					var opts map[string]any
					mustUnmarshal(t, sc.Options, &opts)
					opts["samplingParams"].(map[string]any)["temperature"] = float64(i+1) / 10
					sc.Options = mustMarshal(t, opts)
					scope := textScope(fmt.Sprintf("req-concurrent-%d", i))
					scope.TenantID = tenant.tenant
					hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
						ready <- struct{}{}
						<-release
						payload.Body["extra"] = scope.TenantID
						return ai.KeepPayload(), nil
					}}
					go func() { done <- invokeThinkingSamplingAt(t, ctxFor(t), worlds[i], sc, modeComplete, hooks, scope) }()
				}
				for range 2 {
					waitValue(ev, "calls at barrier", ready)
				}
				close(release)
				for range 2 {
					o := waitValue(ev, "concurrent terminal", done)
					o.record(ev)
					ev.Check("success", o.err == nil, "err=%v", o.err)
				}
				for i, w := range worlds {
					for _, req := range w.provider.Requests() {
						var body map[string]any
						mustUnmarshal(t, req.Body, &body)
						idx := 0
						if req.KeyAlias == p.proto.key(tenantB).alias {
							idx = 1
						}
						ev.Check("tenant ownership", req.KeyAlias == p.proto.key([]tenantKey{tenantA, tenantB}[idx]).alias && body["temperature"] == float64(idx+1)/10 && body["extra"] == []tenantKey{tenantA, tenantB}[idx].tenant, "body=%s key=%s", req.Body, req.KeyAlias)
						marker := float64(20)
						if separate {
							marker += float64(i)
						}
						ev.Check("profile ownership", body["top_k"] == marker, "body=%s", req.Body)
					}
					ev.Record(fmt.Sprintf("client-%d", i), w.client.Catalog())
					ev.Record(fmt.Sprintf("requests-%d", i), w.provider.Requests())
				}
			})
		}
	}
}
