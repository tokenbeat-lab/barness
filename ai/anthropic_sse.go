package ai

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"strings"
)

// sseEvent is one server-sent event: its event name and its data lines
// joined by "\n".
type sseEvent struct {
	name string
	data string
}

// sseDecoder ports pi-ai's own SSE decoder for Anthropic (iterateSseMessages),
// which pi uses instead of the SDK's: comment lines are skipped, only the
// event and data fields are read, an event is dispatched at a blank line when
// it has a name or data, and an event the body ends without a blank line
// after is still dispatched. The Anthropic Go SDK's decoder drops that last
// event and appends a newline to the data, so it is not used.
//
// Lines end at LF, a CR before it dropped, as streamBody counts frames; pi
// also ends a line at a lone CR, which no Anthropic stream sends.
type sseDecoder struct {
	scn *bufio.Scanner
	cur sseEvent
	err error
	eof bool
}

// newSSEDecoder reads body. frameLimit is the policy's MaxFrameBytes, which
// streamBody already enforces; the scanner's own token limit only has to
// stay above it.
func newSSEDecoder(body io.Reader, frameLimit int64) *sseDecoder {
	scn := bufio.NewScanner(body)
	scn.Buffer(nil, int(min(frameLimit, math.MaxInt32-2))+2)
	return &sseDecoder{scn: scn}
}

// next advances to the next event; false at the end of the body or on a read
// error (see err).
func (d *sseDecoder) next() bool {
	if d.eof {
		return false
	}
	var name string
	var data []string
	for d.scn.Scan() {
		line := d.scn.Bytes()
		if len(line) == 0 {
			if name != "" || len(data) > 0 {
				d.cur = sseEvent{name: name, data: strings.Join(data, "\n")}
				return true
			}
			name, data = "", nil
			continue
		}
		if line[0] == ':' {
			continue
		}
		field, value, found := bytes.Cut(line, []byte(":"))
		if found && len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}
		switch string(field) {
		case "event":
			name = string(value)
		case "data":
			data = append(data, string(value))
		}
	}
	d.eof = true
	d.err = d.scn.Err()
	if d.err == nil && (name != "" || len(data) > 0) {
		d.cur = sseEvent{name: name, data: strings.Join(data, "\n")}
		return true
	}
	return false
}

func (d *sseDecoder) event() sseEvent { return d.cur }
