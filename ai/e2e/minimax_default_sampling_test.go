package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// This traces the shared sampling contract through existing public Client
// routes. MiniMax-specific wire constraints belong to its protocol suites.
func TestMiniMaxDefaultSampling(t *testing.T) {
	for _, p := range []struct {
		proto *fixtureProtocol
		empty ai.Options
	}{{responsesProtocol, ai.ResponsesOptions{}}, {chatProtocol, ai.ChatOptions{}}} {
		for _, label := range []string{"", "high", "max"} {
			for _, entry := range scenarioEntries(p.empty) {
				t.Run(p.proto.dir+"/"+label+"/"+entry.name, func(t *testing.T) {
					ev := run.Case(t, "MM01-sampling-"+p.proto.dir+"-"+label+"-"+entry.name)
					fixture, raw := loadFixture(t, p.proto, "thinking-sampling.json")
					a := scenarioText{sc: fixture.Scenarios[0], raw: raw}
					ev.Fixture("synthetic-existing-protocol.json", a.raw)
					w := newWorldWith(t, func(c *ai.Config) {
						cat := ai.BuiltinCatalog()
						for i := range cat.Models {
							m := &cat.Models[i]
							if m.Provider != p.proto.provider || m.API != p.proto.api || m.ID != a.sc.Model {
								continue
							}
							*m = patchModel(*m, json.RawMessage(`{"reasoning":true,"defaultThinkingLevel":"`+label+`","thinkingLevelMap":{"minimal":null,"low":null,"medium":null,"high":"high","xhigh":null,"max":"max"},"samplingParamsByThinkingLevel":{"off":{"temperature":0.1},"high":{"temperature":0.8},"max":{"temperature":0.9}}}`))
						}
						c.Catalog = &cat
					}, tenantA)
					enqueue(ev, w, a.reply(t))
					o := entry.invoke(ctxFor(t), w, textScope("default-sampling"), a.target(), a.request(t))
					o.record(ev)
					ev.Record("catalog", w.client.Catalog())
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("successful single attempt", o.err == nil && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
						return
					}
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					want := map[string]string{"": "0.1", "high": "0.8", "max": "0.9"}[label]
					ev.Check("omission uses declared sampling label", string(fields["temperature"]) == want, "%s", w.provider.Requests()[0].Body)
					_, effort := fields["reasoning_effort"]
					// Native OpenAI Responses already sends none for omission;
					// the sampling label must never replace that with high/max.
					wirePreserved := len(fields["reasoning"]) == 0
					if p.proto.api == ai.APIOpenAIResponses {
						wirePreserved = jsonEqual(fields["reasoning"], json.RawMessage(`{"effort":"none"}`))
					}
					ev.Check("label preserves existing omitted wire", !effort && wirePreserved, "%s", w.provider.Requests()[0].Body)
				})
			}
		}
	}
}
