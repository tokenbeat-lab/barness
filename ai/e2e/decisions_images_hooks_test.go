package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesHooks(t *testing.T) {
	assets := classifierAssets(t)
	for _, name := range []string{"add", "replace", "permission", "policy", "model", "mime-capability", "two-messages", "role", "text-last", "text-null", "text-extra", "detail", "bad-mime", "magic", "crlf", "external", "file", "later-external", "later-huge", "total", "count", "huge-name", "custom", "custom-external", "custom-huge", "whole-json", "state-limit", "typed-huge"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD26-images-hook-"+name)
			code := ai.CodeCallbackFailed
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				switch name {
				case "policy":
					c.Policy.Classifier.Images = nil
					code = ai.CodeTenantDenied
				case "model":
					m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
					m.Input = []ai.Modality{ai.ModalityText}
					m.Capabilities.MaxInputImages = 0
					m.Capabilities.ImageMIMETypes = nil
					code = ai.CodeTenantDenied
				case "mime-capability":
					c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.ImageMIMETypes = []string{"image/webp"}
					code = ai.CodeTenantDenied
				case "later-huge", "typed-huge", "custom-huge", "huge-name", "whole-json", "state-limit", "count", "total":
					code = ai.CodeResourceLimit
				}
				if name == "count" {
					c.Policy.Classifier.Images.MaxImages = 1
				}
				if name == "total" {
					raw, _ := base64.StdEncoding.DecodeString(assets[0].Data)
					c.Policy.Classifier.Images.MaxTotalImageBytes = int64(len(raw))
				}
				if name == "whole-json" {
					c.Policy.MaxRequestBytes = 1500
					c.Policy.MaxImageBytes = 1500
					c.Policy.Classifier.Images.MaxTotalImageBytes = 1500
				}
			})
			if name == "permission" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				code = ai.CodeTenantDenied
			}
			if name == "external" || name == "file" || name == "later-external" || name == "custom-external" {
				code = ai.CodeTenantDenied
			}
			if name == "add" || name == "replace" || name == "custom" {
				enqueueClassifierImages(t, ev, u)
			}
			messages := classifierImageMessages([]ai.Image{assets[0], assets[0]})
			message := messages[0].(map[string]any)
			content := message["content"].([]any)
			huge := "data:image/png;base64," + strings.Repeat("A", 8<<20)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "two-messages":
					messages = append(messages, message)
				case "role":
					message["role"] = "assistant"
				case "text-last":
					content[0], content[1] = content[1], content[0]
				case "text-null":
					content[0].(map[string]any)["text"] = nil
				case "text-extra":
					content = append(content, map[string]any{"type": "input_text", "text": "extra"})
					message["content"] = content
				case "detail":
					content[1].(map[string]any)["detail"] = "low"
				case "bad-mime":
					content[1].(map[string]any)["image_url"] = "data:image/gif;base64," + assets[0].Data
				case "magic":
					content[1].(map[string]any)["image_url"] = "data:image/jpeg;base64," + assets[0].Data
				case "crlf":
					content[1].(map[string]any)["image_url"] = "data:image/png;base64," + assets[0].Data + "\r\n"
				case "external", "later-external":
					index := 1
					if name == "later-external" {
						index = 2
					}
					content[index].(map[string]any)["image_url"] = "https://foreign.invalid/x.png"
				case "file":
					content[1].(map[string]any)["file_id"] = "file-foreign"
				case "later-huge", "custom-huge":
					content[2].(map[string]any)["image_url"] = huge
				case "huge-name":
					p.Body["questions"].([]any)[0].(map[string]any)["name"] = strings.Repeat("x", 5000)
				case "state-limit":
					content[0].(map[string]any)["text"] = strings.Repeat("x", 5000)
				case "whole-json":
					content[0].(map[string]any)["text"] = strings.Repeat("\x01", 300)
				}
				p.Body["input"] = messages
				if name == "typed-huge" {
					p.Body["input"] = []struct {
						Role    string `json:"role"`
						Content []struct {
							Type string  `json:"type"`
							URL  *string `json:"image_url,omitempty"`
							Text string  `json:"text,omitempty"`
						} `json:"content"`
					}{{Role: "user", Content: []struct {
						Type string  `json:"type"`
						URL  *string `json:"image_url,omitempty"`
						Text string  `json:"text,omitempty"`
					}{{Type: "input_text", Text: "ok"}, {Type: "input_image", URL: &huge}}}}
				}
				if strings.HasPrefix(name, "custom") {
					if name == "custom-external" {
						content[2].(map[string]any)["image_url"] = "https://foreign.invalid/x"
					}
					p.Body["input"] = decisionsRaw(mustMarshal(t, messages))
				}
				if name == "replace" {
					return ai.ReplacePayload(p.Body), nil
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-hook-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			runtime.ReadMemStats(&after)
			if name == "add" || name == "replace" || name == "custom" {
				ev.Check("authorized hook adds images", err == nil && len(res.Answers) == 1, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("rejected hook never sends", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			}
			if name == "later-huge" || name == "typed-huge" {
				ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
				ev.Check("known hook image bounded before encoding", after.TotalAlloc-before.TotalAlloc < 2<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsImagesWholeJSONBoundary(t *testing.T) {
	img := classifierAssets(t)[0]
	var wireBytes int
	for _, name := range []string{"measure", "at", "over"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD19-images-whole-json-"+name)
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				if name != "measure" {
					n := wireBytes
					if name == "over" {
						n--
					}
					c.Policy.MaxRequestBytes = int64(n)
					c.Policy.MaxImageBytes = 128
					c.Policy.Classifier.Images.MaxTotalImageBytes = 128
				}
			})
			req := decisionsInput()
			req.Images = []ai.Image{img}
			req.State = json.RawMessage(`"\u0001\u0001"`)
			if name != "over" {
				enqueueClassifierImages(t, ev, u)
			}
			res, err := u.client.Classify(ctxFor(t), textScope("req-image-json-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			if name == "over" {
				checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
			} else {
				if ev.Check("whole encoded boundary accepted", err == nil, "got %v", err) {
					wireBytes = len(u.provider.Requests()[0].Body)
					ev.Record("whole-json-bytes", wireBytes)
				}
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
