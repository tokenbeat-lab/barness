package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
)

func (c Catalog) clone() Catalog {
	out := c
	out.Models = slices.Clone(c.Models)
	for i := range out.Models {
		out.Models[i] = out.Models[i].clone()
	}
	out.ImageModels = slices.Clone(c.ImageModels)
	for i := range out.ImageModels {
		out.ImageModels[i] = out.ImageModels[i].clone()
	}
	out.ClassifierModels = slices.Clone(c.ClassifierModels)
	for i := range out.ClassifierModels {
		out.ClassifierModels[i] = out.ClassifierModels[i].clone()
	}
	return out
}

func (m Model) clone() Model {
	m.Input = slices.Clone(m.Input)
	m.SamplingParams = cloneJSONValues(m.SamplingParams)
	if m.SamplingParamsByThinkingLevel != nil {
		profiles := make(map[ThinkingLevel]map[string]json.RawMessage, len(m.SamplingParamsByThinkingLevel))
		for level, params := range m.SamplingParamsByThinkingLevel {
			profiles[level] = cloneJSONValues(params)
		}
		m.SamplingParamsByThinkingLevel = profiles
	}
	m.Cost.Tiers = slices.Clone(m.Cost.Tiers)
	return m
}

func (m ImageModel) clone() ImageModel {
	m.Input, m.Output = slices.Clone(m.Input), slices.Clone(m.Output)
	m.Capabilities.Sizes = slices.Clone(m.Capabilities.Sizes)
	m.Capabilities.Qualities = slices.Clone(m.Capabilities.Qualities)
	m.Capabilities.ImageSizes = slices.Clone(m.Capabilities.ImageSizes)
	m.Capabilities.AspectRatios = slices.Clone(m.Capabilities.AspectRatios)
	m.Capabilities.InputFidelity = slices.Clone(m.Capabilities.InputFidelity)
	if m.Capabilities.CustomSizes != nil {
		constraints := *m.Capabilities.CustomSizes
		m.Capabilities.CustomSizes = &constraints
	}
	if m.Pricing.PerImage != nil {
		prices := *m.Pricing.PerImage
		prices.Output = slices.Clone(prices.Output)
		m.Pricing.PerImage = &prices
	}
	return m
}

func (m ClassifierModel) clone() ClassifierModel {
	m.Input = slices.Clone(m.Input)
	m.Capabilities.Kinds = slices.Clone(m.Capabilities.Kinds)
	m.Capabilities.DescriptionFormats = slices.Clone(m.Capabilities.DescriptionFormats)
	m.Capabilities.ImageMIMETypes = slices.Clone(m.Capabilities.ImageMIMETypes)
	m.Capabilities.ImageLocations = slices.Clone(m.Capabilities.ImageLocations)
	m.Cost.Tiers = slices.Clone(m.Cost.Tiers)
	return m
}

// Hash is "sha256:" and the hex SHA-256 of c's JSON encoding, which is
// deterministic. CallMetadata.CatalogHash carries it. It fails for a
// catalog NewClient would refuse: a non-finite price cannot be encoded.
func (c Catalog) Hash() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
