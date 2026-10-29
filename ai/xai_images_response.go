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
	usage, reporting, amount, usageFailure := xaiImagesUsage(wire.usage)
	out.Usage = usage
	initial.usageEstimate(reporting, usage, CostEstimate{Status: CostPartial})
	initial.attempts[len(initial.attempts)-1].ReportedCostUSD = amount
	out.ResponseID, out.ResponseModel = wire.id, wire.model
	if ended := contextError(ctx, PhaseResponse); ended != nil {
		return ended
	}
	if r.failure != nil {
		return unaryReadFailure(r.failure)
	}
	if err != nil {
		return xaiImagesProtocolFailure()
	}
	if usageFailure != nil {
		return usageFailure
	}
	if wire.refused {
		return newError(CodeUpstreamError, PhaseResponse, "xAI Images refused the request")
	}
	if wire.count > min(policy.MaxOutputImages, model.Capabilities.MaxOutputImages) {
		return limitFailure(PhaseResponse, "Image.MaxOutputImages", int64(min(policy.MaxOutputImages, model.Capabilities.MaxOutputImages)))
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
		size := base64Size(img.data)
		if size > policy.MaxOutputImageBytes {
			return limitFailure(PhaseResponse, "Image.MaxOutputImageBytes", policy.MaxOutputImageBytes)
		}
		if size > remaining {
			return limitFailure(PhaseResponse, "Image.MaxTotalOutputImageBytes", policy.MaxTotalOutputImageBytes)
		}
		if img.invalid || mime == "" || !validImageBase64(ctx, img.data, mime) {
			return xaiImagesProtocolFailure()
		}
		remaining -= size
		content = append(content, ImageOutputImage{Data: img.data, MimeType: mime})
	}
	initial.usageEstimate(reporting, usage, xaiImagesEstimate(model, request, len(content)))
	out.Content = content
	return nil
}
