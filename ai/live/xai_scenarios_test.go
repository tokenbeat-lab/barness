//go:build live

package live

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"regexp"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func xaiExpected() []string {
	return []string{"entry-stream", "entry-complete", "entry-stream-simple", "entry-complete-simple", "effort-default", "effort-low", "effort-medium", "effort-high", "effort-xhigh", "encrypted-continuation", "tools-auto", "tools-required", "tools-none", "tools-parallel", "user-png", "user-jpeg", "tool-result-image", "cache-usage"}
}
func xaiScenarios() []xaiScenario {
	all := []xaiScenario{}
	for _, entry := range []string{"Stream", "Complete", "StreamSimple", "CompleteSimple"} {
		entry := entry
		id := map[string]string{"Stream": "entry-stream", "Complete": "entry-complete", "StreamSimple": "entry-stream-simple", "CompleteSimple": "entry-complete-simple"}[entry]
		all = append(all, xaiScenario{id: id, run: func(s *xaiSession) { s.text(entry, ai.ThinkingLow) }})
	}
	for _, level := range []ai.ThinkingLevel{"", ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh} {
		level := level
		id := "effort-" + string(level)
		if level == "" {
			id = "effort-default"
		}
		all = append(all, xaiScenario{id: id, run: func(s *xaiSession) {
			s.text("Complete", level)
			if level == ai.ThinkingHigh {
				s.nonzeroReasoning()
			}
		}})
	}
	all = append(all, xaiScenario{"encrypted-continuation", xaiEncryptedContinuation})
	for _, mode := range []string{"auto", "required", "none", "parallel"} {
		mode := mode
		all = append(all, xaiScenario{"tools-" + mode, func(s *xaiSession) { s.tools(mode, false) }})
	}
	for _, mime := range []string{"image/png", "image/jpeg"} {
		mime := mime
		id := "user-png"
		if mime == "image/jpeg" {
			id = "user-jpeg"
		}
		all = append(all, xaiScenario{id, func(s *xaiSession) {
			image := xaiImage(mime)
			s.call("image", "Complete", ai.Request{Messages: []ai.Message{ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: "Name the main color in one word."}, image}}}}, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingLow})
			s.cs.Check("complete inline image sent", len(s.last) == 1 && strings.Contains(s.last[0].RequestBody, "data:"+mime+";base64,"+image.Data), "image bytes missing")
		}})
	}
	all = append(all, xaiScenario{"tool-result-image", func(s *xaiSession) { s.tools("required", true) }}, xaiScenario{"cache-usage", xaiCacheUsage})
	return all
}
func (s *xaiSession) text(entry string, level ai.ThinkingLevel) {
	res := s.call("text", entry, ai.Request{SystemPrompt: "Be concise.", Messages: []ai.Message{ai.UserText("Compute 17 * 23 and briefly state the answer.")}}, ai.ResponsesOptions{ReasoningEffort: level})
	if s.t.Failed() {
		return
	}
	s.cs.Check("text entry returns visible text", strings.TrimSpace(textOf(res.Message)) != "", "no visible answer")
	body := xaiRequest(s.last)
	reasoning, _ := body["reasoning"].(map[string]any)
	effort, _ := reasoning["effort"].(string)
	want := string(level)
	if entry == "StreamSimple" || entry == "CompleteSimple" {
		want = "low"
	}
	s.cs.Check("exact effort preserves omission", effort == want, "effort=%q expected=%q", effort, want)
}
func xaiEncryptedContinuation(s *xaiSession) {
	user := ai.UserText("Remember the product of 17 and 23, and answer briefly.")
	first := s.call("first", "Complete", ai.Request{Messages: []ai.Message{user}}, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingHigh})
	if s.t.Failed() {
		return
	}
	s.nonzeroReasoning()
	native := xaiReasoningItems(s.last)
	s.cs.Check("native encrypted item required", len(native) > 0, "encrypted state missing")
	if s.t.Failed() {
		return
	}
	restored := s.restore(first.Message)
	if s.t.Failed() {
		return
	}
	second := s.call("continued", "Stream", ai.Request{Messages: []ai.Message{user, restored, ai.UserText("Add one to the previous product.")}}, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingLow})
	if s.t.Failed() {
		return
	}
	s.cs.Check("restored native state has no downgrade", second.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "native state downgraded")
	input, _ := xaiRequest(s.last)["input"].([]any)
	for _, item := range native {
		found := false
		for _, in := range input {
			if xaiJSONEqual(item, in) {
				found = true
			}
		}
		s.cs.Check("entire native reasoning item replayed", found, "native item changed")
	}
}
func (s *xaiSession) tools(mode string, imageResult bool) {
	choice := mode
	if mode == "parallel" {
		choice = "required"
	}
	prompt := toolPrompt
	if mode == "parallel" {
		prompt = "Get the weather in both Paris and Tokyo now. Make two separate get_weather calls in parallel, one per city."
	}
	if imageResult {
		prompt = "Use get_weather for Paris; the tool will return a picture. Then describe the picture."
	}
	if mode == "none" {
		prompt = "Say hello without using tools."
	}
	o := ai.ResponsesOptions{ReasoningEffort: ai.ThinkingLow, ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: choice})}
	if mode == "parallel" {
		o.SamplingParams = map[string]json.RawMessage{"parallel_tool_calls": json.RawMessage("true")}
	}
	user := ai.UserText(prompt)
	tools := []ai.Tool{weatherTool}
	req := ai.Request{Messages: []ai.Message{user}, Tools: tools}
	first := s.call("calls", "Stream", req, o)
	if s.t.Failed() {
		return
	}
	s.cs.Check("exact supported tool mode", xaiRequest(s.last)["tool_choice"] == choice, "tool choice changed")
	if mode == "parallel" {
		s.cs.Check("parallel flag sent", xaiRequest(s.last)["parallel_tool_calls"] == true, "parallel flag missing")
	}
	var calls []ai.ToolCall
	for i, c := range first.Message.Content {
		if call, ok := c.(ai.ToolCall); ok {
			_, err := ai.ValidateToolCall(first.Message, i, tools)
			s.cs.Check("Host validates complete function arguments", err == nil, "unexecutable tool call")
			calls = append(calls, call)
		}
	}
	if mode == "none" {
		s.cs.Check("none has zero calls", len(calls) == 0 && first.Message.StopReason == ai.StopReasonStop, "none returned tool")
		return
	}
	want := 1
	if mode == "parallel" {
		want = 2
	}
	if !s.cs.Check("expected independent calls", len(calls) == want && first.Message.StopReason == ai.StopReasonToolUse, "calls=%d expected=%d", len(calls), want) {
		return
	}
	original := xaiToolItems(s.last)
	if !s.cs.Check("raw calls available for correlation", len(original) == want, "missing call capture") {
		return
	}
	req.Messages = append(req.Messages, first.Message)
	// Reverse parallel results to exercise correlation independently of order.
	for i := len(calls) - 1; i >= 0; i-- {
		result := ai.ToolResultText(calls[i], "synthetic-host-result-"+string(rune('a'+i)), false)
		if imageResult {
			result.Content = append(result.Content, xaiImage("image/png"), xaiImage("image/jpeg"))
		}
		req.Messages = append(req.Messages, result)
	}
	o.ToolChoice = ai.Value(ai.ResponsesToolChoice{Mode: "none"})
	final := s.call("results", "Complete", req, o)
	if s.t.Failed() {
		return
	}
	s.cs.Check("tool result concludes", final.Message.StopReason == ai.StopReasonStop, "tool result did not conclude")
	input, _ := xaiRequest(s.last)["input"].([]any)
	for i, item := range original {
		callID, _ := item["call_id"].(string)
		itemID, _ := item["id"].(string)
		functions, outputs := 0, 0
		for _, in := range input {
			m, _ := in.(map[string]any)
			if m["call_id"] == callID {
				if m["type"] == "function_call" && m["id"] == itemID {
					functions++
				}
				output, _ := m["output"].(string)
				marker := "synthetic-host-result-" + string(rune('a'+i))
				if m["type"] == "function_call_output" && (output == marker || imageResult && strings.HasPrefix(output, marker+"\n")) {
					outputs++
				}
			}
		}
		s.cs.Check("native pair preserved once", functions == 1 && outputs == 1, "pair mismatch")
	}
	if imageResult {
		outputIndex := -1
		var images []string
		beforePair := false
		for i, entry := range input {
			item, _ := entry.(map[string]any)
			if item["type"] == "function_call_output" && item["call_id"] == original[0]["call_id"] {
				outputIndex = i
			}
			content, _ := item["content"].([]any)
			for _, value := range content {
				block, _ := value.(map[string]any)
				if block["type"] == "input_image" {
					url, _ := block["image_url"].(string)
					images = append(images, url)
					beforePair = beforePair || outputIndex < 0 || i <= outputIndex
				}
			}
		}
		expected := []string{"data:image/png;base64," + xaiImage("image/png").Data, "data:image/jpeg;base64," + xaiImage("image/jpeg").Data}
		s.cs.Check("paired tool images retain PNG then JPEG order", outputIndex >= 0 && !beforePair && xaiJSONEqual(images, expected), "tool image order or pairing changed")
	}
}
func xaiCacheUsage(s *xaiSession) {
	req := ai.Request{Messages: []ai.Message{ai.UserText("Say hello briefly.")}}
	o := ai.ResponsesOptions{ReasoningEffort: ai.ThinkingLow, SessionID: "xai-synthetic-cache-session", CacheRetention: ai.CacheRetentionShort}
	key := ""
	for _, label := range []string{"first", "repeat"} {
		s.call(label, "Complete", req, o)
		if s.t.Failed() {
			return
		}
		body := xaiRequest(s.last)
		current, _ := body["prompt_cache_key"].(string)
		shape, _ := regexp.MatchString(`^[0-9a-f]{64}$`, current)
		s.cs.Check("cache key is derived and stable", shape && current != o.SessionID && (key == "" || current == key), "invalid cache scope")
		key = current
	}
	// A reported zero is valid; two bounded calls never promise a cache hit.
}
func xaiImage(mime string) ai.Image {
	if mime == "image/png" {
		return ai.Image{Data: redSquarePNG(), MimeType: mime}
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: 220, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return ai.Image{Data: base64.StdEncoding.EncodeToString(buf.Bytes()), MimeType: mime}
}
