package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const dimensionEvidence = `{"edgeMultiple":16,"maxAspectRatio":3,"maxEdge":3840,"minPixels":655360,"maxPixels":8294400,"source":"https://developers.openai.com/api/docs/guides/image-generation (2026-10-08); synthetic catalog"}`

func editScenario(t *testing.T) fixtureScenario {
	t.Helper()
	f, _ := loadFixture(t, imagesProtocol, "editing.json")
	return f.Scenarios[1]
}

func editOptions(t *testing.T, fields map[string]any) ai.ImageOptions {
	t.Helper()
	return imagesOptions(t, mustMarshal(t, fields))
}

func TestOpenAIImagesEditingPreflight(t *testing.T) {
	for _, name := range []string{"host-count", "model-count", "protocol-count", "zero-references", "mask-count", "mask-bytes", "url-length", "request-bytes", "bad-mime", "wrong-mime", "truncated-base64", "broken-tail", "mask-without-reference", "mask-size", "mask-broken-size", "mask-unsupported", "fidelity-unsupported", "fidelity-low-only", "fidelity-invalid", "fidelity-null", "fidelity-generation", "size-multiple", "size-ratio", "size-edge", "size-min-pixels", "size-max-pixels", "size-explicit-bypass", "transparent-jpeg", "compression-png", "quality"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E07-edit-"+name)
			sc := editScenario(t)
			w := newWorldWith(t, func(c *ai.Config) {
				configureImages(c)
				if strings.HasPrefix(name, "size-") {
					patchImageModel(t, c, json.RawMessage(`{"capabilities":{"maxReferenceImages":16,"maxOutputImages":10,"sizes":["auto","16x16"],"customSizes":`+dimensionEvidence+`}}`))
				}
				switch name {
				case "host-count", "mask-count":
					c.Policy.Image.MaxInputImages = 2
				case "model-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 1
				case "protocol-count", "url-length":
					c.Policy.Image.MaxInputImages = 32
					c.Policy.MaxImageBytes, c.Policy.MaxRequestBytes = 32<<20, 64<<20
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 32
				case "zero-references":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 0
					c.Catalog.ImageModels[0].Capabilities.Mask = false
					c.Catalog.ImageModels[0].Capabilities.InputFidelity = nil
				case "mask-bytes":
					c.Policy.MaxImageBytes = 100
				case "mask-unsupported":
					c.Catalog.ImageModels[0].Capabilities.Mask = false
				case "fidelity-unsupported":
					c.Catalog.ImageModels[0].Capabilities.InputFidelity = nil
				case "fidelity-low-only":
					c.Catalog.ImageModels[0].Capabilities.InputFidelity = []string{"low"}
				}
			}, tenantA)
			installImagesBinding(w, tenantA)
			req := imagesInput(t, sc)
			var opts map[string]any
			mustUnmarshal(t, sc.Options, &opts)
			delete(opts, "mask")
			delete(opts, "inputFidelity")
			switch name {
			case "host-count":
				req.ReferenceImages = append(req.ReferenceImages, req.ReferenceImages[0])
			case "protocol-count":
				req.ReferenceImages = make([]ai.Image, 17)
			case "mask-count", "mask-unsupported", "mask-without-reference":
				mustUnmarshal(t, sc.Options, &opts)
				if name == "mask-without-reference" {
					req.ReferenceImages = nil
				}
			case "mask-bytes":
				opts["mask"] = ai.Image{MimeType: "image/png", Data: strings.Repeat("A", 136)}
			case "url-length":
				req.ReferenceImages = []ai.Image{{MimeType: "image/png", Data: strings.Repeat("A", 20971520)}}
			case "request-bytes":
				req.ReferenceImages = []ai.Image{req.ReferenceImages[0], req.ReferenceImages[0], req.ReferenceImages[0]}
				req.ReferenceImages[0].Data = strings.Repeat("A", 500000)
				req.ReferenceImages[1].Data = req.ReferenceImages[0].Data
				req.ReferenceImages[2].Data = req.ReferenceImages[0].Data
			case "bad-mime":
				req.ReferenceImages[0].MimeType = "image/png; charset=utf-8"
			case "wrong-mime":
				req.ReferenceImages[0].MimeType = "image/jpeg"
			case "truncated-base64":
				req.ReferenceImages[0].Data = strings.TrimRight(req.ReferenceImages[0].Data, "=")
			case "broken-tail":
				req.ReferenceImages[0].Data += "!!!!"
			case "mask-size":
				opts["mask"] = req.ReferenceImages[1]
			case "mask-broken-size":
				opts["mask"] = ai.Image{MimeType: "image/png", Data: "iVBORw0KGgo="}
			case "fidelity-invalid":
				opts["inputFidelity"] = "max"
			case "fidelity-null":
				opts["inputFidelity"] = nil
			case "fidelity-unsupported", "fidelity-low-only", "fidelity-generation":
				opts["inputFidelity"] = "high"
				if name == "fidelity-generation" {
					req.ReferenceImages = nil
				}
			case "size-multiple":
				opts["size"] = "1537x864"
			case "size-ratio":
				opts["size"] = "3072x768"
			case "size-edge":
				opts["size"] = "4096x1024"
			case "size-min-pixels", "size-explicit-bypass":
				opts["size"] = "16x16"
			case "size-max-pixels":
				opts["size"] = "3072x3072"
			case "transparent-jpeg":
				opts["background"], opts["outputFormat"] = "transparent", "jpeg"
			case "compression-png":
				opts["outputCompression"] = 0
			case "quality":
				opts["quality"] = "max"
			}
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-edit-preflight-"+name), sc.target(), req, editOptions(t, opts))
			ev.Record("result", res)
			ev.Record("error", errString(err))
			want := ai.CodeInvalidRequest
			if strings.Contains(name, "count") || name == "mask-bytes" || name == "url-length" || name == "request-bytes" || name == "zero-references" {
				want = ai.CodeResourceLimit
			}
			ev.Check("invalid edit refused before credentials and network", errors.Is(err, &ai.Error{Code: want}) && len(w.provider.Requests()) == 0 && w.host.ReadsFor("req-edit-preflight-"+name).Credentials == 0 && len(res.Content) == 0, "got %v", err)
		})
	}
}

func TestOpenAIImagesEditingSnapshot(t *testing.T) {
	for _, pointer := range []bool{false, true} {
		t.Run(map[bool]string{false: "value", true: "pointer"}[pointer], func(t *testing.T) {
			ev := run.Case(t, "P08-E07-edit-snapshot-"+map[bool]string{false: "value", true: "pointer"}[pointer])
			sc := editScenario(t)
			req := imagesInput(t, sc)
			o := imagesOptions(t, sc.Options).(ai.OpenAIImagesOptions)
			w := newWorldWith(t, func(c *ai.Config) {
				configureImages(c)
				bindings := c.Bindings
				c.Bindings = bindingResolverFunc(func(ctx context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
					req.ReferenceImages[0].Data = "caller mutation"
					o.Mask.Data = "caller mask mutation"
					return bindings.ResolveBinding(ctx, scope, id)
				})
			}, tenantA)
			installImagesBinding(w, tenantA)
			enqueue(ev, w, sc.replies(t)...)
			var opts ai.ImageOptions = o
			if pointer {
				opts = &o
			}
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-edit-snapshot"), sc.target(), req, opts)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Check("entry data independently frozen", err == nil && len(w.provider.Requests()) == 1 && jsonEqual(w.provider.Requests()[0].Body, sc.Expect.Request.Body), "got %v", err)
		})
	}
}
