package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

// GLM, Kimi and xAI expose opaque source IDs. Allocate replacements only when
// the target normalization would collapse distinct provenance owners.
// Inspect the complete transcript before renaming so calls from separate
// assistant turns share one collision domain with their matching results.
func opaqueHistoryToolIDRules(target Model, messages []Message, rules historyRules) historyRules {
	if target.Provider == ProviderXAI {
		return rules
	}
	if rules.normalizeToolCallID == nil || (target.API == APIOpenAIResponses && responsesCapabilitiesOf(target.Provider).implementation == responsesGLM) {
		return rules
	}
	if !slices.ContainsFunc(messages, func(message Message) bool {
		a, ok := message.(AssistantMessage)
		return ok && opaqueHistoryIDSource(a)
	}) {
		return rules
	}
	normalize := rules.normalizeToolCallID
	wireID := func(id string) string {
		if target.API == APIOpenAIResponses {
			switch responsesCapabilitiesOf(target.Provider).implementation {
			case responsesKimi:
				if call, err := kimiResponsesCallID(id); err == nil {
					return call
				}
			case responsesMiniMax:
				if call, err := miniMaxResponsesCallID(id); err == nil {
					return call
				}
			}
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
		owner          string
		source         AssistantMessage
	}
	var candidates []candidate
	turn := 0
	for _, message := range messages {
		assistant, ok := message.(AssistantMessage)
		if !ok {
			continue
		}
		turn++
		if assistant.StopReason == StopReasonError || assistant.StopReason == StopReasonAborted {
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
			// overwritten by a replacement assigned to an opaque source call.
			owner := opaqueHistoryOwner(call.ID, assistant, turn)
			reserve(call.ID, owner)
			reserve(normalized, owner)
			if opaqueHistoryIDSource(assistant) {
				candidates = append(candidates, candidate{call.ID, normalized, owner, assistant})
			}
		}
	}
	replacements := map[string]string{}
	slices.SortFunc(candidates, func(a, b candidate) int { return strings.Compare(a.id, b.id) })
	for _, c := range candidates {
		if wireID(c.normalized) != "" && !collisions[wireID(c.normalized)] {
			continue
		}
		if _, done := replacements[c.owner]; done {
			continue
		}
		sum := sha256.Sum256([]byte(c.id))
		// Retain the established generated wire spelling for persisted histories;
		// this prefix carries no Provider attribution or replay authority.
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
		replacements[c.owner] = id
		owners[wireID(id)] = c.owner
	}
	if len(replacements) == 0 {
		return rules
	}
	activeTurn := 0
	previousBegin := rules.beginAssistant
	rules.beginAssistant = func() {
		if previousBegin != nil {
			previousBegin()
		}
		activeTurn++
	}
	rules.normalizeToolCallID = func(id string, source AssistantMessage) string {
		if opaqueHistoryIDSource(source) {
			if replacement, ok := replacements[opaqueHistoryOwner(id, source, activeTurn)]; ok {
				return replacement
			}
		}
		return normalize(id, source)
	}
	return rules
}

// Each source grammar belongs to its protocol. Codec equality only identifies
// opaque ID syntax; replayOrigin still compares the exact real Provider.
func opaqueHistoryIDSource(source AssistantMessage) bool {
	if isKimi(source.Provider) {
		return source.API == APIOpenAICompletions || source.API == APIOpenAIResponses || source.API == APIAnthropicMessages
	}
	switch source.API {
	case APIOpenAICompletions:
		return chatCompatOf(source.Provider).implementation == chatGLM
	case APIOpenAIResponses:
		return source.Provider == ProviderXAI || responsesCapabilitiesOf(source.Provider).implementation == responsesGLM
	case APIAnthropicMessages:
		return anthropicCapabilitiesOf(source.Provider).implementation == anthropicGLM
	}
	return false
}

// Source identity is correlation context, not authorization. Native replay
// separately compares the complete trusted envelope with the destination.
func opaqueHistoryOwner(id string, a AssistantMessage, turn int) string {
	env, _ := a.NativeState.Envelope()
	key, _ := json.Marshal([7]string{string(a.Provider), string(a.API), a.Model, env.TenantID, env.AccountScopeID, id, strconv.Itoa(turn)})
	return string(key)
}
