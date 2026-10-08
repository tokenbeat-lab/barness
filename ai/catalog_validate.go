package ai

import (
	"math"
	"slices"
	"strings"
)

// modelKey includes the operation: IDs shared by different kinds of models
// cannot collide or acquire each other's authorization.
type modelKey struct {
	op       Operation
	provider ProviderID
	api      API
	id       string
}

// validate checks metadata at the configuration boundary. NewClient and Hash
// use the same rules, so every accepted price snapshot can be encoded.
func (c Catalog) validate() error {
	seen := make(map[modelKey]bool)
	check := func(key modelKey, problem string) error {
		if seen[key] {
			problem = "duplicate model identity for operation, provider, API and ID"
		}
		seen[key] = true
		if problem != "" {
			return &ConfigError{Field: "Catalog", Problem: string(key.op) + " model " + key.id + ": " + problem}
		}
		return nil
	}
	for _, m := range c.Models {
		problem := m.Cost.problem()
		if problem == "" {
			problem = modalitiesProblem(m.Input)
		}
		if problem == "" {
			var reserved []string
			switch m.API {
			case APIOpenAIResponses:
				reserved = responsesReservedSampling
			case APIOpenAICompletions:
				reserved = chatReservedSampling
			}
			if reserved != nil {
				problem = validateJSONValues("samplingParams", m.SamplingParams)
				if problem == "" {
					problem = checkReservedKeys(m.SamplingParams, reserved)
				}
			}
		}
		if err := check(modelKey{OperationChat, m.Provider, m.API, m.ID}, problem); err != nil {
			return err
		}
	}
	for _, m := range c.ImageModels {
		if err := check(modelKey{OperationImage, m.Provider, m.API, m.ID}, m.problem()); err != nil {
			return err
		}
	}
	for _, m := range c.ClassifierModels {
		if err := check(modelKey{OperationClassifier, m.Provider, m.API, m.ID}, m.problem()); err != nil {
			return err
		}
	}
	return nil
}

func modalitiesProblem(modalities []Modality) string {
	seen := make(map[Modality]bool)
	for _, m := range modalities {
		if m != ModalityText && m != ModalityImage {
			return "a modality is not text or image"
		}
		if seen[m] {
			return "a modality is declared more than once"
		}
		seen[m] = true
	}
	return ""
}

func (m ImageModel) problem() string {
	if problem := modalitiesProblem(m.Input); problem != "" {
		return problem
	}
	if problem := modalitiesProblem(m.Output); problem != "" {
		return problem
	}
	if len(m.Input) == 0 || !slices.Contains(m.Output, ModalityImage) {
		return "image models require input and image output"
	}
	caps := m.Capabilities
	if caps.MaxReferenceImages < 0 || caps.MaxOutputImages < 1 {
		return "image capacity limits are invalid"
	}
	if caps.MaxReferenceImages > 0 && !slices.Contains(m.Input, ModalityImage) {
		return "reference images require image input"
	}
	if (caps.Mask || len(caps.InputFidelity) > 0) && caps.MaxReferenceImages == 0 {
		return "mask and input fidelity require reference images"
	}
	if caps.CustomSizes != nil {
		d := caps.CustomSizes
		if d.EdgeMultiple < 1 || d.MaxAspectRatio < 1 || d.MaxEdge < d.EdgeMultiple || d.MinPixels < 1 || d.MaxPixels < d.MinPixels || strings.TrimSpace(d.Source) == "" {
			return "custom image sizes require positive, ordered constraints and an evidence source"
		}
	}
	if m.Provider == ProviderOpenAI && m.API == APIOpenAIImages && (!m.Pricing.CacheReadText.IsZero() || !m.Pricing.CacheReadImage.IsZero()) {
		return "direct OpenAI Images does not support cache-read pricing"
	}
	return m.Pricing.problem(m.Input, m.Output)
}

func (p ImagePricing) problem(input, output []Modality) string {
	for _, rate := range []Nullable[float64]{p.InputText, p.InputImage, p.OutputText, p.OutputImage, p.CacheReadText, p.CacheReadImage} {
		if value, ok := rate.Get(); ok && (value < 0 || math.IsNaN(value) || math.IsInf(value, 0)) {
			return "an image price is negative or not finite"
		}
	}
	for _, required := range []struct {
		modalities []Modality
		modality   Modality
		rate       Nullable[float64]
	}{
		{input, ModalityText, p.InputText}, {input, ModalityImage, p.InputImage},
		{output, ModalityText, p.OutputText}, {output, ModalityImage, p.OutputImage},
	} {
		if _, ok := required.rate.Get(); !ok && slices.Contains(required.modalities, required.modality) {
			return "a declared image modality has no price"
		}
	}
	return ""
}

func (m ClassifierModel) problem() string {
	if m.ContextWindow <= 0 {
		return "classifier context window must be positive"
	}
	caps := m.Capabilities
	if len(caps.Kinds) == 0 {
		return "classifier must support at least one question kind"
	}
	seen := make(map[ClassifierQuestionKind]bool)
	for _, kind := range caps.Kinds {
		if kind != ClassifierQuestionChoice && kind != ClassifierQuestionScore && kind != ClassifierQuestionBool {
			return "classifier question kind is not supported"
		}
		if seen[kind] {
			return "classifier question kind is declared more than once"
		}
		seen[kind] = true
	}
	if (seen[ClassifierQuestionChoice] && caps.MaxChoices < 2) || (!seen[ClassifierQuestionChoice] && caps.MaxChoices != 0) {
		return "classifier choice limit does not match its question kinds"
	}
	if seen[ClassifierQuestionScore] {
		if caps.MinScoreLevels < 2 || caps.MaxScoreLevels < caps.MinScoreLevels {
			return "classifier score level range is invalid"
		}
	} else if caps.MinScoreLevels != 0 || caps.MaxScoreLevels != 0 {
		return "classifier score limits require score questions"
	}
	return m.Cost.problem()
}
