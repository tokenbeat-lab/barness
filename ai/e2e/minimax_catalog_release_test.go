package e2e

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Failure inventory: the metadata control must remain present without claiming
// a differential; the actual catalog comparison must retain pinned pi versions,
// real Go provenance and its complete zero-pending record. Missing/forged Go
// provenance is not repaired by the importer. Original historical runs stay as
// recorded. This subprocess uses frozen model lookup only, no vendor or key.
func TestMiniMaxCatalogReleaseAdmission(t *testing.T) {
	ev := run.Case(t, "MM06-catalog-release-admission")
	base := t.TempDir()
	args := []string{"test", "-json", "./ai/e2e", "-count=1", "-run", "^TestMiniMaxPi(Catalog|CatalogDifferences)$"}
	if os.Getenv("BARNESS_MINIMAX_CATALOG_CHILD_RACE") == "1" {
		args = append(args, "-race")
		ev.ReplayEnv("BARNESS_MINIMAX_CATALOG_CHILD_RACE=1")
	}
	cmd := exec.CommandContext(ctxFor(t), "go", args...)
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "BARNESS_AI_PIDIFF=1", "BARNESS_AI_EVIDENCE_DIR="+base)
	output, err := cmd.CombinedOutput()
	ev.Record("producer-process", map[string]any{"argv": args, "failed": err != nil, "outputSHA256": pioracle.SHA256(output), "vendorCalls": 0, "lookupOnly": true})
	if !ev.Check("frozen catalog producers execute", err == nil, "child failed=%t", err != nil) {
		return
	}
	dir, err := release.FindBundle(base)
	if err != nil {
		t.Fatal("controlled producer bundle unavailable")
	}
	raw := miniMaxCatalogArtifact(t, filepath.Join(dir, "manifest.json"))
	ev.Fixture("producer-manifest.json", raw)
	var manifest struct {
		Cases []struct {
			ID, Dir   string
			Artifacts map[string]string
		}
	}
	if err := json.Unmarshal(raw, &manifest); err != nil || len(manifest.Cases) != 2 {
		t.Fatal("controlled catalog case inventory invalid")
	}
	for i, c := range manifest.Cases {
		if !filepath.IsLocal(c.Dir) {
			t.Fatal("controlled case path is not local")
		}
		for name := range c.Artifacts {
			if filepath.Base(name) != name {
				t.Fatal("controlled artifact is not flat")
			}
			data := miniMaxCatalogArtifact(t, filepath.Join(dir, c.Dir, name))
			ev.Fixture([]string{"first-", "second-"}[i]+name, data)
		}
	}
	b, err := release.LoadBundle(dir)
	ev.Record("import", map[string]any{"error": err != nil, "cases": b.Cases, "diffs": b.Diffs, "problems": b.Problems})
	if !ev.Check("complete controlled producer bundle imports", err == nil && len(b.Problems) == 0, "error=%t problems=%d", err != nil, len(b.Problems)) {
		return
	}
	ev.Check("metadata control retained in its own namespace", len(b.Cases) == 2 && slices.Contains(b.Cases, release.Case{ID: "MM06-frozen-catalog-metadata", Status: release.Pass}), "metadata control missing or misclassified")
	ev.Check("one genuine catalog differential admitted", len(b.Diffs) == 1 && b.Diffs[0].ID == "PIDIFF-MM06-catalog-differences" && len(b.Diffs[0].Missing) == 0 && b.Diffs[0].Pending == 0, "diff=%+v", b.Diffs)
	var record pioracle.Record
	raw = miniMaxCatalogArtifact(t, filepath.Join(dir, "PIDIFF-MM06-catalog-differences", "pidiff.json"))
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal("controlled catalog record invalid")
	}
	ev.Check("actual Go provenance retained", record.SDKVersions["go"] == runtime.Version(), "Go provenance absent or invented")
	ev.Check("frozen pi provenance retained", record.PiVersion == "1.1.0" && record.SDKVersions["@earendil-works/pi-ai"] == "1.1.0" && record.PiModelDataSHA256 == "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "frozen baseline changed")
	emptyRequest := []byte(`{"method":"","path":"","query":"","auth":"","headers":null,"body":null}`)
	ev.Check("metadata comparison never claims HTTP requests", record.RequestSHA256["pi"] == pioracle.SHA256(emptyRequest) && record.RequestSHA256["barness"] == pioracle.SHA256(emptyRequest), "unexpected fabricated wire")
	for _, mode := range []string{"missing-go", "wrong-pi"} {
		var changed, envelope map[string]any
		if json.Unmarshal(raw, &changed) != nil {
			t.Fatal("controlled record decode")
		}
		sdks := changed["sdk_versions"].(map[string]any)
		if mode == "missing-go" {
			delete(sdks, "go")
		} else {
			sdks["@earendil-works/pi-ai"] = "0.0.0"
		}
		data := mustMarshal(t, changed)
		path := filepath.Join(dir, "PIDIFF-MM06-catalog-differences", "pidiff.json")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal("controlled tamper write")
		}
		if json.Unmarshal(miniMaxCatalogArtifact(t, filepath.Join(dir, "manifest.json")), &envelope) != nil {
			t.Fatal("controlled manifest decode")
		}
		for _, entry := range envelope["cases"].([]any) {
			c := entry.(map[string]any)
			if c["id"] == "PIDIFF-MM06-catalog-differences" {
				c["artifacts"].(map[string]any)["pidiff.json"] = pioracle.SHA256(data)
			}
		}
		writeReleaseJSON(t, dir, "manifest.json", envelope)
		ev.Fixture(mode+"-pidiff.json", data)
		ev.Fixture(mode+"-manifest.json", miniMaxCatalogArtifact(t, filepath.Join(dir, "manifest.json")))
		invalid, e := release.LoadBundle(dir)
		ev.Record(mode+"-import", map[string]any{"error": e != nil, "diffs": invalid.Diffs, "problems": invalid.Problems})
		ev.Check(mode+" rejected by strict SDK semantics", e == nil && len(invalid.Problems) == 0 && len(invalid.Diffs) == 1 && slices.Contains(invalid.Diffs[0].Missing, "sdk_versions mismatch"), "invalid SDK provenance admitted")
	}
}

func miniMaxCatalogArtifact(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal("controlled catalog artifact unavailable")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (2<<20)+1))
	if err != nil || len(raw) > 2<<20 {
		t.Fatal("controlled catalog artifact exceeds finite bound")
	}
	return raw
}
