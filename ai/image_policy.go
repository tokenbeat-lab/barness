package ai

import (
	"context"
	"unicode/utf8"
)

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

func (p *ResourcePolicy) checkImageInput(ctx context.Context, req ImagesRequest, mask *Image, phase Phase) *Error {
	if p.Image == nil {
		return newError(CodeInvalidRequest, PhaseScope, "the client policy does not enable image generation")
	}
	count := len(req.ReferenceImages)
	if mask != nil {
		count++
	}
	if count > p.Image.MaxInputImages {
		return limitFailure(phase, "Image.MaxInputImages", int64(p.Image.MaxInputImages))
	}
	if len(req.ReferenceImages) > openAIReferenceLimit {
		return limitFailure(phase, "OpenAI.MaxReferenceImages", openAIReferenceLimit)
	}
	remaining := p.MaxRequestBytes
	if int64(len(req.Prompt)) > remaining {
		return limitFailure(phase, "MaxRequestBytes", p.MaxRequestBytes)
	}
	remaining -= int64(len(req.Prompt))
	// Every known length is checked before any image is copied or decoded.
	for i := 0; i < count; i++ {
		image := imageInputAt(req.ReferenceImages, mask, i)
		if failure := p.checkInlineImageSize(image.Data, image.MimeType, phase); failure != nil {
			return failure
		}
		length := int64(len(image.Data)) + int64(len(image.MimeType)) + int64(len("data:;base64,"))
		if length > remaining {
			return limitFailure(phase, "MaxRequestBytes", p.MaxRequestBytes)
		}
		remaining -= length
	}
	if !utf8.ValidString(req.Prompt) || utf8.RuneCountInString(req.Prompt) < 1 || utf8.RuneCountInString(req.Prompt) > 32000 {
		return imageInputFailure(phase, "image prompt requires 1 to 32000 valid Unicode characters")
	}
	if mask != nil && len(req.ReferenceImages) == 0 {
		return imageInputFailure(phase, "image masks require reference images")
	}
	for i := 0; i < count; i++ {
		image := imageInputAt(req.ReferenceImages, mask, i)
		if !isImageMediaType(image.MimeType) || !validImageBase64(ctx, image.Data, image.MimeType) {
			if failure := contextError(ctx, phase); failure != nil {
				return failure
			}
			return imageInputFailure(phase, "input images require supported MIME and strict inline base64")
		}
	}
	if mask != nil && !imageMaskDimensionsMatch(req.ReferenceImages[0], *mask) {
		return imageInputFailure(phase, "image mask dimensions must match the first reference image")
	}
	return nil
}

func imageInputAt(images []Image, mask *Image, i int) Image {
	if i < len(images) {
		return images[i]
	}
	return *mask
}

func imageInputFailure(phase Phase, message string) *Error {
	code := CodeInvalidRequest
	if phase == PhaseRequest {
		code = CodeCallbackFailed
	}
	return newError(code, phase, message)
}
