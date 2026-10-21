package minimaxlive_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Exercise actual live.test initialization with real canonical authority in an
// isolated synthetic Git owner. No Key or live switch is supplied to the child.
func TestMiniMaxRecovery7HarnessDeadlineProcess(t *testing.T) {
	t.Parallel()
	admitRenewalOwner(t)
	root, e := minimaxlive.Root()
	if e != nil {
		t.Fatal(e)
	}
	live := filepath.Join(t.TempDir(), "live.test")
	c := exec.Command("go", "test", "-tags", "live", "-c", "-o", live, "./ai/live")
	c.Dir = root
	if raw, e := c.CombinedOutput(); e != nil {
		t.Fatalf("build live: %v %s", e, raw)
	}
	dir := t.TempDir()
	for n, raw := range map[string]string{"go.mod": "module deadline-fixture\nrequire github.com/openai/openai-go/v3 v3.0.0\n", "go.sum": "", "ai/runtime.go": "package ai\n", "ai/internal/testkit/finalbudget/guard.go": "package finalbudget\n", "ai/internal/testkit/finalbudget/estimate.go": "package finalbudget\n"} {
		p := filepath.Join(dir, n)
		os.MkdirAll(filepath.Dir(p), 0700)
		os.WriteFile(p, []byte(raw), 0600)
	}
	sources, e := minimaxlive.Sources()
	if e != nil {
		t.Fatal(e)
	}
	for name, raw := range sources {
		p := filepath.Join(dir, "ai/live/testdata/minimax", name)
		os.MkdirAll(filepath.Dir(p), 0700)
		os.WriteFile(p, raw, 0600)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "go.sum", "ai"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "source"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if raw, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git %v %s", e, raw)
		}
	}
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	c = exec.CommandContext(ctx, exe, "-test.run=^TestMiniMaxRecovery7HarnessDeadlineChild$", "-test.v")
	c.Dir = dir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_MINIMAX_DEADLINE_CHILD=" + live, "BARNESS_MINIMAX_DEADLINE_EVIDENCE=" + os.Getenv("BARNESS_MINIMAX_DEADLINE_EVIDENCE")}
	raw, e := c.CombinedOutput()
	t.Log(string(raw))
	if e != nil {
		t.Fatalf("actual harness deadline process: %v", e)
	}
}

