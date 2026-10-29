package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

func TestXAIImagesGenericProjectionCannotPublish(t *testing.T) {
	ev := run.Case(t, "XAIIMG-qualification-projection-refused")
	dir := t.TempDir()
	matrix := supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "xai-images", Operation: "image", Provider: "xai", API: "xai-images"}}}
	before, _ := supportmatrix.Encode(matrix)
	matrixPath := filepath.Join(dir, "matrix.json")
	mustWriteXAINative(t, matrixPath, before)
	cat := xai.ImagesCandidateCatalog()
	hash, _ := cat.Hash()
	at := time.Date(2026, 10, 29, 1, 20, 3, 0, time.UTC)
	report := supportmatrix.Report{Schema: 3, Combo: "xai-images", Operation: "image", Provider: "xai", API: "xai-images", Endpoint: xaiimages.Endpoint, Model: xaiimages.Model, SDK: xaiimages.SDK, AccountAlias: "synthetic", GitCommit: "724e988132ac5bf6ac393db156e07081eb0e2427", CatalogVersion: cat.Version, CatalogHash: hash, StartedAt: at, FinishedAt: at.Add(time.Minute), Expected: xaiimages.Expected(), Budget: supportmatrix.Budget{MaxCalls: 6, MaxAttempts: 6, HTTPAttempts: 6, CallsUsed: 6, MaxImages: 7, ImagesUsed: 7, CallSeconds: 180, ProcessSeconds: 1200}}
	for _, scene := range xaiimages.Plan() {
		report.Scenarios = append(report.Scenarios, supportmatrix.ScenarioResult{ID: scene.ID, CaseID: "XAIIMG-live-" + scene.ID, Model: xaiimages.Model, Outcome: supportmatrix.Pass, Calls: 1, Attempts: 1, Images: scene.Outputs, DurationMS: 1, ProviderRequestIDs: []string{"synthetic-" + scene.ID}})
	}
	path := filepath.Join(dir, "live-report.json")
	writeXAINativeJSON(t, path, report)
	cmd := exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/supportmatrix", "-matrix", matrixPath, path)
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	code := imagineProcessExit(err)
	ev.Record("process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(raw), "vendorCalls": 0})
	after, readErr := os.ReadFile(matrixPath)
	ev.Check("a generic PASS projection cannot replace native qualification", code == 1 && readErr == nil && bytes.Equal(before, after), "exit=%d unchanged=%v", code, bytes.Equal(before, after))
}
