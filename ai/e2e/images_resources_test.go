package e2e

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestOpenAIImagesResources(t *testing.T) {
	for _, name := range []string{"above-frame", "output-limit", "image-limit", "total-limit", "count-limit", "model-count-limit", "missing-requested-image", "bad-json", "trailing-json", "duplicate-data", "duplicate-usage", "read-abort", "read-idle", "canceled", "response-panic", "error-limit"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E08-resources-"+name)
			sc := firstImagesScenario(t)
			g := probe.New()
			w, b := imagesTrackedWorld(t, func(c *ai.Config) {
				c.Probe = g
				c.Policy.MaxFrameBytes = 16
				c.Policy.ReadIdleTimeout = 100 * time.Millisecond
				if name == "output-limit" {
					c.Policy.MaxOutputBytes = 4096
					c.Policy.MaxToolJSONBytes = 1024
				}
				if name == "image-limit" {
					c.Policy.Image.MaxOutputImageBytes = 16
				}
				if name == "total-limit" {
					c.Policy.Image.MaxTotalOutputImageBytes = 100
					c.Policy.Image.MaxOutputImageBytes = 100
				}
				if name == "count-limit" {
					c.Policy.Image.MaxOutputImages = 1
				}
				if name == "model-count-limit" {
					c.Catalog.ImageModels[0].Capabilities.MaxOutputImages = 1
				}
				if name == "error-limit" {
					c.Policy.MaxErrorBodyBytes = 16
				}
			})
			w.updateBindingOf(tenantA, "images", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			reply := sc.replies(t)[0]
			var body map[string]any
			mustUnmarshal(t, []byte(sc.Replies[0].Body), &body)
			switch name {
			case "output-limit":
				reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", 5000))...)
			case "total-limit", "count-limit", "model-count-limit":
				data := body["data"].([]any)
				body["data"] = append(data, data[0])
				reply.Chunks[0] = mustMarshal(t, body)
			case "bad-json":
				reply.Chunks[0] = []byte("{")
			case "trailing-json":
				reply.Chunks[0] = append(reply.Chunks[0], []byte(" {}")...)
			case "duplicate-data":
				reply.Chunks[0] = []byte(`{"data":[],"data":[]}`)
			case "duplicate-usage":
				reply.Chunks[0] = []byte(`{"data":[],"usage":{},"usage":{}}`)
			case "read-abort":
				reply.End = provider.EndAbort
			case "read-idle", "canceled":
				reply.End = provider.EndHold
			case "error-limit":
				reply = fixtureReply{Status: 503, Body: strings.Repeat("x", 20)}.script(t, imagesProtocol, "")
			}
			reply.Script = joinChunks(reply.Chunks)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			h := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				if name == "canceled" {
					time.AfterFunc(10*time.Millisecond, cancel)
				}
				if name == "response-panic" {
					panic("synthetic image callback panic")
				}
				return nil
			}}
			enqueue(ev, w, reply)
			var res ai.ImagesResult
			var err error
			var recovered any
			options := ai.OpenAIImagesOptions{}
			if name == "missing-requested-image" || name == "total-limit" {
				options.N = ai.Value(2)
			}
			func() {
				defer func() { recovered = recover() }()
				res, err = w.client.WithHooks(h).GenerateImages(ctx, textScope("req-resource-"+name), sc.target(), imagesInput(t, sc), options)
			}()
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			want, phase := ai.CodeProtocol, ai.PhaseResponse
			switch name {
			case "above-frame":
				want = ""
			case "output-limit", "image-limit", "total-limit", "count-limit", "model-count-limit", "error-limit":
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
			if name == "response-panic" {
				ev.Check("panic retained", recovered == "synthetic image callback panic", "got %v", recovered)
			} else if want == "" {
				ev.Check("unary ignores frame bound", err == nil, "got %v", err)
			} else {
				ev.Check("failure has empty atomic output", errors.Is(err, &ai.Error{Code: want, Phase: phase}) && len(res.Content) == 0, "got %v", err)
			}
			ev.Check("no replay after successful headers or byte limit", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
			ev.Record("resources", map[string]int64{"bodies": b.Open(), "calls": g.ActiveCalls(), "permits": g.Permits(), "events": g.QueuedEvents(), "waiters": g.AdmissionWaiters()})
			ev.Check("all resources released", b.Open() == 0 && g.ActiveCalls() == 0 && g.Permits() == 0 && g.QueuedEvents() == 0 && g.AdmissionWaiters() == 0, "resources remain")
		})
	}
}

