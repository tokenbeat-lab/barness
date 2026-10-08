//go:build live

package live

import (
	"bytes"
	"github.com/tokenbeat-lab/barness/ai"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// Capture limits: enough for the smoke's short turns and an error body,
// small enough that a runaway stream cannot fill the bundle.
const (
	maxCapturedRequest  = 256 << 10
	maxCapturedResponse = 512 << 10
)

// exchange is one HTTP request the Client sent and what came back, redacted
// at the source: credential headers are replaced and key-shaped text is
// masked before anything is stored.
type exchange struct {
	Method         string            `json:"method"`
	URL            string            `json:"url"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	RequestBody    string            `json:"requestBody,omitempty"`
	Status         int               `json:"status,omitempty"`
	// ResponseHeaderNames lists every response header; values are kept only
	// for ResponseHeaders' allowlist (ids, content type, rate limits).
	ResponseHeaderNames []string          `json:"responseHeaderNames,omitempty"`
	ResponseHeaders     map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody        string            `json:"responseBody,omitempty"`
	Truncated           bool              `json:"truncated,omitempty"`
	TransportError      string            `json:"transportError,omitempty"`
	ElapsedMS           int64             `json:"elapsedMs"`
}

// recorder is the Client's transport in a live process: the default
// transport (no proxy from the environment, as the Client's own default)
// with every exchange recorded.
type recorder struct {
	next http.RoundTripper

	mu           sync.Mutex
	exchanges    []*exchange
	observations []ai.Observation
	observed     chan struct{}
}

func newRecorder() *recorder {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = nil
	return &recorder{next: t, observed: make(chan struct{}, 1)}
}

// take returns the exchanges recorded since the last take.
func (r *recorder) take() []exchange {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]exchange, len(r.exchanges))
	for i, e := range r.exchanges {
		out[i] = *e
	}
	r.exchanges = nil
	return out
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	started := time.Now()
	e := &exchange{Method: req.Method, URL: redactURL(req), RequestHeaders: redactHeaders(req.Header)}
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
		e.RequestBody = redactText(string(body[:min(len(body), maxCapturedRequest)]))
		e.Truncated = len(body) > maxCapturedRequest
	}
	r.mu.Lock()
	r.exchanges = append(r.exchanges, e)
	r.mu.Unlock()

	res, err := r.next.RoundTrip(req)
	r.mu.Lock()
	defer r.mu.Unlock()
	e.ElapsedMS = time.Since(started).Milliseconds()
	if err != nil {
		e.TransportError = redactText(err.Error())
		return nil, err
	}
	e.Status = res.StatusCode
	for name := range res.Header {
		e.ResponseHeaderNames = append(e.ResponseHeaderNames, strings.ToLower(name))
		if keptHeaderValue.MatchString(name) {
			if e.ResponseHeaders == nil {
				e.ResponseHeaders = map[string]string{}
			}
			e.ResponseHeaders[strings.ToLower(name)] = res.Header.Get(name)
		}
	}
	slices.Sort(e.ResponseHeaderNames)
	res.Body = &capturingBody{ReadCloser: res.Body, rec: r, e: e, started: started}
	return res, nil
}

// capturingBody copies what the Client reads, up to the capture limit. A
// canceled call may close it while a read is in flight, so the copy and the
// exchange it fills are guarded by the recorder's lock.
type capturingBody struct {
	io.ReadCloser
	rec     *recorder
	e       *exchange
	started time.Time
	buf     bytes.Buffer
	once    sync.Once
}

func (b *capturingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.rec.mu.Lock()
	if room := maxCapturedResponse - b.buf.Len(); room > 0 {
		b.buf.Write(p[:min(n, room)])
		if n > room {
			b.e.Truncated = true
		}
	} else if n > 0 {
		b.e.Truncated = true
	}
	b.rec.mu.Unlock()
	if err != nil {
		b.flush()
	}
	return n, err
}

func (b *capturingBody) Close() error {
	b.flush()
	return b.ReadCloser.Close()
}

func (b *capturingBody) flush() {
	b.once.Do(func() {
		b.rec.mu.Lock()
		defer b.rec.mu.Unlock()
		b.e.ResponseBody = redactText(b.buf.String())
		b.e.ElapsedMS = time.Since(b.started).Milliseconds()
	})
}

// credentialHeader names request headers that carry a credential in any
// first-phase protocol.
var credentialHeader = regexp.MustCompile(`(?i)^(authorization|x-api-key|x-goog-api-key|api-key|cookie|proxy-authorization)$`)

// keptHeaderValue names response headers whose values are kept: vendor
// request and trace ids, content type, retry and rate-limit state. Others
// (organization, project, cookies) are recorded by name only.
var keptHeaderValue = regexp.MustCompile(`(?i)(request[-_]?id|trace[-_]?id|^content-type$|^retry-after|ratelimit)`)

// keyShaped matches vendor key forms, including masked echoes such as
// DeepSeek's "sk-****abcd" in a 401 body.
var keyShaped = regexp.MustCompile(`sk-[A-Za-z0-9_\-*]{4,}|AIza[0-9A-Za-z_\-]{20,}`)

func redactText(s string) string { return keyShaped.ReplaceAllString(s, "[REDACTED-KEY]") }

func redactHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for name, values := range h {
		v := strings.Join(values, ", ")
		if credentialHeader.MatchString(name) {
			v = "[REDACTED]"
		}
		out[strings.ToLower(name)] = redactText(v)
	}
	return out
}

func redactURL(req *http.Request) string {
	u := *req.URL
	q := u.Query()
	if q.Has("key") {
		q.Set("key", "[REDACTED]")
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// Observe stores only the public metadata record. Its JSON is independently
// audited with the existing Observer field allowlist.
func (r *recorder) Observe(o ai.Observation) error {
	r.mu.Lock()
	r.observations = append(r.observations, o)
	r.mu.Unlock()
	select {
	case r.observed <- struct{}{}:
	default:
	}
	return nil
}
func (r *recorder) observationsOf(id string) []ai.Observation {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for {
		r.mu.Lock()
		var out []ai.Observation
		finished := false
		for _, o := range r.observations {
			if o.Call.RequestID == id {
				out = append(out, o)
				finished = finished || o.Kind == ai.ObservationCallFinished
			}
		}
		r.mu.Unlock()
		if finished {
			return out
		}
		select {
		case <-r.observed:
		case <-timer.C:
			return out
		}
	}
}
