package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxExtensionRoutes(t *testing.T) {
	raw, err := os.ReadFile("testdata/minimax-pidiff/extensions.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		ID                                            string `json:"id"`
		API                                           ai.API `json:"api"`
		Model, Kind, Source, PidiffSkip, ExpectedText string
		Events                                        []json.RawMessage `json:"events"`
		Fixture                                       string            `json:"fixture"`
	}
	mustUnmarshal(t, raw, &cases)
	if len(cases) != 5 {
		t.Fatal("expected two M3 route extensions and three M3.1 model extensions")
	}
	for _, sc := range cases {
		var empty ai.Options
		switch sc.API {
		case ai.APIOpenAICompletions:
			empty = ai.ChatOptions{}
		case ai.APIOpenAIResponses:
			empty = ai.ResponsesOptions{}
		case ai.APIAnthropicMessages:
			empty = ai.AnthropicOptions{}
		default:
			t.Fatal("unknown extension API")
		}
		for _, entry := range scenarioEntries(empty) {
			t.Run(sc.ID+"/"+entry.name, func(t *testing.T) {
				ev := run.Case(t, "MM06-extension-"+sc.ID+"-"+entry.name)
				ev.Fixture("extension-inventory.json", raw)
				ev.Record("pi-status", map[string]string{"classification": "extension", "piParity": "NOT_RUN", "pidiffSkip": sc.PidiffSkip, "source": sc.Source, "kind": sc.Kind})
				ev.Check("explicit scoped pi skip", sc.PidiffSkip != "" && sc.Source != "" && (sc.Kind == "route" || sc.Kind == "model"), "%+v", sc)
				registered, routeExtension := loadLedger(t).Route("minimax-cn", string(sc.API))
				ev.Check("exact route inventory", sc.API == ai.APIAnthropicMessages && !routeExtension || sc.API != ai.APIAnthropicMessages && routeExtension && registered.Classification == pioracle.Extension, "%+v", registered)
				var w *world
				var target ai.Target
				var req ai.Request
				var reply provider.Reply
				switch sc.API {
				case ai.APIOpenAICompletions:
					w, target, req = newMiniMaxChatWorld(t, nil), miniMaxChatTarget(sc.Model), miniMaxChatRequest()
					reply = chatEvents(t, sc.Events, "")
				case ai.APIOpenAIResponses:
					w, target, req = newMiniMaxResponsesWorld(t, nil), minimaxResponsesTarget(sc.Model), minimaxResponsesRequest()
					reply = sseEvents(t, sc.Events, provider.FramingLF)
				case ai.APIAnthropicMessages:
					w, target, req = newMiniMaxMessagesWorld(t, nil), miniMaxMessagesTarget(sc.Model), miniMaxMessagesRequest()
					script := miniMaxMessagesFixture(t, sc.Fixture, sc.Model)
					ev.Fixture("own-messages.sse", script)
					reply = provider.SSE([][]byte{script})
				}
				ev.Fixture("own-events.json", mustMarshal(t, sc.Events))
				enqueue(ev, w, reply)
				o := entry.invoke(ctxFor(t), w, textScope(ev.ID()), target, req)
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Record("catalog", w.client.Catalog())
				ev.Check("own extension is independently executed", o.err == nil && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1 && textOf(o.result.Message) == sc.ExpectedText && o.result.Message.StopReason == ai.StopReasonStop, "%v %+v", o.err, o.result.Message)
				ev.Check("actual route/model identity preserved", o.result.Message.Provider == ai.ProviderMiniMaxCN && o.result.Message.API == sc.API && o.result.Message.Model == sc.Model, "%+v", o.result.Message)
				if sc.Kind == "model" {
					ev.Check("new model cannot borrow M3 frozen identity", sc.Model == "MiniMax-M3.1-Flash-Preview" && strings.Contains(sc.PidiffSkip, "M3.1"), "%+v", sc)
				}
			})
		}
	}
}
