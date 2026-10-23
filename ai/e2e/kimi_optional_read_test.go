package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiOptionalCacheRead(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses} {
			for _, sc := range []struct {
				name, read, legacy string
				want               int64
				invalid            bool
			}{
				{"missing", "", "", 0, false}, {"null", "null", "", 0, false}, {"zero", "0", "", 0, false}, {"positive", "20", "", 20, false},
				{"legacy", "", "30", 30, false}, {"null-legacy", "null", "30", 30, false}, {"nested-wins", "20", "30", 20, false},
				{"negative", "-1", "", 0, true}, {"fraction", "1.00000000000000001", "", 0, true}, {"overflow", "9223372036854775808", "", 0, true}, {"string", `"20"`, "", 0, true}, {"legacy-invalid", "", "-1", 0, true},
			} {
				if sc.legacy != "" && api != ai.APIOpenAICompletions {
					continue
				}
				t.Run(string(market)+"/"+string(api)+"/"+sc.name, func(t *testing.T) {
					ev := run.Case(t, "KIMI-"+string(market)+"-"+string(api)+"-K10-optional-read-"+sc.name)
					rec := newRecorder()
					configure := func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 }
					w := newKimiChatWorld(t, market, nil, configure)
					target := kimiChatTarget()
					input, output, details, outdetails := "prompt_tokens", "completion_tokens", "prompt_tokens_details", "completion_tokens_details"
					if api == ai.APIOpenAIResponses {
						w = newKimiResponsesWorld(t, market, nil, configure)
						target = kimiResponsesTarget()
						input, output, details, outdetails = "input_tokens", "output_tokens", "input_tokens_details", "output_tokens_details"
					}
					usage := map[string]any{input: 120, output: 50, "total_tokens": 170, details: map[string]any{"cache_write_tokens": 0}, outdetails: map[string]any{"reasoning_tokens": 10}}
					if sc.read != "" {
						usage[details].(map[string]any)["cached_tokens"] = json.RawMessage(sc.read)
					}
					if sc.legacy != "" {
						usage["cached_tokens"] = json.RawMessage(sc.legacy)
					}
					raw := mustMarshal(t, usage)
					ev.Record("reported-usage", json.RawMessage(raw))
					if api == ai.APIOpenAICompletions {
						enqueue(ev, w, chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":` + string(raw) + `}`)}, ""))
					} else {
						enqueue(ev, w, kimiResponsesReply(t, json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[`+kimiResponsesMessage+`],"usage":`+string(raw)+`}}`)))
					}
					scope := textScope("optional-read-" + string(market) + "-" + string(api) + "-" + sc.name)
					result, err := w.client.Complete(ctxFor(t), scope, target, kimiChatRequest(), nil)
					ev.Record("result", result)
					obs := rec.awaitCall(ev, scope.RequestID)
					ev.Record("observations", obs)
					if sc.invalid {
						ev.Check("reported invalid read rejects", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", err)
						return
					}
					if !ev.Check("optional read succeeds", err == nil && len(result.Metadata.Attempts) == 1, "%v", err) {
						return
					}
					a := result.Metadata.Attempts[0]
					u := result.Message.Usage
					ev.Check("read default changes reporting certainty", a.UsageReporting == ai.UsageComplete && a.CostEstimate.Status == ai.CostResolved && u.CacheRead == sc.want && u.Input == 120-sc.want, "%+v", a)
					ev.Check("all delivered usage snapshots agree", len(obs) == 4 && obs[2].Attempt.Usage == u && obs[2].Attempt.CostEstimate == a.CostEstimate && obs[3].Usage == u && a.Usage == u, "%+v", obs)
				})
			}
		}
	}
}

func TestKimiChinaRealOptionalReadReplay(t *testing.T) {
	ev := run.Case(t, "KIMI-CN-CHAT-K10-real-optional-read-replay")
	raw, err := os.ReadFile("testdata/kimi-cn-chat/optional-read.sse")
	if err != nil {
		t.Fatal(err)
	}
	origin, err := os.ReadFile("testdata/kimi-cn-chat/optional-read-origin.json")
	if err != nil {
		t.Fatal(err)
	}
	var p struct{ FixtureSHA256 string }
	mustUnmarshal(t, origin, &p)
	sum := sha256.Sum256(raw)
	ev.Check("fixture retains exact captured response bytes", hex.EncodeToString(sum[:]) == p.FixtureSHA256 && p.FixtureSHA256 == "b2b11202e8e178f527dd0014558c6f130eb56426d1f6b4a73db63d7dc552b185", "%s", p.FixtureSHA256)
	ev.Fixture("original.sse", raw)
	ev.Record("origin", json.RawMessage(origin))
	w := newKimiChatWorld(t, ai.ProviderMoonshotCN, nil)
	enqueue(ev, w, provider.SSE([][]byte{raw}))
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), nil))
	o.record(ev)
	if !ev.Check("actual first response replays successfully", o.err == nil && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
		return
	}
	a := o.result.Metadata.Attempts[0]
	u := o.result.Message.Usage
	reason, _ := u.Reasoning.Get()
	ev.Check("missing read is complete resolved usage", a.UsageReporting == ai.UsageComplete && a.CostEstimate.Status == ai.CostResolved && u.Input == 94 && u.Output == 196 && reason == 178 && u.CacheRead == 0 && u.CacheWrite == 0, "%+v", a)
}
