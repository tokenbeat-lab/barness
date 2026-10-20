package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// Failure inventory precedes the event writer: fragmented lines, invalid JSON,
// missing action, negative elapsed, oversized lines and failed output writes
// must fail the command, rather than silently produce incomplete timing.
func TestEventMetadataStream(t *testing.T) {
	var out bytes.Buffer
	w := newEventWriter(&out)
	for _, part := range []string{`{"Time":"2026-10-20T00:00:00Z","Action":"output","Package":"p","Test":"TestA","Output":"sensitive fixture `, `body"}`, "\n", `{"Action":"pass","Package":"p","Test":"TestA","Elapsed":1.25}`} {
		if _, err := w.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Finish(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "fixture") || strings.Contains(out.String(), "Output") || !strings.Contains(out.String(), `"Elapsed":1.25`) {
		t.Fatalf("wrong metadata: %s", out.String())
	}
}
func TestEventStreamRejectsIncompleteTiming(t *testing.T) {
	for _, raw := range []string{"not-json\n", "{}\n", "{\"Action\":\"pass\",\"Elapsed\":-1}\n", strings.Repeat("x", (1<<20)+1)} {
		var out bytes.Buffer
		w := newEventWriter(&out)
		_, err := w.Write([]byte(raw))
		if err == nil {
			err = w.Finish()
		}
		if err == nil {
			t.Fatal("invalid timing stream accepted")
		}
	}
	w := newEventWriter(failedEventSink{})
	if _, err := w.Write([]byte("{\"Action\":\"pass\"}\n")); err == nil {
		t.Fatal("failed event output ignored")
	}
}

type failedEventSink struct{}

func (failedEventSink) Write(p []byte) (int, error) { return 0, errors.New("fixture write failure") }

// Stderr is a raw log stream, not test JSON. Its write failure must cancel
// the same child process rather than leave it blocked on an unread pipe.
func TestRawLogStreamCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var log bytes.Buffer
	w := &eventOutput{log: &log, cancel: cancel}
	if _, err := w.Write([]byte("raw compiler diagnostic")); err != nil {
		t.Fatal(err)
	}
	w.log = failedEventSink{}
	if _, err := w.Write([]byte("failed write")); err == nil || ctx.Err() == nil {
		t.Fatal("log failure did not cancel the child")
	}
}
