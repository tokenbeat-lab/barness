//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// The scenarios below settle what issue 34 leaves to the live smoke: the
// vendors accept the requests barness sends to the models ADR-0018 lists
// without their optional features, and gpt-5.4-pro is not served on Chat
// Completions. A refusal keeps the model out of the catalog; the
// conclusion goes to ADR-0018 and the support matrix.

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
		if len(ex) > 0 && ex[len(ex)-1].RequestHeaders["Anthropic-Beta"] != "" {
			leaked = append(leaked, "Anthropic-Beta: "+ex[len(ex)-1].RequestHeaders["Anthropic-Beta"])
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

// chatProUnavailable confirms why the catalog leaves gpt-5.4-pro off Chat
// Completions (ADR-0013 决策二, ADR-0018 决策三): a client whose catalog
// lists it there sends the request, and the vendor refuses the model.
var chatProUnavailable = scenario{id: "pro-model-unavailable", model: "gpt-5.4-pro", run: func(s *session) {
	pro := slices.IndexFunc(catalog.Models, func(m ai.Model) bool { return m.API == ai.APIOpenAIResponses && m.ID == s.model })
	if !s.check("the pro model is listed on Responses", pro >= 0, "catalog %s", catalog.Version) {
		return
	}
	probe := ai.Catalog{Version: catalog.Version + "+chat-pro-probe", Models: slices.Clone(catalog.Models)}
	m := catalog.Models[pro]
	m.API = ai.APIOpenAICompletions
	probe.Models = append(probe.Models, m)
	client, err := newClient(s.env.combo, cfg.key, s.env.rec, probe)
	if !s.check("probe client assembles", err == nil, "%v", err) || !s.spend() {
		return
	}
	res, err := client.CompleteSimple(s.ctx, newScope(), s.target(), ai.Request{Messages: []ai.Message{ai.UserText(shortPrompt)}},
		ai.SimpleOptions{MaxTokens: ai.Value(256)})
	s.observe("probe", res, err)
	var e *ai.Error
	switch {
	case err == nil:
		s.check("the vendor refuses the pro model on Chat", false,
			"served (stop %q): list gpt-5.4-pro on Chat Completions under a new catalog version and record it in ADR-0018", res.Message.StopReason)
	case errors.As(err, &e) && slices.Contains(retryableCodes, e.Code):
		s.ok("probe", err)
	case errors.As(err, &e) && (e.HTTPStatus == 404 || e.HTTPStatus == 400) && strings.Contains(strings.ToLower(e.Message), "model"):
		// Only a refusal of the model confirms it; a 401, 403 or a 400 about
		// something else says nothing about Chat serving it.
		s.note("refused with HTTP %d (%s): %s", e.HTTPStatus, e.Code, truncate(e.Message, 300))
	default:
		s.check("the vendor refuses the pro model on Chat", false, "unexpected failure: %s", errText(err))
	}
}}

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
