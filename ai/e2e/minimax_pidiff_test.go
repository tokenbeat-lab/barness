package e2e

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// M3 Messages is already a frozen native route. Each side consumes the same
// independently sourced synthetic SSE via its real public entry, with no
// model patch, field pruning or replacement pi package.
func TestMiniMaxPiMessages(t *testing.T) {
	for _, entry := range []string{"stream", "streamSimple"} {
		for _, scenario := range []string{"text", "tools"} {
			t.Run(scenario+"/"+entry, func(t *testing.T) {
				o := openOracle(t)
				rec := newRecorder()
				w := newMiniMaxMessagesConfigured(t, nil, func(c *ai.Config) {
					c.Observer = rec
					c.Policy.MaxQueuedObservations = 32
				})
				srv := provider.New(map[string]string{anthropicA.secret: anthropicA.alias})
				defer srv.Close()
				req := miniMaxMessagesRequest()
				if scenario == "tools" {
					req = miniMaxToolRequest()
				}
				stages := []string{scenario}
				if scenario == "tools" {
					stages = append(stages, "tool-replay")
				}
				for _, stage := range stages {
					t.Run(stage, func(t *testing.T) {
						ev := run.Case(t, "PIDIFF-MM06-messages-"+stage+"-"+entry)
						ev.ReplayEnv(pioracle.EnableEnv + "=1")
						file := "text.sse"
						if stage == "tools" {
							file = "tools.sse"
						}
						script := miniMaxMessagesFixture(t, file, "MiniMax-M3")
						ev.Fixture("synthetic-"+file, script)
						enqueue(ev, w, provider.SSE([][]byte{script}))
						srv.Enqueue(provider.SSE([][]byte{script}))
						scope := textScope(ev.ID())
						var stream *ai.Stream
						options := map[string]any{"maxRetries": 0}
						if entry == "streamSimple" {
							simple := ai.SimpleOptions{}
							if scenario == "tools" {
								simple.Reasoning = ai.ThinkingHigh
								options["reasoning"] = "high"
							}
							stream = w.client.StreamSimple(ctxFor(t), scope, miniMaxMessagesTarget("MiniMax-M3"), req, simple)
						} else {
							var full ai.Options
							if scenario == "tools" {
								full = ai.AnthropicOptions{ThinkingEnabled: ai.Value(true)}
								options["thinkingEnabled"] = true
							}
							stream = w.client.Stream(ctxFor(t), scope, miniMaxMessagesTarget("MiniMax-M3"), req, full)
						}
						b, _ := barnessObservation(t, ev, w, stream, 0)
						res, callErr := stream.Result()
						ev.Record("observer", rec.awaitCall(ev, scope.RequestID))
						ev.Check("public MiniMax turn succeeds", callErr == nil, "%v", callErr)
						piRun, err := o.Run(ctxFor(t), pioracle.Case{API: string(ai.APIAnthropicMessages), Provider: "minimax", Model: "MiniMax-M3", BaseURL: srv.URL() + "/anthropic", APIKey: anthropicA.secret, Entry: entry, Context: piContext(t, req), Options: mustMarshal(t, options)})
						if !ev.Check("frozen MiniMax turn executes", err == nil, "%v", err) {
							return
						}
						pi, _ := piObservation(ev, srv, piRun)
						miniMaxCompare(t, ev, o, w, piRun, pi, b, script)
						wantCalls := 1
						if stage == "tool-replay" {
							wantCalls = 2
						}
						ev.Check("same bounded inference count", len(w.provider.Requests()) == wantCalls && len(srv.Requests()) == wantCalls && len(res.Metadata.Attempts) == 1, "barness=%d pi=%d", len(w.provider.Requests()), len(srv.Requests()))
						if stage != "tools" {
							ev.Check("literal text and stop agree independently", textOf(res.Message) == "Synthetic reply." && res.Message.StopReason == ai.StopReasonStop && strings.Contains(string(pi.Result), `"text":"Synthetic reply."`), "%s", pi.Result)
							return
						}
						if !ev.Check("Host receives two complete tools", res.Message.StopReason == ai.StopReasonToolUse && len(res.Message.Content) == 4, "%+v", res.Message.Content) {
							return
						}
						req.Messages = append(req.Messages, res.Message)
						for i, label := range []string{"alpha", "beta"} {
							call, err := ai.ValidateToolCall(res.Message, i+2, req.Tools)
							var args struct{ Label string }
							if !ev.Check("Host validates independent tool input", err == nil && call.Decode(&args) == nil && args.Label == label, "%v", err) {
								return
							}
							req.Messages = append(req.Messages, ai.ToolResultText(res.Message.Content[i+2].(ai.ToolCall), "Synthetic result.", false))
						}
					})
				}
			})
		}
	}
}

