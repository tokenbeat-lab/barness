package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestClassifierRequestRefusals(t *testing.T) {
	cases := map[string]func(*ai.ClassifierRequest){
		"no-questions":  func(r *ai.ClassifierRequest) { r.Questions = nil },
		"empty-key":     func(r *ai.ClassifierRequest) { r.Questions[""] = r.Questions["intent"] },
		"nil-question":  func(r *ai.ClassifierRequest) { r.Questions["intent"] = nil },
		"invalid-state": func(r *ai.ClassifierRequest) { r.State = json.RawMessage("42") },
		"state-limit":   func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`"` + strings.Repeat("x", 4096) + `"`) },
		"question-limit": func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Instructions = json.RawMessage(`"` + strings.Repeat("x", 4096) + `"`)
			r.Questions["intent"] = q
		},
		"count-limit": func(r *ai.ClassifierRequest) {
			for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
				r.Questions[k] = r.Questions["intent"]
			}
		},
		"no-criteria": func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Criteria = nil
			r.Questions["intent"] = q
		},
		"256-choices": func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Criteria = map[string]json.RawMessage{}
			for i := range 256 {
				q.Criteria[string(rune(1000+i))] = json.RawMessage(`"x"`)
			}
			r.Questions["intent"] = q
		},
		"empty-instructions": func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Instructions = json.RawMessage(`""`)
			r.Questions["intent"] = q
		},
		"invalid-criterion": func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Criteria["refund"] = json.RawMessage("false")
			r.Questions["intent"] = q
		},
	}
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E07-request-"+name)
			w := scenarioWorld(t, sc)
			r := classifierInput(t, sc)
			change(&r)
			res, err := w.client.Classify(ctxFor(t), textScope("req-invalid-"+name), sc.target(), r, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			var ae *ai.Error
			ev.Check("scope refusal", errors.As(err, &ae) && ae.Phase == ai.PhaseScope && (ae.Code == ai.CodeInvalidRequest || ae.Code == ai.CodeResourceLimit), "got %v", err)
			ev.Check("no provider or credential access", len(w.provider.Requests()) == 0 && w.host.ReadsFor("req-invalid-"+name).Credentials == 0, "sent %d", len(w.provider.Requests()))
		})
	}
}
func TestClassifierHooksAndRetry(t *testing.T) {
	f, raw := loadFixture(t, classifierProtocol, "choice.json")
	base := f.Scenarios[0]
	for _, name := range []string{"final-questions", "bad-question", "model-widen", "extra-authority", "header-widen", "response-failure", "retry"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E04-"+name)
			ev.Fixture("fixture.json", raw)
			w := scenarioWorld(t, base)
			payloads, responses, headers := 0, 0, 0
			sentinel := errors.New("synthetic host failure")
			hooks := ai.Hooks{TransformHeaders: func(_ context.Context, scope ai.CallScope, h http.Header) error {
				headers++
				ev.Check("header scope operation", scope.Operation == ai.OperationClassifier, "got %+v", scope)
				if name == "header-widen" {
					h.Set("Authorization", "Bearer foreign")
				}
				return nil
			},
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					payloads++
					ev.Check("payload operation", p.Operation == ai.OperationClassifier && p.ModelID == "host-jev", "got %+v", p)
					switch name {
					case "final-questions":
						p.Body["questions"] = map[string]any{"changed": map[string]any{"type": "choice", "instructions": "New question", "criteria": map[string]any{"only": "One"}}}
					case "bad-question":
						p.Body["questions"] = map[string]any{"bad": map[string]any{"type": "choice", "instructions": 42, "criteria": map[string]any{"x": "x"}}}
					case "model-widen":
						p.Body["model"] = "other"
					case "extra-authority":
						p.Body["endpoint"] = "https://other.example"
					}
					return ai.KeepPayload(), nil
				},
				OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
					responses++
					ev.Check("response operation", r.Operation == ai.OperationClassifier && r.ModelID == "host-jev", "got %+v", r)
					if name == "response-failure" {
						return sentinel
					}
					return nil
				}}
			reply := base.replies(t)[0]
			if name == "final-questions" {
				reply = fixtureReply{Status: 200, Body: `{"model":"version","answers":{"changed":{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1}}}`}.script(t, classifierProtocol, "")
			}
			if name == "retry" {
				w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				enqueue(ev, w, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
			}
			enqueue(ev, w, reply)
			res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-hooks-"+name), base.target(), classifierInput(t, base), ai.TypeSafeOptions{})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			want := ai.Code("")
			switch name {
			case "bad-question", "response-failure":
				want = ai.CodeCallbackFailed
			case "model-widen", "header-widen", "extra-authority":
				want = ai.CodeTenantDenied
			}
			if want != "" {
				ev.Check("callback failure classification", errors.Is(err, &ai.Error{Code: want, Phase: ai.PhaseRequest}), "got %v", err)
				ev.Check("answers empty", len(res.Answers) == 0, "got %+v", res.Answers)
			} else {
				ev.Check("success", err == nil, "got %v", err)
			}
			ev.Check("headers run once", headers == 1, "got %d", headers)
			expectedPayloads := 1
			if name == "header-widen" {
				expectedPayloads = 0
			}
			ev.Check("payload runs once outside retry", payloads == expectedPayloads, "got %d", payloads)
			if name == "final-questions" {
				_, ok := res.Answers["changed"].(ai.ChoiceAnswer)
				ev.Check("answers follow final questions", ok && len(res.Answers) == 1, "got %+v", res.Answers)
			}
			if name == "response-failure" {
				ev.Check("host cause retained", errors.Is(err, sentinel), "got %v", err)
			}
			sent := len(w.provider.Requests())
			if name == "retry" {
				ev.Check("two attempts but one response callback", sent == 2 && len(res.Metadata.Attempts) == 2 && responses == 1, "got %d/%d", sent, responses)
			}
			if name == "bad-question" || name == "model-widen" || name == "header-widen" || name == "extra-authority" {
				ev.Check("no request sent", sent == 0, "got %d", sent)
			}
		})
	}
}
func TestClassifierBodyLimitsAndRelease(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	for _, name := range []string{"above-frame", "output-limit", "error-limit", "bad-json", "read-abort", "read-idle", "canceled"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-"+name)
			bodies := provider.TrackBodies(provider.LoopbackTransport())
			gauges := probe.New()
			w := newWorldWith(t, func(c *ai.Config) {
				configureClassifier(c)
				c.Transport = bodies
				c.Probe = gauges
				c.Policy.MaxFrameBytes = 16
				c.Policy.MaxOutputBytes = 4096
				c.Policy.MaxToolJSONBytes = 1024
				c.Policy.ReadIdleTimeout = 50 * time.Millisecond
				if name == "output-limit" {
					c.Policy.MaxOutputBytes = 1024
				}
				if name == "error-limit" {
					c.Policy.MaxErrorBodyBytes = 16
				}
			}, tenantA)
			installClassifierBinding(w, tenantA)
			w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			reply := sc.replies(t)[0]
			switch name {
			case "above-frame":
				reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", 100))...)
			case "output-limit":
				reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", 1500))...)
			case "error-limit":
				reply = fixtureReply{Status: 529, Body: strings.Repeat("x", 30)}.script(t, classifierProtocol, "")
			case "bad-json":
				reply = fixtureReply{Status: 200, Body: "{"}.script(t, classifierProtocol, "")
			case "read-abort":
				reply.End = provider.EndAbort
			case "read-idle", "canceled":
				reply.End = provider.EndHold
			}
			reply.Script = joinChunks(reply.Chunks)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			if name == "canceled" {
				reply.OnHold = nil
			}
			enqueue(ev, w, reply)
			hc := w.client.WithHooks(ai.Hooks{})
			if name == "canceled" {
				hc = w.client.WithHooks(ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					time.AfterFunc(10*time.Millisecond, cancel)
					return nil
				}})
			}
			res, err := hc.Classify(ctx, textScope("req-body-"+name), sc.target(), classifierInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			want := ai.CodeProtocol
			phase := ai.PhaseResponse
			switch name {
			case "above-frame":
				want = ""
			case "output-limit", "error-limit":
				want = ai.CodeResourceLimit
			case "read-abort":
				want = ai.CodeTransport
			case "read-idle":
				want = ai.CodeDeadlineExceeded
			case "canceled":
				want = ai.CodeCanceled
			}
			if name == "error-limit" {
				phase = ai.PhaseRequest
			}
			if want == "" {
				ev.Check("unary JSON ignores frame limit", err == nil, "got %v", err)
			} else {
				ev.Check("classified read failure", errors.Is(err, &ai.Error{Code: want, Phase: phase}), "got %v", err)
			}
			ev.Check("never replay successful response or oversized error", len(w.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata)
			ev.Check("all resources released", bodies.Open() == 0 && gauges.Permits() == 0 && gauges.ActiveCalls() == 0 && gauges.QueuedEvents() == 0, "bodies=%d permits=%d calls=%d events=%d", bodies.Open(), gauges.Permits(), gauges.ActiveCalls(), gauges.QueuedEvents())
		})
	}
}
func TestClassifierPolicyAndSnapshots(t *testing.T) {
	for _, name := range []string{"questions", "state", "question"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-D1-"+name)
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 1, MaxStateBytes: 1, MaxQuestionBytes: 1}
			switch name {
			case "questions":
				p.Classifier.MaxQuestions = 0
			case "state":
				p.Classifier.MaxStateBytes = 0
			case "question":
				p.Classifier.MaxQuestionBytes = 0
			}
			w := newWorld(t, tenantA)
			_, err := ai.NewClient(ai.Config{Policy: p, Bindings: w.host, Credentials: w.host})
			ev.Record("error", errString(err))
			ev.Check("invalid subpolicy rejected", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
		})
	}
	t.Run("copies", func(t *testing.T) {
		ev := run.Case(t, "P07-E07-copies")
		rec := newRecorder()
		rec.gate = make(chan struct{})
		defer close(rec.gate)
		var policy *ai.ClassifierPolicy
		w := newWorldWith(t, func(c *ai.Config) {
			configureClassifier(c)
			policy = c.Policy.Classifier
			c.Observer = rec
			c.Policy.MaxQueuedObservations = 1
		}, tenantA)
		installClassifierBinding(w, tenantA)
		policy.MaxQuestions = 0
		policy.MaxStateBytes = 0
		policy.MaxQuestionBytes = 0
		f, _ := loadFixture(t, classifierProtocol, "choice.json")
		sc := f.Scenarios[0]
		req := classifierInput(t, sc)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, _ http.Header) error {
			req.State[0] = 'x'
			q := req.Questions["intent"].(ai.ChoiceQuestion)
			q.Criteria["refund"][0] = 'x'
			delete(req.Questions, "intent")
			return nil
		}}
		enqueue(ev, w, sc.replies(t)...)
		res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-copy"), sc.target(), req, nil)
		ev.Record("result", res)
		ev.Check("input and policy independently copied", err == nil && len(res.Answers) == 1, "got %v", err)
		ev.Check("observer congestion does not change result", w.client.ObserverStats().Dropped > 0, "got %+v", w.client.ObserverStats())
	})
}

// A trusted host may recover a callback panic outside the library. Preserve
// that panic while releasing resources, as the existing chat path did.
func TestClassifierCallbackPanicRelease(t *testing.T) {
	ev := run.Case(t, "P07-E08-callback-panic")
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	gauges := probe.New()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	w := newWorldWith(t, func(c *ai.Config) { configureClassifier(c); c.Probe = gauges; c.Transport = bodies }, tenantA)
	installClassifierBinding(w, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _ = w.client.WithHooks(ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { panic("synthetic callback panic") }}).Classify(ctxFor(t), textScope("req-panic"), sc.target(), classifierInput(t, sc), nil)
	}()
	ev.Check("host can recover its own panic", recovered == "synthetic callback panic", "got %v", recovered)
	ev.Check("unwinding releases call body and permit", bodies.Open() == 0 && gauges.ActiveCalls() == 0 && gauges.Permits() == 0, "bodies=%d calls=%d permits=%d", bodies.Open(), gauges.ActiveCalls(), gauges.Permits())
}
