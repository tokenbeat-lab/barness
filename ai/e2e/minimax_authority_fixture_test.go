package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

type miniMaxAuthorityFixture struct {
	dir, source, git string
	report           minimaxlive.Report
}

func miniMaxAuthorityBuild(t *testing.T, name, pkg string, native bool) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), name)
	args := []string{"build", "-o", binary, pkg}
	if native {
		source, err := minimaxlive.ExecutionSource()
		if err != nil {
			t.Fatal(err)
		}
		args = []string{"test", "-c", "-tags", "live", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource=" + livegrant.Hash(mustMarshal(t, source)), "-o", binary, pkg}
	}
	cmd := exec.CommandContext(ctxFor(t), "go", args...)
	cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v %s", name, err, output)
	}
	return binary
}

func miniMaxAuthorityCapture(t *testing.T, ev *evidence.Case, binary, route string) miniMaxAuthorityFixture {
	t.Helper()
	binaryHash, err := livegrant.FileHash(binary)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
	scope, err := minimaxlive.ProcessScope(id, "controlled", "candidate", binaryHash)
	if err != nil {
		t.Fatal(err)
	}
	grant := livegrant.Grant{Schema: 1, ID: "synthetic-mm20-native-" + route + "-" + evidence.Digest([]byte(run.Dir()))[:16], AccountAlias: "synthetic-account", RegionAlias: "synthetic-region", ExpiresAt: time.Now().UTC().Add(time.Hour), Scope: scope}
	approval := filepath.Join(t.TempDir(), "grant.json")
	miniMaxAuthorityWrite(t, approval, grant)
	base := filepath.Join(filepath.Dir(run.Dir()), "minimax-authority-native", filepath.Base(run.Dir()), evidence.Digest([]byte(t.Name()))[:12], route)
	cmd := miniMaxFreshLedgerCommand(t, binary, route, approval, base)
	output, err := cmd.CombinedOutput()
	ev.Record("native-process", map[string]any{"argv": cmd.Args, "success": err == nil, "output": regionalSafeOutput(output), "vendorCalls": 0, "kind": "controlled-synthetic-fixture"})
	if err != nil {
		t.Fatalf("controlled native capture: %v %s", err, output)
	}
	paths, err := filepath.Glob(filepath.Join(base, "*", minimaxlive.ReportFile))
	if err != nil || len(paths) != 1 {
		t.Fatal("one controlled native capture required")
	}
	dir := filepath.Dir(paths[0])
	r, err := minimaxlive.Verify(dir, false)
	if err != nil || !r.Complete || len(r.Calls) != 14 || r.Mode != "controlled" {
		t.Fatalf("controlled fixture did not verify: %v", err)
	}
	return miniMaxAuthorityFixture{dir: dir, git: miniMaxEnvValue(cmd.Env, "GIT_DIR"), report: r}
}

