package ai

import (
	"slices"

	"github.com/openai/openai-go/v3/responses"
)

// glmTextDone takes authoritative text, rather than treating it as another
// delta. The item.done event closes the block; a terminal can also close it
// when only a text.done was supplied. No replay item is fabricated here.
func (p *responsesParser) glmTextDone(ev responses.ResponseStreamEventUnion) {
	kind := "message"
	if ev.Type == "response.reasoning_text.done" || ev.Type == "response.reasoning_summary_text.done" {
		kind = "reasoning"
	}
	key := outputIndexOf(ev)
	slot, ok := p.slots[key]
	if !ok {
		p.open(key, responses.ResponseOutputItemUnion{Type: kind})
		slot = p.slots[key]
	}
	if slot.kind != kind {
		return
	}
	text := ev.Text
	if ev.Type == "response.refusal.done" {
		text = ev.Refusal
	}
	slot.done = &text
	p.slots[key] = slot
}

// finishGLMOutput handles legal terminal-only output and done-only streams.
// Finished indexes prevent the terminal snapshot from duplicating item.done.
func (p *responsesParser) finishGLMOutput(output []responses.ResponseOutputItemUnion) {
	for i, item := range output {
		key := responsesIndex{kind: "number", value: int64(i)}
		if !p.glmFinished[key] {
			p.itemDone(key, item)
		}
	}
	// Block indexes define stable presentation order even when provider output
	// indexes were announced or finished out of order.
	keys := make([]responsesIndex, 0, len(p.slots))
	for key := range p.slots {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b responsesIndex) int { return p.slots[a].index - p.slots[b].index })
	for _, key := range keys {
		slot := p.slots[key]
		if slot.kind != "message" && slot.kind != "reasoning" {
			continue
		}
		if slot.kind == "message" {
			text := p.out.textContent(slot.index)
			if slot.done != nil {
				text = *slot.done
			}
			p.out.textEnd(slot.index, text, "")
		} else {
			text := p.out.thinkingText(slot.index)
			if slot.done != nil {
				text = *slot.done
			}
			p.out.thinkingEnd(slot.index, text, reasoningSignature(p.glmNative[key]))
		}
		delete(p.slots, key)
	}
}
