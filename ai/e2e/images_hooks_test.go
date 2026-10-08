package e2e

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func firstImagesScenario(t *testing.T) fixtureScenario {
	t.Helper()
	f, _ := loadFixture(t, imagesProtocol, "generation.json")
	return f.Scenarios[0]
}

func TestOpenAIImagesHooks(t *testing.T) {
	cases := []string{"final-format", "response-format", "final-n", "invalid-n", "nil-body", "bad-number", "long-prompt", "huge-prompt", "model", "header", "callback-error", "response-error", "retry", "default-no-retry"}
	// Every forbidden field is checked independently; false/null still adds
	// the forbidden field rather than authorizing a different operation.
	for _, key := range []string{"stream", "partial_images", "response_format", "user", "images", "image", "mask", "input_fidelity", "file_id", "image_url", "previous_response_id", "store", "tools", "endpoint", "operation", "unknown"} {
		cases = append(cases, "field-"+key)
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E04-hooks-"+name)
			sc := firstImagesScenario(t)
			w := scenarioWorld(t, sc)
			if name == "retry" {
				w.updateBindingOf(tenantA, "images", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			}
			payloads, responses, headers := 0, 0, 0
			sentinel := errors.New("synthetic host cause")
			h := ai.Hooks{TransformHeaders: func(_ context.Context, s ai.CallScope, header http.Header) error {
				headers++
				ev.Check("image header scope and hidden credential", s.Operation == ai.OperationImage && header.Get("Authorization") == "[REDACTED]", "got %+v", s)
				if name == "header" {
					header.Set("Authorization", "Bearer foreign")
				}
				return nil
			}, OnPayload: func(_ context.Context, s ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				ev.Check("image payload identity", s.Operation == ai.OperationImage && p.Operation == ai.OperationImage && p.API == ai.APIOpenAIImages && p.ModelID == sc.Model, "got %+v", p)
				switch name {
				case "final-format", "response-format":
					p.Body["output_format"] = "jpeg"
				case "final-n":
					p.Body["n"] = 2
				case "invalid-n":
					p.Body["n"] = 11
				case "bad-number":
					p.Body["n"] = "2"
				case "nil-body":
					return ai.ReplacePayload(nil), nil
				case "long-prompt":
					p.Body["prompt"] = strings.Repeat("字", 32001)
				case "huge-prompt":
					p.Body["prompt"] = strings.Repeat("x", 2<<20)
				case "model":
					p.Body["model"] = "unauthorized"
				case "callback-error":
					return ai.KeepPayload(), sentinel
				}
				if strings.HasPrefix(name, "field-") {
					p.Body[strings.TrimPrefix(name, "field-")] = false
				}
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, s ai.CallScope, r ai.ResponseInfo) error {
				responses++
				ev.Check("image response metadata", s.Operation == ai.OperationImage && r.Operation == ai.OperationImage && r.Status == 200 && r.Header.Get("x-request-id") == "req-images", "got %+v", r)
				if name == "response-error" {
					return sentinel
				}
				return nil
			}}
			reply := sc.replies(t)[0]
			if name == "response-format" {
				reply.Chunks[0] = []byte(strings.Replace(sc.Replies[0].Body, `"data":`, `"output_format":"png","data":`, 1))
				reply.Script = string(reply.Chunks[0])
			}
			if name == "final-format" {
				reply = fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: `{"data":[{"b64_json":"/9j/2Q=="}]}`}.script(t, imagesProtocol, "")
			}
			if name == "final-n" {
				var body map[string]any
				mustUnmarshal(t, []byte(sc.Replies[0].Body), &body)
				data := body["data"].([]any)
				body["data"] = append(data, data[0])
				reply = fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: string(mustMarshal(t, body))}.script(t, imagesProtocol, "")
			}
			if name == "retry" || name == "default-no-retry" {
				enqueue(ev, w, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":{"message":"limited"}}`}.script(t, imagesProtocol, ""))
			}
			enqueue(ev, w, reply)
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-hooks-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			want := ai.CodeCallbackFailed
			switch name {
			case "final-format", "response-format", "final-n", "retry":
				want = ""
			case "model", "header":
				want = ai.CodeTenantDenied
			case "huge-prompt":
				want = ai.CodeResourceLimit
			case "default-no-retry":
				want = ai.CodeRateLimited
			}
			if strings.HasPrefix(name, "field-") && name != "field-unknown" {
				want = ai.CodeTenantDenied
			}
			if want == "" {
				ev.Check("successful allowed callback", err == nil, "got %v", err)
			} else {
				ev.Check("classified callback", errors.Is(err, &ai.Error{Code: want, Phase: ai.PhaseRequest}) && len(res.Content) == 0, "got %v", err)
			}
			ev.Check("one header and payload outside retry", headers == 1 && payloads == map[bool]int{true: 0, false: 1}[name == "header"], "got %d/%d", headers, payloads)
			if name == "callback-error" || name == "response-error" {
				ev.Check("host cause preserved", errors.Is(err, sentinel), "got %v", err)
			}
			if name == "final-format" {
				img, ok := res.Content[0].(ai.ImageOutputImage)
				ev.Check("final request determines MIME", ok && img.MimeType == "image/jpeg", "got %+v", res.Content)
			}
			if name == "response-format" {
				img, ok := res.Content[0].(ai.ImageOutputImage)
				ev.Check("response format takes precedence", ok && img.MimeType == "image/png", "got %+v", res.Content)
			}
			if name == "final-n" {
				ev.Check("final quantity frozen", len(res.Content) == 2, "got %+v", res.Content)
			}
			if name == "retry" {
				ev.Check("two frozen attempts one response callback", len(w.provider.Requests()) == 2 && len(res.Metadata.Attempts) == 2 && responses == 1 && jsonEqual(w.provider.Requests()[0].Body, w.provider.Requests()[1].Body), "got %+v", res.Metadata)
			}
			if want != "" && name != "response-error" && name != "default-no-retry" {
				ev.Check("rejected before HTTP", len(w.provider.Requests()) == 0 && responses == 0, "got %d", len(w.provider.Requests()))
			}
		})
	}
}

