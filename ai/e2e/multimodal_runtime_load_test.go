package e2e

import (
	"context"
	"encoding/base64"
	"runtime"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func checkMultimodalSuccess(t *testing.T, ev *evidence.Case, r mixedRoute, req ai.ClassifierRequest, res ai.ClassifierResult, err error) {
	if !ev.Check("complete mixed answers", err == nil && len(res.Answers) == len(req.Questions), "got %v", err) {
		return
	}
	expected := []ai.Image{req.Images[1], req.Images[0]}
	if r.api == ai.APIOpenAIDecisions {
		expected = append(expected, req.Images[1])
	}
	checkImageInputs(ev, res, expected)
	ev.Check("one legal priced usage", res.Usage.Input > 0 && res.Usage.Cost.Total > 0, "lost usage")
	if r.api == ai.APITypeSafeSystemOne {
		ev.Check("negative existence does not suppress Box", jsonEqual(mustMarshal(t, res.Answers["location"]), []byte(jevBoxAnswer)) && res.Answers["approved"].(ai.BoolAnswer).Probability == 0, "conditional or changed answer")
		score := res.Answers["urgency"].(ai.ScoreAnswer)
		ev.Check("public legend points into final images", jsonEqual(mustMarshal(t, score.LegendImageReferences), []byte(`[{"imageIndex":1,"placement":"value","pointer":"/0"}]`)) && string(score.Legend[0]) == "null", "wrong legend")
	}
}

// Normal and race both run these four concurrent calls, including real image
// validation, complete response reads, all answers, captures and Observer.
func TestMultimodalRuntimeDesignLoad(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC09-design-load-"+entry)
			ended, release := make(chan struct{}, 4), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			var eof *mixedEOFTransport
			u := newMultimodalWorld(t, ev, func(c *ai.Config) {
				eof = &mixedEOFTransport{next: c.Transport, ended: func() { ended <- struct{}{} }, release: release}
				c.Transport = eof
			})
			var reqs [4]ai.ClassifierRequest
			var results [4]ai.ClassifierResult
			var errs [4]error
			for n, k := range []tenantKey{tenantA, tenantB} {
				for j, r := range multimodalRoutes {
					i := 2*n + j
					reqs[i] = multimodalInput(t, r, k)
					multimodalEnqueue(ev, u, r, k, multimodalReply(t, ev, r, false))
					ev.Record(string(r.api)+"-"+k.tenant+"-input", reqs[i])
				}
			}
			heap := startHeapSampler(u.gauges.Permits)
			defer heap.finish()
			var before, atEOF runtime.MemStats
			runtime.ReadMemStats(&before)
			var wg sync.WaitGroup
			for n, k := range []tenantKey{tenantA, tenantB} {
				for j, r := range multimodalRoutes {
					i := 2*n + j
					wg.Go(func() {
						results[i], errs[i] = multimodalCall(ctxFor(t), u, r, k, entry, "req-load-"+k.tenant+r.id+entry, reqs[i], ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
							return ai.KeepPayload(), nil
						}})
					})
				}
			}
			for range 4 {
				waitFor(ev, "complete body at EOF", ended)
			}
			runtime.GC()
			runtime.ReadMemStats(&atEOF)
			ev.Check("all four permits and bodies held", u.gauges.Permits() == 4 && u.bodies.Open() == 4 && u.adm.Held() == 4, "load was not simultaneous")
			allocation := atEOF.TotalAlloc - before.TotalAlloc
			ev.Check("bounded batch allocation", allocation < 64<<20, "allocated %d", allocation)
			unblock()
			wg.Wait()
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			fullAllocation := after.TotalAlloc - before.TotalAlloc
			ev.Check("complete batch allocation includes response decoding", fullAllocation < 64<<20, "allocated %d", fullAllocation)
			use := heap.finish()
			ev.Check("bounded live growth", use.LiveGrowth() < 64<<20, "grew %d", use.LiveGrowth())
			rooms := map[string]headroom{}
			for _, capture := range u.provider.Requests() {
				rooms["request-"+capture.KeyAlias+capture.Path] = room(int64(len(capture.Body)), 64<<10)
			}
			for n, k := range []tenantKey{tenantA, tenantB} {
				for j, r := range multimodalRoutes {
					i := 2*n + j
					checkMultimodalSuccess(t, ev, r, reqs[i], results[i], errs[i])
					recordMultimodalCall(ev, u, reqs[i], results[i], errs[i])
					ev.Check("same binding and correct tenant", results[i].Metadata.TenantID == k.tenant && results[i].Metadata.BindingID == r.id && results[i].Metadata.API == r.api, "wrong attribution")
				}
			}
			for _, capture := range u.provider.Requests() {
				k := tenantA
				if capture.KeyAlias == tenantB.alias {
					k = tenantB
				}
				r := multimodalRoutes[0]
				if capture.Path == r.prefix(k)+"/systemone" {
				} else {
					r = multimodalRoutes[1]
				}
				req := multimodalInput(t, r, k)
				if r.api == ai.APITypeSafeSystemOne {
					var native struct {
						Assets map[string]struct {
							Data string `json:"data_base64"`
						}
						State struct {
							Images []string
							State  struct{ Photo string }
						}
					}
					mustUnmarshal(t, capture.Body, &native)
					ev.Check("Jev shares sources and repeats ordered context", len(native.Assets) == 2 && len(native.State.Images) == 3 && native.State.Images[0] == native.State.Images[2] && native.Assets[native.State.Images[0]].Data == req.Images[1].Data && native.Assets[native.State.State.Photo].Data == req.Images[0].Data, "cross-tenant image")
				} else {
					var native struct {
						Input []struct {
							Content []struct {
								URL string `json:"image_url"`
							}
						}
					}
					mustUnmarshal(t, capture.Body, &native)
					if ev.Check("Decisions has ordered repeated blocks", len(native.Input) == 1 && len(native.Input[0].Content) == 4, "wrong input") {
						for i, index := range []int{1, 0, 1} {
							ev.Check("credential matches ordered tenant file", native.Input[0].Content[i+1].URL == "data:"+req.Images[index].MimeType+";base64,"+req.Images[index].Data, "cross-tenant block")
						}
					}
				}
			}
			var sourceBytes, expandedBytes int64
			for _, img := range reqs[0].Images {
				raw, _ := base64.StdEncoding.DecodeString(img.Data)
				sourceBytes += int64(len(raw))
			}
			for _, ref := range reqs[0].ImageReferences {
				raw, _ := base64.StdEncoding.DecodeString(reqs[0].Images[ref.ImageIndex].Data)
				expandedBytes += int64(len(raw))
			}
			rooms["source"] = room(sourceBytes, 64<<10)
			rooms["expanded"] = room(expandedBytes, 128<<10)
			rooms["references"] = room(6, 8)
			ev.Record("load", map[string]any{"calls": 4, "tenants": 2, "protocols": 2, "sourceImagesPerCall": 2, "jevOccurrences": 6, "decisionsOccurrences": 3, "headroom": rooms, "heap": use, "totalAllocAtEOF": allocation, "totalAlloc": fullAllocation, "memoryCeiling": 64 << 20, "heldBodiesAtEOF": 4, "httpBytesAtEOF": eof.read.Load(), "realService": "NOT_RUN"})
			u.released(ev)
		})
	}
}

