package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

var googleImagesProtocol = &fixtureProtocol{dir: "google-images", binding: "images", api: ai.APIGoogleInteractions, provider: ai.ProviderGoogle, key: googleKey, account: "acct-tenant-a-google"}

// Synthetic capabilities and prices exercise a host-supplied catalog. Real
// model inclusion and modality-shape confirmation belong to issue 14.
func configureGoogleImages(c *ai.Config) {
	configureImages(c)
	m := &c.Catalog.ImageModels[0]
	m.Provider, m.API, m.ID, m.Name = ai.ProviderGoogle, googleImagesProtocol.api, "host-google-image", "Synthetic Google image model"
	m.Output = []ai.Modality{ai.ModalityText, ai.ModalityImage}
	m.Capabilities = ai.ImageCapabilities{MaxReferenceImages: 14, MaxOutputImages: 4, AspectRatios: []string{"1:1", "16:9"}, ImageSizes: []string{"1K", "2K", "4K"}}
	m.Pricing.OutputText = ai.Value(4.0)
}
func installGoogleImagesBinding(w *world, k tenantKey) {
	b := geminiBinding(k, w.provider.URL())
	b.BindingID, b.API, b.Operation = "images", googleImagesProtocol.api, ai.OperationImage
	b.AllowedModels = []string{"host-google-image"}
	w.host.PutBinding(b)
}
func firstGoogleImagesScenario(t *testing.T) fixtureScenario {
	t.Helper()
	f, _ := loadFixture(t, googleImagesProtocol, "generation.json")
	return f.Scenarios[0]
}
func TestGoogleImagesFixtures(t *testing.T) {
	for _, name := range []string{"generation.json", "failures.json", "options.json", "usage.json", "editing.json", "editing-invalid-input.json", "editing-failures.json", "editing-usage.json", "editing-options.json"} {
		f, raw := loadFixture(t, googleImagesProtocol, name)
		for _, sc := range f.Scenarios {
			t.Run(sc.ID, func(t *testing.T) {
				ev := run.Case(t, "P09-"+sc.ID)
				runScenario(t, ev, sc, raw, modeComplete)
			})
		}
	}
}
