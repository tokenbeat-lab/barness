package ai

import (
	"encoding/json"
	"slices"
	"sync"
)

// PartialView is the live response-so-far of one call, pi's `partial`. Every
// non-terminal event of a stream refers to the same view, and the view keeps
// growing after an event was published: it shows the call as it is when read,
// which may be ahead of the event being handled, never behind it. Use the
// event's own fields for what that event reported.
//
// The view is safe to read from any goroutine while the call runs. Once the
// call has reached its terminal the view equals the final message and never
// changes again.
//
// A streaming tool call's Arguments are parsed from its RawArguments when the
// view is read, not on every delta (issue 32): pi re-parses the whole text on
// each delta, which costs the square of the argument size. The value read is
// the one pi holds at that point. The first read after a delta parses under
// the view's write lock, briefly holding up the producer and other readers;
// reading after every delta costs what pi pays, reading less often costs less.
type PartialView struct {
	mu  sync.RWMutex
	msg AssistantMessage
	// unparsed are the tool call blocks whose Arguments still await the
	// display parse of their current RawArguments.
	unparsed []int
}

func newPartialView(msg AssistantMessage) *PartialView { return &PartialView{msg: msg} }

// Snapshot returns a copy of the message as it is now. While the call runs its
// StopReason is StopReasonPending. The copy shares no storage with the view.
func (v *PartialView) Snapshot() AssistantMessage {
	v.mu.RLock()
	if len(v.unparsed) == 0 {
		defer v.mu.RUnlock()
		return v.msg.clone()
	}
	v.mu.RUnlock()
	v.mu.Lock()
	defer v.mu.Unlock()
	v.parseArguments()
	return v.msg.clone()
}

// MarshalJSON encodes the current snapshot, as serializing pi's live partial does.
func (v *PartialView) MarshalJSON() ([]byte, error) { return json.Marshal(v.Snapshot()) }

// read inspects the message under the view's read lock. A streaming tool
// call's Arguments may not be parsed yet; its RawArguments are current.
func (v *PartialView) read(inspect func(*AssistantMessage)) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	inspect(&v.msg)
}

// update applies one change of the producing call under the view's lock.
func (v *PartialView) update(change func(*AssistantMessage)) {
	v.mu.Lock()
	defer v.mu.Unlock()
	change(&v.msg)
}

// streamToolArguments replaces tool call block i's raw argument text; its
// Arguments are parsed from it when the view is next read.
func (v *PartialView) streamToolArguments(i int, raw string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	c := v.msg.Content[i].(ToolCall)
	c.RawArguments = raw
	v.msg.Content[i] = c
	if !slices.Contains(v.unparsed, i) {
		v.unparsed = append(v.unparsed, i)
	}
}

// endToolArguments settles tool call block i with its final raw text and
// arguments, and returns the call.
func (v *PartialView) endToolArguments(i int, raw string, args UncheckedArguments) ToolCall {
	v.mu.Lock()
	defer v.mu.Unlock()
	c := v.msg.Content[i].(ToolCall)
	c.RawArguments, c.Arguments = raw, args
	v.msg.Content[i] = c
	v.unparsed = slices.DeleteFunc(v.unparsed, func(k int) bool { return k == i })
	return c
}

// parseArguments gives every unparsed tool call the display parse of its raw
// text, as pi computed it on the latest delta. The caller holds the lock.
func (v *PartialView) parseArguments() {
	for _, i := range v.unparsed {
		c := v.msg.Content[i].(ToolCall)
		c.Arguments = displayArguments(c.RawArguments)
		v.msg.Content[i] = c
	}
	v.unparsed = v.unparsed[:0]
}
