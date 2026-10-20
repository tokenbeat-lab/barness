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

const renewalID = "current-source-eleven-renewal-1"

// Auxiliary public authority/process seam. The fixture is a separate Git owner;
// its grant is synthetic and can never authorize the real repository.
func TestFormalRenewalAuthorityProcess(t *testing.T) {
	for _, mode := range []string{"inactive-option", "data-only-commit", "uncommitted-source", "read-only", "pending", "human-missing", "human-pending", "source-stale", "source-commit", "old-drift", "caps", "foreign-owner", "unknown-new-field"} {
		t.Run(mode, func(t *testing.T) { testRenewalFixture(t, mode) })
	}
}
func testRenewalFixture(t *testing.T, mode string) {
	t.Helper()
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "ai"), 0700)
	for n, raw := range map[string]string{"go.mod": "module renewal-fixture\n", "go.sum": "", "ai/runtime.go": "package ai\n"} {
		if e := os.WriteFile(filepath.Join(root, n), []byte(raw), 0600); e != nil {
			t.Fatal(e)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "go.sum", "ai/runtime.go"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "source"}} {
		c := exec.Command("git", args...)
		c.Dir = root
		if o, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git %v %s", e, o)
		}
	}
	exe, _ := os.Executable()
	c := exec.Command(exe, "-test.run=^TestFormalRenewalAuthorityChild$", "-test.v")
	c.Dir = root
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_RENEWAL_CHILD=" + mode}
	o, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("mode=%s %v\n%s", mode, e, o)
	}
	t.Log(strings.TrimSpace(string(o)))
}

