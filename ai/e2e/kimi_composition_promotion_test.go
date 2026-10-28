package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func kimiPromotionCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "regionalpromote")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalpromote")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(raw))
	}
	return binary
}

// Only the retained private authority's ref is changed. Both captures run
// actual processes with exact source maps and different real Git HEADs.
func kimiPromotionFixtureCommit(t *testing.T, parent bool) {
	t.Helper()
	private := os.Getenv("GIT_DIR")
	if private == "" {
		t.Fatal("private authority required")
	}
	args := []string{"--git-dir=" + filepath.Join("../..", ".git"), "rev-parse", "HEAD"}
	if parent {
		args[len(args)-1] = "HEAD^"
	}
	cmd := exec.CommandContext(ctxFor(t), "git", args...)
	raw, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	cmd = exec.CommandContext(ctxFor(t), "git", "--git-dir="+private, "update-ref", "refs/heads/fixture", string(bytes.TrimSpace(raw)))
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(raw))
	}
}

func kimiCompositionPromotion(t *testing.T, dirs []string) {
	t.Helper()
	ev := run.Case(t, "KCP-six-paired-controlled-promotion")
	binary := kimiPromotionCLI(t)
	before, err := os.ReadFile("../catalog_kimi.json")
	if err != nil {
		t.Fatal(err)
	}
	processNumber := 0
	invoke := func(inputs []string, path string, activate bool) ([]byte, error) {
		args := []string{"-kimi", "-controlled", "-proposal", path}
		if activate {
			args = []string{"-kimi", "-receipt", path}
		}
		cmd := exec.CommandContext(ctxFor(t), binary, append(args, inputs...)...)
		cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
		raw, err := cmd.CombinedOutput()
		processNumber++
		ev.Record(fmt.Sprintf("promotion-process-%02d", processNumber), map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
		return raw, err
	}
	path := filepath.Join(t.TempDir(), "proposal.json")
	raw, err := invoke(dirs, path, false)
	if !ev.Check("six actual nine-plus-one pairs accepted", err == nil, "%v %s", err, raw) {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Schema             int
		Kind, Mode, Target string
		Candidates         []struct {
			Combo, Composition string
			Qualification      supportmatrix.KimiQualification
			Inputs             map[string]string
		}
		Catalog struct {
			Active bool
			Models []any
		}
	}
	mustUnmarshal(t, data, &got)
	ev.Check("dedicated paired contract remains nonproduction", got.Schema == 2 && got.Kind == "controlled-kimi-catalog-proposal" && got.Mode == "controlled" && got.Target == "ai/catalog_kimi.json" && !got.Catalog.Active && len(got.Catalog.Models) == 6 && len(got.Candidates) == 6, "%s", data)
	measurements := 0
	commits := map[string]bool{}
	binaries := map[string]bool{}
	for _, p := range got.Candidates {
		q := p.Qualification
		ev.Check("fixed coverage and transitive inputs "+p.Combo, len(q.Measurements) == 2 && len(q.Coverage) == 10 && len(q.Expected) == 10 && q.Mode == "controlled" && p.Composition != "" && len(p.Inputs) > 100, "%+v", q)
		for i, m := range q.Measurements {
			measurements++
			commits[m.Original.GitCommit] = true
			binaries[m.Original.Regional.BinarySHA256] = true
			count := 9
			if i == 1 {
				count = 1
			}
			ev.Check("preserve independent calls grant producer time "+p.Combo+m.Kind, m.Original.Budget.CallsUsed == count && m.Original.Regional.GrantID != "" && m.Evidence.Producer.SHA256 != "" && !m.Original.StartedAt.IsZero(), "%+v", m)
		}
	}
	ev.Check("twelve independent measurements accept differing commits and binaries", measurements == 12 && len(commits) == 2 && len(binaries) == 2, "measurements=%d commits=%d binaries=%d", measurements, len(commits), len(binaries))
	ev.Fixture("controlled-proposal.json", data)
	_, err = invoke(dirs, path, false)
	ev.Check("exclusive proposal refuses overwrite", err != nil, "overwrote")
	after, _ := os.ReadFile(path)
	ev.Check("exclusive proposal bytes preserved", bytes.Equal(data, after), "changed")
	refused := func(label string, inputs []string, activate bool) {
		t.Helper()
		out := filepath.Join(t.TempDir(), "refused.json")
		if activate {
			base, _ := filepath.Abs("../../.evidence/barness-ai-kimi/promotion")
			if err := os.MkdirAll(base, 0700); err != nil {
				t.Fatal(err)
			}
			parent, err := os.MkdirTemp(base, "paired-controlled-refusal-")
			if err != nil {
				t.Fatal(err)
			}
			out = filepath.Join(parent, "receipt.json")
		}
		raw, err := invoke(inputs, out, activate)
		ev.Check("reject "+label, err != nil, "%s", raw)
		_, err = os.Lstat(out)
		ev.Check("no rejected receipt "+label, os.IsNotExist(err), "receipt exists")
	}
	refused("five", dirs[:5], false)
	duplicate := append([]string{}, dirs...)
	duplicate[5] = dirs[0]
	refused("duplicate", duplicate, false)
	refused("controlled cannot activate", dirs, true)
	descriptor := filepath.Join(dirs[0], supportmatrix.KimiCompositionFile)
	original, err := os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	mustUnmarshal(t, original, &c)
	baseline, err := filepath.Abs(filepath.Join("../..", c["baseline"].(map[string]any)["bundle"].(string)))
	if err != nil {
		t.Fatal(err)
	}
	refused("old nine alone", append([]string{baseline}, dirs[1:]...), false)
	for _, kind := range []string{"cross-route", "wrong-pin", "nested"} {
		var bad map[string]any
		mustUnmarshal(t, original, &bad)
		switch kind {
		case "cross-route":
			bad["combo"] = "kimi-intl-chat"
		case "wrong-pin":
			bad["baseline"].(map[string]any)["manifestSHA256"] = "bad"
		case "nested":
			bad["supplement"].(map[string]any)["bundle"] = repoRelative(t, dirs[0])
		}
		if err := os.WriteFile(descriptor, mustMarshal(t, bad), 0600); err != nil {
			t.Fatal(err)
		}
		refused(kind, dirs, false)
	}
	if err := os.WriteFile(descriptor, original, 0600); err != nil {
		t.Fatal(err)
	}
	check := func(string, string) ([]byte, error) {
		return invoke(dirs, filepath.Join(t.TempDir(), "producer-refused.json"), false)
	}
	kimiProducerFailures(t, ev, c, descriptor, baseline, check)
	reportPath := filepath.Join(baseline, supportmatrix.RegionalReportFile)
	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"mixed-mode", "source", "sdk", "model"} {
		var bad map[string]any
		mustUnmarshal(t, report, &bad)
		switch kind {
		case "mixed-mode":
			bad["regional"].(map[string]any)["mode"] = "live"
		case "source":
			bad["regional"].(map[string]any)["sourceSha256"] = "other"
		case "sdk":
			bad["sdk"] = "other"
		case "model":
			bad["model"] = "other"
		}
		if err := os.WriteFile(reportPath, mustMarshal(t, bad), 0600); err != nil {
			t.Fatal(err)
		}
		// Pin the altered report so the consumer must inspect native integrity.
		var badC map[string]any
		mustUnmarshal(t, original, &badC)
		badC["baseline"].(map[string]any)["reportSHA256"] = evidence.Digest(mustMarshal(t, bad))
		if err := os.WriteFile(descriptor, mustMarshal(t, badC), 0600); err != nil {
			t.Fatal(err)
		}
		refused(kind, dirs, false)
	}
	if err := os.WriteFile(reportPath, report, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(descriptor, original, 0600); err != nil {
		t.Fatal(err)
	}
	var r supportmatrix.Report
	mustUnmarshal(t, report, &r)
	ledger := filepath.Join(os.Getenv("GIT_DIR"), "barness-kimi-grants", "id-"+evidence.Digest([]byte(r.Regional.GrantID))+".json")
	ledgerBytes, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledger, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	refused("trusted ledger", dirs, false)
	if err := os.WriteFile(ledger, ledgerBytes, 0600); err != nil {
		t.Fatal(err)
	}
	raw, err = invoke(dirs, filepath.Join(t.TempDir(), "restored.json"), false)
	ev.Check("restored original inputs remain accepted", err == nil, "%v %s", err, raw)
	// A descriptor is only the root of the inputs. New review files must also
	// stay outside its referenced native/producer and trusted authority folders.
	supplement, err := filepath.Abs(filepath.Join("../..", c["supplement"].(map[string]any)["bundle"].(string)))
	if err != nil {
		t.Fatal(err)
	}
	for i, protected := range []string{baseline, filepath.Dir(baseline), supplement, filepath.Join(os.Getenv("GIT_DIR"), "barness-kimi-grants")} {
		out := filepath.Join(protected, fmt.Sprintf("promotion-must-not-write-%d.json", i))
		raw, err := invoke(dirs, out, false)
		ev.Check(fmt.Sprintf("reject output in transitive original directory %d", i), err != nil, "%s", raw)
		_, err = os.Lstat(out)
		ev.Check(fmt.Sprintf("no file added to original directory %d", i), os.IsNotExist(err), "wrote original evidence/authority directory")
	}

	after, err = os.ReadFile("../catalog_kimi.json")
	ev.Check("all controlled processes preserve production catalog bytes", err == nil && bytes.Equal(before, after), "changed")
}
