package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
)

func imageAuthorityFailure() *Error {
	return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change image authority or operation")
}

func (p *ResourcePolicy) checkOpenAIImagesPayload(body map[string]any, model ImageModel, editing bool) *Error {
	if body["model"] != model.ID {
		return imageAuthorityFailure()
	}
	for key := range body {
		if slices.Contains([]string{"model", "prompt", "n", "size", "quality", "background", "output_format", "output_compression", "moderation"}, key) {
			continue
		}
		if editing && slices.Contains([]string{"images", "mask", "input_fidelity"}, key) {
			continue
		}
		if slices.Contains([]string{"stream", "partial_images", "response_format", "user", "images", "image", "mask", "input_fidelity", "file_id", "image_url", "previous_response_id", "previous_interaction_id", "store", "tools", "endpoint", "operation", "api", "provider", "api_key", "credential", "headers", "background_execution"}, key) {
			return imageAuthorityFailure()
		}
		return imageInputFailure(PhaseRequest, "payload callback added an unsupported image field")
	}
	if editing && body["images"] == nil {
		return imageAuthorityFailure()
	}
	if _, ok := jsonPayloadSize(body, p.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	if editing {
		if failure := p.checkImageReferencePayload(reflect.ValueOf(body["images"]), body["mask"] != nil, model.Capabilities); failure != nil {
			return failure
		}
	}
	if mask, exists := body["mask"]; exists {
		if mask == nil {
			return imageInputFailure(PhaseRequest, "image mask must be an inline image object")
		}
		if failure := p.checkInlineImagePayload(reflect.ValueOf(mask)); failure != nil {
			return failure
		}
	}
	return nil
}

func (p *ResourcePolicy) checkFinalImageCount(n int, mask bool, caps ImageCapabilities) *Error {
	if n > openAIReferenceLimit {
		return limitFailure(PhaseRequest, "OpenAI.MaxReferenceImages", openAIReferenceLimit)
	}
	if n > caps.MaxReferenceImages {
		return limitFailure(PhaseRequest, "ImageModel.MaxReferenceImages", int64(caps.MaxReferenceImages))
	}
	count := n
	if mask {
		count++
	}
	if count > p.Image.MaxInputImages {
		return limitFailure(PhaseRequest, "Image.MaxInputImages", int64(p.Image.MaxInputImages))
	}
	return nil
}

func (p *ResourcePolicy) checkInlineURL(url string) *Error {
	if len(url) > openAIInlineURLLimit {
		return limitFailure(PhaseRequest, "OpenAI.MaxInlineImageURLCharacters", openAIInlineURLLimit)
	}
	if !strings.HasPrefix(url, "data:") {
		return imageAuthorityFailure()
	}
	img, ok := parseOpenAIImageURL(url)
	if !ok {
		return imageInputFailure(PhaseRequest, "input image must be an inline base64 data URL")
	}
	return p.checkInlineImageSize(img.Data, img.MimeType, PhaseRequest)
}

func (p *ResourcePolicy) checkOpenAIImagesFinal(ctx context.Context, wire openAIImagesRequest, caps ImageCapabilities, editing bool) *Error {
	if editing != (len(wire.Images) > 0) {
		return imageAuthorityFailure()
	}
	if !editing && (!wire.Mask.IsZero() || !wire.InputFidelity.IsZero()) {
		return imageAuthorityFailure()
	}
	if wire.Mask.IsNull() {
		return imageInputFailure(PhaseRequest, "image mask must be an inline image object")
	}
	if failure := p.checkFinalImageCount(len(wire.Images), !wire.Mask.IsZero(), caps); failure != nil {
		return failure
	}
	// Validate all known URL lengths before allocating the domain snapshot.
	for _, ref := range wire.Images {
		if failure := p.checkInlineURL(ref.ImageURL); failure != nil {
			return failure
		}
	}
	if mask, ok := wire.Mask.Get(); ok {
		if failure := p.checkInlineURL(mask.ImageURL); failure != nil {
			return failure
		}
	}
	req := ImagesRequest{Prompt: wire.Prompt, ReferenceImages: make([]Image, len(wire.Images))}
	for i, ref := range wire.Images {
		req.ReferenceImages[i], _ = parseOpenAIImageURL(ref.ImageURL)
	}
	options := wire.options()
	if failure := p.checkImageInput(ctx, req, options.Mask, PhaseRequest); failure != nil {
		return failure
	}
	return p.checkImageCapabilities(req, options, caps, PhaseRequest)
}

// A callback can supply raw JSON or typed containers. Enforce the resource
// reference schema at decoding too, so those forms cannot bypass authority.
func (r *openAIInlineImage) UnmarshalJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return imageInputFailure(PhaseRequest, "inline image must be an object")
	}
	seen := false
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return imageInputFailure(PhaseRequest, "invalid inline image field")
		}
		if key == "file_id" {
			return imageAuthorityFailure()
		}
		if key != "image_url" || seen {
			return imageInputFailure(PhaseRequest, "inline image requires one image_url field")
		}
		seen = true
		var url Nullable[string]
		if d.Decode(&url) != nil || url.IsNull() {
			return imageInputFailure(PhaseRequest, "image_url must be a string")
		}
		r.ImageURL, _ = url.Get()
	}
	if _, err := d.Token(); err != nil || !seen {
		return imageInputFailure(PhaseRequest, "inline image requires image_url")
	}
	return nil
}
