//go:build live

package live

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// These are runner/process E2E checks. Intentional child FAIL reports remain
// verifiable artifacts; another provider's fixture cannot prove this route.
func TestGoogleImagesHarnessProcess(t *testing.T) {
	if os.Getenv("BARNESS_AI_GOOGLE_IMAGES_CASE") != "" {
		return
	}
	for _, name := range []string{"refused", "environment-retry", "resolution", "reference-count", "no-key", "foreign-key", "narrowed", "exhaustion"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			cmd := exec.Command(os.Args[0], "-test.run=^TestGoogleImagesFailureChild$", "-test.v")
			for _, kv := range os.Environ() {
				key, _, _ := strings.Cut(kv, "=")
				if strings.HasPrefix(key, "BARNESS_AI_LIVE") || strings.HasPrefix(key, "BARNESS_AI_EVIDENCE") || strings.Contains(strings.ToUpper(key), "KEY") {
					continue
				}
				cmd.Env = append(cmd.Env, kv)
			}
			cmd.Env = append(cmd.Env, "BARNESS_AI_GOOGLE_IMAGES_CASE="+name, envLive+"=1", envCombo+"=google-interactions-image", envAlias+"=controlled@offline", "BARNESS_AI_EVIDENCE_DIR="+dir)
			if name != "no-key" {
				cmd.Env = append(cmd.Env, keyVar("google-interactions-image")+"=key:google-image-test")
			}
			if name == "foreign-key" {
				cmd.Env = append(cmd.Env, keyVar("google-gemini")+"=key:foreign-chat")
			}
			output, err := cmd.CombinedOutput()
			fails := slices.Contains([]string{"refused", "resolution", "reference-count", "foreign-key", "exhaustion"}, name)
			if (err != nil) != fails {
				t.Fatalf("unexpected child result: %v\n%s", err, output)
			}
			paths, _ := filepath.Glob(filepath.Join(dir, "*", "live-report.json"))
			if len(paths) != 1 {
				t.Fatal("missing child report")
			}
			report, err := supportmatrix.LoadReport(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			cs := run.Case(t, "P09-live-process-"+name)
			cs.Record("report", report)
			hash, err := ai.BuiltinCatalog().Hash()
			if err != nil {
				t.Fatal(err)
			}
			cs.Check("included smoke uses published catalog and prices", report.CatalogVersion == ai.BuiltinCatalog().Version && report.CatalogHash == hash, "live still used a candidate catalog")
			matrix := supportmatrix.Matrix{Schema: 2, Rows: []supportmatrix.Row{{Combo: "google-interactions-image", Operation: "image", Provider: "google", API: "google-interactions"}}}
			merged, err := supportmatrix.Merge(matrix, report)
			refuses := name == "foreign-key" || name == "narrowed"
			cs.Check("invalid report never partially merges", (err != nil) == refuses, "merge=%v", err)
			if err == nil {
				cs.Check("only complete own success passes", (merged.Rows[0].AllPassedAt != nil) == (name == "environment-retry"), "wrong support outcome")
			}
			cs.Check("all failures and retries stay within four images", report.Budget.CallsUsed <= 4 && report.Budget.ImagesUsed <= 4 && report.Budget.HTTPAttempts <= 4, "overspent")
			if name == "resolution" || name == "reference-count" || name == "no-key" || name == "foreign-key" {
				cs.Check("guard sends nothing", report.Budget.HTTPAttempts == 0 && report.Budget.CallsUsed == 0, "request escaped guard")
			}
			if name == "environment-retry" {
				cs.Check("retry consumption retained", report.Budget.CallsUsed == 3 && report.Budget.EnvironmentRetries == 1 && report.Scenarios[0].Attempts == 2, "lost retry")
			}
			if name == "no-key" {
				cs.Check("missing credential stays NOT_RUN", report.Scenarios[0].Outcome == supportmatrix.NotRun && report.Scenarios[1].Outcome == supportmatrix.NotRun, "invented acceptance")
			}
			if name == "exhaustion" {
				cs.Check("exhausted budget is FAIL", report.Budget.CallsUsed == 4 && report.Scenarios[1].Outcome == supportmatrix.Fail && report.Scenarios[1].ErrorCategory == supportmatrix.CategoryBudget, "exhaustion hidden")
			}
		})
	}
}

func TestGoogleImagesFailureChild(t *testing.T) {
	name := os.Getenv("BARNESS_AI_GOOGLE_IMAGES_CASE")
	if name == "" {
		t.Skip("controlled child only")
	}
	i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "google-interactions-image" })
	if i < 0 {
		t.Fatal("missing combo")
	}
	c := &combos[i]
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	count := 0
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count++
		status := 200
		response := `{"id":"controlled","status":"completed","steps":[{"type":"model_output","content":[{"type":"image","mime_type":"image/png","data":"` + testPNG(t, 1024, 1024) + `"}]}],"usage":{"total_input_tokens":7,"total_output_tokens":1120,"total_thought_tokens":22,"total_tokens":1149,"input_tokens_by_modality":[{"modality":"text","tokens":7},{"modality":"image","tokens":0}],"output_tokens_by_modality":[{"modality":"text","tokens":0},{"modality":"image","tokens":1120}]}}`
		if name == "refused" {
			status = 400
			response = `{"error":{"message":"Image delivery mode is not supported.","code":"invalid_request"}}`
		}
		if name == "environment-retry" && count == 1 || name == "exhaustion" && count == 3 {
			status = 503
			response = `{"error":{"message":"temporary failure"}}`
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response)), Request: req}, nil
	})
	client, err := newClient(c, "key:google-image-test", rec, googleImageProbeCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	env := &liveEnv{combo: c, client: client, rec: rec}
	scenarios := scenariosOf(c)
	if name == "narrowed" {
		scenarios = scenarios[:1]
	}
	for _, sc := range scenarios {
		t.Run(sc.id, func(t *testing.T) {
			cs := run.Case(t, "P09-live-child-"+sc.id)
			if name == "resolution" || name == "reference-count" || name == "exhaustion" {
				id := sc.id
				sc.run = func(s *session) {
					req := ai.ImagesRequest{Prompt: "One circle."}
					opts := googleImageSmokeOptions()
					if name == "resolution" {
						opts.ImageSize = ai.Value("2K")
					}
					if name == "reference-count" {
						ref, _ := smokeImageInputs(t)
						req.ReferenceImages = []ai.Image{ref, ref}
					}
					n := 1
					if name == "exhaustion" {
						n = 2
					}
					for j := 0; j < n; j++ {
						res, err := s.generateGoogleImage(id, req, opts)
						if !s.ok(id, err) {
							return
						}
						s.googleImageTurn(res)
					}
				}
			}
			runScenario(t, cs, c, true, env, sc)
		})
	}
}

func TestGoogleImagesHarnessReplay(t *testing.T) {
	artifact, err := evidence.NewRun("-tags live ./ai/live")
	if err != nil {
		t.Fatal(err)
	}
	c := &combos[slices.IndexFunc(combos, func(c combo) bool { return c.name == "google-interactions-image" })]
	t.Run("dependent-case", func(t *testing.T) { cs := artifact.Case(t, "P09-live-replay-command"); setLiveReplay(cs, c) })
	if err := artifact.Finish(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(artifact.Dir(), "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Cases []struct{ Replay string } }
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Cases) != 1 || !strings.Contains(manifest.Cases[0].Replay, "-run '^TestLive$/^google-interactions-image$'") {
		t.Fatal("replay must select the complete own Google combination")
	}
}
