package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
)

// Formal process cases own their publication assertions independently from the
// candidate driver. Both use the already agreed public compiled-process seam.
func TestXAIProposedBuiltinCatalog(t *testing.T) {
	cs := run.Case(t, "XAI-E01-proposed-builtin-snapshot")
	cat := ai.BuiltinCatalog()
	m, ok := cat.Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, "grok-4.7")
	cs.Record("proposed-catalog", cat)
	cs.Check("reviewable final catalog contains the fixed own model", ok && reflect.DeepEqual(m, xaiFoundationModel()), "proposed xAI facts missing or different")
	candidate := xai.CandidateCatalog()
	before, _ := candidate.Hash()
	builtinHash, err := cat.Hash()
	cs.Check("final and candidate snapshots stay distinct", err == nil && candidate.Version != cat.Version && before != builtinHash, "catalog snapshots alias")
	if ok {
		m.Input[0] = ai.ModalityImage
		m.Cost.Tiers[0].Input = 900
		fresh, _ := ai.BuiltinCatalog().Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, "grok-4.7")
		cs.Check("Builtin views cannot mutate final facts", fresh.Cost.Tiers[0].Input == 4 && fresh.Input[0] == ai.ModalityText, "mutable catalog leaked")
	}
}

func TestXAIFormalHarnessProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cs := run.Case(t, "E17-xai-formal-process-build")
	binary := filepath.Join(t.TempDir(), "xai-formal.test")
	build := exec.CommandContext(ctx, "go", "test", "-c", "-tags=live", "-o", binary, "./ai/live")
	build.Dir = "../.."
	build.Env = xaiProcessEnv(t)
	output, err := build.CombinedOutput()
	cs.Record("build", map[string]any{"argv": build.Args, "exitCode": xaiFormalExit(err), "output": string(output)})
	if !cs.Check("isolated formal binary builds", err == nil, "build failed") {
		return
	}
	binaryBytes, readErr := os.ReadFile(binary)
	if !cs.Check("actual formal binary is retained by digest", readErr == nil, "binary metadata unavailable") {
		return
	}
	binaryHash := evidence.Digest(binaryBytes)
	cs.Record("binary", map[string]any{"sha256": binaryHash, "bytes": len(binaryBytes)})
	baseline, code, _ := xaiFormalChild(t, ctx, binary, "no-switch", "^TestXAIFormal$", nil)
	if baseline == nil {
		return
	}
	cs.Check("formal no-switch is explicit NOT_RUN", code == 0 && baseline["outcome"] == "NOT_RUN", "unexpected formal no-switch outcome")
	cs.Check("native bootstrap identifies the actual built binary", baseline["binarySha256"] == binaryHash, "binary identity differs")
	t.Run("controlled-plan", func(t *testing.T) {
		path := xaiWriteFormalGrant(t, xaiFormalGrant(t, baseline))
		r, code, dir := xaiFormalChild(t, ctx, binary, "controlled-plan", "^TestXAIFormal$", []string{"BARNESS_XAI_CONTROLLED=fixture", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + path})
		if r == nil {
			return
		}
		caseProof := run.Case(t, "E17-xai-formal-projection")
		b := r["budget"].(map[string]any)
		caseProof.Check("complete own bounded controlled plan", code == 0 && r["outcome"] == "PASS" && r["complete"] == true && b["callsUsed"] == float64(24) && b["httpAttempts"] == float64(24), "incomplete own plan")
		raw, err := os.ReadFile(filepath.Join(dir, "live-report.json"))
		var projected map[string]any
		decodeErr := json.Unmarshal(raw, &projected)
		if !caseProof.Check("standard formal projection exists", err == nil && decodeErr == nil, "standard projection unavailable") {
			return
		}
		pb, ok := projected["budget"].(map[string]any)
		caseProof.Check("standard budget preserves inclusive normalized accounting", ok && pb["maxCalls"] == float64(24) && pb["maxAttempts"] == float64(24) && pb["maxRetriesPerScenario"] == float64(0) && pb["maxOutputTokensPerCall"] == float64(1024) && pb["callsUsed"] == float64(24) && pb["httpAttempts"] == float64(24) && pb["inputTokensUsed"] == float64(768) && pb["outputTokensUsed"] == float64(26616) && b["reportedTokens"] == float64(27384), "standard budget differs")
		caseProof.Check("standard projection preserves all own scenes and pins", projected["schema"] == float64(3) && projected["combo"] == "xai-responses" && projected["catalogHash"] == r["catalogHash"] && projected["gitCommit"] == r["gitCommit"] && reflect.DeepEqual(projected["expected"], r["expected"]) && reflect.DeepEqual(projected["scenarios"], r["scenarios"]), "projection differs")
		for _, scene := range []string{"entry-stream", "entry-stream-simple"} {
			raw, err := os.ReadFile(filepath.Join(dir, "E17-xai-"+scene, "text-events.json"))
			var events []map[string]json.RawMessage
			decodeErr := json.Unmarshal(raw, &events)
			valid := err == nil && decodeErr == nil && len(events) > 1
			if valid {
				valid = string(events[0]["type"]) == `"start"` && string(events[len(events)-1]["type"]) == `"done"` && len(events[len(events)-1]["Message"]) > 0
			}
			caseProof.Check("formal event artifact keeps explicit lifecycle "+scene, valid, "typed formal event lifecycle unavailable")
		}

	})
	for _, name := range []string{"no-key", "no-grant", "wrong-stage", "wrong-mode", "candidate-catalog", "source", "binary", "sdk", "endpoint", "model", "expiry", "budget", "alias", "null-usd", "field-case", "duplicate-field", "mixed-credentials"} {
		t.Run(name, func(t *testing.T) {
			g := xaiFormalGrant(t, baseline)
			env := []string{"BARNESS_XAI_CONTROLLED=fixture", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global"}
			want := "FAIL"
			switch name {
			case "no-key":
				env = []string{"BARNESS_AI_LIVE=1"}
				want = "NOT_RUN"
			case "no-grant":
				env = []string{"BARNESS_AI_LIVE=1", "BARNESS_AI_LIVE_KEY_XAI_RESPONSES=synthetic-formal-only"}
				want = "NOT_RUN"
			case "wrong-stage":
				g["stage"] = "candidate"
			case "wrong-mode":
				g["mode"] = "live"
			case "candidate-catalog":
				cat := xai.CandidateCatalog()
				hash, _ := cat.Hash()
				g["catalogVersion"], g["catalogHash"] = cat.Version, hash
			case "source":
				g["sourceSha256"] = strings.Repeat("0", 64)
			case "binary":
				g["binarySha256"] = strings.Repeat("0", 64)
			case "sdk":
				g["sdk"] = "substituted-sdk"
			case "endpoint":
				g["endpoint"] = "https://us-east-1.api.x.ai/v1"
			case "model":
				g["model"] = "grok-4"
			case "expiry":
				g["expiresAt"] = time.Now().Add(-time.Hour).Format(time.RFC3339Nano)
			case "budget":
				g["maxAttempts"] = 25
			case "alias":
				g["regionAlias"] = "other"
			case "null-usd":
				g["maxUSD"] = nil
			case "field-case":
				g["Schema"] = g["schema"]
				delete(g, "schema")
			case "mixed-credentials":
				env = append(env, "OPENAI_API_KEY=synthetic-other-route")
			}
			if name != "no-key" && name != "no-grant" {
				path := xaiWriteFormalGrant(t, g)
				if name == "duplicate-field" {
					raw, _ := os.ReadFile(path)
					raw = append([]byte(`{"maxCalls":24,`), raw[1:]...)
					if err := os.WriteFile(path, raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
				env = append(env, "BARNESS_AI_LIVE_PERMISSIONS_FILE="+path)
			}
			r, code, _ := xaiFormalChild(t, ctx, binary, name, "^TestXAIFormal$", env)
			if r == nil {
				return
			}
			proof := run.Case(t, "E17-xai-formal-gate-"+name)
			b := r["budget"].(map[string]any)
			proof.Check("formal gate refuses before HTTP", r["outcome"] == want && (code == 0) == (want == "NOT_RUN") && b["httpAttempts"] == float64(0) && b["callsUsed"] == float64(0), "formal gate outcome differs")
		})
	}
	t.Run("claim-cancel-reuse", func(t *testing.T) {
		path := xaiWriteFormalGrant(t, xaiFormalGrant(t, baseline))
		env := []string{"BARNESS_XAI_CONTROLLED=cancel", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + path}
		r, code, _ := xaiFormalChild(t, ctx, binary, "cancel", "^TestXAIFormal$", env)
		if r == nil {
			return
		}
		proof := run.Case(t, "E17-xai-formal-claim-on-failure")
		b := r["budget"].(map[string]any)
		proof.Check("cancellation consumes one grant and one attempt without retry", code != 0 && r["outcome"] == "FAIL" && r["claimId"] != "" && b["callsUsed"] == float64(1) && b["httpAttempts"] == float64(1), "cancellation claim missing")
		r, code, _ = xaiFormalChild(t, ctx, binary, "claim-reuse", "^TestXAIFormal$", env)
		if r != nil {
			proof.Check("consumed formal grant cannot be reused", code != 0 && r["budget"].(map[string]any)["httpAttempts"] == float64(0), "grant was reused")
		}
	})
	for _, mode := range []string{"final-write", "projection-write", "final-sync", "final-audit"} {
		t.Run(mode, func(t *testing.T) {
			path := xaiWriteFormalGrant(t, xaiFormalGrant(t, baseline))
			r, code, dir := xaiFormalChild(t, ctx, binary, mode, "^TestXAIFormal$", []string{"BARNESS_XAI_CONTROLLED=" + mode, "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + path})
			if r == nil {
				return
			}
			proof := run.Case(t, "E17-xai-formal-finalization-"+mode)
			b := r["budget"].(map[string]any)
			proof.Check("late formal I/O failure preserves non-PASS budget", code != 0 && r["outcome"] == "FAIL" && r["complete"] == false && r["auditPassed"] == false && b["callsUsed"] == float64(24) && b["httpAttempts"] == float64(24), "late failure lost budget or retained PASS")
			_, projectionErr := os.Stat(filepath.Join(dir, "live-report.json"))
			proof.Check("late failure unpublishes generic successful projection", os.IsNotExist(projectionErr), "generic formal projection survived")
			var fault map[string]any
			raw, readErr := os.ReadFile(filepath.Join(dir, "xai-finalization-failure.json"))
			decodeErr := json.Unmarshal(raw, &fault)
			steps := map[string]string{"final-write": "pending-write", "projection-write": "projection-pending-write", "final-sync": "directory-sync", "final-audit": "final-audit"}
			proof.Check("retained actual filesystem failure step", readErr == nil && decodeErr == nil && fault["step"] == steps[mode], "formal failure proof missing")
			var prior map[string]any
			raw, readErr = os.ReadFile(filepath.Join(dir, "xai-formal-report.unfinalized.json"))
			decodeErr = json.Unmarshal(raw, &prior)
			proof.Check("original pre-finalized report remains non-PASS", readErr == nil && decodeErr == nil && prior["outcome"] == "RUNNING" && prior["complete"] == false, "original report lost")
		})
	}

}

func xaiFormalExit(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(*exec.ExitError); ok {
		return e.ExitCode()
	}
	return -1
}

func xaiFormalChild(t *testing.T, ctx context.Context, binary, name, entry string, env []string) (map[string]any, int, string) {
	t.Helper()
	cs := run.Case(t, "E17-xai-formal-process-"+name)
	out, err := os.MkdirTemp(run.Dir(), "xai-formal-child-"+name+"-")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, binary, "-test.run="+entry, "-test.timeout=1800s")
	cmd.Dir = "../.."
	cmd.Env = append(xaiProcessEnv(t), "BARNESS_AI_LIVE_COMBO=xai-responses", "BARNESS_AI_EVIDENCE_DIR="+out)
	cmd.Env = append(cmd.Env, env...)
	output, runErr := cmd.CombinedOutput()
	code := xaiFormalExit(runErr)
	cs.Record("process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(output)})
	files, err := filepath.Glob(filepath.Join(out, "*", "xai-formal-report.json"))
	if !cs.Check("own formal native report exists", err == nil && len(files) == 1, "reports=%d", len(files)) {
		return nil, code, ""
	}
	dir := filepath.Dir(files[0])
	err = filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, e := filepath.Rel(dir, path)
		if e != nil {
			return e
		}
		raw, e := os.ReadFile(path)
		if e == nil {
			cs.Fixture("native-"+strings.ReplaceAll(filepath.ToSlash(relative), "/", "__"), raw)
		}
		return e
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(files[0])
	var report map[string]any
	if err != nil || json.Unmarshal(raw, &report) != nil {
		t.Fatal("formal report is unreadable")
	}
	cs.Check("formal identity stays explicit", report["kind"] == "xai-formal" && report["stage"] == "formal" && report["combo"] == "xai-responses", "wrong report identity")
	budget, ok := report["budget"].(map[string]any)
	cs.Check("ordinary children never call a vendor", ok && budget["vendorHTTPAttempts"] == float64(0), "vendor transport escaped")
	cs.Check("ordinary formal shape is never a real qualification", report["formalQualified"] == false, "controlled qualification escaped")
	return report, code, dir
}

func xaiFormalGrant(t *testing.T, baseline map[string]any) map[string]any {
	t.Helper()
	g := map[string]any{"schema": 1, "id": fmt.Sprintf("formal-controlled-%d", time.Now().UnixNano()), "mode": "controlled", "stage": "formal", "combo": "xai-responses", "provider": "xai", "api": "openai-responses", "endpoint": "https://api.x.ai/v1", "model": "grok-4.7", "accountAlias": "synthetic", "regionAlias": "global", "expiresAt": time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), "maxCalls": 24, "maxAttempts": 24, "maxRequestedOutputTokens": 24576, "maxReservedInputTokens": 196608, "maxUSD": 1.0}
	for _, key := range []string{"sourceSha256", "binarySha256", "catalogVersion", "catalogHash", "sdk", "expected"} {
		g[key] = baseline[key]
	}
	return g
}

func xaiWriteFormalGrant(t *testing.T, grant map[string]any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "grant.json")
	raw, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
