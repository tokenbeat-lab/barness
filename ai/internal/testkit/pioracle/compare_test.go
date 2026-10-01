package pioracle

import (
	"encoding/json"
	"reflect"
	"testing"
)

// The comparator is the oracle's judgement; these are the ways it could
// silently get a verdict wrong (spec Testing Decisions §5).

func obs(result string, events ...string) Observation {
	o := Observation{
		Request: Request{Method: "POST", Path: "/v1/responses", Auth: "key:a", Headers: map[string]string{}, Body: json.RawMessage(`{}`)},
		Result:  json.RawMessage(result),
	}
	for _, e := range events {
		o.Events = append(o.Events, json.RawMessage(e))
	}
	return o
}

func diffPaths(t *testing.T, pi, barness Observation) []string {
	t.Helper()
	diffs, err := Compare(pi, barness)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{}
	for _, d := range diffs {
		paths = append(paths, string(d.Kind)+" "+d.Path)
	}
	return paths
}

func expectDiffs(t *testing.T, pi, barness Observation, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if got := diffPaths(t, pi, barness); !reflect.DeepEqual(got, want) {
		t.Errorf("diffs\n got  %q\n want %q", got, want)
	}
}

func TestCompareObjectKeyOrderIsNotADifference(t *testing.T) {
	expectDiffs(t, obs(`{"a":1,"b":{"x":1,"y":2}}`), obs(`{"b":{"y":2,"x":1},"a":1}`))
}

func TestCompareArraysAreNotSorted(t *testing.T) {
	expectDiffs(t, obs(`{"content":["a","b"]}`), obs(`{"content":["b","a"]}`),
		"changed result.content[0]", "changed result.content[1]")
}

func TestCompareEventsAreNotReordered(t *testing.T) {
	expectDiffs(t, obs(`{}`, `{"type":"start"}`, `{"type":"done"}`), obs(`{}`, `{"type":"done"}`, `{"type":"start"}`),
		"changed events[0].type", "changed events[1].type")
}

