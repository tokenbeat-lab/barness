package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The retained image inventory predates the catalog migration. The historical
// migration comparison uses the complete pre-policy execution source from its
// actual receipt. The later policy edit is not an automatic qualification.
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
	historicalRaw, err := os.ReadFile("testdata/xai-images/pre-model-policy-source.json")
	if err != nil {
		t.Fatal(err)
	}
	fixtureSHA := sha256.Sum256(historicalRaw)
	if hex.EncodeToString(fixtureSHA[:]) != "ff2820d91594825bad54b85dda78d0d143a01c1817196f1ee8fb6f7366e3b598" {
		t.Fatal("historical source fixture changed")
	}
	var historical struct {
		Receipt       string            `json:"receipt"`
		ReceiptSHA256 string            `json:"receiptSha256"`
		SourceCommit  string            `json:"sourceCommit"`
		Files         map[string]string `json:"files"`
	}
	if err := json.Unmarshal(historicalRaw, &historical); err != nil {
		t.Fatal(err)
	}
	if historical.Receipt != ".evidence/catalog-model-id-policy/e2e/20261102T041548.575496Z-zus0m6ad/receipt.json" || historical.ReceiptSHA256 != "784ce7876e163ac8993c3ab540b914b72157437f5792c86489365de8110eb03c" || historical.SourceCommit != "9584d3d6f095aefa4ab73bdc75c39a705b7931d6" || len(historical.Files) == 0 {
		t.Fatal("historical source provenance mismatch")
	}
	for _, mode := range []string{"reviewed-historical", "policy-current", "later-helper", "changed-origin", "wrong-provider", "wrong-api", "missing-runtime"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CAT-SOURCE-"+mode)
			old, now := maps.Clone(measured), maps.Clone(historical.Files)
			if mode == "policy-current" {
				now = maps.Clone(current)
			}
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
			ev.Check("historical migration is applicable; later source needs review", (err == nil) == (mode == "reviewed-historical"), "%v", err)
			ev.Record("review", map[string]any{"mode": mode, "provider": provider, "api": api, "error": errString(err), "vendorCalls": 0})
		})
	}
}
