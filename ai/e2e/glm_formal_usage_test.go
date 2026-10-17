package e2e

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func extractGLMFormalArchive(t *testing.T, archive, hash string) string {
	t.Helper()
	raw, e := os.ReadFile("../../" + archive)
	if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != hash {
		t.Fatal("immutable archive unavailable")
	}
	dir := t.TempDir()
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.Contains(f.Name, "..") || filepath.IsAbs(f.Name) || f.UncompressedSize64 > 16<<20 {
			t.Fatal("invalid capture path")
		}
		p := filepath.Join(dir, f.Name)
		os.MkdirAll(filepath.Dir(p), 0755)
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(io.LimitReader(r, 16<<20+1))
		r.Close()
		if e != nil || len(b) > 16<<20 {
			t.Fatal("capture read")
		}
		os.WriteFile(p, b, 0600)
	}
	return dir
}

// The literals were independently summed from the immutable captured Attempt
// usage before implementation: input+cache-read+cache-write once; output once.
// Two cancellations are unknown on Chat/Responses, all Messages attempts partial.
func TestGLMFormalKnownUsage(t *testing.T) {
	var index struct {
		Routes map[string]struct{ Archive, ArchiveSHA256, ReportPath string }
	}
	raw, e := os.ReadFile("../live/testdata/glm-formal/live-captures.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("formal captures unavailable")
	}
	for route, want := range map[string][2]int{"chat": {1725, 8312}, "responses": {1648, 7839}, "messages": {2660, 3879}} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GLM12-reported-known-usage-"+route)
			entry := index.Routes[route]
			dir := extractGLMFormalArchive(t, entry.Archive, entry.ArchiveSHA256)
			r, e := supportmatrix.NewGLMReport(filepath.Dir(filepath.Join(dir, entry.ReportPath)))
			ev.Check("captured formal source accepted", e == nil, "%v", e)
			ev.Check("reported known totals retained", r.Budget.InputTokensUsed == want[0] && r.Budget.OutputTokensUsed == want[1], "got input=%d output=%d", r.Budget.InputTokensUsed, r.Budget.OutputTokensUsed)
			ev.Record("known-lower-bound", map[string]any{"input": want[0], "output": want[1], "note": "reported-known lower bounds; partial/cancelled/missing usage remains unknown; reasoning already within output; not billing"})
		})
	}
}