func TestOpenAIImagesOptionsAndInput(t *testing.T) {
	for _, name := range []string{"max-prompt", "long-prompt", "request-budget", "input-count", "input-bytes", "invalid-utf8", "custom-size", "bad-custom-size", "unsupported-transparency", "policy-quantity", "model-quantity", "compression-zero", "typed-nil-options"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E07-input-"+name)
			sc := firstImagesScenario(t)
			w := newWorldWith(t, func(c *ai.Config) {
				configureImages(c)
				if name == "custom-size" || name == "bad-custom-size" {
					c.Catalog.ImageModels[0].Capabilities.CustomSizes = true
				}
				if name == "unsupported-transparency" {
					c.Catalog.ImageModels[0].Capabilities.TransparentBackground = false
				}
				if name == "policy-quantity" {
					c.Policy.Image.MaxOutputImages = 1
				}
				if name == "model-quantity" {
					c.Catalog.ImageModels[0].Capabilities.MaxOutputImages = 1
				}
			}, tenantA)
			installImagesBinding(w, tenantA)
			req, opts := imagesInput(t, sc), ai.OpenAIImagesOptions{}
			switch name {
			case "max-prompt":
				req.Prompt = strings.Repeat("字", 32000)
			case "long-prompt":
				req.Prompt = strings.Repeat("字", 32001)
			case "request-budget":
				req.Prompt = strings.Repeat("x", 2<<20)
			case "input-count":
				req.ReferenceImages = make([]ai.Image, 17)
			case "input-bytes":
				req.ReferenceImages = []ai.Image{{Data: strings.Repeat("A", 1<<20), MimeType: "image/png"}}
			case "invalid-utf8":
				req.Prompt = string([]byte{0xff})
			case "custom-size":
				opts.Size = ai.Value("1536x864")
			case "bad-custom-size":
				opts.Size = ai.Value("1537x864")
			case "unsupported-transparency":
				opts.Background = ai.Value("transparent")
			case "policy-quantity", "model-quantity":
				opts.N = ai.Value(2)
			case "compression-zero":
				opts.OutputFormat = ai.Value("jpeg")
				opts.OutputCompression = ai.Value(0)
			}
			reply := sc.replies(t)[0]
			if name == "compression-zero" {
				reply.Chunks[0] = []byte(strings.Replace(sc.Replies[0].Body, `"data":`, `"output_format":"png","data":`, 1))
				reply.Script = string(reply.Chunks[0])
			}
			enqueue(ev, w, reply)
			var options ai.ImageOptions = opts
			if name == "typed-nil-options" {
				options = (*ai.OpenAIImagesOptions)(nil)
			}
			res, err := w.client.GenerateImages(ctxFor(t), textScope("req-input-"+name), sc.target(), req, options)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "max-prompt" || name == "custom-size" || name == "compression-zero" {
				ev.Check("valid boundary succeeds", err == nil, "got %v", err)
			} else {
				var ae *ai.Error
				ev.Check("invalid input refused before credential and HTTP", errors.As(err, &ae) && len(w.provider.Requests()) == 0 && w.host.ReadsFor("req-input-"+name).Credentials == 0, "got %v", err)
			}
		})
	}
	t.Run("closed-options-and-output", func(t *testing.T) {
		ev := run.Case(t, "P08-E07-closed-types")
		options := reflect.TypeFor[ai.ImageOptions]()
		ev.Check("simple and chat options are excluded", !reflect.TypeFor[ai.SimpleOptions]().Implements(options) && !reflect.TypeFor[ai.ResponsesOptions]().Implements(options), "options leaked")
		ev.Check("chat image excluded from output", !reflect.TypeFor[ai.Image]().Implements(reflect.TypeFor[ai.ImageOutput]()), "output leaked")
		var decoded map[string]any
		mustUnmarshal(t, mustMarshal(t, ai.Usage{}), &decoded)
		_, present := decoded["modalities"]
		ev.Check("chat usage serialization unchanged", !present, "got %+v", decoded)
	})
}

// Track body closure and permits through the same controlled HTTP transport
// used by chat/classifier; there are no image-specific internal test hooks.
func imagesTrackedWorld(t *testing.T, configure func(*ai.Config)) (*world, *provider.BodyTracker) {
	t.Helper()
	b := provider.TrackBodies(provider.LoopbackTransport())
	w := newWorldWith(t, func(c *ai.Config) {
		configureImages(c)
		c.Transport = b
		if configure != nil {
			configure(c)
		}
	}, tenantA)
	installImagesBinding(w, tenantA)
	return w, b
}
