package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog"
)

func TestLocalCatalogFileFailures(t *testing.T) {
	valid := string(localCatalogFile(`{"version":"file-01","models":[` + localChatJSON + `]}`))
	failures := []struct {
		name, raw string
		reason    catalog.FileReason
	}{
		{"invalid-json", `{"hash":`, catalog.FileInvalidJSON},
		{"trailing-json", valid + `{}`, catalog.FileInvalidJSON},
		{"root-array", `[]`, catalog.FileInvalidStructure},
		{"root-null", `null`, catalog.FileInvalidStructure},
		{"unknown-root", strings.Replace(valid, `"hash":`, `"extra":true,"hash":`, 1), catalog.FileInvalidStructure},
		{"wrong-hash-shape", strings.Replace(valid, `sha256:`, `sha1:`, 1), catalog.FileInvalidHash},
		{"bad-hash", strings.Replace(valid, `file-01`, `file-02`, 1), catalog.FileHashMismatch},
		{"unknown-catalog-field", string(localCatalogFile(`{"version":"file-01","models":[],"future":true}`)), catalog.FileInvalidStructure},
		{"null-models", string(localCatalogFile(`{"version":"file-01","models":null}`)), catalog.FileInvalidStructure},
		{"missing-models", string(localCatalogFile(`{"version":"file-01"}`)), catalog.FileInvalidStructure},
		{"object-models", string(localCatalogFile(`{"version":"file-01","models":{}}`)), catalog.FileInvalidStructure},
		{"null-images", string(localCatalogFile(`{"version":"file-01","models":[],"imageModels":null}`)), catalog.FileInvalidStructure},
		{"string-classifiers", string(localCatalogFile(`{"version":"file-01","models":[],"classifierModels":"bad"}`)), catalog.FileInvalidStructure},
		{"empty-version", string(localCatalogFile(`{"version":"","models":[]}`)), catalog.FileInvalidVersion},
		{"whitespace-version", string(localCatalogFile(`{"version":" bad ","models":[]}`)), catalog.FileInvalidVersion},
		{"duplicate-top-field", string(localCatalogFile(`{"version":"one","version":"two","models":[]}`)), catalog.FileInvalidStructure},
	}
	for _, failure := range failures {
		t.Run(failure.name, func(t *testing.T) {
			ev := run.Case(t, "CAT-T09-file-"+failure.name)
			effective, report, err := catalog.Load([]byte(failure.raw))
			var fileError *catalog.FileError
			ev.Record("result", map[string]any{"catalog": effective, "report": report, "error": errString(err)})
			ev.Check("whole file rejected with file-level reason", errors.As(err, &fileError) && fileError.Reason == failure.reason, "err=%v", err)
			ev.Check("no usable partial result or fallback", effective.Version == "" && len(effective.Models) == 0 && len(effective.ImageModels) == 0 && len(effective.ClassifierModels) == 0 && len(report.Models) == 0, "catalog=%+v report=%+v", effective, report)
		})
	}
	t.Run("read-failure", func(t *testing.T) {
		ev := run.Case(t, "CAT-T09-local-file-read-failure")
		_, _, err := catalog.LoadFile(filepath.Join(t.TempDir(), "missing.json"))
		var fileError *catalog.FileError
		ev.Check("missing explicit file does not fall back", errors.As(err, &fileError) && fileError.Reason == catalog.FileReadFailed, "err=%v", err)
	})
}

func TestLocalCatalogOriginalDigest(t *testing.T) {
	ev := run.Case(t, "CAT-T10-original-digest-before-model-parse")
	body := `{"version":"digest-01","models":[` + localChatJSON + `,{"provider":"openai","api":"openai-responses","id":"future","constraint":{"number":1e2,"label":"\u0061"}}]}`
	compact := localCatalogFile(body)
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, compact, "", "  "); err != nil {
		t.Fatal(err)
	}
	first, report, err := catalog.Load(compact)
	second, spacedReport, spacedErr := catalog.Load(pretty.Bytes())
	ev.Record("original-compact-file", json.RawMessage(compact))
	ev.Record("report", report)
	ev.Check("outside whitespace alone does not change either identity", err == nil && spacedErr == nil && first.Version == second.Version && report.SourceHash == spacedReport.SourceHash && report.EffectiveHash == spacedReport.EffectiveHash && report.Loaded == 1 && report.Failed == 1, "err=%v spaced=%v report=%+v", err, spacedErr, report)
	changes := []struct{ name, from, to string }{
		{"unknown-field-number", "1e2", "100"},
		{"unknown-field-escape", `\u0061`, `a`},
		{"unknown-field-value", `"constraint"`, `"newConstraint"`},
	}
	for _, change := range changes {
		mutated := strings.Replace(string(compact), change.from, change.to, 1)
		_, _, err := catalog.Load([]byte(mutated))
		var fileError *catalog.FileError
		ev.Check("original spelling covered: "+change.name, errors.As(err, &fileError) && fileError.Reason == catalog.FileHashMismatch, "err=%v", err)
	}
	path := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(path, compact, 0600); err != nil {
		t.Fatal(err)
	}
	fromFile, fileReport, fileErr := catalog.LoadFile(path)
	if err := os.WriteFile(path, []byte(`broken`), 0600); err != nil {
		t.Fatal(err)
	}
	ev.Check("explicit local file loads the selected immutable bytes", fileErr == nil && fromFile.Version == "digest-01" && fileReport.SourceHash == report.SourceHash, "err=%v", fileErr)
}
