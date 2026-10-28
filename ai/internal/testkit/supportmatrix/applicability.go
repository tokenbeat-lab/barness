package supportmatrix

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// RouteCatalog keeps the original snapshot identity while checking only models
// actually measured by this route. It grants no execution permission.
func RouteCatalog(dir string, r Report) (ai.Catalog, error) {
	current := ai.BuiltinCatalog()
	measured, err := evidence.MeasuredCatalog(dir, r.CatalogVersion, r.CatalogHash, current)
	if err != nil {
		return measured, err
	}
	models := []string{r.Model}
	seen := map[string]bool{r.Model: true}
	for _, s := range r.Scenarios {
		// This explicit negative probe proves absence on Chat, never support.
		// A later activation must invalidate that negative measurement.
		if r.Combo == "openai-chat" && s.ID == "pro-model-unavailable" && s.Model == "gpt-5.4-pro" {
			_, oldExists := measured.Lookup(ai.ProviderOpenAI, ai.APIOpenAICompletions, s.Model)
			_, currentExists := current.Lookup(ai.ProviderOpenAI, ai.APIOpenAICompletions, s.Model)
			if oldExists || currentExists {
				return measured, fmt.Errorf("negative model probe changed: %s", s.Model)
			}
			continue
		}
		if s.Outcome == Unsupported || seen[s.Model] {
			continue
		}
		models = append(models, s.Model)
		seen[s.Model] = true
	}
	err = evidence.VerifyRouteCatalog(measured, current, ai.Operation(r.Operation), ai.ProviderID(r.Provider), ai.API(r.API), models)
	return measured, err
}
