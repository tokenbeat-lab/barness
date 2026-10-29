package ai

import "slices"

// XAIImagesOptions selects the synchronous global Imagine image operation.
// Omitted quality and resolution retain vendor defaults; null is invalid.
type XAIImagesOptions struct {
	N           Nullable[int]    `json:"n,omitzero"`
	AspectRatio Nullable[string] `json:"aspectRatio,omitzero"`
	Resolution  Nullable[string] `json:"resolution,omitzero"`
	Quality     Nullable[string] `json:"quality,omitzero"`
}

func (XAIImagesOptions) imageOptions()    {}
func (XAIImagesOptions) api() API         { return APIXAIImages }
func (o XAIImagesOptions) clone() Options { return o }
func (o XAIImagesOptions) quantity() int {
	if n, ok := o.N.Get(); ok {
		return n
	}
	return 1
}
func (o XAIImagesOptions) validate() string {
	if o.N.IsNull() || o.AspectRatio.IsNull() || o.Resolution.IsNull() || o.Quality.IsNull() {
		return "Imagine options require omitted fields or values"
	}
	if n := o.quantity(); n < 1 || n > 10 {
		return "Imagine quantity must be 1 to 10"
	}
	if r, ok := o.Resolution.Get(); ok && !slices.Contains([]string{"1k", "1.5k", "2k"}, r) {
		return "unsupported Imagine resolution"
	}
	if q, ok := o.Quality.Get(); ok && !slices.Contains([]string{"low", "medium", "auto"}, q) {
		return "unsupported Imagine quality"
	}
	if a, ok := o.AspectRatio.Get(); ok && !slices.Contains([]string{"1:1", "3:4", "4:3", "9:16", "16:9", "2:3", "3:2", "9:19.5", "19.5:9", "9:20", "20:9", "1:2", "2:1", "21:9", "5:2", "auto"}, a) {
		return "unsupported Imagine aspect ratio"
	}
	return ""
}
func (o XAIImagesOptions) check(c ImageCapabilities, p *ImagePolicy, refs int) string {
	if problem := o.validate(); problem != "" {
		return problem
	}
	if o.quantity() > min(c.MaxOutputImages, p.MaxOutputImages) {
		return "Imagine quantity exceeds the effective output limit"
	}
	if refs > 0 && o.quantity() != 1 {
		return "Imagine editing requires one output"
	}
	if refs == 1 && !o.AspectRatio.IsZero() {
		return "single-image editing requires omitted aspect ratio"
	}
	for _, v := range []struct {
		value   Nullable[string]
		allowed []string
	}{{o.Resolution, c.ImageSizes}, {o.Quality, c.Qualities}, {o.AspectRatio, c.AspectRatios}} {
		if x, ok := v.value.Get(); ok && !slices.Contains(v.allowed, x) {
			return "Imagine option is not authorized by the model"
		}
	}
	return ""
}
