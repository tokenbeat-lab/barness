package e2e

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// Every new accepted wire/cost difference is independently constrained by
// literal source-backed expectations, before classification can excuse it.
func miniMaxCheckPiWires(t *testing.T, ev *evidence.Case, pi, barness pioracle.Observation) {
	t.Helper()
	tools := strings.Contains(ev.ID(), "-tools-") || strings.Contains(ev.ID(), "-tool-replay-")
	replay := strings.Contains(ev.ID(), "-tool-replay-")
	simple := strings.HasSuffix(ev.ID(), "-streamSimple")
	for _, side := range []struct {
		name  string
		obs   pioracle.Observation
		max   int
		rates [3]float64
	}{
		{"pi", pi, 512000, [3]float64{.3, 1.2, .06}},
		{"barness", barness, 131072, [3]float64{1, 2, .1}},
	} {
		r := side.obs.Request
		var body struct {
			Model    string                                `json:"model"`
			Max      int                                   `json:"max_tokens"`
			Thinking json.RawMessage                       `json:"thinking"`
			Messages []struct{ Content []json.RawMessage } `json:"messages"`
			Tools    []map[string]json.RawMessage          `json:"tools"`
		}
		mustUnmarshal(t, r.Body, &body)
		ev.Check(side.name+" exact binding and default budget", r.Method == "POST" && r.Path == "/anthropic/v1/messages" && r.Query == "beta=true" && r.Auth == anthropicA.alias && body.Model == "MiniMax-M3" && body.Max == side.max, "%s", r.Body)
		ev.Check(side.name+" length derived from complete body", r.Headers["Content-Length"] == strconv.Itoa(len(r.Body)), "%v", r.Headers["Content-Length"])
		wantThinking := json.RawMessage(nil)
		if side.name == "pi" && simple {
			wantThinking = json.RawMessage(`{"type":"disabled"}`)
		}
		if tools {
			wantThinking = json.RawMessage(`{"type":"adaptive"}`)
			if side.name == "pi" {
				wantThinking = json.RawMessage(`{"type":"enabled","budget_tokens":1024,"display":"summarized"}`)
				if simple {
					wantThinking = json.RawMessage(`{"type":"enabled","budget_tokens":16384,"display":"summarized"}`)
				}
			}
		}
		ev.Check(side.name+" own default or explicit thinking contract", len(wantThinking) == 0 && len(body.Thinking) == 0 || jsonEqual(body.Thinking, wantThinking), "%s", body.Thinking)
		wantBrowser, wantBeta := "", ""
		if side.name == "pi" {
			wantBrowser = "true"
			if tools {
				wantBeta = "interleaved-thinking-2025-05-14"
			}
		}
		ev.Check(side.name+" explicit browser and thinking beta", r.Headers["Anthropic-Dangerous-Direct-Browser-Access"] == wantBrowser && r.Headers["Anthropic-Beta"] == wantBeta, "%v", r.Headers)
		mi, ci := 0, 0
		if replay {
			mi, ci = 2, 1
		}
		boundary := bodyFields(t, body.Messages[mi].Content[ci])
		if side.name == "pi" {
			ev.Check("pi frozen ephemeral boundary", jsonEqual(boundary["cache_control"], []byte(`{"type":"ephemeral"}`)), "%s", r.Body)
		} else {
			ev.Check("MiniMax has no active-cache claim", len(boundary["cache_control"]) == 0, "%s", r.Body)
		}
		if tools {
			for _, key := range []string{"cache_control", "eager_input_streaming"} {
				want := json.RawMessage(nil)
				if side.name == "pi" {
					want = json.RawMessage(`true`)
					if key == "cache_control" {
						want = json.RawMessage(`{"type":"ephemeral"}`)
					}
				}
				ev.Check(side.name+" own tool "+key, len(want) == 0 && len(body.Tools[0][key]) == 0 || jsonEqual(body.Tools[0][key], want), "%s", r.Body)
			}
		}
		if replay {
			ev.Check(side.name+" complete signed thinking survives tool continuation", jsonEqual(body.Messages[1].Content[0], []byte(`{"type":"thinking","thinking":"Synthetic reasoning.","signature":"sig-part1-part2"}`)), "%s", r.Body)
		}
		var result struct {
			Usage struct {
				Input, Output, CacheRead, CacheWrite, TotalTokens float64
				CacheWrite1h                                      *float64
				Cost                                              map[string]float64
			}
		}
		mustUnmarshal(t, side.obs.Result, &result)
		u := result.Usage
		input, output, total := 5., 3., 10.
		if tools && !replay {
			input, output, total = 7, 11, 20
		}
		ev.Check(side.name+" literal token counts remain equal", u.Input == input && u.Output == output && u.CacheRead == 2 && u.CacheWrite == 0 && u.TotalTokens == total, "%s", side.obs.Result)
		ev.Check(side.name+" unsupported one-hour usage stays explicit", side.name == "pi" && u.CacheWrite1h != nil && *u.CacheWrite1h == 0 || side.name == "barness" && u.CacheWrite1h == nil, "%s", side.obs.Result)
		// Fixed worked examples: candidate rates are synthetic; neither side
		// is being certified as the current standard API/M Plan price here.
		wantCosts := map[string]float64{"input": input * side.rates[0] / 1e6, "output": output * side.rates[1] / 1e6, "cacheRead": 2 * side.rates[2] / 1e6, "cacheWrite": 0}
		wantCosts["total"] = wantCosts["input"] + wantCosts["output"] + wantCosts["cacheRead"]
		for key, want := range wantCosts {
			ev.Check(side.name+" sourced/test-only cost "+key, math.Abs(u.Cost[key]-want) < 1e-15, "got=%g want=%g", u.Cost[key], want)
		}
	}
}

func TestMiniMaxPiLedgerBounds(t *testing.T) {
	ev := run.Case(t, "MM06-ledger-bounds")
	l := loadLedger(t)
	diff := pioracle.Diff{Path: "request.body.max_tokens", Kind: pioracle.Changed}
	ev.Check("exact new case is explained", l.Classify("anthropic-messages", "PIDIFF-MM06-messages-text-stream", []pioracle.Diff{diff}).Pass, "missing exact declaration")
	ev.Check("unlisted case stays pending", !l.Classify("anthropic-messages", "unlisted-minimax-case", []pioracle.Diff{diff}).Pass, "case scope widened")
	ev.Check("other protocol stays pending", !l.Classify("openai-completions", "PIDIFF-MM06-messages-text-stream", []pioracle.Diff{diff}).Pass, "protocol scope widened")
	diff.Path = "request.body.model"
	ev.Check("unlisted field stays pending", !l.Classify("anthropic-messages", "PIDIFF-MM06-messages-text-stream", []pioracle.Diff{diff}).Pass, "whole request exempted")
	_, blanket := l.Route("minimax", "anthropic-messages")
	ev.Check("existing MiniMax Messages route never exempted", !blanket, "native route hidden")
}
