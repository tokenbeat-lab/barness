// Package provider is the local controlled Provider used by barness-ai E2E
// tests: a loopback HTTP server that replays scripted responses frame by frame
// and captures what it actually received (spec Testing Decisions §3).
//
// It stands in for the external vendor only. barness-ai's own request building,
// SDK and HTTP stack run unmodified against it.
package provider

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
)

// Reply is one scripted HTTP response. Chunks are written and flushed one by
// one, so framing (including byte-by-byte splits) reaches the client exactly as
// scripted.
type Reply struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header,omitempty"`
	Chunks [][]byte          `json:"-"`
	// Script is the concatenated body, kept for the evidence bundle.
	Script string `json:"script"`
	// OnReceive, when set, runs after the request is captured and before any
	// of the reply is written, so a test can change host state while a call is
	// verifiably in flight.
	OnReceive func() `json:"-"`
	// End is how the reply ends once every chunk is written.
	End End `json:"end,omitempty"`
	// OnHold, when set, runs once an EndHold reply has written its chunks and
	// starts holding the connection open.
	OnHold func() `json:"-"`
	// AfterChunk, when set, runs after chunk i was written and flushed, so a
	// test can pause the reply until its client reached a known state.
	AfterChunk func(i int) `json:"-"`
	// OnFinish observes handler completion, including failed writes. Raw errors
	// are intended only for audited synthetic E2E evidence.
	OnFinish func(ReplyOutcome) `json:"-"`
}

// ReplyOutcome distinguishes an interrupted fixture from a completed script.
type ReplyOutcome struct {
	RequestID        string `json:"requestId,omitempty"`
	End              End    `json:"end"`
	WrittenBytes     int64  `json:"writtenBytes"`
	Chunks           int    `json:"chunks"`
	WriteError       string `json:"writeError,omitempty"`
	ContextError     string `json:"contextError,omitempty"`
	WriteBufferWaits int64  `json:"writeBufferWaits,omitempty"`
}

// End selects how a reply ends after its chunks.
type End string

const (
	// EndClose finishes the body normally (the zero value).
	EndClose End = ""
	// EndAbort cuts the connection without terminating the chunked body, as
	// a crashed upstream or a dropped network path does.
	EndAbort End = "abort"
	// EndHold keeps the connection open, sending nothing more, until the
	// client goes away or the server is closed.
	EndHold End = "hold"
	// EndDrop cuts the connection after reading the request and before
	// writing any response, so the client sees a connection failure with no
	// HTTP status. Status, headers and chunks are ignored.
	EndDrop End = "drop"
	// EndSilent reads the request and then holds the connection open
	// without writing any response, not even headers, until the client goes
	// away or the server is closed: an upstream that accepted the request
	// and stalls. Status, headers and chunks are ignored.
	EndSilent End = "silent"
)

// Request is a redacted capture of what the server received.
type Request struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Query  string `json:"query"`
	// KeyAlias is the alias of the key received, as a bearer token, an
	// x-api-key header (Anthropic) or an x-goog-api-key header (Gemini);
	// "<none>" when absent and "<unknown>" when the key is not one the test
	// registered.
	KeyAlias string            `json:"key_alias"`
	Header   map[string]string `json:"header"`
	Body     json.RawMessage   `json:"body"`
}

// Server is a scripted loopback Provider.
type Server struct {
	srv       *httptest.Server
	aliases   map[string]string
	closing   chan struct{} // closed by Close; releases held replies
	closeOnce sync.Once

	mu       sync.Mutex
	replies  []Reply
	routes   map[string][]Reply // by path prefix; see EnqueueAt
	requests []Request
}

// New starts a server. aliases maps each test key the server may receive to
// the alias reported in captures; real key material never leaves the server.
func New(aliases map[string]string) *Server {
	return NewWithListener(aliases, nil)
}

// NewWithListener owns listener; nil uses httptest's loopback listener. It
// permits socket-fault scripts while the public Client's HTTP stack stays real.
func NewWithListener(aliases map[string]string, listener net.Listener) *Server {
	s := &Server{aliases: aliases, closing: make(chan struct{})}
	s.srv = httptest.NewUnstartedServer(http.HandlerFunc(s.handle))
	if listener != nil {
		_ = s.srv.Listener.Close()
		s.srv.Listener = listener
	}
	s.srv.Listener = writeBufferListener{s.srv.Listener}
	s.srv.Config.ConnContext = writeBufferContext
	s.srv.Start()
	return s
}

// URL is the server base URL (http://127.0.0.1:port).
func (s *Server) URL() string { return s.srv.URL }

// Close releases held replies and stops the server.
func (s *Server) Close() {
	s.closeOnce.Do(func() { close(s.closing) })
	s.srv.Close()
}

// Enqueue appends replies served to subsequent requests in order.
func (s *Server) Enqueue(replies ...Reply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replies = append(s.replies, replies...)
}

// EnqueueAt appends replies served, in order, to subsequent requests whose
// path starts with pathPrefix, whatever order requests to other prefixes
// arrive in. It lets concurrent callers configured with different endpoints
// on this one server each get their own script. A request matching no
// prefix takes from the Enqueue queue. Prefixes must not overlap.
func (s *Server) EnqueueAt(pathPrefix string, replies ...Reply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.routes == nil {
		s.routes = map[string][]Reply{}
	}
	s.routes[pathPrefix] = append(s.routes[pathPrefix], replies...)
}

