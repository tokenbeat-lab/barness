package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Real controlled full-fourteen FAIL and independent single PASS, with native
// exits and budgets intact. Never delete rows from a full successful packet.
func TestMiniMaxThinkingComposition(t *testing.T) {
	kimiHistoryPrivateStore(t)
	ev := run.Case(t, "MMTS-composition")
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := kimiRetainedBinary(t, miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true), "minimax")
	consumer := miniMaxAuthorityBuild(t, "releasegate", "./ai/release/cmd/releasegate", false)
	id, _ := minimaxlive.Select("messages", minimaxlive.M3, "standard")
	h, _ := livegrant.FileHash(binary)
	full, _ := minimaxlive.ProcessScope(id, "controlled", "candidate", h)
	baseline, _, cmd := mmThinkingProcess(t, ev, binary, livegrant.Grant{Schema: 1, Scope: full}, "thinking-on-read-error", "")
	if !ev.Check("full native baseline really fails", cmd.ProcessState.ExitCode() == 1, "exit=%d", cmd.ProcessState.ExitCode()) {
		return
	}
	manifest, _ := os.ReadFile(filepath.Join(baseline, "manifest.json"))
	g := mmThinkingTemplate(t, ev, cli, binary, evidence.Digest(manifest))
	single, r, cmd := mmThinkingProcess(t, ev, binary, g, "success-messages", "")
	if !ev.Check("single native process actually succeeds", cmd.ProcessState.ExitCode() == 0 && r.Outcome == "PASS" && !r.Complete && len(r.Calls) == 1 && r.Budget == (minimaxlive.Budget{Calls: 1, ReservedAttempts: 1, Attempts: 1, RequestedOutput: 2048, ReservedInput: 2048}), "exit=%d outcome=%s failure=%s budget=%+v", cmd.ProcessState.ExitCode(), r.Outcome, r.Failure, r.Budget) {
		return
	}
	for _, d := range []string{baseline, single} {
		_, e := minimaxlive.VerifyCapture(d, false)
		ev.Check("partial cannot independently certify "+filepath.Base(d), e != nil, "accepted")
	}
	dir := filepath.Join(filepath.Dir(run.Dir()), "minimax-compositions", filepath.Base(run.Dir()))
	if e := os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	desc := map[string]any{"schema": 1, "kind": "minimax-messages-thinking-composition", "combo": "minimax-messages", "model": minimaxlive.M3, "baseline": mmThinkingRef(t, baseline), "supplement": mmThinkingRef(t, single)}
	path := filepath.Join(dir, "minimax-thinking-composition.json")
	miniMaxAuthorityWrite(t, path, desc)
	check := func(mode, stage string) ([]byte, error) {
		c := exec.CommandContext(ctxFor(t), consumer, "-inspect-bundle", dir, "-mode", mode, "-stage", stage)
		c.Dir, c.Env = "../..", regionalChildEnvironment()
		return c.CombinedOutput()
	}
	out, e := check("controlled", "candidate")
	ev.Record("consumer", map[string]any{"output": regionalSafeOutput(out), "error": errString(e)})
	if !ev.Check("fixed fourteen contributions verify", e == nil, "%v %s", e, out) {
		return
	}
	var q struct {
		Qualification, Mode string
		Coverage            map[string]any
		Measurements        []json.RawMessage
		ExcludedFailures    []json.RawMessage
	}
	mustUnmarshal(t, out, &q)
	ev.Check("two measurements and original failed case remain", q.Qualification == "CONTROLLED_COMPLETE" && q.Mode == "controlled" && len(q.Coverage) == 14 && len(q.Measurements) == 2 && len(q.ExcludedFailures) == 1, "%s", out)
	_, e = check("live", "candidate")
	ev.Check("controlled never becomes live", e != nil, "accepted")
	for _, name := range []string{"same-source", "manifest", "report", "producer", "unknown", "nested"} {
		var bad map[string]any
		mustUnmarshal(t, mustMarshal(t, desc), &bad)
		switch name {
		case "same-source":
			bad["supplement"] = bad["baseline"]
		case "manifest":
			bad["baseline"].(map[string]any)["manifestSHA256"] = strings.Repeat("b", 64)
		case "report":
			bad["baseline"].(map[string]any)["reportSHA256"] = strings.Repeat("b", 64)
		case "producer":
			bad["baseline"].(map[string]any)["producer"].(map[string]any)["sha256"] = strings.Repeat("b", 64)
		case "unknown":
			bad["pass"] = true
		case "nested":
			bad["baseline"].(map[string]any)["bundle"] = repoRelative(t, dir)
		}
		miniMaxAuthorityWrite(t, path, bad)
		_, e = check("controlled", "candidate")
		ev.Check("reject "+name, e != nil, "accepted")
	}
	miniMaxAuthorityWrite(t, path, desc)
	mmThinkingProducerFailures(t, ev, desc, path, baseline, check)
	nativeOut := dir + "-formal-refusal"
	native := exec.CommandContext(ctxFor(t), consumer, "-live-bundle", repoRelative(t, dir), "-out", nativeOut)
	native.Dir, native.Env = "../..", regionalChildEnvironment()
	o, e := native.CombinedOutput()
	var proof struct {
		Pass   bool
		Inputs map[string]string
	}
	data, re := os.ReadFile(filepath.Join(nativeOut, "live-consumer.json"))
	if re == nil {
		mustUnmarshal(t, data, &proof)
	}
	ev.Check("formal refusal retains full input closure", e != nil && re == nil && !proof.Pass && len(proof.Inputs) > 150, "%v %s inputs=%d", e, o, len(proof.Inputs))
	for _, b := range []string{baseline, single} {
		for _, n := range []string{"manifest.json", minimaxlive.ReportFile, minimaxlive.SealFile, "process-grant.json"} {
			key := "@minimax/" + repoRelative(t, filepath.Join(b, n))
			ev.Check("input "+key, proof.Inputs[key] != "", "missing")
		}
	}
	ev.Record("composition-pointer", map[string]any{"dir": dir, "descriptor": desc, "actualVendorCalls": 0, "privateGit": os.Getenv("GIT_DIR")})
}

