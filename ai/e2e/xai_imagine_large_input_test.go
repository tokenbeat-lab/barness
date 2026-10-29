package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type imagineDirect struct {
	calls  int
	length int
	hash   string
}

func (r *imagineDirect) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++
	h := sha256.New()
	n, err := io.Copy(h, req.Body)
	r.length = int(n)
	r.hash = fmt.Sprintf("%x", h.Sum(nil))
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"` + imaginePNG + `"}]}`)), Request: req}, nil
}

// This bounded boundary case proves no OpenAI URL ceiling leaks into xAI.
// Record size/hash rather than duplicating the 22 MiB synthetic request.
func TestXAIImagineHostInputLimit(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-native-input-size")
	transport := &imagineDirect{}
	w := newImagineWorld(t, func(c *ai.Config) {
		c.Policy.MaxImageBytes = 24 << 20
		c.Policy.MaxRequestBytes = 32 << 20
		c.Transport = transport
	})
	data := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 16<<20)...)
	ref := ai.Image{Data: base64.StdEncoding.EncodeToString(data), MimeType: "image/png"}
	data = nil
	res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-host-limit"), imagineTarget(), ai.ImagesRequest{Prompt: "Edit.", ReferenceImages: []ai.Image{ref}}, nil)
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Record("request-size-hash", map[string]any{"bytes": transport.length, "sha256": transport.hash})
	ev.Check("Host permits larger xAI inline input", err == nil && transport.calls == 1 && transport.length > 20<<20, "err=%v calls=%d bytes=%d", err, transport.calls, transport.length)
}
