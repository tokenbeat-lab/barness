//go:build live

package live

import (
	"encoding/json"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
)

// drained is a consumed Stream.
type drained struct {
	events []ai.Event
	res    ai.Result
	err    error
	// streamErr is Err() after Next returned false.
	streamErr error
}

// drain consumes st to its end, calling on for each event as it arrives.
// ai/e2e has its own consumers; the two test packages are separate programs
// and these few lines are not shared to avoid a testkit package for them.
func drain(st *ai.Stream, on func(ai.Event)) drained {
	defer st.Close()
	var d drained
	for st.Next() {
		e := st.Event()
		d.events = append(d.events, e)
		if on != nil {
			on(e)
		}
	}
	d.streamErr = st.Err()
	d.res, d.err = st.Result()
	return d
}

// events asserts the stream's structure: start first, one terminal event
// last that carries the Result's message, and every content block opened,
// fed and closed under its own index.
func (s *session) events(label string, d drained) {
	s.t.Helper()
	if !s.check(label+": events", len(d.events) > 1, "%d events", len(d.events)) {
		return
	}
	_, started := d.events[0].(ai.StartEvent)
	s.check(label+": start first", started, "first event %T", d.events[0])
	terminals := 0
	var final ai.AssistantMessage
	for _, e := range d.events {
		switch t := e.(type) {
		case ai.DoneEvent:
			terminals++
			final = t.Message
		case ai.ErrorEvent:
			terminals++
			final = t.Message
		}
	}
	last := d.events[len(d.events)-1]
	_, done := last.(ai.DoneEvent)
	_, failed := last.(ai.ErrorEvent)
	s.check(label+": one terminal event, last", terminals == 1 && (done || failed), "%d terminal events, last %T", terminals, last)
	a, _ := json.Marshal(final)
	b, _ := json.Marshal(d.res.Message)
	s.check(label+": terminal message is the Result's", string(a) == string(b), "terminal %s\nresult %s", a, b)
	s.check(label+": Err matches Result", errText(d.streamErr) == errText(d.err), "Err %v, Result %v", d.streamErr, d.err)
	if problem := blockProblem(d.events); problem != "" {
		s.check(label+": content blocks", false, "%s", problem)
	}
}

// blockProblem checks every block event against its block's lifetime. An
// aborted stream may leave its last block open.
func blockProblem(events []ai.Event) string {
	open := map[int]string{}
	closed := map[int]bool{}
	for i, e := range events {
		idx, kind, phase := blockOf(e)
		switch {
		case phase == "":
			continue
		case phase == "start":
			if _, dup := open[idx]; dup || closed[idx] {
				return fmt.Sprintf("event %d opens block %d twice", i, idx)
			}
			open[idx] = kind
		case open[idx] != kind:
			return fmt.Sprintf("event %d (%s %s) for block %d, which is %q", i, kind, phase, idx, open[idx])
		case phase == "end":
			delete(open, idx)
			closed[idx] = true
		}
	}
	if _, aborted := events[len(events)-1].(ai.ErrorEvent); len(open) > 0 && !aborted {
		return fmt.Sprintf("blocks left open: %v", open)
	}
	return ""
}

func blockOf(e ai.Event) (index int, kind, phase string) {
	switch t := e.(type) {
	case ai.TextStartEvent:
		return t.ContentIndex, "text", "start"
	case ai.TextDeltaEvent:
		return t.ContentIndex, "text", "delta"
	case ai.TextEndEvent:
		return t.ContentIndex, "text", "end"
	case ai.ThinkingStartEvent:
		return t.ContentIndex, "thinking", "start"
	case ai.ThinkingDeltaEvent:
		return t.ContentIndex, "thinking", "delta"
	case ai.ThinkingEndEvent:
		return t.ContentIndex, "thinking", "end"
	case ai.ToolCallStartEvent:
		return t.ContentIndex, "toolcall", "start"
	case ai.ToolCallDeltaEvent:
		return t.ContentIndex, "toolcall", "delta"
	case ai.ToolCallEndEvent:
		return t.ContentIndex, "toolcall", "end"
	}
	return 0, "", ""
}
