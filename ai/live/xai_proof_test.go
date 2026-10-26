//go:build live

package live

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func xaiRequest(ex []exchange) map[string]any {
	var body map[string]any
	if len(ex) == 1 {
		_ = json.Unmarshal([]byte(ex[0].RequestBody), &body)
	}
	return body
}
func xaiTerminal(ex []exchange) map[string]any {
	if len(ex) != 1 {
		return nil
	}
	var lines []string
	decode := func() map[string]any {
		var event map[string]any
		if json.Unmarshal([]byte(strings.Join(lines, "\n")), &event) == nil && event["type"] == "response.completed" {
			r, _ := event["response"].(map[string]any)
			return r
		}
		return nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(ex[0].ResponseBody, "\r\n", "\n"), "\n") {
		if line == "" {
			if response := decode(); response != nil {
				return response
			}
			lines = nil
			continue
		}
		if data, ok := strings.CutPrefix(line, "data:"); ok {
			lines = append(lines, strings.TrimPrefix(data, " "))
		}
	}
	return decode()
}
func (s *xaiSession) streamProof(label string, d drained) {
	if !s.cs.Check("stream lifecycle "+label, len(d.events) > 1, "too few events") {
		return
	}
	_, started := d.events[0].(ai.StartEvent)
	_, done := d.events[len(d.events)-1].(ai.DoneEvent)
	_, failed := d.events[len(d.events)-1].(ai.ErrorEvent)
	terminals := 0
	var terminal ai.AssistantMessage
	for _, e := range d.events {
		switch v := e.(type) {
		case ai.DoneEvent:
			terminals++
			terminal = v.Message
		case ai.ErrorEvent:
			terminals++
			terminal = v.Message
		}
	}
	s.cs.Check("one terminal and matching Result "+label, started && (done || failed) && terminals == 1 && xaiJSONEqual(terminal, d.res.Message) && errText(d.err) == errText(d.streamErr) && blockProblem(d.events) == "", "stream lifecycle mismatch")
}

func xaiJSONEqual(a, b any) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}
func xaiToolItems(ex []exchange) []map[string]any {
	var out []map[string]any
	items, _ := xaiTerminal(ex)["output"].([]any)
	for _, v := range items {
		item, _ := v.(map[string]any)
		if item["type"] == "function_call" {
			out = append(out, item)
		}
	}
	return out
}
func xaiReasoningItems(ex []exchange) []map[string]any {
	var out []map[string]any
	items, _ := xaiTerminal(ex)["output"].([]any)
	for _, v := range items {
		item, _ := v.(map[string]any)
		encrypted, _ := item["encrypted_content"].(string)
		if item["type"] == "reasoning" && encrypted != "" {
			out = append(out, item)
		}
	}
	return out
}
func (s *xaiSession) nonzeroReasoning() {
	usage, _ := xaiTerminal(s.last)["usage"].(map[string]any)
	details, _ := usage["output_tokens_details"].(map[string]any)
	r, _ := details["reasoning_tokens"].(float64)
	s.cs.Check("own nonzero reasoning arithmetic", r > 0, "nonzero reasoning required; no retries to manufacture it")
}

func (s *xaiSession) wireProof(res ai.Result) {
	body := xaiRequest(s.last)
	include, _ := body["include"].([]any)
	hasEncrypted := false
	for _, v := range include {
		hasEncrypted = hasEncrypted || v == "reasoning.encrypted_content"
	}
	s.cs.Check("fixed stateless request", body["model"] == "grok-4.7" && body["store"] == false && body["stream"] == true && hasEncrypted && body["previous_response_id"] == nil && body["conversation"] == nil, "stateless request contract missing")
	response := xaiTerminal(s.last)
	actual, _ := response["model"].(string)
	s.cs.Check("actual response model retained", actual != "" && (res.Message.ResponseModel == actual || actual == res.Message.Model && res.Message.ResponseModel == ""), "response model missing or changed")
	usage, _ := response["usage"].(map[string]any)
	details, _ := usage["output_tokens_details"].(map[string]any)
	inputDetails, _ := usage["input_tokens_details"].(map[string]any)
	p, pok := usage["input_tokens"].(float64)
	o, ook := usage["output_tokens"].(float64)
	r, rok := details["reasoning_tokens"].(float64)
	total, tok := usage["total_tokens"].(float64)
	cache, cok := inputDetails["cached_tokens"].(float64)
	b := total - p
	u := res.Message.Usage
	reasoning, reported := u.Reasoning.Get()
	s.cs.Check("raw usage resolves without double counting", pok && ook && rok && tok && cok && r <= b && (b == o || b == o+r) && int64(p-cache) == u.Input && int64(cache) == u.CacheRead && int64(b) == u.Output && reported && int64(r) == reasoning && int64(total) == u.TotalTokens, "wire/native arithmetic differs")
}

func (s *xaiSession) restore(m ai.AssistantMessage) ai.AssistantMessage {
	// Host-owned persistence uses typed block tags; JSON itself never restores
	// trust. This mirrors the public contract without a new serialization API.
	type block struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	}
	type record struct {
		Metadata ai.AssistantMessage    `json:"metadata"`
		Envelope ai.NativeStateEnvelope `json:"envelope"`
		Blocks   []block                `json:"blocks"`
	}
	env, ok := m.NativeState.Envelope()
	if !s.cs.Check("native provenance exists", ok, "missing provenance") {
		return ai.AssistantMessage{}
	}
	stored := record{Metadata: m, Envelope: env}
	stored.Metadata.Content = nil
	for _, c := range m.Content {
		kind := ""
		switch c.(type) {
		case ai.Thinking:
			kind = "thinking"
		case ai.Text:
			kind = "text"
		case ai.ToolCall:
			kind = "tool"
		}
		raw, _ := json.Marshal(c)
		stored.Blocks = append(stored.Blocks, block{kind, raw})
	}
	raw, err := json.Marshal(stored)
	s.cs.Record("persisted-turn", json.RawMessage(raw))
	if !s.cs.Check("persist JSON", err == nil, "persistence failed") {
		return ai.AssistantMessage{}
	}
	var decoded record
	err = json.Unmarshal(raw, &decoded)
	if !s.cs.Check("restore JSON", err == nil, "restore failed") {
		return ai.AssistantMessage{}
	}
	restored := decoded.Metadata
	for _, b := range decoded.Blocks {
		var c ai.AssistantContent
		switch b.Kind {
		case "thinking":
			var v ai.Thinking
			err = json.Unmarshal(b.Data, &v)
			c = v
		case "text":
			var v ai.Text
			err = json.Unmarshal(b.Data, &v)
			c = v
		case "tool":
			var v ai.ToolCall
			err = json.Unmarshal(b.Data, &v)
			c = v
		}
		if !s.cs.Check("typed block restored", err == nil && c != nil, "unknown or invalid block") {
			return ai.AssistantMessage{}
		}
		restored.Content = append(restored.Content, c)
	}
	_, trusted := restored.NativeState.Envelope()
	s.cs.Check("JSON does not restore trust", !trusted, "untrusted persistence acquired provenance")
	restored.NativeState, err = ai.TrustNativeState(newScope(), decoded.Envelope)
	s.cs.Check("Host explicitly restores trust", err == nil && reflect.DeepEqual(restored.Content, m.Content), "restored native content changed")
	return restored
}
