package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestGeminiUsageAndRetry is P03/E11 and E05: usage, cache reads, thinking
// tokens, reporting completeness and cost (testdata/gemini/usage.json), and
// the binding's explicit retry of the initial request as pi's Google path
// retries it (testdata/gemini/retry.json).
func TestGeminiUsageAndRetry(t *testing.T) {
	for _, file := range []string{"usage.json", "retry.json"} {
		f, raw := loadFixture(t, geminiProtocol, file)
		for _, sc := range f.Scenarios {
			t.Run(sc.ID, func(t *testing.T) {
				ev := run.Case(t, "P03-"+sc.ID)
				runScenario(t, ev, sc, raw, modeStream)
			})
		}
	}

	// Each attempt is recorded on its own: the refused one reported nothing
	// and kept the delay the provider asked for only as error metadata, the
	// streamed one its usage; the payload callback ran once, outside the
	// retry.
	t.Run("attempts-and-payload-once", func(t *testing.T) {
		f, raw := loadFixture(t, geminiProtocol, "retry.json")
		sc := scenarioByID(t, f, "retry-ignores-retry-after")
		ev := run.Case(t, "P03-E05-attempts-and-payload-once")
		ev.Fixture("fixture.json", raw)
		w := scenarioWorld(t, sc)
		enqueue(ev, w, sc.replies(t)...)
		payloads := 0
		hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
			payloads++
			return ai.KeepPayload(), nil
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-retry-attempts"), sc.target(), sc.request(t), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		a := res.Metadata.Attempts
		if !ev.Check("two attempts", len(a) == 2, "got %+v", a) {
			return
		}
		ev.Check("the first failed with its status, reported no usage and backed off instead of the requested 120s",
			a[0].HTTPStatus == 503 && a[0].Code == ai.CodeUpstreamError && a[0].UsageReporting == ai.UsageUnreported &&
				a[0].Usage == ai.Usage{} && a[0].RetryDelay > 0 && a[0].RetryDelay <= 500_000_000, "got %+v", a[0])
		ev.Check("the second streamed and reported the message's usage", a[1].HTTPStatus == 200 && a[1].Code == "" &&
			a[1].UsageReporting == ai.UsageComplete && a[1].Usage == res.Message.Usage, "got %+v", a[1])
		ev.Check("payload callback ran once", payloads == 1, "got %d", payloads)
		reqs := w.provider.Requests()
		ev.Check("both attempts sent the same body", len(reqs) == 2 && jsonEqual(reqs[0].Body, reqs[1].Body), "got %d requests", len(reqs))
	})
}
