package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// This keyless CLI checks the very admission/claim boundary used before HTTP.
// Every grant here says controlled; none can authorize a vendor process.
func TestGrokHistoryFreshProcessGrant(t *testing.T) {
	privateGrantStore(t)
	binary := filepath.Join(t.TempDir(), "processgrant")
	build := exec.Command("go", "build", "-o", binary, "./ai/live/cmd/processgrant")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	raw, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("build grant boundary: %v %s", err, raw)
	}
	for _, route := range []string{"openai-responses", "openai-chat", "deepseek-responses", "deepseek-chat", "anthropic-messages", "google-gemini", "zhipu-chat", "zhipu-messages", "minimax-chat", "minimax-responses", "minimax-messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-FRESH-GRANT-"+route)
			probe := exec.CommandContext(ctxFor(t), binary)
			probe.Dir, probe.Env = "../..", xaiProcessEnv(t)
			if !privateGrantProcess(t, ev, probe) {
				return
			}
			call := func(args ...string) ([]byte, error) {
				c := exec.Command(binary, args...)
				c.Dir, c.Env = "../..", xaiProcessEnv(t)
				out, e := c.CombinedOutput()
				ev.Record("process-"+time.Now().Format("150405.000000000"), map[string]any{"argv": c.Args, "output": string(out), "success": e == nil})
				return out, e
			}
			raw, e := call("-template", route, "-mode", "controlled", "-stage", "candidate", "-binary", binary)
			if e != nil {
				t.Fatalf("template: %v %s", e, raw)
			}
			var grant map[string]any
			if e = json.Unmarshal(raw, &grant); e != nil {
				t.Fatal(e)
			}
			sourceRaw, sourceErr := call("-source", route, "-mode", "controlled", "-stage", "candidate")
			var facts map[string]any
			_ = json.Unmarshal(sourceRaw, &facts)
			ev.Check("prebuild source facts match exact template scope", sourceErr == nil && facts["sourceSha256"] == grant["scope"].(map[string]any)["sourceSha256"], "%s", sourceRaw)
			grant["id"] = "synthetic-" + route + "-" + time.Now().Format("20060102150405000000000")
			grant["accountAlias"], grant["regionAlias"] = "synthetic-account", "synthetic-region"
			grant["expiresAt"] = time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
			file := filepath.Join(t.TempDir(), "grant.json")
			check := func(v map[string]any, claim bool) error {
				b, _ := json.Marshal(v)
				if err := os.WriteFile(file, b, 0600); err != nil {
					t.Fatal(err)
				}
				args := []string{"-check", file, "-route", route, "-mode", "controlled", "-stage", "candidate", "-binary", binary}
				if claim {
					args = append(args, "-claim")
				}
				_, err := call(args...)
				return err
			}
			for _, mutation := range []string{"expired", "route", "binary", "source", "scenes", "budget", "account", "unknown", "null"} {
				var bad map[string]any
				b, _ := json.Marshal(grant)
				_ = json.Unmarshal(b, &bad)
				scope := bad["scope"].(map[string]any)
				switch mutation {
				case "unknown":
					bad["unexpected"] = true
				case "null":
					scope["maxCalls"] = nil
				case "expired":
					bad["expiresAt"] = time.Now().Add(-time.Hour).Format(time.RFC3339Nano)
				case "account":
					bad["accountAlias"] = ""
				case "route":
					scope["route"] = "mixed-route"
				case "binary":
					scope["binarySha256"] = "wrong"
				case "source":
					scope["sourceSha256"] = "wrong"
				case "scenes":
					scope["scenes"] = []string{"text-stream"}
				case "budget":
					scope["maxCalls"] = 999
				}
				ev.Check("refuses "+mutation+" before any HTTP", check(bad, false) != nil, "accepted mutation")
			}
			ev.Check("new exact controlled grant is claimable once", check(grant, true) == nil, "fresh grant rejected")
			ev.Check("same approval cannot be reused", check(grant, true) != nil, "reused grant accepted")
		})
	}
}

// Check the public subprocess before burning even a synthetic approval. Shared
// real records are not fixture storage, including for deliberate partial writes.
func privateGrantProcess(t *testing.T, ev *evidence.Case, cmd *exec.Cmd) bool {
	t.Helper()
	common := miniMaxGitProbe(t, cmd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	real := *cmd
	real.Env = nil
	for _, pair := range cmd.Env {
		if !strings.HasPrefix(pair, "GIT_") {
			real.Env = append(real.Env, pair)
		}
	}
	realCommon := miniMaxGitProbe(t, &real, "rev-parse", "--path-format=absolute", "--git-common-dir")
	declared := miniMaxEnvValue(cmd.Env, "GIT_DIR")
	ev.Record("private-claim-namespace", map[string]any{"declared": declared, "common": common, "realCommon": realCommon, "vendorCalls": 0})
	return ev.Check("synthetic claims use a declared private common Git store", declared != "" && common == declared && common != realCommon, "declared=%s common=%s real=%s", declared, common, realCommon)
}
