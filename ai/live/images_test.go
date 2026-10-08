//go:build live

package live

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func imagesBudget() supportmatrix.Budget {
	return supportmatrix.Budget{MaxCalls: 4, MaxRetries: 1, MaxImages: 4, MaxImageWidth: 1024, MaxImageHeight: 1024}
}

// A host-owned sourced probe, separate from the support declaration. A failed
// live edit cannot cause the candidate to enter BuiltinCatalog (ADR-0018).
func imageProbeCatalog(t *testing.T) ai.Catalog {
	t.Helper()
	cat, err := loadImageProbeCatalog()
	if err != nil {
		t.Fatal(err)
	}
	return cat
}
func loadImageProbeCatalog() (ai.Catalog, error) {
	data, err := os.ReadFile("../e2e/testdata/openai-images/official.json")
	if err != nil {
		return ai.Catalog{}, err
	}
	var facts struct{ Model ai.ImageModel }
	if err = json.Unmarshal(data, &facts); err != nil {
		return ai.Catalog{}, err
	}
	cat := ai.BuiltinCatalog()
	cat.Version += "+openai-images-probe-2026-10-08"
	// Keep this source distinct even after inclusion: the live probe must be
	// able to test optional mask support that a prior report may have refused.
	cat.ImageModels = []ai.ImageModel{facts.Model}
	return cat, nil
}

func imageSmokeOptions(format, quality string) ai.OpenAIImagesOptions {
	return ai.OpenAIImagesOptions{N: ai.Value(1), Size: ai.Value("1024x1024"), Quality: ai.Value(quality), OutputFormat: ai.Value(format)}
}

// Reserve before the public call, including calls that fail or are retried.
// Binding.Retry stays disabled, so a reservation permits at most one attempt.
func imageBudgetAllows(n int, size string) bool {
	return n == 1 && size == "1024x1024" && budget.CallsUsed < budget.MaxCalls && n <= budget.MaxImages-budget.ImagesUsed
}
func (s *session) generateImage(label string, req ai.ImagesRequest, opts ai.OpenAIImagesOptions) (ai.ImagesResult, error) {
	n, _ := opts.N.Get()
	size, _ := opts.Size.Get()
	mu.Lock()
	ok := imageBudgetAllows(n, size)
	if ok {
		budget.CallsUsed++
		budget.ImagesUsed += n
	}
	mu.Unlock()
	if !ok {
		s.category = supportmatrix.CategoryBudget
		s.check("image budget", false, "call, image or resolution budget exhausted")
		return ai.ImagesResult{}, errors.New("image smoke budget exhausted")
	}
	s.calls++
	s.images += n
	scope := newScope()
	res, err := s.env.client.GenerateImages(s.ctx, scope, s.target(), req, opts)
	ex := s.env.rec.take()
	s.lastExchanges = ex
	s.httpAttempts += len(ex)
	mu.Lock()
	budget.HTTPAttempts += len(ex)
	for _, a := range res.Metadata.Attempts {
		budget.InputTokensUsed += int(a.Usage.Input)
		budget.OutputTokensUsed += int(a.Usage.Output)
	}
	mu.Unlock()
	for _, a := range res.Metadata.Attempts {
		if a.ProviderRequestID != "" {
			s.requestIDs = append(s.requestIDs, a.ProviderRequestID)
		}
	}
	format, _ := opts.OutputFormat.Get()
	verification, validationErr := verifyImages(res, format)
	reporting := string(ai.UsageUnreported)
	if len(res.Metadata.Attempts) > 0 {
		reporting = string(res.Metadata.Attempts[len(res.Metadata.Attempts)-1].UsageReporting)
	}
	path := "/v1/images/generations"
	if len(req.ReferenceImages) > 0 {
		path = "/v1/images/edits"
	}
	for _, e := range ex {
		call := supportmatrix.ImageCall{Path: path, Status: e.Status, ProviderRequestID: e.ResponseHeaders["x-request-id"], OutputFormat: format, Requested: n, UsageReporting: reporting}
		if err == nil && validationErr == nil {
			call.Verified = 1
			call.Width = verification.Width
			call.Height = verification.Height
		}
		u, parseErr := url.Parse(e.URL)
		if parseErr == nil {
			call.Path = u.Path
		}
		s.imageCalls = append(s.imageCalls, call)
		s.check("fixed official JSON image route", parseErr == nil && e.Method == "POST" && u.Scheme == "https" && u.Host == "api.openai.com" && u.Path == path && u.RawQuery == "" && e.RequestHeaders["content-type"] == "application/json", "unexpected image route or protocol")
		s.checkImagePayload(e, req, opts)
	}
	name := fmt.Sprintf("attempt-%d-%s", s.attempt, label)
	s.cs.Record(name+"-result", map[string]any{"result": res, "error": errText(err)})
	s.cs.Record(name+"-exchanges", ex)
	s.cs.Record(name+"-image-check", map[string]any{"validation": verification, "error": errText(validationErr)})
	if err == nil && validationErr == nil {
		block := res.Content[0].(ai.ImageOutputImage)
		data, _ := base64.StdEncoding.DecodeString(block.Data)
		extension := format
		if extension == "jpeg" {
			extension = "jpg"
		}
		s.cs.Fixture(name+"-output."+extension, data)
	}
	observations := s.env.rec.observationsOf(scope.RequestID)
	s.cs.Record("observations-"+name, observations)
	s.check("Observer reaches terminal metadata", len(observations) > 0 && observations[len(observations)-1].Kind == ai.ObservationCallFinished, "missing terminal observation")
	return res, err
}

