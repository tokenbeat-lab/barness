package e2e

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiMessagesSnapshots(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K14-snapshots")
			rec := newRecorder()
			w := newKimiMessagesConfigured(t, market, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
			entered, resume := make(chan struct{}), make(chan struct{})
			req := miniMaxToolRequest()
			var retained map[string]any
			hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, h http.Header) error {
				ev.Check("Bearer is redacted in header hook", h.Get("Authorization") == "[REDACTED]", "%+v", h)
				close(entered)
				select {
				case <-resume:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				retained = p.Body
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				retained["model"] = "after-serialization"
				return nil
			}}
			enqueue(ev, w, kimiMessagesReply())
			scope := textScope("kimi-snapshots")
			stream := w.client.WithHooks(hooks).Stream(ctxFor(t), scope, kimiMessagesTarget, req, ai.AnthropicOptions{Effort: ai.AnthropicEffortLow})
			defer stream.Close()
			waitFor(ev, "request snapshot established at header hook", entered)
			req.Tools[0].Parameters[0] = '['
			req.Messages[0] = ai.UserText("caller-mutated-input")
			close(resume)
			o := drain(stream)
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("mutable caller and callback data cannot affect active request", o.err == nil && len(w.provider.Requests()) == 1 && !strings.Contains(string(w.provider.Requests()[0].Body), "caller-mutated") && !strings.Contains(string(w.provider.Requests()[0].Body), "after-serialization"), "%v", o.err) {
				return
			}
			records := rec.awaitCall(ev, scope.RequestID)
			for _, record := range records {
				if record.Kind == ai.ObservationAttemptStarted {
					ev.Check("cost state begins unreported", record.Attempt != nil && record.Attempt.CostEstimate.Status == ai.CostUnreported, "%+v", record)
				}
				if record.Kind == ai.ObservationAttemptFinished {
					ev.Check("AttemptFinished receives final independent cost", record.Attempt != nil && reflect.DeepEqual(*record.Attempt, o.result.Metadata.Attempts[0]), "%+v", record)
				}
				if record.Kind == ai.ObservationCallFinished {
					ev.Check("CallFinished receives final cost snapshot", reflect.DeepEqual(record.Call.Attempts, o.result.Metadata.Attempts), "%+v", record)
				}
			}
			original := o.result.Metadata.Attempts[0].CostEstimate
			o.result.Metadata.Attempts[0].CostEstimate.Status = ai.CostResolved
			again, _ := stream.Result()
			ev.Check("Result cost estimate owns its values", again.Metadata.Attempts[0].CostEstimate == original, "%+v", again.Metadata.Attempts)
		})
	}
}
