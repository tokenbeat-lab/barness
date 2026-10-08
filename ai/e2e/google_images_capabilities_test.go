package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGoogleImagesCapabilities(t *testing.T) {
	for _, name := range []string{"foreign-options", "typed-nil", "openai-with-google-options", "wrong-operation", "wrong-provider", "v1", "query-key", "prefixed-model", "disabled-policy", "reference-input", "entry-snapshot", "header-key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E07-capability-"+name)
			sc := firstGoogleImagesScenario(t)
			options := ai.GoogleImagesOptions{ImageSize: ai.Value("1K")}
			w := newWorldWith(t, func(c *ai.Config) {
				configureGoogleImages(c)
				if name == "disabled-policy" {
					c.Policy.Image = nil
				}
				if name == "entry-snapshot" {
					resolver := c.Bindings
					c.Bindings = bindingResolverFunc(func(ctx context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
						options.ImageSize = ai.Value("512")
						return resolver.ResolveBinding(ctx, scope, id)
					})
				}
				if name == "prefixed-model" {
					c.Catalog.ImageModels[0].ID = "models/host-google-image"
				}
				if name == "openai-with-google-options" {
					configureImages(c)
				}
			}, tenantA)
			installGoogleImagesBinding(w, tenantA)
			w.updateBindingOf(tenantA, "images", func(b *ai.Binding) {
				switch name {
				case "wrong-operation":
					b.Operation = ai.OperationChat
				case "wrong-provider":
					b.ProviderID = ai.ProviderOpenAI
				case "v1":
					b.Endpoint = strings.TrimSuffix(b.Endpoint, "v1beta") + "v1"
				case "query-key":
					b.Endpoint += "?key=synthetic"
				case "prefixed-model":
					b.AllowedModels = []string{"models/host-google-image"}
				}
			})
			target := sc.target()
			if name == "prefixed-model" {
				target.ModelID = "models/host-google-image"
			}
			if name == "openai-with-google-options" {
				installImagesBinding(w, tenantA)
				target.ModelID = "host-image"
			}
			req := imagesInput(t, sc)
			if name == "reference-input" {
				req.ReferenceImages = []ai.Image{{MimeType: "image/png"}}
			}
			// Use the existing accepted image fixture as reference bytes; this route
			// still refuses references before credential access until issue 13.
			if name == "reference-input" {
				var content []ai.ImageOutputImage
				mustUnmarshal(t, sc.Expect.Content, &content)
				req.ReferenceImages[0].Data = content[0].Data
			}
			var opts ai.ImageOptions = &options
			if name == "foreign-options" {
				opts = ai.OpenAIImagesOptions{}
			}
			if name == "typed-nil" {
				opts = (*ai.GoogleImagesOptions)(nil)
			}
			h := ai.Hooks{}
			if name == "header-key" {
				h.TransformHeaders = func(_ context.Context, _ ai.CallScope, hdr http.Header) error {
					hdr.Set("X-Goog-Api-Key", "foreign")
					return nil
				}
			}
			enqueue(ev, w, sc.replies(t)...)
			id := "req-google-capability-" + name
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope(id), target, req, opts)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			if name == "entry-snapshot" {
				ev.Check("entry options independently frozen", err == nil && options.ImageSize == ai.Value("512"), "got %v", err)
				if reqs := w.provider.Requests(); len(reqs) == 1 {
					ev.Check("original image size sent", strings.Contains(string(reqs[0].Body), `"image_size":"1K"`), "got %s", reqs[0].Body)
				}
			} else {
				code := ai.CodeInvalidRequest
				phase := ai.PhaseCapability
				if name == "wrong-operation" {
					code = ai.CodeTenantDenied
				}
				if name == "disabled-policy" {
					phase = ai.PhaseScope
				}
				if name == "header-key" {
					code = ai.CodeTenantDenied
					phase = ai.PhaseRequest
				}
				ev.Check("no network on rejected capabilities", errors.Is(err, &ai.Error{Code: code, Phase: phase}) && len(w.provider.Requests()) == 0, "got %v", err)
				if name != "header-key" {
					ev.Check("no credential read", w.host.ReadsFor(id).Credentials == 0, "got %+v", w.host.ReadsFor(id))
				}
			}
		})
	}
}
