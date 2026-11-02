package ai

// imageDefinitionProblem compares a coherent model declaration with the option
// and input guards implemented by its protocol. General malformed metadata is
// still diagnosed by the existing Catalog validation contract.
func imageDefinitionProblem(m ImageModel) string {
	c := m.Capabilities
	switch m.API {
	case APIGoogleInteractions:
		// GoogleImagesOptions and the frozen response_format accept only
		// aspect_ratio/image_size. The payload guard rejects mask/background;
		// there is no request path for the other declared settings.
		if c.Mask || c.TransparentBackground || c.CustomSizes != nil || len(c.InputFidelity) != 0 || len(c.Sizes) != 0 || len(c.Qualities) != 0 {
			return "Google image options do not implement the declared capability"
		}
		if c.MaxReferenceImages > googleImagesReferenceLimit {
			return "Google image reference declaration exceeds the implemented input guard"
		}
	case APIOpenAIImages:
		if len(c.AspectRatios) != 0 || len(c.ImageSizes) != 0 {
			return "OpenAI image options do not implement aspect ratio or image size declarations"
		}
		if c.MaxReferenceImages > openAIReferenceLimit {
			return "OpenAI image reference declaration exceeds the implemented input guard"
		}
		// Use the same closed quantity/quality/fidelity constraints as real
		// requests. Model metadata must not advertise a value those reject.
		if problem := (OpenAIImagesOptions{N: Value(c.MaxOutputImages)}).validate(); problem != "" {
			return problem
		}
		for _, quality := range c.Qualities {
			if problem := (OpenAIImagesOptions{Quality: Value(quality)}).validate(); problem != "" {
				return problem
			}
		}
		for _, fidelity := range c.InputFidelity {
			if problem := (OpenAIImagesOptions{InputFidelity: Value(fidelity)}).validate(); problem != "" {
				return problem
			}
		}
	}
	return ""
}
