package ai

import (
	"bytes"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// OpenAIImagesOptions exposes only synchronous generation settings. Model
// capabilities authorize sizes/qualities/transparency; resource budgets
// always come from ImagePolicy. Editing settings belong to issue 10.
type OpenAIImagesOptions struct {
	N                 Nullable[int]    `json:"n,omitzero"`
	Size              Nullable[string] `json:"size,omitzero"`
	Quality           Nullable[string] `json:"quality,omitzero"`
	Background        Nullable[string] `json:"background,omitzero"`
	OutputFormat      Nullable[string] `json:"outputFormat,omitzero"`
	OutputCompression Nullable[int]    `json:"outputCompression,omitzero"`
	Moderation        Nullable[string] `json:"moderation,omitzero"`
}

func (OpenAIImagesOptions) imageOptions()    {}
func (OpenAIImagesOptions) api() API         { return APIOpenAIImages }
func (o OpenAIImagesOptions) clone() Options { return o }
func (o OpenAIImagesOptions) validate() string {
	for _, null := range []bool{o.N.IsNull(), o.Size.IsNull(), o.Quality.IsNull(), o.Background.IsNull(), o.OutputFormat.IsNull(), o.OutputCompression.IsNull(), o.Moderation.IsNull()} {
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
	if s, ok := o.Size.Get(); ok && !slices.Contains(c.Sizes, s) && !(c.CustomSizes && openAIImageDimensions(s)) {
		return "model does not support this image size"
	}
	if q, ok := o.Quality.Get(); ok && !slices.Contains(c.Qualities, q) {
		return "model does not support this image quality"
	}
	if b, _ := o.Background.Get(); b == "transparent" && !c.TransparentBackground {
		return "model does not support transparent backgrounds"
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

// These are protocol limits, not model-name inference. A catalog must opt
// into custom dimensions; ordinary models accept their explicit Sizes only.
func openAIImageDimensions(size string) bool {
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
	return errW == nil && errH == nil && width > 0 && height > 0 && width <= 3840 && height <= 3840 && min(width, height) <= 2160 && width%16 == 0 && height%16 == 0 && width <= 3*height && height <= 3*width
}

type openAIImagesRequest struct {
	Model             string           `json:"model"`
	Prompt            string           `json:"prompt"`
	N                 Nullable[int]    `json:"n,omitzero"`
	Size              Nullable[string] `json:"size,omitzero"`
	Quality           Nullable[string] `json:"quality,omitzero"`
	Background        Nullable[string] `json:"background,omitzero"`
	OutputFormat      Nullable[string] `json:"output_format,omitzero"`
	OutputCompression Nullable[int]    `json:"output_compression,omitzero"`
	Moderation        Nullable[string] `json:"moderation,omitzero"`
}

func (w openAIImagesRequest) options() OpenAIImagesOptions {
	return OpenAIImagesOptions{N: w.N, Size: w.Size, Quality: w.Quality, Background: w.Background, OutputFormat: w.OutputFormat, OutputCompression: w.OutputCompression, Moderation: w.Moderation}
}

func (p *ResourcePolicy) checkOpenAIImagesPayload(body map[string]any, model string) *Error {
	if body["model"] != model {
		return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	for key := range body {
		if slices.Contains([]string{"model", "prompt", "n", "size", "quality", "background", "output_format", "output_compression", "moderation"}, key) {
			continue
		}
		if slices.Contains([]string{"stream", "partial_images", "response_format", "user", "images", "image", "mask", "input_fidelity", "file_id", "image_url", "previous_response_id", "previous_interaction_id", "store", "tools", "endpoint", "operation", "api", "provider", "api_key", "credential", "headers", "background_execution"}, key) {
			return newError(CodeTenantDenied, PhaseRequest, "payload callback may not add image authority or operation fields")
		}
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback added an unsupported image field")
	}
	if _, ok := jsonPayloadSize(body, p.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	if prompt, ok := body["prompt"].(string); ok && (!utf8.ValidString(prompt) || utf8.RuneCountInString(prompt) < 1 || utf8.RuneCountInString(prompt) > 32000) {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid image prompt")
	}
	return nil
}

func decodeOpenAIImagesRequest(body []byte) (openAIImagesRequest, error) {
	var wire openAIImagesRequest
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	err := d.Decode(&wire)
	return wire, err
}
