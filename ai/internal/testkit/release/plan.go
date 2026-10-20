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
const requiredReleaseItems = `OD-R01 OD-R02 OD-R03 OD-R04 OD-R05 OD-R06 OD-R07 OD-R08 OD-R09 OD-R10 OD-R11 OD-R12 OD-R13 OD-R14 OD-R15 OD-R16 OD-R17 OD-R18 OD-R19 OD-R20 OD07-frozen-decisions C11-ChineseEffect H5-design-load-artifacts T03-mixed-operation-isolation C08-mixed-lifecycle C09-mixed-observer H5-mixed-design-load C10-mixed-hosts V6-google-images-live H2-google-images-catalog C04-google-images-delivery-omission V4-google-images-editing H5-google-images-input-budgets C04-google-images-edit-final V4-google-images-generation C04-google-images-final H5-google-images-budgets T09-google-images-modalities V6-openai-images-live H2-openai-images-catalog V4-openai-images-editing H5-image-input-budgets C04-image-edit-final V4-openai-images-generation H5-image-budgets C04-image-final-payload T09-image-usage V6-typesafe-live H2-typesafe-catalog C08-unary-lifecycle H5-unary-budgets C04-unary-callbacks T09-unary-observer V4-typesafe-mixed H2-typesafe-parity T04-typesafe-final V4-typesafe-choice T01 T02 T03 T04 T05 T06 T07 T08 T09 T10 T11 T12 C01 C02 C03 C04 C05 C06 C07 C08 C09 C10 C11 V4-responses V4-anthropic V4-gemini V4-chat V4-deepseek V6-live H1 H2 H3 H4 H5 D1 D1-examples D2 S6-1-3 S6-4-7 S6-8-10 TS-R01 TS-R02 TS-R03 TS-R04 TS-R05 TS-R06 TS-R07 TS-R08 TS-R09 TS-R10 TS-R11 TS-R12`

const requiredImageItems = `UC-R01 UC-R02 UC-R03 UC-R04 UC-R05 UC-R06 UC-R07 UC-R08 UC-R09 UC-R10 UC-R11 UC-R12 UC-R13 UC-R14 UC-R15 UC-R16 UC-R17 UC-T01 UC-T02 UC-T03 UC-T04 UC-T05 UC-T06 UC-T07 UC-T08 UC-T09 UC-T10 UC-T11 UC-T12 UC-T13 UC-T14 UC-T15 UC-T16 UC-US01 UC-US03 UC-US17 UC-US23 UC-US25 UC-US32 UC-US33 UC-US41 UC-US42 UC-US43 UC-US44 UC-US45 UC-US46 UC-US47 UC-US48 UC-US49 UC-US50 UC-current-live UC-mixed-images UC-visual-effects`

// Every GLM requirement, acceptance group and story is independently required.
const requiredGLMItems = `GLM-R01 GLM-R02 GLM-R03 GLM-R04 GLM-R05 GLM-R06 GLM-R07 GLM-R08 GLM-R09 GLM-R10 GLM-R11 GLM-R12 GLM-R13 GLM-R14 GLM-R15 GLM-R16 GLM-R17 GLM-R18 GLM-R19 GLM-R20 GLM-R21 GLM-R22 GLM-A01 GLM-A02 GLM-A03 GLM-A04 GLM-A05 GLM-A06 GLM-A07 GLM-A08 GLM-A09 GLM-A10 GLM-A11 GLM-A12 GLM-A13 GLM-A14 GLM-A15 GLM-US01 GLM-US02 GLM-US03 GLM-US04 GLM-US05 GLM-US06 GLM-US07 GLM-US08 GLM-US09 GLM-US10 GLM-US11 GLM-US12 GLM-US13 GLM-US14 GLM-US15 GLM-US16 GLM-US17 GLM-US18 GLM-US19 GLM-US20 GLM-US21 GLM-US22 GLM-US23 GLM-US24 GLM-US25 GLM-US26 GLM-US27 GLM-US28 GLM-US29 GLM-US30 GLM-US31 GLM-US32 GLM-US33 GLM-US34 GLM-US35 GLM-US36 GLM-US37 GLM-US38 GLM-US39 GLM-US40 GLM-US41 GLM-US42 GLM-US43 GLM-US44 GLM-US45 GLM-US46 GLM-US47 GLM-US48 GLM-US49 GLM-US50 GLM-US51 GLM-US52 GLM-US53 GLM-US54 GLM-US55 GLM-US56 GLM-US57 GLM-US58 GLM-US59 GLM-US60 GLM-US61 GLM-US62`

// Hash of the approved normalized evidence mappings in traceability.json.
// Editorial text and ordering are excluded. Changing a regex or requirement's
// evidence selection is a release-contract revision, not a way to fill a gap.
const approvedMappingHash = "d79b29621b04b5fa46b7f907cecfa079b051dbef2ff1937d4f84b85a14f03df2"

