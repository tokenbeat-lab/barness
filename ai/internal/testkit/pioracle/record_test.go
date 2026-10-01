package pioracle

import (
	"encoding/json"
	"testing"
)

// The ways a record could misreport where the two sides first diverge.

func TestRecordLocatesFirstDiffPendingAndEvent(t *testing.T) {
	v := Verdict{Findings: []Finding{
		{Diff: Diff{Path: "request.headers.Accept", Kind: Changed}, Classification: Extension},
		{Diff: Diff{Path: "events[3].partial", Kind: OnlyPi}, Classification: Extension},
		{Diff: Diff{Path: "events[5].delta", Kind: Changed}, Classification: Pending},
		{Diff: Diff{Path: "result.usage", Kind: Changed}, Classification: Pending},
	}, Pending: 2}
	r := (&Oracle{}).Record("case", "openai-responses", v, Sides{})
	if r.FirstDiff == nil || r.FirstDiff.Path != "request.headers.Accept" {
		t.Errorf("first diff %+v", r.FirstDiff)
	}
	if r.FirstPending == nil || r.FirstPending.Path != "events[5].delta" {
		t.Errorf("first pending %+v", r.FirstPending)
	}
	if r.FirstEventIndex == nil || *r.FirstEventIndex != 3 {
		t.Errorf("first event index %v, want 3 (first diff inside events)", r.FirstEventIndex)
	}
	if r.Gate != GateFail {
		t.Errorf("gate %q with pending findings", r.Gate)
	}
}

func TestRecordWithoutDifferencesPasses(t *testing.T) {
	r := (&Oracle{}).Record("case", "openai-responses", Verdict{Findings: []Finding{}, Pass: true}, Sides{})
	if r.Gate != GatePass || r.FirstDiff != nil || r.FirstPending != nil || r.FirstEventIndex != nil {
		t.Errorf("record %+v", r)
	}
}

func TestRecordHashesWholeRequestsAndFrames(t *testing.T) {
	a := Sides{PiRequest: json.RawMessage(`{"method":"POST","body":{}}`), BarnessRequest: json.RawMessage(`{"method":"GET","body":{}}`), Frames: []byte("x")}
	r := (&Oracle{}).Record("case", "p", Verdict{Pass: true}, a)
	if r.RequestSHA256["pi"] == r.RequestSHA256["barness"] {
		t.Error("requests differing only outside the body hash equal")
	}
	if r.FrameSHA256 != SHA256([]byte("x")) {
		t.Errorf("frame hash %s", r.FrameSHA256)
	}
}
