package e2e

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMultimodalRuntimeObserver(t *testing.T) {
	for _, mode := range []string{"slow", "congested", "error", "panic"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(mode+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC09-observer-"+mode+"-"+entry)
				rec := newRecorder()
				var release sync.Once
				if mode == "slow" || mode == "congested" {
					rec.gate = make(chan struct{})
					defer release.Do(func() { close(rec.gate) })
				}
				if mode == "error" {
					rec.fail = errors.New("synthetic private Observer error")
				}
				if mode == "panic" {
					rec.panics = true
				}
				u := newMultimodalWorld(t, ev, func(c *ai.Config) {
					c.Observer = rec
					if mode == "congested" {
						c.Policy.MaxQueuedObservations = 1
					}
				})
				var wg sync.WaitGroup
				results := make([]ai.ClassifierResult, 4)
				errs := make([]error, 4)
				for n, k := range []tenantKey{tenantA, tenantB} {
					for j, r := range multimodalRoutes {
						i := 2*n + j
						multimodalEnqueue(ev, u, r, k, multimodalReply(t, ev, r, false))
						wg.Go(func() {
							req := multimodalInput(t, r, k)
							results[i], errs[i] = multimodalCall(ctxFor(t), u, r, k, entry, "req-observer-"+mode+entry+k.tenant+r.id, req, ai.Hooks{})
						})
					}
				}
				wg.Wait()
				for n, k := range []tenantKey{tenantA, tenantB} {
					for j, r := range multimodalRoutes {
						i := 2*n + j
						checkMultimodalSuccess(t, ev, r, multimodalInput(t, r, k), results[i], errs[i])
						ev.Record(results[i].Metadata.RequestID+"-result", results[i])
						ev.Record(results[i].Metadata.RequestID+"-input", multimodalInput(t, r, k))
					}
				}
				r := multimodalRoutes[0]
				req := multimodalInput(t, r, tenantA)
				body := `{"error":"synthetic private error"}`
				multimodalEnqueue(ev, u, r, tenantA, provider.Reply{Status: 503, Chunks: [][]byte{[]byte(body)}, Script: body})
				failed, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-observer-failed-"+mode+entry, req, ai.Hooks{})
				ev.Check("Observer cannot change failure", errors.Is(err, &ai.Error{Code: ai.CodeUpstreamError}) && len(failed.Metadata.Attempts) == 1, "got %v", err)
				ev.Record("failed-result", failed)
				ev.Record("failed-input", req)
				ev.Record("failed-error", errString(err))
				if rec.gate != nil {
					release.Do(func() { close(rec.gate) })
				}
				eventually(ev, "all 20 records accounted", func() bool { s := u.client.ObserverStats(); return s.Delivered+s.Failed+s.Dropped == 20 })
				stats := u.client.ObserverStats()
				switch mode {
				case "slow":
					ev.Check("slow preserves terminals", stats.Delivered == 20 && stats.Dropped == 0, "got %+v", stats)
				case "congested":
					ev.Check("bounded queue reports drops", stats.Dropped > 0, "got %+v", stats)
				default:
					ev.Check("failure only counts deliveries", stats.Failed == 20 && stats.Delivered == 0, "got %+v", stats)
				}
				obs := observationText(t, rec.all())
				for _, forbidden := range []string{"synthetic private", "data_base64", "image_url", "confidence", `"top":`, "/state/photo", "<image:", tenantA.secret, tenantB.secret} {
					ev.Check("Observer metadata only", !strings.Contains(obs, forbidden), "leaked content")
				}
				for _, img := range expectedImageInputs(t, jevContextImages(t)) {
					ev.Check("Observer excludes image hash", !strings.Contains(obs, img["sha256"]), "image hash leaked")
				}

				ev.Record("observer-records", rec.all())
				ev.Record("observer-stats", stats)
				ev.Record("requests", u.provider.Requests())
				u.released(ev)
			})
		}
	}
}
