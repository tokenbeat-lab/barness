package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIImagineEditingAndOptions(t *testing.T) {
	for _, name := range []string{"single", "multiple", "multi-output", "six-inputs", "single-ratio", "typed-nil", "wrong-options", "null-quality", "bad-resolution", "bad-quality", "zero-n", "eleven-n", "empty-prompt", "wrong-mime", "bad-base64", "multi-ratio", "three-resolutions", "ten-output"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-input-"+name)
			w := newImagineWorld(t)
			req := ai.ImagesRequest{Prompt: "Edit ordered images."}
			var opts ai.ImageOptions = ai.XAIImagesOptions{N: ai.Value(1), Resolution: ai.Value("1k"), Quality: ai.Value("low")}
			o := opts.(ai.XAIImagesOptions)
			success := false
			n := 1
			switch name {
			case "single":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}
				success = true
			case "multiple", "multi-ratio":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/png"}, {Data: "/9j/2Q==", MimeType: "image/jpeg"}}
				success = true
				if name == "multi-ratio" {
					o.AspectRatio = ai.Value("16:9")
				}
			case "multi-output":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}
				o.N = ai.Value(2)
			case "six-inputs":
				for range 6 {
					req.ReferenceImages = append(req.ReferenceImages, ai.Image{Data: imaginePNG, MimeType: "image/png"})
				}
			case "single-ratio":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}
				o.AspectRatio = ai.Value("auto")
			case "typed-nil":
				var v *ai.XAIImagesOptions
				opts = v
			case "wrong-options":
				opts = ai.GoogleImagesOptions{}
			case "null-quality":
				o.Quality = ai.Null[string]()
			case "bad-resolution":
				o.Resolution = ai.Value("4k")
			case "bad-quality":
				o.Quality = ai.Value("high")
			case "zero-n":
				o.N = ai.Value(0)
			case "eleven-n":
				o.N = ai.Value(11)
			case "empty-prompt":
				req.Prompt = ""
			case "wrong-mime":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/jpeg"}}
			case "bad-base64":
				req.ReferenceImages = []ai.Image{{Data: imaginePNG + "\n", MimeType: "image/png"}}
			case "three-resolutions":
				o.Resolution = ai.Value("1.5k")
				success = true
			case "ten-output":
				o.N = ai.Value(10)
				success = true
				n = 10
			}
			if name != "typed-nil" && name != "wrong-options" {
				opts = o
			}
			data := make([]any, n)
			for i := range data {
				data[i] = map[string]any{"b64_json": imaginePNG}
			}
			enqueue(ev, w, imagineReply(string(mustMarshal(t, map[string]any{"data": data}))))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-input-"+name), imagineTarget(), req, opts)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Record("error", errString(err))
			ev.Check("entry contract", (err == nil) == success, "err=%v", err)
			if success {
				requests := w.provider.Requests()
				ev.Check("one request", len(requests) == 1, "count=%d", len(requests))
				if len(requests) == 1 {
					var wire map[string]json.RawMessage
					mustUnmarshal(t, requests[0].Body, &wire)
					if len(req.ReferenceImages) == 1 {
						ev.Check("single JSON resource", wire["image"] != nil && wire["images"] == nil, "wire=%s", requests[0].Body)
					}
					if len(req.ReferenceImages) > 1 {
						var refs []map[string]string
						mustUnmarshal(t, wire["images"], &refs)
						ev.Check("reference order", len(refs) == 2 && strings.Contains(refs[0]["url"], imaginePNG) && strings.Contains(refs[1]["url"], "/9j/2Q=="), "refs=%+v", refs)
					}
				}
			} else {
				ev.Check("no HTTP on invalid input", len(w.provider.Requests()) == 0, "requests=%+v", w.provider.Requests())
				ev.Check("no secret on invalid entry", w.host.ReadsFor("imagine-input-"+name).Credentials == 0, "reads=%+v", w.host.ReadsFor("imagine-input-"+name))
			}
		})
	}
}

func TestXAIImagineHookGuard(t *testing.T) {
	for _, name := range []string{"quality", "model", "response-format", "generate-to-edit", "deferred", "stream", "store", "unknown", "raw-duplicate", "null-prompt", "null-image", "external-url", "duplicate-resource", "typed-images", "over-count", "retry"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-hook-"+name)
			w := newImagineWorld(t)
			req := ai.ImagesRequest{Prompt: "Draw."}
			calls := 0
			success := name == "quality" || name == "typed-images" || name == "retry"
			o := ai.XAIImagesOptions{Resolution: ai.Value("1k"), Quality: ai.Value("low")}
			if name == "retry" {
				w.updateBindingOf(tenantA, "imagine", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				enqueue(ev, w, imagineReplyStatus(429, `{"error":{"message":"limited"}}`))
			}
			if name == "null-image" || name == "external-url" || name == "duplicate-resource" || name == "typed-images" || name == "over-count" {
				req.ReferenceImages = []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}
			}
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				calls++
				switch name {
				case "quality":
					p.Body["quality"] = "medium"
				case "model":
					p.Body["model"] = "foreign"
				case "response-format":
					p.Body["response_format"] = "url"
				case "generate-to-edit":
					p.Body["image"] = map[string]any{"type": "image_url", "url": "data:image/png;base64," + imaginePNG}
				case "deferred", "stream", "store", "unknown":
					p.Body[name] = false
				case "raw-duplicate":
					p.Body["n"] = json.RawMessage(`1,"n":2`)
				case "null-prompt":
					p.Body["prompt"] = nil
				case "null-image":
					p.Body["image"] = nil
				case "external-url":
					p.Body["image"] = map[string]any{"type": "image_url", "url": "https://example.invalid/a.png"}
				case "duplicate-resource":
					p.Body["image"] = json.RawMessage(`{"type":"image_url","url":"data:image/png;base64,` + imaginePNG + `","url":"data:image/png;base64,` + imaginePNG + `"}`)
				case "typed-images":
					delete(p.Body, "image")
					type image struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					}
					p.Body["images"] = []image{{"image_url", "data:image/png;base64," + imaginePNG}}
				case "over-count":
					delete(p.Body, "image")
					p.Body["images"] = make([]map[string]string, 6)
				}
				return ai.KeepPayload(), nil
			}}
			enqueue(ev, w, imagineReply(`{"data":[{"b64_json":"`+imaginePNG+`"}]}`))
			res, err := w.client.WithHooks(hooks).GenerateImages(ctxFor(t), textScope("imagine-hook-"+name), imagineTarget(), req, o)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Record("error", errString(err))
			ev.Check("Hook contract", (err == nil) == success, "err=%v", err)
			ev.Check("Hook called once", calls == 1, "calls=%d", calls)
			if !success {
				var ae *ai.Error
				ev.Check("reject before sending", errors.As(err, &ae) && len(w.provider.Requests()) == 0, "err=%v", err)
			}
			if name == "retry" {
				requests := w.provider.Requests()
				ev.Check("frozen retry", len(requests) == 2 && string(requests[0].Body) == string(requests[1].Body), "requests=%+v", requests)
			}
		})
	}
}
func imagineReplyStatus(status int, body string) provider.Reply {
	r := imagineReply(body)
	r.Status = status
	r.Header["retry-after-ms"] = "0"
	return r
}
