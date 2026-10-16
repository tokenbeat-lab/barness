package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMultimodalRuntimeInputLimits(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			for _, mode := range []string{"single", "total", "references", "expanded", "framing", "hook-file", "hook-total", "hook-references", "hook-expanded", "hook-body"} {
				if entry == "plain" && strings.HasPrefix(mode, "hook-") {
					continue
				}
				t.Run(r.id+"/"+entry+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "UC09-limits-"+r.id+"-"+entry+"-"+mode)
					req := multimodalInput(t, r, tenantA)
					var total, expanded, largest int64
					for _, img := range req.Images {
						raw, _ := base64.StdEncoding.DecodeString(img.Data)
						total += int64(len(raw))
						largest = max(largest, int64(len(raw)))
					}
					for _, ref := range req.ImageReferences {
						raw, _ := base64.StdEncoding.DecodeString(req.Images[ref.ImageIndex].Data)
						expanded += int64(len(raw))
					}
					// For framing, first measure a complete legal encoded request. Random
					// keys have fixed width. A second Client uses exactly one byte less.
					var nativeBytes int64
					if mode == "framing" {
						control := newMultimodalWorld(t, ev, nil)
						multimodalEnqueue(ev, control, r, tenantA, multimodalReply(t, ev, r, false))
						res, e := multimodalCall(ctxFor(t), control, r, tenantA, entry, "req-framing-control-"+r.id+entry, req, ai.Hooks{})
						checkMultimodalSuccess(t, ev, r, req, res, e)
						recordMultimodalCall(ev, control, req, res, e)
						control.released(ev)
						nativeBytes = int64(len(control.provider.Requests()[0].Body))
					}
					u := newMultimodalWorld(t, ev, func(c *ai.Config) {
						switch mode {
						case "single":
							c.Policy.MaxImageBytes = largest - 1
						case "total":
							c.Policy.Classifier.Images.MaxTotalImageBytes = total - 1
						case "references":
							c.Policy.Classifier.Images.References.MaxReferences = len(req.ImageReferences) - 1
						case "expanded":
							c.Policy.Classifier.Images.References.MaxExpandedImageBytes = expanded - 1
						case "framing":
							c.Policy.MaxRequestBytes = nativeBytes - 1
							c.Policy.MaxImageBytes = nativeBytes - 1
							c.Policy.Classifier.Images.MaxTotalImageBytes = nativeBytes - 1
						}
					})
					hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						if !strings.HasPrefix(mode, "hook-") {
							return ai.KeepPayload(), nil
						}
						if mode == "hook-body" {
							if r.api == ai.APIOpenAIDecisions {
								p.Body["input"] = strings.Repeat("x", 128<<10)
							} else {
								p.Body["extra"] = strings.Repeat("x", 128<<10)
							}
							return ai.KeepPayload(), nil
						}
						images := append([]ai.Image(nil), req.Images...)
						if mode == "hook-file" {
							images[0].Data = base64.StdEncoding.EncodeToString(make([]byte, (32<<10)+1))
						}
						if mode == "hook-total" {
							images[0].Data = base64.StdEncoding.EncodeToString(make([]byte, 40<<10))
							images[1] = images[0]
						}
						if r.api == ai.APIOpenAIDecisions {
							p.Body["input"] = classifierImageMessages(images)
							if mode == "hook-references" || mode == "hook-expanded" {
								n := 9
								if mode == "hook-expanded" {
									n = 8
									images[0].Data = base64.StdEncoding.EncodeToString(make([]byte, 20<<10))
								}
								for range n {
									images = append(images, images[0])
								}
								images = images[2:]
								p.Body["input"] = classifierImageMessages(images)
							}
						} else {
							assets := p.Body["assets"].(map[string]any)
							var keys []string
							for key, a := range assets {
								keys = append(keys, key)
								if mode == "hook-file" {
									a.(map[string]any)["data_base64"] = images[0].Data
								}
								if mode == "hook-total" {
									a.(map[string]any)["data_base64"] = images[0].Data
								}
							}
							if mode == "hook-references" || mode == "hook-expanded" {
								n := 9
								if mode == "hook-expanded" {
									n = 8
									assets[keys[0]].(map[string]any)["data_base64"] = base64.StdEncoding.EncodeToString(make([]byte, 20<<10))
								}
								repeated := make([]string, n)
								for i := range repeated {
									repeated[i] = keys[0]
								}
								p.Body["state"] = map[string]any{"state": "judge", "images": repeated}
							}
						}
						return ai.KeepPayload(), nil
					}}
					res, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-limit-"+r.id+entry+mode, req, hooks)
					phase, credentials := ai.PhaseScope, 0
					if strings.HasPrefix(mode, "hook-") || (mode == "framing" && r.api == ai.APIOpenAIDecisions) {
						phase, credentials = ai.PhaseRequest, 1
					}
					ev.Check("phase-specific finite budget", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}) && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0 && u.host.ReadsFor(res.Metadata.RequestID).Credentials == credentials, "got %v", err)
					ev.Record("limits-input", map[string]any{"sourceBytes": total, "expandedBytes": expanded, "largest": largest, "nativeBytes": nativeBytes, "mode": mode})
					recordMultimodalCall(ev, u, req, res, err)
					u.released(ev)
				})
			}
		}
	}
}

