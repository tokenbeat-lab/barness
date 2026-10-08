package supportmatrix

import (
	"encoding/json"
	"math"
	"slices"
	"strings"
)

func GoogleImagesExpected() []string { return []string{"generation", "reference-edit"} }

// A Google report proves only this operation and protocol. Required image
// capabilities are never optional, and narrowed reports cannot partly publish.
func checkGoogleImages(r Report) error {
	if r.Combo != "google-interactions-image" {
		return nil
	}
	if r.Operation != "image" || r.Provider != "google" || r.API != "google-interactions" || r.Model != "gemini-nano-banana-2.1" || r.Endpoint != "https://generativelanguage.googleapis.com/v1beta" {
		return refuse("Google image report does not name its fixed v1beta route and model")
	}
	expected := GoogleImagesExpected()
	if err := checkExpected(r, expected); err != nil {
		return err
	}
	if len(r.Scenarios) != len(expected) {
		return refuse("Google image report must cover every required capability")
	}
	if slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.Outcome != NotRun }) && (strings.TrimSpace(r.AccountAlias) == "" || strings.TrimSpace(r.SDK) == "") {
		return refuse("Google image report lacks account alias or implementation provenance")
	}
	b := r.Budget
	if b.MaxCalls != 4 || b.MaxImages != 4 || b.MaxRetries != 1 || b.MaxImageWidth != 1024 || b.MaxImageHeight != 1024 || b.CallsUsed < 0 || b.CallsUsed > 4 || b.ImagesUsed < 0 || b.ImagesUsed > 4 || b.HTTPAttempts < 0 || b.HTTPAttempts > b.CallsUsed || b.InputTokensUsed < 0 || b.OutputTokensUsed < 0 || b.EnvironmentRetries < 0 || b.EnvironmentRetries > 2 {
		return refuse("Google image report has an invalid or exceeded budget")
	}
	calls, images, attempts, retries := 0, 0, 0, 0
	for i, s := range r.Scenarios {
		if s.ID != expected[i] || s.Model != r.Model || s.Outcome == Unsupported {
			return refuse("required Google image capability %q has incorrect order, model or outcome", s.ID)
		}
		if s.Outcome != NotRun && strings.TrimSpace(s.CaseID) == "" {
			return refuse("Google image capability %q lacks its evidence case", s.ID)
		}
		if s.Calls < 0 || s.Images != s.Calls || s.Attempts < 0 || s.Attempts > s.Calls || len(s.ImageCalls) != s.Attempts || s.Retries < 0 || s.Retries > 1 || s.Retries > max(s.Calls-1, 0) {
			return refuse("Google image capability %q has inconsistent counters", s.ID)
		}
		if s.Outcome == NotRun && (s.Calls != 0 || s.Attempts != 0) {
			return refuse("NOT_RUN Google image capability consumed requests")
		}
		for _, a := range s.ImageCalls {
			if a.Path != "/v1beta/interactions" || a.Requested != 1 || a.Verified < 0 || a.Verified > 1 || a.Width < 0 || a.Height < 0 || a.Width > 1024 || a.Height > 1024 || !slices.Contains([]string{"png", "jpeg", "webp"}, a.OutputFormat) || !slices.Contains([]string{"complete", "partial", "unreported"}, a.UsageReporting) {
				return refuse("Google image capability %q has invalid wire or raster evidence", s.ID)
			}
		}
		if s.Outcome == Pass {
			if s.Calls == 0 || s.Attempts == 0 {
				return refuse("Google image capability %q has no actual call", s.ID)
			}
			a := s.ImageCalls[len(s.ImageCalls)-1]
			g := a.Google
			if a.Status != 200 || a.Verified != 1 || a.Width != 1024 || a.Height != 1024 || g == nil || !g.FixedRequest || g.Status != "completed" || g.Continuation || g.URI || g.InlineImages != 1 || (g.ResponseModel != "" && g.ResponseModel != r.Model) || (a.ProviderRequestID != "" && !slices.Contains(s.ProviderRequestIDs, a.ProviderRequestID)) {
				return refuse("Google image capability %q lacks synchronous fixed-request inline evidence", s.ID)
			}
			if len(g.OutputTypes) == 0 || slices.ContainsFunc(g.OutputTypes, func(t string) bool { return t != "image" && t != "text" }) || countImageTypes(g.OutputTypes) != 1 || !googleThoughtEvidence(g, a.UsageReporting) {
				return refuse("Google image capability %q lacks complete content or thought evidence", s.ID)
			}
		}
		calls += s.Calls
		images += s.Images
		attempts += s.Attempts
		retries += s.Retries
	}
	if calls != b.CallsUsed || images != b.ImagesUsed || attempts != b.HTTPAttempts || retries != b.EnvironmentRetries {
		return refuse("Google image counters disagree with scenarios")
	}
	return nil
}

func countImageTypes(types []string) int {
	n := 0
	for _, t := range types {
		if t == "image" {
			n++
		}
	}
	return n
}

func googleThoughtEvidence(g *GoogleImageEvidence, reporting string) bool {
	if g.ThoughtInModalities == nil || g.ThoughtTokens <= 0 || g.InputTokens < 0 || g.OutputTokens < 0 || g.TotalTokens < 0 {
		return false
	}
	var u struct {
		Input           *int64 `json:"total_input_tokens"`
		Output          *int64 `json:"total_output_tokens"`
		Thought         *int64 `json:"total_thought_tokens"`
		Total           *int64 `json:"total_tokens"`
		InputModalities []struct {
			Modality string
			Tokens   *int64
		} `json:"input_tokens_by_modality"`
		Modalities []struct {
			Modality string
			Tokens   *int64
		} `json:"output_tokens_by_modality"`
	}
	if json.Unmarshal(g.Usage, &u) != nil || u.Input == nil || u.Output == nil || u.Thought == nil || u.Total == nil || *u.Input != g.InputTokens || *u.Output != g.OutputTokens || *u.Thought != g.ThoughtTokens || *u.Total != g.TotalTokens {
		return false
	}
	seen := map[string]bool{}
	sum := int64(0)
	for _, m := range u.Modalities {
		if (m.Modality != "text" && m.Modality != "image") || seen[m.Modality] || m.Tokens == nil || *m.Tokens < 0 || sum > math.MaxInt64-*m.Tokens {
			return false
		}
		seen[m.Modality] = true
		sum += *m.Tokens
	}

	inputSeen := map[string]bool{}
	for _, m := range u.InputModalities {
		if (m.Modality != "text" && m.Modality != "image") || inputSeen[m.Modality] || m.Tokens == nil || *m.Tokens < 0 {
			return false
		}
		inputSeen[m.Modality] = true
	}
	complete := seen["text"] && seen["image"] && inputSeen["text"] && inputSeen["image"]
	expectedReporting := "partial"
	if complete {
		expectedReporting = "complete"
	}
	if reporting != expectedReporting {
		return false
	}
	// Image-only output details exclude text thought; absent text counts
	// retain partial reporting and cannot be fabricated into complete usage.
	if len(seen) == 1 && seen["image"] && !*g.ThoughtInModalities {
		return sum <= g.OutputTokens
	}
	if !seen["text"] || !seen["image"] {
		return false
	}
	if *g.ThoughtInModalities {
		return g.OutputTokens <= math.MaxInt64-g.ThoughtTokens && sum == g.OutputTokens+g.ThoughtTokens
	}
	return sum == g.OutputTokens
}
