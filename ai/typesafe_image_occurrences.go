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
// Walk State, then sorted questions: Instructions before Criteria, with sorted
// object/Choice/Bool keys and original array/Score order. A projected context
// envelope has images before state in that order. Only its topology is retained
// from the builder; every content value, asset, permission and budget is rechecked.
func typeSafeImageOccurrences(ctx context.Context, req typeSafeRequest, locations []ClassifierImageLocation, contextWrapped bool, maxReferences int) ([]string, *Error) {
	if failure := typeSafeImageKeys(ctx, req); failure != nil {
		return nil, failure
	}
	var occurrences []string
	sharedVisible := false
	localVisible := map[string]bool{}
	questionName := ""
	add := func(key string, location ClassifierImageLocation) *Error {
		if !slices.Contains(locations, location) {
			return typeSafeRequestFailure()
		}
		if len(occurrences) >= maxReferences {
			return limitFailure(PhaseRequest, "Classifier.Images.References.MaxReferences", int64(maxReferences))
		}
		if location == ClassifierImageLocationContext || location == ClassifierImageLocationState {
			sharedVisible = true
		} else if location == ClassifierImageLocationInstructions {
			localVisible[questionName] = true
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
	var walk func(any, ClassifierImageLocation) *Error
	walk = func(value any, location ClassifierImageLocation) *Error {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return failure
		}
		switch node := value.(type) {
		case string:
			if _, exists := req.Assets[node]; exists {
				return add(node, location)
			}
			if contains(node) {
				return typeSafeRequestFailure()
			}
		case map[string]any:
			for _, key := range slices.Sorted(maps.Keys(node)) {
				if contains(key) {
					return typeSafeRequestFailure()
				}
				if failure := walk(node[key], location); failure != nil {
					return failure
				}
			}
		case []any:
			for _, item := range node {
				if failure := walk(item, location); failure != nil {
					return failure
				}
			}
		}
		return nil
	}
	state := req.State
	if contextWrapped {
		var envelope typeSafeContext
		// A Hook may remove every context occurrence while keeping local uses.
		// Preserve the builder topology: [] is valid, missing/null is not.
		if decodeClassifierJSON(state, &envelope) != nil || envelope.Images == nil || !classifierState(envelope.State) || contains("state") || contains("images") {
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
	if failure := walk(tree, ClassifierImageLocationState); failure != nil {
		return nil, failure
	}
	for _, name := range slices.Sorted(maps.Keys(req.Questions)) {
		questionName = name
		q := req.Questions[name]
		contents := []json.RawMessage{q.instructions}
		for _, key := range slices.Sorted(maps.Keys(q.choices)) {
			contents = append(contents, q.choices[key])
		}
		contents = append(contents, q.levels...)
		if q.boolean != nil {
			contents = append(contents, q.boolean.False, q.boolean.True)
		}
		for i, raw := range contents {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if decoder.Decode(&tree) != nil {
				return nil, typeSafeRequestFailure()
			}
			location := ClassifierImageLocationCriteria
			if i == 0 {
				location = ClassifierImageLocationInstructions
			}
			if failure := walk(tree, location); failure != nil {
				return nil, failure
			}
		}
	}
	for name, q := range req.Questions {
		if q.kind == "box" && !sharedVisible && !localVisible[name] {
			return nil, typeSafeRequestFailure()
		}
	}
	return occurrences, nil
}
