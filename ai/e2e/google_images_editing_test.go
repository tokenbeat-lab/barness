package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func googleEditScenario(t *testing.T) fixtureScenario {
	t.Helper()
	f, _ := loadFixture(t, googleImagesProtocol, "editing.json")
	return f.Scenarios[0]
}

func TestGoogleImagesEditingCallbacks(t *testing.T) {
	raw, err := os.ReadFile("testdata/google-images/editing-callbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			ID    string          `json:"id"`
			Input json.RawMessage `json:"input"`
			Code  ai.Code         `json:"code"`
		} `json:"cases"`
	}
	mustUnmarshal(t, raw, &f)
	for _, sc := range f.Cases {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P09-E04-edit-"+sc.ID)
			ev.Fixture("editing-callbacks.json", raw)
			input := googleEditScenario(t)
			w := scenarioWorld(t, input)
			enqueue(ev, w, input.replies(t)...)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["input"] = sc.Input
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-google-edit-hook-"+sc.ID), input.target(), imagesInput(t, input), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			if sc.Code == "" {
				ev.Check("final references sent", err == nil && len(w.provider.Requests()) == 1, "got %v", err)
				if reqs := w.provider.Requests(); len(reqs) == 1 {
					var body map[string]json.RawMessage
					mustUnmarshal(t, reqs[0].Body, &body)
					ev.Check("final input independent and ordered", jsonEqual(body["input"], sc.Input), "got %s", body["input"])
				}
			} else {
				ev.Check("final edit refused before network", errors.Is(err, &ai.Error{Code: sc.Code, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0 && len(res.Content) == 0, "got %v", err)
			}
		})
	}
}

func TestGoogleImagesEditingBudgets(t *testing.T) {
	large := base64.StdEncoding.EncodeToString(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 1500000)...))
	decimal := base64.StdEncoding.EncodeToString(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 15000000)...))
	for _, name := range []string{"host-count", "model-count", "protocol-count", "image-bytes", "request-bytes", "encoding-overhead", "protocol-bytes", "fourteen-accepted", "decimal-protocol-bytes", "exact-request-bytes", "one-byte-over-request"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E07-edit-budget-"+name)
			sc := googleEditScenario(t)
			budget := int64(len(mustMarshal(t, sc.Expect.Request.Body)))
			w := newWorldWith(t, func(c *ai.Config) {
				configureGoogleImages(c)
				switch name {
				case "host-count":
					c.Policy.Image.MaxInputImages = 1
				case "model-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 1
				case "image-bytes":
					c.Policy.MaxImageBytes = 16
				case "encoding-overhead":
					c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes = 1024, 512
				case "exact-request-bytes", "one-byte-over-request":
					c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes = budget, 100
					if name == "one-byte-over-request" {
						c.Policy.MaxRequestBytes--
					}
				case "protocol-bytes", "decimal-protocol-bytes":
					c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes = 64<<20, 16<<20
				case "protocol-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 16
				}
			}, tenantA)
			installGoogleImagesBinding(w, tenantA)
			req := imagesInput(t, sc)
			switch name {
			case "protocol-count", "fourteen-accepted", "protocol-bytes":
				count := 14
				if name == "protocol-count" {
					count++
				}
				ref := req.ReferenceImages[0]
				if name == "protocol-bytes" {
					ref.Data = large
				}
				req.ReferenceImages = make([]ai.Image, count)
				for i := range req.ReferenceImages {
					req.ReferenceImages[i] = ref
				}
			case "decimal-protocol-bytes":
				req.ReferenceImages = []ai.Image{{Data: decimal, MimeType: "image/png"}}
			case "request-bytes":
				req.ReferenceImages[0].Data = large
			case "encoding-overhead":
				req.Prompt = strings.Repeat("\"", 600)
			}
			enqueue(ev, w, sc.replies(t)...)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			id := "req-google-edit-budget-" + name
			res, err := w.client.GenerateImages(ctxFor(t), textScope(id), sc.target(), req, nil)
			runtime.ReadMemStats(&after)
			ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "fourteen-accepted" || name == "exact-request-bytes" {
				ev.Check("inclusive protocol count", err == nil && len(w.provider.Requests()) == 1, "got %v", err)
			} else {
				ev.Check("budgets fail before credentials and network", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit}) && w.host.ReadsFor(id).Credentials == 0 && len(w.provider.Requests()) == 0, "got %v reads=%+v", err, w.host.ReadsFor(id))
				ev.Check("known input rejected before encoded copies", after.TotalAlloc-before.TotalAlloc < 4<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			}
		})
	}
}

