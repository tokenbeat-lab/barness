package ai

import (
	"encoding/json"
)

// Decode the shared event shape without coercing Kimi counters through int64
// first. The raw usage boundary checks exact integers (including exponent
// syntax), missing/null presence, containment and addition before publishing.
func decodeKimiMessagesEvent(raw string, e *anthropicEvent) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return err
	}
	delete(object, "usage")
	if message, present := object["message"]; present && string(message) != "null" {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(message, &fields); err != nil {
			return err
		}
		delete(fields, "usage")
		encoded, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		object["message"] = encoded
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, e)
}

// Messages deltas replace supplied counters and preserve omitted ones. Null
// replaces a previous value with missing, and supplied TTL/detail containers
// replace their previous breakdown so stale certainty never survives.
func (p *anthropicParser) kimiUsageEvent(ev sseEvent) *Error {
	var object map[string]json.RawMessage
	_ = json.Unmarshal([]byte(ev.data), &object)
	raw := object["usage"]
	if ev.name == "message_start" {
		var message map[string]json.RawMessage
		_ = json.Unmarshal(object["message"], &message)
		raw = message["usage"]
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var patch map[string]json.RawMessage
	if json.Unmarshal(raw, &patch) != nil || patch == nil {
		return kimiUsageFailure()
	}
	merged := map[string]json.RawMessage{}
	for key, value := range p.kimiCounts {
		merged[key] = value
	}
	for key, value := range patch {
		merged[key] = value
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return kimiUsageFailure()
	}
	u, reporting, estimate, failure := kimiUsage(p.model, encoded, nil)
	if failure != nil {
		return failure
	}
	p.kimiCounts = merged
	p.out.usage(u)
	p.attempt.usageEstimate(reporting, u, estimate)
	return nil
}
