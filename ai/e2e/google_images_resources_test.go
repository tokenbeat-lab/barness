package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGoogleImagesResources(t *testing.T) {
	for _, editing := range []bool{false, true} {
		for _, name := range []string{"above-frame", "output-limit", "image-limit", "total-limit", "count-limit", "model-count-limit", "bad-json", "trailing-json", "duplicate-usage", "duplicate-status", "duplicate-steps", "duplicate-block", "read-abort", "read-idle", "canceled", "response-panic", "response-error", "error-limit", "retry", "no-retry"} {
			label := name
			if editing {
				label = "edit-" + name
			}
			t.Run(label, func(t *testing.T) {
				ev := run.Case(t, "P09-E08-resources-"+label)
				sc := firstGoogleImagesScenario(t)
				if editing {
					sc = googleEditScenario(t)
				}
				g := probe.New()
				rec := newRecorder()
				b := provider.TrackBodies(provider.LoopbackTransport())
				w := newWorldWith(t, func(c *ai.Config) {
					configureGoogleImages(c)
					c.Probe = g
					c.Observer = rec
					c.Policy.MaxQueuedObservations = 16
					c.Transport = b
					c.Policy.MaxFrameBytes = 16
					c.Policy.ReadIdleTimeout = 100 * time.Millisecond
					switch name {
					case "output-limit":
						c.Policy.MaxOutputBytes = 4096
						c.Policy.MaxToolJSONBytes = 1024
					case "image-limit":
						c.Policy.Image.MaxOutputImageBytes = 16
					case "total-limit":
						c.Policy.Image.MaxOutputImageBytes = 100
						c.Policy.Image.MaxTotalOutputImageBytes = 100
					case "count-limit":
						c.Policy.Image.MaxOutputImages = 1
					case "model-count-limit":
						c.Catalog.ImageModels[0].Capabilities.MaxOutputImages = 1
					case "error-limit":
						c.Policy.MaxErrorBodyBytes = 16
					}
				}, tenantA)
				installGoogleImagesBinding(w, tenantA)
				if name != "no-retry" {
					w.updateBindingOf(tenantA, "images", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				}
				reply := sc.replies(t)[0]
				var body map[string]any
				mustUnmarshal(t, []byte(sc.Replies[0].Body), &body)
				switch name {
				case "output-limit":
					reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", 5000))...)
				case "total-limit", "count-limit", "model-count-limit":
					step := body["steps"].([]any)[0].(map[string]any)
					images := step["content"].([]any)
					step["content"] = append(images, images[0])
					reply.Chunks[0] = mustMarshal(t, body)
				case "bad-json":
					reply.Chunks[0] = []byte("{")
				case "trailing-json":
					reply.Chunks[0] = append(reply.Chunks[0], []byte(" {}")...)
				case "duplicate-usage":
					reply.Chunks[0] = []byte(`{"status":"completed","steps":[],"usage":{},"usage":{}}`)
				case "duplicate-status":
					reply.Chunks[0] = []byte(`{"status":"completed","status":"completed","steps":[]}`)
				case "duplicate-steps":
					reply.Chunks[0] = []byte(`{"status":"completed","steps":[],"steps":[]}`)
				case "duplicate-block":
					reply.Chunks[0] = []byte(`{"status":"completed","steps":[{"type":"model_output","content":[{"type":"text","text":"a","text":"b"}]}]}`)
				case "read-abort":
					reply.End = provider.EndAbort
				case "read-idle", "canceled":
					reply.End = provider.EndHold
				case "error-limit":
					reply = fixtureReply{Status: 503, Body: strings.Repeat("x", 20)}.script(t, googleImagesProtocol, "")
				}
				reply.Script = joinChunks(reply.Chunks)
				if name == "retry" || name == "no-retry" {
					enqueue(ev, w, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":{"message":"limited"}}`}.script(t, googleImagesProtocol, ""))
				}
				enqueue(ev, w, reply)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				h := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					ev.Check("permit and body held before parse", g.Permits() == 1 && b.Open() == 1, "permits=%d bodies=%d", g.Permits(), b.Open())
					if name == "canceled" {
						time.AfterFunc(10*time.Millisecond, cancel)
					}
					if name == "response-panic" {
						panic("synthetic Google callback panic")
					}
					if name == "response-error" {
						return errors.New("synthetic Google callback failure")
					}
					return nil
				}}
				var res ai.ImagesResult
				var err error
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					res, err = w.client.WithHooks(h).GenerateImages(ctx, textScope("req-google-resource-"+label), sc.target(), imagesInput(t, sc), nil)
				}()
				ev.Record("result", res)
				ev.Record("error", errString(err))
				ev.Record("requests", w.provider.Requests())
				code, phase := ai.CodeProtocol, ai.PhaseResponse
				switch name {
				case "above-frame", "retry":
					code = ""
				case "output-limit", "image-limit", "total-limit", "count-limit", "model-count-limit", "error-limit":
					code = ai.CodeResourceLimit
				case "read-abort":
					code = ai.CodeTransport
				case "read-idle":
					code = ai.CodeDeadlineExceeded
				case "canceled":
					code = ai.CodeCanceled
				case "response-error":
					code = ai.CodeCallbackFailed
					phase = ai.PhaseRequest
				case "no-retry":
					code = ai.CodeRateLimited
					phase = ai.PhaseRequest
				}
				if name == "error-limit" {
					phase = ai.PhaseRequest
				}
				if name == "response-panic" {
					ev.Check("panic unwinds lifetime", recovered == "synthetic Google callback panic", "got %v", recovered)
				} else if code == "" {
					ev.Check("bounded success", err == nil, "got %v", err)
				} else {
					ev.Check("atomic failure", errors.Is(err, &ai.Error{Code: code, Phase: phase}) && len(res.Content) == 0, "got %v", err)
				}
				wantRequests := 1
				if name == "retry" {
					wantRequests = 2
				}
				if name == "trailing-json" || name == "read-abort" {
					want := *sc.Expect.Usage
					ev.Check("reported usage survives parse/read failure", res.Usage.Input == want.Input && res.Usage.Output == want.Output && res.Usage.TotalTokens == want.TotalTokens, "got %+v", res.Usage)
					ev.Check("attempt retains consumption", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageComplete && res.Metadata.Attempts[0].Usage.Output == want.Output, "got %+v", res.Metadata.Attempts)
					eventually(ev, "observer retains consumption", func() bool {
						obs := rec.all()
						return len(obs) == 4 && obs[len(obs)-1].Usage.Output == want.Output
					})
					ev.Record("observations", rec.all())
				}
				ev.Check("no success replay", len(w.provider.Requests()) == wantRequests, "got %d", len(w.provider.Requests()))
				ev.Record("resources", map[string]int64{"bodies": b.Open(), "calls": g.ActiveCalls(), "permits": g.Permits(), "waiters": g.AdmissionWaiters()})
				ev.Check("released all resources", b.Open() == 0 && g.ActiveCalls() == 0 && g.Permits() == 0 && g.AdmissionWaiters() == 0, "resources remain")
			})
		}
	}
}

