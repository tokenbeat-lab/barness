//go:build live

package live

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestJevFormalProtocolProcess(t *testing.T) {
	cs := run.Case(t, "UC13-formal-protocol-process")
	formalProtocolBundle(t, cs)
}

func formalProtocolBundle(t *testing.T, cs *evidence.Case) string {
	t.Helper()
	base := filepath.Join(run.Dir(), t.Name()+"-source")
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	proof := filepath.Join(base, "proof.json")
	raw, _ := json.Marshal(multimodalAuthorization{Route: "jev", Model: "jev-1.14.0", Service: "https://api.typesafe.ai/v1/systemone", AccountAlias: "controlled", RegionAlias: "offline", AuthorizationReference: "ticket13-controlled", DedicatedCredential: true, MaxCalls: 12, MaxAttempts: 12})
	if err := os.WriteFile(proof, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMultimodalFormal$", "-test.timeout=25s")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TMPDIR=" + os.TempDir(), "BARNESS_MM_ROUTE=jev", "BARNESS_MM_FORMAL=1", "BARNESS_MM_CONTROLLED=success", "BARNESS_MM_LIVE=1", "BARNESS_MM_KEY=controlled-mm-credential", "BARNESS_MM_AUTHORIZATION=" + proof, "BARNESS_AI_EVIDENCE_DIR=" + base}
	out, err := cmd.CombinedOutput()
	cs.Record("process", map[string]any{"failed": err != nil})
	paths, _ := filepath.Glob(filepath.Join(base, "*", "multimodal-report.json"))
	if !cs.Check("formal public producer runs", err == nil && len(paths) == 1, "failed: %v", err) {
		return ""
	}
	var r multimodalExecution
	raw, _ = os.ReadFile(paths[0])
	json.Unmarshal(raw, &r)
	cs.Fixture("formal-protocol-report.json", raw)
	cat := ai.BuiltinCatalog()
	hash, _ := cat.Hash()
	cs.Check("independent formal identity and full two-entry capability set", r.Kind == "jev-formal-protocol" && r.CatalogHash == hash && r.CatalogVersion == cat.Version && r.Complete && len(r.Expected) == 12 && len(r.Scenarios) == 12, "wrong identity")
	catPath := filepath.Join(base, "current-catalog.json")
	raw, _ = json.Marshal(cat)
	os.WriteFile(catPath, raw, 0600)
	verify := exec.Command("python3", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py", "--formal-catalog", catPath, filepath.Dir(paths[0]))
	out, err = verify.CombinedOutput()
	cs.Record("verification", map[string]any{"failed": err != nil, "output": string(out)})
	cs.Check("formal evidence independently verifies", err == nil, "%s", out)
	verify = exec.Command("python3", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py", filepath.Dir(paths[0]))
	out, err = verify.CombinedOutput()
	cs.Check("candidate verifier refuses formal identity", err != nil, "%s", out)
	return filepath.Dir(paths[0])
}

func TestJevFormalBundleProcess(t *testing.T) {
	cs := run.Case(t, "UC13-formal-bundle-process")
	protocol := formalProtocolBundle(t, cs)
	if protocol == "" {
		return
	}
	effects := filepath.Join(run.Dir(), "formal-effects-source")
	visual := exec.Command("go", "run", "./ai/examples/visualeval", "-formal", "-route", "jev", "-out", effects)
	visual.Dir = "../.."
	visual.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TMPDIR=" + os.TempDir()}
	out, err := visual.CombinedOutput()
	if err != nil {
		t.Fatalf("effects: %v %s", err, out)
	}
	bundle := filepath.Join(run.Dir(), "formal-support-package")
	cmd := exec.Command("go", "run", "./ai/live/cmd/jevformal", "-protocol", protocol, "-effects", effects, "-out", bundle)
	cmd.Dir = "../.."
	cmd.Env = visual.Env
	out, err = cmd.CombinedOutput()
	cs.Record("bundle-producer", map[string]any{"failed": err != nil})
	if !cs.Check("public command seals independent model evidence", err == nil, "formal bundle command failed") {
		return
	}
	report, loadErr := supportmatrix.LoadReport(filepath.Join(bundle, supportmatrix.ReportFile))
	state, imageBytes, questions, images := 0, 0, 0, 0
	for _, scenario := range report.Scenarios {
		state += scenario.StateBytes
		imageBytes += scenario.DecodedImageBytes
		questions += scenario.Questions
		images += scenario.Images
	}
	cs.Check("formal scenario counters reconcile to actual protocol budget", loadErr == nil && state == report.Budget.StateBytesSubmitted && imageBytes == report.Budget.DecodedImageBytesSubmitted && questions == report.Budget.QuestionsUsed && images == report.Budget.ImagesUsed, "missing actual per-scenario resources")
	cmd = exec.Command("go", "run", "./ai/live/cmd/jevformal", "-verify", bundle)
	cmd.Dir = "../.."
	cmd.Env = visual.Env
	out, err = cmd.CombinedOutput()
	cs.Record("bundle-verifier", map[string]any{"failed": err != nil})
	cs.Check("sealed controlled pipeline verifies without real claim", err == nil, "verify failed")
	cmd = exec.Command("go", "run", "./ai/live/cmd/jevformal", "-verify", bundle, "-require-live")
	cmd.Dir = "../.."
	cmd.Env = visual.Env
	out, err = cmd.CombinedOutput()
	cs.Check("controlled bundle rejected as real support", err != nil, "controlled promoted")
	matrixPath := filepath.Join(run.Dir(), "atomic-matrix.json")
	m := supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses"}, {Combo: supportmatrix.JevFormalCombo, Operation: "classifier", Provider: "typesafe", API: "typesafe-system-one", Model: "jev-1.14.0"}}}
	before, _ := json.Marshal(m)
	os.WriteFile(matrixPath, before, 0600)
	good := supportmatrix.Report{Schema: 3, Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses", FinishedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Expected: []string{"text-stream"}, Scenarios: []supportmatrix.ScenarioResult{{ID: "text-stream", Outcome: supportmatrix.Pass}}}
	goodPath := filepath.Join(run.Dir(), "older-good-report.json")
	raw, _ := json.Marshal(good)
	os.WriteFile(goodPath, raw, 0600)
	cmd = exec.Command("go", "run", "./ai/live/cmd/supportmatrix", "-matrix", matrixPath, goodPath, bundle)
	cmd.Dir = "../.."
	cmd.Env = visual.Env
	_, err = cmd.CombinedOutput()
	after, _ := os.ReadFile(matrixPath)
	cs.Check("public batch merge refuses controlled and never partially writes", err != nil && bytes.Equal(before, after), "partial matrix update")
	formalTamperCases(t, cs, bundle, visual.Env)
	_ = out
}
