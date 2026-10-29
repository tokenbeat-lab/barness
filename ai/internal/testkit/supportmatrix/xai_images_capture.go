package supportmatrix

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
	_ "golang.org/x/image/webp"
)

type xaiImagesExchange struct {
	Method              string            `json:"method"`
	URL                 string            `json:"url"`
	RequestHeaders      map[string]string `json:"requestHeaders"`
	RequestBody         string            `json:"requestBody"`
	Status              int               `json:"status"`
	ResponseHeaderNames []string          `json:"responseHeaderNames"`
	ResponseHeaders     map[string]string `json:"responseHeaders"`
	ResponseBody        string            `json:"responseBody"`
	Truncated           bool              `json:"truncated,omitempty"`
	TransportError      string            `json:"transportError,omitempty"`
	ElapsedMS           int64             `json:"elapsedMs"`
}
type xaiImagesOutput struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}
type xaiImagesResult struct {
	Content       []xaiImagesOutput   `json:"content"`
	ResponseID    ai.Nullable[string] `json:"responseId,omitzero"`
	ResponseModel ai.Nullable[string] `json:"responseModel,omitzero"`
	Usage         ai.Usage            `json:"usage"`
	StopReason    ai.StopReason       `json:"stopReason"`
	ErrorMessage  string              `json:"errorMessage,omitempty"`
	Metadata      ai.CallMetadata     `json:"metadata"`
}
type xaiImagesReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func verifyXAIImagesCase(root *os.Root, n xaiImagesNative, scene xaiimages.Scene, cases []xaiImagesManifestCase, seen map[string]bool, measured xaiImagesMeasuredFacts) (ScenarioResult, int64, error) {
	row := ScenarioResult{ID: scene.ID, CaseID: "XAIIMG-live-" + scene.ID, Model: xaiimages.Model, Outcome: Pass, Calls: 1, Attempts: 1, Images: scene.Outputs}
	var c xaiImagesManifestCase
	for _, candidate := range cases {
		if candidate.ID == row.CaseID {
			c = candidate
		}
	}
	if c.ID == "" || len(c.Artifacts) != 5+scene.Outputs || len(c.Fixtures) != scene.Outputs {
		return row, 0, refuse("Imagine case artifact set incomplete")
	}
	path := func(name string) string { return c.Dir + "/" + name }
	for _, name := range []string{"result.json", "error.json", "http.json", "observations.json", "assertions.json"} {
		if c.Artifacts[name] == "" {
			return row, 0, refuse("Imagine mandatory capture artifact missing")
		}
	}
	var exchanges []xaiImagesExchange
	if e := xaiReadJSON(root, path("http.json"), &exchanges); e != nil || len(exchanges) != 1 {
		return row, 0, refuse("Imagine exactly one complete HTTP exchange required")
	}
	ex := exchanges[0]
	endpoint := "/images/generations"
	if scene.References > 0 {
		endpoint = "/images/edits"
	}
	if ex.Method != "POST" || ex.URL != xaiimages.Endpoint+endpoint || ex.Status != 200 || ex.Truncated || ex.TransportError != "" || ex.ElapsedMS < 0 || ex.ElapsedMS > 180000 || ex.RequestHeaders["authorization"] != "[REDACTED]" || ex.RequestHeaders["content-type"] != "application/json" || ex.ResponseHeaders["x-request-id"] == "" {
		return row, 0, refuse("Imagine native HTTP route or delivery differs")
	}
	for key := range ex.RequestHeaders {
		if key != "authorization" && key != "content-type" && key != "user-agent" {
			return row, 0, refuse("Imagine unapproved request header")
		}
	}
	if e := verifyXAIImagesRequest(ex.RequestBody, scene); e != nil {
		return row, 0, e
	}
	var result xaiImagesResult
	if e := xaiReadJSON(root, path("result.json"), &result); e != nil {
		return row, 0, e
	}
	var callError string
	if e := xaiReadJSON(root, path("error.json"), &callError); e != nil || callError != "" {
		return row, 0, refuse("Imagine successful capture has an error")
	}
	if result.StopReason != ai.StopReasonStop || result.ErrorMessage != "" || len(result.Content) != scene.Outputs || !result.ResponseID.IsZero() || !result.ResponseModel.IsZero() || !reflect.DeepEqual(result.Usage, ai.Usage{}) {
		return row, 0, refuse("Imagine atomic result or missing-token contract differs")
	}
	meta := result.Metadata
	if meta.TenantID != "live-smoke" || meta.JobID != "live-smoke" || meta.RequestID == "" || meta.BindingID != "xai-images" || meta.Operation != ai.OperationImage || !meta.Resolved || meta.ProviderID != ai.ProviderXAI || meta.API != ai.APIXAIImages || meta.ModelID != xaiimages.Model || meta.AccountScopeID != "live-xai-images" || meta.BindingVersion != "live" || meta.CredentialVersion != "live" || meta.CatalogVersion != n.Scope.CatalogVersion || meta.CatalogHash != n.Scope.CatalogHash || len(meta.Attempts) != 1 || !reflect.DeepEqual(meta.NativeStateDowngrades, ai.NativeStateDowngrades{}) {
		return row, 0, refuse("Imagine result attribution differs")
	}
	if seen[meta.RequestID] {
		return row, 0, refuse("Imagine logical call reused across scenarios")
	}
	seen[meta.RequestID] = true
	ticks, e := verifyXAIImagesResponse(root, path, ex.ResponseBody, scene, result, c, measured)
	if e != nil {
		return row, 0, e
	}
	attempt := meta.Attempts[0]
	usd, known := attempt.ReportedCostUSD.Get()
	if attempt.AttemptID != meta.RequestID+"#1" || attempt.HTTPStatus != 200 || attempt.ProviderRequestID != ex.ResponseHeaders["x-request-id"] || attempt.Code != "" || attempt.RetryDelay != 0 || attempt.UsageReporting != ai.UsageUnreported || !reflect.DeepEqual(attempt.Usage, result.Usage) || !known || !xaiAmountEqual(usd, float64(ticks)/1e10) || usd < 0 || usd > scene.MaximumUSD+1e-12 || !reflect.DeepEqual(attempt.CostEstimate, xaiImagesIndependentEstimate(scene)) {
		return row, 0, refuse("Imagine independent token, estimate or reported charge differs")
	}
	if e = verifyXAIImagesObserver(root, path("observations.json"), n, result); e != nil {
		return row, 0, e
	}
	row.ProviderRequestIDs = []string{attempt.ProviderRequestID}
	row.DurationMS = ex.ElapsedMS
	return row, ticks, nil
}
func verifyXAIImagesRequest(body string, scene xaiimages.Scene) error {
	var fields map[string]json.RawMessage
	if regionalDecode([]byte(body), &fields) != nil {
		return refuse("Imagine final request JSON invalid")
	}
	expected := []string{"model", "prompt", "response_format", "n", "quality", "resolution"}
	if scene.AspectRatio != "" {
		expected = append(expected, "aspect_ratio")
	}
	if scene.References == 1 {
		expected = append(expected, "image")
	} else if scene.References > 1 {
		expected = append(expected, "images")
	}
	if len(fields) != len(expected) {
		return refuse("Imagine final request has extra lifecycle fields")
	}
	for _, key := range expected {
		if len(fields[key]) == 0 || string(fields[key]) == "null" {
			return refuse("Imagine final request required value missing")
		}
	}
	var w struct {
		Model, Prompt, Quality, Resolution string
		N                                  int
		ResponseFormat                     string `json:"response_format"`
		AspectRatio                        string `json:"aspect_ratio"`
		Image                              *xaiImagesReference
		Images                             []xaiImagesReference
	}
	if regionalDecode([]byte(body), &w) != nil || w.Model != xaiimages.Model || w.N != scene.Outputs || w.Quality != scene.Quality || w.Resolution != scene.Resolution || w.AspectRatio != scene.AspectRatio || w.ResponseFormat != "b64_json" {
		return refuse("Imagine final charged options differ")
	}
	prompt := "Draw a simple red circle on a plain white background."
	if scene.References > 0 {
		prompt = "Combine the supplied images in their given order. Keep a simple white background."
	}
	if w.Prompt != prompt {
		return refuse("Imagine capture does not contain original selected prompt")
	}
	refs := w.Images
	if w.Image != nil {
		refs = []xaiImagesReference{*w.Image}
	}
	if len(refs) != scene.References || (scene.References == 1) != (w.Image != nil) {
		return refuse("Imagine ordered resource shape differs")
	}
	for i, ref := range refs {
		if ref.Type != "image_url" || strings.ContainsAny(ref.URL, "\r\n") {
			return refuse("Imagine reference is not strict inline data")
		}
		prefix, data, ok := strings.Cut(ref.URL, ";base64,")
		if !ok {
			return refuse("Imagine reference is external")
		}
		raw, e := base64.StdEncoding.Strict().DecodeString(data)
		if e != nil || len(raw) > 512<<10 {
			return refuse("Imagine input image bytes invalid")
		}
		cfg, format, e := image.DecodeConfig(bytes.NewReader(raw))
		if e != nil || cfg.Width < 1 || cfg.Height < 1 || prefix != "data:image/"+format {
			return refuse("Imagine input MIME/header differs")
		}
		want := "aad089a87c8160c6a87bc472bf9a314abda9fb7d8e5aa93732b1577b2da7d10b"
		if scene.ID == "single-edit" {
			want = "613b1633a46e9d3a7c1769c0b323dd5de6a3783b3f6c98e5ab118c34c152e8f3"
		}
		if scene.ID == "ordered-edit" && i == 1 {
			want = "335348a182a3fc65e0a1a59a53555cdea0c2cd60373acf44b091a0304ff2504a"
		}
		if glmHash(raw) != want {
			return refuse("Imagine reference bytes or order differ from original scene")
		}
	}
	return nil
}
func verifyXAIImagesResponse(root *os.Root, path func(string) string, body string, scene xaiimages.Scene, result xaiImagesResult, c xaiImagesManifestCase, measured xaiImagesMeasuredFacts) (int64, error) {
	var fields map[string]json.RawMessage
	if regionalDecode([]byte(body), &fields) != nil || len(fields) != 2 || fields["data"] == nil || fields["usage"] == nil {
		return 0, refuse("Imagine response is not original synchronous inline shape")
	}
	var entries []map[string]json.RawMessage
	var usage map[string]json.RawMessage
	if regionalDecode(fields["data"], &entries) != nil || len(entries) != scene.Outputs || regionalDecode(fields["usage"], &usage) != nil || len(usage) != 1 {
		return 0, refuse("Imagine original response count or token presence differs")
	}
	var ticks int64
	if len(usage["cost_in_usd_ticks"]) == 0 || string(usage["cost_in_usd_ticks"]) == "null" || json.Unmarshal(usage["cost_in_usd_ticks"], &ticks) != nil || ticks < 0 || ticks != measured.Ticks {
		return 0, refuse("Imagine reported ticks must be a nonnegative int64")
	}
	total := 0
	for i, entry := range entries {
		var data, mime string
		if len(entry) != 2 || json.Unmarshal(entry["b64_json"], &data) != nil || json.Unmarshal(entry["mime_type"], &mime) != nil || len(data)%4 != 0 || strings.ContainsAny(data, "\r\n") || len(data) > 12<<20 {
			return 0, refuse("Imagine wire image delivery invalid")
		}
		raw, e := base64.StdEncoding.Strict().DecodeString(data)
		expected := measured.Images[i]
		if e != nil || len(raw) > 8<<20 || glmHash(raw) != expected.SHA256 || mime != expected.MimeType {
			return 0, refuse("Imagine output image bytes invalid")
		}
		total += len(raw)
		cfg, format, e := image.DecodeConfig(bytes.NewReader(raw))
		if e != nil || cfg.Width < 1 || cfg.Height < 1 || mime != "image/"+format || total > 16<<20 {
			return 0, refuse("Imagine output image MIME/raster differs")
		}
		block := result.Content[i]
		name := fmt.Sprintf("output-%d.image", i)
		original, e := regionalRead(root, path(name))
		if e != nil || block.Type != "image" || block.Data != data || block.MimeType != mime || !bytes.Equal(original, raw) || c.Fixtures[name] != glmHash(raw) {
			return 0, refuse("Imagine original output file, wire and public result disagree")
		}
	}
	return ticks, nil
}
func xaiImagesIndependentEstimate(scene xaiimages.Scene) ai.CostEstimate {
	prices := xai.ImagesCandidateCatalog().ImageModels[0].Pricing.PerImage
	input, _ := prices.InputImage.Get()
	low, high := math.Inf(1), math.Inf(-1)
	for _, r := range prices.Output {
		if r.ImageSize != scene.Resolution || scene.Quality != "auto" && r.Quality != scene.Quality {
			continue
		}
		rate, _ := r.Price.Get()
		low = math.Min(low, rate)
		high = math.Max(high, rate)
	}
	min := ai.UsageCost{Input: input * float64(scene.References), Output: low * float64(scene.Outputs)}
	max := ai.UsageCost{Input: min.Input, Output: high * float64(scene.Outputs)}
	min.Total = min.Input + min.Output
	max.Total = max.Input + max.Output
	status := ai.CostResolved
	if scene.Quality == "auto" {
		status = ai.CostBounded
	}
	return ai.CostEstimate{Status: status, Minimum: ai.Value(min), Maximum: ai.Value(max)}
}

