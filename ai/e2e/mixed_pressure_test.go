package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Parameters, not whole response/result evidence, make the probe replayable.
// Provider captures at most one bounded request per admitted call. Response
// chunks are shared per route; host-owned generation happens before baseline.
type mixedLoad struct {
	Calls            int `json:"calls"`
	Tenants          int `json:"tenants"`
	HistoryBytes     int `json:"historyBytes"`
	ChatOutputBytes  int `json:"chatOutputBytes"`
	References       int `json:"references"`
	InputImageBytes  int `json:"inputImageBytes"`
	MaskBytes        int `json:"maskBytes"`
	OutputImages     int `json:"outputImages"`
	OutputImageBytes int `json:"outputImageBytes"`
	StateBytes       int `json:"stateBytes"`
	Questions        int `json:"questions"`
	InstructionBytes int `json:"instructionBytes"`
}

func mixedDesign(calls, tenants int) mixedLoad {
	return mixedLoad{Calls: calls, Tenants: tenants, HistoryBytes: 128 << 10, ChatOutputBytes: 16 << 10,
		References: 2, InputImageBytes: 256 << 10, MaskBytes: 256 << 10, OutputImages: 2, OutputImageBytes: 512 << 10,
		StateBytes: 128 << 10, Questions: 8, InstructionBytes: 4000}
}

// A real 1x1 PNG plus zero padding has exact known wire size. The public
// contract validates MIME/signature/base64, not provider pixel complexity.
func mixedPNG(t *testing.T, size int) ai.Image {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jBMsAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	if size > len(data) {
		data = append(data, make([]byte, size-len(data))...)
	}
	return ai.Image{Data: base64.StdEncoding.EncodeToString(data), MimeType: "image/png"}
}

func mixedLoadClassifier(l mixedLoad) ai.ClassifierRequest {
	q := map[string]ai.ClassifierQuestion{}
	for i := range l.Questions {
		q[fmt.Sprintf("q%d", i)] = ai.BoolQuestion{Instructions: json.RawMessage(`"` + strings.Repeat("q", l.InstructionBytes) + `"`)}
	}
	return ai.ClassifierRequest{State: json.RawMessage(`"` + strings.Repeat("s", l.StateBytes-2) + `"`), Questions: q}
}

func mixedLoadReply(t *testing.T, r mixedRoute, l mixedLoad) provider.Reply {
	t.Helper()
	if r.op == ai.OperationChat {
		return pressureReply(coalesce(pressureText(t, l.ChatOutputBytes/4), 64))
	}
	var body any
	if r.op == ai.OperationClassifier {
		answers := map[string]any{}
		for i := range l.Questions {
			answers[fmt.Sprintf("q%d", i)] = map[string]any{"type": "noul", "noul": 0.8}
		}
		body = map[string]any{"model": "synthetic-response", "answers": answers, "usage": map[string]int{"input_tokens": 296, "output_tokens": 20}}
	} else {
		image := mixedPNG(t, l.OutputImageBytes)
		images := make([]any, l.OutputImages)
		for i := range images {
			if r.provider == ai.ProviderGoogle {
				images[i] = map[string]any{"type": "image", "mime_type": "image/png", "data": image.Data}
			} else {
				images[i] = map[string]any{"b64_json": image.Data}
			}
		}
		if r.provider == ai.ProviderGoogle {
			body = map[string]any{"id": "pressure-image", "status": "completed", "steps": []any{map[string]any{"type": "model_output", "content": images}}, "usage": map[string]int{"total_input_tokens": 7, "total_output_tokens": 20, "total_tokens": 27}}
		} else {
			body = map[string]any{"data": images, "usage": map[string]int{"input_tokens": 7, "output_tokens": 20, "total_tokens": 27}}
		}
	}
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{mustMarshal(t, body)}}
}

