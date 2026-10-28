package evidence

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// These finite byte transitions were reviewed for the three retained routes.
// Reasons and replay limits are in .scratch/provider-source-scope/source-review.md.
// This is neither a filename exemption nor execution authority: any later edit,
// unknown addition/removal, SDK change or another route still fails closed. The
// native consumer still checks original inventories, grants, captures and cases.
//
//go:embed reviewed_sources.json
var reviewedSources []byte

type sourceReview struct {
	Provider, API, File, Before, After, Reason string
	InventorySHA256                            string
}

func reviewedSourceChange(file, before, after, provider, api string, measured map[string]string) bool {
	var reviews []sourceReview
	if json.Unmarshal(reviewedSources, &reviews) != nil {
		return false
	}
	for _, r := range reviews {
		if r.Provider == provider && r.API == api && r.File == file && r.Before == before && r.After == after && r.Reason != "" {
			// The old GLM collector omitted clock/probe. Recovery is allowed
			// only for an exact original inventory whose complete execution
			// snapshot independently retained these same bytes. Never infer
			// a missing source from the current checkout or an arbitrary map.
			if r.InventorySHA256 != "" {
				raw, err := json.Marshal(measured)
				if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != r.InventorySHA256 {
					continue
				}
			}
			return true
		}
	}
	return false
}

// VerifyVisualSource additionally protects the example's generator and scorer.
// Only its read-only verifier may evolve without a measurement-source review.
func VerifyVisualSource(measured, current map[string]string, provider, api string) error {
	for _, files := range []map[string]string{measured, current} {
		for name := range files {
			if strings.HasPrefix(name, "ai/examples/visualeval/") && name != "ai/examples/visualeval/verify.go" && measured[name] != current[name] && !reviewedSourceChange(name, measured[name], current[name], provider, api, measured) {
				return fmt.Errorf("evaluation execution source changed: %s", name)
			}
		}
	}
	return VerifyRouteSource(measured, current, provider, api)
}
