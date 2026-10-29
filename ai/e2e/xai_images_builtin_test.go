package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
)

func TestXAIImagesBuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-builtin-discovery")
	raw, err := os.ReadFile("testdata/xai-images/official.json")
	if err != nil {
		t.Fatal(err)
	}
	var facts struct {
		ActivationVersion string
		Model             ai.ImageModel
	}
	mustUnmarshal(t, raw, &facts)
	ev.Fixture("official.json", raw)
	reads := 0
	w := newWorld(t, tenantA)
	credentials := credentialResolverFunc(func(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
		reads++
		return ai.Credential{}, ai.ErrAccessDenied
	})
	client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: credentials})
	if !ev.Check("default Client constructs", err == nil, "err=%v", err) {
		return
	}
	cat := client.Catalog()
	model, ok := cat.LookupImage(ai.ProviderXAI, ai.APIXAIImages, imagineModel)
	ev.Record("image-model", model)
	ev.Check("default catalog equals real-accepted fixed model and prices", ok && reflect.DeepEqual(model, facts.Model), "model=%+v", model)
	// Historical activation facts remain fixed; the root catalog may advance
	// later for another Provider. The preceding version is pinned by the Git
	// snapshot at 07de754232c647cc18ae5bce5e8551da42ff14aa.
	ev.Record("historical-activation", map[string]string{"version": facts.ActivationVersion, "previousCatalogVersion": "2026-10-26.1", "previousCatalogCommit": "07de754232c647cc18ae5bce5e8551da42ff14aa"})
	ev.Check("historical activation revision remains original", facts.ActivationVersion == "2026-10-29.1", "activation=%s", facts.ActivationVersion)
	ev.Check("current catalog revision follows its release snapshot", cat.Version == currentCatalogVersion(t) && cat.Version != "2026-10-26.1", "version=%s", cat.Version)
	ev.Check("discovery never reads credentials", reads == 0, "reads=%d", reads)
	candidate := xai.ImagesCandidateCatalog()
	candidateHash, _ := candidate.Hash()
	ev.Check("candidate keeps original independent measurement identity", candidate.Version == "xai-images-candidate-2026-10-29" && candidateHash == "sha256:72f8a0e1a88b9e19402e499203500060e7116386934bf816cd8fff38c2a43670", "candidate=%s %s", candidate.Version, candidateHash)
	if !ok || model.Pricing.PerImage == nil {
		return
	}
	hash := catalogHash(t, cat)
	model.Pricing.PerImage.Output[0].Price = ai.Value(999.0)
	model.Capabilities.AspectRatios[0] = "mutated"
	view := client.Catalog()
	for i := range view.ImageModels {
		if view.ImageModels[i].Provider == ai.ProviderXAI && view.ImageModels[i].API == ai.APIXAIImages {
			view.ImageModels[i].Pricing.PerImage.InputImage = ai.Value(888.0)
		}
	}
	fresh := client.Catalog()
	freshCandidateHash, _ := xai.ImagesCandidateCatalog().Hash()
	ev.Check("public model and catalog views own nested prices and capabilities", catalogHash(t, fresh) == hash && catalogHash(t, ai.BuiltinCatalog()) == hash && freshCandidateHash == candidateHash, "mutable catalog escaped")
}

type imagineLiveContract struct {
	Evidence json.RawMessage `json:"evidence"`
	Cases    []struct {
		ID              string            `json:"id"`
		RequestBody     string            `json:"requestBody"`
		ResponseBody    string            `json:"responseBody"`
		Status          int               `json:"status"`
		ResponseHeaders map[string]string `json:"responseHeaders"`
		ReportedUSD     float64           `json:"reportedUSD"`
		RawTicks        int64             `json:"rawTicks"`
		UsageReporting  ai.UsageReporting `json:"usageReporting"`
		CostEstimate    ai.CostEstimate   `json:"costEstimate"`
		Images          []struct {
			SHA256   string `json:"sha256"`
			MimeType string `json:"mimeType"`
			Bytes    int    `json:"bytes"`
		} `json:"images"`
		ResponseID    ai.Nullable[string] `json:"responseID"`
		ResponseModel ai.Nullable[string] `json:"responseModel"`
	} `json:"cases"`
}

