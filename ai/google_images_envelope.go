package ai

import (
	"encoding/json"
	"errors"
	"io"
)

type googleImagesEnvelope struct {
	id, model                         Nullable[string]
	status                            string
	usage                             json.RawMessage
	content                           []ImageOutput
	imageCount                        int
	invalid, continuation, diagnostic bool
}

// Stream the bounded response step by step, and content block by content
// block. No full-body or full-step JSON copy coexists with the image output.
// Continue through invalid output to retain unambiguous usage and identity.
func readGoogleImagesEnvelope(d *json.Decoder, limit int) (googleImagesEnvelope, error) {
	w := googleImagesEnvelope{}
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return w, errors.New("invalid interaction envelope")
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
		if key == "steps" {
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
				step, err := readGoogleImageStep(d, limit-w.imageCount)
				if err != nil {
					return w, err
				}
				w.invalid = w.invalid || step.invalid
				if step.kind == "model_output" {
					w.imageCount += step.imageCount
					w.content = append(w.content, step.content...)
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
		case "id", "model":
			value, valid := rawString(raw)
			if !valid || value == "" {
				w.invalid = true
			}
			field := &w.id
			if key == "model" {
				field = &w.model
			}
			if duplicate {
				*field = Nullable[string]{}
			} else if valid {
				*field = Value(value)
			}
		case "status":
			var valid bool
			w.status, valid = rawString(raw)
			w.invalid = w.invalid || !valid
		case "usage":
			if duplicate {
				w.usage = nil
			} else {
				w.usage = raw
			}
		case "continuation_token":
			w.continuation = true
		case "errors":
			var faults []json.RawMessage
			if json.Unmarshal(raw, &faults) != nil {
				w.invalid = true
			} else {
				w.diagnostic = len(faults) > 0
			}
		}
	}
	if _, err := d.Token(); err != nil {
		return w, err
	}
	if err := d.Decode(new(json.RawMessage)); err != io.EOF {
		return w, errors.New("interaction response requires one JSON value")
	}
	return w, nil
}

type googleImageStep struct {
	kind       string
	content    []ImageOutput
	imageCount int
	invalid    bool
}

func readGoogleImageStep(d *json.Decoder, limit int) (googleImageStep, error) {
	step := googleImageStep{}
	contentInvalid := false
	t, err := d.Token()
	if err != nil {
		return step, err
	}
	if t != json.Delim('{') {
		step.invalid = true
		return step, skipImageJSONToken(d, t)
	}
	seen := map[string]bool{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return step, err
		}
		key := token.(string)
		if seen[key] {
			step.invalid = true
		}
		seen[key] = true
		if key == "content" {
			t, err := d.Token()
			if err != nil {
				return step, err
			}
			if t != json.Delim('[') {
				contentInvalid = true
				if err := skipImageJSONToken(d, t); err != nil {
					return step, err
				}
				continue
			}
			for d.More() {
				var raw json.RawMessage
				if err := d.Decode(&raw); err != nil {
					return step, err
				}
				block, invalid := decodeGoogleImageBlock(raw)
				contentInvalid = contentInvalid || invalid
				if _, ok := block.(ImageOutputImage); ok {
					step.imageCount++
					if step.imageCount > limit {
						continue
					}
				}
				if block != nil {
					step.content = append(step.content, block)
				}
			}
			if _, err := d.Token(); err != nil {
				return step, err
			}
			continue
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return step, err
		}
		if key == "type" {
			var valid bool
			step.kind, valid = rawString(raw)
			step.invalid = step.invalid || !valid
		}
	}
	if _, err := d.Token(); err != nil {
		return step, err
	}
	if step.kind == "" {
		step.invalid = true
	}
	// Thinking/tool steps are not output. Their content has a different schema
	// and must not be published or treated as an image. Top-level syntax still
	// has to be valid; the stateless operation never executes any returned tool.
	if step.kind != "model_output" {
		step.content = nil
		step.imageCount = 0
	} else {
		step.invalid = step.invalid || contentInvalid
	}
	return step, nil
}

func decodeGoogleImageBlock(raw json.RawMessage) (ImageOutput, bool) {
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil || duplicate {
		return nil, true
	}
	kind, ok := rawString(fields["type"])
	if !ok {
		return nil, true
	}
	if _, uri := fields["uri"]; uri {
		return nil, true
	}
	switch kind {
	case "text":
		text, ok := rawString(fields["text"])
		return ImageOutputText{Text: text}, !ok
	case "image":
		data, hasData := rawString(fields["data"])
		mime, hasMime := rawString(fields["mime_type"])
		return ImageOutputImage{Data: data, MimeType: mime}, !hasData || !hasMime || !isImageMediaType(mime)
	default:
		return nil, true
	}
}
