package e2e

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// Failure inventory: the one-call purpose cannot borrow a standard grant,
// widen cases/budget, or acquire human authority from a template.
func TestKimiHistorySupplementTemplate(t *testing.T) {
	binary := kimiTemplateBinary(t)
	for _, route := range kimiQualificationRoutes {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KHS-template-"+route)
			cmd := exec.CommandContext(ctxFor(t), binary, "-template", route, "-stage", "candidate", "-purpose", "xai-history-supplement", "-baseline-manifest-sha256", strings.Repeat("a", 64))
			cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			if !ev.Check("fixed single-scene template exists", err == nil, "%v", err) {
				return
			}
			var g map[string]any
			if json.Unmarshal(raw, &g) != nil {
				t.Fatal("invalid template")
			}
			ev.Check("new scope is one bounded call", g["schema"] == float64(2) && g["purpose"] == "xai-history-supplement" && g["maxCalls"] == float64(1) && g["maxAttempts"] == float64(1) && g["maxReservedInputTokens"] == float64(4096) && g["maxRequestedOutputTokens"] == float64(4096), "%s", raw)
			ev.Check("original manifest and exact case bound", g["baselineManifestSha256"] == strings.Repeat("a", 64) && string(mustMarshal(t, g["scenarioIds"])) == "[\"kimi-k3/xai-origin-history\"]", "%s", raw)
			ev.Check("template cannot authorize spend", g["maxUSD"] == nil || g["maxUSD"] == float64(0), "%s", raw)
		})
	}
}

func TestKimiHistorySupplementTemplateScope(t *testing.T) {
	binary := kimiTemplateBinary(t)
	for name, args := range map[string][]string{
		"source":         {"-source", "regional", "-purpose", "xai-history-supplement"},
		"legacy":         {"-source", "openai-chat", "-baseline-manifest-sha256", strings.Repeat("a", 64)},
		"check":          {"-check", "missing", "-purpose", "xai-history-supplement"},
		"formal":         {"-template", "kimi-cn-chat", "-stage", "formal", "-purpose", "xai-history-supplement"},
		"other-provider": {"-template", "zai-chat", "-stage", "candidate", "-purpose", "xai-history-supplement", "-baseline-manifest-sha256", strings.Repeat("a", 64)},
		"missing-pin":    {"-template", "kimi-cn-chat", "-stage", "candidate", "-purpose", "xai-history-supplement"},
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "KHS-template-scope-"+name)
			cmd := exec.CommandContext(ctxFor(t), binary, args...)
			cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
			out, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(out), "vendorCalls": 0})
			ev.Check("scope flags cannot be silently ignored", err != nil, "%s", regionalSafeOutput(out))
		})
	}
}
