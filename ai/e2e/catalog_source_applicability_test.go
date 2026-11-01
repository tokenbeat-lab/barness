package e2e

import (
	"encoding/json"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The retained image inventory predates the catalog migration. Only these
// exact reviewed bytes may carry its source applicability forward. A later
// edit, different origin, or different route still needs a new review.
func TestCatalogMaintenanceSourceReview(t *testing.T) {
	raw, err := os.ReadFile("testdata/xai-images/measured-source.json")
	if err != nil {
		t.Fatal(err)
	}
	var measured map[string]string
	if err := json.Unmarshal(raw, &measured); err != nil {
		t.Fatal(err)
	}
	current, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"reviewed-current", "later-helper", "changed-origin", "wrong-provider", "wrong-api", "missing-runtime"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CAT-SOURCE-"+mode)
			old, now := maps.Clone(measured), maps.Clone(current)
			provider, api := "xai", "xai-images"
			switch mode {
			case "later-helper":
				now["ai/catalog_definitions.go"] = strings.Repeat("f", 64)
			case "changed-origin":
				old["ai/call_runtime.go"] = strings.Repeat("f", 64)
			case "wrong-provider":
				provider = "deepseek"
			case "wrong-api":
				api = "openai-responses"
			case "missing-runtime":
				delete(now, "ai/images.go")
			}
			err := evidence.VerifyRouteSource(old, now, provider, api)
			ev.Check("only exact reviewed migration source is applicable", (err == nil) == (mode == "reviewed-current"), "%v", err)
			ev.Record("review", map[string]any{"mode": mode, "provider": provider, "api": api, "error": errString(err), "vendorCalls": 0})
		})
	}
}
