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
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestMiniMaxAdditiveReleaseContract(t *testing.T) {
	ev := run.Case(t, "MM10-release-contract")
	raw, e := os.ReadFile("../release/traceability.json")
	if e != nil {
		t.Fatal(e)
	}
	plan, e := release.ParseTraceMap(raw)
	if e != nil {
		t.Fatal(e)
	}
	ev.Check("seventeen own formal support identities", len(plan.LiveCombos) == 17 && release.CheckReleasePlan(plan).Current, "count=%d contract=%+v", len(plan.LiveCombos), release.CheckReleasePlan(plan))
	matrix, e := supportmatrix.LoadMatrix("../live/support-matrix.json")
	if e != nil {
		t.Fatal(e)
	}
	ev.Check("seventeen own rows preserve old identities", len(matrix.Rows) == 17, "rows=%d", len(matrix.Rows))
	routes := map[string]bool{}
	for _, row := range matrix.Rows {
		routes[row.Operation+"/"+row.Provider+"/"+row.API] = true
	}
	ev.Check("sixteen distinct protocol routes", len(routes) == 16, "routes=%d", len(routes))
	ev.Check("eighth frozen differential is own M3 Messages", len(plan.Differential.Required) == 8 && slices.Contains(plan.Differential.Required, "MM-MESSAGES"), "%v", plan.Differential.Required)
	for _, route := range []string{"chat", "responses", "messages"} {
		combo := "minimax-" + route
		i := slices.IndexFunc(matrix.Rows, func(r supportmatrix.Row) bool { return r.Combo == combo })
		ev.Check("own MiniMax row "+combo, i >= 0 && matrix.Rows[i].ValidIdentity(), "missing exact own row")
		changed := plan
		changed.LiveCombos = slices.DeleteFunc(slices.Clone(plan.LiveCombos), func(c string) bool { return c == combo })
		ev.Check("removing own proof fails "+combo, !release.CheckReleasePlan(changed).Current, "missing proof accepted")
	}
	for prefix, count := range map[string]int{"MM-R": 28, "MM-T": 4, "MM-A": 20, "MM-US": 80} {
		for i := 1; i <= count; i++ {
			id := fmt.Sprintf("%s%02d", prefix, i)
			at := slices.IndexFunc(plan.Items, func(item release.TraceItem) bool { return item.ID == id })
			ev.Check("finite required mapping "+id, at >= 0 && len(plan.Items[at].RequiredCases) > 0, "missing finite own cases")
			if at >= 0 && i == 1 {
				// One mutation per independently required category; all finite sets
				// above are checked without repeatedly hashing the same large map.
				changed := plan
				changed.Items = slices.Clone(plan.Items)
				changed.Items[at].RequiredCases = nil
				ev.Check("finite selection cannot be removed "+id, !release.CheckReleasePlan(changed).Current, "empty finite mapping accepted")
			}
		}
	}
	for _, prefix := range []string{"GLM-R01", "GLM-A15", "GLM-US62", "UC-R17", "UC-US50", "OD-R20", "TS-R12"} {
		ev.Check("old contract retained "+prefix, slices.ContainsFunc(plan.Items, func(item release.TraceItem) bool { return item.ID == prefix }), "old contract removed")
	}
}

// Failure inventory for the added public consumer: six absent/failed cases,
// shrunken original4/8 or added3/6 batches, changed protocol/routes/dimensions,
// missing headroom, excessive memory, incomplete captures/results/Observer and
// leaked resources must all fail independently in normal and race evidence.
func TestMiniMaxReleaseRequiresOwnDesignLoads(t *testing.T) {
	ev := run.Case(t, "MM10-release-required-loads")
	report := release.Evaluate(release.Inputs{})
	gate := slices.IndexFunc(report.Gates, func(g release.Gate) bool { return g.ID == release.GateEvidence })
	if gate < 0 {
		t.Fatal("required artifacts gate missing")
	}
	detail := strings.Join(report.Gates[gate].Problems, "\n")
	for _, host := range []string{"local", "cloud"} {
		for _, route := range []string{"chat", "responses", "messages"} {
			name := "minimax-design-load-" + host + "-" + route
			ev.Check("normal missing own load "+name, strings.Contains(detail, name+":"), "missing load was not required")
			ev.Check("race missing own load "+name, strings.Contains(detail, "race-"+name+":"), "missing race load was not required")
		}
	}
}

// These are actual public Client load captures, imported through the same
// bundle and consumer boundary the release command uses. Rehashed mutations
// exercise semantic checks independently of the artifact hash inventory.
func TestMiniMaxReleaseLoadConsumer(t *testing.T) {
	ev := run.Case(t, "MM10-release-load-consumer")
	base := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-count=1", "-timeout=90s", "./ai/e2e", "-run", "^TestMiniMaxSharedDesignLoad$")
	command.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+v)
		}
	}
	command.Env = append(command.Env, "BARNESS_AI_PRESSURE=1", "BARNESS_AI_EVIDENCE_DIR="+base)
	output, err := command.CombinedOutput()
	if !ev.Check("independent complete six-load process succeeds", err == nil, "%v %s", err, output) {
		return
	}
	dir, err := release.FindBundle(base)
	if err != nil {
		t.Fatal(err)
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
	race := release.RaceEvidence(dir, release.TraceMap{}, release.Ledger{})
	for _, check := range checks {
		at := slices.IndexFunc(race, func(c release.EvidenceCheck) bool { return c.Name == "race-"+check.Name })
		ev.Check("independent race consumer "+check.Name, at >= 0 && race[at].Status == release.Pass, "race consumer missing or failed")
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
