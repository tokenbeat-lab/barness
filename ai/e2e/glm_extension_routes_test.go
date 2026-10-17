package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// Standard China routes are distinct from both frozen Coding catalogs. A route
// registration is never a successful native pi differential observation.
func TestGLMExtensionRoutes(t *testing.T) {
	for _, route := range glmSharedRoutes {
		t.Run(route.name, func(t *testing.T) {
			ev := run.Case(t, "GLM10-extension-route-"+route.name)
			registered, ok := loadLedger(t).Route(string(ai.ProviderZhipu), string(route.api))
			ev.Record("route", registered)
			ev.Check("China standard route is explicitly an extension", ok && registered.Classification == pioracle.Extension && registered.Ref != "", "%+v", registered)
			var cat ai.Catalog
			switch route.name {
			case "chat":
				cat = glm.ChatCatalog()
			case "responses":
				cat = glm.ResponsesCatalog()
			default:
				cat = glm.MessagesCatalog()
			}
			ev.Record("candidate-catalog", cat)
			client := newGLMSharedWorld(t, nil).client
			for _, model := range []string{"glm-5.3", "glm-5.2"} {
				_, found := client.Catalog().Lookup(ai.ProviderZhipu, route.api, model)
				ev.Check("own Host catalog resolves "+model, found, "%v", found)
			}
		})
	}
}
