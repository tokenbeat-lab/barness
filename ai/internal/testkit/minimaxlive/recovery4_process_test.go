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

const recovery4ID = "provider-sites-final-acceptance-4"

// These public fixtures deliberately keep recovery3 constructors intact so
// legacy authorization can regress independently from the new fixed boundary.
// Failure ways are established before the implementation: absent or stale
// authority, forged closed history, reused process identity, exhaustion, and
// preflight acquiring execution rights. Every fixture is an independent owner.
func TestRecovery4AuthorityProcess(t *testing.T) {
	t.Parallel()
	root, err := minimaxlive.Root()
	if err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "minimaxverify")
	c := exec.Command("go", "build", "-o", cli, "./ai/live/cmd/minimaxverify")
	c.Dir = root
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	for _, mode := range strings.Fields("facts valid-three historical-source human-duplicate review-duplicate human-case-alias review-case-alias human-nested-duplicate pending missing human-missing human-pending human-scope human-old human-source policy review-source source owner unknown-grant seal stop-pin old-ledger old-tail failed-report passed-report parent-unknown old3-inactive old3-active old3-misuse duplicate concurrent preflight preflight-no-grant preflight-claimed caps output input cancellation fifo linked oversize old-source-commit archive human-record counters forged-pass old1-third old2-third old3-third unstarted-claim missing-nonce linked-worktree same-session-process human-reference") {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			admitRenewalOwner(t)
			dir := t.TempDir()
			for name, raw := range map[string]string{"go.mod": "module recovery-fixture\n", "go.sum": "", "ai/runtime.go": "package ai\n"} {
				p := filepath.Join(dir, name)
				os.MkdirAll(filepath.Dir(p), 0700)
				os.WriteFile(p, []byte(raw), 0600)
			}
			for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "go.sum", "ai"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "source"}} {
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				if out, e := cmd.CombinedOutput(); e != nil {
					t.Fatalf("git: %v %s", e, out)
				}
			}
			exe, _ := os.Executable()
			cmd := exec.Command(exe, "-test.run=^TestRecovery4AuthorityChild$", "-test.v")
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY4_CHILD=" + mode, "BARNESS_RECOVERY4_CLI=" + cli}
			out, e := cmd.CombinedOutput()
			if e != nil {
				t.Fatalf("%s: %v\n%s", mode, e, out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}

func TestRecovery4AuthorityChild(t *testing.T) {
	mode := os.Getenv("BARNESS_RECOVERY4_CHILD")
	if mode == "" {
		t.Skip("synthetic process only")
	}
	if mode == "claim-child" {
		recovery4ClaimChild(t)
		return
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	f.seedSecondCohort(t)
	f.seedThirdCohort(t)
	cli := os.Getenv("BARNESS_RECOVERY4_CLI")
	command := func(flag string) ([]byte, error) {
		c := exec.Command(cli, flag)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		return c.CombinedOutput()
	}
	out, e := command("-recovery4-source")
	if e != nil {
		t.Fatalf("recovery-source: %v %s", e, out)
	}
	var facts map[string]any
	if json.Unmarshal(out, &facts) != nil || facts["renewalId"] != recovery4ID {
		t.Fatalf("wrong recovery facts %s", out)
	}
	if mode == "facts" {
		fmt.Println("PASS keyless recovery4 facts retain15 and closed1/2/3 with failed13/12 and own third39/36")
		return
	}
	f.installRecovery4Grant(t, facts)
	if mode == "historical-source" {
		if facts["sourceCommit"] == facts["thirdCohort"].(map[string]any)["nativeSourceCommit"] {
			t.Fatal("fixture did not change current selected source")
		}
		if out, e := command("-recovery4-preflight=chat"); e != nil {
			t.Fatalf("current authority with historical source: %v %s", e, out)
		}
		fmt.Println("PASS closed3 old source differs while current4 preflight qualifies")
		return
	}
	if recovery4Mutation(t, f, mode) {
		if _, e := command("-recovery4-preflight=chat"); e == nil {
			t.Fatal("invalid recovery preflight accepted", mode)
		}
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recovery4ID)
		if _, e := minimaxlive.StageLedgerPath("formal"); e == nil {
			t.Fatal("invalid recovery admitted", mode)
		}
		fmt.Println("REJECTED", mode)
		return
	}
	if mode == "old3-active" {
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recoveryID)
		if _, e := minimaxlive.StageLedgerPath("formal"); e == nil {
			t.Fatal("closed ID2 reopened")
		}
		return
	}
	if mode == "preflight" || mode == "preflight-claimed" {
		for i := 0; i < 2; i++ {
			if out, e := command("-recovery4-preflight=chat"); e != nil {
				t.Fatalf("preflight: %v %s", e, out)
			}
		}
		if _, e := os.Lstat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
			t.Fatal("preflight consumed nonce")
		}
		if mode == "preflight" {
			fmt.Println("PASS repeated keyless preflight acquired no nonce or call")
			return
		}
	}
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recovery4ID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "recovery-chat", Identity: id}
	if mode == "concurrent" {
		recovery4Concurrent(t, f)
		return
	}
	if mode == "old3-misuse" {
		l.Path = filepath.Join(f.root, ".evidence/minimax-renewals", recoveryID, "formal.jsonl")
		if l.BeginRenewal() == nil {
			t.Fatal("ID3 borrowed old ledger")
		}
		return
	}
	if mode == "same-session-process" {
		l.Session = "same-label"
	}
	if e = l.BeginRenewal(); e != nil {
		t.Fatal(e)
	}
	if mode == "linked-worktree" || mode == "same-session-process" {
		dir := f.root
		if mode == "linked-worktree" {
			dir = filepath.Join(t.TempDir(), "checkout")
			c := exec.Command("git", "worktree", "add", "--detach", dir, "HEAD")
			c.Dir = f.root
			if out, e := c.CombinedOutput(); e != nil {
				t.Fatalf("worktree: %v %s", e, out)
			}
		}
		exe, _ := os.Executable()
		c := exec.Command(exe, "-test.run=^TestRecovery4AuthorityChild$")
		c.Dir = dir
		c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY4_CHILD=claim-child"}
		if out, e := c.CombinedOutput(); e == nil {
			t.Fatalf("consumed route resumed from second process/worktree: %s", out)
		}
		fmt.Println("PASS consumed nonce rejects second process and common-owner worktree")
		return
	}
	if mode == "duplicate" || mode == "preflight-claimed" || mode == "cancellation" {
		if l.BeginRenewal() == nil {
			t.Fatal("duplicate begin accepted")
		}
		if _, e := command("-recovery4-preflight=chat"); e == nil {
			t.Fatal("consumed nonce preflight accepted")
		}
		if _, e := os.Lstat(path + ".lock"); !os.IsNotExist(e) {
			t.Fatal("admission failed to release lock")
		}
		if _, e := os.Lstat(path); !os.IsNotExist(e) {
			t.Fatal("begin invented a call")
		}
		fmt.Println("PASS one-use claim survives cancellation with zero calls and released lock")
		return
	}
	if mode == "output" || mode == "input" {
		output, input := 4096, 2048
		if mode == "input" {
			output, input = 1, 120000
		}
		if _, e = l.Reserve("full-stream", output, input); e != nil {
			t.Fatal(e)
		}
		if _, e = l.Reserve("full-complete", output, input); mode == "input" && e == nil {
			t.Fatal("input cap bypass")
		}
		if mode == "output" {
			if _, e = l.Reserve("full-complete", 4097, 1); e == nil {
				t.Fatal("output cap bypass")
			}
		}
		return
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		own := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "recovery-" + route, Identity: id}
		if route != "chat" {
			if e = own.BeginRenewal(); e != nil {
				t.Fatal(e)
			}
		}
		for _, p := range historicalM3Plan() {
			if _, e = own.Reserve(p.ID, p.Output, 2048); e != nil {
				t.Fatal(e)
			}
			if p.ID != "preflight-error" {
				if _, e = own.Attempt(p.ID); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	rows, e := minimaxlive.ReadLedger(path)
	if e != nil || len(rows) != 75 {
		t.Fatalf("new rows %d %v", len(rows), e)
	}
	for _, r := range rows {
		if r.RenewalID != recovery4ID {
			t.Fatal("old round borrowed")
		}
	}
	f.checkHistory(t)
	fmt.Println("PASS own recovery4 39/36 with all closed1/2/3 facts untouched")
}

func recovery4ClaimChild(t *testing.T) {
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recovery4ID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "same-label", Identity: id}
	if e = l.BeginRenewal(); e != nil {
		t.Fatal(e)
	}
}
func recovery4Concurrent(t *testing.T, f *finalRenewalFixture) {
	exe, _ := os.Executable()
	commands := []*exec.Cmd{}
	for i := 0; i < 2; i++ {
		c := exec.Command(exe, "-test.run=^TestRecovery4AuthorityChild$")
		c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY4_CHILD=claim-child"}
		commands = append(commands, c)
		if e := c.Start(); e != nil {
			t.Fatal(e)
		}
	}
	passed := 0
	for _, c := range commands {
		if c.Wait() == nil {
			passed++
		}
	}
	if passed != 1 {
		t.Fatalf("want exactly one admission got %d", passed)
	}
	if _, e := os.Lstat(filepath.Join(f.base, "formal.jsonl.lock")); !os.IsNotExist(e) {
		t.Fatal("concurrent admission leaked lock")
	}
}
