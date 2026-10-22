package ai

import (
	"encoding/json"
	"github.com/openai/openai-go/v3/responses"
	"strings"
)

type kimiResponsesState struct {
	finished map[responsesIndex]responsesSlot
}

func newKimiResponsesState(p ProviderID) *kimiResponsesState {
	if !isKimi(p) {
		return nil
	}
	return &kimiResponsesState{finished: map[responsesIndex]responsesSlot{}}
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
