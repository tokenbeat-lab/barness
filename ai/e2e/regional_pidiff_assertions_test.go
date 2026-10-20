package e2e

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// Explicit source-backed literals constrain each scoped exception before the
// ledger sees it. All remaining request, event and result fields stay strict.
func regionalFrozenAssertions(t *testing.T, ev *evidence.Case, r regionalFrozenRoute, model, stage, entry string, pi, b pioracle.Observation) {
	t.Helper()
	tools := stage != "text"
	replay := stage == "tool-replay"
	for _, side := range []struct {
		name string
		obs  pioracle.Observation
	}{{"pi", pi}, {"barness", b}} {
		q := side.obs.Request
		body := bodyFields(t, q.Body)
		path, query := "/v1/chat/completions", ""
		if r.api == ai.APIAnthropicMessages {
			path, query = "/anthropic/v1/messages", "beta=true"
		}
		ev.Check(side.name+" exact route/auth/model", q.Method == "POST" && q.Path == path && q.Query == query && q.Auth == tenantA.alias && jsonEqual(body["model"], mustMarshal(t, model)), "%+v", q)
		ev.Check(side.name+" length is exact complete bytes", q.Headers["Content-Length"] == strconv.Itoa(len(q.Body)), "%v", q.Headers)
		var usage struct {
			Provider string
			API      string
			Model    string
			Usage    struct {
				Input, Output, CacheRead, CacheWrite, TotalTokens float64
				CacheWrite1h                                      *float64
				Cost                                              map[string]float64
			}
		}
		mustUnmarshal(t, side.obs.Result, &usage)
		ev.Check(side.name+" actual result identity", usage.Provider == string(r.provider) && usage.API == string(r.api) && usage.Model == model, "%s", side.obs.Result)
		rates := [3]float64{1.4, 4.4, .26}
		input, output, cached, total := 10., 8., 2., 20.
		if r.provider == ai.ProviderZAI {
			if tools {
				input, output, cached = 12, 8, 0
			}
			if replay {
				input, output, total = 30, 12, 42
			}
			want := json.RawMessage(nil)
			if side.name == "pi" {
				want = json.RawMessage(`{"include_usage":true}`)
			}
			ev.Check(side.name+" GLM stream usage request shape", len(want) == 0 && len(body["stream_options"]) == 0 || jsonEqual(body["stream_options"], want), "%s", q.Body)
			if !tools {
				want = nil
				if side.name == "pi" {
					want = json.RawMessage(`{"type":"disabled"}`)
				}
				ev.Check(side.name+" GLM omitted default stays explicit", len(want) == 0 && len(body["thinking"]) == 0 || jsonEqual(body["thinking"], want), "%s", q.Body)
			} else {
				want = json.RawMessage(`{"type":"enabled"}`)
				if side.name == "pi" {
					want = json.RawMessage(`{"type":"enabled","clear_thinking":false}`)
				}
				ev.Check(side.name+" exact enabled GLM switch", jsonEqual(body["thinking"], want) && string(body["tool_stream"]) == "true" && string(body["reasoning_effort"]) == `"high"`, "%s", q.Body)
				var listed []struct{ Function map[string]json.RawMessage }
				mustUnmarshal(t, body["tools"], &listed)
				ev.Check(side.name+" exactly two function declarations", len(listed) == 2, "%s", q.Body)
				for _, tool := range listed {
					ev.Check(side.name+" explicit strict false versus unclaimed strict", side.name == "pi" && string(tool.Function["strict"]) == "false" || side.name == "barness" && len(tool.Function["strict"]) == 0, "%s", q.Body)
				}
			}
		} else {
			rates = [3]float64{.3, 1.2, .06}
			if side.name == "barness" {
				rates = [3]float64{.3124070217, 1.2496280869, .0624814043}
			}
			input, output, cached, total = 5, 3, 2, 10
			if tools && !replay {
				input, output, total = 7, 11, 20
			}
			max := 512000
			if side.name == "barness" {
				max = 2048
			}
			ev.Check(side.name+" independent model ceiling", string(body["max_tokens"]) == strconv.Itoa(max), "%s", q.Body)
			thinking := json.RawMessage(nil)
			if tools {
				thinking = json.RawMessage(`{"type":"adaptive"}`)
				if side.name == "pi" {
					thinking = json.RawMessage(`{"type":"enabled","budget_tokens":1024,"display":"summarized"}`)
					if entry == "streamSimple" {
						thinking = json.RawMessage(`{"type":"enabled","budget_tokens":16384,"display":"summarized"}`)
					}
				}
			} else if side.name == "pi" && entry == "streamSimple" {
				thinking = json.RawMessage(`{"type":"disabled"}`)
			}
			ev.Check(side.name+" exact CN thinking contract", len(thinking) == 0 && len(body["thinking"]) == 0 || jsonEqual(body["thinking"], thinking), "%s", q.Body)
			browser, beta := "", ""
			if side.name == "pi" {
				browser = "true"
				if tools {
					beta = "interleaved-thinking-2025-05-14"
				}
			}
			ev.Check(side.name+" explicit browser/beta presence", q.Headers["Anthropic-Dangerous-Direct-Browser-Access"] == browser && q.Headers["Anthropic-Beta"] == beta, "%v", q.Headers)
			var messages []struct{ Content []map[string]json.RawMessage }
			mustUnmarshal(t, body["messages"], &messages)
			mi, ci := 0, 0
			if replay {
				mi, ci = 2, 1
			}
			cache := json.RawMessage(nil)
			if side.name == "pi" {
				cache = json.RawMessage(`{"type":"ephemeral"}`)
			}
			ev.Check(side.name+" only the exact ephemeral boundary", len(cache) == 0 && len(messages[mi].Content[ci]["cache_control"]) == 0 || jsonEqual(messages[mi].Content[ci]["cache_control"], cache), "%s", q.Body)
			if tools {
				var declarations []map[string]json.RawMessage
				mustUnmarshal(t, body["tools"], &declarations)
				ev.Check("one own tool declaration", len(declarations) == 1, "%s", q.Body)
				for _, d := range declarations {
					ev.Check(side.name+" tool cache boundary", len(cache) == 0 && len(d["cache_control"]) == 0 || jsonEqual(d["cache_control"], cache), "%s", q.Body)
					ev.Check(side.name+" eager stream capability", side.name == "pi" && string(d["eager_input_streaming"]) == "true" || side.name == "barness" && len(d["eager_input_streaming"]) == 0, "%s", q.Body)
				}
			}
			ev.Check(side.name+" no invented one-hour writes", side.name == "pi" && usage.Usage.CacheWrite1h != nil && *usage.Usage.CacheWrite1h == 0 || side.name == "barness" && usage.Usage.CacheWrite1h == nil, "%s", side.obs.Result)
		}
		u := usage.Usage
		ev.Check(side.name+" literal scripted usage", u.Input == input && u.Output == output && u.CacheRead == cached && u.CacheWrite == 0 && u.TotalTokens == total, "%s", side.obs.Result)
		costs := map[string]float64{"input": input * rates[0] / 1e6, "output": output * rates[1] / 1e6, "cacheRead": cached * rates[2] / 1e6, "cacheWrite": 0}
		costs["total"] = costs["input"] + costs["output"] + costs["cacheRead"]
		for k, want := range costs {
			ev.Check(side.name+" exact independent cost "+k, math.Abs(u.Cost[k]-want) < 1e-15, "got %g want %g", u.Cost[k], want)
		}
	}
}
