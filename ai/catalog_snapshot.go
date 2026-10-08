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
	m.Cost.Tiers = slices.Clone(m.Cost.Tiers)
	return m
}

func (m ImageModel) clone() ImageModel {
	m.Input, m.Output = slices.Clone(m.Input), slices.Clone(m.Output)
	m.Capabilities.Sizes = slices.Clone(m.Capabilities.Sizes)
	m.Capabilities.ImageSizes = slices.Clone(m.Capabilities.ImageSizes)
	m.Capabilities.AspectRatios = slices.Clone(m.Capabilities.AspectRatios)
	m.Capabilities.InputFidelity = slices.Clone(m.Capabilities.InputFidelity)
	return m
}

func (m ClassifierModel) clone() ClassifierModel {
	m.Capabilities.Kinds = slices.Clone(m.Capabilities.Kinds)
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
