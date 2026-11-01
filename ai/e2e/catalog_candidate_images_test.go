package e2e

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func catalogImageContract(t *testing.T, ev *evidence.Case, cat ai.Catalog, i catalogContractIdentity, scene string) {
	if scene == "load-query" {
		catalogContractQuery(ev, cat, i)
		return
	}
	m, _ := cat.LookupImage(i.Provider, i.API, i.ID)
	w, _ := newCatalogContractWorld(t, cat, i)
	req := ai.ImagesRequest{Prompt: "Catalog image."}
	if scene == "authorization" {
		catalogContractDenied(ev, w, i, func(scope ai.CallScope, target ai.Target) error {
			_, err := w.client.GenerateImages(ctxFor(t), scope, target, req, nil)
			return err
		})
		return
	}
	call := func(req ai.ImagesRequest, opts ai.ImageOptions, n int) {
		enqueue(ev, w, catalogImageReply(t, i, req, n))
		result, err := w.client.GenerateImages(ctxFor(t), textScope("catalog-image"), catalogContractTarget(i), req, opts)
		ev.Record("result", result)
		ev.Record("error", errString(err))
		ev.Check("declared image request completes", err == nil && len(result.Content) >= n, "%v %+v", err, result)
		if err != nil {
			return
		}
		catalogContractMetadata(ev, cat, i, result.Metadata)
		requests := w.provider.Requests()
		body := requests[len(requests)-1].Body
		var fields map[string]json.RawMessage
		mustUnmarshal(t, body, &fields)
		var id string
		_ = json.Unmarshal(fields["model"], &id)
		ev.Check("exact image wire identity", id == i.ID, "%s", body)
		if i.API == ai.APIXAIImages {
			catalogImagePerImage(ev, m, opts, len(req.ReferenceImages), n, result)
		} else {
			text, image, out, thought := float64(10), float64(0), float64(20), float64(0)
			if len(req.ReferenceImages) > 0 {
				text, image = 9, 1
			}
			if i.API == ai.APIGoogleInteractions {
				text, image, out, thought = 7, 0, 20, 22
				if len(req.ReferenceImages) > 0 {
					text, image = 6, 1
				}
			}
			a, _ := m.Pricing.InputText.Get()
			b, _ := m.Pricing.InputImage.Get()
			c, _ := m.Pricing.OutputImage.Get()
			d, _ := m.Pricing.OutputText.Get()
			expected := (text*a + image*b + out*c + thought*d) / 1e6
			ev.Check("independent modality pricing", math.Abs(result.Usage.Cost.Total-expected) < 1e-12, "got=%+v expected=%g", result.Usage, expected)
			ev.Check("complete image usage", len(result.Metadata.Attempts) == 1 && result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", result.Metadata.Attempts)
		}
	}
	if scene == "generation" {
		call(req, nil, 1)
	}
	if scene == "options" {
		switch i.API {
		case ai.APIOpenAIImages:
			for _, size := range m.Capabilities.Sizes {
				for _, quality := range m.Capabilities.Qualities {
					call(req, ai.OpenAIImagesOptions{Size: ai.Value(size), Quality: ai.Value(quality)}, 1)
				}
			}
			if m.Capabilities.TransparentBackground {
				call(req, ai.OpenAIImagesOptions{Background: ai.Value("transparent")}, 1)
			}
			if m.Capabilities.MaxOutputImages > 1 {
				call(req, ai.OpenAIImagesOptions{N: ai.Value(m.Capabilities.MaxOutputImages)}, m.Capabilities.MaxOutputImages)
			}
		case ai.APIGoogleInteractions:
			for _, ratio := range m.Capabilities.AspectRatios {
				for _, size := range m.Capabilities.ImageSizes {
					call(req, ai.GoogleImagesOptions{AspectRatio: ai.Value(ratio), ImageSize: ai.Value(size)}, 1)
				}
			}
		case ai.APIXAIImages:
			for _, ratio := range m.Capabilities.AspectRatios {
				for _, size := range m.Capabilities.ImageSizes {
					for _, quality := range m.Capabilities.Qualities {
						call(req, ai.XAIImagesOptions{AspectRatio: ai.Value(ratio), Resolution: ai.Value(size), Quality: ai.Value(quality)}, 1)
					}
				}
			}
			call(req, ai.XAIImagesOptions{N: ai.Value(m.Capabilities.MaxOutputImages)}, m.Capabilities.MaxOutputImages)
		default:
			t.Fatalf("code_gap: no image contract %s", i.API)
		}
	}
	if scene == "edit-boundaries" {
		if m.Capabilities.MaxReferenceImages > 0 {
			req.ReferenceImages = make([]ai.Image, m.Capabilities.MaxReferenceImages)
			for n := range req.ReferenceImages {
				req.ReferenceImages[n] = ai.Image{Data: catalogContractPNG, MimeType: "image/png"}
			}
			call(req, nil, 1)
			if i.API == ai.APIOpenAIImages {
				for _, fidelity := range m.Capabilities.InputFidelity {
					call(req, ai.OpenAIImagesOptions{InputFidelity: ai.Value(fidelity)}, 1)
				}
				if m.Capabilities.Mask {
					mask := req.ReferenceImages[0]
					call(req, ai.OpenAIImagesOptions{Mask: &mask}, 1)
				}
			}
			req.ReferenceImages = append(req.ReferenceImages, req.ReferenceImages[0])
			before := len(w.provider.Requests())
			_, err := w.client.GenerateImages(ctxFor(t), textScope("catalog-too-many-references"), catalogContractTarget(i), req, nil)
			ev.Check("reference bound refuses before send", err != nil && len(w.provider.Requests()) == before, "%v", err)
		}
		before := len(w.provider.Requests())
		var invalid ai.ImageOptions
		switch i.API {
		case ai.APIOpenAIImages:
			invalid = ai.OpenAIImagesOptions{N: ai.Value(m.Capabilities.MaxOutputImages + 1)}
		case ai.APIXAIImages:
			invalid = ai.XAIImagesOptions{N: ai.Value(m.Capabilities.MaxOutputImages + 1)}
		case ai.APIGoogleInteractions:
			invalid = ai.GoogleImagesOptions{ImageSize: ai.Value("unsupported-catalog-size")}
		}
		_, err := w.client.GenerateImages(ctxFor(t), textScope("catalog-bad-image-options"), catalogContractTarget(i), ai.ImagesRequest{Prompt: "Catalog image."}, invalid)
		ev.Check("invalid declared bounds refuse before send", err != nil && len(w.provider.Requests()) == before, "%v", err)
	}
	ev.Record("requests", w.provider.Requests())
}
func catalogImageReply(t *testing.T, i catalogContractIdentity, req ai.ImagesRequest, n int) provider.Reply {
	var body string
	data := strings.Repeat(`{"b64_json":"`+catalogContractPNG+`"},`, n)
	data = strings.TrimSuffix(data, ",")
	if i.API == ai.APIGoogleInteractions {
		text, image := 7, 0
		if len(req.ReferenceImages) > 0 {
			text, image = 6, 1
		}
		body = string(mustMarshal(t, map[string]any{"id": "catalog-image", "model": i.ID, "status": "completed", "steps": []any{map[string]any{"type": "model_output", "content": []any{map[string]any{"type": "image", "mime_type": "image/png", "data": catalogContractPNG}}}}, "usage": map[string]any{"total_input_tokens": 7, "total_output_tokens": 20, "total_thought_tokens": 22, "total_tokens": 49, "total_cached_tokens": 0, "input_tokens_by_modality": []any{map[string]any{"modality": "text", "tokens": text}, map[string]any{"modality": "image", "tokens": image}}, "output_tokens_by_modality": []any{map[string]any{"modality": "text", "tokens": 0}, map[string]any{"modality": "image", "tokens": 20}}}}))
	} else if i.API == ai.APIXAIImages {
		body = `{"data":[` + data + `],"usage":{"cost_in_usd_ticks":400000000}}`
	} else {
		text, image := 10, 0
		if len(req.ReferenceImages) > 0 {
			text, image = 9, 1
		}
		body = string(mustMarshal(t, map[string]any{"data": json.RawMessage("[" + data + "]"), "usage": map[string]any{"input_tokens": 10, "output_tokens": 20, "total_tokens": 30, "input_tokens_details": map[string]any{"text_tokens": text, "image_tokens": image}, "output_tokens_details": map[string]any{"text_tokens": 0, "image_tokens": 20}}}))
	}
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Script: body, Chunks: [][]byte{[]byte(body)}}
}
func catalogImagePerImage(ev *evidence.Case, m ai.ImageModel, opts ai.ImageOptions, refs, n int, result ai.ImagesResult) {
	ev.Check("per-image never invents token cost", result.Usage.Cost.Total == 0, "%+v", result.Usage)
	if len(result.Metadata.Attempts) != 1 {
		return
	}
	attempt := result.Metadata.Attempts[0]
	charge, known := attempt.ReportedCostUSD.Get()
	ev.Check("reported charge separately retained", known && charge == 0.04, "%+v", attempt)
	selected, ok := opts.(ai.XAIImagesOptions)
	if !ok {
		ev.Check("omitted dimensions keep estimate partial", attempt.CostEstimate.Status == ai.CostPartial, "%+v", attempt.CostEstimate)
		return
	}
	size, s := selected.Resolution.Get()
	quality, q := selected.Quality.Get()
	if !s || !q {
		return
	}
	low, high := math.Inf(1), math.Inf(-1)
	for _, row := range m.Pricing.PerImage.Output {
		if row.ImageSize == size && (quality == "auto" || row.Quality == quality) {
			price, _ := row.Price.Get()
			low = min(low, price)
			high = max(high, price)
		}
	}
	input, _ := m.Pricing.PerImage.InputImage.Get()
	wantLow, wantHigh := input*float64(refs)+low*float64(n), input*float64(refs)+high*float64(n)
	lo, lok := attempt.CostEstimate.Minimum.Get()
	hi, hok := attempt.CostEstimate.Maximum.Get()
	ev.Check("independent per-image range", lok && hok && math.Abs(lo.Total-wantLow) < 1e-12 && math.Abs(hi.Total-wantHigh) < 1e-12, "got=%+v low=%g high=%g", attempt.CostEstimate, wantLow, wantHigh)
}
