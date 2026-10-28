//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"strings"
)

func miniMaxOptions(p minimaxlive.Case) ai.Options {
	p.ID = strings.TrimPrefix(p.ID, "research-")
	id := minimaxReport.Identity
	level := ai.ThinkingLevel("")
	if p.ID == "thinking-on" || p.ID == "tool-call" || p.ID == "tool-result-native" {
		level = ai.ThinkingHigh
	}
	if p.ID == "thinking-control" {
		level = ai.ThinkingLevel("")
		if id.Model == minimaxlive.M31 {
			level = ai.ThinkingMax
		}
	}
	n := p.Output
	if p.ID == "preflight-error" {
		n = 131073
	}
	switch id.API {
	case ai.APIOpenAICompletions:
		o := ai.ChatOptions{MaxTokens: ai.Value(n), ReasoningEffort: level}
		if p.ID == "thinking-control" && id.Model == minimaxlive.M3 {
			o.SamplingParams = map[string]json.RawMessage{"thinking": json.RawMessage(`{"type":"disabled"}`)}
		}
		return o
	case ai.APIOpenAIResponses:
		o := ai.ResponsesOptions{MaxTokens: ai.Value(n), ReasoningEffort: level}
		if strings.HasPrefix(p.ID, "cache-") {
			o.SessionID = "synthetic-fixed-prefix"
			o.CacheRetention = ai.CacheRetentionShort
		}
		return o
	default:
		o := ai.AnthropicOptions{MaxTokens: ai.Value(n)}
		if p.ID == "thinking-control" && id.Model == minimaxlive.M3 {
			o.ThinkingEnabled = ai.Value(false)
		}
		if level != "" {
			o.ThinkingEnabled = ai.Value(level != ai.ThinkingLevel("off"))
		}
		if id.Model == minimaxlive.M31 && level != "" {
			o.Effort = ai.AnthropicEffort(level)
		}
		return o
	}
}
func (s *minimaxSession) request(p minimaxlive.Case) (ai.Request, error) {
	p.ID = strings.TrimPrefix(p.ID, "research-")
	if p.ID == grokhistory.Scenario {
		return grokhistory.Request(ai.CallScope{TenantID: liveTenant})
	}
	req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic protocol probe. Reply with the integer 42.")}}
	if p.ID == "thinking-on" && minimaxReport.Identity.Model == minimaxlive.M3 {
		// Adaptive M3 may omit reasoning output for a trivial direct-answer task.
		// Exercise a bounded calculation while still requiring its own thinking.
		req.Messages = []ai.Message{ai.UserText("Synthetic reasoning task. Find the smallest positive integer n that gives remainders 7, 11, and 13 when divided by 17, 19, and 23 respectively. Work through the congruences and verify all three remainders before giving the final integer.")}
	}
	if p.ID == "tool-call" {
		req.Messages = []ai.Message{ai.UserText("Synthetic arithmetic task. You MUST call the add function with a=17 and b=25 before answering. Think through the calculation first.")}
		req.Tools = []ai.Tool{glmAdd}
	}
	if p.ID == "tool-result-native" {
		if s.tool.Message.StopReason != ai.StopReasonToolUse {
			return req, fmt.Errorf("own tool prerequisite missing")
		}
		req = s.toolRequest
		req.Messages = append([]ai.Message{}, req.Messages...)
		msg := s.tool.Message
		env, ok := msg.NativeState.Envelope()
		if !ok {
			return req, fmt.Errorf("own native provenance missing")
		}
		// Exercise the Host's persisted state integrity/trust boundary as documented.
		raw, e := json.Marshal(msg)
		if e != nil {
			return req, e
		}
		restored, e := minimaxlive.DecodeMessage(raw)
		if e != nil {
			return req, e
		}
		restored.NativeState, e = ai.TrustNativeState(newScope(), env)
		if e != nil {
			return req, e
		}
		req.Messages = append(req.Messages, restored)
		count := 0
		for i, part := range restored.Content {
			call, ok := part.(ai.ToolCall)
			if !ok {
				continue
			}
			v, e := ai.ValidateToolCall(restored, i, req.Tools)
			if e != nil {
				return req, e
			}
			var args struct{ A, B int }
			if e = v.Decode(&args); e != nil || args.A != 17 || args.B != 25 {
				return req, fmt.Errorf("unexpected synthetic tool arguments")
			}
			req.Messages = append(req.Messages, ai.ToolResultText(call, "42", false))
			count++
		}
		if count == 0 {
			return req, fmt.Errorf("own tool absent")
		}
	}
	if strings.HasPrefix(p.ID, "cache-") {
		req.Messages = []ai.Message{ai.UserText(strings.Repeat("Synthetic prefix, no personal data. ", 64) + "Reply 42.")}
	}
	if p.ID == "truncate" {
		req.Messages = []ai.Message{ai.UserText("Write a long explanation of all integers from one to one hundred.")}
	}
	if p.ID == "cancel" {
		req.Messages = []ai.Message{ai.UserText("Write the integers from one to one thousand slowly and explain each integer.")}
	}
	return req, nil
}
