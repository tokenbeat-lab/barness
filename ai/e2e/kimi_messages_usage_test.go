package e2e

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiMessagesUsage(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, c := range []struct {
			name, usage string
			report      ai.UsageReporting
			status      ai.CostEstimateStatus
			long        int64
		}{
			{"unknown-ttl", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`, ai.UsageComplete, ai.CostBounded, -1},
			{"short", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80,"cache_creation":{"ephemeral_5m_input_tokens":80,"ephemeral_1h_input_tokens":0}}`, ai.UsageComplete, ai.CostResolved, 0},
			{"long", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80,"cache_creation":{"ephemeral_5m_input_tokens":0,"ephemeral_1h_input_tokens":80}}`, ai.UsageComplete, ai.CostResolved, 80},
			{"mixed", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80,"cache_creation":{"ephemeral_5m_input_tokens":30,"ephemeral_1h_input_tokens":50}}`, ai.UsageComplete, ai.CostResolved, 50},
			{"partial-ttl", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80,"cache_creation":{"ephemeral_1h_input_tokens":50}}`, ai.UsageComplete, ai.CostBounded, 50},
			{"missing-read", `{"input_tokens":100,"cache_creation_input_tokens":0}`, ai.UsagePartial, ai.CostPartial, -1},
			{"null-read", `{"input_tokens":100,"cache_read_input_tokens":null,"cache_creation_input_tokens":0}`, ai.UsagePartial, ai.CostPartial, -1},
			{"missing-write", `{"input_tokens":100,"cache_read_input_tokens":20}`, ai.UsagePartial, ai.CostPartial, -1},
			{"null-write", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":null}`, ai.UsagePartial, ai.CostPartial, -1},
			{"zero-write", `{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":0}`, ai.UsageComplete, ai.CostResolved, -1},
			{"integer-exponent", `{"input_tokens":1e2,"cache_read_input_tokens":2e1,"cache_creation_input_tokens":8e1}`, ai.UsageComplete, ai.CostBounded, -1},
		} {
			t.Run(string(market)+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K10-"+c.name)
				w := newKimiMessagesWorld(t, market, nil)
				enqueue(ev, w, provider.SSE([][]byte{[]byte(kimiMessagesScript(c.usage))}))
				o := drain(w.client.Stream(ctxFor(t), textScope("kimi-usage"), kimiMessagesTarget, kimiMessagesRequest(), nil))
				o.record(ev)
				if !ev.Check("valid report succeeds", o.err == nil && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
					return
				}
				a := o.result.Metadata.Attempts[0]
				u := o.result.Message.Usage
				ev.Check("reported counts and cost certainty independent", a.UsageReporting == c.report && a.CostEstimate.Status == c.status, "%+v", a)
				wantRead := int64(20)
				if c.name == "missing-read" || c.name == "null-read" {
					wantRead = 0
				}
				ev.Check("Messages input excludes cache and reasoning is optional", u.Input == 100 && u.CacheRead == wantRead && u.Output == 50 && u.Reasoning.IsZero() && u == a.Usage, "%+v", u)
				long, present := u.CacheWrite1h.Get()
				ev.Check("only actual TTL observed", present == (c.long >= 0) && (!present || long == c.long), "%+v", u)
				if market == ai.ProviderMoonshot && c.name == "unknown-ttl" {
					lo, _ := a.CostEstimate.Minimum.Get()
					hi, _ := a.CostEstimate.Maximum.Get()
					ev.Check("spec worked cost interval", math.Abs(lo.Total-.001296) < 1e-15 && math.Abs(hi.Total-.001536) < 1e-15 && u.Cost == lo, "%+v", a.CostEstimate)
				}
			})
		}
	}
}

func TestKimiMessagesInvalidUsage(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, c := range []struct{ name, usage string }{
			{"negative", `{"output_tokens":-1}`},
			{"fraction", `{"output_tokens":1.00000000000000001}`},
			{"overflow", `{"output_tokens":9223372036854775808}`},
			{"sum-overflow", `{"input_tokens":9223372036854775807,"output_tokens":1}`},
			{"reasoning-exceeds-output", `{"output_tokens":50,"output_tokens_details":{"thinking_tokens":51}}`},
			{"negative-ttl", `{"output_tokens":50,"cache_creation":{"ephemeral_1h_input_tokens":-1}}`},
			{"ttl-exceeds-write", `{"output_tokens":50,"cache_creation":{"ephemeral_1h_input_tokens":81}}`},
			{"ttl-gap", `{"output_tokens":50,"cache_creation":{"ephemeral_5m_input_tokens":30,"ephemeral_1h_input_tokens":40}}`},
			{"total-conflict", `{"output_tokens":50,"total_tokens":249}`},
		} {
			t.Run(string(market)+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K10-invalid-"+c.name)
				w := newKimiMessagesWorld(t, market, nil)
				script := strings.ReplaceAll(kimiMessagesScript(`{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`), `"usage":{"output_tokens":50}`, `"usage":`+c.usage)
				enqueue(ev, w, provider.SSE([][]byte{[]byte(script)}))
				o := drain(w.client.Stream(ctxFor(t), textScope("kimi-invalid-usage"), kimiMessagesTarget, kimiMessagesRequest(), nil))
				o.record(ev)
				ev.Check("invalid counter never appears as successful use", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				ev.Check("previous lawful report retained", o.result.Message.Usage.Input == 100 && o.result.Message.Usage.Output == 0 && o.result.Message.Usage.CacheWrite == 80, "%+v", o.result.Message.Usage)
			})
		}
	}
}
