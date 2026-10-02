package ai

import "encoding/json"

// chatUsage converts, classifies and prices one call's Chat Completions
// usage with the model's rates. reasoningExpected is whether the call asked
// for reasoning: only then is a missing reasoning count a gap in the report.
// DeepSeek omits completion_tokens_details when thinking is off (observed
// live 2026-10-02), and a call without thinking has no reasoning to count,
// so its usage stays complete (maintainer decision 2026-10-03, issue 33).
type chatUsage struct {
	model             Model
	reasoningExpected bool
}

// chatUsageCounts are the counts pi's parseChunkUsage reads; each stays raw
// so a missing or null count can be told from a reported one.
type chatUsageCounts struct {
	PromptTokens        json.RawMessage `json:"prompt_tokens"`
	CompletionTokens    json.RawMessage `json:"completion_tokens"`
	PromptTokensDetails *struct {
		CachedTokens     json.RawMessage `json:"cached_tokens"`
		CacheWriteTokens json.RawMessage `json:"cache_write_tokens"`
	} `json:"prompt_tokens_details"`
	// PromptCacheHitTokens is DeepSeek's cache read count, CachedTokens
	// Kimi's.
	PromptCacheHitTokens    json.RawMessage `json:"prompt_cache_hit_tokens"`
	CachedTokens            json.RawMessage `json:"cached_tokens"`
	CompletionTokensDetails *struct {
		ReasoningTokens json.RawMessage `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

// of converts a reported usage object as pi's parseChunkUsage does: cache
// reads from the first reported of OpenAI's, DeepSeek's and Kimi's fields,
// cache writes (OpenRouter's) counted apart, input without either, and the
// total summed from the parts rather than taken from the provider.
func (c chatUsage) of(raw json.RawMessage) (Usage, UsageReporting) {
	var r chatUsageCounts
	_ = json.Unmarshal(raw, &r)
	var cachedRaw, writeRaw, reasoningRaw json.RawMessage
	if d := r.PromptTokensDetails; d != nil {
		cachedRaw, writeRaw = d.CachedTokens, d.CacheWriteTokens
	}
	if d := r.CompletionTokensDetails; d != nil {
		reasoningRaw = d.ReasoningTokens
	}
	// pi's `cached_tokens ?? prompt_cache_hit_tokens ?? cached_tokens ?? 0`.
	cached, cacheReported := int64(0), false
	for _, v := range []json.RawMessage{cachedRaw, r.PromptCacheHitTokens, r.CachedTokens} {
		if n, ok := usageCount(v); ok {
			cached, cacheReported = n, true
			break
		}
	}
	prompt, promptReported := usageCount(r.PromptTokens)
	output, outputReported := usageCount(r.CompletionTokens)
	reasoning, reasoningReported := usageCount(reasoningRaw)
	cacheWrite, _ := usageCount(writeRaw)
	u := Usage{
		Input:      max(0, prompt-cached-cacheWrite),
		Output:     output,
		CacheRead:  cached,
		CacheWrite: cacheWrite,
		Reasoning:  Value(reasoning),
	}
	u.TotalTokens = u.Input + u.Output + u.CacheRead + u.CacheWrite
	u.Cost = c.model.Cost.estimate(u)
	reporting := UsagePartial
	if promptReported && outputReported && cacheReported && (reasoningReported || !c.reasoningExpected) {
		reporting = UsageComplete
	}
	return u, reporting
}

// usageCount is a reported token count; missing, null and anything but a
// number are not reported and read as 0, as pi's `|| 0` and `?? 0`.
func usageCount(raw json.RawMessage) (int64, bool) {
	var n float64
	if len(raw) == 0 || raw[0] == 'n' || json.Unmarshal(raw, &n) != nil {
		return 0, false
	}
	return int64(n), true
}

// report replaces the message's usage with a chunk's, as pi does for every
// chunk that carries one, and records it on the attempt.
func (p *chatParser) report(raw json.RawMessage) {
	u, reporting := p.usage.of(raw)
	p.out.usage(u)
	p.attempt.usage(reporting, u)
}
