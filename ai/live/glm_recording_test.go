//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tokenbeat-lab/barness/ai"
)

// 2MiB raw SSE leaves room for worst-case JSON string escaping plus the 256KiB
// request inside the existing 16MiB artifact contract. Overflow stays truncated
// and cannot qualify. The outer transport sets the limit before Body escapes
// RoundTrip, hence before its first Read; no reader can race this initialization.
const glmResponseCaptureLimit = 2 << 20

type glmCaptureTransport struct{ recorder *recorder }

func (t glmCaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := t.recorder.RoundTrip(req)
	if err != nil {
		return response, err
	}
	body, ok := response.Body.(*capturingBody)
	if !ok {
		response.Body.Close()
		return nil, fmt.Errorf("GLM recorder body missing")
	}
	t.recorder.mu.Lock()
	body.limit = glmResponseCaptureLimit
	t.recorder.mu.Unlock()
	return response, nil
}

type glmRecordedEvent struct {
	Type  ai.EventType       `json:"type"`
	Call  ai.CallAttribution `json:"call"`
	Event ai.Event           `json:"event"`
}
type glmEventLog struct {
	Schema        int                 `json:"schema"`
	PartialPolicy string              `json:"partialPolicy"`
	Events        []json.RawMessage   `json:"events"`
	FinalPartial  ai.AssistantMessage `json:"finalPartial"`
	Overflow      bool                `json:"overflow"`
	bytes         int
	partial       *ai.PartialView
}

func newGLMEventLog() *glmEventLog {
	return &glmEventLog{Schema: 1, PartialPolicy: "event-owned fields in order; shared live Partial recorded once after drain", Events: []json.RawMessage{}}
}

func (l *glmEventLog) append(envelope ai.EventEnvelope) error {
	e := envelope.Event
	// A PartialView is one live shared object, not an immutable per-event
	// snapshot (ai/event.go). Keep all immutable fields and every terminal;
	// sample the shared view once after draining instead of O(deltas*output).
	switch v := e.(type) {
	case ai.StartEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.TextStartEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.TextDeltaEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.TextEndEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ThinkingStartEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ThinkingDeltaEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ThinkingEndEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ToolCallStartEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ToolCallDeltaEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	case ai.ToolCallEndEvent:
		l.partial = v.Partial
		v.Partial = nil
		e = v
	}
	raw, err := json.Marshal(glmRecordedEvent{Type: e.Type(), Call: envelope.Call, Event: e})
	if err != nil {
		l.Overflow = true
		return err
	}
	terminal := e.Type() == ai.EventDone || e.Type() == ai.EventError
	// Reserve room for terminal + final snapshot; never turn lost events into
	// PASS. Cancellation after overflow still drains and records its terminal.
	if len(l.Events) >= 8192 || l.bytes+len(raw) > 4<<20 {
		l.Overflow = true
		if !terminal {
			return fmt.Errorf("GLM event evidence exceeds finite bound")
		}
	}
	l.Events = append(l.Events, raw)
	l.bytes += len(raw)
	return nil
}
func (l *glmEventLog) finish() {
	if l.partial != nil {
		l.FinalPartial = l.partial.Snapshot()
	}
	l.partial = nil
}
