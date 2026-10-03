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

// The scenarios below settle what issues 21 and 22 left to the live smoke:
// where Chat usage arrives, DeepSeek's answers to thinking with a forced
// tool, to the prompt cache fields and to the high/max levels, and its error
// body and request id header. Each records what it saw in its note.

// chatUsagePosition: usage arrives before [DONE], in the finish_reason chunk
// or a separate chunk (ADR-0015 Consequences).
var chatUsagePosition = scenario{id: "usage-position", run: func(s *session) {
	if !s.spend() {
		return
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText(shortPrompt)}}
	d := drain(s.env.client.Stream(s.ctx, newScope(), s.target(), req, s.env.combo.full(256, toolsDefault)), nil)
	ex := s.observe("stream", d.res, d.err)
	if !s.ok("stream", d.err) || !s.turn("stream", d.res, ai.StopReasonStop) || len(ex) == 0 {
		return
	}
	pos, problem := usagePosition(ex[len(ex)-1].ResponseBody)
	if s.check("usage before [DONE]", problem == "", "%s", problem) {
		s.note("usage %s", pos)
	}
}}

// chatChoice is the part of a Chat chunk's choice usagePosition reads.
type chatChoice struct {
	FinishReason *string `json:"finish_reason"`
}

// usagePosition reads a Chat SSE body: where the first chunk with usage
// stands relative to the finish_reason chunk and [DONE].
func usagePosition(body string) (string, string) {
	usageAt, finishAt, doneAt := -1, -1, -1
	n := 0
	for line := range strings.SplitSeq(body, "\n") {
		data, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "[DONE]" {
			doneAt = n
			break
		}
		var chunk struct {
			Usage   json.RawMessage `json:"usage"`
			Choices []chatChoice    `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) == nil {
			if usageAt < 0 && len(chunk.Usage) > 0 && string(chunk.Usage) != "null" {
				usageAt = n
			}
			if finishAt < 0 && slices.ContainsFunc(chunk.Choices, func(c chatChoice) bool { return c.FinishReason != nil }) {
				finishAt = n
			}
		}
		n++
	}
	switch {
	case usageAt < 0:
		return "", "no chunk carried usage"
	case doneAt < 0:
		return "", "no [DONE] captured"
	case usageAt == finishAt:
		return "in the finish_reason chunk", ""
	case usageAt > finishAt:
		return fmt.Sprintf("in a separate chunk %d after the finish_reason chunk", usageAt-finishAt), ""
	}
	return fmt.Sprintf("in chunk %d, before the finish_reason chunk %d", usageAt, finishAt), ""
}

// forcedToolWithThinking: thinking on with tool_choice required (ADR-0015
// 决策三). A refusal or an honored call keeps the adapter as it is; a
// silently ignored choice fails, since the adapter must then refuse the
// combination before sending.
var forcedToolWithThinking = scenario{id: "forced-tool-with-thinking", run: func(s *session) {
	if !s.spend() {
		return
	}
	opts := ai.ChatOptions{MaxTokens: ai.Value(2048), ReasoningEffort: ai.ThinkingLow,
		ToolChoice: ai.Value(ai.ChatToolChoice{Mode: "required"})}
	req := ai.Request{Messages: []ai.Message{ai.UserText(toolPrompt)}, Tools: []ai.Tool{weatherTool}}
	res, err := s.env.client.Complete(s.ctx, newScope(), s.target(), req, opts)
	ex := s.observe("forced", res, err)
	body := requestBody(ex)
	s.check("sent thinking enabled with tool_choice required", jsonAt(body, "thinking", "type") == "enabled" &&
		jsonAt(body, "tool_choice") == "required", "request %s", truncate(fmt.Sprint(body), 400))
	var e *ai.Error
	switch {
	case errors.As(err, &e) && e.Code == ai.CodeInvalidRequest:
		s.note("refused with HTTP %d as ADR-0015 决策三 expects: %s", e.HTTPStatus, truncate(e.Message, 300))
	case err != nil:
		s.ok("forced", err)
	case res.Message.StopReason == ai.StopReasonToolUse:
		s.note("honored: the model called the tool with thinking on; ADR-0015 决策三 needs no change")
	default:
		s.check("forced choice not silently ignored", false,
			"stop reason %q without a tool call: per ADR-0015 决策三 the Chat adapter must refuse thinking + a forced choice before sending",
			res.Message.StopReason)
	}
}}

// promptCacheLongRetention: the derived prompt_cache_key and 24h retention
// pi sends with long retention are accepted (ADR-0015 决策四).
var promptCacheLongRetention = scenario{id: "prompt-cache-long-retention", run: func(s *session) {
	if !s.spend() {
		return
	}
	opts := ai.ChatOptions{MaxTokens: ai.Value(256), CacheRetention: ai.CacheRetentionLong, SessionID: "live-smoke-session"}
	res, err := s.env.client.Complete(s.ctx, newScope(), s.target(), ai.Request{Messages: []ai.Message{ai.UserText(shortPrompt)}}, opts)
	ex := s.observe("cache", res, err)
	body := requestBody(ex)
	s.check("sent prompt_cache_key and 24h retention", jsonAt(body, "prompt_cache_key") != "" &&
		jsonAt(body, "prompt_cache_retention") == "24h", "request %s", truncate(fmt.Sprint(body), 400))
	var e *ai.Error
	if errors.As(err, &e) && e.Code == ai.CodeInvalidRequest {
		s.check("prompt cache fields accepted", false,
			"refused (%s): per ADR-0015 决策四 stop sending them to DeepSeek and register the difference in the ledger", truncate(e.Message, 300))
		return
	}
	if s.ok("cache", err) && s.turn("cache", res, ai.StopReasonStop) {
		s.note("accepted; cache read %d tokens", res.Message.Usage.CacheRead)
	}
}}

// reasoningLevel: the service accepts the model's level (ADR-0014 决策三,
// ADR-0015 决策二). A refused level fails; the catalog then maps it to null
// under a new version.
func reasoningLevel(model string, level ai.ThinkingLevel) scenario {
	return scenario{id: "reasoning-level-" + string(level), model: model, run: func(s *session) {
		m, _ := s.modelInfo()
		wire, mapped := levelValue(m.ThinkingLevelMap, level)
		if !mapped {
			s.unsupported = fmt.Sprintf("the catalog maps %s's %s level to null", s.model, level)
			return
		}
		if !s.spend() {
			return
		}
		opts := ai.SimpleOptions{Reasoning: level, MaxTokens: ai.Value(maxOutputTokens)}
		req := ai.Request{Messages: []ai.Message{ai.UserText("What is 17 × 23? Reply with the number only.")}}
		res, err := s.env.client.CompleteSimple(s.ctx, newScope(), s.target(), req, opts)
		ex := s.observe(string(level), res, err)
		body := requestBody(ex)
		sent := jsonAt(body, "reasoning", "effort") // Responses
		if s.env.combo.api == ai.APIOpenAICompletions {
			sent = jsonAt(body, "reasoning_effort")
		}
		s.check("level sent", sent == wire, "sent %q, catalog maps %s to %q", sent, level, wire)
		var e *ai.Error
		if errors.As(err, &e) && e.Code == ai.CodeInvalidRequest {
			s.check("level accepted", false, "refused (%s): map %s to null for %s under a new catalog version", truncate(e.Message, 300), level, s.model)
			return
		}
		if !s.ok(string(level), err) {
			return
		}
		r := res.Message
		s.check("ended normally", r.StopReason == ai.StopReasonStop || r.StopReason == ai.StopReasonLength, "stop reason %q", r.StopReason)
		reasoning, _ := r.Usage.Reasoning.Get()
		s.check("usage reported", r.Usage.Output > 0, "usage %+v", r.Usage)
		s.note("accepted; stop %s, %d output tokens of which %d reasoning", r.StopReason, r.Usage.Output, reasoning)
	}}
}

func levelValue(m ai.ThinkingLevelMap, l ai.ThinkingLevel) (string, bool) {
	var entry ai.Nullable[string]
	switch l {
	case ai.ThinkingHigh:
		entry = m.High
	case ai.ThinkingMax:
		entry = m.Max
	default:
		return "", false
	}
	if entry.IsNull() {
		return "", false
	}
	if v, ok := entry.Get(); ok {
		return v, true
	}
	return string(l), true
}

// requestBody is the last captured request body as JSON.
func requestBody(ex []exchange) map[string]any {
	var v map[string]any
	_ = json.Unmarshal([]byte(lastRequestBody(ex)), &v)
	return v
}

// lastRequestBody is the last captured request body, or "".
func lastRequestBody(ex []exchange) string {
	if len(ex) == 0 {
		return ""
	}
	return ex[len(ex)-1].RequestBody
}

// jsonAt is the string at path in v, or "".
func jsonAt(v map[string]any, path ...string) string {
	var cur any = v
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[p]
	}
	s, _ := cur.(string)
	return s
}
