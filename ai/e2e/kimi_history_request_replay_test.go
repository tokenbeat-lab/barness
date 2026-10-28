package e2e

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"path/filepath"
	"testing"
)

// A resealed, changed Host input must not be hidden by replay's payload hook.
// The current encoder, not a synthetic prompt plus captured-body replacement,
// must justify equivalence for the retained original request.
func TestKimiHistoryRequestReencoding(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary := kimiNineFixtureBinary(t)
	for _, route := range kimiQualificationRoutes[:3] {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KHS-request-reencoding-"+route)
			original := kimiNineFixture(t, ev, binary, route)
			if original == "" {
				return
			}
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			name := supportmatrix.RegionalCaseID(route, "kimi-k3/full-complete") + "/request.json"
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			var input map[string]any
			mustUnmarshal(t, raw, &input)
			input["Messages"] = []any{map[string]any{"content": []any{map[string]any{"text": "resealed invented input"}}}}
			if err := os.WriteFile(filepath.Join(dir, name), mustMarshal(t, input), 0600); err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			b, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
			mustUnmarshal(t, b, &m)
			for _, c := range m["cases"].([]any) {
				c := c.(map[string]any)
				if c["dir"] == filepath.Dir(name) {
					c["artifacts"].(map[string]any)["request.json"] = evidence.Digest(mustMarshal(t, input))
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), mustMarshal(t, m), 0600); err != nil {
				t.Fatal(err)
			}
			if err := supportmatrix.SealRegionalBundle(dir); err != nil {
				t.Fatal(err)
			}
			_, err = supportmatrix.VerifyKimiBaselineCapture(dir, "controlled")
			ev.Check("current Host encoding rejects even self-consistent inventory reseal", err != nil, "accepted changed public request")
			ev.Record("replay-refusal", map[string]any{"error": errString(err), "vendorCalls": 0})
		})
	}
}
