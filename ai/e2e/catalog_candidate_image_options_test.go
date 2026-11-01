package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func catalogImageOptionsWire(t *testing.T, ev *evidence.Case, raw []byte, opts ai.ImageOptions) {
	var fields map[string]json.RawMessage
	mustUnmarshal(t, raw, &fields)
	check := func(key string, value any) {
		want := mustMarshal(t, value)
		ev.Check("declared image option wire "+key, jsonEqual(fields[key], want), "got=%s want=%s", fields[key], want)
	}
	switch o := opts.(type) {
	case ai.OpenAIImagesOptions:
		if v, ok := o.N.Get(); ok {
			check("n", v)
		}
		if v, ok := o.Size.Get(); ok {
			check("size", v)
		}
		if v, ok := o.Quality.Get(); ok {
			check("quality", v)
		}
		if v, ok := o.Background.Get(); ok {
			check("background", v)
		}
		if v, ok := o.InputFidelity.Get(); ok {
			check("input_fidelity", v)
		}
		if o.Mask != nil {
			ev.Check("declared mask reaches wire", len(fields["mask"]) > 0, "%s", raw)
		}
	case ai.GoogleImagesOptions:
		var format map[string]json.RawMessage
		mustUnmarshal(t, fields["response_format"], &format)
		fields = format
		if v, ok := o.ImageSize.Get(); ok {
			check("image_size", v)
		}
		if v, ok := o.AspectRatio.Get(); ok {
			check("aspect_ratio", v)
		}
	case ai.XAIImagesOptions:
		if v, ok := o.N.Get(); ok {
			check("n", v)
		}
		if v, ok := o.Resolution.Get(); ok {
			check("resolution", v)
		}
		if v, ok := o.Quality.Get(); ok {
			check("quality", v)
		}
		if v, ok := o.AspectRatio.Get(); ok {
			check("aspect_ratio", v)
		}
	}
}
func catalogImageCustomSizes(t *testing.T, ev *evidence.Case, w *world, i catalogContractIdentity, m ai.ImageModel, req ai.ImagesRequest, call func(ai.ImagesRequest, ai.ImageOptions, int)) {
	c := m.Capabilities.CustomSizes
	side := int(math.Ceil(math.Sqrt(float64(c.MinPixels))))
	if rem := side % c.EdgeMultiple; rem != 0 {
		side += c.EdgeMultiple - rem
	}
	if side < 1 || side > c.MaxEdge || int64(side) > c.MaxPixels/int64(side) || int64(side)*int64(side) < c.MinPixels {
		t.Fatal("code_gap: declared custom image constraints have no square contract fixture")
	}
	call(req, ai.OpenAIImagesOptions{Size: ai.Value(fmt.Sprintf("%dx%d", side, side))}, 1)
	for _, size := range []string{"0x1", fmt.Sprintf("%dx%d", c.MaxEdge+1, c.MaxEdge+1)} {
		before := len(w.provider.Requests())
		_, err := w.client.GenerateImages(ctxFor(t), textScope("catalog-custom-size-outside"), catalogContractTarget(i), req, ai.OpenAIImagesOptions{Size: ai.Value(size)})
		ev.Check("custom dimension boundary refused before send", err != nil && len(w.provider.Requests()) == before, "size=%s err=%v", size, err)
	}
}
