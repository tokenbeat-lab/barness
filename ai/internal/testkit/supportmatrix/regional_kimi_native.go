package supportmatrix

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// L05 is one exact L01 prefix plus one continuation. It is independent of
// tools and cannot append a different assistant whose reasoning would appear
// to justify an otherwise missing native source.
func verifyKimiRegionalNative(root *os.Root, r Report, s ScenarioResult, source string, first ai.AssistantMessage, body map[string]any) error {
	var before, after struct {
		Messages []json.RawMessage
		Tools    []json.RawMessage
	}
	raw, err := regionalRead(root, source+"/request.json")
	if err != nil || regionalDecode(raw, &before) != nil {
		return refuse("Kimi native source request missing")
	}
	raw, err = regionalRead(root, s.CaseID+"/request.json")
	if err != nil || regionalDecode(raw, &after) != nil || len(before.Messages) != 1 || len(after.Messages) != 3 || len(after.Tools) != 0 {
		return refuse("Kimi native continuation differs from fixed L01 prefix")
	}
	var last struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if regionalDecode(after.Messages[2], &last) != nil || len(last.Content) != 1 || last.Content[0].Text != "Continue that exact conversation with a short answer." {
		return refuse("Kimi native continuation is not the declared user turn")
	}
	var xs []regionalExchange
	raw, err = regionalRead(root, source+"/exchanges.json")
	if err != nil || regionalDecode(raw, &xs) != nil || len(xs) != 1 {
		return refuse("Kimi native source wire missing")
	}
	var old map[string]any
	if regionalDecode([]byte(xs[0].RequestBody), &old) != nil {
		return refuse("Kimi native source wire invalid")
	}
	field := "messages"
	if r.API == "openai-responses" {
		field = "input"
	}
	prior, current := regionalArray(old[field]), regionalArray(body[field])
	if len(prior) != 1 || len(current) < 3 || !reflect.DeepEqual(prior[0], current[0]) {
		return refuse("Kimi actual wire does not extend the source user prefix")
	}
	tail, _ := current[len(current)-1].(map[string]any)
	text, _ := tail["content"].(string)
	if text == "" {
		text = regionalPartText(tail["content"])
	}
	if tail["role"] != "user" || text != last.Content[0].Text {
		return refuse("Kimi actual wire continuation differs")
	}
	if r.API != "openai-responses" {
		if len(current) != 3 {
			return refuse("Kimi wire contains extra native turns")
		}
		assistant, _ := current[1].(map[string]any)
		if assistant["role"] != "assistant" {
			return refuse("Kimi source assistant is missing")
		}
		if r.API == "openai-completions" {
			var visible strings.Builder
			for _, b := range first.Content {
				if t, ok := b.(ai.Text); ok {
					visible.WriteString(t.Text)
				}
			}
			if assistant["content"] != visible.String() {
				return refuse("Kimi source visible assistant differs")
			}
		}
		return nil
	}
	// Kimi input accepts only the source's summary projection. Output IDs,
	// status, encrypted placeholders and content fields are not input fields.
	var expected []any
	for _, block := range first.Content {
		switch b := block.(type) {
		case ai.Thinking:
			var native map[string]any
			if json.Unmarshal([]byte(b.Signature), &native) != nil || native["type"] != "reasoning" {
				return refuse("Kimi source native reasoning is missing")
			}
			expected = append(expected, map[string]any{"type": "reasoning", "summary": native["summary"]})
		case ai.Text:
			expected = append(expected, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": b.Text}}})
		default:
			return refuse("Kimi L01 source contains an unrequested tool block")
		}
	}
	if !reflect.DeepEqual(current[1:len(current)-1], expected) {
		return refuse("Kimi native input summary projection differs from L01 output")
	}
	return nil
}
