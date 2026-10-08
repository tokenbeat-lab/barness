package ai

import (
	"encoding/json"
	"maps"
	"math"
	"slices"
	"strconv"
)

// Options are the complete, protocol-specific options for the full entry
// points (Stream, Complete, Classify). Each implementation belongs to exactly one API
// and must match the binding's API; there is no namespace keyed by vendor
// name. A nil Options means protocol defaults.
//
// Options carry no authority: no key, endpoint, header, transport or
// function. Their JSON shape is pi-ai's option shape, so presence (unset,
// null, value) survives decoding.
type Options interface {
	api() API
	// clone copies everything the caller could still mutate.
	clone() Options
	// validate reports why the options are unusable, or "".
	validate() string
}

// SimpleOptions are the protocol-neutral options for StreamSimple and
// CompleteSimple (pi-ai's SimpleStreamOptions). They are mapped per model
// onto the binding's protocol: the reasoning level is clamped to what the
// model supports and translated by its level map, and the output budget
// defaults to the model's and is clamped to its context window.
type SimpleOptions struct {
	// Reasoning is the reasoning level; empty leaves reasoning to the model's
	// default (its "off" mapping, where it has one).
	Reasoning ThinkingLevel `json:"reasoning,omitempty"`
	// ThinkingBudgets replace the default token budget per level on
	// token-budget protocols.
	ThinkingBudgets ThinkingBudgets `json:"thinkingBudgets,omitzero"`
	// ToolChoice selects auto or no tool use; unset leaves it to the
	// provider, null is sent as null.
	ToolChoice Nullable[ToolChoice] `json:"toolChoice,omitzero"`
	// Temperature is sent as given, null included.
	Temperature Nullable[float64] `json:"temperature,omitzero"`
	// MaxTokens caps output tokens; unset or null takes the model's maximum.
	// Either way it is clamped to what the context window leaves.
	MaxTokens Nullable[int] `json:"maxTokens,omitzero"`
	// SamplingParams are merged key by key over the model's own and applied
	// by OpenAI-compatible protocols after, and over, the named request
	// fields. Other protocols ignore them. Keys that would change the model,
	// history, tools or server-side state are refused.
	SamplingParams map[string]json.RawMessage `json:"samplingParams,omitempty"`
	// CacheRetention is the prompt cache retention; empty is "short".
	CacheRetention CacheRetention `json:"cacheRetention,omitempty"`
	// SessionID enables session-scoped prompt caching and affinity. It is
	// never sent as is: the library derives the identifiers it sends from it,
	// the tenant and the vendor account.
	SessionID string `json:"sessionId,omitempty"`
	// Metadata is sent by protocols with a metadata field, as those protocols
	// define; Responses does not send it. The library never adds the tenant.
	Metadata map[string]json.RawMessage `json:"metadata,omitempty"`
	// TimeoutMs is the protocol's request timeout for each attempt, passed on
	// as the protocol defines it; 0 keeps the protocol's default. The policy's
	// limits apply when they are earlier. It must not be negative.
	TimeoutMs int `json:"timeoutMs,omitempty"`
}

func (o SimpleOptions) clone() SimpleOptions {
	o.SamplingParams = cloneJSONValues(o.SamplingParams)
	o.Metadata = cloneJSONValues(o.Metadata)
	return o
}

// validate checks the options before any mapping.
func (o SimpleOptions) validate() string {
	if o.Reasoning != "" && !o.Reasoning.valid() {
		return "reasoning must be minimal, low, medium, high, xhigh or max"
	}
	if problem := o.ThinkingBudgets.validate(); problem != "" {
		return problem
	}
	if c, ok := o.ToolChoice.Get(); ok && c != ToolChoiceAuto && c != ToolChoiceNone {
		return "tool choice must be auto or none"
	}
	if problem := validateJSONValues("metadata", o.Metadata); problem != "" {
		return problem
	}
	return validateCommon(o.Temperature, o.MaxTokens, o.CacheRetention, o.SamplingParams, o.TimeoutMs)
}

// resolve is pi's buildBaseOptions for model m and the request: MaxTokens
// resolved to the clamped output budget. Sampling defaults are applied once
// by the OpenAI-compatible request builders, on full and simple entries.
func (o SimpleOptions) resolve(m Model, req Request) SimpleOptions {
	maxTokens := m.MaxTokens
	if v, ok := o.MaxTokens.Get(); ok {
		maxTokens = v
	}
	o.MaxTokens = Value(clampMaxTokensToContext(m, normalizeTranscript(req), maxTokens))
	return o
}

// CacheRetention is the prompt cache retention preference.
type CacheRetention string

const (
	CacheRetentionNone  CacheRetention = "none"
	CacheRetentionShort CacheRetention = "short"
	CacheRetentionLong  CacheRetention = "long"
)

// validateCommon checks the options every protocol shares.
func validateCommon(temperature Nullable[float64], maxTokens Nullable[int], retention CacheRetention, sampling map[string]json.RawMessage, timeoutMs int) string {
	if t, ok := temperature.Get(); ok && (math.IsNaN(t) || math.IsInf(t, 0)) {
		return "temperature must be a finite number"
	}
	if timeoutMs < 0 {
		return "timeoutMs must not be negative"
	}
	if n, ok := maxTokens.Get(); ok && n < 0 {
		return "maxTokens must not be negative"
	}
	switch retention {
	case "", CacheRetentionNone, CacheRetentionShort, CacheRetentionLong:
	default:
		return "cache retention must be none, short or long"
	}
	return validateJSONValues("samplingParams", sampling)
}

// validateJSONValues checks that every value of an arbitrary option object
// is one JSON value. Keys are checked in order so the error is deterministic.
func validateJSONValues(field string, values map[string]json.RawMessage) string {
	for _, k := range slices.Sorted(maps.Keys(values)) {
		if !json.Valid(values[k]) {
			return field + "." + k + " is not a JSON value"
		}
	}
	return ""
}

// cloneJSONValues deep-copies an arbitrary option object.
func cloneJSONValues(values map[string]json.RawMessage) map[string]json.RawMessage {
	if values == nil {
		return nil
	}
	out := make(map[string]json.RawMessage, len(values))
	for k, v := range values {
		out[k] = slices.Clone(v)
	}
	return out
}

// checkReservedKeys refuses samplingParams keys that would cross the call's
// boundary when applied over the request: the authorized model, the
// prepared history and its native state, the declared tools (which could
// add hosted tools with their own targets and credentials), server-side
// state or stored prompts of the account, the cache key derived per tenant,
// and the adapter's own wire contract. reserved lists the protocol's names
// for them. Every other key keeps pi's "last overrides named fields".
func checkReservedKeys(sampling map[string]json.RawMessage, reserved []string) string {
	for _, k := range reserved {
		if _, ok := sampling[k]; ok {
			return "samplingParams may not set " + strconv.Quote(k)
		}
	}
	return ""
}