func TestMiniMaxRecovery7HarnessDeadlineChild(t *testing.T) {
	live := os.Getenv("BARNESS_MINIMAX_DEADLINE_CHILD")
	if live == "" {
		t.Skip("synthetic process only")
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	f.seedSecondCohort(t)
	f.seedThirdCohort(t)
	f.seedFourthCohort(t)
	f.seedFifthCohort(t)
	f.seedSixthCohort(t)
	facts, e := minimaxlive.Recovery7Facts()
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(facts)
	var fields map[string]any
	json.Unmarshal(raw, &fields)
	f.installRecovery7Grant(t, fields)
	for _, tc := range []struct {
		name, mode, stage, model, renewal string
		authorized                        bool
	}{
		{"trusted-formal", "live", "formal", minimaxlive.M3, recovery7ID, true},
		{"unapproved-id7", "live", "formal", minimaxlive.M3, recovery7ID, false},
		{"controlled", "controlled", "formal", minimaxlive.M3, recovery7ID, false},
		{"replay", "replay", "formal", minimaxlive.M3, recovery7ID, false},
		{"candidate", "live", "candidate", minimaxlive.M3, recovery7ID, false},
		{"research", "live", "candidate", minimaxlive.M3, recovery7ID, false},
		{"research-formal", "live", "formal", minimaxlive.M3, recovery7ID, false},
		{"m31", "live", "formal", minimaxlive.M31, recovery7ID, false},
		{"closed6", "live", "formal", minimaxlive.M3, recovery6ID, false},
		{"unrenewed", "live", "candidate", minimaxlive.M3, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			grant := filepath.Join(f.base, "grant.json")
			if tc.name == "unapproved-id7" {
				old := mustReadFixture(t, grant)
				os.Remove(grant)
				defer os.WriteFile(grant, old, 0600)
			}
			out := t.TempDir()
			entry, purpose := "^TestMiniMaxCandidate$", "standard"
			if tc.stage == "formal" {
				entry = "^TestMiniMaxFormal$"
			}
			if tc.model == minimaxlive.M31 {
				purpose = "mplan"
			}
			env := []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_EVIDENCE_DIR=" + out, "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=" + tc.model, "BARNESS_MINIMAX_PURPOSE=" + purpose, "BARNESS_MINIMAX_FORMAL_RENEWAL=" + tc.renewal}
			if tc.stage == "formal" {
				env = append(env, "BARNESS_MINIMAX_FORMAL=1")
			}
			if tc.mode == "controlled" {
				env = append(env, "BARNESS_MINIMAX_CONTROLLED=success-chat")
			}
			if tc.mode == "replay" {
				env = append(env, "BARNESS_MINIMAX_REPLAY=synthetic")
				entry = "^TestMiniMaxReplay$"
			}
			if tc.name == "research" || tc.name == "research-formal" {
				env = append(env, "BARNESS_MINIMAX_RESEARCH=1")
				entry = "^TestMiniMaxResearch$"
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			c := exec.CommandContext(ctx, live, "-test.run="+entry, "-test.timeout=30s")
			c.Env = env
			start := time.Now().UTC()
			log, err := c.CombinedOutput()
			paths, _ := filepath.Glob(filepath.Join(out, "*", minimaxlive.ReportFile))
			if len(paths) != 1 {
				t.Fatalf("own report: %v %s", err, log)
			}
			b := mustReadFixture(t, paths[0])
			manifest := mustReadFixture(t, filepath.Join(filepath.Dir(paths[0]), "manifest.json"))
			var m struct {
				Package string `json:"package"`
			}
			if json.Unmarshal(manifest, &m) != nil || m.Package != "-tags live -timeout 30s ./ai/live" {
				t.Errorf("manifest must record actual child timeout: %s", manifest)
			}
			var r minimaxlive.Report
			if e := json.Unmarshal(b, &r); e != nil {
				t.Fatal(e)
			}
			code := 0
			if err != nil {
				code = c.ProcessState.ExitCode()
			}
			if tc.renewal != "" && !tc.authorized && code != 1 {
				t.Errorf("unapproved or nonformal renewal must reject: exit=%d output=%s", code, log)
			}
			receipt := map[string]any{"argv": c.Args, "startedAt": start, "finishedAt": time.Now().UTC(), "exitCode": code, "report": json.RawMessage(b), "manifest": json.RawMessage(manifest), "output": string(log), "credentialsRead": false, "vendorCalls": 0, "syntheticOwner": f.root}
			if evidence := os.Getenv("BARNESS_MINIMAX_DEADLINE_EVIDENCE"); evidence != "" {
				os.MkdirAll(evidence, 0700)
				b, _ := json.MarshalIndent(receipt, "", "  ")
				if e := os.WriteFile(filepath.Join(evidence, tc.name+".json"), b, 0600); e != nil {
					t.Fatal(e)
				}
			}
			want := 1800 * time.Second
			if tc.authorized {
				want = 3600 * time.Second
				if r.FormalRenewalGrantSHA256 != f.grantPin || r.Mode != "live" || r.Stage != "formal" {
					t.Fatalf("real fixed7 authority was not reached: %s", b)
				}
			}
			if got := r.Deadline.Sub(r.StartedAt); got != want {
				t.Errorf("actual initialize deadline: got %v want %v", got, want)
			}
			if len(r.Calls) != 0 || r.Complete || r.Budget.Calls != 0 || r.Budget.Attempts != 0 {
				t.Errorf("keyless probe must not run formal qualification: %s", b)
			}
		})
	}
	f.checkHistory(t)
}