// This is a synthetic verifier-contract fixture, never actual live evidence.
// Public Client captures come from the controlled producer. Only test metadata
// is attributed to a synthetic formal grant in an owned private Git namespace.
func miniMaxAuthorityFormalFixture(t *testing.T, source miniMaxAuthorityFixture, binary string, stages ...string) miniMaxAuthorityFixture {
	t.Helper()
	stage := "formal"
	if len(stages) > 0 {
		stage = stages[0]
	}
	f := source
	f.dir = filepath.Join(t.TempDir(), source.report.Session)
	if err := os.CopyFS(f.dir, os.DirFS(source.dir)); err != nil {
		t.Fatal(err)
	}
	f.source = filepath.Join(t.TempDir(), "synthetic-source")
	for name := range f.report.ExecutionSource {
		raw, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(f.source, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Even the historical origin lookup, which deliberately clears Git env,
	// discovers only this fixture's private store through its own .git file.
	if err := os.WriteFile(filepath.Join(f.source, ".git"), []byte("gitdir: "+f.git+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(filepath.Dir(f.git), ".evidence", "minimax-launches", "mm-global-formal-m3-"+f.report.Identity.Route+".json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0700); err != nil {
		t.Fatal(err)
	}
	miniMaxAuthorityWrite(t, legacy, map[string]any{"state": "COMPLETED", "route": "mm-global-formal-m3-" + f.report.Identity.Route, "sourceCommit": strings.Repeat("0", 40), "exitCode": 0, "credentialsRead": true, "launcherAudit": map[string]bool{"passed": true}, "origin": map[string]any{"schema": 1}})
	cat := ai.BuiltinCatalog()
	if stage == "candidate" {
		cat = minimaxlive.CandidateCatalog(f.report.Identity)
	}
	hash, _ := cat.Hash()
	oldHash, oldVersion := f.report.CatalogHash, f.report.CatalogVersion
	inventory, err := minimaxlive.Inventory(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	for name := range inventory {
		path := filepath.Join(f.dir, filepath.FromSlash(name))
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		raw = bytes.ReplaceAll(raw, []byte(oldHash), []byte(hash))
		raw = bytes.ReplaceAll(raw, []byte(oldVersion), []byte(cat.Version))
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, "minimax-catalog.json"), cat)
	f.report.Mode, f.report.Stage, f.report.Kind = "live", stage, "minimax-formal"
	if stage == "candidate" {
		f.report.Kind = "minimax-candidate-only"
	}
	f.report.CatalogHash, f.report.CatalogVersion = hash, cat.Version
	binaryHash, err := livegrant.FileHash(binary)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := minimaxlive.ProcessScope(f.report.Identity, "live", stage, binaryHash)
	if err != nil {
		t.Fatal(err)
	}
	grant := livegrant.Grant{Schema: 1, ID: "synthetic-mm20-formal-" + f.report.Identity.Route + "-" + evidence.Digest([]byte(run.Dir()))[:16], AccountAlias: f.report.AccountAlias, RegionAlias: f.report.RegionAlias, ExpiresAt: time.Now().UTC().Add(time.Hour), Scope: scope}
	raw := mustMarshal(t, grant)
	claim, err := livegrant.Claim(grant, raw, f.report.Session, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	f.report.ProcessGrantSHA256 = livegrant.Hash(raw)
	path, err := livegrant.LedgerPath(claim)
	if err != nil {
		t.Fatal(err)
	}
	ledger := minimaxlive.Ledger{Path: path, Stage: stage, Session: f.report.Session, Identity: f.report.Identity, ProcessGrantSHA256: f.report.ProcessGrantSHA256}
	for i := range f.report.Calls {
		call := &f.report.Calls[i]
		call.Reservation, err = ledger.Reserve(call.ID, call.Reservation.Output, call.Reservation.Input)
		if err != nil {
			t.Fatal(err)
		}
		for j := range call.Attempts {
			call.Attempts[j], err = ledger.Attempt(call.ID)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	rows, err := minimaxlive.ReadLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	f.report.FinishedAt = time.Now().UTC()
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, "minimax-ledger.json"), minimaxlive.LedgerEvidence{Receipts: rows, Predecessors: map[string]map[string][]byte{}})
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, "process-grant.json"), grant)
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, "process-authorization-receipt.json"), claim)
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, minimaxlive.ReportFile), f.report)
	var manifest map[string]any
	data, err := os.ReadFile(filepath.Join(f.dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, data, &manifest)
	versions := manifest["versions"].(map[string]any)
	versions["minimax_mode"], versions["minimax_kind"] = "live", f.report.Kind
	miniMaxAuthorityWrite(t, filepath.Join(f.dir, "manifest.json"), manifest)
	miniMaxAuthoritySeal(t, f.dir)
	return f
}

func miniMaxAuthorityWrite(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.WriteFile(path, mustMarshal(t, value), 0600); err != nil {
		t.Fatal(err)
	}
}

func miniMaxAuthoritySeal(t *testing.T, dir string) {
	t.Helper()
	files, err := minimaxlive.Inventory(dir)
	if err != nil {
		t.Fatal(err)
	}
	miniMaxAuthorityWrite(t, filepath.Join(dir, minimaxlive.SealFile), minimaxlive.Seal{Schema: 1, Files: files})
}

func miniMaxAuthorityCLI(t *testing.T, ev *evidence.Case, binary string, fixture miniMaxAuthorityFixture, args ...string) (int, string) {
	t.Helper()
	cmd := exec.CommandContext(ctxFor(t), binary, "-inspect-bundle", fixture.dir, "-mode", "live", "-stage", "formal")
	cmd.Dir, cmd.Env = fixture.source, append(xaiProcessEnv(t), "GIT_DIR="+fixture.git, "GIT_WORK_TREE="+fixture.source)
	output, err := cmd.CombinedOutput()
	if cmd.ProcessState == nil {
		t.Fatalf("qualification CLI did not start: %v", err)
	}
	exit := cmd.ProcessState.ExitCode()
	ev.Record("consumer-process", map[string]any{"argv": cmd.Args, "dir": cmd.Dir, "exitCode": exit, "output": regionalSafeOutput(output), "syntheticContractOnly": true, "vendorCalls": 0})
	return exit, string(output)
}

// The store is created by miniMaxFreshLedgerCommand under t.TempDir. Preserve
// synthetic guards and ledger bytes before cleanup; never inspect real claims.
func miniMaxAuthorityRecordStore(t *testing.T, ev *evidence.Case, git string) {
	t.Helper()
	store := filepath.Join(git, "barness-live-process-grants")
	entries, err := os.ReadDir(store)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			t.Fatal("unexpected synthetic authority entry")
		}
		raw, err := os.ReadFile(filepath.Join(store, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name()] = raw
	}
	ev.Record("synthetic-private-authority", map[string]any{"commonGitDir": git, "files": files, "actualLive": false, "vendorCalls": 0})
}