// CheckReleasePlan validates the complete own live identities and frozen differential acceptance plan before
// the command collects evidence. Like CheckSnapshot, it returns a verified
// input for Evaluate rather than allowing a mapping to assert its own status.
func CheckReleasePlan(m TraceMap) Snapshot {
	s := Snapshot{Name: "release-plan"}
	var problems []string
	for _, combo := range strings.Fields("openai-responses anthropic-messages google-gemini openai-chat deepseek-responses deepseek-chat typesafe-classifier openai-images google-interactions-image openai-decisions typesafe-classifier-jev-1.14.0 zhipu-chat zhipu-responses zhipu-messages minimax-chat minimax-responses minimax-messages zai-chat zai-responses zai-messages minimax-cn-chat minimax-cn-responses minimax-cn-messages") {
		if count(m.LiveCombos, combo) != 1 {
			problems = append(problems, "required live combo "+combo)
		}
	}
	if len(m.LiveCombos) != 23 {
		problems = append(problems, "want exactly twenty-three live identities")
	}
	for _, scenario := range strings.Fields("E01 E02 E03 E04 E05 E06 E07 E08 E09 E10 E11 P01 P02 P03 P04 P05 P06 P07 P08 P09 P10 D1 D2 PRESSURE UC GLM-CHAT GLM-RESPONSES GLM-MESSAGES GLM-ACCEPTANCE MM-CHAT MM-RESPONSES MM-MESSAGES MM-ACCEPTANCE RP-ACCEPTANCE RP-ZAI-CHAT RP-MINIMAX-CN-MESSAGES RP-EXTENSIONS") {
		if count(m.P0, scenario) != 1 {
			problems = append(problems, "required P0 "+scenario)
		}
	}
	for _, scenario := range strings.Fields("P01 P02 P03 P04 P06 P07 P10 MM-MESSAGES RP-ZAI-CHAT RP-MINIMAX-CN-MESSAGES") {
		if count(m.Differential.Required, scenario) != 1 {
			problems = append(problems, "required differential "+scenario)
		}
	}
	if len(m.Differential.Required) != 10 {
		problems = append(problems, "want exactly ten frozen differentials")
	}
	for _, route := range []RouteRef{{Scenario: "P05", Provider: "deepseek", API: "openai-responses"}, {Scenario: "P08", Provider: "openai", API: "openai-images"}, {Scenario: "P09", Provider: "google", API: "google-interactions"}, {Scenario: "GLM-CHAT", Provider: "zhipu", API: "openai-completions"}, {Scenario: "GLM-RESPONSES", Provider: "zhipu", API: "openai-responses"}, {Scenario: "GLM-MESSAGES", Provider: "zhipu", API: "anthropic-messages"}, {Scenario: "MM-CHAT", Provider: "minimax", API: "openai-completions"}, {Scenario: "MM-RESPONSES", Provider: "minimax", API: "openai-responses"}, {Scenario: "RP-EXTENSIONS", Provider: "zai", API: "openai-responses"}, {Scenario: "RP-EXTENSIONS", Provider: "zai", API: "anthropic-messages"}, {Scenario: "RP-EXTENSIONS", Provider: "minimax-cn", API: "openai-completions"}, {Scenario: "RP-EXTENSIONS", Provider: "minimax-cn", API: "openai-responses"}} {
		if !slices.Contains(m.Differential.Routes, route) {
			problems = append(problems, "required extension route "+route.Scenario)
		}
	}
	for _, id := range strings.Fields(requiredReleaseItems + " " + requiredImageItems + " " + requiredGLMItems) {
		if !slices.ContainsFunc(m.Items, func(i TraceItem) bool { return i.ID == id }) {
			problems = append(problems, "required trace item "+id)
		}
	}
	for prefix, total := range map[string]int{"MM-R": 28, "MM-T": 4, "MM-A": 20, "MM-US": 80} {
		for n := 1; n <= total; n++ {
			id := fmt.Sprintf("%s%02d", prefix, n)
			at := slices.IndexFunc(m.Items, func(i TraceItem) bool { return i.ID == id })
			if at < 0 || len(m.Items[at].RequiredCases) == 0 {
				problems = append(problems, "required finite MiniMax trace item "+id)
			}
		}
	}

	for _, prefix := range []string{"REGIONAL-R", "REGIONAL-RP"} {
		for n := 1; n <= 20; n++ {
			id := fmt.Sprintf("%s%02d", prefix, n)
			at := slices.IndexFunc(m.Items, func(i TraceItem) bool { return i.ID == id })
			if at < 0 || len(m.Items[at].RequiredCases) == 0 {
				problems = append(problems, "required finite regional trace item "+id)
			}
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
		ID            string   `json:"id"`
		Scenarios     []string `json:"scenarios"`
		Cases         string   `json:"cases"`
		RequiredCases []string `json:"requiredCases,omitempty"`
	}
	normalized := struct {
		Scenarios map[string]string `json:"scenarios"`
		Items     []item            `json:"items"`
	}{Scenarios: m.Scenarios}
	for _, i := range m.Items {
		scenarios := slices.Clone(i.Scenarios)
		slices.Sort(scenarios)
		required := slices.Clone(i.RequiredCases)
		slices.Sort(required)
		normalized.Items = append(normalized.Items, item{i.ID, scenarios, i.Cases, required})
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
