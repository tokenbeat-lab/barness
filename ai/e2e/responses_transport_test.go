package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestResponsesTransport covers the SDK and transport hardening the Responses
// adapter must apply regardless of SDK defaults (spec I5, H1).
func TestResponsesTransport(t *testing.T) {
	t.Run("sdk-retries-disabled", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "P01-H1-sdk-retries-disabled")
		// The SDK would retry a 500 twice by default; barness-ai's default is 0.
		enqueue(ev, w, provider.Reply{Status: http.StatusInternalServerError,
			Header: map[string]string{"Content-Type": "application/json"},
			Chunks: [][]byte{[]byte(`{"error":{"message":"boom","type":"server_error"}}`)}})
		res, err := w.client.Complete(ctxFor(t), textScope("req-no-retry"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Record("requests", w.provider.Requests())
		ev.Check("exactly one attempt reached the provider", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
		checkFailed(ev, res, err)
	})

	t.Run("cookie-jar-disabled", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "P01-H1-cookie-jar-disabled")
		reply := sseReply(t, f, provider.FramingLF)
		reply.Header["Set-Cookie"] = "affinity=tenant-a-session; Path=/"
		enqueue(ev, w, reply, reply)
		for i := range 2 {
			_, err := w.client.Complete(ctxFor(t), textScope("req-cookie-"+string(rune('a'+i))), textTarget(f), textRequest(f), nil)
			ev.Check("call succeeds", err == nil, "call %d: %v", i, err)
		}
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if ev.Check("two requests", len(reqs) == 2, "got %d", len(reqs)) {
			_, sent := reqs[1].Header["Cookie"]
			ev.Check("no cookie replayed on the next call", !sent, "Cookie=%q", reqs[1].Header["Cookie"])
		}
	})

	t.Run("redirect-not-followed", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "P01-H1-redirect-not-followed")
		elsewhere := provider.New(map[string]string{tenantA.secret: tenantA.alias})
		defer elsewhere.Close()
		enqueue(ev, w, provider.Reply{Status: http.StatusTemporaryRedirect,
			Header: map[string]string{"Location": elsewhere.URL() + "/v1/responses"}})
		res, err := w.client.Complete(ctxFor(t), textScope("req-redirect"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Record("requests", map[string]any{"origin": w.provider.Requests(), "redirect_target": elsewhere.Requests()})
		ev.Check("redirect target received nothing", len(elsewhere.Requests()) == 0, "got %d", len(elsewhere.Requests()))
		checkFailed(ev, res, err)
	})

	t.Run("no-environment-fallback", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "P01-H1-no-environment-fallback")
		decoy := provider.New(map[string]string{"sk-env-leak": "key:environment"})
		defer decoy.Close()
		// Everything the OpenAI SDK's default client options would read.
		t.Setenv("OPENAI_API_KEY", "sk-env-leak")
		t.Setenv("OPENAI_BASE_URL", decoy.URL()+"/v1")
		t.Setenv("OPENAI_ORG_ID", "org-env-leak")
		t.Setenv("OPENAI_PROJECT_ID", "proj-env-leak")
		t.Setenv("OPENAI_CUSTOM_HEADERS", "X-Env-Leak: 1")
		run.RedactSecret("sk-env-leak", "key:environment")
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))

		res, err := w.client.Complete(ctxFor(t), textScope("req-env"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("environment base URL not used", len(decoy.Requests()) == 0, "decoy got %d", len(decoy.Requests()))
		checkTextRequest(ev, w, f, tenantA, false)
		if reqs := w.provider.Requests(); len(reqs) == 1 {
			for _, h := range []string{"Openai-Organization", "Openai-Project", "X-Env-Leak"} {
				_, leaked := reqs[0].Header[h]
				ev.Check("no environment header "+h, !leaked, "%s=%q", h, reqs[0].Header[h])
			}
		}
	})
}

// startHardeningCase is startTextCase without a pre-scripted reply.
func startHardeningCase(t *testing.T, caseID string) (*evidence.Case, *world, textFixture) {
	t.Helper()
	ev := run.Case(t, caseID)
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("fixture.json", raw)
	return ev, newWorld(t, tenantA), f
}

// enqueue scripts replies on the provider and records them as the case's
// response script.
func enqueue(ev *evidence.Case, w *world, replies ...provider.Reply) {
	w.provider.Enqueue(replies...)
	ev.Record("response-script", replies)
}

func sseReply(t *testing.T, f textFixture, framing provider.Framing) provider.Reply {
	t.Helper()
	return sseEvents(t, f.Events, framing)
}

// sseEvents scripts a 200 SSE reply carrying events in the given framing.
func sseEvents(t *testing.T, events []json.RawMessage, framing provider.Framing) provider.Reply {
	t.Helper()
	chunks, err := provider.EncodeSSE(events, framing)
	if err != nil {
		t.Fatal(err)
	}
	return provider.SSE(chunks)
}

// checkFailed asserts the failed-call contract: a complete error message and a
// classified, non-nil error.
func checkFailed(ev *evidence.Case, res ai.Result, err error) {
	var aiErr *ai.Error
	ev.Check("error is a classified *ai.Error", errors.As(err, &aiErr), "err=%v", err)
	ev.Check("message ends with StopReason error", res.Message.StopReason == ai.StopReasonError, "got %q", res.Message.StopReason)
	ev.Check("message carries errorMessage", res.Message.ErrorMessage != "", "empty")
	ev.Check("failed message still carries its timestamp", res.Message.Timestamp > 0, "got %d", res.Message.Timestamp)
}
