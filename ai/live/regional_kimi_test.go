//go:build live

package live

import (
	"encoding/json"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Each entry's actual full protocol shape is saved for independent raw replay.
// Full stream keeps K3's default max; canonical simple levels clamp separately.
func kimiRegionalOptions(m ai.Model, budget int, kind string) ai.Options {
	effort := ai.ThinkingLevel("")
	if kind == "full-complete" {
		effort = ai.ThinkingHigh
	}
	switch m.API {
	case ai.APIOpenAICompletions:
		o := ai.ChatOptions{MaxTokens: ai.Value(budget), ReasoningEffort: effort}
		if kind == "tool-call" {
			o.ToolChoice = ai.Value(ai.ChatToolChoice{Mode: "required"})
		}
		return o
	case ai.APIOpenAIResponses:
		return ai.ResponsesOptions{MaxTokens: ai.Value(budget), ReasoningEffort: effort}
	default:
		o := ai.AnthropicOptions{MaxTokens: ai.Value(budget), Effort: ai.AnthropicEffort(effort)}
		if kind == "tool-call" {
			o.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "any"})
		}
		return o
	}
}

func kimiRegionalEffectiveOptions(m ai.Model, budget int, kind string) ai.Options {
	if kind != "simple-stream" && kind != "simple-complete" {
		return kimiRegionalOptions(m, budget, kind)
	}
	effort := ai.ThinkingLow
	if kind == "simple-complete" {
		effort = ai.ThinkingHigh
	}
	switch m.API {
	case ai.APIOpenAICompletions:
		return ai.ChatOptions{MaxTokens: ai.Value(budget), ReasoningEffort: effort}
	case ai.APIOpenAIResponses:
		return ai.ResponsesOptions{MaxTokens: ai.Value(budget), ReasoningEffort: effort}
	default:
		return ai.AnthropicOptions{MaxTokens: ai.Value(budget), Effort: ai.AnthropicEffort(effort)}
	}
}

func (g *regionalHTTPGate) reserveUSD(raw []byte, input int) error {
	if !supportmatrix.RegionalIsKimi(g.process.report.Combo) {
		return nil
	}
	var wire struct {
		Model         string `json:"model"`
		MaxTokens     int    `json:"max_tokens"`
		MaxCompletion int    `json:"max_completion_tokens"`
		MaxOutput     int    `json:"max_output_tokens"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return fmt.Errorf("Kimi encoded reservation invalid")
	}
	output := max(wire.MaxTokens, wire.MaxCompletion, wire.MaxOutput)
	amount, err := supportmatrix.RegionalReservedUSD(g.process.report.Combo, wire.Model, input, output)
	e := g.process.report.Regional
	if err != nil || e.ReservedUSD+amount > e.MaxUSD {
		return fmt.Errorf("Kimi explicit USD budget exhausted")
	}
	e.ReservedUSD += amount
	return nil
}
