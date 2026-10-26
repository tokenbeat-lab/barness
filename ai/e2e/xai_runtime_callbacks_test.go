package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIRuntimeCallbackFailures(t *testing.T) {
	for _, phase := range []string{"headers", "payload", "response"} {
		name := phase + "-error"
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E14-callback-"+name)
			rec := newRecorder()
			w := newXAIRuntimeWorld(t, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
			enqueue(ev, w.world, xaiReply(t))
			var order []string
			failure := errors.New("private-prompt-ciphertext-host-diagnostic")
			visit := func(p string) error {
				order = append(order, p)
				if p != phase {
					return nil
				}
				return failure
			}
			h := ai.Hooks{TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { return visit("headers") }, OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				return ai.KeepPayload(), visit("payload")
			}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { return visit("response") }}
			scope := textScope("xai-callback-" + name)
			o := drain(w.client.WithHooks(h).Stream(ctxFor(t), scope, xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("callback-order", order)
			ev.Record("requests", w.provider.Requests())
			wantOrder, wantRequests := "headers", 0
			if phase == "payload" {
				wantOrder = "headers,payload"
			}
			if phase == "response" {
				wantOrder, wantRequests = "headers,payload,response", 1
			}
			ev.Check("callbacks fail once before domain start", errors.Is(o.err, &ai.Error{Code: ai.CodeCallbackFailed, Phase: ai.PhaseRequest}) && strings.Join(order, ",") == wantOrder && !hasEvent(o.events, "start") && len(w.provider.Requests()) == wantRequests, "%v events=%v order=%v", o.err, eventSummary(o.events), order)
			obs := rec.awaitCall(ev, scope.RequestID)
			ev.Check("callback text stays outside public diagnostics and Observer", !strings.Contains(errString(o.err), "private-") && !strings.Contains(o.result.Message.ErrorMessage, "private-") && !strings.Contains(string(mustMarshal(t, obs)), "private-"), "%v", o.err)
			glmSharedOneTerminal(ev, o)
			w.released(ev)
		})
	}
}

func TestXAIRuntimeRedirectRefused(t *testing.T) {
	ev := run.Case(t, "XAI-E14-redirect-refused")
	w := newXAIRuntimeWorld(t, nil)
	r := provider.Reply{Status: 307, Header: map[string]string{"Location": "https://untrusted.invalid/collect", "Content-Type": "application/json"}, Script: `{}`, Chunks: [][]byte{[]byte(`{}`)}}
	enqueue(ev, w.world, r)
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	var ae *ai.Error
	ev.Check("redirect stops at original response", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseRequest}) && errors.As(o.err, &ae) && ae.HTTPStatus == 307 && len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%v %+v", o.err, o.result.Metadata)
	w.released(ev)
}

// xAI preserves the existing synchronous Host panic contract (spec R19).
// Running this through Complete lets the Host recover without crashing the
// test process; Stream deliberately retains the public propagation policy.
func TestXAIRuntimeCallbackPanicPropagation(t *testing.T) {
	for _, phase := range []string{"headers", "payload", "response"} {
		t.Run(phase, func(t *testing.T) {
			ev := run.Case(t, "XAI-E14-callback-panic-propagates-"+phase)
			w := newXAIRuntimeWorld(t, nil)
			enqueue(ev, w.world, xaiReply(t))
			failure := errors.New("synthetic Host panic")
			h := ai.Hooks{}
			switch phase {
			case "headers":
				h.TransformHeaders = func(context.Context, ai.CallScope, http.Header) error { panic(failure) }
			case "payload":
				h.OnPayload = func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) { panic(failure) }
			case "response":
				h.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error { panic(failure) }
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_, _ = w.client.WithHooks(h).Complete(ctxFor(t), textScope("xai-panic-"+phase), xaiTarget(), xaiRequest(), nil)
			}()
			ev.Record("requests", w.provider.Requests())
			ev.Record("host-recovered-original-panic", recovered == failure)
			ev.Check("Host receives its original panic", recovered == failure, "original panic returned=%v", recovered == failure)
			w.released(ev)
		})
	}
}
