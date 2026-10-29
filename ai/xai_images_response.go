package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
)

func decodeXAIImagesResponse(ctx context.Context, reader io.Reader, model ImageModel, request xaiImagesRequest, policy ImagePolicy, initial *initialRequest, out *ImagesResult) *Error {
	r := &imageResponseReader{Reader: reader}
	wire, err := readXAIImagesEnvelope(json.NewDecoder(r), min(policy.MaxOutputImages, model.Capabilities.MaxOutputImages))
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	if r.failure != nil {
		return unaryReadFailure(r.failure)
	}
	reporting, usage, amount, usageFailure := xaiImagesUsage(wire.usage)
	out.Usage = usage
	initial.usage(reporting, usage)
	_ = amount
	out.ResponseID, out.ResponseModel = wire.id, wire.model
	if err != nil {
		return xaiImagesProtocolFailure()
	}
	if usageFailure != nil {
		return usageFailure
	}
	if wire.refused {
		return newError(CodeUpstreamError, PhaseResponse, "xAI Images refused the request")
	}
	if wire.invalid || wire.count != request.options().quantity() {
		return xaiImagesProtocolFailure()
	}
	content := make([]ImageOutput, 0, len(wire.images))
	remaining := policy.MaxTotalOutputImageBytes
	for _, img := range wire.images {
		mime := img.mime
		if mime == "" {
			var header [12]byte
			n, _ := io.ReadFull(base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(img.data)), header[:])
			for _, candidate := range []string{"image/png", "image/jpeg", "image/webp"} {
				if imageHeaderMatches(header[:n], candidate) {
					mime = candidate
					break
				}
			}
		}
		if img.invalid || mime == "" || !validImageBase64(ctx, img.data, mime) {
			return xaiImagesProtocolFailure()
		}
		size := base64Size(img.data)
		if size > policy.MaxOutputImageBytes {
			return limitFailure(PhaseResponse, "Image.MaxOutputImageBytes", policy.MaxOutputImageBytes)
		}
		if size > remaining {
			return limitFailure(PhaseResponse, "Image.MaxTotalOutputImageBytes", policy.MaxTotalOutputImageBytes)
		}
		remaining -= size
		content = append(content, ImageOutputImage{Data: img.data, MimeType: mime})
	}
	out.Content = content
	return nil
}