func mmThinkingProcess(t *testing.T, ev *evidence.Case, binary string, g livegrant.Grant, script, variant string) (string, minimaxlive.Report, *exec.Cmd) {
	t.Helper()
	g.ID = "synthetic-mmts-" + evidence.Digest([]byte(run.Dir() + t.Name() + script + variant))[:24]
	g.AccountAlias, g.RegionAlias = "synthetic-account", "synthetic-region"
	if g.ExpiresAt.IsZero() {
		g.ExpiresAt = time.Now().UTC().Add(time.Hour)
	}
	base := filepath.Join(filepath.Dir(run.Dir()), "minimax-thinking-native", filepath.Base(run.Dir()), g.ID)
	file := filepath.Join(t.TempDir(), "grant.json")
	entry, timeout, count := "^TestMiniMaxThinkingOnSupplement$", "180s", "1"
	if len(g.Scope.Scenes) > 1 && variant != "single-with-full" || variant == "full-with-single" {
		entry, timeout = "^TestMiniMaxCandidate$", "1800s"
	}
	if variant == "entry" {
		entry = "^TestMiniMaxFormal$"
	}
	if variant == "timeout" {
		timeout = "181s"
	}
	if variant == "count" {
		count = "2"
	}
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run="+entry, "-test.count="+count, "-test.timeout="+timeout)
	cmd.Dir, cmd.Env = "../..", append(regionalChildEnvironment(), "BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=messages", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-mmts-key", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_CONTROLLED="+script, "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=synthetic-region", "BARNESS_MINIMAX_GRANT="+file, "BARNESS_AI_EVIDENCE_DIR="+base)
	if !privateGrantProcess(t, ev, cmd) {
		t.Fatal("unowned private namespace")
	}
	miniMaxAuthorityWrite(t, file, g)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now().UTC()
	e := cmd.Run()
	finish := time.Now().UTC()
	ev.Record("process-"+g.ID, map[string]any{"argv": cmd.Args, "error": errString(e), "output": regionalSafeOutput(append(stdout.Bytes(), stderr.Bytes()...)), "vendorCalls": 0})
	if cmd.ProcessState == nil {
		t.Fatal("producer never started", e)
	}
	paths, _ := filepath.Glob(filepath.Join(base, "*", minimaxlive.ReportFile))
	if len(paths) != 1 {
		t.Fatal("native report missing", string(stdout.Bytes()), string(stderr.Bytes()))
	}
	dir := filepath.Dir(paths[0])
	r, e := minimaxlive.Load(dir)
	if e != nil {
		t.Fatal(e)
	}
	files := map[string]string{}
	commandRaw := mustMarshal(t, map[string]any{"schema": 1, "argv": cmd.Args, "stage": "e2e", "route": "minimax-messages"})
	if e := os.WriteFile(filepath.Join(filepath.Dir(dir), "command.json"), commandRaw, 0600); e != nil {
		t.Fatal(e)
	}
	files["command.json"] = evidence.Digest(commandRaw)

	for n, b := range map[string][]byte{"stdout.jsonl": stdout.Bytes(), "stderr.log": stderr.Bytes(), "exit-code": []byte(fmt.Sprint(cmd.ProcessState.ExitCode()) + "\n")} {
		p := filepath.Join(filepath.Dir(dir), n)
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
		files[n] = evidence.Digest(b)
	}
	source := map[string]string{}
	inputs := map[string]string{}
	for n, h := range r.ExecutionSource {
		if strings.HasSuffix(n, ".md") {
			inputs[n] = h
		} else {
			source[n] = h
		}
	}
	manifest, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
	status := "PASS"
	if cmd.ProcessState.ExitCode() != 0 {
		status = "FAIL"
	}
	receipt := map[string]any{"schema": 1, "stage": "e2e", "route": "minimax-messages", "status": status, "exitCode": cmd.ProcessState.ExitCode(), "argv": cmd.Args, "source": map[string]any{"commit": r.GitCommit, "files": source}, "selectedInputs": inputs, "files": files, "commandSHA256": files["command.json"], "startedAt": start, "finishedAt": finish, "bundles": []any{map[string]string{"path": filepath.Base(dir), "manifestSHA256": evidence.Digest(manifest)}}}
	if cmd.ProcessState.ExitCode() != 0 {
		receipt["bundles"] = []any{}
	}
	miniMaxAuthorityWrite(t, filepath.Join(filepath.Dir(dir), "producer-receipt.json"), receipt)
	return dir, r, cmd
}
func mmThinkingRef(t *testing.T, dir string) map[string]any {
	t.Helper()
	hash := func(p string) string {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return evidence.Digest(b)
	}
	receipt := filepath.Join(filepath.Dir(dir), "producer-receipt.json")
	return map[string]any{"bundle": repoRelative(t, dir), "manifestSHA256": hash(filepath.Join(dir, "manifest.json")), "reportSHA256": hash(filepath.Join(dir, minimaxlive.ReportFile)), "sealSHA256": hash(filepath.Join(dir, minimaxlive.SealFile)), "producer": map[string]string{"path": repoRelative(t, receipt), "sha256": hash(receipt)}}
}

