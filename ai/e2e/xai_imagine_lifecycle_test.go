package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// The protocol has a new entry, but cancellation and admission retain the
// existing public unary contract. These route regressions precede any fixes.
func TestXAIImagineLifecycle(t *testing.T) {
	for _, mode := range []string{"cancel-before-send", "cancel-headers", "cancel-read", "header-timeout", "read-timeout", "response-callback", "response-panic"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-lifecycle-"+mode)
			g := probe.New()
			var bodies *provider.BodyTracker
			w := newImagineWorld(t, func(c *ai.Config) {
				c.Probe = g
				c.Policy.ResponseHeaderTimeout = 100 * time.Millisecond
				c.Policy.ReadIdleTimeout = 100 * time.Millisecond
				bodies = provider.TrackBodies(c.Transport)
				c.Transport = bodies
			})
			b := imagineBinding()
			if mode != "header-timeout" {
				b.Retry.MaxRetries = 1
			}
			w.host.PutBinding(b)
			ready := make(chan struct{})
			reply := imagineReply(`{"data":[{"b64_json":"` + imaginePNG + `"}]}`)
			if mode == "cancel-headers" || mode == "header-timeout" {
				reply = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
			}
			if mode == "cancel-read" || mode == "read-timeout" {
				reply.End = provider.EndHold
			}
			if mode != "cancel-before-send" {
				enqueue(ev, w, reply)
			}
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			if mode == "cancel-before-send" {
				cancel()
			}
			hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				if mode == "cancel-read" {
					cancel()
				}
				if mode == "response-callback" {
					return errors.New("synthetic response rejection")
				}
				if mode == "response-panic" {
					panic("synthetic Imagine panic")
				}
				return nil
			}}
			type terminal struct {
				res      ai.ImagesResult
				err      error
				panicked any
			}
			done := make(chan terminal, 1)
			go func() {
				v := terminal{}
				defer func() { v.panicked = recover(); done <- v }()
				v.res, v.err = w.client.WithHooks(hooks).GenerateImages(ctx, textScope("imagine-"+mode), imagineTarget(), ai.ImagesRequest{Prompt: "Draw a circle."}, nil)
			}()
			if mode == "cancel-headers" {
				waitFor(ev, "request reaches headers wait", ready)
				cancel()
			}
			v := waitValue(ev, "unary settles", done)
			ev.Record("result", v.res)
			ev.Record("error", errString(v.err))
			ev.Record("requests", w.provider.Requests())
			if mode == "response-panic" {
				ev.Check("panic retains caller contract", v.panicked == "synthetic Imagine panic", "panic=%v", v.panicked)
			} else {
				want := ai.CodeCanceled
				if mode == "header-timeout" || mode == "read-timeout" {
					want = ai.CodeDeadlineExceeded
				}
				if mode == "response-callback" {
					want = ai.CodeCallbackFailed
				}
				ev.Check("precise terminal and atomic content", errors.Is(v.err, &ai.Error{Code: want}) && len(v.res.Content) == 0, "err=%v", v.err)
			}
			requests := 1
			if mode == "cancel-before-send" {
				requests = 0
			}
			// A header timeout may use the operator's explicit initial retry;
			// all other selected failures occur before send or after 2xx.
			if mode != "header-timeout" {
				ev.Check("no replay", len(w.provider.Requests()) == requests, "requests=%d", len(w.provider.Requests()))
			}
			ev.Record("resources", map[string]int64{"calls": g.ActiveCalls(), "permits": g.Permits(), "waiters": g.AdmissionWaiters(), "bodies": bodies.Open()})
			ev.Check("all unary resources released", g.ActiveCalls() == 0 && g.Permits() == 0 && g.AdmissionWaiters() == 0 && bodies.Open() == 0, "resource leak")
		})
	}
}

func TestXAIImagineEntrySnapshot(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-entry-snapshot")
	req := ai.ImagesRequest{Prompt: "Edit.", ReferenceImages: []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}}
	opts := ai.XAIImagesOptions{Quality: ai.Value("low"), Resolution: ai.Value("1k")}
	w := newImagineWorld(t, func(c *ai.Config) {
		bindings := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
			req.ReferenceImages[0].Data = "caller mutation"
			opts.Quality = ai.Value("medium")
			return bindings.ResolveBinding(ctx, scope, id)
		})
	})
	enqueue(ev, w, imagineReply(`{"data":[{"b64_json":"`+imaginePNG+`"}]}`))
	res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-snapshot"), imagineTarget(), req, &opts)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	ev.Check("entry snapshot succeeds", err == nil && len(w.provider.Requests()) == 1, "err=%v", err)
	if len(w.provider.Requests()) == 1 {
		var wire map[string]any
		mustUnmarshal(t, w.provider.Requests()[0].Body, &wire)
		image, _ := wire["image"].(map[string]any)
		ev.Check("resolvers cannot mutate options or reference", wire["quality"] == "low" && image["url"] == "data:image/png;base64,"+imaginePNG, "wire=%+v", wire)
	}
}

func TestXAIImaginePinnedRevocation(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-pinned-revocation")
	w := newImagineWorld(t)
	arrived, resume := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(resume) }) }
	defer unblock()
	reply := imagineReply(`{"data":[{"b64_json":"` + imaginePNG + `"}]}`)
	reply.OnReceive = func() { close(arrived); <-resume }
	enqueue(ev, w, reply)
	type terminal struct {
		res ai.ImagesResult
		err error
	}
	done := make(chan terminal, 1)
	go func() {
		r, e := w.client.GenerateImages(ctxFor(t), textScope("imagine-pinned"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, nil)
		done <- terminal{r, e}
	}()
	waitFor(ev, "first call pinned and sent", arrived)
	b := imagineBinding()
	b.Enabled, b.Version = false, "disabled-v2"
	w.host.PutBinding(b)
	denied, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-revoked"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, nil)
	ev.Record("denied", denied)
	ev.Record("denied-error", errString(err))
	ev.Check("revocation denies new calls before credentials", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}) && len(denied.Metadata.Attempts) == 0 && w.host.ReadsFor("imagine-revoked").Credentials == 0, "err=%v", err)
	unblock()
	v := waitValue(ev, "pinned call finishes", done)
	ev.Record("pinned", v.res)
	ev.Record("requests", w.provider.Requests())
	ev.Check("inflight snapshot survives revoke", v.err == nil && v.res.Metadata.BindingVersion == "v1" && len(w.provider.Requests()) == 1, "err=%v metadata=%+v", v.err, v.res.Metadata)
}
