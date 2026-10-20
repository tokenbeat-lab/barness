package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
)

// Other protocols retain their existing normalization. Only GLM
// source IDs that would collide on the target wire need a local replacement;
// ordinary GLM histories and every history without this source stay unchanged.
// Inspect the complete transcript before renaming so calls from separate
// assistant turns share one collision domain with their matching results.
func glmHistoryToolIDRules(target Model, messages []Message, rules historyRules) historyRules {
	if rules.normalizeToolCallID == nil || (target.API == APIOpenAIResponses && responsesCapabilitiesOf(target.Provider).implementation == responsesGLM) {
		return rules
	}
	if !slices.ContainsFunc(messages, func(message Message) bool {
		a, ok := message.(AssistantMessage)
		return ok && glmHistoryIDSource(a)
	}) {
		return rules
	}
	normalize := rules.normalizeToolCallID
	wireID := func(id string) string {
		if target.API == APIOpenAIResponses {
			id, _, _ = strings.Cut(id, "|")
		}
		return id
	}
	owners, collisions := map[string]string{}, map[string]bool{}
	reserve := func(id, sourceID string) {
		id = wireID(id)
		if previous, ok := owners[id]; ok && previous != sourceID {
			collisions[id] = true
		} else {
			owners[id] = sourceID
		}
	}
	type candidate struct {
		id, normalized string
		source         AssistantMessage
	}
	var candidates []candidate
	for _, message := range messages {
		assistant, ok := message.(AssistantMessage)
		if !ok || assistant.StopReason == StopReasonError || assistant.StopReason == StopReasonAborted {
			continue
		}
		for _, block := range assistant.Content {
			call, ok := block.(ToolCall)
			if !ok {
				continue
			}
			normalized := normalize(call.ID, assistant)
			// Reserve both spellings: native same-model legacy histories do
			// not normalize, while cross-model histories do. Neither may be
			// overwritten by a replacement assigned to a GLM source call.
			reserve(call.ID, call.ID)
			reserve(normalized, call.ID)
			if glmHistoryIDSource(assistant) {
				candidates = append(candidates, candidate{call.ID, normalized, assistant})
			}
		}
	}
	replacements := map[string]string{}
	slices.SortFunc(candidates, func(a, b candidate) int { return strings.Compare(a.id, b.id) })
	for _, c := range candidates {
		if wireID(c.normalized) != "" && !collisions[wireID(c.normalized)] {
			continue
		}
		if _, done := replacements[c.id]; done {
			continue
		}
		sum := sha256.Sum256([]byte(c.id))
		id := normalize("glm_"+hex.EncodeToString(sum[:16]), c.source)
		if _, occupied := owners[wireID(id)]; occupied {
			// A caller can deliberately preoccupy the hash spelling. There
			// are at most len(owners) occupied IDs, so this bounded numeric
			// allocation always has a free spelling (all stay below 40).
			for n := 0; n <= len(owners); n++ {
				id = normalize("glm_collision_"+strconv.Itoa(n), c.source)
				if _, occupied := owners[wireID(id)]; !occupied {
					break
				}
			}
		}
		replacements[c.id] = id
		owners[wireID(id)] = c.id
	}
	if len(replacements) == 0 {
		return rules
	}
	rules.normalizeToolCallID = func(id string, source AssistantMessage) string {
		if glmHistoryIDSource(source) {
			if replacement, ok := replacements[id]; ok {
				return replacement
			}
		}
		return normalize(id, source)
	}
	return rules
}

// Each source grammar belongs to its protocol. Codec equality only identifies
// opaque ID syntax; replayOrigin still compares the exact real Provider.
func glmHistoryIDSource(source AssistantMessage) bool {
	switch source.API {
	case APIOpenAICompletions:
		return chatCompatOf(source.Provider).implementation == chatGLM
	case APIOpenAIResponses:
		return responsesCapabilitiesOf(source.Provider).implementation == responsesGLM
	case APIAnthropicMessages:
		return anthropicCapabilitiesOf(source.Provider).implementation == anthropicGLM
	}
	return false
}
