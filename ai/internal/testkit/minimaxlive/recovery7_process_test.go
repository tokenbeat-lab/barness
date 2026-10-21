package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Public failure ways precede this fixed-round implementation: historical6
// drift, missing/newly relabelled approval, incomplete first18 proof, changed
// source/guard, widened caps and consumed claims. One synthetic owner retains
// every predecessor once; mutations are restored before the next CLI process.
func TestRecovery7AuthorityProcess(t *testing.T) {
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
	for n, raw := range map[string]string{"go.mod": "module recovery7-fixture\nrequire github.com/openai/openai-go/v3 v3.0.0\n", "go.sum": "", "ai/runtime.go": "package ai\n", "ai/internal/testkit/finalbudget/guard.go": "package finalbudget\n", "ai/internal/testkit/finalbudget/estimate.go": "package finalbudget\n"} {
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
	c = exec.Command(exe, "-test.run=^TestRecovery7AuthorityChild$", "-test.v")
	c.Dir = dir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY7_CHILD=synthetic", "BARNESS_RECOVERY7_CLI=" + cli}
	if os.Getenv("BARNESS_RECOVERY7_FACTS_ONLY") != "" {
		c.Env = append(c.Env, "BARNESS_RECOVERY7_FACTS_ONLY=1")
	}
	if os.Getenv("BARNESS_RECOVERY7_DURATION_ONLY") != "" {
		c.Env = append(c.Env, "BARNESS_RECOVERY7_DURATION_ONLY=1")
	}
	if os.Getenv("BARNESS_RECOVERY7_SELECTION_ONLY") != "" {
		c.Env = append(c.Env, "BARNESS_RECOVERY7_SELECTION_ONLY=1")
	}
	raw, e := c.CombinedOutput()
	t.Log(string(raw))
	if e != nil {
		t.Fatalf("synthetic current7 process: %v", e)
	}
}

func TestRecovery7AuthorityChild(t *testing.T) {
	if os.Getenv("BARNESS_RECOVERY7_CHILD") == "" {
		t.Skip("synthetic process only")
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	f.seedSecondCohort(t)
	f.seedThirdCohort(t)
	f.seedFourthCohort(t)
	f.seedFifthCohort(t)
	f.seedSixthCohort(t)
	cli := os.Getenv("BARNESS_RECOVERY7_CLI")
	command := func(flag string) ([]byte, error) {
		c := exec.Command(cli, flag)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		return c.CombinedOutput()
	}
	out, e := command("-recovery7-source")
	if e != nil {
		t.Fatalf("facts %v %s", e, out)
	}
	var facts map[string]any
	if json.Unmarshal(out, &facts) != nil || facts["renewalId"] != recovery7ID {
		t.Fatalf("facts identity %s", out)
	}
	// Independent dictionary encoding matches the controller's sorted compact
	// JSON contract, rather than copying the Go CLI's supplied history pin.
	canonical, e := json.Marshal(facts["sixthCohort"])
	if e != nil || minimaxlive.Hash(canonical) != facts["sixthCohortHistorySha256"] {
		t.Fatalf("public sixth history digest differs from independent sorted dictionary: %v got=%v want=%s", e, facts["sixthCohortHistorySha256"], minimaxlive.Hash(canonical))
	}
	if os.Getenv("BARNESS_RECOVERY7_FACTS_ONLY") != "" {
		return
	}
	for _, name := range []string{"unexpected.json", "unexpected-link", "unexpected-fifo", "unexpected-directory-alias"} {
		t.Run("closed-selection-"+name, func(t *testing.T) {
			p := filepath.Join(f.root, sixthOut+name)
			defer os.Remove(p)
			if name == "unexpected-fifo" {
				if e := syscall.Mkfifo(p, 0600); e != nil {
					t.Fatal(e)
				}
			} else if name == "unexpected-directory-alias" {
				os.Symlink(filepath.Join(f.root, sixthOut), p)
			} else if name == "unexpected-link" {
				os.Symlink(filepath.Join(f.root, sixthOut+"not-opened"), p)
			} else {
				os.WriteFile(p, []byte("unselected"), 0600)
			}
			if raw, e := command("-recovery7-source"); e == nil {
				t.Fatalf("unselected closed sixth entry admitted %s", raw)
			}
		})
	}
	if os.Getenv("BARNESS_RECOVERY7_SELECTION_ONLY") != "" {
		return
	}
	if _, e = command("-recovery7-preflight=chat"); e == nil {
		t.Fatal("missing actual7 authorization admitted")
	}
	f.installRecovery7Grant(t, facts)
	if out, e = command("-recovery7-preflight=chat"); e != nil {
		t.Fatalf("valid current7 preflight %v %s", e, out)
	}
	f.verifySeventhLifetimeBoundary(t)
	if os.Getenv("BARNESS_RECOVERY7_DURATION_ONLY") != "" {
		return
	}
	if _, e = os.Lstat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
		t.Fatal("preflight consumed nonce")
	}
	if _, _, e = minimaxlive.RenewalAuthority(recovery6ID, true); e == nil {
		t.Fatal("closed6 active entry reopened")
	}
	g, _, e := minimaxlive.RenewalAuthority(recovery6ID, false)
	if e != nil || g.Schema != 7 || g.SourceCommit == f.commit {
		t.Fatalf("closed6 historical source should remain independently readable: %v", e)
	}
	// Restore each exact original byte string after the refusal. The actual
	// canonical repository is never an input to this synthetic owner.
	for _, tc := range []struct {
		name, path, key string
		value           any
	}{
		{"old-actual-approval", ".evidence/provider-sites/final-acceptance-20261020/recovery6-authorization.json", "decision", "PENDING"},
		{"old-UNKNOWN", ".evidence/minimax-launches/" + recovery6ID + "-mm-global-formal-m3-chat.json", "budget", map[string]any{"state": "REPORTED", "status": "PASS", "calls": 4, "attempts": 4}},
		{"old-guard-zero", sixthOut + "mm-global-formal-m3-chat-usd.json", "reservedMicrousd", 0},
		{"old-tail", ".evidence/minimax-renewals/" + recovery6ID + "/chat.json", "processNonce", strings.Repeat("0", 64)},
		{"old-seal", sixthClosure, "claimedHTTPAttempts", 0},
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
			if raw, e := command("-recovery7-source"); e == nil {
				t.Fatalf("changed sixth history accepted %s", raw)
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
			if raw, e := command("-recovery7-preflight=chat"); e == nil {
				t.Fatalf("invalid7 grant accepted %s", raw)
			}
		})
	}
	for _, key := range []string{"exitCode", "credentialsRead"} {
		t.Run("first-consumer-missing-"+key, func(t *testing.T) {
			p := filepath.Join(f.root, seventhOut+"formal-zai-responses-consume-command.json")
			before := mustReadFixture(t, p)
			defer os.WriteFile(p, before, 0600)
			var x map[string]any
			json.Unmarshal(before, &x)
			delete(x, key)
			b, _ := json.Marshal(x)
			os.WriteFile(p, b, 0600)
			if raw, e := command("-recovery7-preflight=chat"); e == nil {
				t.Fatalf("incomplete current first consumer accepted %s", raw)
			}
		})
	}
	t.Run("new-ref-required", func(t *testing.T) {
		paths := []string{filepath.Join(f.base, "human-approval.json"), filepath.Join(f.base, "source-review.json"), filepath.Join(f.base, "grant.json"), filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery7-authorization.json")}
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
			d["requestReference"] = "synthetic-private-new6-no-paid"
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
		if raw, e := command("-recovery7-preflight=chat"); e == nil {
			t.Fatalf("old6 response relabelled into7: %s", raw)
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
			if raw, e := command("-recovery7-preflight=chat"); e == nil {
				t.Fatalf("ambiguous human accepted %s", raw)
			}
		})
	}
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recovery7ID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "seventh-chat", Identity: id}
	if e = l.BeginRenewal(); e != nil {
		t.Fatal(e)
	}
	if e = l.BeginRenewal(); e == nil {
		t.Fatal("duplicate7 nonce admitted")
	}
	if _, e = command("-recovery7-preflight=chat"); e == nil {
		t.Fatal("consumed7 route admitted by preflight")
	}
	if _, e = l.Reserve("full-stream", 1024, 2048); e != nil {
		t.Fatal(e)
	}
	if _, e = l.Attempt("full-stream"); e != nil {
		t.Fatal(e)
	}
	rows, e := minimaxlive.ReadLedger(path)
	if e != nil || len(rows) != 2 || rows[0].RenewalID != recovery7ID || rows[0].RenewalGrantSHA256 != f.grantPin {
		t.Fatalf("new7 ledger authority: %v", e)
	}
	os.Remove(path)
	if _, e = l.Reserve("full-complete", 1024, 2048); e == nil {
		t.Fatal("deleted7 ledger reset consumed nonce")
	}
	f.checkHistory(t)
	fmt.Println("PASS fixed7 scope/closed6/current-first18/new-human/once-only nonce; synthetic only")
}
