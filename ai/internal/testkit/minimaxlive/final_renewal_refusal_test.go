package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func finalRenewalRejectedMode(mode string) bool {
	for _, m := range []string{"missing", "pending", "history-missing", "history-added", "history-modified", "history-linked", "history-parent-linked", "history-launcher", "history-open", "human-missing", "human-pending", "human-plan", "human-history", "human-purpose", "human-routes", "caps", "wrong-id", "candidate", "source-stale", "source-uncommitted", "foreign-owner"} {
		if mode == m {
			return true
		}
	}
	return false
}

func finalRenewalVariant(t *testing.T, f *finalRenewalFixture, mode string) bool {
	t.Helper()
	legacy := filepath.Join(f.root, ".evidence/minimax-renewals", renewalID)
	mutate := func(path string, key string, value any) {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var d map[string]any
		if e = json.Unmarshal(b, &d); e != nil {
			t.Fatal(e)
		}
		d[key] = value
		b, _ = json.Marshal(d)
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	switch mode {
	case "budget-source", "regional-price-source":
		path := filepath.Join(f.root, "ai/internal/testkit/finalbudget/guard.go")
		if mode == "regional-price-source" {
			path = filepath.Join(f.root, "ai/internal/testkit/regional/prices.json")
		}
		if e := os.WriteFile(path, []byte("package finalbudget\n// changed budget policy\n"), 0600); e != nil {
			t.Fatal(e)
		}
		if _, _, e := minimaxlive.RenewalAuthority(finalRenewalID, true); e == nil {
			t.Fatal("changed spend guard retained execution authority")
		}
		fmt.Println("REJECTED changed spend guard execution source")
		return true
	case "facts":
		facts, e := minimaxlive.RenewalFacts()
		if e != nil || facts.RenewalID != finalRenewalID || facts.PreviousRenewalID != renewalID || facts.OriginalCalls != 39 || facts.OriginalAttempts != 36 || facts.PreviousCalls != 39 || facts.PreviousAttempts != 36 || len(facts.ClosedHistory) != 15 || facts.ClosedHistorySHA256 != f.grant["closedHistorySha256"] {
			t.Fatalf("current fixed native facts incomplete: %+v error=%v", facts, e)
		}
		for name, raw := range f.history {
			if facts.ClosedHistory[name] != minimaxlive.Hash(raw) {
				t.Fatalf("native predecessor selection differs: %s", name)
			}
		}
		fmt.Println("PASS current fixed ID with all15 canonical predecessor pins")
		return true
	case "closed-active":
		// Deleting the old close marker cannot reactivate an old grant.
		os.Remove(filepath.Join(legacy, "finalize-intent.json"))
		if _, _, e := minimaxlive.RenewalAuthority(renewalID, true); e == nil {
			t.Fatal("closed original renewal reacquired execution authority")
		}
		fmt.Println("REJECTED permanently closed original renewal")
		return true
	case "missing":
		os.Remove(filepath.Join(f.base, "grant.json"))
	case "pending":
		f.grant["decision"] = "PENDING"
		f.write(t, "grant.json", f.grant)
	case "history-missing":
		os.Remove(filepath.Join(legacy, "chat.json"))
	case "history-added":
		os.WriteFile(filepath.Join(legacy, "unexpected.json"), []byte("{}"), 0600)
	case "history-modified":
		b, _ := os.ReadFile(filepath.Join(legacy, "formal.jsonl"))
		os.WriteFile(filepath.Join(legacy, "formal.jsonl"), append(b, '\n'), 0600)
	case "history-linked":
		p := filepath.Join(legacy, "source-review.json")
		b, _ := os.ReadFile(p)
		os.WriteFile(filepath.Join(f.root, "copied-review.json"), b, 0600)
		os.Remove(p)
		if e := os.Symlink(filepath.Join(f.root, "copied-review.json"), p); e != nil {
			t.Fatal(e)
		}
	case "history-parent-linked":
		p := filepath.Join(f.root, ".evidence/minimax-stages")
		copy := filepath.Join(f.root, "copied-stage")
		if e := os.Rename(p, copy); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(copy, p); e != nil {
			t.Fatal(e)
		}
	case "history-launcher":
		p := filepath.Join(f.root, ".evidence/minimax-launches/mm-global-formal-m3-chat.json")
		mutate(p, "state", "PREPARED")
	case "history-open":
		mutate(filepath.Join(legacy, "finalize-intent.json"), "state", "PREPARED")
	case "human-missing":
		os.Remove(filepath.Join(f.base, "human-approval.json"))
	case "human-pending", "human-plan", "human-history", "human-purpose", "human-routes":
		key, value := "decision", any("PENDING")
		switch mode {
		case "human-plan":
			key, value = "planSHA256", strings.Repeat("d", 64)
		case "human-history":
			key, value = "historySHA256", strings.Repeat("d", 64)
		case "human-purpose":
			key, value = "purpose", "mplan"
		case "human-routes":
			key, value = "routes", []string{"chat", "responses"}
		}
		p := filepath.Join(f.base, "human-approval.json")
		mutate(p, key, value)
		b, _ := os.ReadFile(p)
		f.grant["humanApprovalSha256"] = minimaxlive.Hash(b)
		f.write(t, "grant.json", f.grant)
	case "caps":
		f.grant["maxCalls"] = 61
		f.write(t, "grant.json", f.grant)
	case "source-stale":
		f.grant["executionSourceSha256"] = strings.Repeat("d", 64)
		f.write(t, "grant.json", f.grant)
	case "source-uncommitted":
		os.WriteFile(filepath.Join(f.root, "ai/runtime.go"), []byte("package ai\n// uncommitted\n"), 0600)
		s, _ := minimaxlive.ExecutionSource()
		b, _ := json.Marshal(s)
		f.grant["executionSourceSha256"] = minimaxlive.Hash(b)
		f.write(t, "grant.json", f.grant)
	case "source-option":
		f.grant["sourceCommit"] = "--output=" + filepath.Join(f.root, "option-output")
		f.write(t, "grant.json", f.grant)
		if _, _, e := minimaxlive.RenewalAuthority(finalRenewalID, false); e == nil {
			t.Fatal("malformed commit accepted")
		}
		if _, e := os.Stat(filepath.Join(f.root, "option-output")); !os.IsNotExist(e) {
			t.Fatal("commit option wrote output")
		}
		return true
	case "foreign-owner":
		f.grant["commonRootSha256"] = strings.Repeat("d", 64)
		f.write(t, "grant.json", f.grant)
	case "wrong-id":
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", "invented-round")
		if _, e := minimaxlive.StageLedgerPath("formal"); e == nil {
			t.Fatal("caller-selected new round accepted")
		}
		return true
	case "candidate":
		os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", finalRenewalID)
		if _, e := minimaxlive.StageLedgerPath("candidate"); e == nil {
			t.Fatal("candidate borrowed final formal authority")
		}
		return true
	}
	return false
}

func testFinalRenewalConsumption(t *testing.T, f *finalRenewalFixture, path, mode string) {
	t.Helper()
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "final-chat", Identity: id}
	if mode == "preflight-purpose" {
		l.Identity.Purpose = "mplan"
		begin, ok := any(l).(interface{ BeginRenewal() error })
		if !ok || begin.BeginRenewal() == nil {
			t.Fatal("nonstandard identity acquired a final route claim")
		}
		if _, e := os.Stat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
			t.Fatal("rejected identity consumed authorized standard route")
		}
		return
	}
	if mode == "preflight-claim" || mode == "preflight-crash" || mode == "launcher-preflight" {
		if mode == "launcher-preflight" {
			for range 2 {
				g, pin, e := minimaxlive.RenewalPreflight("chat")
				if e != nil || g.RenewalID != finalRenewalID || pin != f.grantPin {
					t.Fatalf("read-only launcher guard rejected current unclaimed route: %v", e)
				}
			}
			if _, e := os.Stat(filepath.Join(f.base, "chat.json")); !os.IsNotExist(e) {
				t.Fatal("read-only launcher guard consumed child's claim")
			}
		}
		begin, ok := any(l).(interface{ BeginRenewal() error })
		if !ok {
			t.Fatal("Host cannot claim the approved route before reading its credential")
		}
		if e := begin.BeginRenewal(); e != nil {
			t.Fatal(e)
		}
		if _, e := os.Stat(path); !os.IsNotExist(e) {
			t.Fatal("preflight claim invented a Logical Call")
		}
		if e := begin.BeginRenewal(); e == nil {
			t.Fatal("route preflight claimed twice")
		}
		if mode == "launcher-preflight" {
			if _, _, e := minimaxlive.RenewalPreflight("chat"); e == nil {
				t.Fatal("launcher guard accepted an already claimed route")
			}
			if _, _, e := minimaxlive.RenewalPreflight("responses"); e != nil {
				t.Fatalf("independent unclaimed response route rejected: %v", e)
			}
		}
		if mode == "preflight-crash" {
			testFinalRenewalContenders(t, f.root, "same-session-process")
			return
		}
		if _, e := l.Reserve("full-stream", 1024, 2048); e != nil {
			t.Fatalf("claiming process cannot make its first reservation: %v", e)
		}
		if _, e := l.Attempt("full-stream"); e != nil {
			t.Fatal(e)
		}
		os.Remove(path)
		if _, e := l.Reserve("full-complete", 1024, 2048); e == nil {
			t.Fatal("consumed pending claim revived after ledger deletion")
		}
		f.checkHistory(t)
		fmt.Println("PASS pre-secret claim consumed exactly once, deletion remains closed")
		return
	}
	name, output, input := "full-stream", 1024, 2048
	switch mode {
	case "arbitrary-path":
		l.Path = filepath.Join(f.root, "fresh.jsonl")
	case "unregistered":
		name = "invented-case"
	case "duplicate":
		name, output = "preflight-error", 1
	case "excess-output":
		output = 4097
	case "excess-input":
		input = 120001
	case "partial-claim":
		os.WriteFile(filepath.Join(f.base, "chat.json"), []byte(`{"schema":`), 0600)
	case "concurrent":
		testFinalRenewalContenders(t, f.root, mode)
		f.checkHistory(t)
		return
	}
	_, e := l.Reserve(name, output, input)
	for _, m := range []string{"arbitrary-path", "unregistered", "excess-output", "excess-input", "partial-claim"} {
		if mode == m {
			if e == nil {
				t.Fatalf("invalid reservation accepted: %s", mode)
			}
			fmt.Println("REJECTED", mode, e)
			return
		}
	}
	if e != nil {
		t.Fatal(e)
	}
	switch mode {
	case "new-session":
		l.Session = "new-process"
	case "deleted-ledger":
		os.Remove(path)
	case "partial-ledger":
		raw, _ := os.ReadFile(path)
		os.WriteFile(path, append(raw, []byte(`{"sequence":`)...), 0600)
	case "changed-grant":
		f.grant["planSha256"] = strings.Repeat("d", 64)
		f.write(t, "grant.json", f.grant)
	case "duplicate":
		_, e = l.Reserve(name, output, input)
		if e == nil {
			t.Fatal("unregistered duplicate accepted")
		}
		return
	case "linked-worktree", "same-session-process":
		testFinalRenewalContenders(t, f.root, mode)
		f.checkHistory(t)
		return
	}
	_, e = l.Reserve("full-complete", 1024, 2048)
	if e == nil {
		t.Fatalf("reused/restarted final authority: %s", mode)
	}
	f.checkHistory(t)
	fmt.Println("REJECTED", mode, e)
}