// Pins are refreshed only for synthetic negative copies. Source/exit/entry
// contradictions must still be rejected by semantics, not merely outer hashes.
func mmThinkingProducerFailures(t *testing.T, ev *evidence.Case, desc map[string]any, path, baseline string, check func(string, string) ([]byte, error)) {
	t.Helper()
	original := filepath.Join(filepath.Dir(baseline), "producer-receipt.json")
	raw, e := os.ReadFile(original)
	if e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"exit", "entry", "count", "source", "selected-conflict", "missing-selected", "binary", "missing-log", "route", "schema", "command"} {
		var p map[string]any
		mustUnmarshal(t, raw, &p)
		switch kind {
		case "route":
			p["route"] = "minimax-chat"
		case "schema":
			p["schema"] = 2
		case "command":
			p["commandSHA256"] = strings.Repeat("b", 64)
		case "exit":
			p["exitCode"] = 0
			p["status"] = "PASS"
		case "entry":
			a := p["argv"].([]any)
			a[1] = "-test.run=^TestMiniMaxThinkingOnSupplement$"
		case "count":
			a := p["argv"].([]any)
			a = append(a, "-test.count=2")
			p["argv"] = a
		case "source":
			p["source"].(map[string]any)["commit"] = strings.Repeat("0", 40)
		case "selected-conflict":
			for n := range p["selectedInputs"].(map[string]any) {
				p["source"].(map[string]any)["files"].(map[string]any)[n] = strings.Repeat("b", 64)
				break
			}
		case "missing-selected":
			for n := range p["selectedInputs"].(map[string]any) {
				delete(p["selectedInputs"].(map[string]any), n)
				break
			}
		case "binary":
			p["argv"].([]any)[0] = "/missing-producer"
		case "missing-log":
			p["files"].(map[string]any)["stderr.log"] = strings.Repeat("b", 64)
		}
		badPath := filepath.Join(filepath.Dir(original), "negative-"+kind+"-producer.json")
		miniMaxAuthorityWrite(t, badPath, p)
		var d map[string]any
		mustUnmarshal(t, mustMarshal(t, desc), &d)
		d["baseline"].(map[string]any)["producer"] = map[string]string{"path": repoRelative(t, badPath), "sha256": minimaxlive.Hash(mustMarshal(t, p))}
		miniMaxAuthorityWrite(t, path, d)
		out, e := check("controlled", "candidate")
		ev.Check("producer semantic rejection "+kind, e != nil, "error=%v outputSHA=%s", e, evidence.Digest(out))
	}
	miniMaxAuthorityWrite(t, path, desc)
}
