package ai

import "slices"

// GoogleImagesOptions requests an aspect ratio and resolution authorized by
// the ImageModel. Nil uses Google's defaults. ImagePolicy bounds the returned
// collection; this protocol does not promise an exact number of images.
type GoogleImagesOptions struct {
	AspectRatio Nullable[string] `json:"aspectRatio,omitzero"`
	ImageSize   Nullable[string] `json:"imageSize,omitzero"`
}

func (GoogleImagesOptions) imageOptions()    {}
func (GoogleImagesOptions) api() API         { return APIGoogleInteractions }
func (o GoogleImagesOptions) clone() Options { return o }
func (o GoogleImagesOptions) validate() string {
	if o.AspectRatio.IsNull() || o.ImageSize.IsNull() {
		return "Google image options must be omitted or contain a value"
	}
	return ""
}
func (o GoogleImagesOptions) check(c ImageCapabilities) string {
	if problem := o.validate(); problem != "" {
		return problem
	}
	if ratio, ok := o.AspectRatio.Get(); ok && !slices.Contains(c.AspectRatios, ratio) {
		return "model does not support this image aspect ratio"
	}
	if size, ok := o.ImageSize.Get(); ok && !slices.Contains(c.ImageSizes, size) {
		return "model does not support this image size"
	}
	return ""
}
