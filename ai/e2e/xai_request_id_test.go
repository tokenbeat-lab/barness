package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIRequestIDAttribution(t *testing.T) {
	for _, entry := range outcomeEntries {
		for _, mode := range []string{"success", "http-401", "http-429", "protocol"} {
			for _, present := range []bool{true, false} {
				header := "absent"
				if present {
					header = "present"
				}
				name := entry.name + "/" + mode + "/" + header
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "XAI-E14-request-id-"+entry.name+"-"+mode+"-"+header)
					rec := newRecorder()
					w := newXAIWorld(t, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
					reply := xaiReply(t)
					code, phase, status := ai.Code(""), ai.Phase(""), 200
					switch mode {
					case "http-401", "http-429":
						status, code, phase = 401, ai.CodeUpstreamAuth, ai.PhaseRequest
						if mode == "http-429" {
							status, code = 429, ai.CodeRateLimited
						}
						body := `{"error":{"message":"private-body-and-ciphertext"},"id":"body-response-id"}`
						reply = provider.Reply{Status: status, Header: map[string]string{"Content-Type": "application/json"}, Script: body, Chunks: [][]byte{[]byte(body)}}
					case "protocol":
						code, phase = ai.CodeProtocol, ai.PhaseStream
						reply = provider.SSE([][]byte{[]byte("data: {\"type\":\"response.output_text.delta\",\"output_index\":0,\"item_id\":\"orphan\",\"content_index\":0,\"delta\":\"private-body-and-ciphertext\"}\n\n")})
					}
					reply.Header["Request-Id"] = "wrong-request-id-header"
					reply.Header["X-Ds-Trace-Id"] = "wrong-deepseek-header"
					reply.Header["Traceparent"] = "wrong-traceparent-header"
					want := ""
					if present {
						want = "synthetic-xai-response-request-id"
						reply.Header["X-Request-Id"] = want
					}
					enqueue(ev, w, reply)
					scope := textScope("host-request-id-" + strings.ReplaceAll(name, "/", "-"))
					o := entry.invoke(ctxFor(t), w, scope, xaiTarget(), xaiRequest())
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("response-header-source", map[string]string{"header": "x-request-id", "syntheticValue": want})
					if code == "" {
						ev.Check("successful response remains successful", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop, "%v", o.err)
					} else {
						var ae *ai.Error
						ev.Check("failure retains classification and response header ID", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}) && errors.As(o.err, &ae) && ae.HTTPStatus == status && ae.ProviderRequestID == want, "%v", o.err)
						ev.Check("provider body stays outside diagnostics", !strings.Contains(errString(o.err), "private-"), "%v", o.err)
					}
					if ev.Check("one actual attempt retains only the selected header ID", len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "attempts=%d requests=%d", len(o.result.Metadata.Attempts), len(w.provider.Requests())) {
						a := o.result.Metadata.Attempts[0]
						ev.Check("Attempt attributes vendor ID without replacing logical identity", a.HTTPStatus == status && a.ProviderRequestID == want && o.result.Metadata.RequestID == scope.RequestID && a.AttemptID == scope.RequestID+"#1", "%+v", a)
					}
					obs := rec.awaitCall(ev, scope.RequestID)
					checkCallRecords(ev, obs, scope, xaiTarget().BindingID, o, 1)
					if len(obs) > 0 {
						fin := obs[len(obs)-1]
						ev.Check("terminal Observer retains response header ID", len(fin.Call.Attempts) == 1 && fin.Call.Attempts[0].ProviderRequestID == want, "%+v", fin.Call.Attempts)
						if code != "" {
							ev.Check("terminal observed error retains response header ID", fin.Error != nil && fin.Error.ProviderRequestID == want, "%+v", fin.Error)
						}
					}
					ev.Check("Observer excludes provider body", !strings.Contains(string(mustMarshal(t, obs)), "private-"), "body reached Observer")
				})
			}
		}
	}
}

func TestXAIRequestIDExistingProviders(t *testing.T) {
	for _, sc := range []struct{ name, binding, model, want string }{
		{"openai", "primary", "gpt-4.1-mini", "synthetic-openai-request-id"},
		{"deepseek", "deepseek", "deepseek-flash", "synthetic-deepseek-request-id"},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E14-request-id-existing-"+sc.name)
			w := newWorld(t, tenantA)
			frames := xaiFrames()
			for i, frame := range frames {
				frames[i] = []byte(strings.ReplaceAll(string(frame), "grok-4.7", sc.model))
			}
			reply := xaiReply(t, frames...)
			reply.Header["X-Request-Id"] = "synthetic-openai-request-id"
			reply.Header["X-Ds-Trace-Id"] = "synthetic-deepseek-request-id"
			enqueue(ev, w, reply)
			res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), ai.Target{BindingID: sc.binding, ModelID: sc.model}, xaiRequest(), nil)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Check("existing Provider keeps its own request ID source", err == nil && len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].ProviderRequestID == sc.want, "%v %+v", err, res.Metadata.Attempts)
		})
	}
}
