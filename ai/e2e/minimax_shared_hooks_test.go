package e2e

import (
	"context"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
	"strings"
	"testing"
)

func TestMiniMaxSharedCallbackPanic(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, mode := range []string{"headers", "payload", "response"} {
			t.Run(r.name+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM05-hook-panic-"+r.name+mode)
				w := newMiniMaxSharedWorld(t, nil)
				req, first, _, raw := r.script(t, "MiniMax-M3")
				ev.Fixture("synthetic-source.json", raw)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedA), first)
				private := errors.New("synthetic-private-callback-panic")
				hooks := ai.Hooks{}
				switch mode {
				case "headers":
					hooks.TransformHeaders = func(context.Context, ai.CallScope, http.Header) error { panic(private) }
				case "payload":
					hooks.OnPayload = func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) { panic(private) }
				default:
					hooks.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error { panic(private) }
				}
				var o outcome
				var escaped bool
				func() {
					defer func() { escaped = recover() != nil }()
					o.result, o.err = w.client.WithHooks(hooks).Complete(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-panic-"+r.name+mode), r.target("MiniMax-M3"), req, r.options())
				}()
				ev.Check("panic remains one classified logical failure", !escaped && errors.Is(o.err, &ai.Error{Code: ai.CodeCallbackFailed, Phase: ai.PhaseRequest}), "escaped=%t error=%v", escaped, o.err)
				if escaped {
					return
				}
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				want := 0
				if mode == "response" {
					want = 1
				}
				ev.Check("only actual attempts retained", len(o.result.Metadata.Attempts) == want && len(w.provider.Requests()) == want, "%+v", o.result.Metadata)
				ev.Check("panic remains private and cause reachable", errors.Is(o.err, private) && !strings.Contains(o.result.Message.ErrorMessage, private.Error()), "%v", o.err)
				w.released(ev)
			})
		}
	}
}
