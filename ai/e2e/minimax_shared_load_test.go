package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func miniMaxSharedLoadReply(t *testing.T, r miniMaxSharedRoute, size int) provider.Reply {
	t.Helper()
	text := strings.Repeat("x", size)
	switch r.name {
	case "chat":
		return chatEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"choices":[{"delta":{"content":%q},"finish_reason":"stop"}],"usage":{"prompt_tokens":32768,"completion_tokens":4096,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":0}}}`, text))}, "")
	case "responses":
		return minimaxResponsesReply(t, fmt.Sprintf(`{"type":"response.completed","response":{"id":"mm05-load","model":"MiniMax-M3","status":"completed","output":[{"type":"message","id":"msg-load","role":"assistant","content":[{"type":"output_text","text":%q}]}],"usage":{"input_tokens":32768,"output_tokens":4096,"input_tokens_details":{"cached_tokens":4},"output_tokens_details":{"reasoning_tokens":0}}}}`, text))
	default:
		return provider.SSE([][]byte{[]byte(fmt.Sprintf("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"mm05-load\",\"model\":\"MiniMax-M3\",\"role\":\"assistant\",\"content\":[],\"usage\":{\"input_tokens\":32764,\"cache_read_input_tokens\":4,\"cache_creation_input_tokens\":0,\"output_tokens\":0}}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":%q}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":4096}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n", text))})
	}
}

