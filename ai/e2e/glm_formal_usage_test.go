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

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Locate independently pinned historical bytes without relying on the original
// local capture directory. The native consumer validates this same input on its
// own; a helper success never certifies a current support row.
func extractGLMFormalArchive(t *testing.T, hash string) string {
	t.Helper()
	raw, e := os.ReadFile(filepath.Join("../live/testdata/glm-formal/archives", hash+".zip"))
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

// Expected totals are independent literal data beside the pinned raw archive
// identity, reconciled from captured Attempts before this consumer is run.
// Keeping measured facts in data lets a new capture renew them without a Go
// source edit that would invalidate already frozen Jev evidence.
// Two cancellations are unknown on Chat/Responses, all Messages attempts partial.
func TestGLMFormalKnownUsage(t *testing.T) {
	var index struct {
		OriginalLiveSourceCommit string
		Routes                   map[string]struct{ Archive, ArchiveSHA256, ReportPath, ReportSHA256 string }
	}
	raw, e := os.ReadFile("../live/testdata/glm-formal/live-captures.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("formal captures unavailable")
	}
	var expected struct {
		Schema       int
		SourceCommit string
		Routes       map[string]struct {
			ArchiveSHA256, ReportSHA256                             string
			InputSide, Output, UnknownUsageAttempts, ActualAttempts int
		}
	}
	expectationBytes, readErr := os.ReadFile("../live/testdata/glm-formal/known-usage.json")
	if readErr != nil || json.Unmarshal(expectationBytes, &expected) != nil || expected.Schema != 1 {
		t.Fatal("independent measured expectations unavailable")
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GLM12-reported-known-usage-"+route)
			entry := index.Routes[route]
			literal := expected.Routes[route]
			want := [2]int{literal.InputSide, literal.Output}
			ev.Check("expectations bound to exact immutable source/report/archive", expected.SourceCommit == index.OriginalLiveSourceCommit && literal.ArchiveSHA256 == entry.ArchiveSHA256 && literal.ReportSHA256 == entry.ReportSHA256, "expectation identity drift")
			ev.Check("independently reconciled measured totals present", want[0] > 0 && want[1] > 0, "missing measured expectations")
			dir := extractGLMFormalArchive(t, entry.ArchiveSHA256)
			base := filepath.Dir(filepath.Join(dir, entry.ReportPath))
			_, e := supportmatrix.NewGLMReport(base)
			// Read immutable original facts without certifying them as current.
			// The archive and raw report remain bound to independent source pins.
			formalBytes, readErr := os.ReadFile(filepath.Join(dir, entry.ReportPath))
			ev.Check("historical formal report retains independent hash", readErr == nil && fmt.Sprintf("%x", sha256.Sum256(formalBytes)) == entry.ReportSHA256, "original report changed")
			r, loadErr := supportmatrix.LoadReport(filepath.Join(base, "live-report.json"))
			if !ev.Check("immutable historical support projection readable", loadErr == nil, "%v", loadErr) {
				return
			}
			ev.Check("historical actual capture refuses current catalog/source scope", glmCurrentScopeRefusal(e, r.CatalogHash), "%v", e)
			input, output, attempts, unknown := glmHistoricalAttemptUsage(t, base, r)
			ev.Check("raw Attempts independently preserve known totals and unknowns", input == want[0] && output == want[1] && attempts == literal.ActualAttempts && unknown == literal.UnknownUsageAttempts, "got input=%d output=%d attempts=%d unknown=%d", input, output, attempts, unknown)
			ev.Check("historical projection retains same known lower bounds", r.Budget.InputTokensUsed == want[0] && r.Budget.OutputTokensUsed == want[1], "got input=%d output=%d", r.Budget.InputTokensUsed, r.Budget.OutputTokensUsed)
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
			dir := extractGLMFormalArchive(t, entry.ArchiveSHA256)
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

// Reconcile original per-call Results against the pre-existing independent
// Python totals; never call the current support producer to refresh old facts.
func glmHistoricalAttemptUsage(t *testing.T, base string, r supportmatrix.Report) (input, output, attempts, unknown int) {
	t.Helper()
	for _, scenario := range r.Scenarios {
		if scenario.Attempts == 0 {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(base, scenario.CaseID, "attempt-1-result.json"))
		if err != nil {
			t.Fatal("historical Attempt original unavailable")
		}
		var captured struct {
			Result struct{ Metadata ai.CallMetadata }
		}
		if json.Unmarshal(raw, &captured) != nil {
			t.Fatal("historical Attempt original invalid")
		}
		for _, a := range captured.Result.Metadata.Attempts {
			attempts++
			input += int(a.Usage.Input + a.Usage.CacheRead + a.Usage.CacheWrite)
			output += int(a.Usage.Output)
			if a.UsageReporting != ai.UsageComplete {
				unknown++
			}
		}
	}
	return
}
