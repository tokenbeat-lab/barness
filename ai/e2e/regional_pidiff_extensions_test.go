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

func TestRegionalProviderExtensions(t *testing.T) {
	raw, err := os.ReadFile("testdata/regional-pidiff/extensions.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Provider                                  ai.ProviderID
		API                                       ai.API
		Models                                    []string
		Source, PidiffSkip, ExpectedText, Fixture string
		ExpectedInput, ExpectedOutput             int64
		Events                                    []json.RawMessage
	}
	mustUnmarshal(t, raw, &cases)
	if len(cases) != 4 {
		t.Fatal("exactly four missing frozen routes")
	}
	for _, sc := range cases {
		for _, model := range sc.Models {
			for _, entry := range []string{"stream", "complete", "streamSimple", "completeSimple"} {
				t.Run(string(sc.Provider)+"/"+string(sc.API)+"/"+model+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "RP06-extension-"+string(sc.Provider)+"-"+string(sc.API)+"-"+model+"-"+entry)
					ev.Fixture("own-extension-inventory.json", raw)
					l := loadLedger(t)
					registered, ok := l.Route(string(sc.Provider), string(sc.API))
					ev.Check("exact extension route registered", ok && registered.Classification == pioracle.Extension, "%+v", registered)
					ev.Check("source and explicit frozen non-route reason retained", sc.Source != "" && sc.PidiffSkip != "", "missing provenance")
					ev.Record("pi-status", map[string]string{"classification": "extension", "piParity": "NOT_RUN", "pidiffSkip": sc.PidiffSkip, "source": sc.Source})
					r := regionalFrozenRoute{provider: sc.Provider, api: sc.API, models: sc.Models}
					w, target := regionalFrozenWorld(t, r)
					target.ModelID = model
					var reply provider.Reply
					if sc.Fixture != "" {
						script, e := os.ReadFile("testdata/regional-pidiff/" + sc.Fixture)
						if e != nil {
							t.Fatal(e)
						}
						script = []byte(strings.ReplaceAll(string(script), "MODEL", model))
						ev.Fixture("own-protocol.sse", script)
						reply = provider.SSE([][]byte{script})
					} else {
						events := make([]json.RawMessage, len(sc.Events))
						for i, e := range sc.Events {
							events[i] = json.RawMessage(strings.ReplaceAll(string(e), "MODEL", model))
						}
						if sc.API == ai.APIOpenAICompletions {
							reply = chatEvents(t, events, "")
						} else {
							reply = sseEvents(t, events, provider.FramingLF)
						}
					}
					ev.Record("response-script", reply)
					w.provider.Enqueue(reply)
					scope, req := textScope(ev.ID()), ai.Request{Messages: []ai.Message{ai.UserText("Independent regional extension.")}}
					var o outcome
					switch entry {
					case "stream":
						o = drain(w.client.Stream(ctxFor(t), scope, target, req, nil))
					case "complete":
						o.result, o.err = w.client.Complete(ctxFor(t), scope, target, req, nil)
					case "streamSimple":
						o = drain(w.client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{}))
					default:
						o.result, o.err = w.client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{})
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("own-candidate-catalog", w.client.Catalog())
					ev.Check("own fixture executes correct complete identity", o.err == nil && o.result.Metadata.ProviderID == sc.Provider && o.result.Metadata.API == sc.API && o.result.Metadata.ModelID == model && textOf(o.result.Message) == sc.ExpectedText && o.result.Message.StopReason == ai.StopReasonStop, "%v %+v", o.err, o.result)
					ev.Check("scripted usage not mistaken for free", o.result.Message.Usage.Input == sc.ExpectedInput && o.result.Message.Usage.Output == sc.ExpectedOutput && o.result.Message.Usage.Cost.Total > 0, "%+v", o.result.Message.Usage)
					ev.Check("one own send not a borrowed differential", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata)
				})
			}
		}
	}
	for _, r := range regionalFrozenRoutes {
		if _, ok := loadLedger(t).Route(string(r.provider), string(r.api)); ok {
			t.Fatal("actual frozen route must never be exempt")
		}
	}
}
