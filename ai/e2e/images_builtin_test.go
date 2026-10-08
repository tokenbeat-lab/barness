package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image"
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Official facts and this route's own live acceptance are separate from the
// synthetic protocol fixtures. These tests use the existing public Client seam.
func TestOpenAIImagesBuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "P08-E03-live-catalog-inclusion")
	raw, err := os.ReadFile("testdata/openai-images/official.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("official.json", raw)
	var facts struct {
		Retrieved string
		Model     ai.ImageModel
	}
	mustUnmarshal(t, raw, &facts)
	cat := ai.BuiltinCatalog()
	ev.Record("catalog", cat)
	model, ok := cat.LookupImage(ai.ProviderOpenAI, ai.APIOpenAIImages, "gpt-image-2.5-sunburst-2026-09-08")
	ev.Check("fixed model equals sourced verified capability and price snapshot", ok && reflect.DeepEqual(model, facts.Model), "got %+v", model)
	ev.Check("dated source and new version", facts.Retrieved == "2026-10-08" && cat.Version == "2026-10-08.5", "version=%s", cat.Version)
	_, alias := cat.LookupImage(ai.ProviderOpenAI, ai.APIOpenAIImages, "gpt-image-2.5-sunburst")
	ev.Check("moving alias not included", !alias, "alias listed")
	ev.Check("no cached or derived per-image billing", model.Pricing.CacheReadText.IsZero() && model.Pricing.CacheReadImage.IsZero() && len(model.Capabilities.InputFidelity) == 0, "unconfirmed configuration")
}

func TestOpenAIImagesLiveWireReplay(t *testing.T) {
	raw, err := os.ReadFile("testdata/openai-images/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			ID              string            `json:"id"`
			Request         json.RawMessage   `json:"request"`
			Status          int               `json:"status"`
			ResponseHeaders map[string]string `json:"responseHeaders"`
			Response        json.RawMessage   `json:"response"`
			Usage           ai.Usage          `json:"usage"`
			UsageReporting  ai.UsageReporting `json:"usageReporting"`
			Validation      struct {
				SHA256 string `json:"sha256"`
				Bytes  int    `json:"bytes"`
				Format string `json:"format"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"validation"`
		} `json:"cases"`
	}
	mustUnmarshal(t, raw, &fixture)
	for _, tc := range fixture.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			ev := run.Case(t, "P08-E01-live-replay-"+tc.ID)
			ev.Fixture("live-contract.json", raw)
			w := newWorldWith(t, func(c *ai.Config) {
				c.Policy.MaxOutputBytes = 16 << 20
				c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 2, MaxOutputImages: 1, MaxOutputImageBytes: 8 << 20, MaxTotalOutputImageBytes: 8 << 20}
			}, tenantA)
			b := primaryBinding(tenantA, w.provider.URL())
			b.BindingID, b.API, b.Operation = "images", ai.APIOpenAIImages, ai.OperationImage
			b.AllowedModels = []string{"gpt-image-2.5-sunburst-2026-09-08"}
			w.host.PutBinding(b)
			var wire struct {
				Prompt string
				Images []struct {
					ImageURL string `json:"image_url"`
				}
				Mask *struct {
					ImageURL string `json:"image_url"`
				}
				N            int
				Size         string
				Quality      string
				OutputFormat string `json:"output_format"`
				Background   string
			}
			mustUnmarshal(t, tc.Request, &wire)
			req := ai.ImagesRequest{Prompt: wire.Prompt}
			for _, ref := range wire.Images {
				req.ReferenceImages = append(req.ReferenceImages, inlineLiveImage(t, ref.ImageURL))
			}
			opts := ai.OpenAIImagesOptions{N: ai.Value(wire.N), Size: ai.Value(wire.Size), Quality: ai.Value(wire.Quality), OutputFormat: ai.Value(wire.OutputFormat)}
			if wire.Background != "" {
				opts.Background = ai.Value(wire.Background)
			}
			if wire.Mask != nil {
				mask := inlineLiveImage(t, wire.Mask.ImageURL)
				opts.Mask = &mask
			}
			enqueue(ev, w, provider.Reply{Status: tc.Status, Header: tc.ResponseHeaders, Chunks: [][]byte{tc.Response}, Script: string(tc.Response)})
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-live-image-"+tc.ID), ai.Target{BindingID: b.BindingID, ModelID: b.AllowedModels[0]}, req, opts)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("live response accepted offline", err == nil && len(res.Content) == 1, "err=%v", err) {
				return
			}
			requests := w.provider.Requests()
			path := "/v1/images/edits"
			if tc.ID == "generation" {
				path = "/v1/images/generations"
			}
			ev.Check("complete native request and original route", len(requests) == 1 && requests[0].Path == path && jsonEqual(requests[0].Body, tc.Request), "native request mismatch")
			block := res.Content[0].(ai.ImageOutputImage)
			data, err := base64.StdEncoding.Strict().DecodeString(block.Data)
			if err != nil {
				t.Fatal(err)
			}
			header, format, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			decoded, _, err := image.Decode(bytes.NewReader(data))
			sum := sha256.Sum256(data)
			ev.Check("complete raster, MIME, dimensions and captured checksum", err == nil && format == tc.Validation.Format && block.MimeType == "image/"+format && header.Width == 1024 && header.Height == 1024 && decoded.Bounds().Dx() == 1024 && decoded.Bounds().Dy() == 1024 && len(data) == tc.Validation.Bytes && hex.EncodeToString(sum[:]) == tc.Validation.SHA256, "invalid recorded image")
			ev.Check("actual modality usage and known-only prices preserved", reflect.DeepEqual(res.Usage, tc.Usage) && res.Metadata.Attempts[0].UsageReporting == tc.UsageReporting, "got %+v", res.Usage)
			ev.Check("current catalog snapshot attributed", res.Metadata.CatalogVersion == "2026-10-08.5" && res.Metadata.CatalogHash == builtinCatalogHash, "wrong snapshot")
		})
	}
}

func inlineLiveImage(t *testing.T, url string) ai.Image {
	t.Helper()
	const prefix = "data:image/png;base64,"
	if len(url) <= len(prefix) || url[:len(prefix)] != prefix {
		t.Fatal("fixture must use inline PNG")
	}
	return ai.Image{MimeType: "image/png", Data: url[len(prefix):]}
}
