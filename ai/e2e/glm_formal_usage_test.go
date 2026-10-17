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

// A coherent editable inventory cannot hide an additional HTTP call from the
// fixed route counters. The fixtures below clone already captured bytes only.
func TestGLMFormalExtraCaptures(t *testing.T) {
	var index struct {
		Routes map[string]struct{ Archive, ArchiveSHA256, ReportPath string }
	}
	raw, e := os.ReadFile("../live/testdata/glm-formal/live-captures.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("formal captures unavailable")
	}
	entry := index.Routes["chat"]
	for _, name := range []string{"extra-manifest-case", "hidden-second-attempt", "unmanifest-capture"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "GLM12-extra-capture-"+name)
			dir := extractGLMFormalArchive(t, entry.Archive, entry.ArchiveSHA256)
			base := filepath.Dir(filepath.Join(dir, entry.ReportPath))
			data, _ := os.ReadFile(filepath.Join(base, "manifest.json"))
			var manifest map[string]any
			json.Unmarshal(data, &manifest)
			first := "GLM11-glm-5.3-full-stream"
			if name == "extra-manifest-case" {
				target := "GLM11-hidden-http"
				if e := os.CopyFS(filepath.Join(base, target), os.DirFS(filepath.Join(base, first))); e != nil {
					t.Fatal(e)
				}
				var clone map[string]any
				for _, c := range manifest["cases"].([]any) {
					if c.(map[string]any)["id"] == first {
						encoded, _ := json.Marshal(c)
						json.Unmarshal(encoded, &clone)
					}
				}
				clone["id"], clone["dir"] = target, target
				manifest["cases"] = append(manifest["cases"].([]any), clone)
				p := filepath.Join(base, target, "assertions.json")
				b, _ := os.ReadFile(p)
				var assertions map[string]any
				json.Unmarshal(b, &assertions)
				assertions["case_id"] = target
				b, _ = json.Marshal(assertions)
				os.WriteFile(p, b, 0600)
			} else {
				target := filepath.Join(base, first, "attempt-2-exchanges.json")
				if name == "unmanifest-capture" {
					target = filepath.Join(base, "unmanifest-http", "attempt-1-exchanges.json")
					os.MkdirAll(filepath.Dir(target), 0755)
				}
				b, _ := os.ReadFile(filepath.Join(base, first, "attempt-1-exchanges.json"))
				os.WriteFile(target, b, 0600)
			}
			for _, value := range manifest["cases"].([]any) {
				c := value.(map[string]any)
				files, _ := os.ReadDir(filepath.Join(base, c["dir"].(string)))
				hashes := map[string]string{}
				for _, f := range files {
					b, _ := os.ReadFile(filepath.Join(base, c["dir"].(string), f.Name()))
					hashes[f.Name()] = fmt.Sprintf("%x", sha256.Sum256(b))
				}
				c["artifacts"] = hashes
			}
			data, _ = json.Marshal(manifest)
			os.WriteFile(filepath.Join(base, "manifest.json"), data, 0600)
			inventory := map[string]string{}
			filepath.WalkDir(base, func(p string, d os.DirEntry, e error) error {
				if e != nil || d.IsDir() {
					return e
				}
				rel, _ := filepath.Rel(base, p)
				if rel != "glm-integrity.json" && rel != "live-report.json" {
					b, _ := os.ReadFile(p)
					inventory[filepath.ToSlash(rel)] = fmt.Sprintf("%x", sha256.Sum256(b))
				}
				return nil
			})
			data, _ = json.Marshal(map[string]any{"schema": 1, "files": inventory})
			os.WriteFile(filepath.Join(base, "glm-integrity.json"), data, 0600)
			_, err := supportmatrix.NewGLMReport(base)
			ev.Check("finite owned call/capture contract rejects "+name, err != nil, "coherently rehashed hidden capture accepted")
			ev.Record("scope", "Synthetic copy of actual immutable bytes; extra capture is never a real vendor call")
		})
	}
}
