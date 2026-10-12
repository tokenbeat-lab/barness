package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
)

type decisionsMessage struct {
	Role    string             `json:"role"`
	Content []decisionsContent `json:"content"`
}
type decisionsContent struct {
	Type     string  `json:"type"`
	Text     *string `json:"text,omitempty"`
	ImageURL *string `json:"image_url,omitempty"`
}

// The final native input is a closed union: text or exactly one user message
// with text first and ordered inline images. Schema never grants authority.
func (c *Client) checkDecisionsInput(ctx context.Context, raw []byte, model ClassifierModel, allowImages bool) *Error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return decisionsSchemaFailure()
	}
	if raw[0] == '"' {
		var text string
		if json.Unmarshal(raw, &text) != nil {
			return decisionsSchemaFailure()
		}
		return c.checkDecisionsText(text)
	}
	var messages []json.RawMessage
	if json.Unmarshal(raw, &messages) != nil || len(messages) != 1 {
		return decisionsSchemaFailure()
	}
	var message struct {
		Role    string            `json:"role"`
		Content []json.RawMessage `json:"content"`
	}
	if decodeClassifierJSON(messages[0], &message) != nil || message.Role != "user" || len(message.Content) < 2 {
		return decisionsSchemaFailure()
	}
	images := make([]Image, 0, len(message.Content)-1)
	for i, part := range message.Content {
		if i == 0 {
			var text struct {
				Type string  `json:"type"`
				Text *string `json:"text"`
			}
			if decodeClassifierJSON(part, &text) != nil || text.Type != "input_text" || text.Text == nil {
				return decisionsSchemaFailure()
			}
			if failure := c.checkDecisionsText(*text.Text); failure != nil {
				return failure
			}
		} else {
			var image struct {
				Type string  `json:"type"`
				URL  *string `json:"image_url"`
			}
			if decodeClassifierJSON(part, &image) != nil || image.Type != "input_image" || image.URL == nil {
				return decisionsSchemaFailure()
			}
			img, ok := parseOpenAIImageURL(*image.URL)
			if !ok || !classifierImageMIME(img.MimeType) {
				return decisionsSchemaFailure()
			}
			images = append(images, img)
		}
	}
	if !allowImages || c.policy.Classifier.Images == nil {
		return decisionsImageDenied()
	}
	if !slices.Contains(model.Input, ModalityImage) {
		return decisionsImageDenied()
	}
	for _, img := range images {
		if !slices.Contains(model.Capabilities.ImageMIMETypes, img.MimeType) {
			return decisionsImageDenied()
		}
	}
	if failure := c.policy.checkClassifierImageSizes(images, PhaseRequest); failure != nil {
		return failure
	}
	if len(images) > min(model.Capabilities.MaxInputImages, decisionsImageLimit) {
		return limitFailure(PhaseRequest, "ClassifierModel.MaxInputImages", int64(min(model.Capabilities.MaxInputImages, decisionsImageLimit)))
	}
	return checkClassifierImageEncoding(ctx, images, PhaseRequest)
}

func (c *Client) checkDecisionsText(text string) *Error {
	if int64(len(text)) > c.policy.Classifier.MaxStateBytes {
		return limitFailure(PhaseRequest, "Classifier.MaxStateBytes", c.policy.Classifier.MaxStateBytes)
	}
	return nil
}

func decisionsImageDenied() *Error {
	return newError(CodeTenantDenied, PhaseRequest, "payload callback may not exceed pinned classifier image authority")
}