func imagineBuiltinWorld(t *testing.T) *world {
	t.Helper()
	return newImagineWorld(t, func(c *ai.Config) {
		c.Catalog = nil
		c.Policy.MaxOutputBytes = 32 << 20
		c.Policy.MaxImageBytes = 512 << 10
		c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 5, MaxOutputImages: 2, MaxOutputImageBytes: 8 << 20, MaxTotalOutputImageBytes: 16 << 20}
	})
}
func TestXAIImagesBuiltinLiveReplay(t *testing.T) {
	raw, err := os.ReadFile("testdata/xai-images/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture imagineLiveContract
	mustUnmarshal(t, raw, &fixture)
	manifest := run.Case(t, "XAI-IMG-builtin-live-replay-origin")
	manifest.Fixture("live-contract.json", raw)
	manifest.Check("six original real-capture cases", len(fixture.Cases) == 6, "cases=%d", len(fixture.Cases))
	manifest.Record("original-measurement", fixture.Evidence)
	ids := make([]string, 0, len(fixture.Cases))
	ticks := int64(0)
	images := 0
	for _, tc := range fixture.Cases {
		ids = append(ids, tc.ID)
		ticks += tc.RawTicks
		images += len(tc.Images)
	}
	manifest.Check("original six scenarios and consumption inventory", slices.Equal(ids, []string{"generation-low", "generation-medium", "generation-auto", "single-edit", "ordered-edit", "five-reference-edit"}) && ticks == 5100000000 && images == 7, "ids=%v ticks=%d images=%d", ids, ticks, images)
	for _, tc := range fixture.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-builtin-live-replay-"+tc.ID)
			w := imagineBuiltinWorld(t)
			var wire struct {
				Prompt      string
				N           int
				AspectRatio string `json:"aspect_ratio"`
				Resolution  string
				Quality     string
				Image       *struct{ URL string }
				Images      []struct{ URL string }
			}
			mustUnmarshal(t, []byte(tc.RequestBody), &wire)
			req := ai.ImagesRequest{Prompt: wire.Prompt}
			if wire.Image != nil {
				req.ReferenceImages = append(req.ReferenceImages, imagineLiveReference(t, wire.Image.URL))
			}
			for _, ref := range wire.Images {
				req.ReferenceImages = append(req.ReferenceImages, imagineLiveReference(t, ref.URL))
			}
			opts := ai.XAIImagesOptions{N: ai.Value(wire.N), Resolution: ai.Value(wire.Resolution), Quality: ai.Value(wire.Quality)}
			if wire.AspectRatio != "" {
				opts.AspectRatio = ai.Value(wire.AspectRatio)
			}
			enqueue(ev, w, provider.Reply{Status: tc.Status, Header: tc.ResponseHeaders, Script: tc.ResponseBody, Chunks: [][]byte{[]byte(tc.ResponseBody)}})
			result, err := w.client.GenerateImages(ctxFor(t), textScope("builtin-live-"+tc.ID), imagineTarget(), req, opts)
			ev.Record("result", result)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("default Client replays original paid response offline", err == nil && len(result.Content) == len(tc.Images), "err=%v content=%d", err, len(result.Content)) {
				return
			}
			requests := w.provider.Requests()
			path := "/v1/images/generations"
			if len(req.ReferenceImages) > 0 {
				path = "/v1/images/edits"
			}
			ev.Check("original frozen JSON wire and endpoint unchanged", len(requests) == 1 && requests[0].Path == path && string(requests[0].Body) == tc.RequestBody, "requests=%d", len(requests))
			for i, want := range tc.Images {
				image, ok := result.Content[i].(ai.ImageOutputImage)
				decoded, e := base64.StdEncoding.Strict().DecodeString(image.Data)
				digest := sha256.Sum256(decoded)
				ev.Check("original image hash "+string(rune('0'+i)), ok && e == nil && hex.EncodeToString(digest[:]) == want.SHA256 && len(decoded) == want.Bytes && image.MimeType == want.MimeType, "image=%d bytes=%d mime=%s", i, len(decoded), image.MimeType)
			}
			ev.Check("tokens remain actually unreported", reflect.DeepEqual(result.Usage, ai.Usage{}) && len(result.Metadata.Attempts) == 1 && result.Metadata.Attempts[0].UsageReporting == tc.UsageReporting, "usage=%+v", result.Usage)
			if len(result.Metadata.Attempts) == 1 {
				a := result.Metadata.Attempts[0]
				amount, known := a.ReportedCostUSD.Get()
				ev.Check("native charge remains separate from snapshot estimate", known && amount == tc.ReportedUSD && tc.RawTicks >= 0 && float64(tc.RawTicks)/1e10 == amount && reflect.DeepEqual(a.CostEstimate, tc.CostEstimate), "attempt=%+v", a)
				ev.Check("vendor request identity retained", a.ProviderRequestID == tc.ResponseHeaders["x-request-id"], "attempt=%+v", a)
			}
			id, idKnown := result.ResponseID.Get()
			wantID, wantIDKnown := tc.ResponseID.Get()
			model, modelKnown := result.ResponseModel.Get()
			wantModel, wantModelKnown := tc.ResponseModel.Get()
			ev.Check("response identity presence remains original", id == wantID && idKnown == wantIDKnown && model == wantModel && modelKnown == wantModelKnown, "id=%q model=%q", id, model)
			ev.Check("result pins current catalog", result.Metadata.CatalogVersion == currentCatalogVersion(t) && result.Metadata.CatalogHash == catalogHash(t, ai.BuiltinCatalog()), "metadata=%+v", result.Metadata)
		})
	}
}
func TestXAIImagesBuiltinLimits(t *testing.T) {
	for _, name := range []string{"three-outputs", "six-references", "unverified-ratio"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-builtin-preflight-"+name)
			reads := 0
			w := newImagineWorld(t, func(c *ai.Config) {
				c.Catalog = nil
				c.Credentials = credentialResolverFunc(func(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
					reads++
					return ai.Credential{}, ai.ErrAccessDenied
				})
				c.Policy.Image.MaxInputImages = 6
				c.Policy.Image.MaxOutputImages = 10
			})
			req := ai.ImagesRequest{Prompt: "Draw."}
			opts := ai.XAIImagesOptions{N: ai.Value(1), Resolution: ai.Value("1k"), Quality: ai.Value("low")}
			switch name {
			case "three-outputs":
				opts.N = ai.Value(3)
			case "six-references":
				for range 6 {
					req.ReferenceImages = append(req.ReferenceImages, ai.Image{Data: imaginePNG, MimeType: "image/png"})
				}
			case "unverified-ratio":
				opts.AspectRatio = ai.Value("21:9")
			}
			res, err := w.client.GenerateImages(ctxFor(t), textScope("builtin-limit-"+name), imagineTarget(), req, opts)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("wider candidate capability does not authorize builtin call", err != nil && reads == 0 && len(w.provider.Requests()) == 0 && len(res.Metadata.Attempts) == 0, "err=%v credentialReads=%d", err, reads)
		})
	}
}

// Unlike the older OpenAI fixture's fixed PNG input, the real Imagine edit
// captures reuse JPEG output as input. Preserve each original data URI's MIME.
func imagineLiveReference(t *testing.T, url string) ai.Image {
	t.Helper()
	header, data, ok := strings.Cut(url, ";base64,")
	mime, inline := strings.CutPrefix(header, "data:")
	if !ok || !inline || (mime != "image/png" && mime != "image/jpeg" && mime != "image/webp") {
		t.Fatal("real Imagine fixture requires a supported inline image")
	}
	return ai.Image{Data: data, MimeType: mime}
}
