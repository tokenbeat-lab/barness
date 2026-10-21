package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Public failure ways precede this fixed-round implementation: historical5
// drift, missing/newly relabelled approval, incomplete first18 proof, changed
// source/guard, widened caps and consumed claims. One synthetic owner retains
// every predecessor once; mutations are restored before the next CLI process.
func TestRecovery6AuthorityProcess(t *testing.T) {
	t.Parallel()
	admitRenewalOwner(t)
	root, e := minimaxlive.Root()
	if e != nil {
		t.Fatal(e)
	}
	cli := filepath.Join(t.TempDir(), "minimaxverify")
	c := exec.Command("go", "build", "-o", cli, "./ai/live/cmd/minimaxverify")
	c.Dir = root
	if raw, e := c.CombinedOutput(); e != nil {
		t.Fatalf("build %v %s", e, raw)
	}
	dir := t.TempDir()
	for n, raw := range map[string]string{"go.mod": "module recovery6-fixture\n", "go.sum": "", "ai/runtime.go": "package ai\n", "ai/internal/testkit/finalbudget/guard.go": "package finalbudget\n", "ai/internal/testkit/finalbudget/estimate.go": "package finalbudget\n"} {
		p := filepath.Join(dir, n)
		os.MkdirAll(filepath.Dir(p), 0700)
		os.WriteFile(p, []byte(raw), 0600)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "go.sum", "ai"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "source"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if raw, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git %v %s", e, raw)
		}
	}
	exe, _ := os.Executable()
	c = exec.Command(exe, "-test.run=^TestRecovery6AuthorityChild$", "-test.v")
	c.Dir = dir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY6_CHILD=synthetic", "BARNESS_RECOVERY6_CLI=" + cli}
	if os.Getenv("BARNESS_RECOVERY6_SELECTION_ONLY") != "" {
		c.Env = append(c.Env, "BARNESS_RECOVERY6_SELECTION_ONLY=1")
	}
	raw, e := c.CombinedOutput()
	t.Log(string(raw))
	if e != nil {
		t.Fatalf("synthetic current6 process: %v", e)
	}
}

