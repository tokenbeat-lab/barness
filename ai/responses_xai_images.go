package ai

import (
	"context"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

const xaiMaxImageBytes int64 = 20 * 1024 * 1024

func validateXAIResponsesImages(ctx context.Context, m Model, b map[string]any, limits byteLimits, callback bool) *Error {
	code := CodeInvalidRequest
	if callback {
		code = CodeCallbackFailed
	}
	input, _ := b["input"].([]any)
	for _, raw := range input {
		item, _ := raw.(map[string]any)
		parts, _ := item["content"].([]any)
		for _, rawPart := range parts {
			part, _ := rawPart.(map[string]any)
			if part["type"] != "input_image" {
				continue
			}
			if !m.acceptsImages() {
				return newError(code, PhaseRequest, "the selected xAI model accepts text only")
			}
			url, _ := part["image_url"].(string)
			mime, data, ok := xaiInlineImage(url)
			if !ok {
				return newError(code, PhaseRequest, "xAI images require strict inline PNG or JPEG")
			}
			size := base64Size(data)
			if size > xaiMaxImageBytes {
				return limitFailure(PhaseRequest, "xAI.MaxImageBytes", xaiMaxImageBytes)
			}
			if size > limits.image {
				return limitFailure(PhaseRequest, "MaxImageBytes", limits.image)
			}
			if !validImageBase64(ctx, data, mime) {
				if failure := contextError(ctx, PhaseRequest); failure != nil {
					return failure
				}
				return newError(code, PhaseRequest, "xAI images require strict base64 matching their MIME")
			}
			// DecodeConfig validates an actual file header without allocating pixels.
			_, format, err := image.DecodeConfig(base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(data)))
			if err != nil || (mime == "image/png" && format != "png") || (mime == "image/jpeg" && format != "jpeg") {
				return newError(code, PhaseRequest, "xAI image header does not match a supported image")
			}
		}
	}
	return nil
}
func xaiInlineImage(url string) (string, string, bool) {
	for _, mime := range []string{"image/png", "image/jpeg"} {
		if data, ok := strings.CutPrefix(url, "data:"+mime+";base64,"); ok && data != "" {
			return mime, data, true
		}
	}
	return "", "", false
}

func xaiResponsesToolOutput(m ToolResultMessage) ([]any, error) {
	call, err := xaiResponsesCallID(m.ToolCallID)
	if err != nil {
		return nil, err
	}
	var texts []string
	var images []any
	for _, block := range m.Content {
		switch b := block.(type) {
		case Text:
			texts = append(texts, b.Text)
		case Image:
			images = append(images, newResponsesInputImage(b))
		}
	}
	text := strings.Join(texts, "\n")
	if len(images) > 0 {
		if text != "" {
			text += "\n"
		}
		text += "(see attached image)"
	}
	if text == "" {
		text = "(no tool output)"
	}
	out := []any{responsesFunctionCallOutput{Type: "function_call_output", CallID: call, Output: text}}
	if len(images) > 0 {
		out = append(out, responsesUserMessage{Role: "user", Content: images})
	}
	return out, nil
}
