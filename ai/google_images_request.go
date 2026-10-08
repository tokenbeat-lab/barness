package ai

import (
	"encoding/json"
	"net/url"
	"slices"
	"strings"
)

const googleImagesRequestLimit int64 = 20 * 1024 * 1024

type googleImagesRequest struct {
	Model  string             `json:"model"`
	Store  bool               `json:"store"`
	Input  []googleImageInput `json:"input"`
	Format googleImageFormat  `json:"response_format"`
}
type googleImageInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type googleImageFormat struct {
	Type        string           `json:"type"`
	Delivery    string           `json:"delivery"`
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
func (p *ResourcePolicy) checkGoogleImagesPayload(body map[string]any) *Error {
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
	return nil
}

func googleImagesField(key string, allowed []string) *Error {
	if slices.Contains(allowed, key) {
		return nil
	}
	if slices.Contains([]string{"previous_interaction_id", "background", "agent", "agent_config", "tools", "environment", "webhook_config", "continuation_token", "service_tier", "stream", "generation_config", "response_modalities", "labels", "safety_settings", "endpoint", "operation", "uri", "url", "file_id", "data"}, key) {
		return googleImageAuthorityFailure()
	}
	return imageInputFailure(PhaseRequest, "unsupported Google image request field")
}
func googleImageAuthorityFailure() *Error {
	return newError(CodeTenantDenied, PhaseRequest, "image callback may not change the stateless inline Google operation or authorized model")
}

// Decode the final independent bytes, including raw JSON and custom values.
// Presence and duplicates matter: false/null do not authorize forbidden fields.
func decodeGoogleImagesRequest(body []byte, model ImageModel) (googleImagesRequest, *Error) {
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
		if failure := googleImagesField(key, []string{"type", "delivery", "aspect_ratio", "image_size"}); failure != nil {
			return wire, failure
		}
	}
	kind, _ := rawString(format["type"])
	delivery, _ := rawString(format["delivery"])
	if kind != "image" || delivery != "inline" {
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
	if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &input) != nil || len(input) == 0 {
		return wire, imageInputFailure(PhaseRequest, "Google image input must contain a prompt")
	}
	// Reference images have a separate authorization/validation path in issue
	// 13. This generation-only route must not let callbacks introduce that path.
	if len(input) != 1 {
		return wire, googleImageAuthorityFailure()
	}
	part, duplicate, err := unaryJSONEnvelope(input[0])
	if err != nil || duplicate {
		return wire, imageInputFailure(PhaseRequest, "invalid Google image input")
	}
	for key := range part {
		if failure := googleImagesField(key, []string{"type", "text"}); failure != nil {
			return wire, failure
		}
	}
	kind, _ = rawString(part["type"])
	if kind != "text" {
		return wire, googleImageAuthorityFailure()
	}
	text, ok := rawString(part["text"])
	if !ok {
		return wire, imageInputFailure(PhaseRequest, "Google image prompt must be text")
	}
	wire.Input = []googleImageInput{{Type: "text", Text: text}}
	return wire, nil
}
