package minimaxlive

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
)

const MaxRequestReruns = 3

// A complete baseline may repeat a scene once. Research is never restarted.
// The independent capture verifier additionally proves the registered reason;
// budget receipts alone cannot certify that a repeated attempt was appropriate.
func RerunnableCase(id string) bool {
	for _, c := range Plan(Identity{}) {
		if c.ID == id && id != "preflight-error" {
			return true
		}
	}
	return false
}
func RequestRerunnable(result Capture, events []json.RawMessage, status int) bool {
	if result.ErrorPhase != ai.PhaseRequest || len(result.Result.Metadata.Attempts) != 1 || len(result.Result.Message.Content) != 0 {
		return false
	}
	a := result.Result.Metadata.Attempts[0]
	if a.Code != result.ErrorCode || a.HTTPStatus != status || a.HTTPStatus >= 200 && a.HTTPStatus < 400 {
		return false
	}
	switch result.ErrorCode {
	case ai.CodeRateLimited:
		if a.HTTPStatus != 429 {
			return false
		}
	case ai.CodeUpstreamError:
		if a.HTTPStatus < 500 && a.HTTPStatus != 408 && a.HTTPStatus != 409 {
			return false
		}
	case ai.CodeTransport, ai.CodeDeadlineExceeded:
		if a.HTTPStatus != 0 {
			return false
		}
	default:
		return false
	}
	for _, raw := range events {
		var event struct {
			Type ai.EventType `json:"type"`
		}
		if json.Unmarshal(raw, &event) != nil || event.Type != ai.EventStart && event.Type != ai.EventError {
			return false
		}
	}
	return true
}
func Recovered(calls []Call, index int) bool {
	c := calls[index]
	if c.Outcome != "FAIL" || c.Rerun != 0 || index+1 >= len(calls) {
		return false
	}
	next := calls[index+1]
	if next.ID != c.ID || next.Rerun != 1 || next.Outcome != "PASS" {
		return false
	}
	return c.Failure == "transient-request" && RerunnableCase(c.ID) || c.Failure == "tool-capability" && c.ID == "tool-call"
}