func TestComparePresenceIsKept(t *testing.T) {
	cases := []struct {
		name        string
		pi, barness string
		want        []string
	}{
		{"null vs absent", `{"x":null}`, `{}`, []string{"only_pi result.x"}},
		{"zero vs absent", `{}`, `{"x":0}`, []string{"only_barness result.x"}},
		{"empty array vs absent", `{"x":[]}`, `{}`, []string{"only_pi result.x"}},
		{"empty array vs null", `{"x":[]}`, `{"x":null}`, []string{"changed result.x"}},
		{"zero vs null", `{"x":0}`, `{"x":null}`, []string{"changed result.x"}},
		{"empty string vs absent", `{"x":""}`, `{}`, []string{"only_pi result.x"}},
		{"unknown field kept", `{"x":1}`, `{"x":1,"barnessOnly":{"y":1}}`, []string{"only_barness result.barnessOnly"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { expectDiffs(t, obs(c.pi), obs(c.barness), c.want...) })
	}
}

func TestCompareNumbersByValue(t *testing.T) {
	expectDiffs(t, obs(`{"a":0.0000064,"b":1,"c":25}`), obs(`{"a":6.4e-06,"b":1.0,"c":2.5e1}`))
	expectDiffs(t, obs(`{"a":1}`), obs(`{"a":"1"}`), "changed result.a")
	expectDiffs(t, obs(`{"a":9007199254740993}`), obs(`{"a":9007199254740992}`), "changed result.a")
}

func TestCompareMapsTimeFieldsOneToOne(t *testing.T) {
	// Different clocks, same reference structure: equal.
	expectDiffs(t,
		obs(`{"timestamp":100}`, `{"type":"done","message":{"timestamp":100}}`),
		obs(`{"timestamp":555}`, `{"type":"done","message":{"timestamp":555}}`))
	// pi reuses one time where barness has two: the broken reference shows.
	expectDiffs(t,
		obs(`{"timestamp":100}`, `{"type":"done","message":{"timestamp":100}}`),
		obs(`{"timestamp":555}`, `{"type":"done","message":{"timestamp":556}}`),
		"changed result.timestamp")
	// Mapping never hides absence.
	expectDiffs(t, obs(`{"timestamp":100}`), obs(`{}`), "only_pi result.timestamp")
	// Only the explicit message time fields are mapped: a wire or content field
	// that happens to be named timestamp is compared as is.
	pi, barness := obs(`{"content":[{"timestamp":1}]}`, `{"type":"start","timestamp":7}`), obs(`{"content":[{"timestamp":2}]}`, `{"type":"start","timestamp":8}`)
	pi.Request.Body, barness.Request.Body = json.RawMessage(`{"timestamp":1}`), json.RawMessage(`{"timestamp":2}`)
	expectDiffs(t, pi, barness, "changed request.body.timestamp", "changed events[0].timestamp", "changed result.content[0].timestamp")
	// Partial and error messages are message time fields too.
	expectDiffs(t,
		obs(`{"timestamp":1}`, `{"type":"text_start","partial":{"timestamp":1}}`, `{"type":"error","error":{"timestamp":1}}`),
		obs(`{"timestamp":9}`, `{"type":"text_start","partial":{"timestamp":9}}`, `{"type":"error","error":{"timestamp":9}}`))
}

func TestCompareDoesNotMapIDs(t *testing.T) {
	expectDiffs(t, obs(`{"responseId":"resp_1","id":"x"}`), obs(`{"responseId":"resp_2","id":"x"}`), "changed result.responseId")
}

func TestCompareLocatesMissingAndExtraEvents(t *testing.T) {
	expectDiffs(t, obs(`{}`, `{"type":"start"}`, `{"type":"done"}`), obs(`{}`, `{"type":"start"}`), "only_pi events[1]")
	expectDiffs(t, obs(`{}`, `{"type":"start"}`), obs(`{}`, `{"type":"start"}`, `{"type":"done"}`), "only_barness events[1]")
}

func TestCompareReportsRequestThenEventsThenResult(t *testing.T) {
	pi, barness := obs(`{"r":1}`, `{"e":1}`), obs(`{"r":2}`, `{"e":2}`)
	pi.Request.Headers = map[string]string{"User-Agent": "pi", "Accept": "a"}
	barness.Request.Headers = map[string]string{"User-Agent": "go", "Accept": "b"}
	pi.Request.Body = json.RawMessage(`{"model":"m","store":false}`)
	barness.Request.Body = json.RawMessage(`{"model":"m"}`)
	barness.Request.Auth = "key:b"
	barness.Request.Query = "x=1"
	expectDiffs(t, pi, barness,
		"changed request.query",
		"changed request.auth",
		"changed request.headers.Accept",
		"changed request.headers.User-Agent",
		"only_pi request.body.store",
		"changed events[0].e",
		"changed result.r")
}

func TestCompareRejectsInvalidJSON(t *testing.T) {
	if _, err := Compare(obs(`{`), obs(`{}`)); err == nil {
		t.Error("invalid pi JSON compared without error")
	}
	if _, err := Compare(obs(`{}`), obs(`{}`, `nope`)); err == nil {
		t.Error("invalid barness event compared without error")
	}
}

func TestCompareDiffCarriesBothValues(t *testing.T) {
	diffs, err := Compare(obs(`{"stopReason":"stop"}`), obs(`{"stopReason":"length"}`))
	if err != nil || len(diffs) != 1 {
		t.Fatalf("diffs=%v err=%v", diffs, err)
	}
	if string(diffs[0].Pi) != `"stop"` || string(diffs[0].Barness) != `"length"` {
		t.Errorf("values pi=%s barness=%s", diffs[0].Pi, diffs[0].Barness)
	}
}

func TestCompareObservationsWithoutRequest(t *testing.T) {
	// A call refused before any request (connection refused) has no capture
	// on either side; that is agreement, not a decode failure.
	none := obs(`{}`)
	none.Request = Request{}
	expectDiffs(t, none, none)
	// One side sending a request the other did not is still a difference.
	paths := diffPaths(t, obs(`{}`), none)
	if len(paths) == 0 {
		t.Error("a request on one side only was not reported")
	}
}
