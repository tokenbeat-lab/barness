package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGoogleImagesBuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "P09-E03-live-catalog-inclusion")
	raw, err := os.ReadFile("testdata/google-images/official.json")
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
	m, ok := cat.LookupImage(ai.ProviderGoogle, ai.APIGoogleInteractions, "gemini-nano-banana-2.1")
	ev.Check("own live-confirmed model, capabilities and rates", ok && reflect.DeepEqual(m, facts.Model), "model=%+v", m)
	ev.Check("dated source with new content version", facts.Retrieved == "2026-10-08" && cat.Version == "2026-10-08.5", "version=%s", cat.Version)
	for _, id := range []string{"gemini-nano-banana-2.1", "gemini-3.1-flash-image", "gemini-2.5-flash-image", "imagen-4"} {
		_, chat := cat.Lookup(ai.ProviderGoogle, ai.APIGoogleGenerativeAI, id)
		ev.Check("image models cannot leak into chat "+id, !chat, "chat model listed")
		if id != "gemini-nano-banana-2.1" {
			_, included := cat.LookupImage(ai.ProviderGoogle, ai.APIGoogleInteractions, id)
			ev.Check("excluded model stays absent "+id, !included, "excluded model listed")
		}
	}
	ev.Check("only verified capabilities and applicable token rates", m.Capabilities.MaxReferenceImages == 1 && m.Capabilities.MaxOutputImages == 1 && reflect.DeepEqual(m.Capabilities.ImageSizes, []string{"1K"}) && reflect.DeepEqual(m.Capabilities.AspectRatios, []string{"1:1"}) && m.Pricing.CacheReadText.IsZero() && m.Pricing.CacheReadImage.IsZero(), "unverified capability or cache price")
}

func TestGoogleImagesLiveWireReplay(t *testing.T) {
	raw, err := os.ReadFile("testdata/google-images/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			ID              string
			Request         json.RawMessage
			Status          int
			ResponseHeaders map[string]string
			Response        json.RawMessage
			Usage           ai.Usage
			UsageReporting  ai.UsageReporting
			Validation      struct {
				SHA256        string
				Bytes         int
				Format        string
				Width, Height int
			}
		}
	}
	mustUnmarshal(t, raw, &fixture)
	for _, tc := range fixture.Cases {
		for _, variant := range []string{"captured", "partial", "unreported"} {
			t.Run(tc.ID+"/"+variant, func(t *testing.T) {
				ev := run.Case(t, "P09-E01-live-replay-"+tc.ID+"-"+variant)
				ev.Fixture("live-contract.json", raw)
				w := newWorldWith(t, func(c *ai.Config) {
					c.Policy.MaxOutputBytes = 16 << 20
					c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 1, MaxOutputImages: 1, MaxOutputImageBytes: 8 << 20, MaxTotalOutputImageBytes: 8 << 20}
				}, tenantA)
				b := geminiBinding(tenantA, w.provider.URL())
				b.BindingID, b.API, b.Operation = "images", ai.APIGoogleInteractions, ai.OperationImage
				b.AllowedModels = []string{"gemini-nano-banana-2.1"}
				w.host.PutBinding(b)
				var wire struct {
					Input []struct {
						Type, Text, Data string
						MimeType         string `json:"mime_type"`
					}
					Format struct {
						AspectRatio string `json:"aspect_ratio"`
						ImageSize   string `json:"image_size"`
					} `json:"response_format"`
				}
				mustUnmarshal(t, tc.Request, &wire)
				req := ai.ImagesRequest{Prompt: wire.Input[0].Text}
				for _, ref := range wire.Input[1:] {
					req.ReferenceImages = append(req.ReferenceImages, ai.Image{Data: ref.Data, MimeType: ref.MimeType})
				}
				opts := ai.GoogleImagesOptions{AspectRatio: ai.Value(wire.Format.AspectRatio), ImageSize: ai.Value(wire.Format.ImageSize)}
				response := tc.Response
				expected := tc.Usage
				reporting := tc.UsageReporting
				if variant != "captured" {
					var fields map[string]json.RawMessage
					mustUnmarshal(t, response, &fields)
					if variant == "unreported" {
						delete(fields, "usage")
						expected = ai.Usage{}
						reporting = ai.UsageUnreported
					} else {
						var usage map[string]json.RawMessage
						mustUnmarshal(t, fields["usage"], &usage)
						delete(usage, "total_tokens")
						fields["usage"] = mustMarshal(t, usage)
						expected.TotalTokens = expected.Input + expected.Output
						reporting = ai.UsagePartial
					}
					response = mustMarshal(t, fields)
				}
				enqueue(ev, w, provider.Reply{Status: tc.Status, Header: tc.ResponseHeaders, Chunks: [][]byte{response}, Script: string(response)})
				res, err := w.client.GenerateImages(ctxFor(t), textScope("req-google-live-"+tc.ID+variant), ai.Target{BindingID: b.BindingID, ModelID: b.AllowedModels[0]}, req, opts)
				ev.Record("result", res)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("confirmed native response accepted offline", err == nil && len(res.Content) == 1, "err=%v", err) {
					return
				}
				sent := w.provider.Requests()
				ev.Check("same native request, v1beta and omitted delivery", len(sent) == 1 && sent[0].Path == "/v1beta/interactions" && jsonEqual(sent[0].Body, tc.Request), "request mismatch")
				block := res.Content[0].(ai.ImageOutputImage)
				data, err := base64.StdEncoding.Strict().DecodeString(block.Data)
				if err != nil {
					t.Fatal(err)
				}
				config, format, err := image.DecodeConfig(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				raster, _, err := image.Decode(bytes.NewReader(data))
				sum := sha256.Sum256(data)
				ev.Check("raster remains fully readable and matches saved artifact", err == nil && format == tc.Validation.Format && block.MimeType == "image/"+format && config.Width == 1024 && config.Height == 1024 && raster.Bounds().Dx() == 1024 && raster.Bounds().Dy() == 1024 && len(data) == tc.Validation.Bytes && hex.EncodeToString(sum[:]) == tc.Validation.SHA256, "invalid artifact")
				ev.Check("actual identity presence, not attribution fallback", res.ResponseID.IsZero() && res.ResponseModel == ai.Value("gemini-nano-banana-2.1"), "identity invented or lost")
				ev.Check("usage totals, nullable modalities and partial semantics", reflect.DeepEqual(res.Usage, expected) && res.Metadata.Attempts[0].UsageReporting == reporting, "usage=%+v want=%+v", res.Usage, expected)
				if variant == "captured" {
					// Actual responses report only image output detail; thought is text and
					// excluded. Missing non-thought output text remains unknown, never zero.
					thought, _ := res.Usage.Reasoning.Get()
					imageTokens, _ := res.Usage.Modalities.OutputImage.Get()
					ev.Check("thinking billed once and missing text stays unknown", res.Usage.Modalities.OutputText.IsZero() && math.Abs(res.Usage.Cost.Output-(float64(imageTokens)*30+float64(thought)*7.5)/1e6) < 1e-12, "thought double-count or invented text")
				}
				ev.Check("updated immutable price snapshot", res.Metadata.CatalogVersion == "2026-10-08.5" && res.Metadata.CatalogHash == builtinCatalogHash, "wrong snapshot")
			})
		}
	}
}
