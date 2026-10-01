package ai

import (
	"encoding/json"
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
type PartialView struct {
	mu  sync.RWMutex
	msg AssistantMessage
}

func newPartialView(msg AssistantMessage) *PartialView { return &PartialView{msg: msg} }

// Snapshot returns a copy of the message as it is now. While the call runs its
// StopReason is StopReasonPending. The copy shares no storage with the view.
func (v *PartialView) Snapshot() AssistantMessage {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.msg.clone()
}

// MarshalJSON encodes the current snapshot, as serializing pi's live partial does.
func (v *PartialView) MarshalJSON() ([]byte, error) { return json.Marshal(v.Snapshot()) }

// read inspects the message under the view's read lock.
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
