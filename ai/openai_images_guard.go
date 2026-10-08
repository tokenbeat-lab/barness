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
	if editing {
		v, exists := body["images"]
		if !exists || v == nil {
			return imageAuthorityFailure()
		}
		// Count known slices before encoding/decoding; raw JSON is sized by
		// its bytes and validated again after the independent freeze.
		value := reflect.ValueOf(v)
		if (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) && value.Type().Elem().Kind() != reflect.Uint8 {
			n := value.Len()
			if n == 0 {
				return imageAuthorityFailure()
			}
			if failure := p.checkFinalImageCount(n, body["mask"] != nil, model.Capabilities); failure != nil {
				return failure
			}
		}
	}
	if _, ok := jsonPayloadSize(body, p.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	if raw, ok := body["images"].(json.RawMessage); ok {
		if failure := p.checkRawImageReferences(raw, body["mask"] != nil, model.Capabilities); failure != nil {
			return failure
		}
	} else if images := reflect.ValueOf(body["images"]); images.IsValid() && (images.Kind() == reflect.Slice || images.Kind() == reflect.Array) && images.Type().Elem().Kind() != reflect.Uint8 {
		for i := 0; i < images.Len(); i++ {
			if failure := p.checkInlineImagePayload(images.Index(i).Interface()); failure != nil {
				return failure
			}
		}
	}
	if mask, exists := body["mask"]; exists {
		if mask == nil {
			return imageInputFailure(PhaseRequest, "image mask must be an inline image object")
		}
		if failure := p.checkInlineImagePayload(mask); failure != nil {
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

func (p *ResourcePolicy) checkInlineImagePayload(entry any) *Error {
	m := reflect.ValueOf(entry)
	if !m.IsValid() || m.Kind() != reflect.Map || m.Type().Key().Kind() != reflect.String {
		return nil
	} // Custom encodings are checked after freezing.
	fields := m.MapRange()
	for fields.Next() {
		key := fields.Key().String()
		if key == "file_id" {
			return imageAuthorityFailure()
		}
		if key != "image_url" {
			return imageInputFailure(PhaseRequest, "unsupported inline image field")
		}
		value := fields.Value()
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
		}
		if value.Kind() == reflect.String {
			if failure := p.checkInlineURL(value.String()); failure != nil {
				return failure
			}
		}
	}
	return nil
}

// Raw callback arrays are walked one entry at a time and stop at the first
// excess reference; never allocate a slice proportional to attacker input.
func (p *ResourcePolicy) checkRawImageReferences(raw json.RawMessage, mask bool, caps ImageCapabilities) *Error {
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('[') {
		return imageInputFailure(PhaseRequest, "images must be an array")
	}
	n := 0
	for d.More() {
		n++
		if failure := p.checkFinalImageCount(n, mask, caps); failure != nil {
			return failure
		}
		var ref openAIInlineImage
		if err := d.Decode(&ref); err != nil {
			if failure, ok := err.(*Error); ok {
				return failure
			}
			return imageInputFailure(PhaseRequest, "invalid inline image reference")
		}
		if failure := p.checkInlineURL(ref.ImageURL); failure != nil {
			return failure
		}
	}
	if n == 0 {
		return imageAuthorityFailure()
	}
	return nil // The full decoder verifies termination after freezing.
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
