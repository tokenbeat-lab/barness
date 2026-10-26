package ai

import (
	"cmp"
	"encoding/json"
	"strconv"
	"strings"
)

// responsesAssistantItems replays a previous turn prepared by
// prepareTranscript: a signed thinking block as its reasoning item, text as
// output messages and tool calls as function_call items.
func responsesAssistantItems(model Model, m replayedAssistant, msgIndex int) ([]any, error) {
	if model.Provider == ProviderXAI {
		return xaiResponsesAssistantItems(m)
	}
	if isKimi(model.Provider) {
		return kimiResponsesAssistantItems(m)
	}
	sameProviderAPI := m.Provider == model.Provider && m.API == model.API
	// A same provider/API message whose native state was downgraded is
	// replayed as pi replays another model's.
	differentModel := sameProviderAPI && !m.sameModel
	var items []any
	textIndex := 0
	for _, block := range m.Content {
		switch b := block.(type) {
		case Thinking:
			if b.Signature == "" {
				if responsesCapabilitiesOf(model.Provider).implementation == responsesMiniMax && b.Thinking != "" {
					items = append(items, responsesUserMessage{Role: "assistant", Content: []any{responsesInputText{Type: "output_text", Text: b.Thinking}}})
				}
				if responsesCapabilitiesOf(model.Provider).implementation == responsesGLM && b.Thinking != "" {
					items = append(items, glmResponsesAssistantText(b.Thinking))
				}
				continue
			}
			if responsesCapabilitiesOf(model.Provider).implementation == responsesMiniMax {
				item, err := miniMaxResponsesReasoning(b)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
				continue
			}
			if responsesCapabilitiesOf(model.Provider).implementation == responsesGLM {
				item, err := glmResponsesReasoning(b)
				if err != nil {
					return nil, err
				}
				if item != nil {
					items = append(items, item)
				}
				continue
			}
			// pi sends JSON.parse(thinkingSignature). Only same-model state
			// under a trusted envelope gets here, so a signature that is not
			// JSON is a host fault: invalid_request, never a silent drop.
			item, err := parseJSON(b.Signature)
			if err != nil {
				return nil, err
			}
			items = append(items, json.RawMessage(stringifyJSON(item)))
		case Text:
			if responsesCapabilitiesOf(model.Provider).implementation == responsesMiniMax {
				items = append(items, responsesUserMessage{Role: "assistant", Content: []any{responsesInputText{Type: "output_text", Text: b.Text}}})
				continue
			}
			if responsesCapabilitiesOf(model.Provider).implementation == responsesGLM {
				items = append(items, glmResponsesAssistantText(b.Text))
				continue
			}
			id, phase := parseTextSignature(b.Signature)
			switch {
			case id == "" && textIndex == 0:
				id = "msg_pi_" + strconv.Itoa(msgIndex)
			case id == "":
				id = "msg_pi_" + strconv.Itoa(msgIndex) + "_" + strconv.Itoa(textIndex)
			case len(id) > 64: // OpenAI's item id limit
				id = "msg_" + shortHash(id)
			}
			textIndex++
			items = append(items, responsesOutputMessage{Type: "message", Role: "assistant", Status: "completed", ID: id, Phase: phase,
				Content: []responsesOutputText{{Type: "output_text", Text: b.Text, Annotations: []any{}}}})
		case ToolCall:
			parts := strings.Split(b.ID, "|")
			callID := parts[0]
			if responsesCapabilitiesOf(model.Provider).implementation == responsesMiniMax {
				var err error
				callID, err = miniMaxResponsesCallID(b.ID)
				if err != nil {
					return nil, err
				}
			}
			if responsesCapabilitiesOf(model.Provider).implementation == responsesGLM {
				callID = glmResponsesCallID(b.ID)
			}
			itemID := ""
			if len(parts) > 1 {
				itemID = parts[1]
			}
			// A different model's fc_ id would trip OpenAI's reasoning pairing
			// check, and function_call item ids must start with fc_.
			if responsesCapabilitiesOf(model.Provider).implementation == responsesGLM || responsesCapabilitiesOf(model.Provider).implementation == responsesMiniMax || differentModel || !strings.HasPrefix(itemID, "fc_") {
				itemID = ""
			}
			// Arguments that are not JSON (only a host-built call can have
			// them) make the request unencodable: invalid_request, never a
			// silent {}.
			args, err := parseJSON(string(cmp.Or(b.Arguments, "{}")))
			if err != nil {
				return nil, err
			}
			items = append(items, responsesFunctionCall{Type: "function_call", ID: itemID, CallID: callID, Name: b.Name, Arguments: stringifyJSON(args)})
		}
	}
	return items, nil
}