// Existing TestMixedPressure still runs the complete original4/8 saturation
// batches unchanged. These are additional3/6-call batches: two original routes
// and one MiniMax route per tenant, rotating to cover every original route and
// every MiniMax protocol. All original sizes/policies/memory ceilings remain.
func TestMiniMaxSharedDesignLoad(t *testing.T) {
	for _, p := range []struct {
		name    string
		policy  *ai.ResourcePolicy
		tenants int
		budget  uint64
	}{{"local", localassembly.MixedPolicy(), 1, 256 << 20}, {"cloud", hostintegration.CloudMixedPolicy(), 2, 512 << 20}} {
		for phase, mini := range miniMaxSharedRoutes {
			t.Run(p.name+"/"+mini.name, func(t *testing.T) {
				ev := pressureCase(t, "MM05-design-load-"+p.name+"-"+mini.name)
				l := mixedDesign(p.policy.MaxConcurrentProcess, p.tenants)
				held, release := make(chan struct{}, 3*p.tenants), make(chan struct{})
				var once sync.Once
				unblock := func() { once.Do(func() { close(release) }) }
				defer unblock()
				w := newMiniMaxSharedWorld(t, func(c *ai.Config) { c.Policy = p.policy })
				old := []mixedRoute{mixedRoutes[phase%4], mixedRoutes[(phase+1)%4]}
				ctx, cancel := context.WithTimeout(ctxFor(t), pressureDeadline)
				defer cancel()
				hooks := ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					held <- struct{}{}
					select {
					case <-release:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}}
				ref, mask := mixedPNG(t, l.InputImageBytes), mixedPNG(t, l.MaskBytes)
				imageReq := ai.ImagesRequest{Prompt: "synthetic design edit", ReferenceImages: []ai.Image{ref, ref}}
				classifierReq := mixedLoadClassifier(l)
				chatReq := ai.Request{Messages: []ai.Message{ai.UserText(strings.Repeat("h", l.HistoryBytes))}}
				replies := map[string]provider.Reply{mini.name: miniMaxSharedLoadReply(t, mini, l.ChatOutputBytes)}
				for _, r := range old {
					replies["old-"+r.id] = mixedLoadReply(t, r, l)
				}
				for _, k := range []tenantKey{tenantA, tenantB}[:p.tenants] {
					for _, r := range old {
						w.provider.EnqueueAt(r.prefix(k), replies["old-"+r.id])
					}
					mk := miniMaxSharedA
					if k.tenant == tenantB.tenant {
						mk = miniMaxSharedB
					}
					w.provider.EnqueueAt(mini.prefix(mk), replies[mini.name])
				}
				heap := startHeapSampler(w.gauges.Permits)
				defer heap.finish()
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				results := make([]mixedOutcome, 3*p.tenants)
				var wg sync.WaitGroup
				defer wg.Wait()
				defer cancel()
				defer unblock()
				for n, k := range []tenantKey{tenantA, tenantB}[:p.tenants] {
					for i, r := range old {
						wg.Go(func() {
							scope := scopeFor(k, fmt.Sprintf("mm05-load-%s-%s-old-%s-%s", p.name, mini.name, r.id, k.tenant))
							target := ai.Target{BindingID: r.id, ModelID: mixedModel}
							h := w.client.WithHooks(hooks)
							var o mixedOutcome
							switch r.op {
							case ai.OperationChat:
								res, e := h.Complete(ctx, scope, target, chatReq, nil)
								o = mixedOutcome{metadata: res.Metadata, chat: res, usage: res.Message.Usage, err: e}
							case ai.OperationClassifier:
								res, e := h.Classify(ctx, scope, target, classifierReq, nil)
								o = mixedOutcome{metadata: res.Metadata, classifier: res, usage: res.Usage, err: e}
							default:
								var opts ai.ImageOptions
								if r.provider == ai.ProviderOpenAI {
									opts = ai.OpenAIImagesOptions{N: ai.Value(l.OutputImages), Mask: &mask}
								}
								res, e := h.GenerateImages(ctx, scope, target, imageReq, opts)
								o = mixedOutcome{metadata: res.Metadata, images: res, usage: res.Usage, err: e}
							}
							results[n*3+i] = o
						})
					}
					mk := miniMaxSharedA
					if k.tenant == tenantB.tenant {
						mk = miniMaxSharedB
					}
					wg.Go(func() {
						res, e := w.client.WithHooks(hooks).Complete(ctx, scopeFor(mk, "mm05-load-"+p.name+mini.name+mk.tenant), mini.target("MiniMax-M3"), chatReq, mini.options())
						results[n*3+2] = mixedOutcome{metadata: res.Metadata, chat: res, usage: res.Message.Usage, err: e}
					})
				}
				for range 3 * p.tenants {
					waitForWithin(ev, "every old and MiniMax call holds its HTTP lease", pressureDeadline, held)
				}
				ev.Check("additional batch holds exact concurrent leases", w.gauges.Permits() == int64(3*p.tenants) && w.bodies.Open() == int64(3*p.tenants) && w.adm.Held() == 3*p.tenants, "%d %d %d", w.gauges.Permits(), w.bodies.Open(), w.adm.Held())
				heap.peakGauge.Store(max(heap.peakGauge.Load(), w.gauges.Permits()))
				runtime.GC()
				heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
				unblock()
				waitForWithin(ev, "whole added batch completes", pressureDeadline, doneWhen(wg.Wait))
				runtime.GC()
				heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
				runtime.ReadMemStats(&after)
				use := heap.finish()
				var largestRequest int64
				for _, q := range w.provider.Requests() {
					largestRequest = max(largestRequest, int64(len(q.Body)))
				}
				for _, o := range results {
					ev.Check("original and MiniMax workload succeeds", o.err == nil, "%s %v", o.metadata.BindingID, o.err)
					if o.metadata.Operation == ai.OperationImage {
						ev.Check("all original image outputs", len(o.images.Content) == l.OutputImages, "%+v", o.images)
					}
					if o.metadata.Operation == ai.OperationClassifier {
						ev.Check("all original questions", len(o.classifier.Answers) == l.Questions, "%+v", o.classifier)
					}
					if o.metadata.ProviderID == ai.ProviderMiniMaxCN {
						ev.Check("MiniMax whole16KiB output", len(textOf(o.chat.Message)) == l.ChatOutputBytes, "%d", len(textOf(o.chat.Message)))
					}
					w.record(ev, o)
					o.recordResult(ev, o.metadata.RequestID+"-result")
				}
				var largestFrame, totalResponse int64
				for name, reply := range replies {
					frame, total := largestAndTotal(reply.Chunks)
					totalResponse = max(totalResponse, int64(total))
					if strings.Contains(name, "image") == false {
						largestFrame = max(largestFrame, int64(frame))
					}
					ev.Record("response-"+name, reply)
				}
				// old streamed Chat coalesces frames into network chunks: account for each
				// actual SSE frame instead of treating a coalesced chunk as one frame.
				if phase == 0 {
					largestFrame = 0
					for _, frame := range pressureText(t, l.ChatOutputBytes/4) {
						largestFrame = max(largestFrame, int64(len(frame)))
					}
					f, _ := largestAndTotal(replies[mini.name].Chunks)
					largestFrame = max(largestFrame, int64(f))
				}
				rooms := map[string]headroom{
					"MaxRequestBytes":                             room(largestRequest, p.policy.MaxRequestBytes),
					"MaxFrameBytes":                               room(largestFrame, p.policy.MaxFrameBytes),
					"MaxOutputBytes":                              room(totalResponse, p.policy.MaxOutputBytes),
					"MaxImageBytes":                               room(int64(l.InputImageBytes), p.policy.MaxImageBytes),
					"MaxToolJSONBytes":                            room(0, p.policy.MaxToolJSONBytes),
					"MaxErrorBodyBytes":                           room(0, p.policy.MaxErrorBodyBytes),
					"MaxQueuedEvents":                             room(w.gauges.PeakQueuedEvents(), int64(3*p.tenants*p.policy.MaxQueuedEvents)),
					"MaxQueuedEventBytes":                         room(0, p.policy.MaxQueuedEventBytes),
					"MaxConcurrentPerTenant":                      room(3, int64(p.policy.MaxConcurrentPerTenant)),
					"MaxConcurrentProcess":                        room(int64(3*p.tenants), int64(p.policy.MaxConcurrentProcess)),
					"MaxAdmissionWaiters":                         room(0, int64(p.policy.MaxAdmissionWaiters)),
					"MaxQueuedObservationsConservativeAllRecords": room(int64(4*3*p.tenants), int64(p.policy.MaxQueuedObservations)),
					"Image.MaxInputImages":                        room(int64(l.References+1), int64(p.policy.Image.MaxInputImages)),
					"Image.MaxOutputImages":                       room(int64(l.OutputImages), int64(p.policy.Image.MaxOutputImages)),
					"Image.MaxOutputImageBytes":                   room(int64(l.OutputImageBytes), p.policy.Image.MaxOutputImageBytes),
					"Image.MaxTotalOutputImageBytes":              room(int64(l.OutputImages*l.OutputImageBytes), p.policy.Image.MaxTotalOutputImageBytes),
					"Classifier.MaxQuestions":                     room(int64(l.Questions), int64(p.policy.Classifier.MaxQuestions)),
					"Classifier.MaxStateBytes":                    room(int64(l.StateBytes), p.policy.Classifier.MaxStateBytes),
					"Classifier.MaxQuestionBytes":                 room(int64(l.InstructionBytes+128), p.policy.Classifier.MaxQuestionBytes),
				}
				ev.Check("Complete batch creates no public stream queue", w.gauges.PeakQueuedEvents() == 0, "%d", w.gauges.PeakQueuedEvents())
				for name, h := range rooms {
					ev.Check("additional batch leaves75percent headroom in "+name, h.Ratio <= maxDesignRatio, "%+v", h)
				}
				ev.Check("same memory ceiling including conservative allocations", use.LiveGrowth() <= p.budget && after.TotalAlloc-before.TotalAlloc <= p.budget, "live=%d allocated=%d limit=%d", use.LiveGrowth(), after.TotalAlloc-before.TotalAlloc, p.budget)
				ev.Record("design-load", map[string]any{"baselineCase": "E08-mixed-pressure-" + p.name + "-design-load", "baselineCallsUnchanged": l.Calls, "baselinePermitOccupancy": 1, "additionalCalls": 3 * p.tenants, "load": l, "MiniMaxProtocol": mini.name, "originalRoutes": []string{old[0].id, old[1].id}, "headroom": rooms, "heap": use, "heapBudget": p.budget, "totalAllocDuringCalls": after.TotalAlloc - before.TotalAlloc, "peakPermits": heap.peakGauge.Load(), "queuePeak": w.gauges.PeakQueuedEvents()})
				ev.Record("policy", p.policy)
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}
