package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestKimiQualificationMissingCharges(t *testing.T) {
	privateGrantStore(t)
	binary, digest := kimiQualificationBinary(t)
	for _, mode := range []string{"missing-charge", "missing-native", "expires-during-call", "output-exceeds-reserve"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-"+mode+"-stops-process")
			combo := "kimi-intl-messages"
			route, _ := supportmatrix.RegionalRouteOf(combo)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				reply := kimiQualificationReply(t, route.API, kind)
				if mode == "output-exceeds-reserve" && kind == "full-stream" {
					for i, c := range reply.Chunks {
						reply.Chunks[i] = []byte(strings.ReplaceAll(string(c), `"output_tokens":50`, `"output_tokens":4097`))
					}
				}
				if mode == "expires-during-call" && kind == "full-stream" {
					reply.End = provider.EndHold
				}
				if mode == "missing-native" && kind == "full-stream" {
					reply = kimiQualificationReply(t, route.API, "full-complete")
				}
				if mode == "missing-charge" && kind == "full-complete" {
					for i, c := range reply.Chunks {
						reply.Chunks[i] = []byte(strings.ReplaceAll(string(c), `,"cache_creation_input_tokens":80`, ""))
					}
				}
				srv.Enqueue(reply)
			}
			grant, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
			if err != nil {
				t.Fatal(err)
			}
			grant.Mode = "controlled"
			grant.ID = mode + "-" + evidence.Digest([]byte(run.Dir()))[:12]
			grant.AccountAlias, grant.RegionAlias = "synthetic-kimi-account", "synthetic-kimi-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			if mode == "expires-during-call" {
				grant.ExpiresAt = time.Now().UTC().Add(time.Second)
			}
			grant.MaxUSD = 10
			grant.BinarySHA256 = digest
			path := filepath.Join(t.TempDir(), "grant.json")
			if err := os.WriteFile(path, mustMarshal(t, grant), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(filepath.Dir(run.Dir()), "kimi-partial-native", filepath.Base(run.Dir()), mode))
			began := time.Now()
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			ev.Record("requests", srv.Requests())
			want := 2
			if mode == "missing-native" || mode == "expires-during-call" || mode == "output-exceeds-reserve" {
				want = 1
			}
			ev.Check("missing prerequisite refuses qualification and stops before later calls", err != nil && len(srv.Requests()) == want && (mode != "expires-during-call" || time.Since(began) < 4*time.Second), "%v requests=%d %s", err, len(srv.Requests()), regionalSafeOutput(raw))
		})
	}
}
