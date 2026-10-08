package ai

import (
	"encoding/json"
	"net/url"
	"slices"
	"strings"
)

const googleImagesRequestLimit int64 = 20_000_000
const googleImagesReferenceLimit = 14

type googleImagesRequest struct {
	Model  string             `json:"model"`
	Store  bool               `json:"store"`
	Input  []googleImageInput `json:"input"`
	Format googleImageFormat  `json:"response_format"`
}
type googleImageInput struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Data     string `json:"data,omitempty"`
}
type googleImageFormat struct {
	Type        string           `json:"type"`
	AspectRatio Nullable[string] `json:"aspect_ratio,omitzero"`
	ImageSize   Nullable[string] `json:"image_size,omitzero"`
}

func googleImagesTargetProblem(endpoint, id string) string {
	u, err := url.Parse(endpoint)
	if err != nil || !strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/v1beta") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "Google Interactions requires a v1beta base endpoint without query credentials"
	}
	if id == "" || strings.ContainsAny(id, "/?# \t\r\n") {
		return "Google Interactions requires a bare model ID"
	}
	return ""
}

// Check known host-container sizes before encoding; custom encodings are
// checked again after freezing, just like the other unary protocols.
func (p *ResourcePolicy) checkGoogleImagesPayload(body map[string]any, caps ImageCapabilities) *Error {
	for key := range body {
		if failure := googleImagesField(key, []string{"model", "store", "input", "response_format"}); failure != nil {
			return failure
		}
	}
	if _, ok := jsonPayloadSize(body, p.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	if _, ok := jsonPayloadSize(body, googleImagesRequestLimit); !ok {
		return limitFailure(PhaseRequest, "Google.MaxInlineRequestBytes", googleImagesRequestLimit)
	}
	head := make(map[string]any, len(body))
	for key, value := range body {
		if key != "input" {
			head[key] = value
		}
	}
	remaining := min(p.MaxRequestBytes, googleImagesRequestLimit)
	size, ok := jsonPayloadSize(head, remaining)
	if !ok {
		return p.takeGoogleInputBytes(remaining+1, &remaining)
	}
	// The header includes its braces; inserting input adds a key and comma.
	if failure := p.takeGoogleInputBytes(size+int64(len(`,"input":`)), &remaining); failure != nil {
		return failure
	}
	return p.checkGoogleInputPayload(body["input"], caps, &remaining)
}

func googleImagesField(key string, allowed []string) *Error {
	if slices.Contains(allowed, key) {
		return nil
	}
	if slices.Contains([]string{"previous_interaction_id", "background", "agent", "agent_config", "tools", "environment", "webhook_config", "continuation_token", "service_tier", "stream", "generation_config", "response_modalities", "labels", "safety_settings", "endpoint", "operation", "uri", "url", "file_id", "mask", "delivery"}, key) {
		return googleImageAuthorityFailure()
	}
	return imageInputFailure(PhaseRequest, "unsupported Google image request field")
}
func googleImageAuthorityFailure() *Error {
	return newError(CodeTenantDenied, PhaseRequest, "image callback may not change the stateless inline Google operation or authorized model")
}

// Decode the final independent bytes, including raw JSON and custom values.
// Presence and duplicates matter: false/null do not authorize forbidden fields.
func decodeGoogleImagesRequest(body []byte, model ImageModel, policy *ResourcePolicy) (googleImagesRequest, *Error) {
	wire := googleImagesRequest{}
	fields, duplicate, err := unaryJSONEnvelope(body)
	if err != nil || duplicate {
		return wire, imageInputFailure(PhaseRequest, "invalid Google image request")
	}
	for key := range fields {
		if failure := googleImagesField(key, []string{"model", "store", "input", "response_format"}); failure != nil {
			return wire, failure
		}
	}
	id, ok := rawString(fields["model"])
	if !ok || id != model.ID || string(fields["store"]) != "false" {
		return wire, googleImageAuthorityFailure()
	}
	wire.Model = id
	if _, present := fields["response_format"]; !present {
		return wire, googleImageAuthorityFailure()
	}
	format, duplicate, err := unaryJSONEnvelope(fields["response_format"])
	if err != nil || duplicate {
		return wire, imageInputFailure(PhaseRequest, "invalid Google image response format")
	}
	for key := range format {
		if failure := googleImagesField(key, []string{"type", "aspect_ratio", "image_size"}); failure != nil {
			return wire, failure
		}
	}
	kind, _ := rawString(format["type"])
	if kind != "image" {
		return wire, googleImageAuthorityFailure()
	}
	if json.Unmarshal(fields["response_format"], &wire.Format) != nil {
		return wire, imageInputFailure(PhaseRequest, "invalid Google image options")
	}
	options := GoogleImagesOptions{AspectRatio: wire.Format.AspectRatio, ImageSize: wire.Format.ImageSize}
	if problem := options.check(model.Capabilities); problem != "" {
		return wire, imageInputFailure(PhaseRequest, problem)
	}
	var input []json.RawMessage
	raw := fields["input"]
	remaining := min(policy.MaxRequestBytes, googleImagesRequestLimit)
	if failure := policy.checkGoogleInputPayload(raw, model.Capabilities, &remaining); failure != nil {
		return wire, failure
	}
	if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &input) != nil || len(input) == 0 {
		return wire, imageInputFailure(PhaseRequest, "Google image input must contain a prompt")
	}
	for i, raw := range input {
		part, duplicate, err := unaryJSONEnvelope(raw)
		if err != nil || duplicate {
			return wire, imageInputFailure(PhaseRequest, "invalid Google image input")
		}
		allowed := []string{"type", "text"}
		if i > 0 {
			allowed = []string{"type", "mime_type", "data"}
		}
		for key := range part {
			if failure := googleImagesField(key, allowed); failure != nil {
				return wire, failure
			}
		}
		kind, ok := rawString(part["type"])
		if !ok || (i == 0 && kind != "text") || (i > 0 && kind != "image") {
			return wire, imageInputFailure(PhaseRequest, "Google image input requires prompt then inline images")
		}
		value := googleImageInput{Type: kind}
		if i == 0 {
			value.Text, ok = rawString(part["text"])
		} else {
			value.MimeType, ok = rawString(part["mime_type"])
			if ok {
				value.Data, ok = rawString(part["data"])
			}
		}
		if !ok {
			return wire, imageInputFailure(PhaseRequest, "invalid Google image input fields")
		}
		wire.Input = append(wire.Input, value)
	}
	return wire, nil
}

func (w googleImagesRequest) imagesInput() ImagesRequest {
	req := ImagesRequest{Prompt: w.Input[0].Text}
	for _, part := range w.Input[1:] {
		req.ReferenceImages = append(req.ReferenceImages, Image{Data: part.Data, MimeType: part.MimeType})
	}
	return req
}

func (p *ResourcePolicy) checkGoogleImageCount(count int, caps ImageCapabilities, phase Phase) *Error {
	if count > p.Image.MaxInputImages {
		return limitFailure(phase, "Image.MaxInputImages", int64(p.Image.MaxInputImages))
	}
	if count > caps.MaxReferenceImages {
		return limitFailure(phase, "ImageModel.MaxReferenceImages", int64(caps.MaxReferenceImages))
	}
	if count > googleImagesReferenceLimit {
		return limitFailure(phase, "Google.MaxReferenceImages", googleImagesReferenceLimit)
	}
	return nil
}