func TestOpenAIImagesPolicySnapshot(t *testing.T) {
	for _, field := range []string{"input", "output", "image", "total", "total-over-output", "image-over-total"} {
		t.Run(field, func(t *testing.T) {
			ev := run.Case(t, "P08-D1-policy-"+field)
			w := newWorld(t, tenantA)
			cfg := ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host}
			configureImages(&cfg)
			switch field {
			case "input":
				cfg.Policy.Image.MaxInputImages = 0
			case "output":
				cfg.Policy.Image.MaxOutputImages = 0
			case "image":
				cfg.Policy.Image.MaxOutputImageBytes = 0
			case "total":
				cfg.Policy.Image.MaxTotalOutputImageBytes = 0
			case "total-over-output":
				cfg.Policy.Image.MaxTotalOutputImageBytes = cfg.Policy.MaxOutputBytes + 1
			case "image-over-total":
				cfg.Policy.Image.MaxOutputImageBytes = cfg.Policy.Image.MaxTotalOutputImageBytes + 1
			}
			_, err := ai.NewClient(cfg)
			ev.Record("error", errString(err))
			ev.Check("invalid policy rejected", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
		})
	}
	t.Run("copies", func(t *testing.T) {
		ev := run.Case(t, "P08-E07-policy-copies")
		var policy *ai.ImagePolicy
		var catalog *ai.Catalog
		w := newWorldWith(t, func(c *ai.Config) { configureImages(c); policy = c.Policy.Image; catalog = c.Catalog }, tenantA)
		installImagesBinding(w, tenantA)
		policy.MaxOutputImages = 0
		catalog.ImageModels[0].Capabilities.Sizes[0] = "corrupted"
		catalog.ImageModels[0].Capabilities.Qualities[0] = "corrupted"
		sc := firstImagesScenario(t)
		enqueue(ev, w, sc.replies(t)...)
		res, err := w.client.GenerateImages(ctxFor(t), textScope("req-copies"), sc.target(), imagesInput(t, sc), ai.OpenAIImagesOptions{Quality: ai.Value("auto"), Size: ai.Value("auto")})
		ev.Record("result", res)
		ev.Check("construction snapshot independent", err == nil, "got %v", err)
	})
}

