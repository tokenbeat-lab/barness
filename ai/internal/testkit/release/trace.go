package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
)

// LiveScenario is the reserved scenario whose evidence is the support
// matrix rather than offline cases.
const LiveScenario = "LIVE"

const ArtifactScenario = "ARTIFACTS"

// TraceMap links research items to acceptance scenarios and scenarios to
// evidence case IDs (spec Testing Decisions §6: 需求 → 断言 → 实际证据). It is
// maintained with the research traceability table; the gate computes each
// item's status from evidence alone, so an item that is only mapped is never
// PASS.
type TraceMap struct {
	Schema int `json:"schema"`
	// Scenarios maps a scenario ID (E01, P03, D1, ...) to a regular
	// expression over evidence case IDs.
	Scenarios map[string]string `json:"scenarios"`
	// P0 are the scenarios that must each have at least one passing
	// offline case.
	P0 []string `json:"p0"`
	// Differential names the protocols that must run against frozen pi and
	// the routes outside pi registered as extensions instead.
	Differential DifferentialPlan `json:"differential"`
	// LiveCombos are the support matrix combinations a release requires.
	LiveCombos []string    `json:"liveCombos"`
	Items      []TraceItem `json:"items"`

	patterns map[string]*regexp.Regexp
}

// DifferentialPlan is what the differential gate requires.
type DifferentialPlan struct {
	Required []string   `json:"required"`
	Routes   []RouteRef `json:"routes"`
}

// RouteRef is a scenario proven by its own fixtures because frozen pi has
// no route for it; the ledger must register provider × api.
type RouteRef struct {
	Scenario string `json:"scenario"`
	Provider string `json:"provider"`
	API      string `json:"api"`
}

// TraceItem is one research item (T01, C03, H2, ...).
type TraceItem struct {
	ID          string   `json:"id"`
	Source      string   `json:"source"`
	Requirement string   `json:"requirement"`
	Scenarios   []string `json:"scenarios"`
	// Cases, when set, narrows the scenarios' cases to those matching it.
	Cases string `json:"cases,omitempty"`

	cases *regexp.Regexp
}

// ParseTraceMap reads and validates a traceability map.
func ParseTraceMap(data []byte) (TraceMap, error) {
	var m TraceMap
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return TraceMap{}, fmt.Errorf("trace map: %w", err)
	}
	if m.Schema != 1 {
		return TraceMap{}, fmt.Errorf("trace map: schema %d, want 1", m.Schema)
	}
	m.patterns = map[string]*regexp.Regexp{}
	for id, p := range m.Scenarios {
		if id == LiveScenario || id == ArtifactScenario {
			return TraceMap{}, fmt.Errorf("trace map: scenario %s is reserved for the support matrix", LiveScenario)
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return TraceMap{}, fmt.Errorf("trace map: scenario %s: %w", id, err)
		}
		m.patterns[id] = re
	}
	known := func(id string) bool { _, ok := m.patterns[id]; return ok }
	for _, id := range m.P0 {
		if !known(id) {
			return TraceMap{}, fmt.Errorf("trace map: p0 scenario %s is not defined", id)
		}
	}
	for _, id := range m.Differential.Required {
		if !known(id) {
			return TraceMap{}, fmt.Errorf("trace map: differential scenario %s is not defined", id)
		}
	}
	for _, r := range m.Differential.Routes {
		if !known(r.Scenario) || r.Provider == "" || r.API == "" {
			return TraceMap{}, fmt.Errorf("trace map: differential route %+v needs a defined scenario, a provider and an api", r)
		}
	}
	seen := map[string]bool{}
	for i := range m.Items {
		it := &m.Items[i]
		switch {
		case it.ID == "" || it.Source == "" || it.Requirement == "":
			return TraceMap{}, fmt.Errorf("trace map: item %d needs an id, a source and a requirement", i)
		case seen[it.ID]:
			return TraceMap{}, fmt.Errorf("trace map: item %s appears twice", it.ID)
		case len(it.Scenarios) == 0:
			return TraceMap{}, fmt.Errorf("trace map: item %s names no scenario", it.ID)
		}
		seen[it.ID] = true
		for _, s := range it.Scenarios {
			if s != LiveScenario && s != ArtifactScenario && !known(s) {
				return TraceMap{}, fmt.Errorf("trace map: item %s names undefined scenario %s", it.ID, s)
			}
		}
		if it.Cases != "" {
			re, err := regexp.Compile(it.Cases)
			if err != nil {
				return TraceMap{}, fmt.Errorf("trace map: item %s: %w", it.ID, err)
			}
			it.cases = re
		}
	}
	return m, nil
}

