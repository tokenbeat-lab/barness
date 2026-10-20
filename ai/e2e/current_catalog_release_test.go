package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestCurrentCatalogRelease(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC14-current-release-"+entry)
			u := newJevBoxWorld(t, ev, func(c *ai.Config) { *c.Catalog = ai.BuiltinCatalog() })
			req := jevBoxInput(t)
			enqueueJevBox(t, ev, u, jevBoxAnswer, "0.9")
			var result ai.ClassifierResult
			var err error
			if entry == "plain" {
				result, err = u.client.Classify(ctxFor(t), textScope("req-release-"+entry), jevContextTarget(), req, jevContextOptions())
			} else {
				result, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-release-"+entry), jevContextTarget(), req, jevContextOptions())
			}
			ev.Check("current formal catalog executes image Box and existence", err == nil && len(result.Answers) == 2 && len(result.ImageInputs) == 1 && result.Metadata.CatalogHash == builtinCatalogHash, "public Classify failed: %v", err)
			recordJevContext(ev, u, req, result, err)
			u.released(ev)
			report := releaseProcess(t, ev)
			ev.Check("all seventeen independent identities required", len(report.Live.Rows) == 17, "got %d rows", len(report.Live.Rows))
			for _, identity := range []struct{ combo, model string }{{"typesafe-classifier", "jev-1.13.0"}, {supportmatrix.JevFormalCombo, "jev-1.14.0"}, {"openai-decisions", "gpt-6-luna"}} {
				i := slices.IndexFunc(report.Live.Rows, func(r release.LiveRow) bool { return r.Combo == identity.combo })
				ev.Check("independent model "+identity.combo, i >= 0 && report.Live.Rows[i].Model == identity.model, "model missing or substituted")
			}
		})
	}
}

func TestCurrentReleaseTraceAndBlocker(t *testing.T) {
	ev := run.Case(t, "UC14-trace-and-external-blocker")
	report := releaseProcess(t, ev)
	for _, group := range []struct {
		prefix  string
		numbers []int
	}{
		{"R", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}},
		{"T", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
		{"US", []int{1, 3, 17, 23, 25, 32, 33, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50}},
	} {
		for _, number := range group.numbers {
			id := fmt.Sprintf("UC-%s%02d", group.prefix, number)
			ev.Check("required trace "+id, slices.ContainsFunc(report.Trace, func(r release.TraceResult) bool { return r.ID == id }), "missing trace")
		}
	}
	i := slices.IndexFunc(report.Snapshots, func(s release.Snapshot) bool { return s.Name == "pressure-stream-integrity" })
	if i >= 0 {
		ev.Record("original-blocker", report.Snapshots[i])
	}
	tracker, err := os.ReadFile("../../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md")
	if err != nil {
		t.Fatal("original pressure tracker unavailable")
	}
	// The original tracker is the authority; legitimate resolution must clear
	// this check without editing the integration contract or closing it here.
	statuses := regexp.MustCompile(`(?m)^Status:`).FindAll(tracker, -1)
	resolved := regexp.MustCompile(`(?m)^Status:[ \t]*resolved[ \t]*$`).Match(tracker) && len(statuses) == 1
	ev.Record("tracker-expectation", map[string]bool{"resolved": resolved})
	ev.Check("historical tracker controls release blocker", i >= 0 && report.Snapshots[i].Current == resolved && strings.Contains(report.Snapshots[i].Detail, "07-pressure-stream-integrity"), "tracker and release gate disagree")
	for _, name := range []string{"multimodal-images-plain", "multimodal-images-hooked", "race-multimodal-images-plain", "race-multimodal-images-hooked"} {
		e := releaseEvidence(t, report, name)
		ev.Check("missing public-entry load never passes "+name, e.Status != release.Pass, "missing load passed")
	}
}

// Clone only versioned sources/fixtures: no user credentials or original tracker
// writes. The child runs the same public release CLI assertion in both states.
func TestCurrentReleaseTrackerTransition(t *testing.T) {
	ev := run.Case(t, "UC14-original-blocker-transition")
	clone := t.TempDir()
	const issue = ".scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md"
	list := exec.CommandContext(ctxFor(t), "git", "ls-files", "-z", "--", "ai", "go.mod", "go.sum", issue)
	list.Dir = "../.."
	files, err := list.Output()
	if err != nil {
		t.Fatal("versioned source inventory unavailable")
	}
	for _, name := range strings.Split(string(files), "\x00") {
		if name == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(clone, name)
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, state := range []string{"ready-for-agent", "resolved"} {
		raw := []byte("# Synthetic original tracker fixture\n\nStatus: " + state + "\n")
		ev.Fixture("tracker-"+state+".md", raw)
		if err := os.WriteFile(filepath.Join(clone, issue), raw, 0600); err != nil {
			t.Fatal(err)
		}
		out := t.TempDir()
		cmd := exec.CommandContext(ctxFor(t), "go", "test", "./ai/e2e", "-run", "^TestCurrentReleaseTraceAndBlocker$", "-count=1")
		cmd.Dir = clone
		for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
			if value, ok := os.LookupEnv(key); ok {
				cmd.Env = append(cmd.Env, key+"="+value)
			}
		}
		cmd.Env = append(cmd.Env, "BARNESS_AI_EVIDENCE_DIR="+out)
		err := cmd.Run()
		ev.Check("public gate follows original tracker "+state, err == nil, "state %s rejected: %v", state, err)
		if bundle, e := release.FindBundle(out); e == nil {
			if captured, e := os.ReadFile(filepath.Join(bundle, "UC14-trace-and-external-blocker", "original-blocker.json")); e == nil {
				ev.Fixture("snapshot-"+state+".json", captured)
			}
		}
	}
}
