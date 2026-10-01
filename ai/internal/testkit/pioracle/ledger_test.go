package pioracle

import (
	"strings"
	"testing"
)

const testLedger = `{"decisions": [
  {"protocol": "openai-responses", "path": "result.usage.cost", "kind": "only_pi", "classification": "pending", "decision": "cost lands with ticket 15", "ref": "issues/15"},
  {"protocol": "openai-responses", "path": "request.headers.User-Agent", "classification": "extension", "decision": "Go SDK identifies itself", "ref": "spec §8"},
  {"protocol": "openai-responses", "path": "events[*].partial", "kind": "only_pi", "classification": "pending", "decision": "PartialView lands with ticket 04", "ref": "issues/04"},
  {"protocol": "openai-responses", "path": "result.stopReason", "classification": "fixed", "decision": "stop mapping fixed", "ref": "commit abc"},
  {"protocol": "anthropic-messages", "path": "result.timestamp", "classification": "extension", "decision": "other protocol", "ref": "x"},
  {"protocol": "openai-responses", "path": "result.errorMessage", "kind": "changed", "cases": ["CASE-401", "CASE-403"], "classification": "extension", "decision": "redacted", "ref": "spec I9"}
]}`

func classify(t *testing.T, diffs ...Diff) Verdict {
	t.Helper()
	return classifyCase(t, "CASE-OTHER", diffs...)
}

func classifyCase(t *testing.T, caseID string, diffs ...Diff) Verdict {
	t.Helper()
	l, err := ParseLedger([]byte(testLedger))
	if err != nil {
		t.Fatal(err)
	}
	return l.Classify("openai-responses", caseID, diffs)
}

func TestCaseScopedDecisionAppliesOnlyToItsCases(t *testing.T) {
	diff := Diff{Path: "result.errorMessage", Kind: Changed}
	for _, id := range []string{"CASE-401", "CASE-403"} {
		if v := classifyCase(t, id, diff); !v.Pass || v.Findings[0].Ref != "spec I9" {
			t.Errorf("%s: scoped decision not applied: %+v", id, v)
		}
	}
	if v := classifyCase(t, "CASE-429", diff); v.Pass || v.Findings[0].Ref != "" {
		t.Errorf("scoped decision applied outside its cases: %+v", v)
	}
}

func TestCaseScopedDecisionRejectsEmptyCaseIDs(t *testing.T) {
	_, err := ParseLedger([]byte(`{"decisions": [{"protocol": "p", "path": "result.x", "cases": [""], "classification": "extension", "decision": "d", "ref": "r"}]}`))
	if err == nil {
		t.Error("an empty case id was accepted")
	}
}

func TestUnledgeredDiffIsPendingAndFailsTheGate(t *testing.T) {
	v := classify(t, Diff{Path: "result.timestamp", Kind: OnlyPi})
	if v.Pass || v.Pending != 1 || v.Findings[0].Classification != Pending {
		t.Errorf("verdict %+v", v)
	}
}

func TestOtherProtocolsEntriesDoNotApply(t *testing.T) {
	v := classify(t, Diff{Path: "result.timestamp", Kind: OnlyPi})
	if v.Findings[0].Ref == "x" {
		t.Errorf("anthropic entry applied to responses: %+v", v.Findings[0])
	}
}

func TestExtensionsPassTheGate(t *testing.T) {
	v := classify(t, Diff{Path: "request.headers.User-Agent", Kind: Changed})
	if !v.Pass || v.Findings[0].Classification != Extension || v.Findings[0].Ref != "spec §8" {
		t.Errorf("verdict %+v", v)
	}
}

func TestWildcardMatchesAnyIndexAndDescendants(t *testing.T) {
	v := classify(t,
		Diff{Path: "events[0].partial", Kind: OnlyPi},
		Diff{Path: "events[12].partial.content[0]", Kind: OnlyPi})
	for _, f := range v.Findings {
		if f.Ref != "issues/04" {
			t.Errorf("%s not matched by events[*].partial: %+v", f.Path, f)
		}
	}
}

func TestPatternDoesNotMatchPrefixSiblings(t *testing.T) {
	v := classify(t, Diff{Path: "result.usage.costly", Kind: OnlyPi}, Diff{Path: "events[0].partialX", Kind: OnlyPi})
	for _, f := range v.Findings {
		if f.Ref != "" {
			t.Errorf("%s matched an entry for a different field: %+v", f.Path, f)
		}
	}
}

func TestKindRestrictsTheMatch(t *testing.T) {
	v := classify(t, Diff{Path: "result.usage.cost", Kind: Changed})
	if v.Findings[0].Ref != "" {
		t.Errorf("only_pi entry matched a changed diff: %+v", v.Findings[0])
	}
}

func TestFixedDifferenceReappearingIsARegression(t *testing.T) {
	v := classify(t, Diff{Path: "result.stopReason", Kind: Changed})
	f := v.Findings[0]
	if v.Pass || f.Classification != Pending || !strings.Contains(f.Decision, "regression") {
		t.Errorf("verdict %+v", v)
	}
}

func TestNoDiffsPass(t *testing.T) {
	if v := classify(t); !v.Pass || v.Pending != 0 {
		t.Errorf("verdict %+v", v)
	}
}

func TestMalformedLedgersAreRejected(t *testing.T) {
	cases := map[string]string{
		"unknown classification": `{"decisions":[{"protocol":"p","path":"result.x","classification":"ignored","decision":"d","ref":"r"}]}`,
		"missing decision":       `{"decisions":[{"protocol":"p","path":"result.x","classification":"pending","ref":"r"}]}`,
		"missing ref":            `{"decisions":[{"protocol":"p","path":"result.x","classification":"pending","decision":"d"}]}`,
		"missing protocol":       `{"decisions":[{"path":"result.x","classification":"pending","decision":"d","ref":"r"}]}`,
		"bad section":            `{"decisions":[{"protocol":"p","path":"usage.x","classification":"pending","decision":"d","ref":"r"}]}`,
		"unknown kind":           `{"decisions":[{"protocol":"p","path":"result.x","kind":"gone","classification":"pending","decision":"d","ref":"r"}]}`,
		"unknown field":          `{"decisions":[{"protocol":"p","path":"result.x","classification":"pending","decision":"d","ref":"r","ignore":true}]}`,
		"catch-all pattern":      `{"decisions":[{"protocol":"p","path":"result","classification":"extension","decision":"d","ref":"r"}]}`,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseLedger([]byte(data)); err == nil {
				t.Error("accepted")
			}
		})
	}
}
