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

const finalRenewalID = "provider-sites-final-acceptance-2"

// This independent process owns a separate Git repository and only synthetic
// approval/history. It cannot authorize the shared real repository or HTTP.
func TestSecondRenewalClosedFirstProcess(t *testing.T) {
	for _, mode := range []string{"closed-unused-active", "closed-unused-path", "unused-native-intent", "unused-native-grant", "unused-native-nonce", "cohort-seal", "cohort-human", "cohort-review", "cohort-archive-extra", "cohort-large"} {
		t.Run(mode, func(t *testing.T) { testFinalRenewalFixture(t, mode) })
	}
}

func TestFinalRenewalAuthorityProcess(t *testing.T) {
	for _, mode := range []string{"valid-three", "facts", "closed-active", "missing", "pending", "history-missing", "history-added", "history-modified", "history-linked", "history-parent-linked", "history-launcher", "history-open", "human-missing", "human-pending", "human-plan", "human-history", "human-purpose", "human-routes", "caps", "wrong-id", "candidate", "source-stale", "source-uncommitted", "source-option", "foreign-owner", "arbitrary-path", "unregistered", "new-session", "deleted-ledger", "partial-ledger", "partial-claim", "changed-grant", "duplicate", "excess-output", "excess-input", "linked-worktree", "concurrent", "same-session-process", "preflight-claim", "preflight-crash", "budget-source", "regional-price-source", "launcher-preflight", "preflight-purpose"} {
		t.Run(mode, func(t *testing.T) { testFinalRenewalFixture(t, mode) })
	}
}

func testFinalRenewalFixture(t *testing.T, mode string) {
	t.Helper()
	root := t.TempDir()
	if e := os.Mkdir(filepath.Join(root, "ai"), 0700); e != nil {
		t.Fatal(e)
	}
	for n, raw := range map[string]string{"go.mod": "module final-renewal-fixture\n", "go.sum": "", "ai/runtime.go": "package ai\n", "ai/internal/testkit/finalbudget/guard.go": "package finalbudget\n", "ai/internal/testkit/regional/prices.json": "{\"input\":1}\n"} {
		if e := os.MkdirAll(filepath.Dir(filepath.Join(root, n)), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(root, n), []byte(raw), 0600); e != nil {
			t.Fatal(e)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "go.sum", "ai"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "source"}} {
		c := exec.Command("git", args...)
		c.Dir = root
		if o, e := c.CombinedOutput(); e != nil {
			t.Fatalf("git %v %s", e, o)
		}
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	c := exec.Command(exe, "-test.run=^TestFinalRenewalAuthorityChild$", "-test.v")
	c.Dir = root
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_FINAL_RENEWAL_CHILD=" + mode}
	o, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("mode=%s %v\n%s", mode, e, o)
	}
	t.Log(strings.TrimSpace(string(o)))
}

