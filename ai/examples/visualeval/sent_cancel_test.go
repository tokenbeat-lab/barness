package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type cancelAtEOF struct {
	*bytes.Reader
	cancel context.CancelFunc
}

func (r cancelAtEOF) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.cancel()
	}
	return n, err
}

type readCancelTransport struct {
	body   []byte
	cancel context.CancelFunc
	calls  int
}

func (r *readCancelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++
	req.Body.Close()
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(cancelAtEOF{bytes.NewReader(r.body), r.cancel}), Request: req}, nil
}

func TestVisualSentCancellationBeforeDecode(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := unsentEvidence.Case(t, "VQ20-sent-before-decode-"+entry)
			p, err := load("decisions")
			if err != nil {
				t.Fatal(err)
			}
			a := arguments{mode: "controlled", route: "decisions", entry: entry, calls: 12, workMS: 225000, callMS: 30000}
			r := initial(p, a)
			b, err := newBundle(filepath.Join(unsentEvidence.Dir(), "sent-before-decode-"+entry), "synthetic-unsent-credential")
			if err != nil || b.inputs(p, a, nil, &r) != nil {
				t.Fatal("controlled inputs")
			}
			raw, _ := data.ReadFile("testdata/decisions-responses.json")
			var replies []scriptReply
			if strict(raw, &replies) != nil {
				t.Fatal("native response fixture")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			transport := &readCancelTransport{body: replies[0].Body, cancel: cancel}
			if err = executeSample(ctx, p, a, b, &r, 0, "http://127.0.0.1/v1", "synthetic-unsent-credential", "synthetic-account", transport, ai.Hooks{}); err != nil {
				t.Fatal(err)
			}
			for i := 1; i < len(r.Samples); i++ {
				r.Samples[i].Status = "cancelled"
			}
			finalize(p, &r)
			s := r.Samples[0]
			ev.Record("public-result", s)
			ev.Check("real response-read cancellation", transport.calls == 1 && r.Budget.Calls == 1 && r.Budget.Attempts == 1 && s.Error != nil && s.Error.Code == ai.CodeCanceled && s.Error.Phase == ai.PhaseResponse, "wrong public lifecycle")
			ev.Check("no decoded usage or model invented", same(s.Result.Usage, ai.Usage{}) && same(s.Result.ResponseModel, ai.Nullable[string]{}) && len(s.Result.Answers) == 0 && s.Result.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "decoded state leaked past cancellation")
			wireRaw, err := readBounded(filepath.Join(b.dir, "samples/v01/response.bin"), 64<<10)
			ev.Check("actual complete response bytes captured", err == nil && bytes.Equal(wireRaw, transport.body), "lost native response")
			err = b.finish(r)
			ev.Check("sent cancellation package audits", err == nil, "finish: %v", err)
			err = verify(b.dir, false)
			ev.Check("encoded absent model preserves honest cancellation", err == nil, "verify: %v", err)
			ev.Check("controlled failure never qualifies live", verify(b.dir, true) != nil, "failure promoted")
		})
	}
}
