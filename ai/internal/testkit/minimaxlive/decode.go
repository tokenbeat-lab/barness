package minimaxlive

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
)

// DecodeMessage is the evidence Host's explicit storage DTO conversion. It
// never restores NativeState trust; the Host must separately verify provenance.
func DecodeMessage(raw []byte) (ai.AssistantMessage, error) {
	type alias ai.AssistantMessage
	var out ai.AssistantMessage
	wire := struct {
		*alias
		Content []json.RawMessage `json:"content"`
	}{alias: (*alias)(&out)}
	if e := json.Unmarshal(raw, &wire); e != nil {
		return out, e
	}
	for _, part := range wire.Content {
		var kind struct {
			Type string `json:"type"`
		}
		if e := json.Unmarshal(part, &kind); e != nil {
			return out, e
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(part, &fields)
		if _, ok := fields["text"]; ok {
			kind.Type = "text"
		}
		if _, ok := fields["thinking"]; ok {
			kind.Type = "thinking"
		}
		if _, ok := fields["arguments"]; ok {
			kind.Type = "toolCall"
		}
		switch kind.Type {
		case "text":
			var v ai.Text
			if e := json.Unmarshal(part, &v); e != nil {
				return out, e
			}
			out.Content = append(out.Content, v)
		case "thinking":
			var v ai.Thinking
			if e := json.Unmarshal(part, &v); e != nil {
				return out, e
			}
			out.Content = append(out.Content, v)
		case "toolCall":
			var v ai.ToolCall
			if e := json.Unmarshal(part, &v); e != nil {
				return out, e
			}
			out.Content = append(out.Content, v)
		default:
			return out, fmt.Errorf("unknown captured assistant block")
		}
	}
	return out, nil
}
func (c *Capture) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Result struct {
			Message  json.RawMessage `json:"message"`
			Metadata ai.CallMetadata `json:"metadata"`
		} `json:"result"`
		ErrorCode  ai.Code  `json:"errorCode"`
		ErrorPhase ai.Phase `json:"errorPhase"`
	}
	if e := json.Unmarshal(raw, &wire); e != nil {
		return e
	}
	msg, e := DecodeMessage(wire.Result.Message)
	if e != nil {
		return e
	}
	c.Result = ai.Result{Message: msg, Metadata: wire.Result.Metadata}
	c.ErrorCode = wire.ErrorCode
	c.ErrorPhase = wire.ErrorPhase
	return nil
}