func miniMaxCompare(t *testing.T, ev *evidence.Case, oracle *pioracle.Oracle, w *world, run pioracle.Run, pi, barness pioracle.Observation, frames []byte) {
	t.Helper()
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", barness)
	miniMaxCheckPiWires(t, ev, pi, barness)
	checkObservationDuration(t, ev, "pi", pi)
	checkObservationDuration(t, ev, "barness", barness)
	diffs, err := pioracle.Compare(pi, barness)
	if !ev.Check("complete observations compare", err == nil, "%v", err) {
		return
	}
	verdict := loadLedger(t).Classify(string(ai.APIAnthropicMessages), ev.ID(), diffs)
	ev.Record("pidiff", oracle.Record(ev.ID(), string(ai.APIAnthropicMessages), verdict, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version(), "anthropic-sdk-go": evidence.ModuleVersion("github.com/anthropics/anthropic-sdk-go")}, NodeVersion: run.Node, ModelCatalogHash: catalogHash(t, w.client.Catalog()), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, barness.Request), PiResult: pi.Result, BarnessResult: barness.Result, Frames: frames}))
	var pending []string
	for _, f := range verdict.Findings {
		if f.Classification == pioracle.Pending {
			pending = append(pending, fmt.Sprintf("%s %s", f.Kind, f.Path))
		}
	}
	ev.Check("zero pending MiniMax differences", verdict.Pass, "%s", strings.Join(pending, "; "))
}

func TestMiniMaxPiCatalogDifferences(t *testing.T) {
	ev := run.Case(t, "PIDIFF-MM06-catalog-differences")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	o := openOracle(t)
	models, err := o.Models(ctxFor(t), []pioracle.ModelKey{{Provider: "minimax", API: "anthropic-messages", ID: "MiniMax-M3"}})
	if !ev.Check("frozen M3 metadata available", err == nil && len(models) == 1 && string(models[0]) != "null", "%v", err) {
		return
	}
	w := newMiniMaxMessagesWorld(t, nil)
	m, _ := w.client.Catalog().Lookup(ai.ProviderMiniMax, ai.APIAnthropicMessages, "MiniMax-M3")
	ev.Check("candidate metadata remains explicitly bounded", m.Name == "MiniMax-M3 anthropic-messages candidate (synthetic prices)" && m.DefaultThinkingLevel == "off" && m.MaxTokens == 131072 && jsonEqual(mustMarshal(t, m.Input), []byte(`["text"]`)), "%s", mustMarshal(t, m))
	ev.Check("complete own hard constraints stay fixed", jsonEqual(mustMarshal(t, m.Compat), []byte(`{"supportsLongCacheRetention":false,"thinkingSwitchOnly":true}`)) && jsonEqual(mustMarshal(t, m.ThinkingLevelMap), []byte(`{"off":"disabled","minimal":null,"low":null,"medium":null,"high":"high","xhigh":null,"max":null}`)), "%s", mustMarshal(t, m))
	ev.Check("complete candidate test rates stay synthetic", jsonEqual(mustMarshal(t, m.Cost), []byte(`{"input":1,"output":2,"cacheRead":0.1,"cacheWrite":1}`)), "%+v", m.Cost)
	metadata := jsonObject(t, m)
	metadata["type"], metadata["baseUrl"] = "chat", "https://api.minimax.io/anthropic"
	pi, b := pioracle.Observation{Result: models[0]}, pioracle.Observation{Result: mustMarshal(t, metadata)}
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", b)
	diffs, err := pioracle.Compare(pi, b)
	if !ev.Check("complete metadata compares", err == nil, "%v", err) {
		return
	}
	v := loadLedger(t).Classify(string(ai.APIAnthropicMessages), ev.ID(), diffs)
	ev.Record("pidiff", o.Record(ev.ID(), string(ai.APIAnthropicMessages), v, pioracle.Sides{ModelCatalogHash: catalogHash(t, w.client.Catalog()), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), PiResult: pi.Result, BarnessResult: b.Result, Frames: mustMarshal(t, models)}))
	ev.Check("catalog differences are bounded and reviewed", v.Pass, "%s", mustMarshal(t, v))
}
