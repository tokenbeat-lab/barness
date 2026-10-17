package release

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Public release-plan boundary. The new protocol routes must not erase either
// versioned Jev identity, old P0 requirements or frozen native differentials.
func TestGLMReleaseContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "release", "traceability.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := ParseTraceMap(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.LiveCombos) != 14 || !CheckReleasePlan(plan).Current {
		t.Fatal("complete fourteen-identity formal plan absent")
	}
	for _, combo := range []string{"zhipu-chat", "zhipu-responses", "zhipu-messages", "typesafe-classifier-jev-1.14.0"} {
		t.Run("missing-"+combo, func(t *testing.T) {
			p := plan
			p.LiveCombos = slices.DeleteFunc(slices.Clone(plan.LiveCombos), func(s string) bool { return s == combo })
			if CheckReleasePlan(p).Current {
				t.Fatal("missing own identity accepted")
			}
		})
	}
	for _, combo := range []string{"zhipu-chat", "zhipu-responses", "zhipu-messages"} {
		if len(expectedCapabilities(combo)) != 22 {
			t.Fatal("model-qualified complete GLM set absent")
		}
	}
}
