package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestChatParity uses public entries and scripted wire events. Failure cases
// precede implementation: lost unfinished blocks, arguments.done mistaken for
// item.done, and rejecting a done received after the terminal event.
func TestChatParity(t *testing.T) {
	for _, p := range []struct {
		proto *fixtureProtocol
		files []string
	}{
		{responsesProtocol, []string{"parity.json", "parity-models.json", "parity-errors.json"}},
		{chatProtocol, []string{"parity-models.json", "parity-errors.json"}},
		{anthropicProtocol, []string{"parity-models.json"}},
	} {
		for _, file := range p.files {
			f, raw := loadFixture(t, p.proto, file)
			for _, sc := range f.Scenarios {
				for _, mode := range []callMode{modeStream, modeComplete} {
					t.Run(p.proto.dir+"/"+sc.ID+"/"+string(mode), func(t *testing.T) {
						ev := run.Case(t, "E02-parity-"+p.proto.dir+"-"+sc.ID+"-"+string(mode))
						runScenario(t, ev, sc, raw, mode)
					})
				}
			}
		}
	}
}

// Defaults have the same authorization boundary as call samplingParams,
// including construction with invalid JSON; they cannot be hidden by overrides.
func TestModelSamplingRejects(t *testing.T) {
	for _, api := range []ai.API{ai.APIOpenAIResponses, ai.APIOpenAICompletions} {
		for _, key := range []string{"model", "stream", "tools", "store", "prompt_cache_key", "bad-json"} {
			t.Run(string(api)+"/"+key, func(t *testing.T) {
				ev := run.Case(t, "E04-model-sampling-reject-"+string(api)+"-"+key)
				c := ai.BuiltinCatalog()
				var m ai.Model
				for _, candidate := range c.Models {
					if candidate.Provider == ai.ProviderOpenAI && candidate.API == api && candidate.ID == "gpt-4.1-mini" {
						m = candidate
						break
					}
				}
				m.SamplingParams = map[string]json.RawMessage{key: json.RawMessage(`true`)}
				if key == "bad-json" {
					m.SamplingParams[key] = json.RawMessage(`{`)
				}
				c.Models = []ai.Model{m}
				w := newWorld(t, tenantA)
				_, err := ai.NewClient(ai.Config{Catalog: &c, Policy: validPolicy(), Bindings: w.host, Credentials: w.host})
				var ce *ai.ConfigError
				ev.Check("constructor rejects unsafe defaults", errors.As(err, &ce) && ce.Field == "Catalog", "err=%v", err)
			})
		}
	}
}

// Invalid headers must produce a cancellable finite backoff at every public
// entry. A held fake clock proves no second request and response body release.
func TestInvalidRetryHeaderInterrupted(t *testing.T) {
	f, _ := loadRetryFixture(t)
	text, _ := loadTextFixture(t, "text-basic.json")
	for _, e := range outcomeEntries {
		for _, ending := range []string{"cancel", "deadline"} {
			t.Run(e.name+"/"+ending, func(t *testing.T) {
				ev := run.Case(t, "P01-E05-invalid-header-"+e.name+"-"+ending)
				clk := f.clock()
				held := clk.HoldSleeps()
				w, bodies := newRetryWorld(t, clk, ai.RetryPolicy{MaxRetries: 1})
				body := `{"error":{"message":"busy"}}`
				enqueue(ev, w, provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after": "Infinity"}, Chunks: [][]byte{[]byte(body)}, Script: body})
				ctx, cancel := context.WithCancel(ctxFor(t))
				if ending == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(ctxFor(t), 200*time.Millisecond)
				}
				defer cancel()
				result := make(chan outcome, 1)
				go func() {
					result <- e.invoke(ctx, w, textScope("req-invalid-header"), textTarget(text), textRequest(text))
				}()
				d := waitValue(ev, "backoff started", held)
				ev.Check("finite jittered backoff", d == 450*time.Millisecond, "got %s", d)
				if ending == "cancel" {
					cancel()
				}
				o := waitValue(ev, "call ended", result)
				o.record(ev)
				code := ai.CodeCanceled
				if ending == "deadline" {
					code = ai.CodeDeadlineExceeded
				}
				ev.Check("interruption classification", errors.Is(o.err, &ai.Error{Code: code, Phase: ai.PhaseRequest}), "err=%v", o.err)
				ev.Check("no second attempt", len(w.provider.Requests()) == 1, "requests=%d", len(w.provider.Requests()))
				ev.Check("response closed", bodies.Open() == 0, "open=%d", bodies.Open())
			})
		}
	}
}

// Provider-controlled tool references must pass through the normal error
// redaction boundary when an unfinished call is reported.
func TestUnfinishedToolErrorRedactsKey(t *testing.T) {
	ev := run.Case(t, "P01-E02-unfinished-tool-error-redacts-key")
	w := newWorld(t, tenantA)
	events := []json.RawMessage{
		mustMarshal(t, map[string]any{"type": "response.output_item.added", "output_index": 0,
			"item": map[string]any{"type": "function_call", "id": "fc_secret", "call_id": "call_secret", "name": tenantA.secret, "arguments": "{}"}}),
		json.RawMessage(`{"type":"response.completed","response":{"id":"resp_secret","status":"completed","output":[]}}`),
	}
	enqueue(ev, w, sseEvents(t, events, provider.FramingLF))
	res, err := w.client.Complete(ctxFor(t), textScope("req-unfinished-secret"), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}, ai.Request{Messages: []ai.Message{ai.UserText("Hi")}}, nil)
	ev.Record("result", res)
	ev.Check("protocol failure", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "err=%v", err)
	ev.Check("error text redacts the key", !strings.Contains(res.Message.ErrorMessage, tenantA.secret), "redaction missing")
}
