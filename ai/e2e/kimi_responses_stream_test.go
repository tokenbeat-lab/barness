package e2e

import (
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

func TestKimiResponsesTerminals(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, sc := range []struct {
			name   string
			frames []string
			code   ai.Code
			stop   ai.StopReason
			want   string
		}{
			{"terminal-only", []string{`{"type":"response.completed","response":{"status":"completed","output":[` + kimiResponsesReasoning + `,` + kimiResponsesMessage + `]}}`}, "", ai.StopReasonStop, "Hello Kimi."},
			{"done-only-text", []string{`{"type":"response.output_text.done","output_index":0,"content_index":0,"text":"Hello."}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, "", ai.StopReasonStop, "Hello."},
			{"missing-terminal", []string{`{"type":"response.output_item.done","output_index":0,"item":` + kimiResponsesMessage + `}`}, ai.CodeProtocol, "", ""},
			{"length", []string{`{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[]}}`}, "", ai.StopReasonLength, ""},
			{"filtered", []string{`{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"content_filter"},"output":[]}}`}, ai.CodeUpstreamError, "", ""},
			{"failed", []string{`{"type":"response.failed","response":{"status":"failed","error":{"code":"server_error","message":"private provider text"},"output":[]}}`}, ai.CodeUpstreamError, "", ""},
			{"error", []string{`{"type":"error","code":"server_error","message":"private provider text"}`}, ai.CodeUpstreamError, "", ""},
			{"bad-status", []string{`{"type":"response.completed","response":{"status":"in_progress","output":[]}}`}, ai.CodeProtocol, "", ""},
			{"malformed-json", []string{`{"type":`}, ai.CodeProtocol, "", ""},
			{"nonobject", []string{`[]`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, ai.CodeProtocol, "", ""},
			{"unassociated-delta", []string{`{"type":"response.output_text.delta","output_index":0,"delta":"lost"}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, ai.CodeProtocol, "", ""},
			{"bad-arguments", []string{`{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"f","call_id":"c","name":"lookup","arguments":"{\"q\":"}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, ai.CodeProtocol, "", ""},
			{"duplicate-call-id", []string{`{"type":"response.completed","response":{"status":"completed","output":[{"type":"function_call","id":"a","call_id":"same","name":"lookup","arguments":"{}"},{"type":"function_call","id":"b","call_id":"same","name":"lookup","arguments":"{}"}]}}`}, ai.CodeProtocol, "", ""},
			{"duplicate-item-id", []string{`{"type":"response.completed","response":{"status":"completed","output":[{"type":"message","id":"same","role":"assistant","content":[{"type":"output_text","text":"a"}]},{"type":"message","id":"same","role":"assistant","content":[{"type":"output_text","text":"b"}]}]}}`}, ai.CodeProtocol, "", ""},
			{"changed-tool-identity", []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"f","call_id":"a","name":"lookup","arguments":""}}`, `{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"f","call_id":"b","name":"lookup","arguments":"{}"}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, ai.CodeProtocol, "", ""},
			{"changed-terminal-kind", []string{`{"type":"response.output_item.done","output_index":0,"item":` + kimiResponsesMessage + `}`, `{"type":"response.completed","response":{"status":"completed","output":[` + kimiResponsesReasoning + `]}}`}, ai.CodeProtocol, "", ""},
			{"changed-terminal-arguments", []string{`{"type":"response.output_item.done","output_index":0,"item":` + kimiResponsesCallA + `}`, `{"type":"response.completed","response":{"status":"completed","output":[{"type":"function_call","id":"item|a","call_id":"opaque|α/a","name":"lookup","arguments":"{}"}]}}`}, ai.CodeProtocol, "", ""},
			{"event-after-terminal", []string{`{"type":"response.completed","response":{"status":"completed","output":[]}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}, ai.CodeProtocol, "", ""},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K12-terminal-")+sc.name)
				w := newKimiResponsesWorld(t, market, nil)
				b := w.host.Binding(tenantA.tenant, "kimi-responses")
				b.Retry.MaxRetries = 1
				w.host.PutBinding(b)
				var chunks [][]byte
				for _, f := range sc.frames {
					chunks = append(chunks, []byte("data: "+f+"\n\n"))
				}
				enqueue(ev, w, provider.SSE(chunks))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesToolRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if sc.code == "" {
					ev.Check("normal terminal retained", o.err == nil && o.result.Message.StopReason == sc.stop && textOf(o.result.Message) == sc.want, "%v %+v", o.err, o.result.Message)
				} else {
					ev.Check("bad terminal classified", errors.Is(o.err, &ai.Error{Code: sc.code, Phase: ai.PhaseStream}), "%v", o.err)
				}
				ev.Check("stream never retried", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%v", o.err)
			})
		}
	}
}

func TestKimiResponsesPartConsistency(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, sc := range []struct {
			name   string
			frames []string
		}{
			{"invalid-content-part", []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"m","role":"assistant","content":[]}}`, `{"type":"response.content_part.done","output_index":0,"content_index":0,"part":{"type":"image","text":"private"}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}},
			{"delta-after-arguments-done", []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"f","call_id":"c","name":"lookup","arguments":""}}`, `{"type":"response.function_call_arguments.done","output_index":0,"arguments":"{}"}`, `{"type":"response.function_call_arguments.delta","output_index":0,"delta":" "}`, `{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"f","call_id":"c","name":"lookup","arguments":"{}"}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}},
			{"item-changes-arguments-done", []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"f","call_id":"c","name":"lookup","arguments":""}}`, `{"type":"response.function_call_arguments.done","output_index":0,"arguments":"{\"q\":\"x\"}"}`, `{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"f","call_id":"c","name":"lookup","arguments":"{}"}}`, `{"type":"response.completed","response":{"status":"completed","output":[]}}`}},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K12-parts-")+sc.name)
				w := newKimiResponsesWorld(t, market, nil)
				var chunks [][]byte
				for _, raw := range sc.frames {
					chunks = append(chunks, []byte("data: "+raw+"\n\n"))
				}
				enqueue(ev, w, provider.SSE(chunks))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesToolRequest(), nil))
				o.record(ev)
				ev.Check("contradictory parts rejected", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
			})
		}
	}
}
