package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func xaiTestImage(t *testing.T, mime string) ai.Image {
	t.Helper()
	var b bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var err error
	if mime == "image/png" {
		err = png.Encode(&b, img)
	} else {
		err = jpeg.Encode(&b, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	return ai.Image{Data: base64.StdEncoding.EncodeToString(b.Bytes()), MimeType: mime}
}
func TestXAIImagesAndToolOrder(t *testing.T) {
	ev := run.Case(t, "XAI-E09-image-projection-order")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	png, jpg := xaiTestImage(t, "image/png"), xaiTestImage(t, "image/jpeg")
	a := ai.ToolCall{ID: xaiPublicPair(t, "a", "item-a"), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
	b := ai.ToolCall{ID: xaiPublicPair(t, "b", "item-b"), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
	req := xaiRequest()
	req.Messages = append(req.Messages, ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: "before"}, png, ai.Text{Text: "between"}, jpg}}, xaiTrustedMessage(t, a, b), ai.ToolResultMessage{ToolCallID: a.ID, ToolName: a.Name, Content: []ai.ToolResultContent{ai.Text{Text: "first result"}, png}}, ai.ToolResultMessage{ToolCallID: b.ID, ToolName: b.Name, Content: []ai.ToolResultContent{ai.Text{Text: "second result"}, jpg}})
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("images accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		return
	}
	var input []map[string]any
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
	if !ev.Check("each result retains a following image message", len(input) == 9, "%v", input) {
		return
	}
	for _, sc := range []struct {
		index            int
		call, mime, text string
	}{{5, "a", "image/png", "first result"}, {7, "b", "image/jpeg", "second result"}} {
		output, next := input[sc.index], input[sc.index+1]
		text, _ := output["output"].(string)
		content, _ := next["content"].([]any)
		part, _ := content[0].(map[string]any)
		ev.Check("tool image remains immediately after its own result", output["call_id"] == sc.call && strings.Contains(text, sc.text) && strings.Contains(text, "attached image") && next["role"] == "user" && strings.HasPrefix(part["image_url"].(string), "data:"+sc.mime+";base64,"), "output=%v next=%v", output, next)
	}
}

func TestXAIRejectsInvalidImages(t *testing.T) {
	for _, source := range []string{"user", "tool", "hook"} {
		for _, kind := range []string{"base64", "newline", "padding", "mime", "webp", "header-only", "external", "file"} {
			if source != "hook" && (kind == "external" || kind == "file") {
				continue
			}
			t.Run(source+"/"+kind, func(t *testing.T) {
				ev := run.Case(t, "XAI-E09-image-invalid-"+t.Name())
				w := newXAIWorld(t, nil)
				enqueue(ev, w, xaiReply(t))
				img := xaiTestImage(t, "image/png")
				switch kind {
				case "base64":
					img.Data = "!!!!"
				case "newline":
					img.Data += "\n"
				case "padding":
					img.Data = img.Data[:len(img.Data)-1]
				case "mime":
					img.MimeType = "image/jpeg"
				case "webp":
					img.MimeType = "image/webp"
				case "header-only":
					img.Data = base64.StdEncoding.EncodeToString([]byte("\x89PNG\r\n\x1a\n"))
				}
				req := xaiRequest()
				h := ai.Hooks{}
				if source == "user" {
					req.Messages = append(req.Messages, ai.UserMessage{Content: []ai.UserContent{img}})
				}
				if source == "tool" {
					call := ai.ToolCall{ID: xaiPublicPair(t, "a", "item"), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
					req.Messages = append(req.Messages, xaiTrustedMessage(t, call), ai.ToolResultMessage{ToolCallID: call.ID, Content: []ai.ToolResultContent{img}})
				}
				if source == "hook" {
					h.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						part := map[string]any{"type": "input_image", "image_url": "data:" + img.MimeType + ";base64," + img.Data}
						if kind == "external" {
							part["image_url"] = "https://example.invalid/image.png"
						}
						if kind == "file" {
							delete(part, "image_url")
							part["file_id"] = "foreign"
						}
						p.Body["input"] = []any{map[string]any{"role": "user", "content": []any{part}}}
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
				o.record(ev)
				ev.Check("invalid image refused without HTTP", o.err != nil && len(w.provider.Requests()) == 0, "err=%v", o.err)
			})
		}
	}
}

func TestXAIImageSizeBoundaries(t *testing.T) {
	for _, kind := range []string{"exact-20m", "over-20m", "host-smaller", "hook-host-smaller", "text-only-hook"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-E09-image-size-"+kind)
			img := xaiTestImage(t, "image/png")
			if kind == "exact-20m" || kind == "over-20m" {
				raw, _ := base64.StdEncoding.DecodeString(img.Data)
				size := 20 * 1024 * 1024
				if kind == "over-20m" {
					size++
				}
				raw = append(raw, make([]byte, size-len(raw))...)
				img.Data = base64.StdEncoding.EncodeToString(raw)
			}
			w := newXAIWorld(t, func(c *ai.Catalog) {
				if kind == "text-only-hook" {
					c.Models[0].Input = []ai.Modality{ai.ModalityText}
				}
			}, func(c *ai.Config) {
				c.Policy.MaxImageBytes = 21 * 1024 * 1024
				c.Policy.MaxRequestBytes = 32 * 1024 * 1024
				if kind == "host-smaller" || kind == "hook-host-smaller" {
					c.Policy.MaxImageBytes = 1
				}
			})
			enqueue(ev, w, xaiReply(t))
			req := xaiRequest()
			h := ai.Hooks{}
			if kind == "hook-host-smaller" || kind == "text-only-hook" {
				h.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					p.Body["input"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_image", "image_url": "data:" + img.MimeType + ";base64," + img.Data}}}}
					return ai.KeepPayload(), nil
				}
			} else {
				req.Messages = append(req.Messages, ai.UserMessage{Content: []ai.UserContent{img}})
			}
			o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
			o.record(ev)
			ev.Record("input-size", map[string]any{"encoded": len(img.Data), "requests": len(w.provider.Requests())})
			if kind == "exact-20m" {
				ev.Check("20MiB image fits both limits", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
			} else {
				ev.Check("smaller applicable image limit enforced", o.err != nil && len(w.provider.Requests()) == 0, "%v", o.err)
			}
		})
	}
}
