package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The positive composition uses two real controlled processes: an independently
// tagged nine-call fixture and the production one-call entry. No report rows or
// budgets are removed from a ten-call execution to fabricate a baseline.
func TestKimiHistoryComposition(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary, digest := kimiQualificationBinary(t)
	baselineBinary := kimiNineFixtureBinary(t)
	consumer := kimiQualificationConsumer(t)
	var compositions []string
	for _, route := range kimiQualificationRoutes {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KHS-composition-"+route)
			kimiPromotionFixtureCommit(t, true)
			baseline := kimiNineFixture(t, ev, baselineBinary, route)
			kimiPromotionFixtureCommit(t, false)
			manifest, _ := os.ReadFile(filepath.Join(baseline, "manifest.json"))
			supplement, srv, _ := kimiHistoryCapture(t, ev, binary, digest, route, evidence.Digest(manifest), false)
			defer srv.Close()
			if baseline == "" || supplement == "" {
				return
			}
			dir := filepath.Join(filepath.Dir(run.Dir()), "kimi-compositions", filepath.Base(run.Dir()), route)
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			descriptor := map[string]any{"schema": 1, "kind": "kimi-history-composition", "combo": route, "model": "kimi-k3", "baseline": kimiCompositionRef(t, baseline), "supplement": kimiCompositionRef(t, supplement)}
			path := filepath.Join(dir, "kimi-composition.json")
			if err := os.WriteFile(path, mustMarshal(t, descriptor), 0600); err != nil {
				t.Fatal(err)
			}
			check := func(mode, stage string) ([]byte, error) {
				cmd := exec.CommandContext(ctxFor(t), consumer, "-inspect-bundle", dir, "-mode", mode, "-stage", stage)
				cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
				return cmd.CombinedOutput()
			}
			raw, err := check("controlled", "candidate")
			ev.Record("consumer", map[string]any{"output": regionalSafeOutput(raw), "error": errString(err), "vendorCalls": 0})
			if !ev.Check("two actual processes cover exactly ten cases", err == nil, "%v %s", err, regionalSafeOutput(raw)) {
				return
			}
			var q struct {
				Mode, Qualification string
				Expected            []string
				Measurements        []json.RawMessage
				Coverage            map[string]any
			}
			mustUnmarshal(t, raw, &q)
			ev.Check("controlled remains controlled with both measurements", q.Mode == "controlled" && q.Qualification == "CONTROLLED_COMPLETE" && len(q.Expected) == 10 && len(q.Coverage) == 10 && len(q.Measurements) == 2, "%s", raw)
			nativeOut := filepath.Join(filepath.Dir(dir), route+"-formal-refusal")
			native := exec.CommandContext(ctxFor(t), consumer, "-live-bundle", repoRelative(t, dir), "-out", nativeOut)
			native.Dir, native.Env = "../..", regionalChildEnvironment()
			out, nativeErr := native.CombinedOutput()
			ev.Record("formal-consumer-refusal", map[string]any{"argv": native.Args, "error": errString(nativeErr), "output": regionalSafeOutput(out)})
			var proof struct {
				Pass   bool
				Inputs map[string]string
			}
			data, readErr := os.ReadFile(filepath.Join(nativeOut, "live-consumer.json"))
			if readErr == nil {
				mustUnmarshal(t, data, &proof)
			}
			ev.Check("formal refusal still binds complete transitive input inventory", nativeErr != nil && readErr == nil && !proof.Pass && len(proof.Inputs) > 100, "err=%v inputs=%d", nativeErr, len(proof.Inputs))
			for _, bundle := range []string{baseline, supplement} {
				for _, name := range []string{"manifest.json", "regional-integrity.json", "regional-grant.json", "regional-report.json"} {
					key := "@kimi/" + repoRelative(t, filepath.Join(bundle, name))
					ev.Check("transitive input "+key, proof.Inputs[key] != "", "missing")
				}
				for _, name := range []string{"producer-receipt.json", "stdout.jsonl", "stderr.log", "exit-code"} {
					key := "@kimi/" + repoRelative(t, filepath.Join(filepath.Dir(bundle), name))
					ev.Check("producer input "+key, proof.Inputs[key] != "", "missing")
				}
			}
			if route == "kimi-cn-chat" {
				kimiProducerFailures(t, ev, descriptor, path, baseline, check)
			}
			_, err = check("live", "candidate")
			ev.Check("controlled sources cannot become actual candidate evidence", err != nil, "accepted")
			_, err = check("live", "formal")
			ev.Check("controlled sources cannot become formal", err != nil, "accepted")
			for _, mutation := range []string{"same-source", "cross-route", "wrong-model", "wrong-manifest", "nested", "unknown"} {
				var bad map[string]any
				mustUnmarshal(t, mustMarshal(t, descriptor), &bad)
				switch mutation {
				case "same-source":
					bad["supplement"] = bad["baseline"]
				case "cross-route":
					bad["combo"] = "kimi-cn-chat"
					if route == "kimi-cn-chat" {
						bad["combo"] = "kimi-intl-chat"
					}
				case "wrong-model":
					bad["model"] = "other"
				case "wrong-manifest":
					bad["baseline"].(map[string]any)["manifestSHA256"] = "bad"
				case "nested":
					bad["baseline"].(map[string]any)["bundle"] = repoRelative(t, dir)
				case "unknown":
					bad["passed"] = true
				}
				if err := os.WriteFile(path, mustMarshal(t, bad), 0600); err != nil {
					t.Fatal(err)
				}
				out, e := check("controlled", "candidate")
				ev.Check("reject "+mutation, e != nil, "%s", regionalSafeOutput(out))
			}
			if err := os.WriteFile(path, mustMarshal(t, descriptor), 0600); err != nil {
				t.Fatal(err)
			}
			// The host input is validated by current encoding, not replaced in a hook.
			requestPath := filepath.Join(baseline, supportmatrix.RegionalCaseID(route, "kimi-k3/full-complete"), "request.json")
			prior, _ := os.ReadFile(requestPath)
			var changed map[string]any
			mustUnmarshal(t, prior, &changed)
			changed["Messages"] = []any{map[string]any{"content": []any{map[string]any{"text": "invented host request"}}}}
			if err := os.WriteFile(requestPath, mustMarshal(t, changed), 0600); err != nil {
				t.Fatal(err)
			}
			// Do not reseal originals: this is an explicitly controlled fixture copy.
			// Other resealed semantic mutations are covered in a dedicated test.
			_, err = check("controlled", "candidate")
			ev.Check("transitive original mutation invalidates composition", err != nil, "accepted changed original")
			if err := os.WriteFile(requestPath, prior, 0600); err != nil {
				t.Fatal(err)
			}
			compositions = append(compositions, dir)
			ev.Record("composition-pointer", map[string]any{"dir": dir, "descriptor": descriptor})
		})
	}
	if len(compositions) == 6 {
		kimiCompositionPromotion(t, compositions)
	}
}

