package pioracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Classification is the only way a difference may be explained (spec Testing
// Decisions §5): 已修复 / 规范定义或批准的扩展 / 待处理.
type Classification string

const (
	// Fixed records a difference that was fixed. It never excuses a diff: if it
	// shows up again it is a regression and counts as pending.
	Fixed Classification = "fixed"
	// Extension is behavior the spec defines or a later decision approved.
	Extension Classification = "extension"
	// Pending is an unexplained or not-yet-resolved difference. Any pending
	// finding fails the protocol's differential gate.
	Pending Classification = "pending"
)

// Decision is one reviewed difference in the ledger. Path is a diff path in
// which "[*]" matches any array index; it also covers everything below it.
type Decision struct {
	Protocol       string         `json:"protocol"`
	Path           string         `json:"path"`
	Kind           DiffKind       `json:"kind,omitempty"`
	Classification Classification `json:"classification"`
	// Decision is the handling decision and its reason.
	Decision string `json:"decision"`
	// Ref points at the evidence or follow-up: a spec section, ADR or ticket.
	Ref string `json:"ref"`

	match *regexp.Regexp
}

// Ledger is the maintained set of decisions, versioned with the fixtures.
type Ledger struct {
	decisions []Decision
}

// Finding is a diff with its classification and handling decision.
type Finding struct {
	Diff
	Classification Classification `json:"classification"`
	Decision       string         `json:"decision"`
	Ref            string         `json:"ref,omitempty"`
}

// Verdict is one case's gate outcome.
type Verdict struct {
	Findings []Finding `json:"findings"`
	Pending  int       `json:"pending"`
	Pass     bool      `json:"pass"`
}

// leafSections are request fields that are a single value; a decision may
// name them directly. Every other decision must name something below a
// container, so no entry can excuse a whole request body, event list or result.
var leafSections = map[string]bool{"request.method": true, "request.path": true, "request.query": true, "request.auth": true}

var containerPrefix = regexp.MustCompile(`^(request\.headers\.[^.\[]+|request\.body[.\[]|events\[(\d+|\*)\][.\[]|events\[\d+\]$|result[.\[])`)

// ParseLedger reads and validates a ledger file.
func ParseLedger(data []byte) (Ledger, error) {
	var file struct {
		Decisions []Decision `json:"decisions"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&file); err != nil {
		return Ledger{}, fmt.Errorf("ledger: %w", err)
	}
	for i := range file.Decisions {
		d := &file.Decisions[i]
		if err := d.validate(); err != nil {
			return Ledger{}, fmt.Errorf("ledger decision %d (%s %s): %w", i, d.Protocol, d.Path, err)
		}
		pattern := regexp.QuoteMeta(d.Path)
		pattern = strings.ReplaceAll(pattern, `\[\*\]`, `\[\d+\]`)
		d.match = regexp.MustCompile(`^` + pattern + `($|[.\[])`)
	}
	return Ledger{decisions: file.Decisions}, nil
}

func (d Decision) validate() error {
	switch {
	case d.Protocol == "":
		return fmt.Errorf("protocol is required")
	case d.Decision == "":
		return fmt.Errorf("decision is required")
	case d.Ref == "":
		return fmt.Errorf("ref is required")
	}
	switch d.Classification {
	case Fixed, Extension, Pending:
	default:
		return fmt.Errorf("unknown classification %q", d.Classification)
	}
	switch d.Kind {
	case "", Changed, OnlyPi, OnlyBarness:
	default:
		return fmt.Errorf("unknown kind %q", d.Kind)
	}
	if !leafSections[d.Path] && !containerPrefix.MatchString(d.Path) {
		return fmt.Errorf("path must name a request field or something inside request headers/body, an event or the result")
	}
	return nil
}

// Classify explains each diff with the first matching decision for protocol.
// A diff without a decision is pending ("unreviewed"); a diff matching a
// Fixed decision is a regression and also pending.
func (l Ledger) Classify(protocol string, diffs []Diff) Verdict {
	v := Verdict{Findings: make([]Finding, 0, len(diffs))}
	for _, diff := range diffs {
		f := Finding{Diff: diff, Classification: Pending, Decision: "unreviewed: no ledger decision"}
		if d, ok := l.lookup(protocol, diff); ok {
			f.Classification, f.Decision, f.Ref = d.Classification, d.Decision, d.Ref
			if d.Classification == Fixed {
				f.Classification = Pending
				f.Decision = "regression of a fixed difference: " + d.Decision
			}
		}
		if f.Classification == Pending {
			v.Pending++
		}
		v.Findings = append(v.Findings, f)
	}
	v.Pass = v.Pending == 0
	return v
}

func (l Ledger) lookup(protocol string, diff Diff) (Decision, bool) {
	for _, d := range l.decisions {
		if d.Protocol == protocol && (d.Kind == "" || d.Kind == diff.Kind) && d.match.MatchString(diff.Path) {
			return d, true
		}
	}
	return Decision{}, false
}
