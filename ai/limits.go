package ai

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// byteLimits are the ResourcePolicy byte limits one call enforces. They come
// from the Client's validated policy only; no adapter has capacities of its
// own (spec I9, D1).
type byteLimits struct {
	request, image, frame, toolJSON, errorBody, output int64
}

func (p *ResourcePolicy) byteLimits() byteLimits {
	return byteLimits{
		request:   p.MaxRequestBytes,
		image:     p.MaxImageBytes,
		frame:     p.MaxFrameBytes,
		toolJSON:  p.MaxToolJSONBytes,
		errorBody: p.MaxErrorBodyBytes,
		output:    p.MaxOutputBytes,
	}
}

// limitExceeded is the error a bounded reader returns once the bytes it read
// pass a policy limit; it stops reading there.
type limitExceeded struct {
	field string // the ResourcePolicy field
	limit int64
}

func (e *limitExceeded) Error() string { return limitMessage(e.field, e.limit) }

// failure classifies the limit in the phase it was reached.
func (e *limitExceeded) failure(phase Phase) *Error {
	return limitFailure(phase, e.field, e.limit)
}

// limitMessage is barness's own text: pi has no limits. It names the policy
// field, never content.
func limitMessage(field string, limit int64) string {
	return "exceeded the resource policy's " + field + " (" + strconv.FormatInt(limit, 10) + ")"
}

func limitFailure(phase Phase, field string, limit int64) *Error {
	return newError(CodeResourceLimit, phase, limitMessage(field, limit))
}

// checkRequestBody refuses a final encoded request body over the limit
// before it is sent.
func (l byteLimits) checkRequestBody(body []byte) *Error {
	if int64(len(body)) > l.request {
		return limitFailure(PhaseRequest, "MaxRequestBytes", l.request)
	}
	return nil
}

// checkImages refuses a request holding an image whose bytes exceed the
// limit, wherever it is in the history, before anything resolves. The size is
// the image's own bytes, computed from its base64 text without decoding.
func (l byteLimits) checkImages(req Request) *Error {
	for i, m := range req.Messages {
		var blocks []any
		switch m := m.(type) {
		case UserMessage:
			blocks = anyBlocks(m.Content)
		case ToolResultMessage:
			blocks = anyBlocks(m.Content)
		}
		for _, b := range blocks {
			if img, ok := b.(Image); ok && base64Size(img.Data) > l.image {
				f := limitFailure(PhaseScope, "MaxImageBytes", l.image)
				f.Message = "message " + strconv.Itoa(i) + ": image " + f.Message
				return f
			}
		}
	}
	return nil
}

// base64Size is the number of bytes standard base64 text encodes, padded or
// not. Text that is not base64 is passed on as is (as pi does), and sized by
// the same rule.
func base64Size(data string) int64 {
	n := int64(len(data))
	padding := int64(len(data) - len(strings.TrimRight(data, "=")))
	return n*3/4 - min(padding, 2)
}

// limitBody bounds res's body while it is read: a successful response's body
// by SSE frame/output limits or the unary output limit, any other by the
// error body limit. The SDK
// and adapter read it through the wrapper, so nothing is read past a limit.
type bodyKind int

const (
	bodySSE bodyKind = iota
	bodyJSON
)

func (l byteLimits) limitBody(res *http.Response, kind bodyKind) {
	if res == nil || res.Body == nil {
		return
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if kind == bodyJSON {
			res.Body = &totalBody{ReadCloser: res.Body, limit: l.output, field: "MaxOutputBytes"}
		} else {
			res.Body = &streamBody{ReadCloser: res.Body, frameLimit: l.frame, outputLimit: l.output}
		}
		return
	}
	res.Body = &totalBody{ReadCloser: res.Body, limit: l.errorBody, field: "MaxErrorBodyBytes"}
}

// streamBody bounds an SSE body: each frame (an event's lines through the
// blank line that ends it) by frameLimit,
// and the whole body, the call's streamed output, by outputLimit. A frame
// that never ends fails as soon as it is one byte too long.
type streamBody struct {
	io.ReadCloser
	frameLimit, outputLimit int64

	frame, output int64
	// Lines end at LF, a CR before it dropped, exactly as the SDK's decoder
	// (bufio.ScanLines) splits them; a lone CR ends nothing there either.
	line int64 // bytes of the current line, its terminator excluded
	cr   bool  // the line's last byte was '\r'
	err  error
}

func (b *streamBody) Read(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	n, err := b.ReadCloser.Read(p)
	for i, c := range p[:n] {
		b.frame++
		b.output++
		switch {
		case b.output > b.outputLimit:
			b.err = &limitExceeded{field: "MaxOutputBytes", limit: b.outputLimit}
		case b.frame > b.frameLimit:
			b.err = &limitExceeded{field: "MaxFrameBytes", limit: b.frameLimit}
		}
		if b.err != nil {
			return i, b.err
		}
		if c == '\n' {
			// A line that is empty, or only the '\r' of a CRLF, ends the frame.
			if b.line == 0 || (b.line == 1 && b.cr) {
				b.frame = 0
			}
			b.line, b.cr = 0, false
			continue
		}
		b.line++
		b.cr = c == '\r'
	}
	return n, err
}

// totalBody bounds a unary output or non-2xx diagnostic body.
type totalBody struct {
	io.ReadCloser
	limit, read int64
	field       string
	err         error
}

func (b *totalBody) Read(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if remaining := b.limit - b.read; int64(len(p)) > remaining {
		p = p[:min(int64(len(p)), remaining+1)]
	}
	n, err := b.ReadCloser.Read(p)
	b.read += int64(n)
	if over := b.read - b.limit; over > 0 {
		b.err = &limitExceeded{field: b.field, limit: b.limit}
		return n - int(over), b.err
	}
	return n, err
}
