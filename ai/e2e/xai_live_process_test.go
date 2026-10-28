package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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
	for _, name := range []string{"PATH", "TMPDIR", "GOCACHE", "GOPATH", "GOROOT", "GIT_DIR", "GIT_WORK_TREE"} {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}

func TestXAILiveHarnessProcess(t *testing.T) {
	privateGrantStore(t)
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
		out, makeErr := os.MkdirTemp(run.Dir(), "xai-child-"+name+"-")
		if makeErr != nil {
			t.Fatal(makeErr)
		}
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
		nativeFiles := map[string]string{}
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
			fixtureName := "native-" + strings.ReplaceAll(relative, string(filepath.Separator), "__")
			nativeFiles[filepath.ToSlash(relative)] = fixtureName
			cs.Fixture(fixtureName, data)
			return nil
		})
		cs.Record("native-file-map", nativeFiles)
		if name == "finite-controlled-plan" {
			// The live JPEG rejection reported a 512-total-pixel minimum.
			// Check real scenario payloads, including projected tool images.
			for capture, wantCount := range map[string]int{"E17-xai-user-png/image-http.json": 1, "E17-xai-user-jpeg/image-http.json": 1, "E17-xai-tool-result-image/results-http.json": 2} {
				data, readErr := os.ReadFile(filepath.Join(filepath.Dir(files[0]), capture))
				var attempts []struct {
					RequestBody string `json:"requestBody"`
				}
				decodeErr := json.Unmarshal(data, &attempts)
				if !cs.Check("image scenario capture available "+capture, readErr == nil && decodeErr == nil && len(attempts) == 1, "missing or invalid capture") {
					continue
				}
				var request struct {
					Input []map[string]any `json:"input"`
				}
				if !cs.Check("image request decodes "+capture, json.Unmarshal([]byte(attempts[0].RequestBody), &request) == nil, "invalid captured request") {
					continue
				}
				count := 0
				for _, item := range request.Input {
					content, _ := item["content"].([]any)
					for _, value := range content {
						block, _ := value.(map[string]any)
						if block["type"] != "input_image" {
							continue
						}
						count++
						url, _ := block["image_url"].(string)
						header, encoded, ok := strings.Cut(url, ",")
						raw, decodeErr := base64.StdEncoding.Strict().DecodeString(encoded)
						config, format, imageErr := image.DecodeConfig(bytes.NewReader(raw))
						valid := ok && decodeErr == nil && imageErr == nil && header == "data:image/"+format+";base64" && config.Width*config.Height >= 512
						cs.Check("live image meets captured pixel minimum "+capture, valid, "format=%s dimensions=%dx%d", format, config.Width, config.Height)
					}
				}
				cs.Check("all planned inline images captured "+capture, count == wantCount, "images=%d want=%d", count, wantCount)
			}
		}
		if name == "final-write" {
			info, statErr := os.Stat(filepath.Join(filepath.Dir(files[0]), "xai-candidate-report.pending.json"))
			cs.Check("filesystem refused final candidate write", statErr == nil && info.IsDir(), "expected retained filesystem failure target")
		}
		if name == "final-sync" || name == "final-audit" {
			var fault struct {
				Step        string `json:"step"`
				ClosedFile  bool   `json:"closedFile"`
				IsDirectory bool   `json:"isDirectory"`
			}
			faultRaw, faultErr := os.ReadFile(filepath.Join(filepath.Dir(files[0]), "xai-finalization-failure.json"))
			decodeErr := json.Unmarshal(faultRaw, &fault)
			wantStep, actualFault := "directory-sync", fault.ClosedFile
			if name == "final-audit" {
				wantStep, actualFault = "final-audit", fault.IsDirectory
			}
			cs.Check("retained actual filesystem failure", faultErr == nil && decodeErr == nil && fault.Step == wantStep && actualFault, "failure=%s", faultRaw)
			var audit struct {
				Findings []json.RawMessage `json:"findings"`
			}
			auditRaw, readErr := os.ReadFile(filepath.Join(filepath.Dir(files[0]), "audit.json"))
			decodeErr = json.Unmarshal(auditRaw, &audit)
			cs.Check("failed report has complete final redaction audit", readErr == nil && decodeErr == nil && audit.Findings != nil && len(audit.Findings) == 0, "audit unavailable or failed")
			priorRaw, readErr := os.ReadFile(filepath.Join(filepath.Dir(files[0]), "xai-candidate-report.unfinalized.json"))
			var prior map[string]any
			decodeErr = json.Unmarshal(priorRaw, &prior)
			cs.Check("original nonPASS report retained", readErr == nil && decodeErr == nil && prior["outcome"] == "RUNNING" && prior["complete"] == false && prior["auditPassed"] == false, "prior=%s", priorRaw)
		}
		if name == "finite-controlled-plan" || name == "final-sync" || name == "final-audit" {
			verifyOut := filepath.Join(run.Dir(), "xai-controlled-native-consumer-"+name)
			rootPath, err := filepath.Abs("../..")
			if err != nil {
				t.Fatal(err)
			}
			relativeBundle, err := filepath.Rel(rootPath, filepath.Dir(files[0]))
			if err != nil {
				t.Fatal(err)
			}
			verify := exec.CommandContext(ctx, consumer, "-live-bundle", filepath.ToSlash(relativeBundle), "-out", verifyOut)
			verify.Dir = rootPath
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
			proofRaw, readErr := os.ReadFile(filepath.Join(verifyOut, "live-consumer.json"))
			var proof struct {
				Pass     bool              `json:"pass"`
				Consumer string            `json:"consumer"`
				Error    string            `json:"error"`
				Inputs   map[string]string `json:"inputs"`
			}
			decodeErr := json.Unmarshal(proofRaw, &proof)
			candidateRaw, _ := os.ReadFile(files[0])
			if readErr == nil {
				cs.Fixture("formal-consumer-native.json", proofRaw)
			}
			cs.Check("native formal consumer rejects candidate controlled proof", verifyCode == 1 && readErr == nil && decodeErr == nil && !proof.Pass && proof.Consumer == "releasegate/native-live-v1" && proof.Error == "native live consumer rejected evidence" && proof.Inputs["xai-candidate-report.json"] == evidence.Digest(candidateRaw), "candidate rejection receipt missing or invalid")
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
	for _, fault := range []string{"final-sync", "final-audit"} {
		t.Run(fault, func(t *testing.T) {
			p := write(t, grant(t, "controlled"))
			r, code := child(t, fault, "^TestXAICandidate$", []string{"BARNESS_XAI_CONTROLLED=" + fault, "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + p})
			if r == nil {
				return
			}
			if code != 1 || r["outcome"] == "PASS" || r["complete"] != false || r["auditPassed"] != false || r["claimId"] == "" || r["budget"].(map[string]any)["httpAttempts"] != float64(24) {
				t.Errorf("late filesystem failure must retain consumed nonPASS evidence: code=%d report=%v", code, r)
			}
		})
	}

}