func TestFormalRenewalAuthorityChild(t *testing.T) {
	mode := os.Getenv("BARNESS_RENEWAL_CHILD")
	if mode == "" {
		t.Skip("private process only")
	}
	root, _ := os.Getwd()
	root, _ = filepath.EvalSymlinks(root)
	git := exec.Command("git", "rev-parse", "HEAD")
	raw, e := git.Output()
	if e != nil {
		t.Fatal(e)
	}
	commit := strings.TrimSpace(string(raw))
	old, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		l := minimaxlive.Ledger{Path: old, Stage: "formal", Session: "historical-" + route, Identity: id}
		for _, p := range minimaxlive.Plan(id) {
			if _, e = l.Reserve(p.ID, p.Output, 2048); e != nil {
				t.Fatal(e)
			}
			if p.ID != "preflight-error" {
				if _, e = l.Attempt(p.ID); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	original, _ := os.ReadFile(old)
	receipts, _ := minimaxlive.ReadLedger(old)
	source := map[string]string{}
	for _, n := range []string{"go.mod", "go.sum", "ai/runtime.go"} {
		b, _ := os.ReadFile(n)
		source[n] = minimaxlive.Hash(b)
	}
	sourceRaw, _ := json.Marshal(source)
	base := filepath.Join(root, ".evidence/minimax-renewals", renewalID)
	os.MkdirAll(base, 0700)
	write := func(n string, v any) []byte {
		b, _ := json.Marshal(v)
		if e := os.WriteFile(filepath.Join(base, n), b, 0600); e != nil {
			t.Fatal(e)
		}
		return b
	}
	human := map[string]any{"decision": "APPROVED", "renewalId": renewalID, "maxCalls": 216, "maxAttempts": 252, "requestReference": "synthetic-private-no-paid"}
	review := map[string]any{"decision": "APPROVED", "sourceCommit": commit, "sourceManifestSHA256": strings.Repeat("a", 64), "binariesManifestSHA256": strings.Repeat("b", 64), "planSHA256": strings.Repeat("c", 64)}
	h := write("human-approval.json", human)
	r := write("source-review.json", review)
	grant := map[string]any{"schema": 1, "renewalId": renewalID, "decision": "APPROVED", "commonRootSha256": minimaxlive.Hash([]byte(root)), "sourceCommit": commit, "executionSourceSha256": minimaxlive.Hash(sourceRaw), "originalLedgerSha256": minimaxlive.Hash(original), "originalLedgerTailSha256": receipts[len(receipts)-1].Hash, "originalCalls": 39, "originalAttempts": 36, "maxCalls": 60, "maxAttempts": 72, "perIdentityCalls": 20, "perIdentityAttempts": 24, "maxOutputTokens": 4096, "maxRequestedOutput": 40000, "maxReservedInput": 120000, "humanApprovalSha256": minimaxlive.Hash(h), "sourceReviewSha256": minimaxlive.Hash(r), "sourceManifestSha256": review["sourceManifestSHA256"], "binariesManifestSha256": review["binariesManifestSHA256"], "planSha256": review["planSHA256"]}
	switch mode {
	case "pending":
		grant["decision"] = "PENDING"
	case "human-missing":
		os.Remove(filepath.Join(base, "human-approval.json"))
	case "human-pending":
		human["decision"] = "PENDING"
		grant["humanApprovalSha256"] = minimaxlive.Hash(write("human-approval.json", human))
	case "source-stale":
		grant["executionSourceSha256"] = strings.Repeat("d", 64)
	case "source-commit":
		grant["sourceCommit"] = strings.Repeat("d", 40)
	case "old-drift":
		os.WriteFile(old, append(original, '\n'), 0600)
	case "caps":
		grant["maxCalls"] = 61
	case "foreign-owner":
		grant["commonRootSha256"] = strings.Repeat("d", 64)
	case "unknown-new-field":
		grant["previousCalls"] = 1
	}
	if mode == "uncommitted-source" {
		os.WriteFile(filepath.Join(root, "ai/runtime.go"), []byte("package ai\n// WIP\n"), 0600)
		source["ai/runtime.go"] = minimaxlive.Hash([]byte("package ai\n// WIP\n"))
		changed, _ := json.Marshal(source)
		grant["executionSourceSha256"] = minimaxlive.Hash(changed)
	}
	if mode == "data-only-commit" {
		os.WriteFile(filepath.Join(root, "readme.txt"), []byte("data-only\n"), 0600)
		for _, args := range [][]string{{"add", "readme.txt"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "data-only"}} {
			c := exec.Command("git", args...)
			if o, e := c.CombinedOutput(); e != nil {
				t.Fatalf("git %v %s", e, o)
			}
		}
	}
	if mode == "inactive-option" {
		grant["sourceCommit"] = "--output=" + filepath.Join(root, "option-output")
	}
	write("grant.json", grant)
	if mode == "missing" {
		os.Remove(filepath.Join(base, "grant.json"))
	}
	if mode == "inactive-option" {
		_, _, e := minimaxlive.RenewalAuthority(renewalID, false)
		if e == nil {
			t.Fatal("malformed commit accepted")
		}
		if _, e := os.Stat(filepath.Join(root, "option-output")); !os.IsNotExist(e) {
			t.Fatal("unvalidated commit option wrote external output")
		}
		return
	}
	if mode == "data-only-commit" {
		if _, _, e := minimaxlive.RenewalAuthority(renewalID, false); e != nil {
			t.Fatalf("unchanged selected source rejected after data-only commit: %v", e)
		}
		if _, _, e := minimaxlive.RenewalAuthority(renewalID, true); e == nil {
			t.Fatal("active execution borrowed old frozen HEAD")
		}
		return
	}
	_, _, e = minimaxlive.RenewalAuthority(renewalID, false)
	invalid := map[string]bool{"pending": true, "human-missing": true, "human-pending": true, "source-stale": true, "source-commit": true, "old-drift": true, "caps": true, "foreign-owner": true, "uncommitted-source": true, "missing": true, "unknown-new-field": true}
	if invalid[mode] {
		if e == nil {
			t.Fatalf("invalid renewal acquired authority: %s", mode)
		}
		fmt.Println("REJECTED", mode, e)
		return
	}
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = minimaxlive.RenewalAuthority(renewalID, true); e == nil {
		t.Fatal("historical verification reopened closed execution")
	}
	after, _ := os.ReadFile(old)
	if string(original) != string(after) {
		t.Fatal("original formal ledger changed")
	}
	fmt.Println("PASS original renewal read-only, active permanently closed")
}
