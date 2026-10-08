package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type unaryHostError struct{}

func (*unaryHostError) Error() string { return "synthetic private callback diagnostic" }

func TestUnaryCallbackFailures(t *testing.T) {
	for _, stage := range []string{"headers", "payload", "response"} {
		for _, mode := range []string{"error", "cancel", "deadline"} {
			name := stage + "-" + mode
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P07-E04-unary-callback-"+name)
				u := newUnaryWorld(t, func(c *ai.Config) {
					if mode == "deadline" {
						fitCallTimeout(c.Policy, short)
					}
				})
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
				if stage == "response" {
					enqueue(ev, u.world, u.sc.replies(t)[0])
				}
				sentinel := &unaryHostError{}
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				fail := func(ctx context.Context) error {
					if mode == "cancel" {
						cancel()
					}
					if mode == "deadline" {
						<-ctx.Done()
					}
					return sentinel
				}
				hooks := ai.Hooks{}
				switch stage {
				case "headers":
					hooks.TransformHeaders = func(ctx context.Context, _ ai.CallScope, _ http.Header) error { return fail(ctx) }
				case "payload":
					hooks.OnPayload = func(ctx context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
						return ai.KeepPayload(), fail(ctx)
					}
				case "response":
					hooks.OnResponse = func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error { return fail(ctx) }
				}
				res, err := u.call(t, ctx, "req-callback-"+name, hooks)
				code, stop := ai.CodeCallbackFailed, ai.StopReasonError
				if mode == "cancel" {
					code = ai.CodeCanceled
					stop = ai.StopReasonAborted
				}
				if mode == "deadline" {
					code = ai.CodeDeadlineExceeded
					stop = ai.StopReasonAborted
				}
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, stop)
				var diagnostic *unaryHostError
				ev.Check("host cause Is and As survive context", errors.Is(err, sentinel) && errors.As(err, &diagnostic) && diagnostic == sentinel, "cause missing")
				ev.Check("private callback diagnostic never published", !strings.Contains(res.ErrorMessage, sentinel.Error()), "diagnostic leaked")
				n := 0
				if stage == "response" {
					n = 1
				}
				ev.Check("no callback retry", len(res.Metadata.Attempts) == n && len(u.provider.Requests()) == n, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestUnaryInvalidCallbackAndPolicy(t *testing.T) {
	for _, name := range []string{"nil-body", "unencodable-body", "question-count", "state-bytes", "question-bytes", "disabled"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E04-unary-invalid-"+name)
			u := newUnaryWorld(t, func(c *ai.Config) {
				if name == "disabled" {
					c.Policy.Classifier = nil
				}
			})
			called := 0
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				called++
				switch name {
				case "nil-body":
					return ai.ReplacePayload(nil), nil
				case "unencodable-body":
					p.Body["questions"] = make(chan int)
				case "question-count":
					qs := p.Body["questions"].(map[string]any)
					for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
						qs[k] = qs["intent"]
					}
				case "state-bytes":
					p.Body["state"] = strings.Repeat("x", 4097)
				case "question-bytes":
					p.Body["questions"].(map[string]any)["intent"].(map[string]any)["instructions"] = strings.Repeat("x", 4097)
				}
				return ai.KeepPayload(), nil
			}}
			res, err := u.call(t, ctxFor(t), "req-invalid-"+name, hooks)
			code, phase := ai.CodeCallbackFailed, ai.PhaseRequest
			if name == "disabled" {
				code = ai.CodeInvalidRequest
				phase = ai.PhaseScope
			}
			checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
			ev.Check("invalid final request never admitted or sent", len(u.adm.Requests()) == 0 && len(u.provider.Requests()) == 0 && len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata)
			if name == "disabled" {
				ev.Check("disabled before resolution or callback", called == 0 && u.host.ReadsFor(res.Metadata.RequestID).Bindings == 0, "called=%d", called)
			} else {
				ev.Check("payload called once", called == 1, "got %d", called)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
	for _, field := range []string{"MaxQuestions", "MaxStateBytes", "MaxQuestionBytes"} {
		t.Run("negative-"+field, func(t *testing.T) {
			ev := run.Case(t, "P07-D1-unary-negative-"+field)
			u := newUnaryWorld(t, nil)
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
			switch field {
			case "MaxQuestions":
				p.Classifier.MaxQuestions = -1
			case "MaxStateBytes":
				p.Classifier.MaxStateBytes = -1
			case "MaxQuestionBytes":
				p.Classifier.MaxQuestionBytes = -1
			}
			_, err := ai.NewClient(ai.Config{Policy: p, Bindings: u.host, Credentials: u.host})
			ev.Record("error", errString(err))
			ev.Check("non-positive capacity rejected", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
		})
	}
}
