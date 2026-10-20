package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Separate processes share the trusted authorization domain even when grant
// bytes are copied or a caller changes only the label of the same approval.
func TestRegionalProviderReleaseGrantAliases(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "regional-live.test")
	build := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live", "-o", binary, "./ai/live")
	build.Dir, build.Env = "../..", regionalChildEnvironment()
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, regionalSafeOutput(raw))
	}
	for _, mode := range []string{"copied", "relabeled", "concurrent", "partial", "offset"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "RP18-grant-alias-"+mode)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:regional-controlled"})
			defer srv.Close()
			for range 48 {
				srv.Enqueue(provider.Reply{Status: 400, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic refused"}}`)}})
			}
			grant, err := supportmatrix.RegionalGrantSnapshot("minimax-cn-chat", "candidate")
			if err != nil {
				t.Fatal(err)
			}
			grant.Mode, grant.ID, grant.AccountAlias, grant.RegionAlias = "controlled", "synthetic-"+strings.ReplaceAll(filepath.Base(run.Dir()), ".", "")+"-"+mode, "synthetic-account", "synthetic-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			paths := []string{filepath.Join(t.TempDir(), "original.json"), filepath.Join(t.TempDir(), "alias.json")}
			first := mustMarshal(t, grant)
			if mode == "relabeled" || mode == "offset" {
				grant.ID += "-alias"
			}
			if mode == "offset" {
				// Same approved instant, different RFC3339 offset spelling.
				// A new label and representation must not create more budget.
				grant.ExpiresAt = grant.ExpiresAt.In(time.FixedZone("synthetic-offset", 8*60*60))
			}
			for i, path := range paths {
				raw := first
				if i == 1 {
					raw = mustMarshal(t, grant)
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "partial" {
				// An interrupted reservation consumes the semantic approval
				// even before its receipt can be completed. Keep this partial
				// original for explicit Host reconciliation; never auto-delete.
				common := exec.CommandContext(ctxFor(t), "git", "rev-parse", "--path-format=absolute", "--git-common-dir")
				raw, err := common.Output()
				if err != nil {
					t.Fatal(err)
				}
				grant.ID = ""
				scope, _ := json.Marshal(grant)
				ledger := filepath.Join(strings.TrimSpace(string(raw)), "barness-regional-grants")
				if err := os.MkdirAll(ledger, 0700); err != nil {
					t.Fatal(err)
				}
				f, err := os.OpenFile(filepath.Join(ledger, "scope-"+pioracle.SHA256(scope)+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.WriteString("{"); err != nil {
					t.Fatal(err)
				}
				if err = f.Sync(); err != nil {
					t.Fatal(err)
				}
				if err = f.Close(); err != nil {
					t.Fatal(err)
				}
			}
			commands := make([]*exec.Cmd, 2)
			bases := make([]string, 2)
			if mode == "offset" {
				// Same approved instant, different RFC3339 offset spelling.
				// A new label and representation must not create more budget.
				grant.ExpiresAt = grant.ExpiresAt.In(time.FixedZone("synthetic-offset", 8*60*60))
			}
			for i, path := range paths {
				bases[i] = filepath.Join(filepath.Dir(run.Dir()), "regional-grant-alias-native", filepath.Base(run.Dir()), mode, []string{"first", "alias"}[i])
				c := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=120s")
				c.Dir, c.Env = "../..", regionalChildEnvironment("BARNESS_REGIONAL_ROUTE=minimax-cn-chat", "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE=minimax-cn", "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+bases[i])
				commands[i] = c
			}
			outputs := make([]string, 2)
			beforeAlias := 0
			if mode == "concurrent" {
				type outcome struct {
					index int
					raw   []byte
					err   error
				}
				done := make(chan outcome, 2)
				start := make(chan struct{})
				for i, c := range commands {
					go func(i int, c *exec.Cmd) { <-start; raw, err := c.CombinedOutput(); done <- outcome{i, raw, err} }(i, c)
				}
				close(start)
				for range 2 {
					r := <-done
					outputs[r.index] = regionalSafeOutput(r.raw)
					ev.Check("deliberately failed process", r.err != nil, "index=%d", r.index)
				}
			} else {
				for i, c := range commands {
					raw, err := c.CombinedOutput()
					outputs[i] = regionalSafeOutput(raw)
					ev.Check("deliberately failed process", err != nil, "index=%d", i)
					if i == 0 {
						beforeAlias = len(srv.Requests())
					}
				}
			}
			reservations := 0
			for _, base := range bases {
				matches, _ := filepath.Glob(filepath.Join(base, "*", "regional-authorization-receipt.json"))
				reservations += len(matches)
			}
			rejected := 0
			for _, raw := range outputs {
				if strings.Contains(raw, "grant reservation failed") {
					rejected++
				}
			}
			for i, base := range bases {
				bases[i], _ = filepath.Rel(run.Dir(), base)
			}
			ev.Record("processes", map[string]any{"outputs": outputs, "nativeRoots": bases, "reservations": reservations, "rejectedBeforeSend": rejected, "requests": srv.Requests(), "beforeAlias": beforeAlias, "vendorCalls": 0})
			if mode == "partial" {
				ev.Check("ambiguous approval cannot send", reservations == 0 && rejected == 2 && len(srv.Requests()) == 0, "reservations=%d rejected=%d requests=%d", reservations, rejected, len(srv.Requests()))
				return
			}
			ev.Check("one approval permits only one process", reservations == 1 && rejected == 1 && len(srv.Requests()) > 0, "reservations=%d rejected=%d requests=%d", reservations, rejected, len(srv.Requests()))
			if mode != "concurrent" {
				ev.Check("alias cannot spend consumed approval", len(srv.Requests()) == beforeAlias, "before=%d after=%d", beforeAlias, len(srv.Requests()))
			}
		})
	}
}
