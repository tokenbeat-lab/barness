package e2e

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"strings"
	"testing"
)

func TestMiniMaxSharedFailures(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, sc := range []struct {
			name                  string
			status, retries, want int
		}{{"bad", 400, 2, 1}, {"auth", 401, 2, 1}, {"denied", 403, 2, 1}, {"default429", 429, 0, 1}, {"retry429", 429, 1, 2}, {"retry500", 500, 1, 2}, {"retry503", 503, 1, 2}} {
			t.Run(r.name+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, "MM05-http-"+r.name+sc.name)
				w := newMiniMaxSharedWorld(t, nil)
				miniMaxSharedRetries(w, r, miniMaxSharedA, sc.retries)
				private := "synthetic-private-prompt-and-reasoning"
				body := fmt.Sprintf(`{"error":{"type":"unclassified","message":%q},"base_resp":{"status_code":1008}}`, private)
				bad := provider.Reply{Status: sc.status, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1", "x-should-retry": "true", "X-Request-Id": "synthetic-actual-id", "Request-Id": "synthetic-actual-id"}, Chunks: [][]byte{[]byte(body)}}
				req, first, _, raw := r.script(t, "MiniMax-M3")
				ev.Fixture("synthetic-source.json", raw)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedA), bad, first)
				o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-error-"+r.name+sc.name), r.target("MiniMax-M3"), req, r.options()))
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				ev.Record("error-response", bad)
				ev.Check("only Host bounded retry applies", len(w.provider.Requests()) == sc.want && len(o.result.Metadata.Attempts) == sc.want, "%v %+v", o.err, o.result.Metadata)
				if sc.want == 1 {
					ev.Check("private freeform errors stay out of caller and Observer", o.err != nil && !strings.Contains(o.err.Error(), private) && !strings.Contains(o.result.Message.ErrorMessage, private), "%v", o.err)
					ev.Check("actual diagnostic header preserved", o.result.Metadata.Attempts[0].ProviderRequestID == "synthetic-actual-id", "%+v", o.result.Metadata.Attempts)
				} else {
					ev.Check("retry can succeed without changing route", o.err == nil && o.result.Metadata.ProviderID == ai.ProviderMiniMaxCN, "%v", o.err)
				}
				miniMaxSharedOneTerminal(ev, o)
				w.released(ev)
			})
		}
	}
}

func TestMiniMaxSharedStreamErrorPrivacy(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		t.Run(r.name, func(t *testing.T) {
			ev := run.Case(t, "MM05-stream-error-"+r.name)
			w := newMiniMaxSharedWorld(t, nil)
			miniMaxSharedRetries(w, r, miniMaxSharedA, 2)
			private := "synthetic-private-error-text"
			raw := fmt.Sprintf(`{"type":"error","error":{"type":"api_error","message":%q},"code":"server_error","message":%q}`, private, private)
			reply := provider.SSE([][]byte{[]byte("event: error\ndata: " + raw + "\n\n")})
			w.provider.EnqueueAt(r.prefix(miniMaxSharedA), reply)
			o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-stream-error-"+r.name), r.target("MiniMax-M3"), miniMaxMessagesRequest(), r.options()))
			miniMaxSharedRecord(ev, w, o)
			ev.Record("response-script", reply)
			ev.Record("requests", w.provider.Requests())
			ev.Check("stream error is private with actual status and zero replay", o.err != nil && !strings.Contains(o.err.Error(), private) && len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%v", o.err)
			if e, ok := o.err.(*ai.Error); ok {
				ev.Check("known HTTP200 retained", e.HTTPStatus == 200, "%+v", e)
				if r.name == "messages" {
					ev.Check("documented finite stream type retained", strings.Contains(e.Error(), "api_error"), "%v", e)
				}
			}
			miniMaxSharedOneTerminal(ev, o)
			w.released(ev)
		})
	}
}

func miniMaxSharedRetries(w *mixedWorld, r miniMaxSharedRoute, k tenantKey, n int) {
	b := w.host.Binding(k.tenant, r.target("MiniMax-M3").BindingID)
	b.Retry.MaxRetries = n
	w.host.PutBinding(b)
}

func TestMiniMaxSharedMessagesPermanentType(t *testing.T) {
	r := miniMaxSharedRoutes[2]
	for _, typ := range []string{"invalid_request_error", "authentication_error", "permission_error", "not_found_error", "request_too_large"} {
		t.Run(typ, func(t *testing.T) {
			ev := run.Case(t, "MM05-permanent-"+typ)
			w := newMiniMaxSharedWorld(t, nil)
			miniMaxSharedRetries(w, r, miniMaxSharedA, 2)
			bad := provider.Reply{Status: 429, Header: map[string]string{"Content-Type": "application/json", "x-should-retry": "true", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(fmt.Sprintf(`{"type":"error","error":{"type":%q,"message":"private parameter"}}`, typ))}}
			_, first, _, raw := r.script(t, "MiniMax-M3")
			ev.Fixture("synthetic-source.json", raw)
			w.provider.EnqueueAt(r.prefix(miniMaxSharedA), bad, first)
			o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-permanent-"+typ), r.target("MiniMax-M3"), miniMaxMessagesRequest(), r.options()))
			miniMaxSharedRecord(ev, w, o)
			ev.Record("response-script", bad)
			ev.Record("requests", w.provider.Requests())
			ev.Check("documented permanent type cannot be forced to retry", o.err != nil && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%v %+v", o.err, o.result.Metadata)
			w.released(ev)
		})
	}
}
