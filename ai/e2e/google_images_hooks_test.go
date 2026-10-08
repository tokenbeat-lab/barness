package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGoogleImagesCallbackFixtures(t *testing.T) {
	raw, err := os.ReadFile("testdata/google-images/callbacks.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Source     string `json:"source"`
		PidiffSkip string `json:"pidiffSkip"`
		Cases      []struct {
			ID     string         `json:"id"`
			Set    map[string]any `json:"set"`
			Delete string         `json:"delete"`
			Code   ai.Code        `json:"code"`
		} `json:"cases"`
	}
	mustUnmarshal(t, raw, &f)
	for _, editing := range []bool{false, true} {
		for _, sc := range f.Cases {
			name := sc.ID
			if editing {
				name = "edit-guard-" + name
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P09-E04-"+name)
				ev.Fixture("callbacks.json", raw)
				ev.Check("explicit native differential skip", f.PidiffSkip != "", "missing native skip")
				input := firstGoogleImagesScenario(t)
				if editing {
					input = googleEditScenario(t)
				}
				w := scenarioWorld(t, input)
				enqueue(ev, w, input.replies(t)...)
				headers, payloads, responses := 0, 0, 0
				h := ai.Hooks{
					TransformHeaders: func(_ context.Context, scope ai.CallScope, header http.Header) error {
						headers++
						ev.Check("pinned Google auth redacted", scope.Operation == ai.OperationImage && header.Get("X-Goog-Api-Key") == "[REDACTED]" && header.Get("Authorization") == "", "got %+v", scope)
						return nil
					},
					OnPayload: func(_ context.Context, scope ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						payloads++
						ev.Check("Google image identity", scope.Operation == ai.OperationImage && p.Operation == ai.OperationImage && p.API == googleImagesProtocol.api && p.ModelID == input.Model, "got %+v", p)
						for k, v := range sc.Set {
							p.Body[k] = v
						}
						if sc.Delete != "" {
							delete(p.Body, sc.Delete)
						}
						return ai.KeepPayload(), nil
					},
					OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
						responses++
						ev.Check("authorized identity, independent of response model", r.Operation == ai.OperationImage && r.ModelID == input.Model && r.API == googleImagesProtocol.api && r.Status == 200, "got %+v", r)
						return nil
					},
				}
				res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-google-hook-"+name), input.target(), imagesInput(t, input), nil)
				ev.Record("result", res)
				ev.Record("requests", w.provider.Requests())
				ev.Record("error", errString(err))
				ev.Check("callbacks once", headers == 1 && payloads == 1, "got %d/%d", headers, payloads)
				if sc.Code == "" {
					ev.Check("allowed changes sent", err == nil && responses == 1 && len(w.provider.Requests()) == 1, "got %v", err)
					if reqs := w.provider.Requests(); len(reqs) == 1 {
						var body map[string]json.RawMessage
						mustUnmarshal(t, reqs[0].Body, &body)
						ev.Check("final options preserved", jsonEqual(body["response_format"], mustMarshal(t, sc.Set["response_format"])), "got %s", body["response_format"])
					}
				} else {
					ev.Check("reject before network", errors.Is(err, &ai.Error{Code: sc.Code, Phase: ai.PhaseRequest}) && responses == 0 && len(w.provider.Requests()) == 0 && len(res.Content) == 0, "got %v", err)
				}
			})
		}
	}
}

func TestGoogleImagesRawFinalGuards(t *testing.T) {
	for _, name := range []string{"raw-options", "duplicate-format", "duplicate-input", "raw-uri", "known-large-input", "encoded-large-input", "protocol-request-limit"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E04-final-"+name)
			input := firstGoogleImagesScenario(t)
			w := newWorldWith(t, func(c *ai.Config) {
				configureGoogleImages(c)
				if name == "protocol-request-limit" {
					c.Policy.MaxRequestBytes = 32 << 20
				}
			}, tenantA)
			installGoogleImagesBinding(w, tenantA)
			enqueue(ev, w, input.replies(t)...)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "raw-options":
					p.Body["response_format"] = json.RawMessage(`{"type":"image","image_size":"2K"}`)
				case "duplicate-format":
					p.Body["response_format"] = json.RawMessage(`{"type":"image","type":"text"}`)
				case "duplicate-input":
					p.Body["input"] = json.RawMessage(`[{"type":"text","text":"x","text":"y"}]`)
				case "raw-uri":
					p.Body["input"] = json.RawMessage(`[{"type":"image","uri":"https://example.invalid/image.png"}]`)
				case "known-large-input", "protocol-request-limit":
					p.Body["input"] = []any{map[string]any{"type": "text", "text": strings.Repeat("x", 21<<20)}}
				case "encoded-large-input":
					p.Body["input"] = []any{map[string]any{"type": "text", "text": strings.Repeat("\x00", 200000)}}
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-google-final-"+name), input.target(), imagesInput(t, input), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			code := ai.CodeCallbackFailed
			switch name {
			case "raw-options":
				code = ""
			case "raw-uri":
				code = ai.CodeTenantDenied
			case "known-large-input", "encoded-large-input", "protocol-request-limit":
				code = ai.CodeResourceLimit
			}
			if code == "" {
				ev.Check("allowed raw JSON frozen", err == nil, "got %v", err)
			} else {
				ev.Check("final payload refused", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0, "got %v", err)
			}
		})
	}
}
