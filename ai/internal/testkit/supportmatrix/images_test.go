package supportmatrix

import (
	"errors"
	"reflect"
	"testing"
)

// Public report merger seam. Failure modes precede code in issue 11 FAILURES.md.
func imagesReport() (Matrix, Report) {
	m := baseMatrix()
	m.Rows = append(m.Rows, Row{Combo: "openai-images", Operation: "image", Provider: "openai", API: "openai-images"})
	r := report("openai-images", "openai", "openai-images", t1, pass("json-edit"), pass("generation"), pass("mask-edit"))
	r.Operation, r.Model, r.Endpoint = "image", "gpt-image-2.5-sunburst-2026-09-08", "https://api.openai.com/v1"
	r.Budget = Budget{MaxCalls: 4, MaxRetries: 1, MaxImages: 4, MaxImageWidth: 1024, MaxImageHeight: 1024, CallsUsed: 3, ImagesUsed: 3, HTTPAttempts: 3}
	for i := range r.Scenarios {
		s := &r.Scenarios[i]
		s.Model, s.Calls, s.Attempts, s.Images = r.Model, 1, 1, 1
		s.CaseID = "TEST-P08-" + s.ID
		s.ProviderRequestIDs = []string{"req-test"}
		path := "/v1/images/edits"
		if s.ID == "generation" {
			path = "/v1/images/generations"
		}
		s.ImageCalls = []ImageCall{{Path: path, Status: 200, ProviderRequestID: "req-test", OutputFormat: "png", Requested: 1, Verified: 1, Width: 1024, Height: 1024, UsageReporting: "partial"}}
	}
	return m, r
}
func TestImagesReport(t *testing.T) {
	t.Run("complete-own-evidence", func(t *testing.T) {
		m, r := imagesReport()
		out, err := Merge(m, r)
		if err != nil {
			t.Fatal(err)
		}
		if out.Rows[2].AllPassedAt == nil || !reflect.DeepEqual(out.Rows[:2], m.Rows[:2]) {
			t.Fatal("complete pass crossed routes or missing")
		}
	})
	for name, change := range map[string]func(*Report){
		"blank-alias":  func(r *Report) { r.AccountAlias = " " },
		"blank-sdk":    func(r *Report) { r.SDK = "" },
		"missing-case": func(r *Report) { r.Scenarios[0].CaseID = "" },
		"mask-false-note": func(r *Report) {
			s := &r.Scenarios[2]
			s.Outcome = Unsupported
			s.Note = "mask input is invalid"
			s.ImageCalls[0].Status = 400
			s.ImageCalls[0].Verified = 0
		},
		"operation":              func(r *Report) { r.Operation = "chat" },
		"provider":               func(r *Report) { r.Provider = "deepseek" },
		"api":                    func(r *Report) { r.API = "openai-responses" },
		"alias":                  func(r *Report) { r.Model = "gpt-image-2.5-sunburst" },
		"target":                 func(r *Report) { r.Endpoint = "https://proxy.invalid/v1" },
		"expected-reduced":       func(r *Report) { r.Expected = r.Expected[:1]; r.Scenarios = r.Scenarios[:1] },
		"expected-duplicate":     func(r *Report) { r.Expected[2] = r.Expected[0] },
		"edit-unsupported":       func(r *Report) { r.Scenarios[0].Outcome = Unsupported; r.Scenarios[0].Note = "JSON not supported" },
		"generation-unsupported": func(r *Report) { r.Scenarios[1].Outcome = Unsupported; r.Scenarios[1].Note = "not supported" },
		"overspend":              func(r *Report) { r.Budget.ImagesUsed = 5 },
		"unbounded":              func(r *Report) { r.Budget.MaxImages = 5 },
		"resolution":             func(r *Report) { r.Scenarios[0].ImageCalls[0].Width = 1536 },
		"no-output-check":        func(r *Report) { r.Scenarios[0].ImageCalls[0].Verified = 0 },
		"no-call":                func(r *Report) { r.Scenarios[0].Calls = 0 },
		"no-request-id":          func(r *Report) { r.Scenarios[0].ProviderRequestIDs = nil },
		"wrong-path":             func(r *Report) { r.Scenarios[0].ImageCalls[0].Path = "/v1/responses" },
		"wrong-status":           func(r *Report) { r.Scenarios[0].ImageCalls[0].Status = 400 },
		"wrong-model":            func(r *Report) { r.Scenarios[0].Model = "other" },
		"no-capture":             func(r *Report) { r.Scenarios[0].ImageCalls = nil },
		"counters":               func(r *Report) { r.Budget.HTTPAttempts = 2 },
		"retry-counter":          func(r *Report) { r.Budget.EnvironmentRetries = 1 },
		"edit-after-generation":  func(r *Report) { r.Scenarios[0], r.Scenarios[1] = r.Scenarios[1], r.Scenarios[0] },
	} {
		t.Run(name, func(t *testing.T) {
			m, r := imagesReport()
			before, _ := Encode(m)
			change(&r)
			_, err := Merge(m, r)
			if !errors.Is(err, ErrInvalidReport) {
				t.Fatalf("invalid report accepted: %v", err)
			}
			after, _ := Encode(m)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("refusal mutated matrix")
			}
		})
	}
	t.Run("partial-cannot-pass", func(t *testing.T) {
		m, r := imagesReport()
		r.Scenarios = r.Scenarios[:1]
		r.Budget.CallsUsed = 1
		r.Budget.ImagesUsed = 1
		r.Budget.HTTPAttempts = 1
		out, err := Merge(m, r)
		if err != nil || out.Rows[2].AllPassedAt != nil {
			t.Fatalf("partial fabricated pass: %v", err)
		}
	})
	t.Run("mask-refusal-with-evidence", func(t *testing.T) {
		m, r := imagesReport()
		s := &r.Scenarios[2]
		s.Outcome = Unsupported
		s.Note = "HTTP 400: mask is not supported by this model"
		s.ImageCalls[0].Status = 400
		s.ImageCalls[0].Verified = 0
		s.ImageCalls[0].Width = 0
		s.ImageCalls[0].Height = 0
		out, err := Merge(m, r)
		if err != nil || out.Rows[2].AllPassedAt == nil {
			t.Fatalf("optional refusal not accepted: %v", err)
		}
	})
	t.Run("not-run-preserves-matrix", func(t *testing.T) {
		m, r := imagesReport()
		r.Budget.CallsUsed = 0
		r.Budget.ImagesUsed = 0
		r.Budget.HTTPAttempts = 0
		for i := range r.Scenarios {
			r.Scenarios[i] = ScenarioResult{ID: r.Scenarios[i].ID, Model: r.Model, Outcome: NotRun}
		}
		out, err := Merge(m, r)
		if err != nil || !reflect.DeepEqual(m, out) {
			t.Fatalf("NOT_RUN changed matrix: %v", err)
		}
	})
}
