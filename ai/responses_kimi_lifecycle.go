package ai

import (
	"encoding/json"
	"github.com/openai/openai-go/v3/responses"
	"maps"
	"slices"
	"strings"
)

type kimiResponsesState struct {
	finished    map[responsesIndex]responsesSlot
	identities  map[responsesIndex]kimiResponsesIdentity
	doneItems   map[responsesIndex]json.RawMessage
	doneParts   map[responsesIndex]map[int64]bool
	textParts   map[responsesIndex]map[int64]string
	sequence    int64
	hasSequence bool
}

func newKimiResponsesState(p ProviderID) *kimiResponsesState {
	if !isKimi(p) {
		return nil
	}
	return &kimiResponsesState{finished: map[responsesIndex]responsesSlot{}, identities: map[responsesIndex]kimiResponsesIdentity{}, doneItems: map[responsesIndex]json.RawMessage{}, doneParts: map[responsesIndex]map[int64]bool{}, textParts: map[responsesIndex]map[int64]string{}}
}

// Kimi owns summary as its native reasoning representation. Preserve the raw
// output even with encrypted_content:null; the input projection is separate.
func (p *responsesParser) finishKimiOutput(output []responses.ResponseOutputItemUnion) {
	for i, item := range output {
		key := responsesIndex{kind: "number", value: int64(i)}
		if slot, done := p.kimi.finished[key]; done {
			switch item.Type {
			case "reasoning":
				parts := make([]string, len(item.Summary))
				for j, s := range item.Summary {
					parts[j] = s.Text
				}
				text := strings.Join(parts, "\n\n")
				p.out.view.update(func(m *AssistantMessage) {
					b := m.Content[slot.index].(Thinking)
					b.Thinking = text
					b.Signature = item.RawJSON()
					m.Content[slot.index] = b
				})
			case "message":
				var text strings.Builder
				for _, part := range item.Content {
					if part.Type == "output_text" {
						text.WriteString(part.Text)
					} else {
						text.WriteString(part.Refusal)
					}
				}
				p.out.view.update(func(m *AssistantMessage) {
					m.Content[slot.index] = Text{Text: text.String(), Signature: encodeTextSignature(item.ID, string(item.Phase))}
				})
			case "function_call":
				p.out.toolCallArguments(slot.index, item.Arguments.OfString)
			}
			continue
		}
		p.itemDone(key, item)
	}

	keys := slices.Collect(maps.Keys(p.slots))
	slices.SortFunc(keys, func(a, b responsesIndex) int { return p.slots[a].index - p.slots[b].index })
	for _, key := range keys {
		slot := p.slots[key]
		if slot.kind == "function_call" {
			continue
		}
		text := ""
		if slot.kind == "reasoning" {
			text = p.out.thinkingText(slot.index)
		} else {
			text = p.out.textContent(slot.index)
		}
		if slot.done != nil {
			text = *slot.done
		}
		if slot.kind == "message" {
			p.out.textEnd(slot.index, text, "")
		} else {
			p.out.thinkingEnd(slot.index, text, "")
		}
		p.kimi.finished[key] = slot
		delete(p.slots, key)
	}
}
func (p *responsesParser) kimiResponsesTextDone(ev responses.ResponseStreamEventUnion) {
	key := outputIndexOf(ev)
	kind := "message"
	if strings.Contains(ev.Type, "reasoning_summary") {
		kind = "reasoning"
	}
	slot, ok := p.slots[key]
	if !ok {
		p.open(key, responses.ResponseOutputItemUnion{Type: kind})
		slot = p.slots[key]
	}
	part := ev.ContentIndex
	var raw map[string]json.RawMessage
	_ = json.Unmarshal([]byte(ev.RawJSON()), &raw)
	if v, ok := raw["summary_index"]; ok {
		part, _ = kimiResponsesIndex(v)
	}
	if p.kimi.textParts[key] == nil {
		p.kimi.textParts[key] = map[int64]string{}
	}
	p.kimi.textParts[key][part] = ev.Text
	keys := slices.Sorted(maps.Keys(p.kimi.textParts[key]))
	parts := make([]string, 0, len(keys))
	for _, i := range keys {
		parts = append(parts, p.kimi.textParts[key][i])
	}
	separator := ""
	if kind == "reasoning" {
		separator = "\n\n"
	}
	text := strings.Join(parts, separator)
	slot.done = &text
	p.slots[key] = slot
}
func (p *responsesParser) reportKimiResponsesUsage(r responses.Response) *Error {
	var raw struct {
		Usage  json.RawMessage `json:"usage"`
		Actual json.RawMessage `json:"prompt_cache_options"`
	}
	if json.Unmarshal([]byte(r.RawJSON()), &raw) != nil {
		return kimiUsageFailure()
	}
	u, reporting, estimate, failure := kimiUsage(p.usage.model, raw.Usage, raw.Actual)
	if failure != nil {
		return failure
	}
	p.out.usage(u)
	p.attempt.usageEstimate(reporting, u, estimate)
	return nil
}
