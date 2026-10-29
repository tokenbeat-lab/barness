package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIImagineDuplicateUsageFacts(t *testing.T) {
	cases := []struct {
		name, usage   string
		money         bool
		input, output int64
	}{
		{"duplicate-input", `"cost_in_usd_ticks":400000000,"input_tokens":10,"input_tokens":11,"output_tokens":8`, true, 0, 8},
		{"duplicate-ticks", `"cost_in_usd_ticks":400000000,"cost_in_usd_ticks":500000000,"input_tokens":10,"output_tokens":8`, false, 10, 8},
		{"duplicate-cache", `"cost_in_usd_ticks":400000000,"input_tokens":10,"output_tokens":8,"input_tokens_details":{"cached_tokens":1,"cached_tokens":2,"text_tokens":6,"image_tokens":4}`, true, 10, 8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-review-"+tc.name)
			w := newImagineWorld(t)
			enqueue(ev, w, imagineReply(`{"usage":{`+tc.usage+`},"data":[{"b64_json":"`+imaginePNG+`"}]}`))
			r, e := w.client.GenerateImages(ctxFor(t), textScope("review-"+tc.name), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, nil)
			ev.Record("result", r)
			ev.Record("error", errString(e))
			var failure *ai.Error
			ev.Check("ambiguous report fails atomically", errors.As(e, &failure) && failure.Code == ai.CodeProtocol && len(r.Content) == 0, "error=%v", e)
			if len(r.Metadata.Attempts) == 0 {
				return
			}
			amount, present := r.Metadata.Attempts[0].ReportedCostUSD.Get()
			ev.Check("independent unique charge retained", present == tc.money && (!present || amount == .04), "attempt=%+v", r.Metadata.Attempts[0])
			ev.Check("independent unique tokens retained", r.Usage.Input == tc.input && r.Usage.Output == tc.output, "usage=%+v", r.Usage)
			if tc.name == "duplicate-cache" {
				ev.Check("other unique detail counts retained", r.Usage.Modalities != nil && !r.Usage.Modalities.InputText.IsZero(), "usage=%+v", r.Usage)
			}
		})
	}
}
func TestXAIImagineTypedAggregateBudget(t *testing.T) {
	for _, name := range []string{"slice", "pointer-slice", "embedded", "raw-url-fields"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-review-typed-total-"+name)
			encoded := 0
			w := newImagineWorld(t, func(c *ai.Config) { c.Policy.MaxRequestBytes = 2048; c.Policy.MaxImageBytes = 1024 })
			url := "data:image/png;base64," + base64.StdEncoding.EncodeToString(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 792)...))
			type ref struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			}
			type embedded struct{ ref }
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				delete(p.Body, "image")
				p.Body["prompt"] = imagineCountedPrompt{&encoded}
				refs := []ref{{"image_url", url}, {"image_url", url}}
				switch name {
				case "slice":
					p.Body["images"] = refs
				case "pointer-slice":
					p.Body["images"] = &refs
				case "embedded":
					p.Body["images"] = []embedded{{refs[0]}, {refs[1]}}
				case "raw-url-fields":
					type rawref struct {
						Type string          `json:"type"`
						URL  json.RawMessage `json:"url"`
					}
					raw := json.RawMessage(`"` + url + `"`)
					p.Body["images"] = []rawref{{"image_url", raw}, {"image_url", raw}}
				}
				return ai.KeepPayload(), nil
			}}
			r, e := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("review-total-"+name), imagineTarget(), ai.ImagesRequest{Prompt: "Draw.", ReferenceImages: []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}}, nil)
			ev.Record("result", r)
			ev.Record("error", errString(e))
			ev.Record("host-encodings", encoded)
			var failure *ai.Error
			ev.Check("aggregate limit precedes Host encoding", errors.As(e, &failure) && failure.Code == ai.CodeResourceLimit && encoded == 0 && len(w.provider.Requests()) == 0, "err=%v encoded=%d", e, encoded)
		})
	}
}

type imagineCancelAfterWire struct {
	io.Reader
	once   sync.Once
	cancel context.CancelFunc
}

func (r *imagineCancelAfterWire) Read(p []byte) (int, error) {
	n, e := r.Reader.Read(p)
	if e == io.EOF {
		r.once.Do(func() { time.AfterFunc(time.Millisecond, r.cancel) })
	}
	return n, e
}
func (r *imagineCancelAfterWire) Close() error { return nil }

type imagineValidationTransport struct {
	body   string
	cancel context.CancelFunc
}

func (r imagineValidationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: &imagineCancelAfterWire{Reader: strings.NewReader(r.body), cancel: r.cancel}, Request: req}, nil
}

// EOF establishes that the complete bounded response was delivered. The
// delayed cancellation covers local image verification after wire delivery.
func TestXAIImagineCancelAfterWire(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-review-cancel-validation")
	ctx, cancel := context.WithCancel(ctxFor(t))
	defer cancel()
	data := base64.StdEncoding.EncodeToString(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32<<20)...))
	body := `{"usage":{"cost_in_usd_ticks":400000000},"data":[{"b64_json":"` + data + `"}]}`
	w := newImagineWorld(t, func(c *ai.Config) {
		c.Policy.MaxOutputBytes = 64 << 20
		c.Policy.Image.MaxOutputImageBytes = 40 << 20
		c.Policy.Image.MaxTotalOutputImageBytes = 40 << 20
		c.Transport = imagineValidationTransport{body: body, cancel: cancel}
	})
	r, e := w.client.GenerateImages(ctx, textScope("review-cancel-validation"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, nil)
	ev.Record("result", r)
	ev.Record("error", errString(e))
	var failure *ai.Error
	ev.Check("local verification retains cancellation", errors.As(e, &failure) && failure.Code == ai.CodeCanceled && r.StopReason == ai.StopReasonAborted && len(r.Content) == 0, "err=%v stop=%s", e, r.StopReason)
}
