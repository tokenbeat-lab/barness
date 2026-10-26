package supportmatrix

import (
	"encoding/json"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/tokenbeat-lab/barness/ai"
)

func xaiTerminal(body string) (map[string]json.RawMessage, []map[string]json.RawMessage, map[string]string, error) {
	decoder := ssestream.NewDecoder(&http.Response{Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body))})
	defer decoder.Close()
	var terminal map[string]json.RawMessage
	var frames []map[string]json.RawMessage
	ended := false
	for decoder.Next() {
		event := decoder.Event()
		raw := strings.TrimSpace(string(event.Data))
		if raw == "[DONE]" {
			if terminal == nil || ended {
				return nil, nil, nil, refuse("xAI SSE done marker without unique terminal")
			}
			ended = true
			continue
		}
		if raw == "" {
			continue
		}
		if ended || len(raw) > 256<<10 {
			return nil, nil, nil, refuse("xAI SSE data after terminal or oversized frame")
		}
		var frame map[string]json.RawMessage
		if regionalDecode([]byte(raw), &frame) != nil {
			return nil, nil, nil, refuse("xAI SSE frame malformed")
		}
		var kind string
		if json.Unmarshal(frame["type"], &kind) != nil || event.Type != "" && event.Type != "message" && event.Type != kind {
			return nil, nil, nil, refuse("xAI SSE event name differs")
		}
		frames = append(frames, frame)
		switch kind {
		case "response.completed":
			var complete map[string]json.RawMessage
			if regionalDecode(frame["response"], &complete) != nil || complete == nil {
				return nil, nil, nil, refuse("xAI completed response absent")
			}
			if terminal != nil && !xaiEqual(terminal, complete) {
				return nil, nil, nil, refuse("xAI conflicting completed snapshots")
			}
			terminal = complete
		case "response.failed", "response.incomplete", "error", "response.error":
			return nil, nil, nil, refuse("xAI SSE did not complete successfully")
		default:
			if terminal != nil {
				return nil, nil, nil, refuse("xAI SSE data after completed snapshot")
			}
		}
	}
	if decoder.Err() != nil || terminal == nil {
		return nil, nil, nil, refuse("xAI SSE lacks complete terminal")
	}
	var status, model string
	json.Unmarshal(terminal["status"], &status)
	json.Unmarshal(terminal["model"], &model)
	if status != "completed" || model != xaiModel {
		return nil, nil, nil, refuse("xAI terminal model or outcome differs")
	}
	for _, k := range []string{"error", "incomplete_details"} {
		if v := strings.TrimSpace(string(terminal[k])); v != "" && v != "null" {
			return nil, nil, nil, refuse("xAI successful terminal contains failure detail")
		}
	}
	var output []map[string]json.RawMessage
	if regionalDecode(terminal["output"], &output) != nil || len(output) == 0 {
		return nil, nil, nil, refuse("xAI terminal output absent")
	}
	ordered, visible, err := verifyXAIWire(frames, terminal, output)
	if err != nil {
		return nil, nil, nil, err
	}
	return terminal, ordered, visible, nil
}
func xaiCount(raw json.RawMessage) (int64, error) {
	s := string(raw)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 {
		return n, nil
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil || !xaiFinite(value) || math.Trunc(value) != value || value > float64(math.MaxInt64) {
		return 0, refuse("xAI exact token count invalid")
	}
	if at := strings.IndexAny(s, "eE"); at >= 0 {
		exponent, err := strconv.ParseInt(s[at+1:], 10, 32)
		if err != nil || exponent > 256 || exponent < -256 {
			return 0, refuse("xAI token exponent exceeds exact arithmetic bound")
		}
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok || !r.IsInt() || r.Sign() < 0 || !r.Num().IsInt64() {
		return 0, refuse("xAI raw token count is not a finite nonnegative exact integer")
	}
	return r.Num().Int64(), nil
}
func verifyXAIUsage(c *xaiCall) error {
	var usage, inputs, outputs map[string]json.RawMessage
	if regionalDecode(c.terminal["usage"], &usage) != nil || regionalDecode(usage["input_tokens_details"], &inputs) != nil || regionalDecode(usage["output_tokens_details"], &outputs) != nil {
		return refuse("xAI required raw usage details missing")
	}
	if err := verifyXAIPricingScope(c.terminal, usage, inputs, outputs); err != nil {
		return err
	}
	p, e1 := xaiCount(usage["input_tokens"])
	o, e2 := xaiCount(usage["output_tokens"])
	t, e3 := xaiCount(usage["total_tokens"])
	r, e4 := xaiCount(outputs["reasoning_tokens"])
	cache, e5 := xaiCount(inputs["cached_tokens"])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || p > 8192 || t < p || cache > p {
		return refuse("xAI complete usage or input reservation differs")
	}
	b := t - p
	if r > b || b != o && (o > math.MaxInt64-r || b != o+r) {
		return refuse("xAI raw output/reasoning arithmetic contradicts total")
	}
	u := c.message.Usage
	reported, ok := u.Reasoning.Get()
	if u.Input != p-cache || u.CacheRead != cache || u.CacheWrite != 0 || u.Output != b || u.TotalTokens != t || !ok || reported != r || !u.CacheWrite1h.IsZero() || u.Modalities != nil {
		return refuse("xAI normalized usage differs from raw report")
	}
	model, exists := ai.BuiltinCatalog().Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, xaiModel)
	if !exists {
		return refuse("xAI formal price snapshot missing")
	}
	rates := model.Cost.CostRates
	threshold := int64(-1)
	for _, tier := range model.Cost.Tiers {
		if p > tier.InputTokensAbove && tier.InputTokensAbove > threshold {
			rates, threshold = tier.CostRates, tier.InputTokensAbove
		}
	}
	cost := ai.UsageCost{Input: float64(p-cache) * rates.Input / 1e6, Output: float64(b) * rates.Output / 1e6, CacheRead: float64(cache) * rates.CacheRead / 1e6}
	cost.Total = cost.Input + cost.Output + cost.CacheRead
	a := c.result.Metadata.Attempts[0]
	low, lok := a.CostEstimate.Minimum.Get()
	high, hok := a.CostEstimate.Maximum.Get()
	if a.CostEstimate.Status != ai.CostResolved || !lok || !hok || !xaiCostEqual(u.Cost, cost) || !xaiCostEqual(low, cost) || !xaiCostEqual(high, cost) {
		return refuse("xAI independently calculated cost differs")
	}
	c.input, c.outputTokens, c.tokens, c.usd = p, b, t, high.Total
	return nil
}
func xaiCostEqual(a, b ai.UsageCost) bool {
	return xaiAmountEqual(a.Input, b.Input) && xaiAmountEqual(a.Output, b.Output) && xaiAmountEqual(a.CacheRead, b.CacheRead) && xaiAmountEqual(a.CacheWrite, b.CacheWrite) && xaiAmountEqual(a.Total, b.Total)
}
