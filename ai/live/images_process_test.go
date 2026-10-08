//go:build live

package live

import (
	"encoding/json"
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

// Drive the existing runner through a public Client in isolated test processes.
// Intentional child FAIL results are assertions in the parent, with reports left
// as repeatable evidence. No production target override or extra matrix exists.
func TestImagesHarnessProcess(t *testing.T) {
	if os.Getenv("BARNESS_AI_IMAGES_HARNESS_CASE") != "" {
		return
	}
	for _, name := range []string{"edit-refusal", "environment-retry", "exhaustion", "narrowed", "resolution", "quantity"} {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			cmd := exec.Command(os.Args[0], "-test.run=^TestImagesHarnessFailureChild$", "-test.v")
			for _, kv := range os.Environ() {
				key, _, _ := strings.Cut(kv, "=")
				if strings.HasPrefix(key, "BARNESS_AI_LIVE") || strings.HasPrefix(key, "BARNESS_AI_EVIDENCE") || strings.Contains(strings.ToUpper(key), "KEY") {
					continue
				}
				cmd.Env = append(cmd.Env, kv)
			}
			cmd.Env = append(cmd.Env, "BARNESS_AI_IMAGES_HARNESS_CASE="+name, envLive+"=1", envCombo+"=openai-images", envAlias+"=controlled@offline", keyVar("openai-images")+"=key:image-test", "BARNESS_AI_EVIDENCE_DIR="+base)
			output, err := cmd.CombinedOutput()
			wantFail := name == "edit-refusal" || name == "exhaustion" || name == "resolution" || name == "quantity"
			if (err != nil) != wantFail {
				t.Fatalf("child outcome: %v\n%s", err, output)
			}
			paths, _ := filepath.Glob(filepath.Join(base, "*", "live-report.json"))
			if len(paths) != 1 {
				t.Fatal("child did not produce its own report")
			}
			report, err := supportmatrix.LoadReport(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			cs := run.Case(t, "P08-live-harness-process-"+name)
			cs.Record("report", report)
			matrix := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{{Combo: "openai-images", Operation: "image", Provider: "openai", API: "openai-images"}}}
			merged, err := supportmatrix.Merge(matrix, report)
			if err != nil {
				t.Fatal(err)
			}
			cs.Check("only complete own success passes", (merged.Rows[0].AllPassedAt != nil) == (name == "environment-retry"), "incorrect complete outcome")
			cs.Check("bounded consumption", report.Budget.CallsUsed <= 4 && report.Budget.ImagesUsed <= 4 && report.Budget.HTTPAttempts <= 4, "overspent")
			switch name {
			case "edit-refusal":
				cs.Check("mandatory edit is FAIL, downstream NOT_RUN", report.Scenarios[0].Outcome == supportmatrix.Fail && report.Scenarios[1].Outcome == supportmatrix.NotRun && report.Scenarios[2].Outcome == supportmatrix.NotRun && report.Budget.CallsUsed == 1, "refusal concealed")
			case "environment-retry":
				cs.Check("failed request and retry counted", report.Budget.CallsUsed == 4 && report.Budget.ImagesUsed == 4 && report.Budget.EnvironmentRetries == 1 && report.Scenarios[0].Attempts == 2, "retry consumption lost")
			case "exhaustion":
				cs.Check("budget exhaustion fails", report.Scenarios[2].Outcome == supportmatrix.Fail && report.Scenarios[2].ErrorCategory == supportmatrix.CategoryBudget && report.Budget.CallsUsed == 4, "exhaustion skipped")
			case "resolution", "quantity":
				cs.Check("request budget guard sends nothing", report.Budget.HTTPAttempts == 0 && report.Budget.CallsUsed == 0 && report.Scenarios[0].Outcome == supportmatrix.Fail && report.Scenarios[0].ErrorCategory == supportmatrix.CategoryBudget, "request escaped smoke limits")
			case "narrowed":
				cs.Check("cannot bypass first JSON edit", report.Budget.CallsUsed == 0 && report.Scenarios[0].Outcome == supportmatrix.NotRun, "narrowed run sent a request")
			}
		})
	}
}

func TestImagesHarnessFailureChild(t *testing.T) {
	name := os.Getenv("BARNESS_AI_IMAGES_HARNESS_CASE")
	if name == "" {
		t.Skip("controlled child process only")
	}
	i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "openai-images" })
	c := &combos[i]
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	count := 0
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count++
		var body struct {
			OutputFormat string `json:"output_format"`
		}
		json.NewDecoder(req.Body).Decode(&body)
		status := 200
		response := `{"data":[{"b64_json":"` + testPNG(t, 1024, 1024) + `"}],"output_format":"png"}`
		if name == "edit-refusal" {
			status = 400
			response = `{"error":{"message":"JSON edit rejected"}}`
		}
		if (name == "environment-retry" && count == 1) || (name == "exhaustion" && (count == 1 || count == 3)) {
			status = 503
			response = `{"error":{"message":"temporary vendor failure"}}`
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}, "X-Request-Id": []string{"req-controlled-process"}}, Body: io.NopCloser(strings.NewReader(response)), Request: req}, nil
	})
	client, err := newClient(c, "key:image-test", rec, imageProbeCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	env := &liveEnv{combo: c, client: client, rec: rec}
	scenarios := scenariosOf(c)
	if name == "narrowed" {
		scenarios = scenarios[1:2]
	}
	for _, sc := range scenarios {
		t.Run(sc.id, func(t *testing.T) {
			cs := run.Case(t, "P08-live-harness-child-"+sc.id)
			// Keep process scenarios at the same runner seam, using PNG for all
			// controlled responses; real scenarios independently exercise three codecs.
			original := sc.id
			sc.run = func(s *session) {
				ref, _ := smokeImageInputs(t)
				req := imageRequestForProcess(original, ref)
				opts := imageSmokeOptions("png", "low")
				if name == "resolution" {
					opts.Size = ai.Value("1536x1024")
				}
				if name == "quantity" {
					opts.N = ai.Value(2)
				}
				res, err := s.generateImage(original, req, opts)
				if s.ok(original, err) {
					s.imageTurn(res, "png")
				}
			}
			runScenario(t, cs, c, true, env, sc)
		})
	}
}

func imageRequestForProcess(id string, ref ai.Image) ai.ImagesRequest {
	req := ai.ImagesRequest{Prompt: "A geometric shape."}
	if id != "generation" {
		req.ReferenceImages = []ai.Image{ref}
	}
	return req
}
