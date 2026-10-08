package supportmatrix

import (
	"slices"
	"strings"
)

// ImageCall is bounded protocol and raster validation evidence for one sent
// request. Token consumption, including failed attempts, remains in the bundle's
// public result/attempt records and the report's aggregate token counters.
// These additive schema-2 fields leave classifier/chat report meanings unchanged.
type ImageCall struct {
	Path              string `json:"path"`
	Status            int    `json:"status"`
	ProviderRequestID string `json:"providerRequestId,omitempty"`
	OutputFormat      string `json:"outputFormat"`
	Requested         int    `json:"requested"`
	Verified          int    `json:"verified"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	UsageReporting    string `json:"usageReporting"`
}

func ImagesExpected() []string { return []string{"json-edit", "generation", "mask-edit"} }

func checkImages(r Report) error {
	if r.Combo != "openai-images" {
		return nil
	}
	if r.Operation != "image" || r.Provider != "openai" || r.API != "openai-images" || r.Model != "gpt-image-2.5-sunburst-2026-09-08" || r.Endpoint != "https://api.openai.com/v1" {
		return refuse("image report does not name the fixed official route and model")
	}
	expected := ImagesExpected()
	if len(r.Expected) != len(expected) {
		return refuse("image report lacks its complete capability set")
	}
	for _, id := range expected {
		if !slices.Contains(r.Expected, id) {
			return refuse("image report lacks %q", id)
		}
	}
	if slices.ContainsFunc(r.Scenarios, func(s ScenarioResult) bool { return s.Outcome != NotRun }) && (strings.TrimSpace(r.AccountAlias) == "" || strings.TrimSpace(r.SDK) == "") {
		return refuse("image report lacks account/region alias or implementation provenance")
	}
	b := r.Budget
	if b.MaxCalls != 4 || b.MaxRetries != 1 || b.MaxImages != 4 || b.MaxImageWidth != 1024 || b.MaxImageHeight != 1024 || b.CallsUsed < 0 || b.CallsUsed > b.MaxCalls || b.ImagesUsed < 0 || b.ImagesUsed > b.MaxImages || b.HTTPAttempts < 0 || b.HTTPAttempts > b.CallsUsed || b.InputTokensUsed < 0 || b.OutputTokensUsed < 0 || b.EnvironmentRetries < 0 || b.EnvironmentRetries > len(expected) {
		return refuse("image report has an invalid or exceeded budget")
	}
	calls, images, attempts, retries := 0, 0, 0, 0
	editPassed := false
	lastIndex := -1
	for _, s := range r.Scenarios {
		index := slices.Index(expected, s.ID)
		if index <= lastIndex {
			return refuse("image scenarios do not preserve edit-first ordering")
		}
		lastIndex = index
		if s.Outcome != NotRun && strings.TrimSpace(s.CaseID) == "" {
			return refuse("image capability %q lacks its evidence case", s.ID)
		}
		if s.Model != r.Model || (s.Outcome == Unsupported && s.ID != "mask-edit") {
			return refuse("required image capability %q lacks its fixed model", s.ID)
		}
		if s.Calls < 0 || s.Images < 0 || s.Attempts < 0 || s.Attempts > s.Calls || s.Retries < 0 || s.Retries > b.MaxRetries || s.Calls != s.Images || len(s.ImageCalls) != s.Attempts || s.Retries > max(s.Calls-1, 0) {
			return refuse("image capability %q has inconsistent actual counters", s.ID)
		}
		if s.ID != "json-edit" && s.Outcome != NotRun && !editPassed {
			return refuse("image capability %q ran without a passed JSON edit", s.ID)
		}
		if s.Outcome == Pass || s.Outcome == Unsupported {
			if s.Calls == 0 || s.Attempts == 0 || len(s.ProviderRequestIDs) == 0 || slices.ContainsFunc(s.ProviderRequestIDs, func(id string) bool { return strings.TrimSpace(id) == "" }) {
				return refuse("image capability %q lacks actual request evidence", s.ID)
			}
			a := s.ImageCalls[len(s.ImageCalls)-1]
			if s.Outcome == Pass && (a.Status != 200 || a.Verified != 1 || a.Width != 1024 || a.Height != 1024) {
				return refuse("image capability %q lacks complete raster verification", s.ID)
			}
			note := strings.ToLower(s.Note)
			if s.Outcome == Unsupported && (a.Status != 400 || a.Verified != 0 || !strings.Contains(note, "mask") || !(strings.Contains(note, "not supported") || strings.Contains(note, "unsupported"))) {
				return refuse("mask refusal lacks explicit vendor evidence")
			}
		}
		path := "/v1/images/edits"
		if s.ID == "generation" {
			path = "/v1/images/generations"
		}
		for _, a := range s.ImageCalls {
			if a.Path != path || a.Requested != 1 || a.Verified < 0 || a.Verified > 1 || a.Width < 0 || a.Height < 0 || a.Width > b.MaxImageWidth || a.Height > b.MaxImageHeight || !slices.Contains([]string{"png", "jpeg", "webp"}, a.OutputFormat) || !slices.Contains([]string{"unreported", "partial", "complete"}, a.UsageReporting) || (a.Status > 0 && (strings.TrimSpace(a.ProviderRequestID) == "" || !slices.Contains(s.ProviderRequestIDs, a.ProviderRequestID))) {
				return refuse("image capability %q has invalid wire or resolution evidence", s.ID)
			}
		}
		if s.ID == "json-edit" && s.Outcome == Pass {
			editPassed = true
		}
		calls += s.Calls
		images += s.Images
		attempts += s.Attempts
		retries += s.Retries
	}
	if calls != b.CallsUsed || images != b.ImagesUsed || attempts != b.HTTPAttempts || retries != b.EnvironmentRetries {
		return refuse("image actual counters disagree with its scenarios")
	}
	return nil
}
