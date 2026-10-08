package ai

import (
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

// The JSON edit schema bounds each complete inline URL, not only its bytes.
const openAIInlineURLLimit = 20971520
const openAIReferenceLimit = 16

func (p *ResourcePolicy) checkInlineImageSize(data, mime string, phase Phase) *Error {
	if int64(len(data))+int64(len(mime))+int64(len("data:;base64,")) > openAIInlineURLLimit {
		return limitFailure(phase, "OpenAI.MaxInlineImageURLCharacters", openAIInlineURLLimit)
	}
	if base64Size(data) > p.MaxImageBytes {
		return limitFailure(phase, "MaxImageBytes", p.MaxImageBytes)
	}
	return nil
}

func openAIImageReference(img Image) openAIInlineImage {
	return openAIInlineImage{ImageURL: "data:" + img.MimeType + ";base64," + img.Data}
}

func parseOpenAIImageURL(url string) (Image, bool) {
	mime, data, ok := strings.Cut(url, ";base64,")
	mime, inline := strings.CutPrefix(mime, "data:")
	return Image{Data: data, MimeType: mime}, ok && inline && isImageMediaType(mime)
}

func imageMaskDimensionsMatch(first, mask Image) bool {
	// DecodeConfig reads headers from strict base64 streams. No raster or
	// decoded image copy is allocated, even for enormous dimensions.
	a, _, errA := image.DecodeConfig(base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(first.Data)))
	b, _, errB := image.DecodeConfig(base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(mask.Data)))
	return errA == nil && errB == nil && a.Width > 0 && a.Height > 0 && a.Width == b.Width && a.Height == b.Height
}

func (p *ResourcePolicy) checkImageCapabilities(req ImagesRequest, opts OpenAIImagesOptions, caps ImageCapabilities, phase Phase) *Error {
	if len(req.ReferenceImages) > caps.MaxReferenceImages {
		return limitFailure(phase, "ImageModel.MaxReferenceImages", int64(caps.MaxReferenceImages))
	}
	if len(req.ReferenceImages) == 0 && (opts.Mask != nil || !opts.InputFidelity.IsZero()) {
		return imageInputFailure(phase, "mask and input fidelity require reference images")
	}
	if problem := opts.check(caps, p.Image); problem != "" {
		return imageInputFailure(phase, problem)
	}
	return nil
}
