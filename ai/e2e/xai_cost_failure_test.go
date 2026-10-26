package e2e

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAICostFailureSemantics(t *testing.T) {
	for _, sc := range []struct {
		name   string
		frame  json.RawMessage
		mutate func(*ai.Catalog)
		report ai.UsageReporting
		cost   ai.CostEstimateStatus
		code   ai.Code
		total  int64
	}{
		{name: "failed-usage-unreported", frame: json.RawMessage(`{"type":"response.failed","response":{"status":"failed","model":"grok-4.7","error":{"code":"server_error","message":"private"},"usage":` + xaiUsage + `}}`), report: ai.UsageUnreported, cost: ai.CostUnreported, code: ai.CodeUpstreamError},
		{name: "incomplete-retains-usage", frame: json.RawMessage(`{"type":"response.incomplete","response":{"status":"incomplete","model":"grok-4.7","incomplete_details":{"reason":"max_output_tokens"},"output":[` + xaiMessage + `],"usage":` + xaiUsage + `}}`), report: ai.UsageComplete, cost: ai.CostResolved, total: 41},
		{name: "unsupported-tier", frame: json.RawMessage(strings.Replace(string(xaiTerminal(xaiMessage, xaiUsage)), `"status":"completed"`, `"status":"completed","service_tier":"priority"`, 1)), report: ai.UsageComplete, cost: ai.CostPartial, code: ai.CodeProtocol, total: 41},
		{name: "zero-rates-unknown-cache", frame: xaiTerminal(xaiMessage, `{"input_tokens":32,"total_tokens":151}`), mutate: func(c *ai.Catalog) { c.Models[0].Cost = ai.ModelCost{} }, report: ai.UsagePartial, cost: ai.CostResolved, total: 151},
		{name: "finite-host-rate-overflow", frame: xaiTerminal(xaiMessage, `{"input_tokens":9223372036854775807,"total_tokens":9223372036854775807,"input_tokens_details":{"cached_tokens":0}}`), mutate: func(c *ai.Catalog) { c.Models[0].Cost = ai.ModelCost{CostRates: ai.CostRates{Input: math.MaxFloat64}} }, report: ai.UsageUnreported, cost: ai.CostUnreported, code: ai.CodeProtocol},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E12-failure-"+sc.name)
			w := newXAIWorld(t, sc.mutate)
			enqueue(ev, w, xaiReply(t, sc.frame))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("frames", []json.RawMessage{sc.frame})
			if sc.code != "" {
				ev.Check("failure remains explicit", errors.Is(o.err, &ai.Error{Code: sc.code}), "%v", o.err)
			} else {
				ev.Check("valid uncertain estimate accepted", o.err == nil, "%v", o.err)
			}
			if len(o.result.Metadata.Attempts) == 1 {
				a := o.result.Metadata.Attempts[0]
				ev.Check("failure and cost completeness preserve their own semantics", a.UsageReporting == sc.report && a.CostEstimate.Status == sc.cost && a.Usage.TotalTokens == sc.total && o.result.Message.Usage == a.Usage, "%+v", a)
			}
		})
	}
}

func TestXAIHTTPFailures(t *testing.T) {
	for _, status := range []int{400, 401, 403, 408, 409, 429, 500} {
		t.Run(jsonNumber(status), func(t *testing.T) {
			ev := run.Case(t, "XAI-E10-http-"+jsonNumber(status))
			w := newXAIWorld(t, nil)
			reply := provider.Reply{Status: status, Header: map[string]string{"Content-Type": "application/json", "X-Request-Id": "unverified"}, Chunks: [][]byte{[]byte(`{"error":{"message":"private text and private-cipher","code":"private-code"}}`)}}
			enqueue(ev, w, reply)
			res, err := w.client.Complete(ctxFor(t), textScope("xai-http-"+jsonNumber(status)), xaiTarget(), xaiRequest(), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("reply", reply)
			code := ai.CodeInvalidRequest
			if status == 401 || status == 403 {
				code = ai.CodeUpstreamAuth
			} else if status == 429 {
				code = ai.CodeRateLimited
			} else if status == 408 || status == 409 || status >= 500 {
				code = ai.CodeUpstreamError
			}
			var ae *ai.Error
			ev.Check("HTTP status classified without private echo", errors.Is(err, &ai.Error{Code: code}) && errors.As(err, &ae) && ae.HTTPStatus == status && ae.ProviderRequestID == "" && !strings.Contains(err.Error(), "private"), "%v", err)
			if len(res.Metadata.Attempts) == 1 {
				a := res.Metadata.Attempts[0]
				ev.Check("unaccepted error usage stays unreported", a.UsageReporting == ai.UsageUnreported && a.CostEstimate.Status == ai.CostUnreported, "%+v", a)
			}
		})
	}
}
func jsonNumber(n int) string { b, _ := json.Marshal(n); return string(b) }