func TestRecovery6AuthorityChild(t *testing.T) {
	if os.Getenv("BARNESS_RECOVERY6_CHILD") == "" {
		t.Skip("synthetic process only")
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	f.seedSecondCohort(t)
	f.seedThirdCohort(t)
	f.seedFourthCohort(t)
	f.seedFifthCohort(t)
	cli := os.Getenv("BARNESS_RECOVERY6_CLI")
	command := func(flag string) ([]byte, error) {
		c := exec.Command(cli, flag)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		return c.CombinedOutput()
	}
	out, e := command("-recovery6-source")
	if e != nil {
		t.Fatalf("facts %v %s", e, out)
	}
	var facts map[string]any
	if json.Unmarshal(out, &facts) != nil || facts["renewalId"] != recovery6ID {
		t.Fatalf("facts identity %s", out)
	}
	for _, name := range []string{"unexpected.json", "unexpected-link"} {
		t.Run("closed-selection-"+name, func(t *testing.T) {
			p := filepath.Join(f.root, fifthOut+name)
			defer os.Remove(p)
			if name == "unexpected-link" {
				os.Symlink(filepath.Join(f.root, fifthOut+"not-opened"), p)
			} else {
				os.WriteFile(p, []byte("unselected"), 0600)
			}
			if raw, e := command("-recovery6-source"); e == nil {
				t.Fatalf("unselected closed fifth entry admitted %s", raw)
			}
		})
	}
	if os.Getenv("BARNESS_RECOVERY6_SELECTION_ONLY") != "" {
		return
	}
	if _, e = command("-recovery6-preflight=chat"); e == nil {
		t.Fatal("missing actual6 authorization admitted")
	}
	f.installRecovery6Grant(t, facts)
	if out, e = command("-recovery6-preflight=chat"); e != nil {
		t.Fatalf("valid current6 preflight %v %s", e, out)
	}
	if _, e = os.Lstat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
		t.Fatal("preflight consumed nonce")
	}
	if _, _, e = minimaxlive.RenewalAuthority(recovery5ID, true); e == nil {
		t.Fatal("closed5 active entry reopened")
	}
	g, _, e := minimaxlive.RenewalAuthority(recovery5ID, false)
	if e != nil || g.Schema != 6 || g.SourceCommit == f.commit {
		t.Fatalf("closed5 historical source should remain independently readable: %v", e)
	}
	// Restore each exact original byte string after the refusal. The actual
	// canonical repository is never an input to this synthetic owner.
	for _, tc := range []struct {
		name, path, key string
		value           any
	}{
		{"old-actual-approval", ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json", "decision", "PENDING"},
		{"old-failed-report", fifthOut + "formal-minimax-cn-responses/fifth-formal-minimax-cn-responses/regional-report.json", "regional", map[string]any{"mode": "live", "stage": "formal", "complete": true, "auditPassed": true}},
		{"old-UNKNOWN", ".evidence/minimax-launches/" + recovery5ID + "-formal-minimax-cn-responses.json", "budget", map[string]any{"state": "REPORTED", "status": "PASS", "calls": 9, "attempts": 9}},
		{"old-guard-zero", fifthOut + "formal-minimax-cn-responses-usd.json", "reservedMicrousd", 0},
		{"old-tail", ".evidence/minimax-renewals/" + recovery5ID + "/responses.json", "processNonce", strings.Repeat("0", 64)},
		{"old-seal", fifthClosure, "claimedHTTPAttempts", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(f.root, tc.path)
			before := mustReadFixture(t, p)
			defer os.WriteFile(p, before, 0600)
			var x map[string]any
			json.Unmarshal(before, &x)
			x[tc.key] = tc.value
			b, _ := json.Marshal(x)
			os.WriteFile(p, b, 0600)
			if raw, e := command("-recovery6-source"); e == nil {
				t.Fatalf("changed fifth history accepted %s", raw)
			}
		})
	}
	for _, tc := range []struct {
		name, key string
		value     any
	}{
		{"pending", "decision", "PENDING"}, {"old-round", "renewalId", recovery5ID}, {"budget", "maxCalls", 61}, {"source", "executionSourceSha256", strings.Repeat("0", 64)}, {"guard", "spendingGuardSourceSha256", map[string]string{}}, {"proof-missing", "firstRouteProof", map[string]string{}}, {"old-proof", "firstRouteProof", g.FirstRouteProof},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(f.base, "grant.json")
			before := mustReadFixture(t, p)
			defer os.WriteFile(p, before, 0600)
			var x map[string]any
			json.Unmarshal(before, &x)
			x[tc.key] = tc.value
			b, _ := json.Marshal(x)
			os.WriteFile(p, b, 0600)
			if raw, e := command("-recovery6-preflight=chat"); e == nil {
				t.Fatalf("invalid6 grant accepted %s", raw)
			}
		})
	}
	for _, key := range []string{"exitCode", "credentialsRead"} {
		t.Run("first-consumer-missing-"+key, func(t *testing.T) {
			p := filepath.Join(f.root, sixthOut+"formal-zai-responses-consume-command.json")
			before := mustReadFixture(t, p)
			defer os.WriteFile(p, before, 0600)
			var x map[string]any
			json.Unmarshal(before, &x)
			delete(x, key)
			b, _ := json.Marshal(x)
			os.WriteFile(p, b, 0600)
			if raw, e := command("-recovery6-preflight=chat"); e == nil {
				t.Fatalf("incomplete current first consumer accepted %s", raw)
			}
		})
	}
	t.Run("new-ref-required", func(t *testing.T) {
		paths := []string{filepath.Join(f.base, "human-approval.json"), filepath.Join(f.base, "source-review.json"), filepath.Join(f.base, "grant.json"), filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery6-authorization.json")}
		original := map[string][]byte{}
		for _, p := range paths {
			original[p] = mustReadFixture(t, p)
		}
		defer func() {
			for p, b := range original {
				os.WriteFile(p, b, 0600)
			}
		}()
		docs := []map[string]any{}
		for _, p := range paths {
			var d map[string]any
			json.Unmarshal(original[p], &d)
			d["requestReference"] = "synthetic-private-new5-no-paid"
			docs = append(docs, d)
		}
		b, _ := json.Marshal(docs[3])
		os.WriteFile(paths[3], b, 0600)
		docs[0]["actualHumanApprovalSHA256"] = minimaxlive.Hash(b)
		for i := 0; i < 2; i++ {
			b, _ = json.Marshal(docs[i])
			os.WriteFile(paths[i], b, 0600)
		}
		docs[2]["humanApprovalSha256"] = minimaxlive.Hash(mustReadFixture(t, paths[0]))
		docs[2]["sourceReviewSha256"] = minimaxlive.Hash(mustReadFixture(t, paths[1]))
		b, _ = json.Marshal(docs[2])
		os.WriteFile(paths[2], b, 0600)
		if raw, e := command("-recovery6-preflight=chat"); e == nil {
			t.Fatalf("old5 response relabelled into6: %s", raw)
		}
	})
	for _, kind := range []string{"duplicate", "alias"} {
		t.Run("human-"+kind, func(t *testing.T) {
			p := filepath.Join(f.base, "human-approval.json")
			before := mustReadFixture(t, p)
			defer os.WriteFile(p, before, 0600)
			b := append([]byte(`{"decision":"PENDING",`), before[1:]...)
			if kind == "alias" {
				b = append(append([]byte(nil), before[:len(before)-1]...), []byte(`,"Decision":"APPROVED"}`)...)
			}
			os.WriteFile(p, b, 0600)
			grant := filepath.Join(f.base, "grant.json")
			old := mustReadFixture(t, grant)
			defer os.WriteFile(grant, old, 0600)
			var x map[string]any
			json.Unmarshal(old, &x)
			x["humanApprovalSha256"] = minimaxlive.Hash(b)
			b, _ = json.Marshal(x)
			os.WriteFile(grant, b, 0600)
			if raw, e := command("-recovery6-preflight=chat"); e == nil {
				t.Fatalf("ambiguous human accepted %s", raw)
			}
		})
	}
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recovery6ID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "sixth-chat", Identity: id}
	if e = l.BeginRenewal(); e != nil {
		t.Fatal(e)
	}
	if e = l.BeginRenewal(); e == nil {
		t.Fatal("duplicate6 nonce admitted")
	}
	if _, e = command("-recovery6-preflight=chat"); e == nil {
		t.Fatal("consumed6 route admitted by preflight")
	}
	if _, e = l.Reserve("full-stream", 1024, 2048); e != nil {
		t.Fatal(e)
	}
	if _, e = l.Attempt("full-stream"); e != nil {
		t.Fatal(e)
	}
	rows, e := minimaxlive.ReadLedger(path)
	if e != nil || len(rows) != 2 || rows[0].RenewalID != recovery6ID || rows[0].RenewalGrantSHA256 != f.grantPin {
		t.Fatalf("new6 ledger authority: %v", e)
	}
	os.Remove(path)
	if _, e = l.Reserve("full-complete", 1024, 2048); e == nil {
		t.Fatal("deleted6 ledger reset consumed nonce")
	}
	f.checkHistory(t)
	fmt.Println("PASS fixed6 scope/closed5/current-first18/new-human/once-only nonce; synthetic only")
}
