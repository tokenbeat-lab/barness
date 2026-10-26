package e2e

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var kimiQualificationRoutes = []string{"kimi-cn-chat", "kimi-cn-responses", "kimi-cn-messages", "kimi-intl-chat", "kimi-intl-responses", "kimi-intl-messages"}

func kimiQualificationConsumer(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "regionalverify")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalverify")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build keyless consumer: %v %s", err, regionalSafeOutput(raw))
	}
	return binary
}

func TestKimiQualificationTemplate(t *testing.T) {
	binary := kimiQualificationConsumer(t)
	for _, route := range kimiQualificationRoutes {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+route+"-K16-template")
			cmd := exec.CommandContext(ctxFor(t), binary, "-grant-template", route)
			cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			if !ev.Check("own machine template available without secret or network", err == nil, "%v %s", err, raw) {
				return
			}
			var g map[string]any
			if json.Unmarshal(raw, &g) != nil {
				t.Fatal("template JSON")
			}
			ev.Check("fixed ten calls and finite route-owned token reservations", g["maxCalls"] == float64(10) && g["maxAttempts"] == float64(10) && g["maxRequestedOutputTokens"] == float64(40960) && g["maxReservedInputTokens"] == float64(131072), "%s", raw)
			ev.Check("actual Provider and independent Kimi authorization", strings.HasPrefix(g["provider"].(string), "moonshot") && strings.HasPrefix(g["contract"].(string), "kimi/"), "%s", raw)
			ev.Check("template never grants human spend or executable identity", g["maxUSD"] == nil || g["maxUSD"] == float64(0), "%s", raw)
		})
	}
	cmd := exec.CommandContext(ctxFor(t), binary, "-source")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	ev := run.Case(t, "KIMI-K16-source-inputs")
	var source struct{ Files map[string]string }
	mustUnmarshal(t, raw, &source)
	ev.Check("freeze includes candidate and original sources", source.Files["ai/internal/testkit/kimi/candidate.json"] != "" && source.Files["ai/internal/testkit/kimi/sources/cn-api-messages.md.gz"] != "", "missing Kimi frozen input")
}
