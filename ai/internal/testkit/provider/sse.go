package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Framing selects how scripted SSE events are laid out on the wire. Every
// framing must decode to the same event sequence.
type Framing string

const (
	// FramingLF writes one "event:/data:" block per chunk with LF line endings.
	FramingLF Framing = "lf"
	// FramingCRLF uses CRLF line endings.
	FramingCRLF Framing = "crlf"
	// FramingBytewise flushes every byte as its own chunk.
	FramingBytewise Framing = "bytewise"
	// FramingMultiline splits each JSON payload across several "data:" lines.
	FramingMultiline Framing = "multiline"
	// FramingHeartbeat interleaves SSE comment heartbeats between events.
	FramingHeartbeat Framing = "heartbeat"
	// FramingUnknownEvent interleaves event types the client does not know.
	FramingUnknownEvent Framing = "unknown-event"
)

// Framings lists every framing, in a stable order.
var Framings = []Framing{FramingLF, FramingCRLF, FramingBytewise, FramingMultiline, FramingHeartbeat, FramingUnknownEvent}

// EncodeSSE lays out events (JSON objects carrying a "type" field) as SSE chunks.
func EncodeSSE(events []json.RawMessage, framing Framing) ([][]byte, error) {
	var chunks [][]byte
	for i, ev := range events {
		var head struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(ev, &head); err != nil || head.Type == "" {
			return nil, fmt.Errorf("sse: event %d has no type", i)
		}
		switch framing {
		case FramingHeartbeat:
			chunks = append(chunks, []byte(": keep-alive\n\n"))
		case FramingUnknownEvent:
			unknown := fmt.Sprintf(`{"type":"response.barness_unknown_%d","sequence_number":%d,"payload":{"x":1}}`, i, 10000+i)
			chunks = append(chunks, block("response.barness_unknown", []string{unknown}, "\n"))
		}
		lines := []string{compact(ev)}
		if framing == FramingMultiline {
			lines = indentedLines(ev)
		}
		eol := "\n"
		if framing == FramingCRLF {
			eol = "\r\n"
		}
		chunks = append(chunks, block(head.Type, lines, eol))
	}
	if framing == FramingBytewise {
		var split [][]byte
		for _, c := range chunks {
			for _, b := range c {
				split = append(split, []byte{b})
			}
		}
		chunks = split
	}
	return chunks, nil
}

func block(eventType string, dataLines []string, eol string) []byte {
	var b bytes.Buffer
	b.WriteString("event: " + eventType + eol)
	for _, l := range dataLines {
		b.WriteString("data: " + l + eol)
	}
	b.WriteString(eol)
	return b.Bytes()
}

func compact(ev json.RawMessage) string {
	var b bytes.Buffer
	_ = json.Compact(&b, ev)
	return b.String()
}

func indentedLines(ev json.RawMessage) []string {
	var b bytes.Buffer
	_ = json.Indent(&b, ev, "", "  ")
	return splitLines(b.String())
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}