func TestFinalRenewalAuthorityChild(t *testing.T) {
	mode := os.Getenv("BARNESS_FINAL_RENEWAL_CHILD")
	if mode == "" {
		t.Skip("private synthetic process only")
	}
	f := newFinalRenewalFixture(t)
	f.installGrant(t)
	if finalRenewalVariant(t, f, mode) {
		return
	}
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", finalRenewalID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if finalRenewalRejectedMode(mode) {
		if e == nil {
			t.Fatalf("invalid final renewal acquired authority: %s", mode)
		}
		fmt.Println("REJECTED", mode, e)
		return
	}
	if e != nil {
		t.Fatal(e)
	}
	if path != filepath.Join(f.base, "formal.jsonl") {
		t.Fatalf("incorrect canonical path %s", path)
	}
	if mode != "valid-three" {
		testFinalRenewalConsumption(t, f, path, mode)
		return
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "final-" + route, Identity: id}
		for _, c := range minimaxlive.Plan(id) {
			if _, e = l.Reserve(c.ID, c.Output, 2048); e != nil {
				t.Fatal(e)
			}
			if c.ID != "preflight-error" {
				if _, e = l.Attempt(c.ID); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	rows, e := minimaxlive.ReadLedger(path)
	if e != nil || len(rows) != 75 {
		t.Fatalf("want complete three identities 39/36, rows=%d error=%v", len(rows), e)
	}
	for _, row := range rows {
		if row.RenewalID != finalRenewalID || row.RenewalGrantSHA256 != f.grantPin {
			t.Fatal("new consumption borrowed old authority")
		}
	}
	f.checkHistory(t)
	fmt.Println("PASS independent final39/36, closed original+renewal unchanged")
}

type finalRenewalFixture struct {
	root, base, commit, grantPin string
	original, previous           []minimaxlive.Receipt
	history                      map[string][]byte
	grant                        map[string]any
}

// Construct the closed predecessor independently from the old read-only
// fixture: changing today's grant builder must not rewrite yesterday's facts.
func newFinalRenewalFixture(t *testing.T) *finalRenewalFixture {
	t.Helper()
	root, _ := os.Getwd()
	root, _ = filepath.EvalSymlinks(root)
	c := exec.Command("git", "rev-parse", "HEAD")
	raw, e := c.Output()
	if e != nil {
		t.Fatal(e)
	}
	f := &finalRenewalFixture{root: root, base: filepath.Join(root, ".evidence/minimax-renewals", finalRenewalID), commit: strings.TrimSpace(string(raw)), history: map[string][]byte{}}
	old, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		l := minimaxlive.Ledger{Path: old, Stage: "formal", Session: "original-" + route, Identity: id}
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
	f.original, e = minimaxlive.ReadLedger(old)
	if e != nil {
		t.Fatal(e)
	}
	original, _ := os.ReadFile(old)
	f.history["minimax-stages/formal.jsonl"] = original
	source, e := minimaxlive.ExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	sourceRaw, _ := json.Marshal(source)
	legacy := "minimax-renewals/" + renewalID + "/"
	human := map[string]any{"decision": "APPROVED", "renewalId": renewalID, "maxCalls": 216, "maxAttempts": 252, "requestReference": "synthetic-closed-history-no-paid"}
	review := map[string]any{"decision": "APPROVED", "sourceCommit": f.commit, "sourceManifestSHA256": strings.Repeat("a", 64), "binariesManifestSHA256": strings.Repeat("b", 64), "planSHA256": strings.Repeat("c", 64)}
	h := f.writeHistory(t, legacy+"human-approval.json", human)
	r := f.writeHistory(t, legacy+"source-review.json", review)
	f.grant = map[string]any{"schema": 1, "renewalId": renewalID, "decision": "APPROVED", "commonRootSha256": minimaxlive.Hash([]byte(root)), "sourceCommit": f.commit, "executionSourceSha256": minimaxlive.Hash(sourceRaw), "originalLedgerSha256": minimaxlive.Hash(original), "originalLedgerTailSha256": f.original[len(f.original)-1].Hash, "originalCalls": 39, "originalAttempts": 36, "maxCalls": 60, "maxAttempts": 72, "perIdentityCalls": 20, "perIdentityAttempts": 24, "maxOutputTokens": 4096, "maxRequestedOutput": 40000, "maxReservedInput": 120000, "humanApprovalSha256": minimaxlive.Hash(h), "sourceReviewSha256": minimaxlive.Hash(r), "sourceManifestSha256": review["sourceManifestSHA256"], "binariesManifestSha256": review["binariesManifestSHA256"], "planSha256": review["planSHA256"]}
	oldGrant := f.writeHistory(t, legacy+"grant.json", f.grant)
	oldPin := minimaxlive.Hash(oldGrant)
	previous := ""
	var ledger []byte
	for _, row := range f.original {
		row.RenewalID, row.RenewalGrantSHA256 = renewalID, oldPin
		row.Session = "previous-" + row.Identity.Route
		row.Previous, row.Hash = previous, ""
		b, _ := json.Marshal(row)
		row.Hash = minimaxlive.Hash(b)
		previous = row.Hash
		f.previous = append(f.previous, row)
		b, _ = json.Marshal(row)
		ledger = append(ledger, append(b, '\n')...)
	}
	f.writeHistoryBytes(t, legacy+"formal.jsonl", ledger)
	f.writeHistory(t, legacy+"finalize-intent.json", map[string]any{"state": "COMPLETED", "noAutomaticReruns": true})
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		f.writeHistory(t, legacy+route+".json", map[string]any{"schema": 1, "renewalId": renewalID, "grantSha256": oldPin, "session": "previous-" + route, "identity": id})
		for _, previous := range []bool{false, true} {
			name, session, pin, selected := "mm-global-formal-m3-"+route, "original-"+route, "", ""
			if previous {
				name, session, pin, selected = renewalID+"-"+name, "previous-"+route, oldPin, renewalID
			}
			origin := minimaxlive.Origin{Schema: 1, Stage: "formal", Identity: id, Session: session, SourceCommit: f.commit, FormalRenewalID: selected, FormalRenewalGrantSHA256: pin, SourceManifestSHA256: strings.Repeat("a", 64), BinariesManifestSHA256: strings.Repeat("b", 64), PlanSHA256: strings.Repeat("c", 64), ReportSHA256: strings.Repeat("d", 64), SealSHA256: strings.Repeat("e", 64), LedgerSHA256: strings.Repeat("f", 64)}
			f.writeHistory(t, "minimax-launches/"+name+".json", map[string]any{"state": "COMPLETED", "route": name, "sourceCommit": f.commit, "exitCode": 0, "credentialsRead": true, "launcherAudit": map[string]any{"passed": true}, "origin": origin})
		}
	}
	f.grant["schema"], f.grant["renewalId"] = 3, finalRenewalID
	f.grant["previousRenewalId"] = renewalID
	f.grant["previousLedgerSha256"] = minimaxlive.Hash(ledger)
	f.grant["previousLedgerTailSha256"] = f.previous[len(f.previous)-1].Hash
	f.grant["previousCalls"], f.grant["previousAttempts"] = 39, 36
	pins := map[string]string{}
	for name, b := range f.history {
		pins[name] = minimaxlive.Hash(b)
	}
	pinsRaw, _ := json.Marshal(pins)
	f.grant["closedHistorySha256"] = minimaxlive.Hash(pinsRaw)
	f.seedFirstCohort(t, pins)
	return f
}

