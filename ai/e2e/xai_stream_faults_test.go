package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIStreamFaults(t *testing.T) {
	added := `{"type":"response.output_item.added","output_index":0,"item":{"id":"message|α","type":"message","role":"assistant","content":[]}}`
	done := `{"type":"response.output_item.done","output_index":0,"item":` + xaiMessage + `}`
	final := string(xaiTerminal(xaiMessage, xaiUsage))
	for _, sc := range []struct {
		name   string
		frames []string
		code   ai.Code
		stop   ai.StopReason
	}{
		{name: "duplicate-done", frames: []string{done, done, final}, stop: ai.StopReasonStop},
		{name: "no-terminal", frames: []string{done}, code: ai.CodeProtocol},
		{name: "done-marker-only", frames: []string{`[DONE]`}, code: ai.CodeProtocol},
		{name: "partial-then-done-marker", frames: []string{added, `{"type":"response.output_text.delta","output_index":0,"item_id":"message|α","content_index":0,"delta":"Hello"}`, `[DONE]`}, code: ai.CodeProtocol},
		{name: "orphan-delta", frames: []string{`{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":0,"delta":"lost"}`, final}, code: ai.CodeProtocol},
		{name: "wrong-id", frames: []string{added, `{"type":"response.output_text.delta","output_index":0,"item_id":"wrong","content_index":0,"delta":"lost"}`, final}, code: ai.CodeProtocol},
		{name: "fraction-index", frames: []string{added, `{"type":"response.output_text.delta","output_index":0.5,"item_id":"message|α","content_index":0,"delta":"lost"}`}, code: ai.CodeProtocol},
		{name: "null-index", frames: []string{`{"type":"response.output_item.done","output_index":null,"item":` + xaiMessage + `}`}, code: ai.CodeProtocol},
		{name: "missing-part-index", frames: []string{added, `{"type":"response.output_text.delta","output_index":0,"item_id":"message|α","delta":"lost"}`}, code: ai.CodeProtocol},
		{name: "changed-completed", frames: []string{done, strings.Replace(final, "Hello Grok.", "Changed", 1)}, code: ai.CodeProtocol},
		{name: "dropped-completed", frames: []string{done, string(xaiTerminal("", xaiUsage))}, code: ai.CodeProtocol},
		{name: "changed-prefix", frames: []string{added, `{"type":"response.output_text.delta","output_index":0,"item_id":"message|α","content_index":0,"delta":"Prefix"}`, done}, code: ai.CodeProtocol},
		{name: "conflicting-terminal", frames: []string{final, strings.Replace(final, "Hello Grok.", "Changed", 1)}, code: ai.CodeProtocol},
		{name: "after-terminal", frames: []string{final, added}, code: ai.CodeProtocol},
		{name: "duplicate-item-id", frames: []string{string(xaiTerminal(xaiMessage+`,`+xaiMessage, xaiUsage))}, code: ai.CodeProtocol},
		{name: "duplicate-call-id", frames: []string{string(xaiTerminal(xaiCallA+`,`+strings.Replace(xaiCallB, "call|β", "call|α", 1), xaiUsage))}, code: ai.CodeProtocol},
		{name: "unknown-output", frames: []string{string(xaiTerminal(`{"type":"web_search_call","id":"w"}`, xaiUsage))}, code: ai.CodeProtocol},
		{name: "unknown-status", frames: []string{`{"type":"response.completed","response":{"status":"in_progress","output":[]}}`}, code: ai.CodeProtocol},
		{name: "wrong-model", frames: []string{strings.Replace(final, `"grok-4.7"`, `"grok-4-other"`, 1)}, code: ai.CodeProtocol},
		{name: "bad-json", frames: []string{`{"type":`}, code: ai.CodeProtocol},
		{name: "nonobject", frames: []string{`[]`}, code: ai.CodeProtocol},
		{name: "duplicate-key", frames: []string{`{"type":"response.completed","type":"response.created","response":{}}`}, code: ai.CodeProtocol},
		{name: "length", frames: []string{`{"type":"response.incomplete","response":{"status":"incomplete","model":"grok-4.7","incomplete_details":{"reason":"max_output_tokens"},"output":[` + xaiMessage + `],"usage":` + xaiUsage + `}}`}, stop: ai.StopReasonLength},
		{name: "filter", frames: []string{`{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"content_filter"},"output":[]}}`}, code: ai.CodeUpstreamError},
		{name: "failed", frames: []string{`{"type":"response.failed","response":{"status":"failed","error":{"code":"private-code","message":"private-cipher"},"usage":` + xaiUsage + `}}`}, code: ai.CodeUpstreamError},
		{name: "error", frames: []string{`{"type":"error","code":"private-code","message":"private-cipher"}`}, code: ai.CodeUpstreamError},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E10-"+sc.name)
			w := newXAIWorld(t, nil)
			b := w.host.Binding(tenantA.tenant, "xai-responses")
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			var chunks [][]byte
			for _, frame := range sc.frames {
				chunks = append(chunks, []byte("data: "+frame+"\n\n"))
			}
			reply := provider.SSE(chunks)
			reply.Header = map[string]string{"Content-Type": "text/event-stream", "X-Request-Id": "synthetic-xai-stream-request-id"}
			enqueue(ev, w, reply)
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiToolRequest(), nil))
			o.record(ev)
			ev.Record("frames", sc.frames)
			ev.Record("requests", w.provider.Requests())
			if sc.code == "" {
				ev.Check("valid terminal", o.err == nil && o.result.Message.StopReason == sc.stop, "%v %+v", o.err, o.result.Message)
			} else {
				ev.Check("fault is explicit and safe", errors.Is(o.err, &ai.Error{Code: sc.code, Phase: ai.PhaseStream}) && !strings.Contains(o.err.Error(), "private"), "%v", o.err)
				var ae *ai.Error
				if errors.As(o.err, &ae) {
					ev.Check("received HTTP status and request ID survive stream failure", ae.HTTPStatus == 200 && ae.ProviderRequestID == "synthetic-xai-stream-request-id", "%+v", ae)
				}
			}
			if len(o.result.Metadata.Attempts) == 1 {
				a := o.result.Metadata.Attempts[0]
				ev.Check("attempt keeps status and received request ID", a.HTTPStatus == 200 && a.ProviderRequestID == "synthetic-xai-stream-request-id", "%+v", a)
			}
			ev.Check("stream never retried", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata)
		})
	}
}
