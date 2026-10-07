package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
)

func (c Catalog) clone() Catalog {
	out := Catalog{Version: c.Version, Models: make([]Model, len(c.Models))}
	for i, m := range c.Models {
		m.Input = append([]Modality(nil), m.Input...)
		m.SamplingParams = cloneJSONValues(m.SamplingParams)
		m.Cost.Tiers = slices.Clone(m.Cost.Tiers)
		out.Models[i] = m
	}
	return out
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

// validate refuses a catalog whose prices could not yield a finite,
// non-negative cost.
func (c Catalog) validate() error {
	for _, m := range c.Models {
		if problem := m.Cost.problem(); problem != "" {
			return &ConfigError{Field: "Catalog", Problem: "model " + m.ID + ": " + problem}
		}
		var reserved []string
		switch m.API {
		case APIOpenAIResponses:
			reserved = responsesReservedSampling
		case APIOpenAICompletions:
			reserved = chatReservedSampling
		default:
			continue // These protocols do not apply samplingParams.
		}
		problem := validateJSONValues("samplingParams", m.SamplingParams)
		if problem == "" {
			problem = checkReservedKeys(m.SamplingParams, reserved)
		}
		if problem != "" {
			return &ConfigError{Field: "Catalog", Problem: "model " + m.ID + ": " + problem}
		}
	}
	return nil
}