func TestOpenAIImagesObserverIsolation(t *testing.T) {
	ev := run.Case(t, "P08-E09-observer-isolation")
	sc := firstImagesScenario(t)
	rec := newRecorder()
	rec.gate = make(chan struct{})
	w := newWorldWith(t, func(c *ai.Config) { configureImages(c); c.Observer = rec; c.Policy.MaxQueuedObservations = 16 }, tenantA)
	installImagesBinding(w, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.GenerateImages(ctxFor(t), textScope("req-observer-image"), sc.target(), imagesInput(t, sc), nil)
	ev.Check("success", err == nil, "got %v", err)
	if err != nil {
		close(rec.gate)
		return
	}
	res.Usage.Modalities.InputText = ai.Value(int64(999))
	ev.Check("result and attempt usage independent", res.Metadata.Attempts[0].Usage.Modalities.InputText != res.Usage.Modalities.InputText, "usage aliases")
	res.Metadata.Attempts[0].Usage.Modalities.InputText = ai.Value(int64(888))
	close(rec.gate)
	obs := rec.awaitCall(ev, res.Metadata.RequestID)
	ev.Record("observations", obs)
	for _, o := range obs {
		if o.Kind == ai.ObservationCallFinished {
			n, _ := o.Usage.Modalities.InputText.Get()
			ev.Check("observer owns usage", n == 10, "got %d", n)
			ev.Check("observer attribution", o.Call.Operation == ai.OperationImage && o.Call.API == ai.APIOpenAIImages, "got %+v", o.Call)
		}
		if o.Kind == ai.ObservationAttemptFinished {
			n, _ := o.Attempt.Usage.Modalities.InputText.Get()
			ev.Check("attempt observer owns usage", n == 10, "got %d", n)
		}
	}
	encoded := string(mustMarshal(t, obs))
	ev.Check("observer contains metadata only", !strings.Contains(encoded, imagesInput(t, sc).Prompt) && !strings.Contains(encoded, "b64_json") && !strings.Contains(encoded, "errorMessage"), "content in observer")
}

func TestOpenAIImagesExtensionsRegistered(t *testing.T) {
	ev := run.Case(t, "P08-E11-extensions-registered")
	r, ok := loadLedger(t).Route(string(ai.ProviderOpenAI), string(ai.APIOpenAIImages))
	ev.Record("route", r)
	ev.Check("native route registered independently", ok && r.Classification == pioracle.Extension && r.Ref != "", "got %+v", r)
	for _, name := range []string{"generation.json", "failures.json"} {
		f, _ := loadFixture(t, imagesProtocol, name)
		for _, sc := range f.Scenarios {
			ev.Check("explicit pi skip "+sc.ID, sc.PidiffSkip != "", "missing skip")
		}
	}
	ev.Record("catalog", ai.BuiltinCatalog())
	// Real inclusion is verified independently against sourced facts and this
	// route's own live response replay by TestOpenAIImagesBuiltinCatalog.
}

func TestOpenAIImagesUsageValidation(t *testing.T) {
	for _, name := range []string{"negative", "fractional", "overflow", "null-total", "missing-total", "missing-input-detail", "unknown-detail", "malformed-details", "unreported-null"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E11-usage-"+name)
			sc := firstImagesScenario(t)
			w := scenarioWorld(t, sc)
			var body map[string]any
			mustUnmarshal(t, []byte(sc.Replies[0].Body), &body)
			u := body["usage"].(map[string]any)
			switch name {
			case "negative":
				u["input_tokens"] = -1
			case "fractional":
				u["output_tokens"] = 0.5
			case "overflow":
				u["input_tokens"] = 1e30
			case "null-total":
				u["total_tokens"] = nil
			case "missing-total":
				delete(u, "total_tokens")
			case "missing-input-detail":
				delete(u, "input_tokens_details")
			case "unknown-detail":
				u["input_tokens_details"].(map[string]any)["audio_tokens"] = 5
			case "malformed-details":
				u["input_tokens_details"] = "wrong"
			case "unreported-null":
				body["usage"] = nil
			}
			enqueue(ev, w, fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: string(mustMarshal(t, body))}.script(t, imagesProtocol, ""))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-usage-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "negative" || name == "fractional" || name == "overflow" || name == "malformed-details" {
				ev.Check("invalid usage protocol failure", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Content) == 0, "got %v", err)
			} else {
				ev.Check("partial usage permitted", err == nil, "got %v", err)
				want := ai.UsagePartial
				if name == "unknown-detail" {
					want = ai.UsageComplete
				}
				if name == "unreported-null" {
					want = ai.UsageUnreported
				}
				ev.Check("one completeness axis", res.Metadata.Attempts[0].UsageReporting == want, "got %+v", res.Metadata.Attempts)
				if name == "missing-total" || name == "null-total" {
					ev.Check("missing total uses known sums", res.Usage.TotalTokens == 30, "got %d", res.Usage.TotalTokens)
				}
			}
		})
	}
}

func TestOpenAIImagesDecodedBudgetBoundary(t *testing.T) {
	for _, name := range []string{"at", "above"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E08-image-boundary-"+name)
			sc := firstImagesScenario(t)
			w, _ := imagesTrackedWorld(t, func(c *ai.Config) { c.Policy.Image.MaxOutputImageBytes = 12 })
			data := []byte("RIFF\x04\x00\x00\x00WEBP")
			if name == "above" {
				data = append(data, 0)
			}
			enqueue(ev, w, fixtureReply{Status: 200, Body: `{"output_format":"webp","data":[{"b64_json":"` + base64.StdEncoding.EncodeToString(data) + `"}]}`}.script(t, imagesProtocol, ""))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-boundary-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			if name == "at" {
				ev.Check("inclusive decoded budget", err == nil && len(res.Content) == 1, "got %v", err)
			} else {
				ev.Check("one decoded byte above refused", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseResponse}), "got %v", err)
			}
		})
	}
}
