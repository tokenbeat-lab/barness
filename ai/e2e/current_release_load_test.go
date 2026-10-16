package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestCurrentReleaseLoadImport(t *testing.T) {
	base := t.TempDir()
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "./ai/e2e", "-run", "^TestMultimodalRuntimeDesignLoad$", "-count=1")
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "BARNESS_AI_EVIDENCE_DIR="+base)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("public mixed load producer: %s", output)
	}
	source, err := release.FindBundle(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"intact", "missing-hooked", "budget", "resources", "assertions", "result", "observer", "unlisted"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "UC14-load-import-"+mode)
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS(source)); err != nil {
				t.Fatal(err)
			}
			raw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
			var manifest map[string]any
			mustUnmarshal(t, raw, &manifest)
			for i, v := range manifest["cases"].([]any) {
				c := v.(map[string]any)
				id := c["id"].(string)
				if id == "UC09-design-load-hooked" && mode == "missing-hooked" {
					manifest["cases"] = append(manifest["cases"].([]any)[:i], manifest["cases"].([]any)[i+1:]...)
					break
				}
				if id != "UC09-design-load-plain" {
					continue
				}
				file := map[string]string{"budget": "load.json", "resources": "resources.json", "assertions": "assertions.json", "result": "req-load-tenant-aclassifierplain-result.json", "observer": "observations-req-load-tenant-aclassifierplain.json", "unlisted": "load.json"}[mode]
				if file == "" {
					continue
				}
				artifacts := c["artifacts"].(map[string]any)
				if mode == "observer" {
					os.Remove(filepath.Join(dir, id, file))
					delete(artifacts, file)
				} else if mode == "unlisted" {
					delete(artifacts, file)
				} else {
					raw, _ := os.ReadFile(filepath.Join(dir, id, file))
					var value map[string]any
					mustUnmarshal(t, raw, &value)
					switch mode {
					case "budget":
						value["memoryCeiling"] = 1 << 30
					case "resources":
						delete(value, "bodies")
					case "assertions":
						value["assertions"] = []any{}
					case "result":
						value["imageInputs"] = []any{}
					}
					writeReleaseJSON(t, filepath.Join(dir, id), file, value)
					raw, _ = os.ReadFile(filepath.Join(dir, id, file))
					artifacts[file] = pioracle.SHA256(raw)
				}
			}
			writeReleaseJSON(t, dir, "manifest.json", manifest)
			report := releaseProcess(t, ev, "-bundle", dir, "-race-bundle", dir)
			name := "multimodal-images-plain"
			if mode == "missing-hooked" {
				name = "multimodal-images-hooked"
			}
			for _, prefix := range []string{"", "race-"} {
				got := releaseEvidence(t, report, prefix+name)
				ev.Check("same contract normal and race", (got.Status == release.Pass) == (mode == "intact"), "mode=%s status=%s detail=%s", mode, got.Status, got.Detail)
			}
			// Retain the producer's original manifest and image/HTTP/result contracts.
			raw, _ = os.ReadFile(filepath.Join(source, "manifest.json"))
			ev.Fixture("producer-manifest.json", raw)
			ev.Record("producer", map[string]any{"controlled": true, "cases": 2, "realService": "NOT_RUN", "panic": strings.Contains(string(output), "panic:")})
			raw, _ = json.Marshal(manifest)
			ev.Fixture("import-manifest.json", raw)
		})
	}
}
