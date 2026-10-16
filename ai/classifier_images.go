package ai

import (
	"context"
	"slices"
)

// ClassifierImagePolicy holds finite host budgets, independent of image
// generation. MaxImages/MaxTotalImageBytes bound the source collection;
// References bounds occurrences after reuse. MaxImageBytes and MaxRequestBytes
// also apply to classification.
type ClassifierImagePolicy struct {
	References         *ClassifierImageReferencePolicy
	MaxImages          int
	MaxTotalImageBytes int64
}

const decisionsImageLimit = 128

func (p *ClassifierImagePolicy) validate(requestLimit int64) error {
	if p == nil {
		return nil
	}
	if p.MaxImages <= 0 {
		return policyError("Classifier.Images.MaxImages", "must be positive; unlimited is not supported")
	}
	if p.MaxTotalImageBytes <= 0 || p.MaxTotalImageBytes > requestLimit {
		return policyError("Classifier.Images.MaxTotalImageBytes", "must be positive and must not exceed MaxRequestBytes")
	}
	if p.References != nil {
		if p.References.MaxReferences <= 0 {
			return policyError("Classifier.Images.References.MaxReferences", "must be positive; unlimited is not supported")
		}
		if p.References.MaxExpandedImageBytes <= 0 {
			return policyError("Classifier.Images.References.MaxExpandedImageBytes", "must be positive; unlimited is not supported")
		}
	}
	return nil
}

func classifierImageMIME(mime string) bool {
	return mime == "image/png" || mime == "image/jpeg" || mime == "image/webp"
}

// Count and check every known length before decoding or copying any input.
// Subtraction from the remaining budget avoids cumulative integer overflow.
func (p *ResourcePolicy) checkClassifierImageSizes(images []Image, phase Phase) *Error {
	if len(images) == 0 {
		return nil
	}
	if p.Classifier == nil || p.Classifier.Images == nil {
		return imageInputFailure(phase, "the client policy does not enable classifier images")
	}
	ip := p.Classifier.Images
	if len(images) > ip.MaxImages {
		return limitFailure(phase, "Classifier.Images.MaxImages", int64(ip.MaxImages))
	}
	remaining := ip.MaxTotalImageBytes
	for _, img := range images {
		size := base64Size(img.Data)
		if size > p.MaxImageBytes {
			return limitFailure(phase, "MaxImageBytes", p.MaxImageBytes)
		}
		if size > remaining {
			return limitFailure(phase, "Classifier.Images.MaxTotalImageBytes", ip.MaxTotalImageBytes)
		}
		remaining -= size
	}
	return nil
}

func checkClassifierImageEncoding(ctx context.Context, images []Image, phase Phase) *Error {
	for _, img := range images {
		if failure := contextError(ctx, phase); failure != nil {
			return failure
		}
		if !classifierImageMIME(img.MimeType) || !validImageBase64(ctx, img.Data, img.MimeType) {
			if failure := contextError(ctx, phase); failure != nil {
				return failure
			}
			return imageInputFailure(phase, "classifier images require PNG/JPEG/WebP headers and strict inline base64")
		}
	}
	return contextError(ctx, phase)
}

func classifierImageCapability(images []Image, model ClassifierModel) string {
	if len(images) == 0 {
		return ""
	}
	caps := model.Capabilities
	if !slices.Contains(model.Input, ModalityImage) || len(images) > classifierImageLimit(model) {
		return "classifier images exceed model or protocol capability"
	}
	for _, img := range images {
		if !slices.Contains(caps.ImageMIMETypes, img.MimeType) {
			return "classifier image MIME is not supported by the model"
		}
	}
	return ""
}

// Protocol bounds intersect the explicit Host catalog; IDs never grant images.
func classifierImageLimit(model ClassifierModel) int {
	if model.API == APIOpenAIDecisions {
		return min(model.Capabilities.MaxInputImages, decisionsImageLimit)
	}
	return model.Capabilities.MaxInputImages
}

// Before Binding resolution, an explicitly declared native-asset candidate
// can upload a source once. If an ID exists on multiple protocols, use this
// less restrictive body estimate only; actual Decisions expansion is checked
// again after resolution. This never grants protocol or image authority.
func (c *Client) classifierInlineImages(target Target) bool {
	for _, model := range c.catalog.ClassifierModels {
		if model.ID == target.ModelID && model.API == APITypeSafeSystemOne && slices.Contains(model.Input, ModalityImage) {
			return false
		}
	}
	return true
}
