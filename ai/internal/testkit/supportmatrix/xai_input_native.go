package supportmatrix

import "encoding/json"

func xaiInputFields(item map[string]json.RawMessage, context string, keys ...string) error {
	for key := range item {
		allowed := false
		for _, known := range keys {
			allowed = allowed || key == known
		}
		if !allowed {
			return refuse("xAI %s contains an unsupported control or resource field", context)
		}
	}
	return nil
}
func verifyXAINativeInput(item map[string]json.RawMessage, kind string) error {
	if kind == "reasoning" {
		return verifyXAIReasoning(item)
	}
	if kind == "function_call" {
		if err := xaiInputFields(item, "native function call", "type", "id", "call_id", "name", "arguments"); err != nil {
			return err
		}
		call, ok := xaiWireText(item["call_id"])
		name, nok := xaiWireText(item["name"])
		args, aok := xaiWireText(item["arguments"])
		var object map[string]json.RawMessage
		if !ok || call == "" || !nok || name == "" || !aok || regionalDecode([]byte(args), &object) != nil || object == nil {
			return refuse("xAI native function identity or complete business arguments invalid")
		}
		if raw, present := item["id"]; present {
			if id, ok := xaiWireText(raw); !ok || id == "" {
				return refuse("xAI native function item ID invalid")
			}
		}
		return nil
	}
	if err := xaiInputFields(item, "native function output", "type", "call_id", "output"); err != nil {
		return err
	}
	id, ok := xaiWireText(item["call_id"])
	_, text := xaiWireText(item["output"])
	if !ok || id == "" || !text {
		return refuse("xAI native function output must correlate with text")
	}
	return nil
}
func verifyXAIInputPairs(input []map[string]json.RawMessage) error {
	calls, outputs := map[string]bool{}, map[string]bool{}
	for _, item := range input {
		id := xaiWireString(item["call_id"])
		switch xaiWireString(item["type"]) {
		case "function_call":
			if calls[id] {
				return refuse("xAI native input repeats a function association")
			}
			calls[id] = true
		case "function_call_output":
			if !calls[id] || outputs[id] {
				return refuse("xAI native function output is orphaned or duplicated")
			}
			outputs[id] = true
		}
	}
	for id := range calls {
		if !outputs[id] {
			return refuse("xAI native function history lacks its output")
		}
	}
	return nil
}
