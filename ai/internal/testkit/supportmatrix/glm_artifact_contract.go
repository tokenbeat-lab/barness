package supportmatrix

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"strings"
)

type glmCapturedCase struct {
	ID, Dir, Status string
	Artifacts       map[string]string
}

// The fixed process owns exactly its 22 model-qualified calls and one source
// case. Coherently rehashing an extra case/attempt cannot hide HTTP work from
// the counters. Failed cases may retain a prefix of their owned artifacts.
func checkGLMArtifactContract(r Report, cases []glmCapturedCase, files map[string]string, snapshots map[string]string) error {
	if len(cases) != len(r.Scenarios)+1 || len(files) > 512 {
		return glmBad("finite case/inventory count exceeded")
	}
	root := map[string]bool{"audit.json": true, "formal-glm-report.json": true, "formal-glm-5.3-report.json": true, "formal-glm-5.2-report.json": true, "glm-execution-source.json": true, "manifest.json": true}
	owned := map[string]map[string]bool{}
	source := map[string]bool{"sources.json": true, "assertions.json": true}
	for name := range snapshots {
		source[name] = true
	}
	owned["GLM11-sources"] = source
	results := map[string]Outcome{}
	for _, sc := range r.Scenarios {
		allowed := map[string]bool{}
		for _, name := range []string{"assertions.json", "attempt-1-exchanges.json", "attempt-1-result.json", "budget-after.json", "budget-before.json", "events.json", "observations-1.json", "options.json", "plan.json", "public-request.json", "request-id-source.json", "request.json", "response.sse", "usage-presence.json"} {
			allowed[name] = true
		}
		if strings.HasSuffix(sc.ID, "/"+grokhistory.Scenario) {
			allowed["immutable-input.json"] = true
		}
		if strings.HasSuffix(sc.ID, "/cancel") {
			allowed["capture-scope.json"] = true
		} else {
			allowed["estimated-cost-provenance.json"] = true
		}
		if strings.Contains(sc.ID, "/control-") {
			allowed["single-variable-comparison.json"] = true
		}
		owned[sc.CaseID] = allowed
		results[sc.CaseID] = sc.Outcome
	}
	seen := map[string]bool{}
	for _, c := range cases {
		allowed, ok := owned[c.ID]
		if !ok || seen[c.ID] || c.Dir != c.ID {
			return glmBad("unexpected or duplicate manifest case")
		}
		seen[c.ID] = true
		if c.ID == "GLM11-sources" || results[c.ID] == Pass {
			if len(c.Artifacts) != len(allowed) {
				return glmBad("complete case artifact set differs from fixed contract")
			}
		}
		for name := range c.Artifacts {
			if !allowed[name] {
				return glmBad("unexpected case artifact or hidden attempt")
			}
		}
	}
	for name := range root {
		if files[name] == "" {
			return glmBad("required root report/source artifact missing")
		}
	}
	for name := range files {
		parts := strings.Split(name, "/")
		if len(parts) == 1 {
			if !root[name] {
				return glmBad("unexpected root artifact")
			}
			continue
		}
		if len(parts) != 2 || owned[parts[0]] == nil || !owned[parts[0]][parts[1]] {
			return glmBad("unmanifest or hidden capture artifact")
		}
	}
	return nil
}
