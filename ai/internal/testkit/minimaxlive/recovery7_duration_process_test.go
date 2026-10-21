package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// This public process exercises the provenance gate before a deliberately
// absent manifest. Passing that gate does not certify the incomplete capture.
func (f *finalRenewalFixture) verifySeventhLifetimeBoundary(t *testing.T) {
	t.Helper()
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	source, e := minimaxlive.ExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	sdk, e := minimaxlive.SDKVersion(id)
	if e != nil {
		t.Fatal(e)
	}
	plan := minimaxlive.Plan(id)
	raw, _ := json.Marshal(plan)
	for _, tc := range []struct {
		name     string
		duration time.Duration
		renewal  bool
		nextGate bool
	}{
		{"fixed7-over-old1800", 1801 * time.Second, true, true},
		{"fixed7-at3600", 3600 * time.Second, true, true},
		{"fixed7-over3600", 3600*time.Second + time.Nanosecond, true, false},
		{"older-unrenewed-over1800", 1801 * time.Second, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			start := time.Unix(1, 0).UTC()
			r := minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Mode: "controlled", Stage: "formal", Identity: id, Outcome: "PASS", Complete: true, AuditPassed: true, SDK: sdk, AccountAlias: "synthetic", RegionAlias: "synthetic", Entitlement: "confirmed", ExecutionSource: source, Expected: plan, PlanHash: minimaxlive.Hash(raw), SourceHash: minimaxlive.SourcePin, StartedAt: start, FinishedAt: start.Add(tc.duration)}
			if tc.renewal {
				r.Mode = "live"
				r.FormalRenewalID = recovery7ID
				r.FormalRenewalGrantSHA256 = f.grantPin
				r.GitCommit = f.commit
			}
			b, _ := json.Marshal(r)
			os.WriteFile(filepath.Join(dir, minimaxlive.ReportFile), b, 0600)
			os.WriteFile(filepath.Join(dir, "audit.json"), []byte(`{"findings":[]}`), 0600)
			files, e := minimaxlive.Inventory(dir)
			if e != nil {
				t.Fatal(e)
			}
			b, _ = json.Marshal(minimaxlive.Seal{Schema: 1, Files: files})
			os.WriteFile(filepath.Join(dir, minimaxlive.SealFile), b, 0600)
			_, e = minimaxlive.VerifyCapture(dir, false)
			want := "provenance incomplete"
			if tc.nextGate {
				want = "selected catalog/price snapshot is stale"
			}
			if e == nil || !strings.Contains(e.Error(), want) {
				t.Fatalf("public lifetime gate: want %q got %v", want, e)
			}
		})
	}
}