// Requests returns the captures so far.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

// SSE builds a 200 text/event-stream reply from chunks.
func SSE(chunks [][]byte) Reply {
	var script strings.Builder
	for _, c := range chunks {
		script.Write(c)
	}
	return Reply{
		Status: http.StatusOK,
		Header: map[string]string{"Content-Type": "text/event-stream"},
		Chunks: chunks,
		Script: script.String(),
	}
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	capture := Request{
		Method:   r.Method,
		Path:     r.URL.Path,
		Query:    r.URL.RawQuery,
		KeyAlias: s.requestKeyAlias(r.Header),
		Header:   s.redactHeaders(r.Header),
		Body:     rawJSON(body),
	}

	s.mu.Lock()
	s.requests = append(s.requests, capture)
	reply, ok := s.nextReplyLocked(r.URL.Path)
	s.mu.Unlock()

	if !ok {
		http.Error(w, `{"error":{"message":"no scripted reply"}}`, http.StatusTeapot)
		return
	}
	outcome := ReplyOutcome{End: reply.End, RequestID: r.Header.Get("X-Barness-Test-Call")}
	waitsBefore := writeBufferWaits(r.Context())
	if reply.OnFinish != nil {
		defer func() {
			outcome.WriteBufferWaits = writeBufferWaits(r.Context()) - waitsBefore
			if err := r.Context().Err(); err != nil {
				outcome.ContextError = err.Error()
			}
			reply.OnFinish(outcome)
		}()
	}
	if reply.OnReceive != nil {
		reply.OnReceive()
	}
	if reply.End == EndDrop {
		// net/http drops the connection; nothing has been written yet.
		panic(http.ErrAbortHandler)
	}
	if reply.End == EndSilent {
		s.hold(r)
		return
	}
	for k, v := range reply.Header {
		w.Header().Set(k, v)
	}
	w.WriteHeader(reply.Status)
	flusher, _ := w.(http.Flusher)
	for i, chunk := range reply.Chunks {
		n, err := w.Write(chunk)
		outcome.WrittenBytes += int64(n)
		if err != nil {
			outcome.WriteError = err.Error()
			return
		}
		if flusher != nil {
			if err := http.NewResponseController(w).Flush(); err != nil {
				outcome.WriteError = err.Error()
				return
			}
		}
		outcome.Chunks++
		if reply.AfterChunk != nil {
			reply.AfterChunk(i)
		}
	}
	switch reply.End {
	case EndAbort:
		// net/http drops the connection without the terminating chunk.
		panic(http.ErrAbortHandler)
	case EndHold:
		if reply.OnHold != nil {
			reply.OnHold()
		}
		s.hold(r)
	}
}

// nextReplyLocked takes the reply for a request to path: from the EnqueueAt
// prefix of path, else from the Enqueue queue.
func (s *Server) nextReplyLocked(path string) (Reply, bool) {
	route, matched := "", false
	for prefix := range s.routes {
		if strings.HasPrefix(path, prefix) {
			route, matched = prefix, true
			break
		}
	}
	if matched {
		q := s.routes[route]
		if len(q) == 0 {
			return Reply{}, false
		}
		s.routes[route] = q[1:]
		return q[0], true
	}
	if len(s.replies) == 0 {
		return Reply{}, false
	}
	reply := s.replies[0]
	s.replies = s.replies[1:]
	return reply, true
}

// hold blocks until the client goes away or the server is closed.
func (s *Server) hold(r *http.Request) {
	select {
	case <-r.Context().Done():
	case <-s.closing:
	}
}

// requestKeyAlias is the alias of the request's credential: its bearer
// token, else its x-api-key, else its x-goog-api-key.
func (s *Server) requestKeyAlias(h http.Header) string {
	if key, ok := strings.CutPrefix(h.Get("Authorization"), "Bearer "); ok && key != "" {
		return s.alias(key)
	}
	if key := h.Get("X-Api-Key"); key != "" {
		return s.alias(key)
	}
	if key := h.Get("X-Goog-Api-Key"); key != "" {
		return s.alias(key)
	}
	return "<none>"
}

func (s *Server) keyAlias(authorization string) string {
	key, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || key == "" {
		return "<none>"
	}
	return s.alias(key)
}

func (s *Server) alias(key string) string {
	if alias, ok := s.aliases[key]; ok {
		return alias
	}
	return "<unknown>"
}

// redactHeaders flattens headers and replaces credentials with aliases.
func (s *Server) redactHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := strings.Join(h.Values(k), ", ")
		switch {
		case strings.EqualFold(k, "Authorization"):
			v = "Bearer " + s.keyAlias(v)
		case strings.EqualFold(k, "X-Api-Key"), strings.EqualFold(k, "X-Goog-Api-Key"):
			v = s.alias(v)
		}
		out[k] = v
	}
	return out
}

func rawJSON(body []byte) json.RawMessage {
	if json.Valid(body) {
		return body
	}
	quoted, _ := json.Marshal(string(body))
	return quoted
}