func TestMixedPressure(t *testing.T) {
	for _, p := range []struct {
		name    string
		policy  *ai.ResourcePolicy
		tenants int
		budget  uint64
	}{
		{"local", localassembly.MixedPolicy(), 1, 256 << 20}, {"cloud", hostintegration.CloudMixedPolicy(), 2, 512 << 20},
	} {
		t.Run(p.name, func(t *testing.T) {
			ev := pressureCase(t, "E08-mixed-pressure-"+p.name+"-design-load")
			l := mixedDesign(p.policy.MaxConcurrentProcess, p.tenants)
			var arrivals, ends sync.WaitGroup
			arrivals.Add(l.Calls)
			ends.Add(3 * p.tenants)
			allArrived, allEnds := doneWhen(arrivals.Wait), doneWhen(ends.Wait)
			release := make(chan struct{})
			var releaseOnce sync.Once
			defer releaseOnce.Do(func() { close(release) })
			eof := &mixedEOFTransport{ended: ends.Done, release: release}
			w := newMixedWorld(t, func(c *ai.Config) { c.Policy = p.policy; eof.next = c.Transport; c.Transport = eof }, tenantA)
			if p.tenants == 2 {
				w.install(tenantB)
			}
			ref, mask := mixedPNG(t, l.InputImageBytes), mixedPNG(t, l.MaskBytes)
			imageReq := ai.ImagesRequest{Prompt: "synthetic design edit"}
			for range l.References {
				imageReq.ReferenceImages = append(imageReq.ReferenceImages, ref)
			}
			classifierReq := mixedLoadClassifier(l)
			chatReq := ai.Request{Messages: []ai.Message{ai.UserText(strings.Repeat("h", l.HistoryBytes))}}
			replies := map[string]provider.Reply{}
			var scripts int64
			for _, r := range mixedRoutes {
				reply := mixedLoadReply(t, r, l)
				replies[r.id] = reply
				for _, ch := range reply.Chunks {
					scripts += int64(len(ch))
				}
			}
			// Every request arrives before replies start. The client transport
			// holds each complete unary body at EOF; chat waits in OnResponse.
			for _, k := range []tenantKey{tenantA, tenantB}[:p.tenants] {
				for _, r := range mixedRoutes {
					reply := replies[r.id]
					reply.OnReceive = func() {
						arrivals.Done()
						select {
						case <-allArrived:
						case <-release:
						}
					}
					w.provider.EnqueueAt(r.prefix(k), reply)
				}
			}
			heap := startHeapSampler(w.gauges.Permits)
			defer heap.finish()
			ctx, cancel := context.WithTimeout(ctxFor(t), pressureDeadline)
			defer cancel()
			results := make([]mixedOutcome, l.Calls)
			var wg sync.WaitGroup
			defer wg.Wait()
			defer cancel()
			defer releaseOnce.Do(func() { close(release) })
			var allocationStart runtime.MemStats
			runtime.ReadMemStats(&allocationStart)
			started := time.Now()
			for n, k := range []tenantKey{tenantA, tenantB}[:p.tenants] {
				for i, r := range mixedRoutes {
					wg.Go(func() {
						scope := scopeFor(k, fmt.Sprintf("req-mixed-pressure-%s-%d", p.name, n*4+i))
						target := ai.Target{BindingID: r.id, ModelID: mixedModel}
						o := mixedOutcome{}
						switch r.op {
						case ai.OperationChat:
							hooked := w.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
								select {
								case <-release:
									return nil
								case <-ctx.Done():
									return ctx.Err()
								}
							}})
							res, err := hooked.Complete(ctx, scope, target, chatReq, nil)
							o = mixedOutcome{metadata: res.Metadata, chat: res, usage: res.Message.Usage, err: err}
						case ai.OperationClassifier:
							res, err := w.client.Classify(ctx, scope, target, classifierReq, nil)
							o = mixedOutcome{metadata: res.Metadata, classifier: res, usage: res.Usage, err: err}
						case ai.OperationImage:
							var opts ai.ImageOptions
							if r.provider == ai.ProviderOpenAI {
								opts = ai.OpenAIImagesOptions{N: ai.Value(l.OutputImages), Mask: &mask}
							}
							res, err := w.client.GenerateImages(ctx, scope, target, imageReq, opts)
							o = mixedOutcome{metadata: res.Metadata, images: res, usage: res.Usage, err: err}
						}
						results[n*4+i] = o
					})
				}
			}
			waitForWithin(ev, "all design responses waiting before EOF", pressureDeadline, allEnds)
			var expectedRead int64
			for _, r := range mixedRoutes[1:] {
				for _, ch := range replies[r.id].Chunks {
					expectedRead += int64(len(ch)) * int64(p.tenants)
				}
			}
			ev.Check("client received every unary response byte before EOF", eof.read.Load() == expectedRead, "read %d want %d", eof.read.Load(), expectedRead)
			ev.Check("all mixed permits held through complete body", w.gauges.Permits() == int64(l.Calls), "held %d", w.gauges.Permits())
			runtime.GC() // record the reachable request/body peak, not only garbage
			heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
			heap.peakGauge.Store(max(heap.peakGauge.Load(), w.gauges.Permits()))
			releaseOnce.Do(func() { close(release) })
			waitForWithin(ev, "all design calls end", pressureDeadline, doneWhen(wg.Wait))
			elapsed := time.Since(started)
			var allocationEnd runtime.MemStats
			runtime.ReadMemStats(&allocationEnd)
			allocated := allocationEnd.TotalAlloc - allocationStart.TotalAlloc
			// Keep full validated outputs alive for one GC to count publication.
			runtime.GC()
			heap.peakLive.Store(max(heap.peakLive.Load(), liveHeap()))
			use := heap.finish()
			for _, o := range results {
				ev.Check("mixed design call successful", o.err == nil, "%s: %v", o.metadata.Operation, o.err)
				if o.metadata.Operation == ai.OperationImage {
					ev.Check("all validated images published", len(o.images.Content) == l.OutputImages, "got %d", len(o.images.Content))
				}
				if o.metadata.Operation == ai.OperationClassifier {
					ev.Check("all typed answers published", len(o.classifier.Answers) == l.Questions, "got %d", len(o.classifier.Answers))
				}
				w.record(ev, o)
			}
			var largest, captured int64
			for _, req := range w.provider.Requests() {
				largest = max(largest, int64(len(req.Body)))
				captured += int64(len(req.Body))
			}
			var totalResponse, largestFrame int64
			for _, reply := range replies {
				var total int64
				for _, ch := range reply.Chunks {
					total += int64(len(ch))
				}
				totalResponse = max(totalResponse, total)
			}
			for _, frame := range pressureText(t, l.ChatOutputBytes/4) {
				largestFrame = max(largestFrame, int64(len(frame)))
			}
			var largestQuestion int64
			for _, q := range classifierReq.Questions {
				largestQuestion = max(largestQuestion, int64(len(mustMarshal(t, q))))
			}
			rooms := map[string]headroom{
				"MaxRequestBytes": room(largest, p.policy.MaxRequestBytes), "MaxImageBytes": room(int64(l.InputImageBytes), p.policy.MaxImageBytes),
				"MaxFrameBytes (chat only)": room(largestFrame, p.policy.MaxFrameBytes), "MaxOutputBytes": room(totalResponse, p.policy.MaxOutputBytes),
				"Image.MaxInputImages": room(int64(l.References+1), int64(p.policy.Image.MaxInputImages)), "Image.MaxOutputImages": room(int64(l.OutputImages), int64(p.policy.Image.MaxOutputImages)),
				"Image.MaxOutputImageBytes": room(int64(l.OutputImageBytes), p.policy.Image.MaxOutputImageBytes), "Image.MaxTotalOutputImageBytes": room(int64(l.OutputImages*l.OutputImageBytes), p.policy.Image.MaxTotalOutputImageBytes),
				"Classifier.MaxQuestions": room(int64(l.Questions), int64(p.policy.Classifier.MaxQuestions)), "Classifier.MaxStateBytes": room(int64(l.StateBytes), p.policy.Classifier.MaxStateBytes), "Classifier.MaxQuestionBytes": room(largestQuestion, p.policy.Classifier.MaxQuestionBytes),
			}
			for name, h := range rooms {
				ev.Check("design headroom under "+name, h.Ratio <= maxDesignRatio, "got %+v", h)
			}
			ev.Check("unary legal JSON exceeds SSE frame budget", int64(len(replies["openai-image"].Chunks[0])) > p.policy.MaxFrameBytes, "image response too small")
			ev.Check("reachable growth within explicit memory budget", use.LiveGrowth() <= p.budget, "growth %d budget %d", use.LiveGrowth(), p.budget)
			// TotalAlloc is a deliberately conservative upper bound on new
			// reachable bytes at any instant: it includes every temporary copy
			// even if created and freed between GC samples, plus test/observer
			// allocations. LiveGrowth is the GC-measured lower bound only.
			ev.Check("all allocations bound even unsampled parsing peaks", allocated <= p.budget, "allocated %d budget %d", allocated, p.budget)
			ev.Record("mixed-pressure", map[string]any{"load": l, "headroom": rooms, "heap": use, "heapGrowth": use.LiveGrowth(), "heapBudget": p.budget, "totalAllocDuringCalls": allocated, "clientUnaryBytesAtEOF": eof.read.Load(), "peakPermits": heap.peakGauge.Load(), "providerCapturedRequestBytes": captured, "providerSharedResponseScriptBytes": scripts, "inputImageBase64Bytes": len(ref.Data), "outputImageBase64Bytes": base64.StdEncoding.EncodedLen(l.OutputImageBytes), "duration": elapsed.String(), "callsPerSecond": float64(l.Calls) / elapsed.Seconds(), "captureBoundCalls": l.Calls})
			w.released(ev)
		})
	}
}
