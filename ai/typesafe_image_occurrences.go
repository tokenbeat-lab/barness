package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

// Freeze actual whole-value occurrences independently of the public pointers.
// Walk native State in deterministic object-key/array order. A projected context
// envelope has images before state in that order. Only its topology is retained
// from the builder; every content value, asset, permission and budget is rechecked.
func typeSafeImageOccurrences(ctx context.Context, req typeSafeRequest, locations []ClassifierImageLocation, contextWrapped bool, maxReferences int) ([]string, *Error) {
	if failure := typeSafeImageKeys(ctx, req); failure != nil {
		return nil, failure
	}
	var occurrences []string
	add := func(key string, location ClassifierImageLocation) *Error {
		if !slices.Contains(locations, location) {
			return typeSafeRequestFailure()
		}
		if len(occurrences) >= maxReferences {
			return limitFailure(PhaseRequest, "Classifier.Images.References.MaxReferences", int64(maxReferences))
		}
		occurrences = append(occurrences, key)
		return nil
	}
	contains := func(text string) bool {
		for key := range req.Assets {
			if strings.Contains(text, key) {
				return true
			}
		}
		return false
	}
	var walk func(any) *Error
	walk = func(value any) *Error {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return failure
		}
		switch node := value.(type) {
		case string:
			if _, exists := req.Assets[node]; exists {
				return add(node, ClassifierImageLocationState)
			}
			if contains(node) {
				return typeSafeRequestFailure()
			}
		case map[string]any:
			for _, key := range slices.Sorted(maps.Keys(node)) {
				if contains(key) {
					return typeSafeRequestFailure()
				}
				if failure := walk(node[key]); failure != nil {
					return failure
				}
			}
		case []any:
			for _, item := range node {
				if failure := walk(item); failure != nil {
					return failure
				}
			}
		}
		return nil
	}
	state := req.State
	if contextWrapped {
		var envelope typeSafeContext
		if decodeClassifierJSON(state, &envelope) != nil || len(envelope.Images) == 0 || !classifierState(envelope.State) || contains("state") || contains("images") {
			return nil, typeSafeRequestFailure()
		}
		for _, key := range envelope.Images {
			if _, exists := req.Assets[key]; !exists {
				return nil, typeSafeRequestFailure()
			}
			if failure := add(key, ClassifierImageLocationContext); failure != nil {
				return nil, failure
			}
		}
		state = envelope.State
	}
	var tree any
	decoder := json.NewDecoder(bytes.NewReader(state))
	decoder.UseNumber()
	if decoder.Decode(&tree) != nil {
		return nil, typeSafeRequestFailure()
	}
	if failure := walk(tree); failure != nil {
		return nil, failure
	}
	return occurrences, nil
}
