//go:build live

package live

import (
	"archive/zip"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// A newly generated, explicitly synthetic archive exercises the same immutable
// replay boundary as historical captures. Relabelling this controlled fixture
// only supplies replay test input; it is never published as actual support.
func glmCorrectiveReplay(t *testing.T, cs *evidence.Case, original string) {
	t.Helper()
	root := filepath.Join("..", "..", ".evidence")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	base, err := os.MkdirTemp(root, "glm-corrective-replay-")
	if err != nil {
		t.Fatal(err)
	}
	base, _ = filepath.Abs(base)
	defer os.RemoveAll(base)
	for _, kind := range []string{"glm-formal", "glm-candidate"} {
		path := filepath.Join(base, "testdata", kind)
		os.MkdirAll(path, 0700)
		if err := os.CopyFS(path, os.DirFS(filepath.Join("testdata", kind))); err != nil {
			t.Fatal(err)
		}
	}
	raw, _ := os.ReadFile(filepath.Join(original, "formal-glm-report.json"))
	var report map[string]any
	json.Unmarshal(raw, &report)
	report["mode"] = "live"
	raw, _ = json.Marshal(report)
	archivePath := filepath.Join(base, "synthetic.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(file)
	err = filepath.WalkDir(original, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(original, path)
		writer, err := w.Create("session/" + filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if rel == "formal-glm-report.json" {
			data = raw
		}
		_, err = writer.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	file.Close()
	zipRaw, _ := os.ReadFile(archivePath)
	index := glmCaptureIndex{Schema: 1, OriginalLiveSourceCommit: report["gitCommit"].(string), SourceManifestSHA256: glmSourcePin,
		Routes: map[string]glmCaptureRecord{"responses": {Archive: "unused", ArchiveSHA256: sha256Text(zipRaw), ReportPath: "session/formal-glm-report.json", ReportSHA256: sha256Text(raw)}}}
	indexRaw, _ := json.Marshal(index)
	indexPath := filepath.Join(base, "testdata/glm-formal/live-captures.json")
	os.WriteFile(indexPath, indexRaw, 0600)
	os.WriteFile(filepath.Join(base, "testdata/glm-formal/live-captures.sha256"), []byte(sha256Text(indexRaw)+"\n"), 0600)
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMCaptureReplay$", "-test.timeout=80s")
	cmd.Dir = base
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_GLM_ROUTE=responses", "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_REPLAY_ARCHIVE=" + archivePath, "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "replayed")}
	output, err := cmd.CombinedOutput()
	cs.Record("new-capture-replay-exit", map[string]any{"success": err == nil, "output": string(output)})
	cs.Check("new-current capture replays Result Observer wire and cancellation exactly", err == nil, "%v %s", err, output)
	matches, _ := filepath.Glob(filepath.Join(base, "replayed", "*", "capture-glm-replay-report.json"))
	if !cs.Check("new replay has own report", len(matches) == 1, "reports=%d", len(matches)) {
		return
	}
	data, _ := os.ReadFile(matches[0])
	cs.Fixture("new-capture-replay-report.json", data)
	cs.Check("full new replay comparisons retained", glmRetainCorrectiveChild(cs, filepath.Dir(matches[0]), "replayed-capture") == nil, "replay copy failed")
	var detail glmReport
	json.Unmarshal(data, &detail)
	cs.Check("replay-only zero vendor attempts", detail.Mode == "replay" && detail.Complete && detail.Budget.HTTPAttempts == 0 && detail.Budget.ReplayAttempts == 22, "budget=%+v", detail.Budget)
	_, e := supportmatrix.NewGLMReport(filepath.Dir(matches[0]))
	cs.Check("synthetic replay never current actual certification", e != nil, "certified synthetic replay")
}

func glmRetainCorrectiveChild(cs *evidence.Case, dir, prefix string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.Join(prefix, rel)
		if err = os.MkdirAll(filepath.Join(run.Dir(), cs.ID(), filepath.Dir(name)), 0700); err != nil {
			return err
		}
		raw, err := readDecisionsFile(path, 16<<20)
		if err != nil {
			return err
		}
		cs.Fixture(filepath.ToSlash(name), raw)
		return nil
	})
}
