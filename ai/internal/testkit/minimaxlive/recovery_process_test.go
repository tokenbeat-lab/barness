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

const recoveryID = "provider-sites-final-acceptance-3"
const secondClosure = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-2/seal.json"
const stoppedIndex = ".evidence/provider-sites/final-acceptance-20261020/remaining-stopped/index.json"

// Failure ways are established before the implementation: absent or stale
// authority, forged closed history, reused process identity, exhaustion, and
// preflight acquiring execution rights. Every fixture is an independent owner.
func TestRecoveryAuthorityProcess(t *testing.T) {
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
	for _, mode := range strings.Fields("human-duplicate review-duplicate human-case-alias review-case-alias facts valid-three pending missing human-missing human-pending human-scope human-old human-source policy review-source source owner unknown-grant seal stop-pin old-ledger old-tail failed-report parent-unknown old2-inactive old2-active old2-misuse duplicate concurrent preflight preflight-no-grant preflight-claimed caps output input cancellation fifo linked") {
		t.Run(mode, func(t *testing.T) {
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
			cmd := exec.Command(exe, "-test.run=^TestRecoveryAuthorityChild$", "-test.v")
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY_CHILD=" + mode, "BARNESS_RECOVERY_CLI=" + cli}
			out, e := cmd.CombinedOutput()
			if e != nil {
				t.Fatalf("%s: %v\n%s", mode, e, out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}

func TestRecoveryAuthorityChild(t *testing.T) {
	mode := os.Getenv("BARNESS_RECOVERY_CHILD")
	if mode == "" {
		t.Skip("synthetic process only")
	}
	if mode == "claim-child" {
		recoveryClaimChild(t)
		return
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	f.seedSecondCohort(t)
	cli := os.Getenv("BARNESS_RECOVERY_CLI")
	command := func(flag string) ([]byte, error) {
		c := exec.Command(cli, flag)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		return c.CombinedOutput()
	}
	out, e := command("-recovery-source")
	if e != nil {
		t.Fatalf("recovery-source: %v %s", e, out)
	}
	var facts map[string]any
	if json.Unmarshal(out, &facts) != nil || facts["renewalId"] != recoveryID {
		t.Fatalf("wrong recovery facts %s", out)
	}
	if mode == "facts" {
		fmt.Println("PASS keyless recovery facts retain original39/36 twice and failed13/12")
		return
	}
	f.installRecoveryGrant(t, facts)
	if recoveryMutation(t, f, mode) {
		if _, e := command("-recovery-preflight=chat"); e == nil {
			t.Fatal("invalid recovery preflight accepted", mode)
		}
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recoveryID)
		if _, e := minimaxlive.StageLedgerPath("formal"); e == nil {
			t.Fatal("invalid recovery admitted", mode)
		}
		fmt.Println("REJECTED", mode)
		return
	}
	if mode == "old2-active" {
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", finalRenewalID)
		if _, e := minimaxlive.StageLedgerPath("formal"); e == nil {
			t.Fatal("closed ID2 reopened")
		}
		return
	}
	if mode == "preflight" || mode == "preflight-claimed" {
		for i := 0; i < 2; i++ {
			if out, e := command("-recovery-preflight=chat"); e != nil {
				t.Fatalf("preflight: %v %s", e, out)
			}
		}
		if _, e := os.Lstat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
			t.Fatal("preflight consumed nonce")
		}
	}
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recoveryID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "recovery-chat", Identity: id}
	if mode == "concurrent" {
		recoveryConcurrent(t, f)
		return
	}
	if mode == "old2-misuse" {
		l.Path = filepath.Join(f.root, ".evidence/minimax-renewals", finalRenewalID, "formal.jsonl")
		if l.BeginRenewal() == nil {
			t.Fatal("ID3 borrowed old ledger")
		}
		return
	}
	if e = l.BeginRenewal(); e != nil {
		t.Fatal(e)
	}
	if mode == "duplicate" || mode == "preflight-claimed" || mode == "cancellation" {
		if l.BeginRenewal() == nil {
			t.Fatal("duplicate begin accepted")
		}
		if _, e := command("-recovery-preflight=chat"); e == nil {
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
		for _, p := range minimaxlive.Plan(id) {
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
		if r.RenewalID != recoveryID {
			t.Fatal("old round borrowed")
		}
	}
	f.checkHistory(t)
	fmt.Println("PASS own recovery39/36 with closed failed ID2 untouched")
}

func recoveryClaimChild(t *testing.T) {
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recoveryID)
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
func recoveryConcurrent(t *testing.T, f *finalRenewalFixture) {
	exe, _ := os.Executable()
	commands := []*exec.Cmd{}
	for i := 0; i < 2; i++ {
		c := exec.Command(exe, "-test.run=^TestRecoveryAuthorityChild$")
		c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RECOVERY_CHILD=claim-child"}
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