func repoRelative(t *testing.T, path string) string {
	t.Helper()
	root, _ := filepath.Abs("../..")
	p, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(p) {
		t.Fatal("not repository local", path)
	}
	return filepath.ToSlash(p)
}
func kimiCompositionRef(t *testing.T, dir string) map[string]any {
	t.Helper()
	hash := func(name string) string {
		raw, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		return evidence.Digest(raw)
	}
	receipt := filepath.Join(filepath.Dir(dir), "producer-receipt.json")
	raw, e := os.ReadFile(receipt)
	if e != nil {
		t.Fatal(e)
	}
	return map[string]any{"bundle": repoRelative(t, dir), "manifestSHA256": hash("manifest.json"), "reportSHA256": hash(supportmatrix.RegionalReportFile), "sealSHA256": hash("regional-integrity.json"), "producer": map[string]string{"path": repoRelative(t, receipt), "sha256": evidence.Digest(raw)}}
}

func kimiNineFixtureBinary(t *testing.T) string {
	t.Helper()
	source, e := supportmatrix.RegionalExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "kimi-nine.test")
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live,kimi_baseline_fixture", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource="+supportmatrix.RegionalSourceHash(source), "-o", path, "./ai/live")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatal(e, regionalSafeOutput(out))
	}
	return path
}
func kimiNineFixture(t *testing.T, ev *evidence.Case, binary, route string) string {
	t.Helper()
	info, _ := supportmatrix.RegionalRouteOf(route)
	srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
	defer srv.Close()
	for _, kind := range kimiQualificationKinds()[:9] {
		srv.Enqueue(kimiQualificationReply(t, info.API, kind))
	}
	g, e := supportmatrix.RegionalGrantSnapshot(route, "candidate")
	if e != nil {
		t.Fatal(e)
	}
	g.Mode, g.ID, g.AccountAlias, g.RegionAlias = "controlled", "nine-"+route+"-"+evidence.Digest([]byte(run.Dir() + t.Name()))[:12], "synthetic-kimi-account", "synthetic-kimi-region"
	raw, e := os.ReadFile(binary)
	if e != nil {
		t.Fatal(e)
	}
	g.BinarySHA256 = evidence.Digest(raw)
	g.MaxUSD = 10
	g.ExpiresAt = time.Now().UTC().Add(time.Hour)
	g.MaxCalls, g.MaxAttempts, g.MaxRequestedOutputTokens = 9, 9, 36864
	base := filepath.Join(filepath.Dir(run.Dir()), "kimi-nine-native", filepath.Base(run.Dir()), g.ID)
	file := filepath.Join(t.TempDir(), "grant.json")
	binary = kimiRetainedBinary(t, binary, "nine")
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestKimiBaselineControlled$", "-test.timeout=1800s")
	cmd.Dir = "../.."
	cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+route, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+info.Provider, "BARNESS_REGIONAL_GRANT="+file, "BARNESS_AI_EVIDENCE_DIR="+base)
	if !privateGrantProcess(t, ev, cmd) {
		return ""
	}
	if e := os.WriteFile(file, mustMarshal(t, g), 0600); e != nil {
		t.Fatal(e)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	started := time.Now().UTC()
	e = cmd.Run()
	finished := time.Now().UTC()
	out := append(append([]byte{}, stdout.Bytes()...), stderr.Bytes()...)
	ev.Record("baseline-process", map[string]any{"argv": cmd.Args, "error": errString(e), "output": regionalSafeOutput(out), "requests": len(srv.Requests()), "vendorCalls": 0})
	if !ev.Check("separate controlled fixture actually executed nine", e == nil && len(srv.Requests()) == 9, "%v %s", e, regionalSafeOutput(out)) {
		return ""
	}
	paths, _ := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
	if len(paths) != 1 {
		t.Fatal("baseline missing")
	}
	dir := filepath.Dir(paths[0])
	kimiProducerReceipt(t, cmd, dir, stdout.Bytes(), stderr.Bytes(), started, finished)
	return dir
}
