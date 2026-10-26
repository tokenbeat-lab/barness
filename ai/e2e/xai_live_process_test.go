package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// The agreed public seam is the exact live binary entry and its durable report.
// Every child has a fresh environment and HOME; these checks cannot call a vendor.
func xaiProcessEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"GOENV=off", "GOFLAGS=", "HOME=" + t.TempDir()}
	for _, name := range []string{"PATH", "TMPDIR", "GOCACHE", "GOPATH", "GOROOT"} {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}

func TestXAILiveHarnessProcess(t *testing.T) {
	buildCase := run.Case(t, "E17-xai-process-build")
	exitCode := func(err error) int {
		if err == nil {
			return 0
		}
		if ex, ok := err.(*exec.ExitError); ok {
			return ex.ExitCode()
		}
		return -1
	}
	binary := filepath.Join(t.TempDir(), "xai-live.test")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "test", "-c", "-tags=live", "-o", binary, "./ai/live")
	build.Dir = "../.."
	build.Env = xaiProcessEnv(t)
	output, err := build.CombinedOutput()
	buildCase.Record("live-build", map[string]any{"argv": build.Args, "exitCode": exitCode(err), "output": string(output)})
	if err != nil {
		t.Fatalf("isolated compile: %v %s", err, output)
	}
	consumer := filepath.Join(t.TempDir(), "releasegate")
	consumerBuild := exec.CommandContext(ctx, "go", "build", "-o", consumer, "./ai/release/cmd/releasegate")
	consumerBuild.Dir = "../.."
	consumerBuild.Env = xaiProcessEnv(t)
	output, err = consumerBuild.CombinedOutput()
	buildCase.Record("consumer-build", map[string]any{"argv": consumerBuild.Args, "exitCode": exitCode(err), "output": string(output)})
	if err != nil {
		t.Fatalf("isolated consumer compile: %v %s", err, output)
	}
	for name, path := range map[string]string{"live-binary": binary, "consumer-binary": consumer} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		buildCase.Record(name, map[string]any{"sha256": evidence.Digest(raw), "bytes": len(raw)})
	}

	child := func(t *testing.T, name, entry string, env []string) (map[string]any, int) {
		t.Helper()
		cs := run.Case(t, "E17-xai-process-"+name)
		out := t.TempDir()
		cmd := exec.CommandContext(ctx, binary, "-test.run="+entry, "-test.timeout=30s")
		cmd.Dir = "../.."
		cmd.Env = append(xaiProcessEnv(t), "BARNESS_AI_LIVE_COMBO=xai-responses", evidence.DirEnv+"="+out)
		cmd.Env = append(cmd.Env, env...)
		output, e := cmd.CombinedOutput()
		code := 0
		if e != nil {
			code = -1
			if ex, ok := e.(*exec.ExitError); ok {
				code = ex.ExitCode()
			}
		}
		cs.Record("process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(output)})
		files, _ := filepath.Glob(filepath.Join(out, "*", "xai-candidate-report.json"))
		if !cs.Check("own candidate report exists", len(files) == 1, "reports=%d", len(files)) {
			return nil, code
		}
		filepath.WalkDir(filepath.Dir(files[0]), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				t.Fatal(walkErr)
			}
			if entry.IsDir() {
				return nil
			}
			relative, _ := filepath.Rel(filepath.Dir(files[0]), path)
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			cs.Fixture("native-"+strings.ReplaceAll(relative, string(filepath.Separator), "__"), data)
			return nil
		})
		if name == "finite-controlled-plan" {
			verify := exec.CommandContext(ctx, consumer, "-live-bundle", filepath.Dir(files[0]), "-out", t.TempDir())
			verify.Dir = "../.."
			verify.Env = xaiProcessEnv(t)
			output, verifyErr := verify.CombinedOutput()
			verifyCode := 0
			if verifyErr != nil {
				verifyCode = -1
				if ex, ok := verifyErr.(*exec.ExitError); ok {
					verifyCode = ex.ExitCode()
				}
			}
			cs.Record("formal-consumer", map[string]any{"argv": verify.Args, "exitCode": verifyCode, "output": string(output)})
			cs.Check("native formal consumer rejects candidate controlled proof", verifyCode > 0, "candidate obtained formal qualification")
		}

		raw, _ := os.ReadFile(files[0])
		cs.Fixture("child-report", raw)
		var r map[string]any
		if e = json.Unmarshal(raw, &r); e != nil {
			t.Fatal(e)
		}
		cs.Check("candidate never formal", r["formalQualified"] == false && r["stage"] == "candidate", "formal promotion escaped")
		b := r["budget"].(map[string]any)
		cs.Check("ordinary process has zero vendor HTTP", b["vendorHTTPAttempts"] == float64(0), "vendor transport escaped")
		return r, code
	}
	baseline, baselineCode := child(t, "snapshot", "^TestXAICandidate$", nil)
	if baseline == nil || baselineCode != 0 || baseline["outcome"] != "NOT_RUN" {
		t.Fatal("candidate snapshot unavailable")
	}
	t.Run("no-switch", func(t *testing.T) {
		r, code := child(t, "no-switch", "^TestXAICandidate$", nil)
		if r != nil && (code != 0 || r["outcome"] != "NOT_RUN") {
			t.Errorf("missing switch: %v", r)
		}
	})
	if baseline == nil {
		return
	}
	grant := func(t *testing.T, mode string) map[string]any {
		g := map[string]any{"schema": 1, "id": fmt.Sprintf("controlled-%d", time.Now().UnixNano()), "mode": mode, "stage": "candidate", "combo": "xai-responses", "provider": "xai", "api": "openai-responses", "endpoint": "https://api.x.ai/v1", "model": "grok-4.7", "accountAlias": "synthetic", "regionAlias": "global", "expiresAt": time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), "maxCalls": 24, "maxAttempts": 24, "maxRequestedOutputTokens": 24576, "maxReservedInputTokens": 196608, "maxUSD": 1.0}
		for _, k := range []string{"sourceSha256", "binarySha256", "catalogVersion", "catalogHash", "sdk", "expected"} {
			g[k] = baseline[k]
		}
		return g
	}
	write := func(t *testing.T, g map[string]any) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "grant.json")
		b, _ := json.Marshal(g)
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
		return p
	}
	for _, name := range []string{"no-key", "no-grant", "entry", "identity", "selected-route", "endpoint", "model", "alias", "alias-mismatch", "budget", "source", "catalog", "binary", "expiry", "mixed-credentials", "mixed-route", "invalid-proxy", "null-budget", "zero-usd", "field-case", "duplicate-field"} {
		t.Run(name, func(t *testing.T) {
			g := grant(t, "controlled")
			env := []string{"BARNESS_XAI_CONTROLLED=cancel", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global"}
			entry := "^TestXAICandidate$"
			want := "FAIL"
			switch name {
			case "no-key":
				env = []string{"BARNESS_AI_LIVE=1"}
				want = "NOT_RUN"
			case "no-grant":
				env = []string{"BARNESS_AI_LIVE=1", "BARNESS_AI_LIVE_KEY_XAI_RESPONSES=synthetic-xai-only"}
				want = "NOT_RUN"
			case "entry":
				entry = "TestXAI"
			case "selected-route":
				env = append(env, "BARNESS_AI_LIVE_COMBO=openai-responses")
			case "identity":
				g["provider"] = "openai"
			case "endpoint":
				g["endpoint"] = "https://us-east-1.api.x.ai/v1"
			case "model":
				g["model"] = "grok-4"
			case "alias":
				g["accountAlias"] = "../account"
			case "alias-mismatch":
				g["regionAlias"] = "other"
			case "budget":
				g["maxAttempts"] = 25
			case "source":
				g["sourceSha256"] = "changed"
			case "catalog":
				g["catalogHash"] = "changed"
			case "binary":
				g["binarySha256"] = "changed"
			case "expiry":
				g["expiresAt"] = time.Now().Add(-time.Hour).Format(time.RFC3339Nano)
			case "mixed-credentials":
				env = append(env, "OPENAI_API_KEY=synthetic-other-route")
			case "mixed-route":
				env = append(env, "BARNESS_REGIONAL_ROUTE=zai-responses")
			case "invalid-proxy":
				env = append(env, "BARNESS_AI_LIVE_PROXY=invalid")
			case "zero-usd":
				g["maxUSD"] = 0
			case "field-case":
				g["Schema"] = g["schema"]
				delete(g, "schema")
			case "null-budget":
				g["maxUSD"] = nil
			}
			if name != "no-key" && name != "no-grant" {
				p := write(t, g)
				if name == "duplicate-field" {
					b, _ := os.ReadFile(p)
					b = append([]byte(`{"maxCalls":24,`), b[1:]...)
					os.WriteFile(p, b, 0600)
				}
				env = append(env, "BARNESS_AI_LIVE_PERMISSIONS_FILE="+p)
			}
			r, code := child(t, name, entry, env)
			if r == nil {
				return
			}
			b := r["budget"].(map[string]any)
			if r["outcome"] != want || (code == 0) != (want == "NOT_RUN") || b["httpAttempts"] != float64(0) {
				t.Errorf("gate: code=%d report=%v", code, r)
			}
		})
	}
	t.Run("cancel-and-claim-reuse", func(t *testing.T) {
		p := write(t, grant(t, "controlled"))
		env := []string{"BARNESS_XAI_CONTROLLED=cancel", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + p}
		r, code := child(t, "cancel", "^TestXAICandidate$", env)
		if r == nil {
			return
		}
		b := r["budget"].(map[string]any)
		if code == 0 || r["outcome"] != "FAIL" || r["claimId"] == "" || b["httpAttempts"] != float64(1) {
			t.Errorf("cancel must retain one claimed attempt: %v", r)
		}
		r, code = child(t, "claim-reuse", "^TestXAICandidate$", env)
		if r != nil && (code == 0 || r["budget"].(map[string]any)["httpAttempts"] != float64(0)) {
			t.Errorf("claim reused: %v", r)
		}
	})
	t.Run("finite-controlled-plan", func(t *testing.T) {
		p := write(t, grant(t, "controlled"))
		r, code := child(t, "finite-controlled-plan", "^TestXAICandidate$", []string{"BARNESS_XAI_CONTROLLED=fixture", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + p})
		if r == nil {
			return
		}
		b := r["budget"].(map[string]any)
		if code != 0 || r["outcome"] != "PASS" || r["complete"] != true || len(r["scenarios"].([]any)) != 18 || b["callsUsed"] != float64(24) || b["httpAttempts"] != float64(24) || b["reportedTokens"] != float64(3624) {
			t.Errorf("finite own plan incomplete: code=%d report=%v", code, r)
		}
	})

	t.Run("final-write", func(t *testing.T) {
		p := write(t, grant(t, "controlled"))
		r, code := child(t, "final-write", "^TestXAICandidate$", []string{"BARNESS_XAI_CONTROLLED=final-write", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + p})
		if r != nil && (code == 0 || r["outcome"] == "PASS" || r["complete"] == true || r["budget"].(map[string]any)["httpAttempts"] != float64(24)) {
			t.Errorf("failed finalization passed: %v", r)
		}
	})
}
