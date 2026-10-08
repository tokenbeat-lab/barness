package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"runtime"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMixedImageBudgets(t *testing.T) {
	for _, r := range mixedRoutes[1:3] {
		for _, mode := range []string{"above-frame", "single-image", "total-images", "response", "count", "request", "input-image"} {
			t.Run(r.id+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "E08-mixed-image-budget-"+r.id+"-"+mode)
				w := newMixedWorld(t, func(c *ai.Config) {
					switch mode {
					case "single-image":
						c.Policy.Image.MaxOutputImageBytes = 128 << 10
					case "total-images":
						c.Policy.Image.MaxOutputImageBytes = 512 << 10
						c.Policy.Image.MaxTotalOutputImageBytes = 768 << 10
					case "response":
						c.Policy.MaxOutputBytes = 1 << 20
						c.Policy.Image.MaxOutputImageBytes = 512 << 10
						c.Policy.Image.MaxTotalOutputImageBytes = 1 << 20
					case "count":
						c.Policy.Image.MaxOutputImages = 1
					}
				}, tenantA, tenantB)
				held, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				defer once.Do(func() { close(release) })
				chat := mixedRoutes[0]
				w.provider.EnqueueAt(chat.prefix(tenantB), chat.reply(t))
				sibling := make(chan mixedOutcome, 1)
				go func() {
					sibling <- chat.invokeHooked(ctxFor(t), w.client, scopeFor(tenantB, "req-budget-sibling"), ai.Target{BindingID: chat.id, ModelID: mixedModel}, ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
						close(held)
						select {
						case <-release:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}})
				}()
				waitFor(ev, "other call owns its permit", held)
				l := mixedDesign(4, 1)
				reply := mixedLoadReply(t, r, l)
				w.provider.EnqueueAt(r.prefix(tenantA), reply)
				input := ai.ImagesRequest{Prompt: "synthetic budget edit"}
				if mode == "request" {
					ref := mixedPNG(t, 1<<20)
					input.ReferenceImages = []ai.Image{ref, ref, ref, ref}
				}
				if mode == "input-image" {
					input.ReferenceImages = []ai.Image{mixedPNG(t, (1<<20)+1)}
				}
				var opts ai.ImageOptions
				if r.provider == ai.ProviderOpenAI && mode != "count" {
					opts = ai.OpenAIImagesOptions{N: ai.Value(2)}
				}
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				res, err := w.client.GenerateImages(ctxFor(t), textScope("req-budget-subject"), ai.Target{BindingID: r.id, ModelID: mixedModel}, input, opts)
				runtime.ReadMemStats(&after)
				if mode == "above-frame" {
					ev.Check("whole legal unary JSON succeeds above frame limit", err == nil && len(res.Content) == 2, "got %v", err)
				} else {
					ev.Check("limit fails atomically", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit}) && len(res.Content) == 0, "got %v", err)
					if mode == "request" || mode == "input-image" {
						ev.Check("known size checked before copies and credential", after.TotalAlloc-before.TotalAlloc < 2<<20 && w.host.ReadsFor("req-budget-subject").Credentials == 0 && len(res.Metadata.Attempts) == 0, "allocated %d, reads %+v", after.TotalAlloc-before.TotalAlloc, w.host.ReadsFor("req-budget-subject"))
					} else if mode != "response" {
						ev.Check("failed output retains consumption", res.Usage.Input == 7 && len(res.Metadata.Attempts) == 1, "got %+v", res)
					}
				}
				ev.Check("subject returns only its resources", w.gauges.Permits() == 1 && w.bodies.Open() == 1, "held %d bodies %d", w.gauges.Permits(), w.bodies.Open())
				w.record(ev, mixedOutcome{metadata: res.Metadata, usage: res.Usage, images: res, err: err})
				h := sha256.Sum256(reply.Chunks[0])
				ev.Record("generated-response", map[string]any{"parameters": l, "bytes": len(reply.Chunks[0]), "sha256": hex.EncodeToString(h[:]), "maxFrameBytes": 256 << 10, "allocatedBytes": after.TotalAlloc - before.TotalAlloc})
				once.Do(func() { close(release) })
				o := <-sibling
				w.record(ev, o)
				ev.Check("other result remains successful", o.err == nil, "got %v", o.err)
				w.released(ev)
			})
		}
	}
}
