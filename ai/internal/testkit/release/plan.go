package release

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// This release contract is intentionally independent of editable trace input.
// Updating the acceptance set requires a reviewed release-contract change;
// deleting evidence or mappings cannot shrink the approved requirements.
const requiredReleaseItems = `C11-ChineseEffect H5-design-load-artifacts T03-mixed-operation-isolation C08-mixed-lifecycle C09-mixed-observer H5-mixed-design-load C10-mixed-hosts V6-google-images-live H2-google-images-catalog C04-google-images-delivery-omission V4-google-images-editing H5-google-images-input-budgets C04-google-images-edit-final V4-google-images-generation C04-google-images-final H5-google-images-budgets T09-google-images-modalities V6-openai-images-live H2-openai-images-catalog V4-openai-images-editing H5-image-input-budgets C04-image-edit-final V4-openai-images-generation H5-image-budgets C04-image-final-payload T09-image-usage V6-typesafe-live H2-typesafe-catalog C08-unary-lifecycle H5-unary-budgets C04-unary-callbacks T09-unary-observer V4-typesafe-mixed H2-typesafe-parity T04-typesafe-final V4-typesafe-choice T01 T02 T03 T04 T05 T06 T07 T08 T09 T10 T11 T12 C01 C02 C03 C04 C05 C06 C07 C08 C09 C10 C11 V4-responses V4-anthropic V4-gemini V4-chat V4-deepseek V6-live H1 H2 H3 H4 H5 D1 D1-examples D2 S6-1-3 S6-4-7 S6-8-10`

// Hash of the approved normalized evidence mappings in traceability.json.
// Editorial text and ordering are excluded. Changing a regex or requirement's
// evidence selection is a release-contract revision, not a way to fill a gap.
const approvedMappingHash = "72aab63148ca8fdca6777410856c91ba7d0818c9a5f828bc70449eba429772ae"

// CheckReleasePlan validates the complete nine-route acceptance plan before
// the command collects evidence. Like CheckSnapshot, it returns a verified
// input for Evaluate rather than allowing a mapping to assert its own status.
func CheckReleasePlan(m TraceMap) Snapshot {
	s := Snapshot{Name: "release-plan"}
	var problems []string
	for _, combo := range strings.Fields("openai-responses anthropic-messages google-gemini openai-chat deepseek-responses deepseek-chat typesafe-classifier openai-images google-interactions-image") {
		if count(m.LiveCombos, combo) != 1 {
			problems = append(problems, "required live combo "+combo)
		}
	}
	if len(m.LiveCombos) != 9 {
		problems = append(problems, "want exactly nine live combinations")
	}
	for _, scenario := range strings.Fields("E01 E02 E03 E04 E05 E06 E07 E08 E09 E10 E11 P01 P02 P03 P04 P05 P06 P07 P08 P09 D1 D2 PRESSURE") {
		if count(m.P0, scenario) != 1 {
			problems = append(problems, "required P0 "+scenario)
		}
	}
	for _, scenario := range strings.Fields("P01 P02 P03 P04 P06 P07") {
		if count(m.Differential.Required, scenario) != 1 {
			problems = append(problems, "required differential "+scenario)
		}
	}
	for _, route := range []RouteRef{{Scenario: "P05", Provider: "deepseek", API: "openai-responses"}, {Scenario: "P08", Provider: "openai", API: "openai-images"}, {Scenario: "P09", Provider: "google", API: "google-interactions"}} {
		if !slices.Contains(m.Differential.Routes, route) {
			problems = append(problems, "required extension route "+route.Scenario)
		}
	}
	for _, id := range strings.Fields(requiredReleaseItems) {
		if !slices.ContainsFunc(m.Items, func(i TraceItem) bool { return i.ID == id }) {
			problems = append(problems, "required trace item "+id)
		}
	}
	if mappingHash(m) != approvedMappingHash {
		problems = append(problems, "approved evidence mappings changed")
	}
	s.Current = len(problems) == 0
	if !s.Current {
		s.Detail = fmt.Sprintf("release acceptance contract incomplete: %s", strings.Join(problems, ", "))
	}
	return s
}

func mappingHash(m TraceMap) string {
	type item struct {
		ID        string   `json:"id"`
		Scenarios []string `json:"scenarios"`
		Cases     string   `json:"cases"`
	}
	normalized := struct {
		Scenarios map[string]string `json:"scenarios"`
		Items     []item            `json:"items"`
	}{Scenarios: m.Scenarios}
	for _, i := range m.Items {
		scenarios := slices.Clone(i.Scenarios)
		slices.Sort(scenarios)
		normalized.Items = append(normalized.Items, item{i.ID, scenarios, i.Cases})
	}
	slices.SortFunc(normalized.Items, func(a, b item) int { return strings.Compare(a.ID, b.ID) })
	raw, _ := json.Marshal(normalized) // Only strings and slices; encoding cannot fail.
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func count(values []string, want string) int {
	n := 0
	for _, v := range values {
		if v == want {
			n++
		}
	}
	return n
}
