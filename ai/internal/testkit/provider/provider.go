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
)

// Request is a redacted capture of what the server received.
type Request struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Query  string `json:"query"`
	// KeyAlias is the alias of the bearer key received; "<none>" when absent and
	// "<unknown>" when the key is not one the test registered.
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
	requests []Request
}

// New starts a server. aliases maps each test key the server may receive to
// the alias reported in captures; real key material never leaves the server.
func New(aliases map[string]string) *Server {
	s := &Server{aliases: aliases, closing: make(chan struct{})}
	s.srv = httptest.NewServer(http.HandlerFunc(s.handle))
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
		KeyAlias: s.keyAlias(r.Header.Get("Authorization")),
		Header:   s.redactHeaders(r.Header),
		Body:     rawJSON(body),
	}

	s.mu.Lock()
	s.requests = append(s.requests, capture)
	var reply Reply
	ok := len(s.replies) > 0
	if ok {
		reply = s.replies[0]
		s.replies = s.replies[1:]
	}
	s.mu.Unlock()

	if !ok {
		http.Error(w, `{"error":{"message":"no scripted reply"}}`, http.StatusTeapot)
		return
	}
	if reply.OnReceive != nil {
		reply.OnReceive()
	}
	if reply.End == EndDrop {
		// net/http drops the connection; nothing has been written yet.
		panic(http.ErrAbortHandler)
	}
	for k, v := range reply.Header {
		w.Header().Set(k, v)
	}
	w.WriteHeader(reply.Status)
	flusher, _ := w.(http.Flusher)
	for _, chunk := range reply.Chunks {
		if _, err := w.Write(chunk); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
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
		select {
		case <-r.Context().Done():
		case <-s.closing:
		}
	}
}

func (s *Server) keyAlias(authorization string) string {
	key, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || key == "" {
		return "<none>"
	}
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
		if strings.EqualFold(k, "Authorization") {
			v = "Bearer " + s.keyAlias(v)
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
