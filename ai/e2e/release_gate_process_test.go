package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// OD57 exercises the actual release CLI with retained, deliberately incomplete
// bundles. Inspect the individual gate: a different expected failure must not
// conceal a missing contract check.
func releaseProcess(t *testing.T, ev *evidence.Case, args ...string) release.Report {
	t.Helper()
	offline, out := t.TempDir(), t.TempDir()
	writeReleaseJSON(t, offline, "manifest.json", map[string]any{"package": "./ai/e2e", "git_commit": "synthetic-release-test", "cases": []any{}})
	writeReleaseJSON(t, offline, "audit.json", map[string]any{"findings": []any{}})
	argv := append([]string{"run", "./ai/release/cmd/releasegate", "-bundle", offline, "-out", out}, args...)
	cmd := exec.CommandContext(ctxFor(t), "go", argv...)
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	output, err := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"rejected": err != nil, "panic": strings.Contains(string(output), "panic:")})
	raw, readErr := os.ReadFile(filepath.Join(out, "release-report.json"))
	if !ev.Check("CLI emits a reviewable verdict", readErr == nil, "report missing; exit=%v", err) {
		t.FailNow()
	}
	var report release.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	// Temp paths identify this process only, so keep outcomes in the durable
	// parent artifact without copying host filesystem paths into its audit.
	ev.Record("verdict", map[string]any{"pass": report.Pass, "live": report.Live, "evidence": sanitizedReleaseEvidence(report.Evidence)})
	ev.Check("incomplete offline bundle refuses release without panic", err != nil && !report.Pass && !strings.Contains(string(output), "panic:"), "bad process outcome")
	return report
}
func sanitizedReleaseEvidence(in []release.EvidenceCheck) []release.EvidenceCheck {
	out := slices.Clone(in)
	for i := range out {
		out[i].Bundle = "[test-bundle]"
	}
	return out
}
func writeReleaseJSON(t *testing.T, dir, name string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func releaseEvidence(t *testing.T, r release.Report, name string) release.EvidenceCheck {
	t.Helper()
	for _, e := range r.Evidence {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("required evidence %s absent", name)
	return release.EvidenceCheck{}
}
func TestTenRouteReleaseCLI(t *testing.T) {
	ev := run.Case(t, "OD57-release-ten-route-contract")
	r := releaseProcess(t, ev)
	ev.Check("exactly ten own-route live requirements", len(r.Live.Rows) == 10, "got %d", len(r.Live.Rows))
	for _, name := range []string{"chinese-effect-jev", "chinese-effect-luna", "authorization-revocation"} {
		e := releaseEvidence(t, r, name)
		ev.Check("missing "+name+" refuses delivery", e.Status != release.Pass, "missing evidence passed")
	}
	trace, err := release.LoadTraceMap("../release/traceability.json")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 20; i++ {
		id := fmt.Sprintf("OD-R%02d", i)
		ev.Check("required trace "+id, slices.ContainsFunc(trace.Items, func(it release.TraceItem) bool { return it.ID == id }), "missing %s", id)
	}
}
func TestReleaseEvaluationIdentityAndIntegrity(t *testing.T) {
	// Generate current execution evidence through the public CLI. A retained
	// live bundle becomes stale after legitimate source changes; using it here
	// would make every tampering case pass at the unrelated source-drift guard.
	// Fixtures remain NOT_RUN for live delivery, including this positive control.
	source := currentReleaseEvaluationFixture(t, run.Case(t, "OD57-release-evaluation-source"))
	for _, mode := range []string{"current", "same-package", "wrong-route", "catalog", "budget", "code", "dataset", "config", "manifest", "audit", "files", "captures", "observer", "fixture", "not-run", "failed"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-evaluation-"+mode)
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS(source)); err != nil {
				t.Fatal(err)
			}
			args := []string{"-evaluation-luna", dir}
			if mode == "same-package" {
				args = append(args, "-evaluation-jev", dir)
			}
			if mode == "wrong-route" {
				args = []string{"-evaluation-jev", dir}
			}
			deleted := map[string]string{"manifest": "manifest.json", "audit": "audit.json", "files": "files.json", "captures": "captures.json", "observer": "observations-evaluation.json"}
			if file := deleted[mode]; file != "" {
				if err := os.Remove(filepath.Join(dir, file)); err != nil {
					t.Fatal(err)
				}
			}
			if slices.Contains([]string{"catalog", "budget", "code", "dataset", "config", "fixture", "not-run", "failed"}, mode) {
				raw, err := os.ReadFile(filepath.Join(dir, "evaluation-report.json"))
				if err != nil {
					t.Fatal(err)
				}
				var r map[string]any
				if err = json.Unmarshal(raw, &r); err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "catalog":
					r["config"].(map[string]any)["catalogHash"] = "sha256:old"
				case "budget":
					r["budget"].(map[string]any)["calls"] = 25
				case "code":
					r["codeHash"] = "sha256:" + strings.Repeat("0", 64)
				case "dataset":
					r["datasetHash"] = "sha256:" + strings.Repeat("0", 64)
				case "config":
					r["configHash"] = "sha256:" + strings.Repeat("0", 64)
				case "fixture":
					r["kind"] = "live" // a fixture cannot be promoted to live evidence
				case "not-run":
					r["status"] = "NOT_RUN"
				case "failed":
					r["status"] = "FAIL"
				}
				writeReleaseJSON(t, dir, "evaluation-report.json", r)
				rehashReleaseFile(t, dir, "evaluation-report.json")
			}
			r := releaseProcess(t, ev, args...)
			name := "chinese-effect-luna"
			if mode == "wrong-route" {
				name = "chinese-effect-jev"
			}
			e := releaseEvidence(t, r, name)
			if mode == "current" {
				ev.Check("current fixture is verified without claiming live", e.Status == release.NotRun && strings.Contains(e.Detail, "fixture/COMPLETE"), "got %+v", sanitizedReleaseEvidence([]release.EvidenceCheck{e}))
			} else {
				ev.Check("tampered evidence fails its integrity checks", e.Status == release.Fail, "%s: got %s, %s", mode, e.Status, e.Detail)
				if mode != "code" {
					ev.Check("source drift does not mask the intended integrity check", !strings.Contains(e.Detail, "execution source is stale"), "%s: %s", mode, e.Detail)
				}
			}
		})
	}
}

func currentReleaseEvaluationFixture(t *testing.T, ev *evidence.Case) string {
	t.Helper()
	out := t.TempDir()
	if _, err := evaluationCommand(t, "-route", "luna", "-fixture", "normal", "-out", out); err != nil {
		t.Fatal(err)
	}
	dir, _ := evaluationBundle(t, out, ev)
	return dir
}
func rehashReleaseFile(t *testing.T, dir, name string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "files.json"))
	if err != nil {
		t.Fatal(err)
	}
	var files map[string]string
	if err = json.Unmarshal(raw, &files); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	files[name] = chineseeval.Hash(raw)
	writeReleaseJSON(t, dir, "files.json", files)
}
