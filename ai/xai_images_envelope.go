package ai

import (
	"encoding/json"
	"errors"
	"io"
)

type xaiImagesEnvelope struct {
	images           []openAIImage
	count            int
	usage            json.RawMessage
	id, model        Nullable[string]
	invalid, refused bool
}

// Read one image at a time from the shared bounded response. Keep independent
// usage and identity even when another field fails validation.
func readXAIImagesEnvelope(d *json.Decoder, limit int) (xaiImagesEnvelope, error) {
	w := xaiImagesEnvelope{}
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return w, errors.New("invalid image envelope")
	}
	seen := map[string]bool{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return w, err
		}
		key := token.(string)
		duplicate := seen[key]
		seen[key] = true
		if duplicate {
			w.invalid = true
		}
		if key == "data" {
			token, err := d.Token()
			if err != nil {
				return w, err
			}
			if token != json.Delim('[') {
				w.invalid = true
				if err := skipImageJSONToken(d, token); err != nil {
					return w, err
				}
				continue
			}
			for d.More() {
				var raw json.RawMessage
				if err := d.Decode(&raw); err != nil {
					return w, err
				}
				w.count++
				image, refused := decodeXAIImage(raw)
				w.refused = w.refused || refused
				if w.count <= limit {
					w.images = append(w.images, image)
				}
			}
			if _, err := d.Token(); err != nil {
				return w, err
			}
			continue
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return w, err
		}
		switch key {
		case "usage":
			if duplicate {
				w.usage = nil
			} else {
				w.usage = raw
			}
		case "id", "model":
			target := &w.id
			if key == "model" {
				target = &w.model
			}
			if duplicate {
				*target = Nullable[string]{}
				continue
			}
			if string(raw) == "null" {
				continue
			}
			value, ok := rawString(raw)
			if !ok || value == "" {
				w.invalid = true
			} else {
				*target = Value(value)
			}
		case "request_id", "status", "url", "continuation_token":
			w.invalid = true
		case "error", "refusal":
			if rawTruthy(raw) {
				w.refused = true
			}
		case "respect_moderation":
			var respected bool
			if json.Unmarshal(raw, &respected) != nil || string(raw) == "null" {
				w.invalid = true
			} else if !respected {
				w.refused = true
			}
		}
	}
	if _, err := d.Token(); err != nil {
		return w, err
	}
	if err := d.Decode(new(json.RawMessage)); err != io.EOF {
		return w, errors.New("image response requires one JSON value")
	}
	return w, nil
}
func decodeXAIImage(raw json.RawMessage) (openAIImage, bool) {
	w := openAIImage{}
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil || duplicate {
		w.invalid = true
		return w, false
	}
	var ok bool
	w.data, ok = rawString(fields["b64_json"])
	w.invalid = !ok
	if mime, exists := fields["mime_type"]; exists && string(mime) != "null" {
		w.mime, ok = rawString(mime)
		w.invalid = w.invalid || !ok || !isImageMediaType(w.mime)
	}
	// Presence itself violates inline delivery, including explicit null/empty URL.
	for _, key := range []string{"url", "file_id", "request_id"} {
		if _, exists := fields[key]; exists {
			w.invalid = true
		}
	}
	refused := false
	if raw, exists := fields["respect_moderation"]; exists {
		var respected bool
		if json.Unmarshal(raw, &respected) != nil || string(raw) == "null" {
			w.invalid = true
		} else {
			refused = !respected
		}
	}
	return w, refused
}
