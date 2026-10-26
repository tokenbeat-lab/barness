package minimaxlive_test

import "github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"

// These synthetic closed-cohort fixtures represent the original 13-call,
// 12-HTTP M3 contract (39/36 across three routes). They must not be regenerated
// from today's additive Plan14: that would rewrite historical accounting or
// manufacture authority for the xAI scene. Issue barness-ai-grok/14 retains all
// raw historical bytes and pins; this fixture supplies no current permission.
func historicalM3Plan() []minimaxlive.Case {
	return []minimaxlive.Case{
		{ID: "full-stream", Output: 1024, Research: "default wire and terminal, raw actual model/ID and usage"},
		{ID: "full-complete", Output: 1024, Research: "full Complete text"},
		{ID: "simple-stream", Output: 1024, Research: "simple Stream default selection"},
		{ID: "simple-complete", Output: 1024, Research: "simple Complete default selection"},
		{ID: "thinking-on", Output: 2048, Research: "own explicit reasoning on/high or mandatory high"},
		{ID: "thinking-control", Output: 1024, Research: "binary off or mandatory max"},
		{ID: "tool-call", Output: 4096, Research: "own automatic tool and reasoning in the same assistant turn"},
		{ID: "tool-result-native", Output: 2048, Research: "Host validated tool and original trusted reasoning/native state continuation"},
		{ID: "truncate", Output: 1, Research: "length terminal retains actual known usage"},
		{ID: "cache-first", Output: 1024, Research: "fixed repeated synthetic prefix, observed usage presence only"},
		{ID: "cache-repeat", Output: 1024, Research: "identical prefix; no promised cache hit"},
		{ID: "cancel", Output: 2048, Research: "cancel after visible content, unknown usage remains unknown"},
		{ID: "preflight-error", Output: 1, Research: "invalid output request fails before any Attempt"},
	}
}
