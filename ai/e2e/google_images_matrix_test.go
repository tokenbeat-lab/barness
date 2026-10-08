package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func googleMatrixReport() (supportmatrix.Matrix, supportmatrix.Report) {
	at := time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)
	m := supportmatrix.Matrix{Schema: 2, Rows: []supportmatrix.Row{{Combo: "google-gemini", Operation: "chat", Provider: "google", API: "google-generative-ai"}, {Combo: "google-interactions-image", Operation: "image", Provider: "google", API: "google-interactions"}}}
	r := supportmatrix.Report{Schema: 2, Combo: "google-interactions-image", Operation: "image", Provider: "google", API: "google-interactions", Model: "gemini-nano-banana-2.1", SDK: "direct HTTP v1beta", Endpoint: "https://generativelanguage.googleapis.com/v1beta", AccountAlias: "controlled@offline", FinishedAt: at, Expected: []string{"generation", "reference-edit"}, Budget: supportmatrix.Budget{MaxCalls: 4, MaxRetries: 1, MaxImages: 4, MaxImageWidth: 1024, MaxImageHeight: 1024, CallsUsed: 2, ImagesUsed: 2, HTTPAttempts: 2}}
	included := false
	for _, id := range r.Expected {
		r.Scenarios = append(r.Scenarios, supportmatrix.ScenarioResult{ID: id, CaseID: "P09-controlled-" + id, Model: r.Model, Outcome: supportmatrix.Pass, Calls: 1, Images: 1, Attempts: 1, ProviderRequestIDs: []string{"interaction=controlled"}, ImageCalls: []supportmatrix.ImageCall{{Path: "/v1beta/interactions", Status: 200, ProviderRequestID: "interaction=controlled", OutputFormat: "png", Requested: 1, Verified: 1, Width: 1024, Height: 1024, UsageReporting: "complete", Google: &supportmatrix.GoogleImageEvidence{FixedRequest: true, ResponseID: "controlled", Status: "completed", OutputTypes: []string{"image"}, InlineImages: 1, InputTokens: 7, OutputTokens: 1120, ThoughtTokens: 22, TotalTokens: 1149, ThoughtInModalities: &included, Usage: json.RawMessage(`{"total_input_tokens":7,"total_output_tokens":1120,"total_thought_tokens":22,"total_tokens":1149,"input_tokens_by_modality":[{"modality":"text","tokens":7},{"modality":"image","tokens":0}],"output_tokens_by_modality":[{"modality":"text","tokens":0},{"modality":"image","tokens":1120}]}`)}}}})
	}
	return m, r
}

// Exercise the published CLI with a good report followed by a bad report:
// refusal must leave the matrix byte-identical, including the good update.
func TestGoogleImagesMatrixCLI(t *testing.T) {
	changes := map[string]func(*supportmatrix.Report){
		"image-only-partial": func(r *supportmatrix.Report) {
			for i := range r.Scenarios {
				a := &r.Scenarios[i].ImageCalls[0]
				g := a.Google
				g.ResponseID = ""
				a.ProviderRequestID = ""
				r.Scenarios[i].ProviderRequestIDs = nil
				a.UsageReporting = "partial"
				g.Usage = json.RawMessage(`{"total_input_tokens":7,"total_output_tokens":1120,"total_thought_tokens":22,"total_tokens":1149,"output_tokens_by_modality":[{"modality":"image","tokens":1120}]}`)
			}
		},
		"pass":           func(r *supportmatrix.Report) {},
		"wrong-provider": func(r *supportmatrix.Report) { r.Provider = "openai" },
		"stale":          func(r *supportmatrix.Report) { r.FinishedAt = r.FinishedAt.Add(-time.Second) },
		"image-only-complete": func(r *supportmatrix.Report) {
			r.Scenarios[0].ImageCalls[0].Google.Usage = json.RawMessage(`{"total_input_tokens":7,"total_output_tokens":1120,"total_thought_tokens":22,"total_tokens":1149,"output_tokens_by_modality":[{"modality":"image","tokens":1120}]}`)
		},
		"wrong-operation":    func(r *supportmatrix.Report) { r.Operation = "chat" },
		"wrong-api":          func(r *supportmatrix.Report) { r.API = "google-generative-ai" },
		"wrong-model":        func(r *supportmatrix.Report) { r.Model = "gemini-3.1-flash-image" },
		"wrong-version":      func(r *supportmatrix.Report) { r.Endpoint = "https://generativelanguage.googleapis.com/v1" },
		"missing-capability": func(r *supportmatrix.Report) { r.Expected = r.Expected[:1]; r.Scenarios = r.Scenarios[:1] },
		"narrowed": func(r *supportmatrix.Report) {
			r.Scenarios = r.Scenarios[:1]
			r.Budget.CallsUsed = 1
			r.Budget.ImagesUsed = 1
			r.Budget.HTTPAttempts = 1
		},
		"config": func(r *supportmatrix.Report) {
			r.Scenarios[0].Outcome = supportmatrix.Fail
			r.Scenarios[0].ErrorCategory = supportmatrix.CategoryConfig
		},
		"unsupported-edit": func(r *supportmatrix.Report) {
			r.Scenarios[1].Outcome = supportmatrix.Unsupported
			r.Scenarios[1].Note = "not supported"
		},
		"over-budget":             func(r *supportmatrix.Report) { r.Budget.ImagesUsed = 5 },
		"missing-raster":          func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Verified = 0 },
		"missing-shape":           func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google = nil },
		"wrong-format":            func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].OutputFormat = "heic" },
		"uri":                     func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google.URI = true },
		"pending":                 func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google.Status = "in_progress" },
		"continuation":            func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google.Continuation = true },
		"fixed-item":              func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google.FixedRequest = false },
		"unproven-thought":        func(r *supportmatrix.Report) { r.Scenarios[0].ImageCalls[0].Google.ThoughtInModalities = nil },
		"false-thought-inclusion": func(r *supportmatrix.Report) { v := true; r.Scenarios[0].ImageCalls[0].Google.ThoughtInModalities = &v },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-live-matrix-cli-"+name)
			dir := t.TempDir()
			path := filepath.Join(dir, "matrix.json")
			m, r := googleMatrixReport()
			before, _ := supportmatrix.Encode(m)
			if err := os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			write := func(name string, r supportmatrix.Report) string {
				p := filepath.Join(dir, name)
				data, _ := json.Marshal(r)
				if err := os.WriteFile(p, data, 0600); err != nil {
					t.Fatal(err)
				}
				return p
			}
			wantPass := name == "pass" || name == "image-only-partial"
			if wantPass {
				change(&r)
			}
			good := write("good.json", r)
			args := []string{"run", "./ai/live/cmd/supportmatrix", "-matrix", path, good}
			if !wantPass {
				r.FinishedAt = r.FinishedAt.Add(time.Second)
				change(&r)
				args = append(args, write("bad.json", r))
			}
			cmd := exec.CommandContext(ctxFor(t), "go", args...)
			cmd.Dir = "../.."
			output, err := cmd.CombinedOutput()
			ev.Record("command", map[string]any{"output": string(output), "refused": err != nil})
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if wantPass {
				var got supportmatrix.Matrix
				mustUnmarshal(t, after, &got)
				ev.Record("matrix", got)
				ev.Check("only own image row passes", err == nil && got.Rows[1].AllPassedAt != nil && got.Rows[0].LastRunAt == nil, "err=%v", err)
			} else {
				ev.Check("invalid later report cannot partly write", err != nil && bytes.Equal(before, after), "err=%v unchanged=%v", err, bytes.Equal(before, after))
			}
		})
	}
}