// Inspect the bounded capture independently of Client result metadata: it must
// still be the same native JSON protocol with fixed budgeted options.
func (s *session) checkImagePayload(e exchange, req ai.ImagesRequest, opts ai.OpenAIImagesOptions) {
	var body struct {
		Model        string
		N            int
		Size         string
		OutputFormat string `json:"output_format"`
		Images       []struct {
			ImageURL string `json:"image_url"`
			FileID   string `json:"file_id"`
		}
		Mask *struct {
			ImageURL string `json:"image_url"`
			FileID   string `json:"file_id"`
		}
		InputFidelity *string `json:"input_fidelity"`
	}
	err := json.Unmarshal([]byte(e.RequestBody), &body)
	format, _ := opts.OutputFormat.Get()
	ok := err == nil && body.Model == s.model && body.N == 1 && body.Size == "1024x1024" && body.OutputFormat == format && body.InputFidelity == nil && len(body.Images) == len(req.ReferenceImages) && (body.Mask != nil) == (opts.Mask != nil)
	for _, ref := range body.Images {
		ok = ok && ref.FileID == "" && strings.HasPrefix(ref.ImageURL, "data:image/png;base64,")
	}
	if body.Mask != nil {
		ok = ok && body.Mask.FileID == "" && strings.HasPrefix(body.Mask.ImageURL, "data:image/png;base64,")
	}
	s.check("fixed budget and inline JSON payload", ok, "unexpected or incomplete native request capture")
}

func imagesScenarios() []scenario {
	return []scenario{
		{id: "json-edit", run: func(s *session) {
			ref, _ := smokeImageInputs(s.t)
			res, err := s.generateImage("json-edit", ai.ImagesRequest{Prompt: "Change the blue square to a green square. Keep the simple white background.", ReferenceImages: []ai.Image{ref}}, imageSmokeOptions("png", "low"))
			if s.ok("JSON edit", err) {
				s.imageTurn(res, "png")
			}
		}},
		{id: "generation", run: func(s *session) {
			res, err := s.generateImage("generation", ai.ImagesRequest{Prompt: "A simple red circle centered on a plain white background."}, imageSmokeOptions("jpeg", "medium"))
			if s.ok("generation", err) {
				s.imageTurn(res, "jpeg")
			}
		}},
		{id: "mask-edit", run: func(s *session) {
			ref, mask := smokeImageInputs(s.t)
			opts := imageSmokeOptions("webp", "low")
			opts.Mask = &mask
			opts.Background = ai.Value("transparent")
			res, err := s.generateImage("mask-edit", ai.ImagesRequest{Prompt: "Replace the square inside the transparent mask area with a green circle on a transparent background.", ReferenceImages: []ai.Image{ref}}, opts)
			if explicitMaskRefusal(err, s.lastExchanges) {
				s.unsupported = "HTTP 400: mask is not supported by the fixed model; see this scenario's JSON request and vendor response"
				return
			}
			if s.ok("mask edit", err) {
				s.imageTurn(res, "webp")
			}
		}},
	}
}

func (s *session) imageTurn(res ai.ImagesResult, format string) {
	c := s.env.combo
	m := res.Metadata
	s.check("image terminal and attribution", res.StopReason == ai.StopReasonStop && m.Resolved && m.Operation == ai.OperationImage && m.ProviderID == c.provider && m.API == c.api && m.ModelID == s.model, "wrong image attribution")
	s.check("one successful HTTP attempt", len(m.Attempts) == 1 && m.Attempts[0].HTTPStatus == 200 && m.Attempts[0].ProviderRequestID != "", "missing HTTP acceptance")
	s.check("no invented image response identity", res.ResponseID.IsZero() && res.ResponseModel.IsZero(), "invented response identity")
	_, err := verifyImages(res, format)
	s.check("readable image, format, size and count", err == nil, "%v", err)
	if len(m.Attempts) == 1 {
		s.note("%s: %s output, 1024x1024; usage %s; known-only cost %.8f USD", format, format, m.Attempts[0].UsageReporting, res.Usage.Cost.Total)
	}
}

// Only a specific mask support rejection is optional. Authentication, generic
// JSON rejection, other parameters and environment failures remain FAIL.
func explicitMaskRefusal(err error, ex []exchange) bool {
	var e *ai.Error
	if !errors.As(err, &e) || e.HTTPStatus != 400 || len(ex) != 1 {
		return false
	}
	var body struct {
		Error struct {
			Message string `json:"message"`
			Param   string `json:"param"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(ex[0].ResponseBody), &body) != nil {
		return false
	}
	message := strings.ToLower(body.Error.Message)
	return (body.Error.Param == "mask" || strings.Contains(message, "mask")) && (strings.Contains(message, "not supported") || strings.Contains(message, "unsupported"))
}

// Synthetic geometric input and PNG alpha mask; no external data or downloads.
// Both inputs are 256 square, much smaller than the bounded 1024 output.
func smokeImageInputs(t *testing.T) (ai.Image, ai.Image) {
	t.Helper()
	bounds := image.Rect(0, 0, 256, 256)
	ref := image.NewNRGBA(bounds)
	mask := image.NewNRGBA(bounds)
	draw.Draw(ref, bounds, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(ref, image.Rect(64, 64, 192, 192), image.NewUniform(color.NRGBA{R: 30, G: 70, B: 220, A: 255}), image.Point{}, draw.Src)
	draw.Draw(mask, bounds, image.NewUniform(color.NRGBA{A: 255}), image.Point{}, draw.Src)
	draw.Draw(mask, image.Rect(64, 64, 192, 192), image.Transparent, image.Point{}, draw.Src)
	encode := func(img image.Image) ai.Image {
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			t.Fatal(err)
		}
		return ai.Image{Data: base64.StdEncoding.EncodeToString(b.Bytes()), MimeType: "image/png"}
	}
	return encode(ref), encode(mask)
}
