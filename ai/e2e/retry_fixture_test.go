package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// retryFixture is testdata/responses/retry.json: E05 scenarios run against
// text-basic.json's logical input, whose successful reply is "ok".
type retryFixture struct {
	Now       time.Time       `json:"now"`
	Jitter    float64         `json:"jitter"`
	Scenarios []retryScenario `json:"scenarios"`
}

type retryScenario struct {
	ID string `json:"id"`
	// Retry is the binding's retry policy; absent keeps the zero value.
	Retry   *retrySettings `json:"retry"`
	Replies []retryReply   `json:"replies"`
	Expect  retryExpect    `json:"expect"`
	// PidiffSkip, when set, says why the scenario stays out of the pi
	// differential (pi waits for real, on its own clock).
	PidiffSkip string `json:"pidiffSkip"`
}

// retrySettings are a RetryPolicy in pi's option names. A maxRetryDelayMs
// that is absent keeps the default cap; 0 disables it.
type retrySettings struct {
	MaxRetries      int      `json:"maxRetries"`
	MaxRetryDelayMs *float64 `json:"maxRetryDelayMs"`
}

func (s *retrySettings) policy() ai.RetryPolicy {
	if s == nil {
		return ai.RetryPolicy{}
	}
	p := ai.RetryPolicy{MaxRetries: s.MaxRetries}
	if s.MaxRetryDelayMs != nil {
		d := msDuration(*s.MaxRetryDelayMs)
		p.MaxRetryDelay = &d
	}
	return p
}

// piOptions are the same settings as pi stream options.
func (s *retrySettings) piOptions() map[string]any {
	if s == nil {
		return nil
	}
	opts := map[string]any{"maxRetries": s.MaxRetries}
	if s.MaxRetryDelayMs != nil {
		opts["maxRetryDelayMs"] = *s.MaxRetryDelayMs
	}
	return opts
}

// retryReply is one scripted attempt: the text fixture's successful SSE
// reply ("ok"), a connection dropped before any response ("drop"), an SSE
// reply of events, or a plain HTTP reply.
type retryReply struct {
	OK     bool              `json:"ok"`
	Drop   bool              `json:"drop"`
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
	Events []json.RawMessage `json:"events"`
	End    provider.End      `json:"end"`
}

type retryExpect struct {
	Requests          int           `json:"requests"`
	SleepsMs          []float64     `json:"sleepsMs"`
	StopReason        ai.StopReason `json:"stopReason"`
	Code              ai.Code       `json:"code"`
	Phase             ai.Phase      `json:"phase"`
	HTTPStatus        int           `json:"httpStatus"`
	ProviderRequestID string        `json:"providerRequestId"`
	RetryAfterMs      float64       `json:"retryAfterMs"`
	ErrorMessage      string        `json:"errorMessage"`
	// Content and ResponseID are checked on a failure that kept partial
	// output; a success is checked against the text fixture instead.
	Content    []string `json:"content"`
	ResponseID string   `json:"responseId"`
	// Responses is how often OnResponse ran: once when an initial response
	// arrived, never otherwise.
	Responses int            `json:"responses"`
	Attempts  []retryAttempt `json:"attempts"`
}

type retryAttempt struct {
	HTTPStatus        int     `json:"httpStatus"`
	Code              ai.Code `json:"code"`
	ProviderRequestID string  `json:"providerRequestId"`
	RetryDelayMs      float64 `json:"retryDelayMs"`
}

func loadRetryFixture(t *testing.T) (retryFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "retry.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f retryFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("retry.json: %v", err)
	}
	return f, raw
}

func (f retryFixture) clock() *clock.Clock { return clock.NewFake(f.Now, f.Jitter) }

// replies scripts sc's attempts; ok is the text fixture's reply.
func (sc retryScenario) replies(t *testing.T, text textFixture) []provider.Reply {
	t.Helper()
	out := make([]provider.Reply, 0, len(sc.Replies))
	for _, r := range sc.Replies {
		switch {
		case r.OK:
			out = append(out, sseReply(t, text, provider.FramingLF))
		case r.Drop:
			out = append(out, provider.Reply{End: provider.EndDrop})
		case r.Events != nil:
			reply := sseEvents(t, r.Events, provider.FramingLF)
			reply.End = r.End
			out = append(out, reply)
		default:
			header := map[string]string{"Content-Type": "application/json"}
			for k, v := range r.Header {
				header[k] = v
			}
			out = append(out, provider.Reply{Status: r.Status, Header: header, Chunks: [][]byte{[]byte(r.Body)}, Script: r.Body})
		}
	}
	return out
}

func msDuration(ms float64) time.Duration { return time.Duration(ms * float64(time.Millisecond)) }

// newRetryWorld is tenant A's tracked world on clk, with the binding's retry
// policy set.
func newRetryWorld(t *testing.T, clk *clock.Clock, policy ai.RetryPolicy) (*world, *provider.BodyTracker) {
	t.Helper()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	w := newWorldWith(t, func(c *ai.Config) { c.Transport, c.Clock = bodies, clk }, tenantA)
	w.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = policy })
	return w, bodies
}

// hookCounts counts the trusted callbacks of one call.
type hookCounts struct{ payloads, responses atomic.Int32 }

func (c *hookCounts) hooks() ai.Hooks {
	return ai.Hooks{
		OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
			c.payloads.Add(1)
			return ai.KeepPayload(), nil
		},
		OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
			c.responses.Add(1)
			return nil
		},
	}
}

// hookedEntries are the four public entry points of h, reporting whatever
// happens.
func hookedEntries(h *ai.HookedClient) []outcomeEntry {
	return []outcomeEntry{
		{"stream-full", func(ctx context.Context, _ *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(h.Stream(ctx, scope, target, req, nil))
		}},
		{"stream-simple", func(ctx context.Context, _ *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(h.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{}))
		}},
		{"complete-full", func(ctx context.Context, _ *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := h.Complete(ctx, scope, target, req, ai.ResponsesOptions{})
			return outcome{result: res, err: err}
		}},
		{"complete-simple", func(ctx context.Context, _ *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := h.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{})
			return outcome{result: res, err: err}
		}},
	}
}

// retryPidiffScenario is E05 scenario sc of TestExplicitRetry: text-basic's
// logical input with sc's attempts and retry settings on both sides.
func retryPidiffScenario(t *testing.T, f retryFixture, raw []byte, text textFixture, sc retryScenario) pidiffScenario {
	t.Helper()
	return pidiffScenario{fixture: "retry.json", raw: raw, model: text.Model, req: textRequest(text),
		replies: sc.replies(t, text), requests: sc.Expect.Requests,
		retry: sc.Retry.policy(), clock: f.clock(), piOptions: sc.Retry.piOptions()}
}
