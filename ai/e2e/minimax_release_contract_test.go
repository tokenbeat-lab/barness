package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// These are actual public Client load captures, imported through the same
// bundle and consumer boundary the release command uses. Rehashed mutations
// exercise semantic checks independently of the artifact hash inventory.
func TestMiniMaxReleaseLoadConsumer(t *testing.T) { miniMaxReleaseLoadConsumer(t, false) }

func TestMiniMaxReleasePressureLoadConsumer(t *testing.T) {
	pressureCase(t, "MM10-actual-pressure-import")
	miniMaxReleaseLoadConsumer(t, true)
}

func miniMaxReleaseLoadConsumer(t *testing.T, actual bool) {
	ev := run.Case(t, fmt.Sprintf("MM10-release-load-consumer-%t", actual))
	base := t.TempDir()
	dir := base
	if actual {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, "go", "test", "-count=1", "-timeout=90s", "./ai/e2e", "-run", "^TestMiniMaxSharedDesignLoad$")
		command.Dir = "../.."
		for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
			if value, ok := os.LookupEnv(key); ok {
				command.Env = append(command.Env, key+"="+value)
			}
		}
		command.Env = append(command.Env, "BARNESS_AI_PRESSURE=1", "BARNESS_AI_EVIDENCE_DIR="+base)
		output, err := command.CombinedOutput()
		if !ev.Check("explicit pressure capture succeeds", err == nil, "%v %s", err, output) {
			return
		}
		dir, err = release.FindBundle(base)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		if _, err := release.LoadMiniMaxDesignLoadFixture(base); err != nil {
			t.Fatal(err)
		}
	}

	bundle, err := release.LoadBundle(dir)
	if err != nil {
		t.Fatal(err)
	}
	checks := release.MiniMaxDesignLoadEvidence(bundle)
	ev.Check("exactly six independent consumers", len(checks) == 6, "%d", len(checks))
	for _, check := range checks {
		ev.Check(check.Name, check.Status == release.Pass, "%s", check.Detail)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("source-manifest.json", raw)
	id := "MM05-design-load-local-chat"
	manifestPath := filepath.Join(dir, "manifest.json")
	reportPath := filepath.Join(dir, id, "design-load.json")
	original, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"baseline", "additional", "protocol", "routes", "memory", "headroom", "request", "result", "observer", "resources", "case-status"} {
		t.Run(mutation, func(t *testing.T) {
			var manifest map[string]any
			if json.Unmarshal(raw, &manifest) != nil {
				t.Fatal("manifest decode")
			}
			selected := map[string]any{}
			for _, v := range manifest["cases"].([]any) {
				c := v.(map[string]any)
				if c["id"] == id {
					selected = c
				}
			}
			var value map[string]any
			if json.Unmarshal(original, &value) != nil {
				t.Fatal("load decode")
			}
			leaf := "design-load.json"
			changed := any(value)
			switch mutation {
			case "baseline":
				value["baselineCallsUnchanged"] = 3
			case "additional":
				value["additionalCalls"] = 2
			case "protocol":
				value["MiniMaxProtocol"] = "responses"
			case "routes":
				value["originalRoutes"] = []string{"chat", "classifier"}
			case "memory":
				value["totalAllocDuringCalls"] = 999999999
			case "headroom":
				value["headroom"] = map[string]any{}
			case "request":
				leaf = "requests.json"
				changed = []any{}
			case "result":
				leaf = "mm05-load-localchattenant-a-result.json"
				changed = map[string]any{}
			case "observer":
				leaf = "observations-mm05-load-localchattenant-a.json"
				changed = []any{}
			case "resources":
				leaf = "resources.json"
				changed = map[string]int{"calls": 0, "permits": 0, "waiters": 0, "events": 0, "bodies": 1, "hostPermits": 0}
			case "case-status":
				selected["status"] = "FAIL"
			}
			path := filepath.Join(dir, id, leaf)
			saved, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			bytes, e := json.Marshal(changed)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(path, bytes, 0600); e != nil {
				t.Fatal(e)
			}
			defer os.WriteFile(path, saved, 0600)
			defer os.WriteFile(manifestPath, raw, 0600)
			selected["artifacts"].(map[string]any)[leaf] = fmt.Sprintf("%x", sha256.Sum256(bytes))
			encoded, e := json.Marshal(manifest)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(manifestPath, encoded, 0600); e != nil {
				t.Fatal(e)
			}
			imported, e := release.LoadBundle(dir)
			if e != nil {
				t.Fatal(e)
			}
			ev.Check("rehashed mutation imports without inventory errors "+mutation, len(imported.Problems) == 0, "%v", imported.Problems)
			verdicts := release.MiniMaxDesignLoadEvidence(imported)
			at := slices.IndexFunc(verdicts, func(c release.EvidenceCheck) bool { return c.Name == "minimax-design-load-local-chat" })
			ev.Check("semantic consumer rejects "+mutation, at >= 0 && verdicts[at].Status == release.Fail, "%+v", verdicts)
		})
	}
}
