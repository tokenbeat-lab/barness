package e2e

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var imagesProtocol = &fixtureProtocol{dir: "openai-images", binding: "images", api: ai.APIOpenAIImages, provider: ai.ProviderOpenAI, key: func(k tenantKey) tenantKey { return k }, account: "acct-tenant-a"}

// Synthetic capabilities/rates prove protocol behavior. Official model
// inclusion and live evidence are issue 11, independent of these fixtures.
func configureImages(c *ai.Config) {
	cat := ai.BuiltinCatalog()
	cat.ImageModels = []ai.ImageModel{{Provider: ai.ProviderOpenAI, API: ai.APIOpenAIImages, ID: "host-image", Name: "Synthetic image model", Input: []ai.Modality{ai.ModalityText, ai.ModalityImage}, Output: []ai.Modality{ai.ModalityImage}, Capabilities: ai.ImageCapabilities{MaxReferenceImages: 16, MaxOutputImages: 10, Sizes: []string{"auto", "1024x1024", "1536x1024"}, Qualities: []string{"auto", "low", "medium", "high"}, TransparentBackground: true}, Pricing: ai.ImagePricing{InputText: ai.Value(2.0), InputImage: ai.Value(3.0), OutputImage: ai.Value(5.0), Source: "synthetic fixture rates"}}}
	c.Catalog = &cat
	c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 16, MaxOutputImages: 10, MaxOutputImageBytes: 1024, MaxTotalOutputImageBytes: 4096}
}
func patchImageModel(t *testing.T, c *ai.Config, patch json.RawMessage) {
	t.Helper()
	var base, changes map[string]json.RawMessage
	mustUnmarshal(t, mustMarshal(t, c.Catalog.ImageModels[0]), &base)
	mustUnmarshal(t, patch, &changes)
	for k, v := range changes {
		base[k] = v
	}
	mustUnmarshal(t, mustMarshal(t, base), &c.Catalog.ImageModels[0])
}
func installImagesBinding(w *world, k tenantKey) {
	b := primaryBinding(k, w.provider.URL())
	b.BindingID, b.API, b.Operation = "images", ai.APIOpenAIImages, ai.OperationImage
	b.AllowedModels = []string{"host-image"}
	w.host.PutBinding(b)
}
func imagesInput(t *testing.T, sc fixtureScenario) ai.ImagesRequest {
	t.Helper()
	var r ai.ImagesRequest
	mustUnmarshal(t, sc.Images, &r)
	return r
}
func imagesOptions(t *testing.T, raw json.RawMessage) ai.ImageOptions {
	t.Helper()
	if raw == nil {
		return nil
	}
	var o ai.OpenAIImagesOptions
	mustUnmarshal(t, raw, &o)
	return o
}
func invokeImagesScenario(t *testing.T, ev *evidence.Case, w *world, sc fixtureScenario) {
	t.Helper()
	res, err := w.client.GenerateImages(ctxFor(t), textScope("req-"+sc.ID), sc.target(), imagesInput(t, sc), imagesOptions(t, sc.Options))
	checkImagesScenario(ev, w, sc, res, err)
}
func checkImagesScenario(ev *evidence.Case, w *world, sc fixtureScenario, res ai.ImagesResult, err error) {
	ev.Record("result", res)
	ev.Record("error", errString(err))
	reqs, x := w.provider.Requests(), sc.Expect
	ev.Record("requests", reqs)
	ev.Check("request count", len(reqs) == x.Requests, "got %d want %d", len(reqs), x.Requests)
	if len(reqs) > 0 && x.Request != nil {
		r := reqs[len(reqs)-1]
		ev.Check("generation endpoint and tenant credential", r.Method == "POST" && r.Path == x.Request.Path && r.Query == "" && r.KeyAlias == tenantA.alias, "got %+v", r)
		ev.Check("final request JSON", jsonEqual(r.Body, x.Request.Body), "got %s", r.Body)
	}
	ev.Check("terminal", string(res.StopReason) == x.StopReason, "got %s", res.StopReason)
	if x.Code != "" {
		var ae *ai.Error
		ev.Check("failure classification", errors.As(err, &ae) && ae.Code == x.Code && ae.Phase == x.Phase, "got %v", err)
		ev.Check("failure clears whole output", len(res.Content) == 0 && ae != nil && res.ErrorMessage == ae.Message, "got %+v", res)
		if ae != nil {
			ev.Check("HTTP failure metadata", ae.HTTPStatus == x.HTTPStatus && ae.ProviderRequestID == x.ProviderRequestID, "got %+v", ae)
		}
	} else {
		ev.Check("success", err == nil, "got %v", err)
	}
	if x.Content != nil {
		ev.Check("ordered independent output blocks", jsonEqual(mustMarshal(ev.T(), res.Content), x.Content), "got %+v", res.Content)
	}
	ev.Check("response identity absent", res.ResponseID.IsZero() && res.ResponseModel.IsZero(), "got %+v", res)
	if x.Usage != nil {
		ev.Check("usage retained and modality priced", reflect.DeepEqual(res.Usage, *x.Usage), "got %+v want %+v", res.Usage, *x.Usage)
	}
	ev.Check("image attribution", res.Metadata.Operation == ai.OperationImage, "got %+v", res.Metadata)
	ev.Check("attempt count", len(res.Metadata.Attempts) == x.Requests, "got %+v", res.Metadata.Attempts)
	if len(res.Metadata.Attempts) > 0 {
		a := res.Metadata.Attempts[len(res.Metadata.Attempts)-1]
		ev.Check("usage axis agrees", a.UsageReporting == x.UsageReporting && reflect.DeepEqual(a.Usage, res.Usage), "got %+v", a)
		ev.Check("vendor ID", a.ProviderRequestID == "req-images", "got %+v", a)
	} else {
		ev.Check("unresolved preflight", !res.Metadata.Resolved, "got %+v", res.Metadata)
		ev.Check("no credential access", w.host.ReadsFor("req-"+sc.ID).Credentials == 0, "got %+v", w.host.ReadsFor("req-"+sc.ID))
	}
}

func TestOpenAIImagesFixtures(t *testing.T) {
	for _, name := range []string{"generation.json", "failures.json"} {
		f, raw := loadFixture(t, imagesProtocol, name)
		for _, sc := range f.Scenarios {
			t.Run(sc.ID, func(t *testing.T) { ev := run.Case(t, "P08-"+sc.ID); runScenario(t, ev, sc, raw, modeComplete) })
		}
	}
}
