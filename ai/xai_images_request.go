package ai

import (
	"context"
	"encoding/json"
	"slices"
)

type xaiImageResource struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
type xaiImagesRequest struct {
	Model          string             `json:"model"`
	Prompt         string             `json:"prompt"`
	ResponseFormat string             `json:"response_format"`
	N              Nullable[int]      `json:"n,omitzero"`
	AspectRatio    Nullable[string]   `json:"aspect_ratio,omitzero"`
	Resolution     Nullable[string]   `json:"resolution,omitzero"`
	Quality        Nullable[string]   `json:"quality,omitzero"`
	Image          *xaiImageResource  `json:"image,omitempty"`
	Images         []xaiImageResource `json:"images,omitempty"`
}

func (w xaiImagesRequest) options() XAIImagesOptions {
	return XAIImagesOptions{N: w.N, AspectRatio: w.AspectRatio, Resolution: w.Resolution, Quality: w.Quality}
}
func (w xaiImagesRequest) resources() []xaiImageResource {
	if w.Image != nil {
		return []xaiImageResource{*w.Image}
	}
	return w.Images
}
func (p *ResourcePolicy) checkXAIImageCount(n int, caps ImageCapabilities, phase Phase) *Error {
	if n > 5 {
		return limitFailure(phase, "XAI.MaxReferenceImages", 5)
	}
	if n > caps.MaxReferenceImages {
		return limitFailure(phase, "ImageModel.MaxReferenceImages", int64(caps.MaxReferenceImages))
	}
	if n > p.Image.MaxInputImages {
		return limitFailure(phase, "Image.MaxInputImages", int64(p.Image.MaxInputImages))
	}
	return nil
}
func (p *ResourcePolicy) checkXAIImagesEncoding(w xaiImagesRequest) *Error {
	extra := int64(0)
	w.Images = slices.Clone(w.Images)
	if w.Image != nil {
		v := *w.Image
		extra += int64(len(v.URL)) - 1
		v.URL = "A"
		w.Image = &v
	}
	for i := range w.Images {
		extra += int64(len(w.Images[i].URL)) - 1
		w.Images[i].URL = "A"
	}
	small, err := json.Marshal(w)
	if err != nil {
		return imageInputFailure(PhaseCapability, "Imagine request could not be encoded")
	}
	if int64(len(small))+extra > p.MaxRequestBytes {
		return limitFailure(PhaseCapability, "MaxRequestBytes", p.MaxRequestBytes)
	}
	return nil
}
func (p *ResourcePolicy) checkXAIImagesPayload(body map[string]any, model ImageModel, editing bool) *Error {
	if body["model"] != model.ID || body["response_format"] != "b64_json" {
		return imageAuthorityFailure()
	}
	for key := range body {
		if slices.Contains([]string{"model", "prompt", "response_format", "n", "aspect_ratio", "resolution", "quality", "image", "images"}, key) {
			continue
		}
		if slices.Contains([]string{"stream", "deferred", "file_id", "image_url", "store", "include", "tools", "endpoint", "operation", "api", "provider", "headers", "api_key", "credential", "previous_response_id", "output_format", "upload_url"}, key) {
			return imageAuthorityFailure()
		}
		return imageInputFailure(PhaseRequest, "unsupported Imagine request field")
	}
	_, one := body["image"]
	_, many := body["images"]
	if (!editing && (one || many)) || (editing && !one && !many) {
		return imageAuthorityFailure()
	}
	if _, ok := jsonPayloadSize(body, p.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	return nil
}
func decodeXAIImagesRequest(ctx context.Context, body []byte, model ImageModel, p *ResourcePolicy, editing bool) (xaiImagesRequest, *Error) {
	var w xaiImagesRequest
	fields, duplicate, err := unaryJSONEnvelope(body)
	if err != nil || duplicate {
		return w, imageInputFailure(PhaseRequest, "invalid Imagine request JSON")
	}
	for key := range fields {
		if !slices.Contains([]string{"model", "prompt", "response_format", "n", "aspect_ratio", "resolution", "quality", "image", "images"}, key) {
			return w, imageInputFailure(PhaseRequest, "unsupported Imagine request field")
		}
	}
	if json.Unmarshal(body, &w) != nil {
		return w, imageInputFailure(PhaseRequest, "invalid Imagine request schema")
	}
	if w.Model != model.ID || w.ResponseFormat != "b64_json" {
		return w, imageAuthorityFailure()
	}
	_, one := fields["image"]
	_, many := fields["images"]
	if one && many {
		return w, imageInputFailure(PhaseRequest, "Imagine image and images are mutually exclusive")
	}
	refs := w.resources()
	if editing != (len(refs) > 0) {
		return w, imageAuthorityFailure()
	}
	if failure := p.checkXAIImageCount(len(refs), model.Capabilities, PhaseRequest); failure != nil {
		return w, failure
	}
	req := ImagesRequest{Prompt: w.Prompt}
	for _, r := range refs {
		if r.Type != "image_url" {
			return w, imageAuthorityFailure()
		}
		img, ok := parseOpenAIImageURL(r.URL)
		if !ok {
			return w, imageAuthorityFailure()
		}
		req.ReferenceImages = append(req.ReferenceImages, img)
	}
	if failure := p.checkImageInput(ctx, req, nil, PhaseRequest); failure != nil {
		return w, failure
	}
	if problem := w.options().check(model.Capabilities, p.Image, len(refs)); problem != "" {
		return w, imageInputFailure(PhaseRequest, problem)
	}
	return w, nil
}
func (r *xaiImageResource) UnmarshalJSON(raw []byte) error {
	fields, dup, err := unaryJSONEnvelope(raw)
	if err != nil || dup {
		return imageInputFailure(PhaseRequest, "invalid Imagine inline resource")
	}
	if _, ok := fields["file_id"]; ok {
		return imageAuthorityFailure()
	}
	if len(fields) != 2 {
		return imageInputFailure(PhaseRequest, "Imagine inline resource requires type and url")
	}
	var ok bool
	r.Type, ok = rawString(fields["type"])
	if !ok {
		return imageInputFailure(PhaseRequest, "Imagine inline resource requires type")
	}
	r.URL, ok = rawString(fields["url"])
	if !ok {
		return imageInputFailure(PhaseRequest, "Imagine inline resource requires url")
	}
	return nil
}