func TestGoogleImagesUsageObserver(t *testing.T) {
	ev := run.Case(t, "P09-E09-modality-audit")
	rec := newRecorder()
	w := newWorldWith(t, func(c *ai.Config) { configureGoogleImages(c); c.Observer = rec; c.Policy.MaxQueuedObservations = 16 }, tenantA)
	installGoogleImagesBinding(w, tenantA)
	sc := googleEditScenario(t)
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.GenerateImages(ctxFor(t), textScope("req-google-observer"), sc.target(), imagesInput(t, sc), nil)
	ev.Check("success", err == nil, "got %v", err)
	eventually(ev, "four observer records", func() bool { return len(rec.all()) == 4 })
	obs := rec.all()
	ev.Record("observations", obs)
	for _, o := range obs {
		ev.Check("pinned attribution", o.Call.Operation == ai.OperationImage && (!o.Call.Resolved || (o.Call.ProviderID == ai.ProviderGoogle && o.Call.API == googleImagesProtocol.api && o.Call.ModelID == sc.Model)), "got %+v", o.Call)
	}
	before := observationText(t, obs)
	if res.Usage.Modalities != nil {
		res.Usage.Modalities.OutputImage = ai.Value[int64](999)
	}
	if len(res.Metadata.Attempts) > 0 && res.Metadata.Attempts[0].Usage.Modalities != nil {
		res.Metadata.Attempts[0].Usage.Modalities.InputText = ai.Value[int64](999)
	}
	ev.Check("observer independently owns usage", observationText(t, rec.all()) == before, "observer changed")
	ev.Check("only counts and metadata observed", !strings.Contains(before, "interaction-fixture") && !strings.Contains(before, "露台") && !strings.Contains(before, "iVBOR") && !strings.Contains(before, googleA.secret), "observer content leaked")
	ev.Record("usage", obs[len(obs)-1].Usage)
}

func TestGoogleImagesExtensionRegistration(t *testing.T) {
	ev := run.Case(t, "P09-E11-native-catalog-exemption")
	ledger := loadLedger(t)
	route, ok := ledger.Route(string(ai.ProviderGoogle), string(googleImagesProtocol.api))
	ev.Check("explicit native image route", ok && route.Ref != "", "missing Google extension")
	for _, file := range []string{"generation.json", "failures.json", "options.json", "usage.json", "editing.json", "editing-invalid-input.json", "editing-failures.json", "editing-usage.json", "editing-options.json"} {
		f, _ := loadFixture(t, googleImagesProtocol, file)
		for _, sc := range f.Scenarios {
			ev.Check("explicit differential skip "+sc.ID, sc.PidiffSkip != "", "missing differential skip")
		}
	}
	for _, m := range ai.BuiltinCatalog().ImageModels {
		if m.API == googleImagesProtocol.api {
			ev.Check("own live-confirmed builtin remains an extension", m.Provider == ai.ProviderGoogle && m.ID == "gemini-nano-banana-2.1", "unverified model listed")
		}
	}
	// This protocol does not become a chat route or acquire native continuation.
	var encoded map[string]any
	mustUnmarshal(t, mustMarshal(t, ai.ImagesResult{}), &encoded)
	_, native := encoded["nativeState"]
	ev.Check("diagnostic ID only", !native, "native continuation leaked")
}