func TestGoogleImagesEditingSnapshot(t *testing.T) {
	ev := run.Case(t, "P09-E07-edit-snapshot")
	sc := googleEditScenario(t)
	req := imagesInput(t, sc)
	w := newWorldWith(t, func(c *ai.Config) {
		configureGoogleImages(c)
		bindings := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
			req.ReferenceImages[0] = ai.Image{Data: "caller mutation", MimeType: "text/plain"}
			return bindings.ResolveBinding(ctx, scope, id)
		})
	}, tenantA)
	installGoogleImagesBinding(w, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.GenerateImages(ctxFor(t), textScope("req-google-edit-snapshot"), sc.target(), req, nil)
	checkImagesScenario(ev, w, sc, res, err)
	ev.Check("caller mutation occurred", req.ReferenceImages[0].Data == "caller mutation", "mutation missing")
}

func TestGoogleImagesEditingTypedInput(t *testing.T) {
	for _, name := range []string{"promoted", "byte-data", "custom-key", "omitempty", "pointer-marshaler", "custom-zero", "pointer-zero"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E04-edit-typed-"+name)
			sc := googleEditScenario(t)
			req := imagesInput(t, sc)
			budget := int64(len(mustMarshal(t, sc.Expect.Request.Body)))
			w := newWorldWith(t, func(c *ai.Config) {
				configureGoogleImages(c)
				c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes = budget, 100
			}, tenantA)
			installGoogleImagesBinding(w, tenantA)
			enqueue(ev, w, sc.replies(t)...)
			first := req.ReferenceImages[0]
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				input := p.Body["input"].([]any)
				switch name {
				case "promoted":
					input[1] = googleDataWrapper{googleDataFields{first.Data}, "image", first.MimeType}
				case "byte-data":
					data, err := base64.StdEncoding.DecodeString(first.Data)
					if err != nil {
						return ai.KeepPayload(), err
					}
					input[1] = map[string]any{"type": "image", "mime_type": first.MimeType, "data": data}
				case "custom-key":
					input[1] = map[googleDataKey]string{0: "image", 1: first.MimeType, 2: first.Data}
				case "pointer-marshaler":
					value := struct {
						Type string            `json:"type"`
						Mime string            `json:"mime_type"`
						Data googleEncodedData `json:"data"`
					}{"image", first.MimeType, googleEncodedData(strings.Repeat("A", 8<<20))}
					input[1] = &value
				case "pointer-zero":
					input[1] = struct {
						Type string                   `json:"type"`
						Mime string                   `json:"mime_type"`
						Data string                   `json:"data"`
						Text googlePointerOmittedText `json:"text,omitzero"`
					}{"image", first.MimeType, first.Data, googlePointerOmittedText(strings.Repeat("x", 8192))}
				case "custom-zero":
					input[1] = struct {
						Type string            `json:"type"`
						Mime string            `json:"mime_type"`
						Data string            `json:"data"`
						Text googleOmittedText `json:"text,omitzero"`
					}{"image", first.MimeType, first.Data, googleOmittedText(strings.Repeat("x", 8192))}
				case "omitempty":
					input[1] = struct {
						Type string `json:"type"`
						Text string `json:"text,omitempty"`
						Mime string `json:"mime_type"`
						Data string `json:"data"`
					}{Type: "image", Mime: first.MimeType, Data: first.Data}
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-google-typed-"+name), sc.target(), req, nil)
			checkImagesScenario(ev, w, sc, res, err)
		})
	}
}

// These ordinary host types exercise encoding/json's method selection through
// the public callback, without calling private encoders or validation helpers.
type googleEncodedData string

func (*googleEncodedData) MarshalJSON() ([]byte, error) {
	return []byte(`"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR4nGNgAAIAAAUAAXpeqz8AAAAASUVORK5CYII="`), nil
}

type googleOmittedText string

func (googleOmittedText) IsZero() bool { return true }

type googlePointerOmittedText string

func (*googlePointerOmittedText) IsZero() bool { return true }