// matches reports whether case id is evidence for scenario.
func (m TraceMap) matches(scenario, id string) bool {
	re, ok := m.patterns[scenario]
	return ok && re.MatchString(id)
}

// TraceResult is one item's status, computed from evidence only.
type TraceResult struct {
	ID          string   `json:"id"`
	Source      string   `json:"source"`
	Requirement string   `json:"requirement"`
	Scenarios   []string `json:"scenarios"`
	Status      Status   `json:"status"`
	Offline     Counts   `json:"offline"`
	// Differential counts frozen pi differential cases.
	Differential Counts `json:"differential"`
	// Live counts support matrix combinations.
	Live      Counts `json:"live"`
	Artifacts Counts `json:"artifacts"`
	// Evidence lists a few of the matching case IDs (and live
	// combinations), where the assertions can be read.
	Evidence []string `json:"evidence"`
}

// maxEvidence bounds the examples kept per item; the counts cover all.
const maxEvidence = 6

func traceItems(m TraceMap, cases []Case, live []LiveRow, artifacts []EvidenceCheck) []TraceResult {
	out := make([]TraceResult, 0, len(m.Items))
	for _, it := range m.Items {
		r := TraceResult{ID: it.ID, Source: it.Source, Requirement: it.Requirement, Scenarios: it.Scenarios}
		for _, c := range cases {
			if kindOf(c.ID) == kindLive || !slices.ContainsFunc(it.Scenarios, func(s string) bool { return m.matches(s, c.ID) }) {
				continue
			}
			if it.cases != nil && !it.cases.MatchString(c.ID) {
				continue
			}
			if kindOf(c.ID) == kindDifferential {
				r.Differential.add(c.Status)
			} else {
				r.Offline.add(c.Status)
			}
			if len(r.Evidence) < maxEvidence {
				r.Evidence = append(r.Evidence, c.ID)
			}
		}
		if slices.Contains(it.Scenarios, LiveScenario) {
			for _, row := range live {
				r.Live.add(row.Status)
				if len(r.Evidence) < maxEvidence*2 {
					r.Evidence = append(r.Evidence, "support-matrix:"+row.Combo+"="+string(row.Status))
				}
			}
			if len(live) == 0 {
				r.Live.NotRun++
			}
		}
		if slices.Contains(it.Scenarios, ArtifactScenario) {
			for _, e := range artifacts {
				if it.cases != nil && !it.cases.MatchString(e.Name) {
					continue
				}
				r.Artifacts.add(e.Status)
				if len(r.Evidence) < maxEvidence*2 {
					r.Evidence = append(r.Evidence, "artifact:"+e.Name+"="+string(e.Status))
				}
			}
		}
		all := r.Offline.plus(r.Differential).plus(r.Live).plus(r.Artifacts)
		switch {
		case all.Fail > 0:
			r.Status = Fail
		case all.NotRun > 0:
			r.Status = NotRun
		case all.Pass == 0:
			// No evidence at all, or only UNSUPPORTED: nothing was exercised.
			r.Status = NoEvidence
		default:
			r.Status = Pass
		}
		out = append(out, r)
	}
	return out
}
