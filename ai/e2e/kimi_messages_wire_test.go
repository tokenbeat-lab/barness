package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestKimiMessagesWire(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K03-wire")
			w := newKimiMessagesWorld(t, market, nil)
			enqueue(ev, w, kimiMessagesReply())
			o := drain(w.client.Stream(ctxFor(t), textScope("kimi-wire"), kimiMessagesTarget, kimiMessagesRequest(), ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("request succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			r := w.provider.Requests()[0]
			f := bodyFields(t, r.Body)
			ev.Check("stable resource and version", r.Path == "/anthropic/v1/messages" && r.Query == "" && r.Header["Anthropic-Version"] == "2023-06-01", "%+v", r)
			ev.Check("single Bearer credential", r.Header["Authorization"] != "" && r.Header["X-Api-Key"] == "", "%+v", r.Header)
			ev.Check("no vendor native extensions", r.Header["Anthropic-Beta"] == "" && r.Header["Anthropic-Dangerous-Direct-Browser-Access"] == "" && len(f["thinking"]) == 0, "%+v %s", r.Header, r.Body)
			ev.Check("standalone effort", jsonEqual(f["output_config"], []byte(`{"effort":"high"}`)), "%s", r.Body)
			ev.Check("top level explicit five minute preference", jsonEqual(f["cache_control"], []byte(`{"type":"ephemeral","ttl":"5m"}`)), "%s", r.Body)
		})
	}
}
