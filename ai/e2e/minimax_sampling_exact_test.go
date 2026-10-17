package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxFinalSamplingExactRanges(t *testing.T) {
	for _, route := range miniMaxSharedRoutes {
		for _, sample := range []struct {
			name, field, value string
			valid              bool
		}{
			{"temperature-upper", "temperature", "1.00000000000000001", false},
			{"top-p-upper", "top_p", "1.00000000000000001", false},
			{"tiny-positive", "top_p", "1e-100000000000000000000", true},
			{"tiny-negative", "top_p", "-1e-100000000000000000000", false},
			{"zero", "top_p", "0", true},
		} {
			t.Run(route.name+"/"+sample.name, func(t *testing.T) {
				ev := run.Case(t, "MM10-exact-sampling-"+route.name+"-"+sample.name)
				w := newMiniMaxSharedWorld(t, nil)
				value, valid := sample.value, sample.valid
				if sample.name == "temperature-upper" && route.name != "responses" {
					value = "2.00000000000000001"
				}
				if sample.name == "zero" && route.name == "responses" {
					valid = false
				}
				req, _, reply, _ := route.script(t, "MiniMax-M3")
				w.provider.Enqueue(reply)
				hooked := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					p.Body[sample.field] = json.Number(value)
					return ai.KeepPayload(), nil
				}})
				result, err := hooked.Complete(ctxFor(t), textScope(t.Name()), route.target("MiniMax-M3"), req, route.options())
				ev.Record("result", result)
				ev.Record("error", errString(err))
				ev.Record("requests", w.provider.Requests())
				ev.Check("exact range survives final wire", (err == nil) == valid, "value=%s err=%v", value, err)
				if valid {
					ev.Check("single final request", len(w.provider.Requests()) == 1, "requests=%d", len(w.provider.Requests()))
				} else {
					ev.Check("invalid precise number sends zero Attempt", len(w.provider.Requests()) == 0 && len(result.Metadata.Attempts) == 0, "%+v", result.Metadata)
				}
			})
		}
	}
}

func TestMiniMaxResponsesSamplingMapExactRanges(t *testing.T) {
	for _, field := range []string{"temperature", "top_p"} {
		for _, value := range []string{"1.00000000000000001", "1e-100000000000000000000"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				ev := run.Case(t, "MM10-exact-sampling-map-"+field+"-"+value)
				w := newMiniMaxResponsesWorld(t, nil)
				enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), ai.ResponsesOptions{SamplingParams: map[string]json.RawMessage{field: json.RawMessage(value)}}))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				valid := value != "1.00000000000000001"
				ev.Check("raw sampling merge preserves exact endpoints", (o.err == nil) == valid, "%v", o.err)
				if !valid {
					ev.Check("no invalid raw sampling Attempt", len(o.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "%+v", o.result.Metadata)
				}
			})
		}
	}
}
