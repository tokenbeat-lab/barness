package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxResponsesTerminalDiagnostics(t *testing.T) {
	for _, sc := range []struct {
		name, event, raw string
		failure          bool
	}{
		{"completed", `{"type":"response.completed","response":{"id":"actual-response","model":"MiniMax-M3","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`, "completed", false},
		{"failed-preserves-existing-usage-contract", `{"type":"response.failed","response":{"id":"actual-response","model":"MiniMax-M3","status":"failed","error":{"code":"server_error","message":"synthetic provider failure"},"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`, "failed", true},
		{"filter-known-usage", `{"type":"response.incomplete","response":{"id":"actual-response","model":"MiniMax-M3","status":"incomplete","incomplete_details":{"reason":"content_filter"},"output":[],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`, "incomplete.content_filter", true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM03-terminal-"+sc.name)
			w := newMiniMaxResponsesWorld(t, nil)
			ev.Fixture("synthetic-event.json", []byte(sc.event))
			enqueue(ev, w, minimaxResponsesReply(t, sc.event))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.failure {
				checkFailed(ev, o.result, o.err)
			} else {
				ev.Check("completed succeeds", o.err == nil, "%v", o.err)
			}
			ev.Check("actual provider model preserved", o.result.Message.ResponseModel == "MiniMax-M3", "%+v", o.result.Message)
			if sc.raw == "failed" {
				ev.Check("failed follows existing Responses ID and usage contract", o.result.Message.ResponseID == "" && len(o.result.Metadata.Attempts) == 1 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported && o.result.Message.Usage.TotalTokens == 0, "%+v", o.result)
			} else {
				ev.Check("provider response ID preserved", o.result.Message.ResponseID == "actual-response", "%+v", o.result.Message)
				ev.Check("known attempt usage remains available", len(o.result.Metadata.Attempts) == 1 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsagePartial && o.result.Message.Usage.TotalTokens == 5, "%+v", o.result.Metadata.Attempts)
			}
			ev.Check("raw terminal reason retained", o.result.Message.RawStopReason == sc.raw, "%+v", o.result.Message)
		})
	}
}