func TestMultimodalRuntimeInputSnapshot(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(r.id+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC09-input-snapshot-"+r.id+"-"+entry)
				req := multimodalInput(t, r, tenantA)
				var original ai.ClassifierRequest
				mustUnmarshal(t, mustMarshal(t, req), &original)
				var config *ai.ResourcePolicy
				var catalog *ai.Catalog
				var mutate func()
				u := newMultimodalWorld(t, ev, func(c *ai.Config) {
					config = c.Policy
					catalog = c.Catalog
					previous := c.Bindings
					c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
						if mutate != nil {
							mutate()
						}
						return previous.ResolveBinding(ctx, s, id)
					})
				})
				pinned := u.client.Catalog()
				hash, _ := pinned.Hash()
				mutate = func() {
					req.Images[0] = req.Images[1]
					req.State[1] = 'X'
					req.Questions["intent"] = nil
					req.ImageReferences[0].ImageIndex = 0
					config.Classifier.Images.References.MaxReferences = 1
					config.Classifier.Images.MaxImages = 1
					catalog.ClassifierModels[0].Cost.Input = 999
					catalog.ClassifierModels[1].Cost.Input = 999
				}
				multimodalEnqueue(ev, u, r, tenantA, multimodalReply(t, ev, r, false))
				res, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-input-snapshot-"+r.id+entry, req, ai.Hooks{})
				checkMultimodalSuccess(t, ev, r, original, res, err)
				ev.Check("configuration and catalog pinned", res.Metadata.CatalogHash == hash && res.Metadata.CatalogVersion == pinned.Version && res.Usage.Cost.Input < 1, "mutated snapshot")
				recordMultimodalCall(ev, u, original, res, err)
				if err != nil {
					return
				}
				priorAnswers := mustMarshal(t, res.Answers)
				priorImages := mustMarshal(t, res.ImageInputs)
				priorUsage := res.Usage
				res.ImageInputs[0].SHA256 = "changed"
				res.Answers["intent"] = nil
				if score, ok := res.Answers["urgency"].(ai.ScoreAnswer); ok && len(score.Legend) > 0 {
					score.Legend[0][0] = 'X'
				}
				mutate = nil
				multimodalEnqueue(ev, u, r, tenantA, multimodalReply(t, ev, r, false))
				next, e := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-result-snapshot-"+r.id+entry, original, ai.Hooks{})
				checkMultimodalSuccess(t, ev, r, original, next, e)
				ev.Check("returned values do not mutate future results", jsonEqual(priorImages, mustMarshal(t, next.ImageInputs)) && jsonEqual(priorAnswers, mustMarshal(t, next.Answers)) && priorUsage == next.Usage, "aliased result")
				recordMultimodalCall(ev, u, original, next, e)
				u.released(ev)
			})
		}
	}
}
