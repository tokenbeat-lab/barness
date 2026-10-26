package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// ADR-0049 v2 test vector for the synthetic replay scope, calculated outside
// the Client: tenant-a / xai / acct-xai / synthetic-real-replay-cache-session.
const xaiRealReplayCacheKey = "b2dfee34eea3862c71bbcc9f4ddfe42181fa12e2a6f2ea1bf9a10e8348d71b5b"

func TestXAIRealCaptureProvenance(t *testing.T) {
	ev := run.Case(t, "XAI-E17-real-replay-provenance")
	manifest, captures, raw := loadXAIRealCaptures(t)
	ev.Fixture("provenance.json", raw)
	ev.Check("actual candidate source remains explicit", manifest.Schema == 1 && manifest.Kind == "deidentified-live-replay" && manifest.NativeSourceSHA256 == "05e89a133ceb71b6725ad7a6ee889170f6e49fa7a72845928b202416a7d2838b" && manifest.SourceReportSHA256 == "2c3a8da8b41659c5500c8f840c5da05c51d016f858236b79609cc88fab8dc36f" && strings.Contains(manifest.SourceBundle, "20261026T050451.200594000Z"), "unexpected capture provenance")
	seen := map[string]bool{}
	for _, capture := range captures {
		seen[capture.Scenario] = true

		ev.Check("captured request has only deidentified protocol state", checkXAIRealSanitization(t, capture.Request), "unsanitized captured request")
		for _, line := range strings.Split(capture.ResponseBody, "\n") {
			if strings.HasPrefix(line, "data:") {
				raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if raw != "[DONE]" {
					ev.Check("captured response has only deidentified protocol state", checkXAIRealSanitization(t, []byte(raw)), "unsanitized captured response")
				}
			}
		}
	}
	ev.Check("all actual scenarios and calls are assembled", len(captures) == 24 && len(seen) == 18, "captures=%d scenarios=%d", len(captures), len(seen))
}

func TestXAIRealCaptureReplay(t *testing.T) {
	_, captures, _ := loadXAIRealCaptures(t)
	for _, scenario := range []string{"entry-stream", "entry-complete", "entry-stream-simple", "entry-complete-simple", "effort-default", "effort-low", "effort-medium", "effort-high", "effort-xhigh", "encrypted-continuation", "tools-auto", "tools-required", "tools-none", "tools-parallel", "user-png", "user-jpeg", "tool-result-image", "cache-usage"} {
		t.Run(scenario, func(t *testing.T) {
			ev := run.Case(t, "XAI-E17-real-replay-"+scenario)
			var group []xaiRealCapture
			for _, capture := range captures {
				if capture.Scenario == scenario {
					group = append(group, capture)
				}
			}
			if !ev.Check("source scenario exists", len(group) > 0, "missing scenario") {
				return
			}
			w := newXAIWorld(t, nil)
			var replies []provider.Reply
			for _, capture := range group {
				replies = append(replies, capture.reply())
			}
			enqueue(ev, w, replies...)
			req := group[0].initial(t)
			var previous ai.Result
			cacheKey := ""
			for index, capture := range group {
				ev.Record(capture.Label+"-source-capture", capture)
				if index > 0 {
					req = capture.continueRequest(t, req, previous, group[index-1])
				}
				o := capture.options(t)
				var result outcome
				scope := textScope(t.Name() + "-" + capture.Label)
				switch {
				case scenario == "entry-stream-simple":
					result = drain(w.client.StreamSimple(ctxFor(t), scope, xaiTarget(), req, ai.SimpleOptions{MaxTokens: o.MaxTokens, Reasoning: ai.ThinkingLow}))
				case scenario == "entry-complete-simple":
					result.result, result.err = w.client.CompleteSimple(ctxFor(t), scope, xaiTarget(), req, ai.SimpleOptions{MaxTokens: o.MaxTokens, Reasoning: ai.ThinkingLow})
				case scenario == "entry-stream" || capture.Label == "calls" || capture.Label == "continued":
					result = drain(w.client.Stream(ctxFor(t), scope, xaiTarget(), req, o))
				default:
					result.result, result.err = w.client.Complete(ctxFor(t), scope, xaiTarget(), req, o)
				}
				ev.Record(capture.Label+"-result", result.result)
				ev.Record(capture.Label+"-events", result.events)
				if !ev.Check("actual captured response replays "+capture.Label, result.err == nil && len(w.provider.Requests()) == index+1, "%v", result.err) {
					return
				}
				message := result.result.Message
				ev.Check("visible answer matches captured terminal "+capture.Label, textOf(message) == capture.Expected.Text, "text changed")
				usage, expected := message.Usage, capture.Expected.Usage
				reasoning, reported := usage.Reasoning.Get()
				wantReasoning, _ := expected.Reasoning.Get()
				ev.Check("reported usage normalizes exactly once "+capture.Label, usage.Input == expected.Input && usage.Output == expected.Output && usage.CacheRead == expected.CacheRead && usage.CacheWrite == 0 && usage.TotalTokens == expected.TotalTokens && reported && reasoning == wantReasoning && reasoning > 0 && usage.Input+usage.CacheRead+usage.Output == usage.TotalTokens, "usage=%+v want=%+v", usage, expected)
				ev.Check("attempt reports complete usage "+capture.Label, len(result.result.Metadata.Attempts) == 1 && result.result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "attempt metadata incomplete")
				var native []json.RawMessage
				calls := 0
				for _, block := range message.Content {
					switch value := block.(type) {
					case ai.Thinking:
						native = append(native, json.RawMessage(value.Signature))
					case ai.ToolCall:
						if calls < len(capture.Expected.Calls) {
							want := capture.Expected.Calls[calls]
							ev.Check("tool identity and complete arguments survive "+capture.Label, value.ID != "" && value.Name == want.Name && jsonEqual([]byte(value.RawArguments), []byte(want.Arguments)), "tool changed")
						}
						calls++
					}
				}
				ev.Check("whole sanitized native reasoning retained "+capture.Label, jsonEqual(mustMarshal(t, native), mustMarshal(t, capture.Expected.Reasoning)), "native item changed")
				ev.Check("captured function count and stop reason "+capture.Label, calls == len(capture.Expected.Calls) && (calls > 0 && message.StopReason == ai.StopReasonToolUse || calls == 0 && message.StopReason == ai.StopReasonStop), "calls=%d stop=%s", calls, message.StopReason)
				actual := bodyFields(t, w.provider.Requests()[index].Body)
				want := bodyFields(t, capture.Request)
				if capture.body(t).PromptCacheKey != "" {
					var current string
					mustUnmarshal(t, actual["prompt_cache_key"], &current)
					ev.Check("scoped cache hint matches independent replay scope", current == xaiRealReplayCacheKey && (cacheKey == "" || current == cacheKey), "cache hint changed")
					cacheKey = current
					// The source account's hint was removed during deidentification.
					// Replace only that placeholder with the replay scope's known
					// vector; every actual request field stays in the comparison.
					want["prompt_cache_key"] = mustMarshal(t, xaiRealReplayCacheKey)
				}
				ev.Check("entire request matches captured history pairing and image order "+capture.Label, jsonEqual(mustMarshal(t, actual), mustMarshal(t, want)), "captured request projection changed")
				ev.Check("restored source retains native trust "+capture.Label, result.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "unexpected downgrade")
				previous = result.result
			}
			ev.Record("requests", w.provider.Requests())
		})
	}
}
