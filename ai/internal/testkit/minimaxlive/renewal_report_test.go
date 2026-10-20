package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// A submitted report's copied grant never establishes canonical authority.
func TestFormalRenewalPublicCaptureRefusesNonliveAndImportedGrant(t *testing.T) {
	for _, mode := range []string{"controlled", "replay", "live"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
			r := minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Stage: "formal", Mode: mode, Identity: id, FormalRenewalID: "current-source-eleven-renewal-1", FormalRenewalGrantSHA256: strings.Repeat("a", 64)}
			raw, _ := json.Marshal(r)
			os.WriteFile(filepath.Join(dir, minimaxlive.ReportFile), raw, 0600)
			os.WriteFile(filepath.Join(dir, "grant.json"), []byte(`{"decision":"APPROVED"}`), 0600)
			_, e := minimaxlive.VerifyCapture(dir, true)
			if e == nil || !strings.Contains(e.Error(), "renewal") {
				t.Fatalf("want independent renewal refusal before bundle data, got %v", e)
			}
		})
	}
}
