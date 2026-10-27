package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Failure inventory at the compiled public consumer seam: correct fresh
// authority must be independent from stale legacy origins; missing/tampered
// fresh authority cannot fall back. Mixed history, wrong source/session/ledger,
// incomplete scenes, altered seals and controlled formal promotion stay refused.
func TestMiniMaxFreshAuthorityConsumerProcess(t *testing.T) {
	native := miniMaxAuthorityBuild(t, "minimax-live", "./ai/live", true)
	consumer := miniMaxAuthorityBuild(t, "minimaxverify", "./ai/live/cmd/minimaxverify", false)
	for _, route := range []string{"chat", "responses", "messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "MM20-fresh-authority-"+route)
			controlled := miniMaxAuthorityCapture(t, ev, native, route)
			formal := miniMaxAuthorityFormalFixture(t, controlled, native)
			files, err := minimaxlive.Snapshot(formal.dir)
			if err != nil {
				t.Fatal(err)
			}
			ev.Record("synthetic-formal-fixture", files)
			miniMaxAuthorityRecordStore(t, ev, formal.git)
			exit, output := miniMaxAuthorityCLI(t, ev, consumer, formal, "-formal", "-publish")
			if !ev.Check("fresh complete authority qualifies without stale legacy origin", exit == 0 && strings.Contains(output, "PASS: "+route+" MiniMax-M3 mode=live kind=minimax-formal calls=14 attempts=13"), "exit=%d output=%s", exit, output) {
				return
			}
			projection, err := os.ReadFile(filepath.Join(formal.dir, "live-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("synthetic-projection.json", projection)
			ev.Record("fixture-purpose", map[string]any{"actualLive": false, "producerMode": "controlled", "consumerContractFixture": true, "realCredentials": false, "vendorCalls": 0})
			for _, name := range []string{"missing-grant", "changed-grant", "missing-claim", "untrusted-claim", "changed-session", "mixed-renewal", "foreign-identity", "source-change", "ledger-change", "missing-scene", "terminal-change", "seal-change", "legacy-without-fresh"} {
				t.Run(name, func(t *testing.T) {
					proof := run.Case(t, "MM20-fresh-authority-"+route+"-"+name)
					mutated := formal
					mutated.dir = filepath.Join(t.TempDir(), formal.report.Session)
					if err := os.CopyFS(mutated.dir, os.DirFS(formal.dir)); err != nil {
						t.Fatal(err)
					}
					miniMaxAuthorityMutation(t, mutated.dir, name)
					if name == "untrusted-claim" {
						mutated.git = miniMaxPrivateSourceNamespace(t).dir
					}
					exit, output := miniMaxAuthorityCLI(t, proof, consumer, mutated, "-formal")
					proof.Check("invalid authority or capture is refused", exit == 1, "exit=%d output=%s", exit, output)
					if name != "legacy-without-fresh" {
						proof.Check("invalid fresh authority cannot borrow legacy origin", !strings.Contains(output, "origin launcher"), "%s", output)
					} else {
						proof.Check("legacy still requires its independent origin", strings.Contains(output, "origin launcher receipt invalid"), "%s", output)
					}
					files, err := minimaxlive.Snapshot(mutated.dir)
					if err != nil {
						t.Fatal(err)
					}
					proof.Record("synthetic-rejected-fixture", files)
				})
			}
			t.Run("controlled", func(t *testing.T) {
				proof := run.Case(t, "MM20-fresh-authority-"+route+"-controlled")
				controlled.source = formal.source
				exit, output := miniMaxAuthorityCLI(t, proof, consumer, controlled, "-formal")
				proof.Check("controlled public capture cannot certify formal support", exit == 1 && strings.Contains(output, "controlled and replay cannot certify formal support"), "exit=%d output=%s", exit, output)
			})
		})
	}
}

func miniMaxAuthorityMutation(t *testing.T, dir, name string) {
	t.Helper()
	mutate := func(path string, change func(map[string]any)) {
		raw, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		mustUnmarshal(t, raw, &object)
		change(object)
		miniMaxAuthorityWrite(t, filepath.Join(dir, path), object)
	}
	switch name {
	case "missing-grant", "missing-claim":
		leaf := "process-grant.json"
		if name == "missing-claim" {
			leaf = "process-authorization-receipt.json"
		}
		if err := os.Remove(filepath.Join(dir, leaf)); err != nil {
			t.Fatal(err)
		}
	case "changed-grant":
		mutate("process-grant.json", func(v map[string]any) { v["scope"].(map[string]any)["binarySha256"] = strings.Repeat("0", 64) })
	case "changed-session":
		mutate("process-authorization-receipt.json", func(v map[string]any) { v["session"] = "synthetic-other-session" })
	case "mixed-renewal":
		mutate(minimaxlive.ReportFile, func(v map[string]any) {
			v["formalRenewalId"], v["formalRenewalGrantSha256"] = "synthetic-old-renewal", strings.Repeat("0", 64)
		})
	case "foreign-identity":
		mutate(minimaxlive.ReportFile, func(v map[string]any) {
			v["identity"].(map[string]any)["model"] = minimaxlive.M31
			v["identity"].(map[string]any)["purpose"] = "mplan"
		})
	case "source-change":
		mutate(minimaxlive.ReportFile, func(v map[string]any) {
			v["executionSource"].(map[string]any)["ai/client.go"] = strings.Repeat("0", 64)
		})
	case "ledger-change":
		mutate("minimax-ledger.json", func(v map[string]any) { v["receipts"] = []any{} })
	case "missing-scene":
		mutate(minimaxlive.ReportFile, func(v map[string]any) { v["calls"] = v["calls"].([]any)[1:] })
	case "terminal-change":
		mutate("MM07-thinking-on/result.json", func(v map[string]any) {
			v["result"].(map[string]any)["message"].(map[string]any)["stopReason"] = "error"
		})
		raw, err := os.ReadFile(filepath.Join(dir, "MM07-thinking-on/result.json"))
		if err != nil {
			t.Fatal(err)
		}
		kimiQualificationManifest(t, dir, "MM07-thinking-on", "result.json", raw)
	case "seal-change":
		mutate(minimaxlive.SealFile, func(v map[string]any) { v["files"].(map[string]any)[minimaxlive.ReportFile] = strings.Repeat("0", 64) })
		return
	case "legacy-without-fresh":
		mutate(minimaxlive.ReportFile, func(v map[string]any) { delete(v, "processGrantSha256") })
	}
	miniMaxAuthoritySeal(t, dir)
}