// Reuse the retained capture facts rather than duplicate its seven image hashes
// and six integer charges. The pre-consumer fixture hash is an independent pin;
// the outer consumer source snapshot also binds this repository input.
type xaiImagesMeasuredImage struct {
	SHA256   string `json:"sha256"`
	MimeType string `json:"mimeType"`
}
type xaiImagesMeasuredFacts struct {
	ID     string                   `json:"id"`
	Ticks  int64                    `json:"rawTicks"`
	Images []xaiImagesMeasuredImage `json:"images"`
}

func xaiImagesMeasured() (map[string]xaiImagesMeasuredFacts, error) {
	repository, e := minimaxlive.Root()
	if e != nil {
		return nil, e
	}
	root, e := os.OpenRoot(repository)
	if e != nil {
		return nil, e
	}
	defer root.Close()
	raw, e := regionalRead(root, "ai/e2e/testdata/xai-images/live-contract.json")
	if e != nil || glmHash(raw) != "1b76056460aa2c9711d00860814274b3b156bbcc5c2055bde398c623923520c6" {
		return nil, refuse("Imagine independent retained capture changed")
	}
	var fixture struct {
		Cases []xaiImagesMeasuredFacts `json:"cases"`
	}
	if regionalDecode(raw, &fixture) != nil || len(fixture.Cases) != 6 {
		return nil, refuse("Imagine retained capture incomplete")
	}
	out := map[string]xaiImagesMeasuredFacts{}
	for _, scene := range fixture.Cases {
		out[scene.ID] = scene
	}
	return out, nil
}
