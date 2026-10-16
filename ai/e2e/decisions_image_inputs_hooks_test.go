package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type imageInputsMessage struct {
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
		URL  string `json:"image_url,omitempty"`
	} `json:"content"`
}

func TestDecisionsImageInputsHooks(t *testing.T) {
	for _, name := range []string{"replace", "add", "reorder", "remove", "remove-all", "from-text", "escaped", "typed", "custom", "whole-file"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC03-hook-"+name)
			u := newReferenceWorld(t, ev, nil)
			req := decisionsInput()
			assets := classifierAssets(t)
			req.Images = assets[:2]
			finalImages := []ai.Image{assets[1], assets[0], assets[1]}
			switch name {
			case "replace":
				finalImages = assets[2:]
			case "add":
				finalImages = assets
			case "remove":
				finalImages = assets[1:2]
			case "remove-all":
				finalImages = nil
			case "from-text":
				req.Images = nil
			case "whole-file":
				raw, _ := base64.StdEncoding.DecodeString(assets[0].Data)
				// The existing protocol validates signatures without raster decoding.
				// Two files with the same PNG header must still have different hashes.
				tail := ai.Image{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(append(raw, 1, 2, 3, 4))}
				finalImages = []ai.Image{assets[0], tail}
				ev.Fixture("png-with-tail.bin", append(raw, 1, 2, 3, 4))
			}
			recordClassifierAssets(ev, assets)
			ev.Fixture("public-request.json", mustMarshal(t, req))
			ev.Record("hook-final-images", finalImages)
			enqueueClassifierImages(t, ev, u)
			var borrowed map[string]any
			res, err := u.client.WithHooks(ai.Hooks{
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					borrowed = p.Body
					input := any(classifierImageMessages(finalImages))
					if len(finalImages) == 0 {
						input = "track order"
					}
					switch name {
					case "escaped":
						raw := strings.ReplaceAll(string(mustMarshal(t, input)), "/", `\u002f`)
						input = json.RawMessage(raw)
					case "typed":
						var messages []imageInputsMessage
						mustUnmarshal(t, mustMarshal(t, input), &messages)
						input = messages
					case "custom":
						input = decisionsRaw(mustMarshal(t, input))
					}
					p.Body["input"] = input
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					borrowed["input"] = "changed after freeze"
					return nil
				},
			}).Classify(ctxFor(t), textScope("req-image-hook-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			ev.Check("legal final Hook input succeeds", err == nil, "got %v", err)
			checkImageInputs(ev, res, finalImages)
			requests := u.provider.Requests()
			if ev.Check("one final request", len(requests) == 1, "request missing") {
				var captured map[string]any
				mustUnmarshal(t, requests[0].Body, &captured)
				want := any(classifierImageMessages(finalImages))
				if len(finalImages) == 0 {
					want = "track order"
				}
				ev.Check("identity follows actual sent bytes despite borrowed mutation", reflect.DeepEqual(captured["input"], want), "final image changed")
			}
			recordImageInputsCall(ev, u, res, err)
		})
	}
}

func TestDecisionsImageInputsHookReject(t *testing.T) {
	for _, name := range []string{"mime", "signature", "newline", "base64", "suffix", "substring", "external", "file-id", "policy", "permission", "model", "model-count", "count", "total", "single", "body"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC03-hook-reject-"+name)
			assets := classifierAssets(t)
			code := ai.CodeCallbackFailed
			u := newReferenceWorld(t, ev, func(c *ai.Config) {
				switch name {
				case "policy":
					c.Policy.Classifier.Images = nil
				case "model":
					m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
					m.Input, m.Capabilities.ImageMIMETypes, m.Capabilities.MaxInputImages = []ai.Modality{ai.ModalityText}, nil, 0
				case "model-count":
					c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.MaxInputImages = 1
				case "count":
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 1, MaxExpandedImageBytes: 1 << 19}
				case "total":
					c.Policy.Classifier.Images.MaxTotalImageBytes = 100
				case "single":
					c.Policy.MaxImageBytes = 50
				case "body":
					c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes, c.Policy.Classifier.Images.MaxTotalImageBytes = 500, 500, 500
				}
			})
			if name == "permission" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
			}
			if name == "policy" || name == "permission" || name == "model" || name == "external" || name == "file-id" || name == "substring" {
				code = ai.CodeTenantDenied
			}
			if name == "model-count" || name == "count" || name == "total" || name == "single" || name == "body" {
				code = ai.CodeResourceLimit
			}
			req := decisionsInput()
			ev.Fixture("public-request.json", mustMarshal(t, req))
			recordClassifierAssets(ev, assets)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				input := classifierImageMessages([]ai.Image{assets[0], assets[0]})
				part := input[0].(map[string]any)["content"].([]any)[2].(map[string]any)
				url := part["image_url"].(string)
				switch name {
				case "mime":
					part["image_url"] = strings.Replace(url, "image/png", "image/gif", 1)
				case "signature":
					part["image_url"] = strings.Replace(url, "image/png", "image/jpeg", 1)
				case "newline":
					part["image_url"] = url + "\r\n"
				case "base64":
					part["image_url"] = url[:len(url)-1] + "!"
				case "suffix":
					part["image_url"] = url + "AAAA"
				case "substring":
					part["image_url"] = "prefix " + url
				case "external":
					part["image_url"] = "https://foreign.invalid/a.png"
				case "file-id":
					part["file_id"] = "foreign"
				}
				p.Body["input"] = input
				ev.Record("hook-input", input)
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-reject-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
			checkImageInputs(ev, res, nil)
			ev.Check("no network Attempt for invalid final input", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "sent invalid input")
			recordImageInputsCall(ev, u, res, err)
		})
	}
}
