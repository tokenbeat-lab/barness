package supportmatrix

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyRegionalToolResult(root *os.Root, r Report, s ScenarioResult, wire string, message ai.AssistantMessage) error {
	raw, err := regionalRead(root, RegionalCaseID(r.Combo, s.Model+"/tool-call")+"/result.json")
	var source regionalResult
	if err != nil || json.Unmarshal(raw, &source) != nil {
		return refuse("regional own tool source absent")
	}
	first, err := source.Message.assistant()
	if err != nil {
		return err
	}
	want := map[string]int{}
	for _, b := range first.Content {
		if c, ok := b.(ai.ToolCall); ok {
			id := c.ID
			if r.API == "openai-responses" {
				// Decode the persisted public identity independently at the evidence
				// boundary; these are protocol data, not trusted domain state.
				if r.Provider == "minimax-cn" || RegionalIsKimi(r.Combo) {
					prefix := "mmr1|"
					if RegionalIsKimi(r.Combo) {
						prefix = "kmr1|"
					}
					encoded, ok := strings.CutPrefix(id, prefix)
					raw, e := base64.RawURLEncoding.DecodeString(encoded)
					var pair []string
					if !ok || e != nil || json.Unmarshal(raw, &pair) != nil || len(pair) != 2 {
						return refuse("regional persisted tool identity invalid")
					}
					id = pair[0]
				} else {
					id, _, _ = strings.Cut(id, "|")
					id = strings.NewReplacer("%7C", "|", "%25", "%").Replace(id)
				}
			}
			want[id] = 0
		}
	}
	if len(want) == 0 {
		return refuse("regional own tool source has no calls")
	}
	var body any
	if json.Unmarshal([]byte(wire), &body) != nil {
		return refuse("regional tool result wire invalid")
	}
	valid := true
	outputs := 0
	var walk func(any)
	walk = func(v any) {
		switch v := v.(type) {
		case []any:
			for _, child := range v {
				walk(child)
			}
		case map[string]any:
			id := ""
			var content any
			output := false
			if v["role"] == "tool" {
				id, _ = v["tool_call_id"].(string)
				content = v["content"]
				output = true
			}
			if v["type"] == "function_call_output" {
				id, _ = v["call_id"].(string)
				content = v["output"]
				output = true
			}
			if v["type"] == "tool_result" {
				id, _ = v["tool_use_id"].(string)
				content = v["content"]
				output = true
			}
			if output {
				outputs++
				_, ok := want[id]
				text, _ := content.(string)
				if text == "" {
					text = regionalPartText(content)
				}
				if !ok || text != "REGIONAL42" {
					valid = false
				}
				want[id]++
			}
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(body)
	for _, count := range want {
		valid = valid && count == 1
	}
	text := ""
	for _, b := range message.Content {
		if b, ok := b.(ai.Text); ok {
			text += b.Text
		}
	}
	if !valid || outputs != len(want) || !strings.Contains(text, "REGIONAL42") {
		return refuse("regional Host tool result pair or visible answer differs")
	}
	return nil
}
