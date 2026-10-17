//go:build live

package live

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// The route report is the acceptance marker. It stays incomplete until all
// captures, audit, immutable file reads, model summaries and inventory succeed.
// Staged report wrappers cannot be decoded as accepting candidate reports.
func finishGLM(code int, auditOK bool) int {
	r := glmReportState
	r.AuditPassed = auditOK
	r.Complete = false
	record := func() error {
		if e := run.Record(glmReportName, r); e != nil {
			return e
		}
		for name, v := range glmModelReports() {
			if e := run.Record(name, v); e != nil {
				return e
			}
		}
		return nil
	}
	fail := func(note string) int {
		r.Outcome = "FAIL"
		r.Complete = false
		r.Failure = note
		_ = record()
		return 1
	}
	if !auditOK {
		return fail("redaction audit failed")
	}
	if code != 0 {
		r.Outcome = "FAIL"
	}
	if e := record(); e != nil {
		return fail("provisional report write failed")
	}
	files := map[string]string{}
	err := filepath.WalkDir(run.Dir(), func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "glm-integrity.json" || strings.HasPrefix(d.Name(), "glm-pending-") {
			return nil
		}
		raw, e := readDecisionsFile(path, 16<<20)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(run.Dir(), path)
		if e != nil {
			return e
		}
		files[filepath.ToSlash(rel)] = sha256Text(raw)
		return nil
	})
	if err != nil || os.Getenv("BARNESS_GLM_CONTROLLED") == "inventory-failure" {
		return fail("inventory read failed")
	}
	r.Complete = r.Outcome == "PASS" && code == 0
	final := glmModelReports()
	final[glmReportName] = r
	pending := map[string]string{}
	for name, v := range final {
		temp := "glm-pending-" + name
		if e := run.Record(temp, map[string]any{"sealedCandidate": v}); e != nil {
			return fail("pending report write failed")
		}
		pending[name] = filepath.Join(run.Dir(), temp+".json")
	}
	findings, e := run.Audit()
	if e != nil || len(findings) > 0 {
		r.AuditPassed = false
		return fail("final redaction audit failed")
	}
	// Read the redacted wrapper, encode the exact final artifact bytes and stage
	// every write before any candidate acceptance marker is published.
	staged := map[string]string{}
	defer func() {
		for _, p := range pending {
			_ = os.Remove(p)
		}
		for _, p := range staged {
			_ = os.Remove(p)
		}
	}()
	for name, path := range pending {
		raw, e := readDecisionsFile(path, 2<<20)
		if e != nil {
			return fail("pending report read failed")
		}
		var wrapper struct {
			Candidate json.RawMessage `json:"sealedCandidate"`
		}
		if e = json.Unmarshal(raw, &wrapper); e != nil {
			return fail("pending report decoding failed")
		}
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if e = encoder.Encode(wrapper.Candidate); e != nil {
			return fail("report encoding failed")
		}
		temp := filepath.Join(run.Dir(), "glm-pending-final-"+name+".json")
		if e = os.WriteFile(temp, buf.Bytes(), 0600); e != nil {
			return fail("report staging failed")
		}
		staged[name] = temp
		files[name+".json"] = sha256Text(buf.Bytes())
	}
	raw, e := readDecisionsFile(filepath.Join(run.Dir(), "audit.json"), 1<<20)
	if e != nil {
		return fail("final audit read failed")
	}
	files["audit.json"] = sha256Text(raw)
	if e = run.Record("glm-integrity", map[string]any{"schema": 1, "files": files}); e != nil {
		return fail("inventory write failed")
	}
	for name, path := range staged {
		if name == glmReportName {
			continue
		}
		if e = os.Rename(path, filepath.Join(run.Dir(), name+".json")); e != nil {
			return fail("model report publication failed")
		}
	}
	if e = os.Rename(staged[glmReportName], filepath.Join(run.Dir(), glmReportName+".json")); e != nil {
		return fail("route report publication failed")
	}
	return code
}
