//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// The scenarios below settle what issue 34 leaves to the live smoke: the
// vendors accept requests for the active GPT-5.6/GPT-6 catalog models without
// optional features. The retired pro-model vendor probe was removed with the
// 2026-11-03 catalog retirement; historical captures remain unchanged.

// clockTool is removed and convertTool added by the mid-conversation change.
var (
	clockTool = ai.Tool{
		Name:        "get_time",
		Description: "Returns the current local time in a city.",
		Parameters: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string","description":"City name"}},` +
			`"required":["city"],"additionalProperties":false}`),
	}
	convertTool = ai.Tool{
		Name:        "convert_temperature",
		Description: "Converts a temperature in degrees Celsius to degrees Fahrenheit.",
		Parameters: json.RawMessage(`{"type":"object","properties":{"celsius":{"type":"number"}},` +
			`"required":["celsius"],"additionalProperties":false}`),
	}
)

// optionalFeatureMarkers are what the optional features ADR-0018 leaves out
// would put on the wire: Anthropic's native tool changes and fallback
// models, OpenAI's additional_tools and tool search.
var optionalFeatureMarkers = []string{`"defer_loading"`, `"tool_addition"`, `"tool_removal"`, `"fallbacks"`,
	`"additional_tools"`, `"tool_search_`}

// toolChanges is a multi-turn tool conversation on model with a
// mid-conversation change: the model calls get_weather, a later system
// message removes get_time and adds convert_temperature, the model calls
// the added tool and answers from its result. The simple entry without a
// level sends each model's default reasoning, which its level map keeps
// valid; tool choice is automatic with directive prompts.
func toolChanges(model string) scenario {
	return scenario{id: "tool-changes-" + model, model: model, run: func(s *session) {
		if !s.spend() {
			return
		}
		opts := ai.SimpleOptions{MaxTokens: ai.Value(1024)}
		req := ai.Request{SystemPrompt: "You are a concise assistant.", Tools: []ai.Tool{weatherTool, clockTool},
			Messages: []ai.Message{ai.UserText(toolPrompt)}}
		res, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, opts)
		s.observe("call", res, err)
		if !s.ok("call", err) || !s.turn("call", res, ai.StopReasonToolUse) {
			return
		}
		call, _, ok := s.weatherCall("call", res.Message, req.Tools)
		if !ok || !s.spend() {
			return
		}
		change := ai.SystemMessage{Content: "The get_time tool is no longer available. Use convert_temperature for unit conversions.",
			ToolsRemoved: []string{clockTool.Name}, ToolsAdded: []ai.Tool{convertTool}}
		req.Messages = append(req.Messages, res.Message, ai.ToolResultText(call, weatherResult, false), change,
			ai.UserText("Convert that temperature to Fahrenheit with the convert_temperature tool."))
		res, err = s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, opts)
		ex := s.observe("changed", res, err)
		body := requestBody(ex)
		declared := declaredTools(body)
		s.check("changed: the current tools are declared", slices.Equal(declared, []string{weatherTool.Name, convertTool.Name}),
			"declared %q", declared)
		s.check("changed: the system message is sent", sentStrings(ex)[change.Content], "request %s", truncate(fmt.Sprint(body), 600))
		sent := lastRequestBody(ex)
		leaked := slices.DeleteFunc(slices.Clone(optionalFeatureMarkers), func(m string) bool { return !strings.Contains(sent, m) })
		if len(ex) > 0 && ex[len(ex)-1].RequestHeaders["anthropic-beta"] != "" {
			leaked = append(leaked, "Anthropic-Beta: "+ex[len(ex)-1].RequestHeaders["anthropic-beta"])
		}
		s.check("changed: no optional feature sent", len(leaked) == 0, "sent %q", leaked)
		if !s.ok("changed", err) {
			return
		}
		convert, idx := toolCallNamed(res.Message, convertTool.Name)
		if idx < 0 {
			// Accepting the request is what the scenario proves; whether the
			// model calls the added tool is its own choice.
			s.turn("changed", res, ai.StopReasonStop)
			s.note("answered without calling %s", convertTool.Name)
			return
		}
		_, err = ai.ValidateToolCall(res.Message, idx, []ai.Tool{weatherTool, convertTool})
		if !s.turn("changed", res, ai.StopReasonToolUse) || !s.check("changed: the added tool's call validates", err == nil, "%v", err) ||
			!s.spend() {
			return
		}
		req.Messages = append(req.Messages, res.Message, ai.ToolResultText(convert, `{"fahrenheit":64.4}`, false))
		opts.ToolChoice = ai.Value(ai.ToolChoiceNone)
		res, err = s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, opts)
		s.observe("answer", res, err)
		if s.ok("answer", err) && s.turn("answer", res, ai.StopReasonStop) {
			s.note("called %s after the change and answered", convertTool.Name)
		}
	}}
}

// declaredTools are the tool names of a request body: Responses and
// Anthropic name each tool at its top, Chat under function.
func declaredTools(body map[string]any) []string {
	tools, _ := body["tools"].([]any)
	var names []string
	for _, t := range tools {
		tool, _ := t.(map[string]any)
		name, _ := tool["name"].(string)
		if fn, ok := tool["function"].(map[string]any); ok {
			name, _ = fn["name"].(string)
		}
		names = append(names, name)
	}
	return names
}

// toolCallNamed is the message's first call of name and its index, or -1.
func toolCallNamed(m ai.AssistantMessage, name string) (ai.ToolCall, int) {
	for i, c := range m.Content {
		if call, ok := c.(ai.ToolCall); ok && call.Name == name {
			return call, i
		}
	}
	return ai.ToolCall{}, -1
}
