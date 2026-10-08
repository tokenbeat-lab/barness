package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type openAIImage struct {
	data, mime string
	invalid    bool
}
type openAIImagesEnvelope struct {
	images     []openAIImage
	imageCount int
	format     Nullable[string]
	usage      json.RawMessage
	refused    bool
	invalid    bool
}

// Observe transport read errors separately from JSON syntax errors: both a
// truncated JSON value and an interrupted HTTP body can yield unexpected EOF.
type imageResponseReader struct {
	io.Reader
	failure error
}

func (r *imageResponseReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err != nil && err != io.EOF {
		r.failure = err
	}
	return n, err
}

func decodeOpenAIImagesResponse(ctx context.Context, reader io.Reader, model ImageModel, opts OpenAIImagesOptions, policy ImagePolicy, initial *initialRequest, out *ImagesResult) *Error {
	r := &imageResponseReader{Reader: reader}
	outputLimit := min(policy.MaxOutputImages, model.Capabilities.MaxOutputImages)
	wire, err := readOpenAIImagesEnvelope(json.NewDecoder(r), outputLimit)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	if r.failure != nil {
		return unaryReadFailure(r.failure)
	}
	if err != nil {
		return imageProtocolFailure()
	}
	reporting, usage, failure := openAIImagesUsage(wire.usage, model.Pricing)
	out.Usage = usage
	initial.usage(reporting, usage)
	if failure != nil {
		return failure
	}
	if wire.refused {
		return newError(CodeUpstreamError, PhaseResponse, "OpenAI Images refused the request")
	}
	if wire.invalid {
		return imageProtocolFailure()
	}
	if wire.imageCount > outputLimit {
		return limitFailure(PhaseResponse, "Image.MaxOutputImages", int64(outputLimit))
	}
	if len(wire.images) != opts.quantity() {
		return imageProtocolFailure()
	}
	format := opts.outputFormat()
	if f, ok := wire.format.Get(); ok {
		format = f
	}
	mime := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp"}[format]
	if mime == "" {
		return imageProtocolFailure()
	}
	content := make([]ImageOutput, 0, len(wire.images))
	remaining := policy.MaxTotalOutputImageBytes
	for i, image := range wire.images {
		if ended := contextError(ctx, PhaseResponse); ended != nil {
			return ended
		}
		if image.invalid || image.data == "" || (image.mime != "" && image.mime != mime) {
			return imageProtocolFailure()
		}
		size := base64Size(image.data)
		if size > policy.MaxOutputImageBytes {
			return limitFailure(PhaseResponse, "Image.MaxOutputImageBytes", policy.MaxOutputImageBytes)
		}
		if size > remaining {
			return limitFailure(PhaseResponse, "Image.MaxTotalOutputImageBytes", policy.MaxTotalOutputImageBytes)
		}
		if !validImageBase64(ctx, image.data, mime) {
			if ended := contextError(ctx, PhaseResponse); ended != nil {
				return ended
			}
			return imageProtocolFailure()
		}
		remaining -= size
		content = append(content, ImageOutputImage{Data: image.data, MimeType: mime})
		wire.images[i] = openAIImage{} // Drop the parser's reference as it is transferred.
	}
	out.Content = content
	return nil
}

// Strict base64 is validated in fixed-size chunks, retaining only the file
// header. Decoded images are not kept alongside the published base64 strings.
func validImageBase64(ctx context.Context, data, mime string) bool {
	if len(data)%4 != 0 || strings.ContainsAny(data, "\r\n") {
		return false
	}
	d := base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(data))
	var header [12]byte
	n, err := io.ReadFull(d, header[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return false
	}
	if !imageHeaderMatches(header[:n], mime) {
		return false
	}
	if err != nil {
		return true
	}
	var buffer [4096]byte
	for {
		if ctx.Err() != nil {
			return false
		}
		_, err := d.Read(buffer[:])
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

func imageHeaderMatches(header []byte, mime string) bool {
	switch mime {
	case "image/png":
		return len(header) >= 8 && string(header[:8]) == "\x89PNG\r\n\x1a\n"
	case "image/jpeg":
		return len(header) >= 3 && string(header[:3]) == "\xff\xd8\xff"
	case "image/webp":
		return len(header) >= 12 && string(header[:4]) == "RIFF" && string(header[8:12]) == "WEBP"
	}
	return false
}

// Decode directly from the bounded body. Only one image entry's raw JSON
// exists at a time; the full response body is never collected or copied.
func readOpenAIImagesEnvelope(d *json.Decoder, outputLimit int) (openAIImagesEnvelope, error) {
	w := openAIImagesEnvelope{}
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return w, errors.New("invalid image envelope")
	}
	seen := map[string]bool{}
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return w, err
		}
		key, ok := t.(string)
		if !ok {
			return w, errors.New("invalid image key")
		}
		if seen[key] {
			w.invalid = true
			if key == "usage" {
				w.usage = nil
			}
		}
		duplicate := seen[key]
		seen[key] = true
		if key == "data" {
			t, err := d.Token()
			if err != nil {
				return w, err
			}
			if t != json.Delim('[') {
				w.invalid = true
				if err := skipImageJSONToken(d, t); err != nil {
					return w, err
				}
				continue
			}
			for d.More() {
				var raw json.RawMessage
				if err := d.Decode(&raw); err != nil {
					return w, err
				}
				w.imageCount++
				// Continue the bounded envelope read to retain usage while
				// keeping output records within the effective count budget.
				if w.imageCount <= outputLimit {
					w.images = append(w.images, decodeOpenAIImage(raw))
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
			if !duplicate {
				w.usage = raw
			}
		case "output_format":
			if json.Unmarshal(raw, &w.format) != nil {
				w.invalid = true
			}
		case "error", "refusal":
			if rawTruthy(raw) {
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
func decodeOpenAIImage(raw json.RawMessage) openAIImage {
	w := openAIImage{}
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil || duplicate {
		w.invalid = true
		return w
	}
	data, ok := rawString(fields["b64_json"])
	if !ok {
		w.invalid = true
	}
	w.data = data
	if raw, ok := fields["mime_type"]; ok {
		mime, valid := rawString(raw)
		if !valid || (mime != "image/png" && mime != "image/jpeg" && mime != "image/webp") {
			w.invalid = true
		}
		w.mime = mime
	}
	// URL delivery is never fetched, even if an inline image also exists.
	if rawTruthy(fields["url"]) {
		w.invalid = true
	}
	return w
}
func skipImageJSONToken(d *json.Decoder, t json.Token) error {
	if delim, ok := t.(json.Delim); ok && (delim == '{' || delim == '[') {
		for d.More() {
			next, err := d.Token()
			if err != nil {
				return err
			}
			if err := skipImageJSONToken(d, next); err != nil {
				return err
			}
		}
		_, err := d.Token()
		return err
	}
	return nil
}
func imageProtocolFailure() *Error {
	return newError(CodeProtocol, PhaseResponse, "invalid OpenAI Images response")
}
