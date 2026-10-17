package ai

import (
	"slices"
	"strings"

	"github.com/openai/openai-go/v3/responses"
)

type miniMaxResponsesState struct {
	finished  map[responsesIndex]responsesSlot
	ids       map[responsesIndex]string
	doneParts map[responsesIndex]map[int64]string
}

func newMiniMaxResponsesState(provider ProviderID) *miniMaxResponsesState {
	if provider != ProviderMiniMaxCN {
		return nil
	}
	return &miniMaxResponsesState{finished: map[responsesIndex]responsesSlot{}, ids: map[responsesIndex]string{}, doneParts: map[responsesIndex]map[int64]string{}}
}

func (p *responsesParser) miniMaxTextDone(ev responses.ResponseStreamEventUnion) {
	kind := "message"
	if strings.Contains(ev.Type, "reasoning_") {
		kind = "reasoning"
	}
	key := outputIndexOf(ev)
	slot, ok := p.slots[key]
	if !ok {
		p.open(key, responses.ResponseOutputItemUnion{Type: kind})
		slot = p.slots[key]
	}
	text := ev.Text
	if ev.Type == "response.refusal.done" {
		text = ev.Refusal
	}
	partIndex := ev.ContentIndex
	if p.minimax.doneParts[key] == nil {
		p.minimax.doneParts[key] = map[int64]string{}
	}
	p.minimax.doneParts[key][partIndex] = text
	indexes := make([]int64, 0, len(p.minimax.doneParts[key]))
	for index := range p.minimax.doneParts[key] {
		indexes = append(indexes, index)
	}
	slices.Sort(indexes)
	parts := make([]string, 0, len(indexes))
	for _, index := range indexes {
		parts = append(parts, p.minimax.doneParts[key][index])
	}
	separator := ""
	if kind == "reasoning" {
		separator = "\n\n"
	}
	text = strings.Join(parts, separator)
	slot.done = &text
	p.slots[key] = slot
}

// Terminal output is authoritative, and can be the only source of items.
// Finished slots are updated without publishing a second block-end event.
func (p *responsesParser) finishMiniMaxOutput(output []responses.ResponseOutputItemUnion) {
	for i, item := range output {
		key := responsesIndex{kind: "number", value: int64(i)}
		if slot, done := p.minimax.finished[key]; done {
			p.updateMiniMaxFinished(slot, item)
			continue
		}
		p.itemDone(key, item)
	}
	keys := make([]responsesIndex, 0, len(p.slots))
	for key := range p.slots {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b responsesIndex) int { return p.slots[a].index - p.slots[b].index })
	for _, key := range keys {
		slot := p.slots[key]
		if slot.kind == "function_call" {
			continue
		}
		text := ""
		if slot.kind == "message" {
			text = p.out.textContent(slot.index)
		} else {
			text = p.out.thinkingText(slot.index)
		}
		if slot.done != nil {
			text = *slot.done
		}
		if slot.kind == "message" {
			p.out.textEnd(slot.index, text, "")
		} else {
			p.out.thinkingEnd(slot.index, text, "")
		}
		p.minimax.finished[key] = slot
		delete(p.slots, key)
	}
}

func (p *responsesParser) updateMiniMaxFinished(slot responsesSlot, item responses.ResponseOutputItemUnion) {
	switch item.Type {
	case "message":
		if len(item.Content) == 0 {
			return
		}
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
	case "reasoning":
		parts := []string{}
		for _, part := range item.Content {
			parts = append(parts, part.Text)
		}
		if len(parts) == 0 {
			for _, part := range item.Summary {
				parts = append(parts, part.Text)
			}
		}
		text := strings.Join(parts, "\n\n")
		if len(parts) == 0 {
			text = p.out.thinkingText(slot.index)
		}
		p.out.view.update(func(m *AssistantMessage) {
			b := m.Content[slot.index].(Thinking)
			b.Thinking = text
			b.Signature = item.RawJSON()
			m.Content[slot.index] = b
		})
	case "function_call":
		// A changed authoritative function argument body must still pass the same
		// assembler JSON resource checks as its streamed counterpart.
		p.out.toolCallArguments(slot.index, item.Arguments.OfString)
	}
}
