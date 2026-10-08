package ai

import (
	"context"
	"encoding/json"
	"io"
)

func googleImageProtocolFailure() *Error {
	return newError(CodeProtocol, PhaseResponse, "invalid Google Interactions image response")
}

func decodeGoogleImagesResponse(ctx context.Context, reader io.Reader, model ImageModel, policy ImagePolicy, initial *initialRequest, out *ImagesResult) *Error {
	r := &imageResponseReader{Reader: reader}
	limit := min(policy.MaxOutputImages, model.Capabilities.MaxOutputImages)
	wire, err := readGoogleImagesEnvelope(json.NewDecoder(r), limit)
	// Consumption and diagnostic identity can precede a malformed tail or
	// interrupted read. Keep that reported prefix without publishing output.
	out.ResponseID, out.ResponseModel = wire.id, wire.model
	reporting, usage, failure := googleImagesUsage(wire.usage, model.Pricing)
	out.Usage = usage
	initial.usage(reporting, usage)
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	if r.failure != nil {
		return unaryReadFailure(r.failure)
	}
	if err != nil {
		return googleImageProtocolFailure()
	}
	if failure != nil {
		return failure
	}
	if wire.invalid || wire.continuation {
		return googleImageProtocolFailure()
	}
	switch wire.status {
	case "failed", "cancelled":
		return newError(CodeUpstreamError, PhaseResponse, "Google Interactions image generation "+wire.status)
	case "completed":
	default:
		return googleImageProtocolFailure()
	}
	if wire.imageCount == 0 && wire.diagnostic {
		return newError(CodeUpstreamError, PhaseResponse, "Google Interactions blocked image generation")
	}
	if wire.imageCount > limit {
		return limitFailure(PhaseResponse, "Image.MaxOutputImages", int64(limit))
	}
	if wire.imageCount == 0 {
		return googleImageProtocolFailure()
	}
	remaining := policy.MaxTotalOutputImageBytes
	for _, block := range wire.content {
		if ended := contextError(ctx, PhaseResponse); ended != nil {
			return ended
		}
		img, ok := block.(ImageOutputImage)
		if !ok {
			continue
		}
		size := base64Size(img.Data)
		if size > policy.MaxOutputImageBytes {
			return limitFailure(PhaseResponse, "Image.MaxOutputImageBytes", policy.MaxOutputImageBytes)
		}
		if size > remaining {
			return limitFailure(PhaseResponse, "Image.MaxTotalOutputImageBytes", policy.MaxTotalOutputImageBytes)
		}
		if img.Data == "" || !validImageBase64(ctx, img.Data, img.MimeType) {
			if ended := contextError(ctx, PhaseResponse); ended != nil {
				return ended
			}
			return googleImageProtocolFailure()
		}
		remaining -= size
	}
	out.Content = wire.content
	return nil
}
