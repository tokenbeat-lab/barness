package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func kimiQualificationPromotion(t *testing.T, dirs []string) {
	t.Helper()
	ev := run.Case(t, "KIMI-K16-independent-controlled-proposal")
	binary := filepath.Join(t.TempDir(), "regionalpromote")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalpromote")
	build.Dir, build.Env = "../..", regionalChildEnvironment()
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("promotion build: %v %s", err, raw)
	}
	before, err := os.ReadFile("../catalog_kimi.json")
	if err != nil {
		t.Fatal(err)
	}
	proposal := filepath.Join(t.TempDir(), "controlled.json")
	invoke := func(inputs []string, path string, activate bool) ([]byte, error) {
		args := []string{"-kimi", "-controlled", "-proposal", path}
		if activate {
			args = []string{"-kimi", "-receipt", path}
		}
		cmd := exec.CommandContext(ctxFor(t), binary, append(args, inputs...)...)
		cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
		return cmd.CombinedOutput()
	}
	raw, err := invoke(dirs, proposal, false)
	ev.Record("proposal-process", map[string]any{"error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
	if !ev.Check("all six own controlled proofs produce only nonproduction proposal", err == nil, "%v %s", err, raw) {
		return
	}
	data, err := os.ReadFile(proposal)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Kind, Mode, Target string
		Candidates         []supportmatrix.RegionalPromotionProof
		Catalog            struct {
			Active                    bool
			Version, ProvenanceSHA256 string
			Models                    []any
		}
	}
	mustUnmarshal(t, data, &got)
	ev.Check("proposal has independent six identities source provenance and target", got.Kind == "controlled-kimi-catalog-proposal" && got.Mode == "controlled" && got.Target == "ai/catalog_kimi.json" && !got.Catalog.Active && len(got.Catalog.Models) == 6 && len(got.Candidates) == 6 && got.Catalog.ProvenanceSHA256 != "", "%s", data)
	ev.Fixture("controlled-proposal.json", data)
	for _, mode := range []string{"missing", "duplicate", "controlled-live", "stale", "partial"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-promotion-refusal-"+mode)
			inputs := append([]string{}, dirs...)
			activate := mode == "controlled-live"
			if mode == "missing" {
				inputs = inputs[:5]
			}
			if mode == "duplicate" {
				inputs[5] = inputs[0]
			}
			if mode == "stale" || mode == "partial" {
				copy := filepath.Join(t.TempDir(), filepath.Base(inputs[0]))
				if err := os.CopyFS(copy, os.DirFS(inputs[0])); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(copy, supportmatrix.RegionalReportFile)
				rraw, _ := os.ReadFile(path)
				var r map[string]any
				mustUnmarshal(t, rraw, &r)
				if mode == "stale" {
					r["regional"].(map[string]any)["sourceSha256"] = "stale"
				} else {
					r["scenarios"] = r["scenarios"].([]any)[:8]
				}
				if err := os.WriteFile(path, mustMarshal(t, r), 0600); err != nil {
					t.Fatal(err)
				}
				if err := supportmatrix.SealRegionalBundle(copy); err != nil {
					t.Fatal(err)
				}
				inputs[0] = copy
			}
			path := filepath.Join(t.TempDir(), "refused.json")
			if activate {
				base, err := filepath.Abs(filepath.Join("../..", ".evidence/barness-ai-kimi/promotion"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(base, 0700); err != nil {
					t.Fatal(err)
				}
				parent, err := os.MkdirTemp(base, "controlled-refusal-")
				if err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(parent, "lineage.json")
			}
			raw, err := invoke(inputs, path, activate)
			ev.Record("process", map[string]any{"error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			ev.Check("invalid or controlled proof cannot activate Kimi", err != nil, "%s", raw)
			_, st := os.Lstat(path)
			ev.Check("no rejected lineage output", os.IsNotExist(st), "created receipt")
		})
	}
	after, err := os.ReadFile("../catalog_kimi.json")
	ev.Check("all rejected activations preserve inactive production bytes", err == nil && string(before) == string(after), "%v", err)
}
