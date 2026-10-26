package release

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// expectedCapabilities is the release contract, independent of a supplied
// report's Expected list. Keep extras in sync with the live scenario registry;
// a missing scenario must fail closed rather than shrink the release.
func expectedCapabilities(combo string) []string {
	if expected := supportmatrix.RegionalExpected(combo); len(expected) > 0 {
		return expected
	}
	if strings.HasPrefix(combo, "minimax-") {
		return supportmatrix.MiniMaxExpected()
	}
	if expected := supportmatrix.GLMExpected(combo); len(expected) > 0 {
		return expected
	}
	switch combo {
	case "xai-responses":
		return supportmatrix.XAIExpected()
	case "openai-decisions":
		return supportmatrix.DecisionsExpected()
	case "typesafe-classifier":
		return supportmatrix.ClassifierExpected()
	case supportmatrix.JevFormalCombo:
		return supportmatrix.JevExpected()
	case "openai-images":
		return supportmatrix.ImagesExpected()
	case "google-interactions-image":
		return supportmatrix.GoogleImagesExpected()
	}
	return supportmatrix.GenericExpected(combo)
}

func capabilityProblems(row supportmatrix.Row) []string {
	var problems []string
	expected := expectedCapabilities(row.Combo)
	if len(expected) == 0 || len(row.Capabilities) != len(expected) {
		problems = append(problems, "incomplete capability set")
	}
	for _, id := range expected {
		matches := 0
		for _, c := range row.Capabilities {
			if c.ID != id {
				continue
			}
			matches++
			if c.Model == "" || ((row.Operation != "chat" || row.Provider == "minimax") && c.Model != row.Model) {
				problems = append(problems, id+": missing or substituted capability model")
			}
			if c.Outcome == supportmatrix.Unsupported && (c.Note == "" || row.Provider == "xai" || row.Provider == "zhipu" || row.Provider == "minimax" || row.Provider == "zai" || row.Provider == "minimax-cn" || (row.Operation != "chat" && !(row.Combo == "openai-images" && id == "mask-edit"))) {
				problems = append(problems, id+": required capability cannot be UNSUPPORTED")
			}
		}
		if matches != 1 {
			problems = append(problems, fmt.Sprintf("capability %s occurs %d times", id, matches))
		}
	}
	for _, c := range row.Capabilities {
		if !slices.Contains(expected, c.ID) {
			problems = append(problems, "unexpected capability "+c.ID)
		}
	}
	return problems
}

func staleSDK(sdk string, modules map[string]string) string {
	fields := strings.Fields(sdk)
	if len(fields) < 2 || !strings.Contains(fields[0], "/") {
		return ""
	}
	if now, ok := modules[fields[0]]; ok && now != fields[1] {
		return fmt.Sprintf("last complete pass used %s %s, go.mod now pins %s; rerun the live smoke", fields[0], fields[1], now)
	}
	return ""
}
