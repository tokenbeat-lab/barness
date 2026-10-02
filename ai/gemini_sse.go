package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

// msgGeminiIncomplete is the Google SDK's text for a body that ends inside
// an event, which pi reports as is.
const msgGeminiIncomplete = "Incomplete JSON segment at the end"

// geminiDelimiters end an event in the Google SDK's decoder; the earliest
// one in the buffer wins.
var geminiDelimiters = [][]byte{[]byte("\n\n"), []byte("\r\r"), []byte("\r\n\r\n")}

// geminiDecoder ports the Google SDK's processStreamResponse (@google/genai
// 2.21.0), through which pi reads the stream: the body is cut into events
// at the earliest blank-line delimiter, each event is trimmed, and only an
// event starting with "data:" is a chunk, its trimmed rest being one JSON
// document. Event names, comments and other fields are skipped; a data
// event spread over several data: lines is not joined, so it fails to parse.
// A body that ends with anything but whitespace after its last delimiter
// fails with the SDK's text.
//
// The SDK also fails the stream when a network chunk as a whole is a JSON
// object with an "error" whose code is an HTTP error status. Network chunks
// are not observable here; the decoder applies the check to each event,
// which is the same whenever the provider writes an error as one chunk of
// its own (ADR-0012).
//
// streamBody ends a frame only at LF, so an event ended by "\r\r" counts
// toward the same frame as the next: the frame limit is reached early,
// never late.
type geminiDecoder struct {
	body io.Reader
	buf  []byte
	// scanned is how far buf was searched for a delimiter without finding
	// one, so each byte is searched about once.
	scanned int
	eof     bool
	data    []byte // the current chunk's JSON
	err     error  // a read error, or the stream failure the SDK raises
	read    []byte // the buffer each Read fills
	// upstream classifies a failure the provider reported in the stream,
	// redacting its text (httpFailures.upstream).
	upstream func(msg string) *Error
}

func newGeminiDecoder(body io.Reader, upstream func(string) *Error) *geminiDecoder {
	return &geminiDecoder{body: body, read: make([]byte, 32*1024), upstream: upstream}
}

// geminiStreamError is a failure the Google SDK raises while decoding.
type geminiStreamError struct{ failure *Error }

func (e *geminiStreamError) Error() string { return e.failure.Message }

// next advances to the next data chunk; false at the end of the body or on a
// failure (see err).
func (d *geminiDecoder) next() bool {
	for {
		if event, ok := d.cut(); ok {
			if d.errorChunk(event) {
				return false
			}
			trimmed := bytes.TrimFunc(event, isJSSpace)
			if rest, ok := bytes.CutPrefix(trimmed, []byte("data:")); ok {
				d.data = bytes.TrimFunc(rest, isJSSpace)
				return true
			}
			continue
		}
		if d.eof {
			if d.errorChunk(d.buf) {
				return false
			}
			if len(bytes.TrimFunc(d.buf, isJSSpace)) > 0 {
				d.err = &geminiStreamError{newError(CodeProtocol, PhaseStream, msgGeminiIncomplete)}
			}
			return false
		}
		n, err := d.body.Read(d.read)
		d.buf = append(d.buf, d.read[:n]...)
		switch {
		case errors.Is(err, io.EOF):
			d.eof = true
		case err != nil:
			d.err = err
			return false
		}
	}
}

// cut removes the next complete event from the buffer.
func (d *geminiDecoder) cut() ([]byte, bool) {
	at, size := -1, 0
	for _, delim := range geminiDelimiters {
		// Resume a little before where the last search ended, so a
		// delimiter split across reads is found.
		from := max(0, d.scanned-len(delim)+1)
		if i := bytes.Index(d.buf[from:], delim); i >= 0 && (at < 0 || from+i < at) {
			at, size = from+i, len(delim)
		}
	}
	if at < 0 {
		d.scanned = len(d.buf)
		return nil, false
	}
	event := d.buf[:at]
	d.buf = d.buf[at+size:]
	d.scanned = 0
	return event, true
}

// errorChunk applies the SDK's error check to one event: a JSON object with
// an "error" whose numeric code is 400–599 fails the stream with the SDK's
// ApiError text, "got status: <status>. <the object as JSON>".
func (d *geminiDecoder) errorChunk(event []byte) bool {
	var chunk map[string]json.RawMessage
	if json.Unmarshal(event, &chunk) != nil {
		return false
	}
	raw, ok := chunk["error"]
	if !ok {
		return false
	}
	var fields struct {
		Status json.RawMessage `json:"status"`
		Code   json.RawMessage `json:"code"`
	}
	_ = json.Unmarshal(raw, &fields)
	code, ok := jsNumberOf(fields.Code)
	if !ok || code < 400 || code >= 600 {
		return false
	}
	msg := "got status: " + jsTemplateString(fields.Status) + ". " + compactJSON(bytes.TrimFunc(event, isJSSpace))
	d.err = &geminiStreamError{d.upstream(msg)}
	return true
}

func (d *geminiDecoder) chunk() []byte { return d.data }

// jsNumberOf is a JSON value as JavaScript compares it to a number: a number,
// or a string holding one.
func jsNumberOf(raw json.RawMessage) (float64, bool) {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return 0, false
	}
	switch v := v.(type) {
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimFunc(v, isJSSpace), 64)
		return f, err == nil
	}
	return 0, false
}

// jsTemplateString is a JSON value as a JavaScript template literal shows
// it; a missing value is "undefined".
func jsTemplateString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "undefined"
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return "undefined"
	}
	switch v := v.(type) {
	case nil:
		return "null"
	case string:
		return v
	case float64:
		return jsNumberString(v)
	case bool:
		return strconv.FormatBool(v)
	case []any:
		parts := make([]string, len(v))
		for i, e := range v {
			if e != nil {
				parts[i] = jsTemplateString(jsonOf(e))
			}
		}
		return strings.Join(parts, ",")
	}
	return "[object Object]"
}

func jsonOf(v any) json.RawMessage {
	out, _ := json.Marshal(v)
	return out
}
