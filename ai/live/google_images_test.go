//go:build live

package live

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"testing"
)

// This host-owned probe stays separate from the builtin support declaration.
func loadGoogleImageProbeCatalog() (ai.Catalog, error) {
	data, err := os.ReadFile("../e2e/testdata/google-images/official.json")
	if err != nil {
		return ai.Catalog{}, err
	}
	var facts struct{ Model ai.ImageModel }
	if err = json.Unmarshal(data, &facts); err != nil {
		return ai.Catalog{}, err
	}
	cat := ai.BuiltinCatalog()
	cat.Version += "+google-images-probe-2026-10-08"
	cat.ImageModels = []ai.ImageModel{facts.Model}
	return cat, nil
}
func googleImageProbeCatalog(t *testing.T) ai.Catalog {
	t.Helper()
	cat, err := loadGoogleImageProbeCatalog()
	if err != nil {
		t.Fatal(err)
	}
	return cat
}
func googleImageSmokeOptions() ai.GoogleImagesOptions {
	return ai.GoogleImagesOptions{AspectRatio: ai.Value("1:1"), ImageSize: ai.Value("1K")}
}
func googleImagesScenarios() []scenario {
	return []scenario{
		{id: "generation", run: func(s *session) {
			res, err := s.generateGoogleImage("generation", ai.ImagesRequest{Prompt: "Create an original whimsical still-life photograph of three handmade turquoise ceramic bowls, a yellow paper crane, and a small purple pebble on a pale peach tabletop, soft morning light. No text or logos. Return exactly one image."}, googleImageSmokeOptions())
			if s.ok("generation", err) {
				s.googleImageTurn(res)
			}
		}},
		{id: "reference-edit", run: func(s *session) {
			ref, _ := smokeImageInputs(s.t)
			res, err := s.generateGoogleImage("reference-edit", ai.ImagesRequest{Prompt: "Edit this reference image: change the blue square to a green square, keeping the white background. Return exactly one image.", ReferenceImages: []ai.Image{ref}}, googleImageSmokeOptions())
			if s.ok("reference edit", err) {
				s.googleImageTurn(res)
			}
		}},
	}
}
func (s *session) generateGoogleImage(label string, req ai.ImagesRequest, opts ai.GoogleImagesOptions) (ai.ImagesResult, error) {
	size, _ := opts.ImageSize.Get()
	ratio, _ := opts.AspectRatio.Get()
	mu.Lock()
	allowed := size == "1K" && ratio == "1:1" && len(req.ReferenceImages) <= 1 && imageBudgetAllows(1, "1024x1024")
	if allowed {
		budget.CallsUsed++
		budget.ImagesUsed++
	}
	mu.Unlock()
	if !allowed {
		s.category = supportmatrix.CategoryBudget
		s.check("Google image budget", false, "call, image or 1K resolution budget exhausted")
		return ai.ImagesResult{}, errors.New("Google image smoke budget exhausted")
	}
	s.calls++
	s.images++
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
	verification, validationErr := verifyGoogleImages(res)
	reporting := string(ai.UsageUnreported)
	if len(res.Metadata.Attempts) > 0 {
		reporting = string(res.Metadata.Attempts[len(res.Metadata.Attempts)-1].UsageReporting)
	}
	for _, e := range ex {
		shape := googleImageShape(e.ResponseBody)
		id := ""
		if len(res.Metadata.Attempts) > 0 {
			id = res.Metadata.Attempts[0].ProviderRequestID
		}
		if id == "" && shape.ResponseID != "" {
			id = "interaction=" + shape.ResponseID
		}
		if id != "" {
			s.requestIDs = append(s.requestIDs, id)
		}
		call := supportmatrix.ImageCall{Path: "/v1beta/interactions", Status: e.Status, ProviderRequestID: id, OutputFormat: verification.Format, Requested: 1, UsageReporting: reporting, Google: shape}
		if err == nil && validationErr == nil {
			call.Verified = 1
			call.Width = verification.Width
			call.Height = verification.Height
		}
		if call.OutputFormat == "" {
			call.OutputFormat = "png"
		}
		shape.FixedRequest = s.checkGoogleImagePayload(e, req)
		s.imageCalls = append(s.imageCalls, call)
	}
	name := fmt.Sprintf("attempt-%d-%s", s.attempt, label)
	s.cs.Record(name+"-result", map[string]any{"result": res, "error": errText(err)})
	s.cs.Record(name+"-exchanges", ex)
	s.cs.Record(name+"-image-check", map[string]any{"validation": verification, "error": errText(validationErr), "calls": s.imageCalls})
	if err == nil && validationErr == nil {
		for _, block := range res.Content {
			if img, ok := block.(ai.ImageOutputImage); ok {
				data, _ := base64.StdEncoding.DecodeString(img.Data)
				s.cs.Fixture(name+"-output."+verification.Format, data)
			}
		}
	}
	obs := s.env.rec.observationsOf(scope.RequestID)
	s.cs.Record("observations-"+name, obs)
	s.check("Observer reaches terminal metadata", len(obs) > 0 && obs[len(obs)-1].Kind == ai.ObservationCallFinished, "missing terminal observation")
	return res, err
}
func (s *session) checkGoogleImagePayload(e exchange, req ai.ImagesRequest) bool {
	var body struct {
		Model string
		Store *bool
		Input []struct {
			Type, Text, Data string
			Mime             string `json:"mime_type"`
		}
		Format struct {
			Type     string
			Delivery json.RawMessage `json:"delivery"`
			Size     string          `json:"image_size"`
			Ratio    string          `json:"aspect_ratio"`
		} `json:"response_format"`
	}
	err := json.Unmarshal([]byte(e.RequestBody), &body)
	ok := err == nil && e.Method == "POST" && e.URL == "https://generativelanguage.googleapis.com/v1beta/interactions" && e.RequestHeaders["x-goog-api-key"] == "[REDACTED]" && e.RequestHeaders["authorization"] == "" && e.RequestHeaders["content-type"] == "application/json" && body.Model == s.model && body.Store != nil && !*body.Store && body.Format.Type == "image" && len(body.Format.Delivery) == 0 && body.Format.Size == "1K" && body.Format.Ratio == "1:1" && len(body.Input) == len(req.ReferenceImages)+1
	for i, p := range body.Input {
		if i == 0 {
			ok = ok && p.Type == "text" && p.Text == req.Prompt
		} else {
			ok = ok && p.Type == "image" && p.Mime == req.ReferenceImages[i-1].MimeType && p.Data == req.ReferenceImages[i-1].Data
		}
	}
	return s.check("fixed official stateless inline request", ok, "wrong route, authentication or image output configuration")
}
func (s *session) googleImageTurn(res ai.ImagesResult) {
	m := res.Metadata
	c := s.env.combo
	s.check("image terminal and attribution", res.StopReason == ai.StopReasonStop && m.Resolved && m.Operation == ai.OperationImage && m.ProviderID == c.provider && m.API == c.api && m.ModelID == s.model, "wrong attribution")
	_, err := verifyGoogleImages(res)
	s.check("readable image, format, dimensions and count", err == nil, "%v", err)
	s.check("one HTTP acceptance", len(m.Attempts) == 1 && m.Attempts[0].HTTPStatus == 200, "wrong attempt")
	if len(s.imageCalls) == 0 {
		return
	}
	shape := s.imageCalls[len(s.imageCalls)-1].Google
	id, hasID := res.ResponseID.Get()
	model, hasModel := res.ResponseModel.Get()
	s.check("observed identity presence preserved", hasID == (shape.ResponseID != "") && id == shape.ResponseID && hasModel == (shape.ResponseModel != "") && model == shape.ResponseModel, "identity invented or lost")
	s.check("complete captured terminal and inline evidence", shape.Status == "completed" && shape.InlineImages == 1 && !shape.Continuation && !shape.URI && len(shape.OutputTypes) == len(res.Content), "incomplete terminal shape")
	s.check("nonzero thought inclusion established", shape.ThoughtInModalities != nil && shape.ThoughtTokens > 0, "cannot prove thought billing from this response")
	s.check("native token totals preserved", res.Usage.Input == shape.InputTokens && res.Usage.Output == shape.OutputTokens+shape.ThoughtTokens && res.Usage.TotalTokens == shape.TotalTokens, "usage normalization mismatch")
	// This sourced probe has no cached input: real usage supplies raw modality
	// counts and establishes whether thinking needs a text-rate addition.
	var fields struct {
		Input []struct {
			Modality string
			Tokens   int64
		} `json:"input_tokens_by_modality"`
		Output []struct {
			Modality string
			Tokens   int64
		} `json:"output_tokens_by_modality"`
	}
	json.Unmarshal(shape.Usage, &fields)
	expected := 0.0
	for _, p := range fields.Input {
		expected += float64(p.Tokens) * 1.5 / 1e6
	}
	for _, p := range fields.Output {
		rate := 7.5
		if p.Modality == "image" {
			rate = 30
		}
		expected += float64(p.Tokens) * rate / 1e6
	}
	if shape.ThoughtInModalities != nil && !*shape.ThoughtInModalities {
		expected += float64(shape.ThoughtTokens) * 7.5 / 1e6
	}
	s.check("sourced modality cost includes thinking once", math.Abs(res.Usage.Cost.Total-expected) < 1e-12, "wrong estimated cost")
	s.note("completed, inline, 1K/1:1; response id present=%t model present=%t; thought included in output modalities=%v; usage %s; estimated %.8f USD", hasID, hasModel, shape.ThoughtInModalities != nil && *shape.ThoughtInModalities, m.Attempts[0].UsageReporting, res.Usage.Cost.Total)
}
func googleImageShape(body string) *supportmatrix.GoogleImageEvidence {
	out := &supportmatrix.GoogleImageEvidence{}
	var native struct {
		ID, Model, Status string
		Continuation      json.RawMessage `json:"continuation_token"`
		Usage             json.RawMessage
		Steps             []struct {
			Type    string
			Content []struct {
				Type string
				Data *string
				URI  json.RawMessage `json:"uri"`
			}
		}
	}
	if json.Unmarshal([]byte(body), &native) != nil {
		return out
	}
	out.ResponseID, out.ResponseModel, out.Status, out.Usage = native.ID, native.Model, native.Status, native.Usage
	out.Continuation = len(native.Continuation) > 0
	for _, step := range native.Steps {
		if step.Type == "model_output" {
			for _, p := range step.Content {
				out.OutputTypes = append(out.OutputTypes, p.Type)
				out.URI = out.URI || len(p.URI) > 0
				if p.Type == "image" && p.Data != nil {
					out.InlineImages++
				}
			}
		}
	}
	var u struct {
		Input      int64 `json:"total_input_tokens"`
		Output     int64 `json:"total_output_tokens"`
		Thought    int64 `json:"total_thought_tokens"`
		Total      int64 `json:"total_tokens"`
		Modalities []struct {
			Modality string
			Tokens   int64
		} `json:"output_tokens_by_modality"`
	}
	if json.Unmarshal(native.Usage, &u) == nil {
		out.InputTokens, out.OutputTokens, out.ThoughtTokens, out.TotalTokens = u.Input, u.Output, u.Thought, u.Total
		sum := int64(0)
		seen := []string{}
		for _, m := range u.Modalities {
			sum += m.Tokens
			seen = append(seen, m.Modality)
		}
		if len(seen) == 1 && seen[0] == "image" && u.Thought > 0 && sum <= u.Output {
			// The observed image-only breakdown omits text, including text thought.
			// Missing text stays unknown/partial; it is not filled with a zero or
			// inferred from internal model invocations, which overlap billing totals.
			included := false
			out.ThoughtInModalities = &included
		} else if slices.Contains(seen, "text") && slices.Contains(seen, "image") && u.Thought > 0 {
			included := sum == u.Output+u.Thought
			if sum == u.Output || included {
				out.ThoughtInModalities = &included
			}
		}
	}
	return out
}
