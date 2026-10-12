package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// EOF is measured at Client's HTTP boundary, after io.ReadAll received every
// byte. Provider scripts and Host inputs exist before the allocation baseline.
func runClassifierImageBatch(t *testing.T, ev *evidence.Case, w *mixedWorld, name, kind string, req ai.ClassifierRequest, reply provider.Reply, l mixedLoad, budget uint64, transport *mixedBatchTransport) (map[string]any, int64) {
	t.Helper()
	var arrivals, ends sync.WaitGroup
	arrivals.Add(l.Calls)
	ends.Add(l.Calls)
	allArrived, allEnds := doneWhen(arrivals.Wait), doneWhen(ends.Wait)
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	eof := &mixedEOFTransport{ended: ends.Done, release: release}
	// All three batches use the same Client, including built-in Admission.
	eof.next = transport.next
	transport.barrier.Store(eof)
	for _, k := range []tenantKey{tenantA, tenantB}[:l.Tenants] {
		for range 4 {
			r := reply
			r.OnReceive = func() {
				arrivals.Done()
				select {
				case <-allArrived:
				case <-release:
				}
			}
			w.provider.EnqueueAt(loadClassifierRoute.prefix(k), r)
		}
	}
	captureStart := len(w.provider.Requests())
	heap := startHeapSampler(w.gauges.Permits)
	defer heap.finish()
	ctx, cancel := context.WithTimeout(ctxFor(t), pressureDeadline)
	defer cancel()
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	defer once.Do(func() { close(release) })
	results := make([]ai.ClassifierResult, l.Calls)
	errs := make([]error, l.Calls)
	var start runtime.MemStats
	runtime.ReadMemStats(&start)
	began := time.Now()
	for n, k := range []tenantKey{tenantA, tenantB}[:l.Tenants] {
		for i := range 4 {
			wg.Go(func() {
				scope := scopeFor(k, fmt.Sprintf("req-load-%s-%s-%d", name, kind, n*4+i))
				results[n*4+i], errs[n*4+i] = w.client.Classify(ctx, scope, ai.Target{BindingID: "luna", ModelID: loadClassifierModel}, req, nil)
			})
		}
	}
	waitForWithin(ev, kind+": every complete response awaits EOF", pressureDeadline, allEnds)
	expectedRead := int64(l.Calls * len(reply.Chunks[0]))
	ev.Check(kind+": all unary response bytes retained", eof.read.Load() == expectedRead, "got %d want %d", eof.read.Load(), expectedRead)
	held := w.gauges.Permits()
	ev.Check(kind+": process concurrency occupied", held == int64(l.Calls), "held %d", held)
	runtime.GC()
	heldLive := liveHeap()
	heap.peakLive.Store(max(heap.peakLive.Load(), heldLive))
	heap.peakGauge.Store(max(heap.peakGauge.Load(), held))
	if kind == "cancel" {
		cancel()
	} else {
		once.Do(func() { close(release) })
	}
	waitForWithin(ev, kind+": calls terminate", pressureDeadline, doneWhen(wg.Wait))
	elapsed := time.Since(began)
	var end runtime.MemStats
	runtime.ReadMemStats(&end)
	allocated := end.TotalAlloc - start.TotalAlloc
	runtime.GC()
	publishedLive := liveHeap()
	heap.peakLive.Store(max(heap.peakLive.Load(), publishedLive))
	use := heap.finish()
	successful, canceled := 0, 0
	for i, res := range results {
		if kind == "cancel" {
			ev.Check("cancel after 2xx does not publish answers", errors.Is(errs[i], &ai.Error{Code: ai.CodeCanceled}) && len(res.Answers) == 0, "got %v", errs[i])
			if errs[i] != nil {
				canceled++
			}
		} else {
			ev.Check(kind+": complete typed answers", errs[i] == nil && len(res.Answers) == l.Questions, "got %v / %d", errs[i], len(res.Answers))
			for _, a := range res.Answers {
				v, ok := a.(ai.BoolAnswer)
				ev.Check("predicate value retained", ok && v.Probability == .8, "%+v", a)
			}
			if errs[i] == nil {
				successful++
			}
		}
		ev.Check("own route and one attempt", res.Metadata.Operation == ai.OperationClassifier && res.Metadata.ProviderID == ai.ProviderOpenAI && res.Metadata.API == ai.APIOpenAIDecisions && res.Metadata.ModelID == loadClassifierModel && len(res.Metadata.Attempts) == 1, "%+v", res.Metadata)
		w.record(ev, mixedOutcome{metadata: res.Metadata, usage: res.Usage, classifier: res, err: errs[i]})
		ev.Record(fmt.Sprintf("%s-result-%d", kind, i), res)
	}
	runtime.KeepAlive(results)
	requests := w.provider.Requests()[captureStart:]
	var captured, largest int64
	for _, r := range requests {
		captured += int64(len(r.Body))
		largest = max(largest, int64(len(r.Body)))
	}
	ev.Check("one bounded request copy per call", len(requests) == l.Calls, "got %d", len(requests))
	ev.Record(kind+"-requests", requests)
	ev.Check(kind+": reachable heap lower bound within budget", use.LiveGrowth() <= budget, "%d / %d", use.LiveGrowth(), budget)
	ev.Check(kind+": allocation upper bound within budget", allocated > 0 && allocated <= budget, "%d / %d", allocated, budget)
	resources := map[string]int64{"calls": w.gauges.ActiveCalls(), "permits": w.gauges.Permits(), "waiters": w.gauges.AdmissionWaiters(), "events": w.gauges.QueuedEvents(), "bodies": w.bodies.Open(), "hostPermits": int64(w.adm.Held())}
	for key, n := range resources {
		ev.Check(kind+": released "+key, n == 0, "%d", n)
	}
	return map[string]any{"kind": kind, "calls": l.Calls, "successful": successful, "canceled": canceled, "peakPermits": heap.peakGauge.Load(), "permitsAtEOF": held, "clientUnaryBytesAtEOF": eof.read.Load(), "responseBytesPerCall": len(reply.Chunks[0]), "heap": use, "heapGrowth": use.LiveGrowth(), "heapAtEOF": heldLive, "heapWithResults": publishedLive, "totalAllocDuringCalls": allocated, "providerCapturedRequestBytes": captured, "duration": elapsed.String(), "callsPerSecond": float64(l.Calls) / elapsed.Seconds(), "resources": resources}, largest
}

// Compile-time check on the measurement transport's public boundary.
var _ http.RoundTripper = (*mixedEOFTransport)(nil)

// Swap the measurement barrier only between completed batches. The Client and
// all its admission/observer state survive cancellation and recovery.
type mixedBatchTransport struct {
	next    http.RoundTripper
	barrier atomic.Pointer[mixedEOFTransport]
}

func (t *mixedBatchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if barrier := t.barrier.Load(); barrier != nil {
		return barrier.RoundTrip(req)
	}
	return t.next.RoundTrip(req)
}
