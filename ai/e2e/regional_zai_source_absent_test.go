package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The predeclared full-stream source lacks native reasoning while tool-call
// has it. The complete real process must fail; searching another turn is banned.
func regionalMissingNativeSource(t *testing.T, binary, combo string) {
	t.Run("native-source-absent", func(t *testing.T) {
		ev := run.Case(t, "RP18-"+combo+"-native-source-absent")
		srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:regional-controlled"})
		defer srv.Close()
		route, _ := supportmatrix.RegionalRouteOf(combo)
		for _, model := range supportmatrix.RegionalModels(combo) {
			for _, kind := range []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "tool-call", "tool-result", "native-history", "cancel", "recovery"} {
				reply := regionalTextReply(t, ai.ProviderID(route.Provider), ai.API(route.API), model)
				if kind == "full-stream" && combo == "zai-chat" {
					reply = chatEvents(t, []json.RawMessage{
						json.RawMessage(`{"id":"source-without-native","model":"` + model + `","choices":[{"delta":{"content":"Synthetic source without reasoning."}}]}`),
						json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`),
						json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":8}}`),
					}, "")
				}
				if kind == "tool-call" {
					reply = regionalPipelineToolReply(t, ai.ProviderID(route.Provider), ai.API(route.API), model)
				}
				if kind == "cancel" {
					reply = regionalPipelinePrefix(t, ai.API(route.API))
					reply.End = provider.EndHold
				}
				if kind == "tool-result" || kind == "native-history" {
					for i, raw := range reply.Chunks {
						reply.Chunks[i] = []byte(strings.NewReplacer("Regional text.", "REGIONAL42", "Hello MiniMax.", "REGIONAL42", "Synthetic reply.", "REGIONAL42").Replace(string(raw)))
					}
				}
				reply.Header["X-Request-Id"] = "regional-controlled"
				srv.Enqueue(reply)
			}
		}
		grant, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
		if err != nil {
			t.Fatal(err)
		}
		grant.Mode, grant.ID = "controlled", "absent-native-"+combo+"-"+evidence.Digest([]byte(run.Dir()))[:12]
		grant.AccountAlias, grant.RegionAlias, grant.ExpiresAt = "synthetic-account", "synthetic-region", time.Now().UTC().Add(time.Hour)
		path := filepath.Join(t.TempDir(), "grant.json")
		if err := os.WriteFile(path, mustMarshal(t, grant), 0600); err != nil {
			t.Fatal(err)
		}
		base := filepath.Join(filepath.Dir(run.Dir()), "regional-source-absent-native", filepath.Base(run.Dir()), combo)
		cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
		cmd.Dir, cmd.Env = "../..", regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+base)
		raw, err := cmd.CombinedOutput()
		ev.Record("process", map[string]any{"failed": err != nil, "output": regionalSafeOutput(raw), "argv": cmd.Args[1:], "vendorCalls": 0})
		paths, _ := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
		if !ev.Check("one original absence bundle", len(paths) == 1, "count=%d", len(paths)) {
			return
		}
		ev.Check("missing predeclared source fails actual process", err != nil && len(srv.Requests()) == len(supportmatrix.RegionalExpected(combo)), "failed=%t calls=%d", err != nil, len(srv.Requests()))
		dir := filepath.Dir(paths[0])
		var report supportmatrix.Report
		data, err := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile))
		if err != nil || json.Unmarshal(data, &report) != nil {
			t.Fatal("source-absence report unavailable")
		}
		for _, row := range report.Scenarios {
			_, kind, _ := strings.Cut(row.ID, "/")
			if kind == "native-history" {
				ev.Check("native history remains FAIL "+row.Model, row.Outcome == supportmatrix.Fail, "outcome=%s", row.Outcome)
			}
			if kind == "tool-call" {
				ev.Check("tool reasoning cannot substitute "+row.Model, row.Outcome == supportmatrix.Pass, "outcome=%s", row.Outcome)
				var tool struct {
					Message struct{ Content []struct{ Thinking string } }
				}
				raw, err := os.ReadFile(filepath.Join(dir, row.CaseID, "result.json"))
				if err != nil {
					t.Fatal(err)
				}
				mustUnmarshal(t, raw, &tool)
				positive := false
				for _, block := range tool.Message.Content {
					positive = positive || block.Thinking != ""
				}
				ev.Check("another own turn is native-positive "+row.Model, positive, "positive=%t", positive)
			}
			if kind != "native-history" {
				ev.Check("only fixed native absence fails "+row.ID, row.Outcome == supportmatrix.Pass, "outcome=%s", row.Outcome)
			}
		}
		_, err = supportmatrix.VerifyRegionalControlledCapture(dir)
		ev.Check("consumer cannot certify absent native source", err != nil, "accepted")
		relative, relErr := filepath.Rel(run.Dir(), paths[0])
		if relErr != nil {
			t.Fatal(relErr)
		}
		ev.Record("original-native-manifest", relative)
	})
}
