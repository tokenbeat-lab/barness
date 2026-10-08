package ai

import "unicode/utf8"

// ImagePolicy explicitly enables image generation with finite capacities.
// These are host resource budgets, separate from protocol/model capabilities.
type ImagePolicy struct {
	MaxInputImages           int
	MaxOutputImages          int
	MaxOutputImageBytes      int64
	MaxTotalOutputImageBytes int64
}

func (p *ImagePolicy) validate(output int64) error {
	if p == nil {
		return nil
	}
	for _, v := range []struct {
		name  string
		value int64
	}{
		{"MaxInputImages", int64(p.MaxInputImages)}, {"MaxOutputImages", int64(p.MaxOutputImages)},
		{"MaxOutputImageBytes", p.MaxOutputImageBytes}, {"MaxTotalOutputImageBytes", p.MaxTotalOutputImageBytes},
	} {
		if v.value <= 0 {
			return policyError("Image."+v.name, "must be positive; unlimited is not supported")
		}
	}
	if p.MaxOutputImageBytes > p.MaxTotalOutputImageBytes {
		return policyError("Image.MaxOutputImageBytes", "must not exceed MaxTotalOutputImageBytes")
	}
	if p.MaxTotalOutputImageBytes > output {
		return policyError("Image.MaxTotalOutputImageBytes", "must not exceed MaxOutputBytes")
	}
	return nil
}

func (p *ResourcePolicy) checkImageInput(req ImagesRequest) *Error {
	if p.Image == nil {
		return newError(CodeInvalidRequest, PhaseScope, "the client policy does not enable image generation")
	}
	if len(req.ReferenceImages) > p.Image.MaxInputImages {
		return limitFailure(PhaseScope, "Image.MaxInputImages", int64(p.Image.MaxInputImages))
	}
	remaining := p.MaxRequestBytes
	if int64(len(req.Prompt)) > remaining {
		return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
	}
	remaining -= int64(len(req.Prompt))
	for _, image := range req.ReferenceImages {
		if base64Size(image.Data) > p.MaxImageBytes {
			return limitFailure(PhaseScope, "MaxImageBytes", p.MaxImageBytes)
		}
		if int64(len(image.Data)) > remaining {
			return limitFailure(PhaseScope, "MaxRequestBytes", p.MaxRequestBytes)
		}
		remaining -= int64(len(image.Data))
		if !isImageMediaType(image.MimeType) {
			return newError(CodeInvalidRequest, PhaseScope, "reference image MIME must be an image media type")
		}
	}
	if !utf8.ValidString(req.Prompt) || utf8.RuneCountInString(req.Prompt) < 1 || utf8.RuneCountInString(req.Prompt) > 32000 {
		return newError(CodeInvalidRequest, PhaseScope, "image prompt requires 1 to 32000 valid Unicode characters")
	}
	return nil
}
