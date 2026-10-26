package e2e

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The public native template must reserve the added foreign-history call;
// old nine-call authorizations cannot authorize a ten-call Kimi process.
func TestGrokHistoryKimiGrantTemplates(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "regionalverify")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./ai/live/cmd/regionalverify")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	output, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("isolated consumer build: %v %s", err, output)
	}
	for _, route := range []string{"kimi-cn-chat", "kimi-intl-chat", "kimi-cn-messages", "kimi-intl-messages", "kimi-cn-responses", "kimi-intl-responses"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-KIMI-GRANT-"+route)
			cmd := exec.CommandContext(ctx, binary, "-grant-template", route, "-stage", "candidate")
			cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "output": string(raw), "success": err == nil})
			var grant struct {
				MaxCalls, MaxAttempts, MaxRequestedOutputTokens, MaxReservedInputTokens int
			}
			if !ev.Check("native template command succeeds", err == nil && json.Unmarshal(raw, &grant) == nil, "%v", err) {
				return
			}
			ev.Check("fresh template includes ten fixed calls", grant.MaxCalls == 10 && grant.MaxAttempts == 10, "calls=%d attempts=%d", grant.MaxCalls, grant.MaxAttempts)
			ev.Check("requested output reserves all ten calls", grant.MaxRequestedOutputTokens == 40960, "output=%d", grant.MaxRequestedOutputTokens)
			ev.Check("existing finite input envelope retained", grant.MaxReservedInputTokens == 131072, "input=%d", grant.MaxReservedInputTokens)
		})
	}
}

func TestGrokHistoryNativeFullPlans(t *testing.T) {
	for _, route := range []string{"zai-chat", "zai-messages", "minimax-cn-chat", "minimax-cn-messages", "minimax-cn-responses", "kimi-cn-chat", "kimi-cn-messages", "kimi-cn-responses", "kimi-intl-chat", "kimi-intl-messages", "kimi-intl-responses", "zhipu-chat", "zhipu-messages", "minimax-chat", "minimax-messages", "minimax-responses"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-FULL-PLAN-"+route)
			var expected []string
			want := 10
			switch {
			case route[:6] == "zhipu-":
				expected, want = supportmatrix.GLMExpected(route), 24
			case route[:4] == "zai-":
				expected, want = supportmatrix.RegionalExpected(route), 20
			case route == "minimax-chat" || route == "minimax-messages" || route == "minimax-responses":
				id, _ := minimaxlive.Select(route[8:], minimaxlive.M3, "standard")
				for _, c := range minimaxlive.Plan(id) {
					expected = append(expected, c.ID)
				}
				want = 14
			default:
				expected = supportmatrix.RegionalExpected(route)
			}
			ev.Record("full-native-plan", expected)
			ev.Check("all old capabilities plus one per fixed model", len(expected) == want, "got=%d want=%d", len(expected), want)
			count := 0
			for _, s := range expected {
				if s == grokhistory.Scenario || len(s) > len(grokhistory.Scenario) && s[len(s)-len(grokhistory.Scenario):] == grokhistory.Scenario {
					count++
				}
			}
			models := 1
			if want == 20 || want == 24 {
				models = 2
			}
			ev.Check("added foreign-history scene is mandatory per model", count == models, "got=%d want=%d", count, models)
		})
	}
}

func TestGrokHistoryProducerBuild(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-PRODUCER-BUILD")
	cmd := exec.Command("go", "test", "-c", "-tags", "live", "-o", filepath.Join(t.TempDir(), "live"), "./ai/live")
	cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
	raw, err := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "output": string(raw), "success": err == nil})
	ev.Check("complete public native producer compiles without credentials", err == nil, "%s", raw)
}