func TestMultimodalRuntimeOutputLimits(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			for _, mode := range []string{"exact", "over", "error-over"} {
				t.Run(r.id+"/"+entry+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "UC09-output-"+r.id+"-"+entry+"-"+mode)
					var meter *unaryMeter
					u := newMultimodalWorld(t, ev, func(c *ai.Config) {
						meter = &unaryMeter{next: c.Transport, held: func() bool { return c.Probe.Permits() == 1 }}
						c.Transport = meter
						c.Policy.MaxErrorBodyBytes = 1024
					})
					b := u.host.Binding(tenantA.tenant, r.id)
					b.Retry.MaxRetries = 2
					u.host.PutBinding(b)
					req := multimodalInput(t, r, tenantA)
					reply := multimodalReply(t, ev, r, false)
					callback := reply.OnRequest
					limit := 16 << 10
					reply.OnRequest = func(capture provider.Request) {
						callback(capture)
						n := limit
						if mode == "over" {
							n++
						}
						reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", n-len(reply.Chunks[0])))...)
						ev.Fixture(fmt.Sprintf("response-%s-%x.json", r.id, sha256.Sum256(capture.Body)), reply.Chunks[0])
					}
					if mode == "error-over" {
						limit = 1024
						reply = provider.Reply{Status: 503, Chunks: [][]byte{[]byte(strings.Repeat("x", 1025))}, Script: strings.Repeat("x", 1025)}
					}
					multimodalEnqueue(ev, u, r, tenantA, reply)
					res, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-output-"+r.id+entry+mode, req, ai.Hooks{})
					if mode == "exact" {
						checkMultimodalSuccess(t, ev, r, req, res, err)
						ev.Check("inclusive response bound", meter.read.Load() == int64(limit), "got %d", meter.read.Load())
					} else {
						phase := ai.PhaseResponse
						if mode == "error-over" {
							phase = ai.PhaseRequest
						}
						checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, phase, ai.StopReasonError)
						ev.Check("at most one probe byte", meter.read.Load() == int64(limit+1), "got %d", meter.read.Load())
					}
					ev.Check("one Attempt, permit through body close", len(res.Metadata.Attempts) == 1 && meter.heldAtClose.Load(), "replayed or released early")
					recordMultimodalCall(ev, u, req, res, err)
					ev.Record("body-meter", map[string]any{"limit": limit, "read": meter.read.Load(), "heldAtClose": meter.heldAtClose.Load()})
					u.released(ev)
				})
			}
		}
	}
}