func (f *finalRenewalFixture) installGrant(t *testing.T) {
	t.Helper()
	human := map[string]any{"decision": "APPROVED", "renewalId": finalRenewalID, "maxCalls": 722, "maxAttempts": 758, "campaignMaxCalls": 774, "campaignMaxAttempts": 810, "newReservedMicrousd": 28400000, "firstCohortSealSHA256": f.grant["firstCohortSealSha256"], "requestReference": "synthetic-private-no-paid", "requestText": "synthetic fixture only", "historySHA256": f.grant["closedHistorySha256"], "planSHA256": f.grant["planSha256"], "provider": "minimax", "model": minimaxlive.M3, "purpose": "standard", "routes": []string{"chat", "responses", "messages"}, "renewalMaxCalls": 60, "renewalMaxAttempts": 72}
	review := map[string]any{"decision": "APPROVED", "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "historySHA256": f.grant["closedHistorySha256"], "firstCohortSealSHA256": f.grant["firstCohortSealSha256"]}
	h, r := f.write(t, "human-approval.json", human), f.write(t, "source-review.json", review)
	f.grant["humanApprovalSha256"], f.grant["sourceReviewSha256"] = minimaxlive.Hash(h), minimaxlive.Hash(r)
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
}

func (f *finalRenewalFixture) write(t *testing.T, name string, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(f.base, name)
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	return b
}

func (f *finalRenewalFixture) writeHistory(t *testing.T, name string, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return f.writeHistoryBytes(t, name, b)
}

func (f *finalRenewalFixture) writeHistoryBytes(t *testing.T, name string, b []byte) []byte {
	t.Helper()
	path := filepath.Join(f.root, ".evidence", name)
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	f.history[name] = b
	return b
}

func (f *finalRenewalFixture) checkHistory(t *testing.T) {
	t.Helper()
	for name, want := range f.history {
		b, e := os.ReadFile(filepath.Join(f.root, ".evidence", name))
		if e != nil || string(b) != string(want) {
			t.Fatalf("closed history changed: %s", name)
		}
	}
}
