package release

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

// Rejected input fields must not become diagnostic output (issue 17 review).
func TestRejectedIdentifiersAreNeverRepublished(t *testing.T) {
	const secret = "sk-syntheticDisclosureCredential"
	t.Run("live-combo", func(t *testing.T) {
		dir := t.TempDir()
		r := goodInputs(t).Audits[1].Live
		r.Combo = secret
		writeArtifact(t, dir, "live-report.json", r)
		writeArtifact(t, dir, "manifest.json", map[string]any{"cases": []any{}})
		raw, err := json.Marshal(AuditBundle(dir, nil))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte(secret)) {
			t.Fatal("rejected combo republished")
		}
	})
	t.Run("recorded-audit-rule", func(t *testing.T) {
		dir := t.TempDir()
		writeArtifact(t, dir, "manifest.json", map[string]any{"cases": []any{}})
		writeArtifact(t, dir, "audit.json", map[string]any{"findings": []audit.Finding{{File: "fixture.json", Rule: audit.Rule(secret), Detail: "invalid"}}})
		raw, err := json.Marshal(AuditBundle(dir, nil))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte(secret)) {
			t.Fatal("rejected audit rule republished")
		}
	})
	for _, field := range []string{"case-id", "artifact-name"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			id, name := "E01-synthetic", "result.json"
			if field == "case-id" {
				id += "-" + secret
			} else {
				name = secret + ".json"
			}
			writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]any{{"id": id, "dir": "synthetic", "status": "PASS", "artifacts": map[string]string{name: "missing"}}}})
			b, err := LoadBundle(dir)
			raw, _ := json.Marshal(b)
			if bytes.Contains(raw, []byte(secret)) || (err != nil && bytes.Contains([]byte(err.Error()), []byte(secret))) {
				t.Fatal("rejected offline identifier republished")
			}
			writeArtifact(t, dir, name, map[string]any{"value": secret})
			raw, _ = json.Marshal(AuditBundle(dir, []audit.Forbidden{{Value: secret, Label: "test credential"}}))
			if bytes.Contains(raw, []byte(secret)) {
				t.Fatal("rejected artifact filename republished by audit")
			}
		})
	}
}
