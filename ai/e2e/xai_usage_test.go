package e2e

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIUsageAndCost(t *testing.T) {
	for _, sc := range []struct {
		name, usage                 string
		report                      ai.UsageReporting
		cost                        ai.CostEstimateStatus
		input, cache, output, total int64
		reason                      *int64
		lo, hi                      float64
		invalid                     bool
	}{
		{name: "exclusive-reasoning", usage: `{"input_tokens":32,"output_tokens":9,"total_tokens":151,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsageComplete, cost: ai.CostResolved, input: 24, cache: 8, output: 119, total: 151, reason: xaiI64(110), lo: .000766, hi: .000766},
		{name: "inclusive-reasoning", usage: `{"input_tokens":32,"output_tokens":119,"total_tokens":151,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsageComplete, cost: ai.CostResolved, input: 24, cache: 8, output: 119, total: 151, reason: xaiI64(110), lo: .000766, hi: .000766},
		{name: "unknown-cache", usage: `{"input_tokens":32,"output_tokens":9,"total_tokens":151,"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsagePartial, cost: ai.CostBounded, input: 32, output: 119, total: 151, reason: xaiI64(110), lo: .000730, hi: .000778},
		{name: "cache-null", usage: `{"input_tokens":32,"output_tokens":9,"total_tokens":151,"input_tokens_details":{"cached_tokens":null},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsagePartial, cost: ai.CostBounded, input: 32, output: 119, total: 151, reason: xaiI64(110), lo: .000730, hi: .000778},
		{name: "missing-reasoning", usage: `{"input_tokens":32,"output_tokens":9,"total_tokens":151,"input_tokens_details":{"cached_tokens":8}}`, report: ai.UsagePartial, cost: ai.CostResolved, input: 24, cache: 8, output: 119, total: 151, lo: .000766, hi: .000766},
		{name: "missing-output", usage: `{"input_tokens":32,"total_tokens":151,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsagePartial, cost: ai.CostResolved, input: 24, cache: 8, output: 119, total: 151, reason: xaiI64(110), lo: .000766, hi: .000766},
		{name: "missing-input", usage: `{"output_tokens":9,"total_tokens":151,"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsagePartial, cost: ai.CostPartial, output: 110, total: 151, reason: xaiI64(110)},
		{name: "missing-total", usage: `{"input_tokens":32,"output_tokens":9,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsagePartial, cost: ai.CostPartial, input: 24, cache: 8, output: 110, reason: xaiI64(110)},
		{name: "zero", usage: `{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, report: ai.UsageComplete, cost: ai.CostResolved, reason: xaiI64(0)},
		{name: "exact-exponent", usage: `{"input_tokens":3.2e1,"output_tokens":9.0,"total_tokens":151,"input_tokens_details":{"cached_tokens":8},"output_tokens_details":{"reasoning_tokens":110}}`, report: ai.UsageComplete, cost: ai.CostResolved, input: 24, cache: 8, output: 119, total: 151, reason: xaiI64(110), lo: .000766, hi: .000766},
		{name: "unreported", report: ai.UsageUnreported, cost: ai.CostUnreported},
		{name: "null", usage: `null`, report: ai.UsageUnreported, cost: ai.CostUnreported},
		{name: "empty", usage: `{}`, report: ai.UsageUnreported, cost: ai.CostUnreported},
		{name: "details-null", usage: `{"input_tokens_details":null,"output_tokens_details":null}`, report: ai.UsageUnreported, cost: ai.CostUnreported},
		{name: "fraction", usage: `{"input_tokens":1.00000000000000001}`, invalid: true},
		{name: "negative", usage: `{"input_tokens":-1}`, invalid: true},
		{name: "boolean", usage: `{"input_tokens":true}`, invalid: true},
		{name: "string", usage: `{"input_tokens":"32"}`, invalid: true},
		{name: "overflow", usage: `{"input_tokens":9223372036854775808}`, invalid: true},
		{name: "adversarial-exponent", usage: `{"input_tokens":1e999999999}`, invalid: true},
		{name: "cache-over-input", usage: `{"input_tokens":2,"input_tokens_details":{"cached_tokens":3}}`, invalid: true},
		{name: "total-under-input", usage: `{"input_tokens":32,"total_tokens":31}`, invalid: true},
		{name: "reasoning-over-all-output", usage: `{"input_tokens":32,"total_tokens":151,"output_tokens_details":{"reasoning_tokens":120}}`, invalid: true},
		{name: "third-arithmetic", usage: `{"input_tokens":32,"total_tokens":151,"output_tokens":9,"output_tokens_details":{"reasoning_tokens":100}}`, invalid: true},
		{name: "output-over-all-output", usage: `{"input_tokens":32,"total_tokens":151,"output_tokens":120}`, invalid: true},
		{name: "positive-cache-write", usage: `{"input_tokens":32,"input_tokens_details":{"cache_write_tokens":1}}`, invalid: true},
		{name: "duplicate-count", usage: `{"input_tokens":32,"input_tokens":2}`, invalid: true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E11-"+sc.name)
			rec := newRecorder()
			w := newXAIWorld(t, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
			frame := xaiTerminal(xaiMessage, sc.usage)
			enqueue(ev, w, xaiReply(t, frame))
			scope := textScope("xai-usage-" + sc.name)
			o := drain(w.client.Stream(ctxFor(t), scope, xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("frames", []json.RawMessage{frame})
			obs := rec.awaitCall(ev, scope.RequestID)
			ev.Record("observations", obs)
			if sc.invalid {
				ev.Check("illegal raw counts fail explicitly", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				return
			}
			if !ev.Check("usage accepted", o.err == nil && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
				return
			}
			a := o.result.Metadata.Attempts[0]
			u := a.Usage
			ev.Check("normalized counts and independent certainty", a.UsageReporting == sc.report && a.CostEstimate.Status == sc.cost && u.Input == sc.input && u.CacheRead == sc.cache && u.Output == sc.output && u.TotalTokens == sc.total, "%+v", a)
			r, has := u.Reasoning.Get()
			ev.Check("reasoning presence stays explicit", (sc.reason == nil && !has) || (sc.reason != nil && has && r == *sc.reason), "%+v", u.Reasoning)
			lo, lok := a.CostEstimate.Minimum.Get()
			hi, hok := a.CostEstimate.Maximum.Get()
			resolved := sc.cost == ai.CostResolved || sc.cost == ai.CostBounded
			ev.Check("worked USD range", lok == resolved && hok == resolved && (!resolved || math.Abs(lo.Total-sc.lo) < 1e-12 && math.Abs(hi.Total-sc.hi) < 1e-12 && u.Cost == lo), "%+v", a.CostEstimate)
			ev.Check("message and attempt usage agree", o.result.Message.Usage == a.Usage, "%+v", o.result)
			if len(obs) == 4 {
				ev.Check("observer freezes the same cost and usage", obs[1].Attempt.CostEstimate.Status == ai.CostUnreported && obs[2].Attempt.CostEstimate == a.CostEstimate && obs[2].Attempt.Usage == a.Usage && obs[3].Call.Attempts[0].CostEstimate == a.CostEstimate, "%+v", obs)
			}
		})
	}
}
func xaiI64(n int64) *int64 { return &n }

func TestXAILongContextCost(t *testing.T) {
	for _, sc := range []struct {
		name, usage string
		want        float64
	}{
		{"below", `{"input_tokens":199999,"output_tokens":1000,"total_tokens":200999,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, .405998},
		{"at", `{"input_tokens":200000,"output_tokens":1000,"total_tokens":201000,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, .812000},
		{"above", `{"input_tokens":200001,"output_tokens":1000,"total_tokens":201001,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, .812004},
		{"cached-threshold", `{"input_tokens":200000,"output_tokens":1000,"total_tokens":201000,"input_tokens_details":{"cached_tokens":200000},"output_tokens_details":{"reasoning_tokens":0}}`, .212000},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E12-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t, xaiTerminal(xaiMessage, sc.usage)))
			res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
			ev.Record("result", res)
			ev.Record("usage", json.RawMessage(sc.usage))
			ev.Check("whole-request rates use inclusive prompt threshold", err == nil && math.Abs(res.Message.Usage.Cost.Total-sc.want) < 1e-12 && res.Metadata.Attempts[0].CostEstimate.Status == ai.CostResolved, "%v %+v", err, res)
		})
	}
}
