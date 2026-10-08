package ai

import (
	"bytes"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

// OpenAIImagesOptions exposes synchronous generation and editing settings. Model
// capabilities authorize sizes/qualities/transparency; resource budgets
// always come from ImagePolicy. Mask and InputFidelity require references.
type OpenAIImagesOptions struct {
	N                 Nullable[int]    `json:"n,omitzero"`
	Size              Nullable[string] `json:"size,omitzero"`
	Quality           Nullable[string] `json:"quality,omitzero"`
	Background        Nullable[string] `json:"background,omitzero"`
	OutputFormat      Nullable[string] `json:"outputFormat,omitzero"`
	OutputCompression Nullable[int]    `json:"outputCompression,omitzero"`
	Moderation        Nullable[string] `json:"moderation,omitzero"`
	Mask              *Image           `json:"mask,omitempty"`
	InputFidelity     Nullable[string] `json:"inputFidelity,omitzero"`
}

func (OpenAIImagesOptions) imageOptions() {}
func (OpenAIImagesOptions) api() API      { return APIOpenAIImages }
func (o OpenAIImagesOptions) clone() Options {
	if o.Mask != nil {
		mask := *o.Mask
		o.Mask = &mask
	}
	return o
}
func (o OpenAIImagesOptions) validate() string {
	for _, null := range []bool{o.N.IsNull(), o.Size.IsNull(), o.Quality.IsNull(), o.Background.IsNull(), o.OutputFormat.IsNull(), o.OutputCompression.IsNull(), o.Moderation.IsNull(), o.InputFidelity.IsNull()} {
		if null {
			return "image options must be omitted or contain a value"
		}
	}
	if n, ok := o.N.Get(); ok && (n < 1 || n > 10) {
		return "image quantity must be 1 to 10"
	}
	if f, ok := o.OutputFormat.Get(); ok && !slices.Contains([]string{"png", "jpeg", "webp"}, f) {
		return "unsupported image output format"
	}
	if b, ok := o.Background.Get(); ok && !slices.Contains([]string{"auto", "opaque", "transparent"}, b) {
		return "unsupported image background"
	}
	if q, ok := o.Quality.Get(); ok && !slices.Contains([]string{"auto", "low", "medium", "high", "xhigh", "max"}, q) {
		return "unsupported image quality"
	}
	if m, ok := o.Moderation.Get(); ok && !slices.Contains([]string{"auto", "low"}, m) {
		return "unsupported image moderation"
	}
	if f, ok := o.InputFidelity.Get(); ok && !slices.Contains([]string{"high", "low"}, f) {
		return "unsupported image input fidelity"
	}
	format := o.outputFormat()
	if b, _ := o.Background.Get(); b == "transparent" && format != "png" && format != "webp" {
		return "transparent backgrounds require png or webp"
	}
	if compression, ok := o.OutputCompression.Get(); ok {
		if compression < 0 || compression > 100 {
			return "image compression must be 0 to 100"
		}
		if format != "jpeg" && format != "webp" {
			return "image compression requires jpeg or webp"
		}
	}
	return ""
}
func (o OpenAIImagesOptions) check(c ImageCapabilities, p *ImagePolicy) string {
	if problem := o.validate(); problem != "" {
		return problem
	}
	if o.quantity() > min(c.MaxOutputImages, p.MaxOutputImages) {
		return "image quantity exceeds the effective output limit"
	}
	if s, ok := o.Size.Get(); ok {
		if s == "auto" || c.CustomSizes == nil {
			if !slices.Contains(c.Sizes, s) {
				return "model does not support this image size"
			}
		} else if !c.CustomSizes.allows(s) {
			return "model does not support this image size"
		}
	}
	if q, ok := o.Quality.Get(); ok && !slices.Contains(c.Qualities, q) {
		return "model does not support this image quality"
	}
	if b, _ := o.Background.Get(); b == "transparent" && !c.TransparentBackground {
		return "model does not support transparent backgrounds"
	}
	if o.Mask != nil && !c.Mask {
		return "model does not support masks"
	}
	if f, ok := o.InputFidelity.Get(); ok && !slices.Contains(c.InputFidelity, f) {
		return "model does not support this input fidelity"
	}
	return ""
}
func (o OpenAIImagesOptions) quantity() int {
	if n, ok := o.N.Get(); ok {
		return n
	}
	return 1
}
func (o OpenAIImagesOptions) outputFormat() string {
	if f, ok := o.OutputFormat.Get(); ok {
		return f
	}
	return "png"
}

func (c ImageSizeConstraints) allows(size string) bool {
	w, h, ok := strings.Cut(size, "x")
	if !ok {
		return false
	}
	for _, part := range []string{w, h} {
		if part == "" {
			return false
		}
		for _, b := range []byte(part) {
			if b < '0' || b > '9' {
				return false
			}
		}
	}
	width, errW := strconv.Atoi(w)
	height, errH := strconv.Atoi(h)
	if errW != nil || errH != nil || width < 1 || height < 1 || width > c.MaxEdge || height > c.MaxEdge || width%c.EdgeMultiple != 0 || height%c.EdgeMultiple != 0 {
		return false
	}
	// Division avoids overflow for even unusually large host constraints.
	w64, h64 := int64(width), int64(height)
	if w64 > c.MaxPixels/h64 || (max(w64, h64)-1)/min(w64, h64) >= int64(c.MaxAspectRatio) {
		return false
	}
	return w64*h64 >= c.MinPixels
}

type openAIImagesRequest struct {
	Model             string                      `json:"model"`
	Prompt            string                      `json:"prompt"`
	N                 Nullable[int]               `json:"n,omitzero"`
	Size              Nullable[string]            `json:"size,omitzero"`
	Quality           Nullable[string]            `json:"quality,omitzero"`
	Background        Nullable[string]            `json:"background,omitzero"`
	OutputFormat      Nullable[string]            `json:"output_format,omitzero"`
	OutputCompression Nullable[int]               `json:"output_compression,omitzero"`
	Moderation        Nullable[string]            `json:"moderation,omitzero"`
	Images            []openAIInlineImage         `json:"images,omitempty"`
	Mask              Nullable[openAIInlineImage] `json:"mask,omitzero"`
	InputFidelity     Nullable[string]            `json:"input_fidelity,omitzero"`
}

type openAIInlineImage struct {
	ImageURL string `json:"image_url"`
}

func (w openAIImagesRequest) options() OpenAIImagesOptions {
	o := OpenAIImagesOptions{N: w.N, Size: w.Size, Quality: w.Quality, Background: w.Background, OutputFormat: w.OutputFormat, OutputCompression: w.OutputCompression, Moderation: w.Moderation, InputFidelity: w.InputFidelity}
	if ref, ok := w.Mask.Get(); ok {
		mask, _ := parseOpenAIImageURL(ref.ImageURL)
		o.Mask = &mask
	}
	return o
}

func decodeOpenAIImagesRequest(body []byte) (openAIImagesRequest, error) {
	var wire openAIImagesRequest
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	err := d.Decode(&wire)
	return wire, err
}
