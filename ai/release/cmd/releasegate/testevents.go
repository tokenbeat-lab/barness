package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"time"
)

// Only timing metadata crosses this boundary; original output stays in the
// command log. The line ceiling bounds JSON decoding even for a broken child.
const maxEventLine = 1 << 20

type testEvent struct {
	Time    string   `json:"Time,omitempty"`
	Action  string   `json:"Action"`
	Package string   `json:"Package,omitempty"`
	Test    string   `json:"Test,omitempty"`
	Elapsed *float64 `json:"Elapsed,omitempty"`
}
type eventWriter struct {
	dst     io.Writer
	pending []byte
	err     error
}

func newEventWriter(dst io.Writer) *eventWriter { return &eventWriter{dst: dst} }
func (w *eventWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	size := len(p)
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			i = len(p)
		}
		if len(w.pending)+i > maxEventLine {
			w.err = errors.New("Go test JSON event exceeds line limit")
			return 0, w.err
		}
		w.pending = append(w.pending, p[:i]...)
		p = p[i:]
		if len(p) == 0 {
			break
		}
		p = p[1:]
		if err := w.flush(); err != nil {
			return 0, err
		}
	}
	return size, nil
}
func (w *eventWriter) flush() error {
	if len(bytes.TrimSpace(w.pending)) == 0 {
		w.pending = nil
		return nil
	}
	var e testEvent
	if json.Unmarshal(w.pending, &e) != nil || e.Action == "" {
		w.err = errors.New("invalid Go test JSON event")
		return w.err
	}
	if e.Time != "" {
		if _, err := time.Parse(time.RFC3339Nano, e.Time); err != nil {
			w.err = errors.New("invalid Go test event time")
			return w.err
		}
	}
	if e.Elapsed != nil && (*e.Elapsed < 0 || math.IsInf(*e.Elapsed, 0) || math.IsNaN(*e.Elapsed)) {
		w.err = errors.New("invalid Go test event elapsed time")
		return w.err
	}
	w.pending = nil
	w.err = json.NewEncoder(w.dst).Encode(e)
	return w.err
}
func (w *eventWriter) Finish() error {
	if w.err != nil {
		return w.err
	}
	return w.flush()
}
