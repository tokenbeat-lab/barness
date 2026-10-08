package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

func TestOpenAIImagesEditingHooks(t *testing.T) {
	cases := []string{"format", "response-format", "mask", "reorder", "replace", "retry", "bad-response", "response-error", "remove", "empty", "null-images", "generation-to-edit", "generation-mask", "generation-fidelity", "host-count", "model-count", "protocol-count", "mask-count", "image-bytes", "mask-bytes", "url-length", "invalid-base64", "mask-size", "mask-unsupported", "fidelity-unsupported", "fidelity-invalid", "fidelity-null", "size", "transparent", "compression", "schema", "extra-image-field", "remote", "file-id", "mask-remote", "mask-file-id", "model", "unknown", "endpoint", "operation", "stream", "raw-file-id", "raw-duplicate-url", "raw-mask-null", "raw-count"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E04-edit-"+name)
			sc := editScenario(t)
			g := probe.New()
			w, b := imagesTrackedWorld(t, func(c *ai.Config) {
				c.Probe = g
				switch name {
				case "host-count", "mask-count":
					c.Policy.Image.MaxInputImages = 2
				case "model-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 2
				case "protocol-count", "url-length":
					c.Policy.Image.MaxInputImages = 32
					c.Policy.MaxImageBytes, c.Policy.MaxRequestBytes = 32<<20, 64<<20
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 32
				case "mask-unsupported":
					c.Catalog.ImageModels[0].Capabilities.Mask = false
				case "fidelity-unsupported":
					c.Catalog.ImageModels[0].Capabilities.InputFidelity = nil
				}
			})
			w.updateBindingOf(tenantA, "images", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			req := imagesInput(t, sc)
			if strings.HasPrefix(name, "generation-") {
				req.ReferenceImages = nil
			}
			ref := imagesInput(t, sc).ReferenceImages[0]
			url := "data:image/png;base64," + ref.Data
			payloads, responses := 0, 0
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				switch name {
				case "format", "response-format":
					p.Body["output_format"] = "jpeg"
				case "mask", "mask-count", "mask-unsupported", "generation-mask":
					p.Body["mask"] = map[string]any{"image_url": url}
				case "mask-size":
					p.Body["mask"] = map[string]any{"image_url": "data:image/png;base64," + req.ReferenceImages[1].Data}
				case "remove":
					delete(p.Body, "images")
				case "empty":
					p.Body["images"] = []any{}
				case "null-images":
					p.Body["images"] = nil
				case "reorder":
					images := p.Body["images"].([]any)
					images[0], images[1] = images[1], images[0]
				case "replace":
					return ai.ReplacePayload(map[string]any{"model": sc.Model, "prompt": "replacement", "images": []any{map[string]any{"image_url": url}}}), nil
				case "generation-to-edit":
					p.Body["images"] = []any{map[string]any{"image_url": url}}
				case "host-count", "model-count", "protocol-count":
					n := 3
					if name == "protocol-count" {
						n = 17
					}
					images := make([]any, n)
					for i := range images {
						images[i] = map[string]any{"image_url": url}
					}
					p.Body["images"] = images
				case "image-bytes", "url-length":
					n := 1 << 20
					if name == "url-length" {
						n = 20971520
					}
					p.Body["images"] = []any{map[string]any{"image_url": "data:image/png;base64," + strings.Repeat("A", n)}}
				case "mask-bytes":
					p.Body["mask"] = map[string]any{"image_url": "data:image/png;base64," + strings.Repeat("A", 1<<20)}
				case "invalid-base64":
					p.Body["images"] = []any{map[string]any{"image_url": "data:image/png;base64,!!!!"}}
				case "schema":
					p.Body["images"] = "bad schema"
				case "extra-image-field":
					p.Body["images"] = []any{map[string]any{"image_url": url, "detail": "high"}}
				case "remote":
					p.Body["images"] = []any{map[string]any{"image_url": "https://example.invalid/image.png"}}
				case "file-id":
					p.Body["images"] = []any{map[string]any{"image_url": url, "file_id": "file-fixture"}}
				case "raw-file-id":
					p.Body["images"] = json.RawMessage(`[{"file_id":"file-fixture"}]`)
				case "raw-duplicate-url":
					p.Body["images"] = json.RawMessage(`[{"image_url":"https://example.invalid/image.png","image_url":"` + url + `"}]`)
				case "raw-mask-null":
					p.Body["mask"] = json.RawMessage(`null`)
				case "raw-count":
					p.Body["images"] = json.RawMessage(`[` + strings.Repeat(`{"image_url":"`+url+`"},`, 16) + `{"image_url":"` + url + `"}]`)
				case "mask-remote":
					p.Body["mask"] = map[string]any{"image_url": "https://example.invalid/mask.png"}
				case "mask-file-id":
					p.Body["mask"] = map[string]any{"file_id": "file-fixture"}
				case "fidelity-unsupported", "generation-fidelity":
					p.Body["input_fidelity"] = "high"
				case "fidelity-invalid":
					p.Body["input_fidelity"] = "max"
				case "fidelity-null":
					p.Body["input_fidelity"] = nil
				case "size":
					p.Body["size"] = "1537x864"
				case "transparent":
					p.Body["background"], p.Body["output_format"] = "transparent", "jpeg"
				case "compression":
					p.Body["output_compression"] = 0
				case "model":
					p.Body["model"] = "foreign-model"
				case "endpoint", "operation", "stream", "unknown":
					p.Body[name] = false
				}
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
				responses++
				ev.Check("response metadata retained", r.Operation == ai.OperationImage && r.Header.Get("x-request-id") == "req-images", "got %+v", r)
				if name == "response-error" {
					return errors.New("synthetic response failure")
				}
				return nil
			}}
			reply := sc.replies(t)[0]
			if name == "format" {
				reply = fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: `{"data":[{"b64_json":"/9j/2Q=="}]}`}.script(t, imagesProtocol, "")
			}
			if name == "response-format" {
				reply = fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: strings.Replace(sc.Replies[0].Body, `"data":`, `"output_format":"png","data":`, 1)}.script(t, imagesProtocol, "")
			}
			if name == "bad-response" {
				reply = fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-images"}, Body: strings.Replace(sc.Replies[0].Body, `"data": [`, `"data": [{"b64_json":"!!!!"},`, 1)}.script(t, imagesProtocol, "")
			}
			if name == "retry" {
				enqueue(ev, w, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{}`}.script(t, imagesProtocol, ""))
			}
			enqueue(ev, w, reply)
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-edit-hooks-"+name), sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			want := ai.CodeCallbackFailed
			switch name {
			case "format", "response-format", "mask", "reorder", "replace", "retry":
				want = ""
			case "remove", "empty", "null-images", "generation-to-edit", "generation-mask", "generation-fidelity", "remote", "file-id", "mask-remote", "mask-file-id", "model", "endpoint", "operation", "stream", "raw-file-id":
				want = ai.CodeTenantDenied
			case "host-count", "model-count", "protocol-count", "mask-count", "image-bytes", "mask-bytes", "url-length", "raw-count":
				want = ai.CodeResourceLimit
			case "bad-response":
				want = ai.CodeProtocol
			}
			if want == "" {
				ev.Check("allowed final edit", err == nil && len(res.Content) == 1, "got %v", err)
			} else {
				ev.Check("invalid final edit", errors.Is(err, &ai.Error{Code: want}) && len(res.Content) == 0, "got %v", err)
			}
			n := 0
			if want == "" || name == "bad-response" || name == "response-error" {
				n = 1
			}
			if name == "retry" {
				n = 2
			}
			ev.Check("one callback, no unauthorized send or replay", payloads == 1 && len(w.provider.Requests()) == n && responses == min(n, 1), "payloads=%d responses=%d sends=%d", payloads, responses, len(w.provider.Requests()))
			if n > 0 && len(w.provider.Requests()) > 0 {
				ev.Check("fixed JSON edit endpoint", w.provider.Requests()[0].Path == "/v1/images/edits", "got %s", w.provider.Requests()[0].Path)
			}
			if name == "format" || name == "response-format" {
				if len(res.Content) > 0 {
					wantMIME := "image/jpeg"
					if name == "response-format" {
						wantMIME = "image/png"
					}
					ev.Check("final MIME", res.Content[0].(ai.ImageOutputImage).MimeType == wantMIME, "got %+v", res.Content)
				}
			}
			if name == "retry" && len(w.provider.Requests()) == 2 {
				ev.Check("frozen request retry", jsonEqual(w.provider.Requests()[0].Body, w.provider.Requests()[1].Body) && len(res.Metadata.Attempts) == 2, "got %+v", res.Metadata)
			}
			if name == "bad-response" && len(res.Metadata.Attempts) > 0 {
				ev.Check("usage retained on atomic failure", res.Usage.TotalTokens == 30 && res.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "got %+v", res)
			}
			ev.Record("resources", map[string]int64{"bodies": b.Open(), "calls": g.ActiveCalls(), "permits": g.Permits(), "waiters": g.AdmissionWaiters()})
			ev.Check("edit resources released", b.Open() == 0 && g.ActiveCalls() == 0 && g.Permits() == 0 && g.AdmissionWaiters() == 0, "resources remain")
		})
	}
}
