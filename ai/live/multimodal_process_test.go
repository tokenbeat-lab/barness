//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The probe executable and the independently read bundle are the public seams.
// Children receive a synthetic credential only, never the parent's environment.
func TestMultimodalCandidateProcess(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, fault := range []string{"success", "missing-key", "switch-off", "missing-proof", "foreign-key", "foreign-token", "wrong-service", "wrong-model", "wrong-catalog", "calls-budget", "attempts-budget", "questions-budget", "sources-budget", "references-budget", "source-bytes-budget", "expanded-bytes-budget", "state-budget", "body-budget", "deadline", "process-deadline", "429", "malformed", "audit", "manifest", "inventory"} {
			t.Run(route+"/"+fault, func(t *testing.T) {
				cs := run.Case(t, "MM11-process-"+route+"-"+fault)
				base := t.TempDir()
				proof := filepath.Join(base, "proof.json")
				model, service := "jev-1.14.0", "https://api.typesafe.ai/v1/systemone"
				if route == "decisions" {
					model, service = "gpt-6-luna", "https://api.openai.com/v1/decisions"
				}
				if fault == "wrong-model" {
					model = "unapproved-model"
				}
				if fault == "wrong-service" {
					service = "https://unapproved.invalid/classify"
				}
				raw, _ := json.Marshal(map[string]any{"route": route, "model": model, "service": service, "accountAlias": "controlled", "regionAlias": "offline", "authorizationReference": "controlled-process-fixture", "dedicatedCredential": true, "maxCalls": 12, "maxAttempts": 12})
				if err := os.WriteFile(proof, raw, 0600); err != nil {
					t.Fatal(err)
				}
				env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TMPDIR=" + os.TempDir(), "BARNESS_MM_CONTROLLED=" + fault, "BARNESS_MM_ROUTE=" + route, "BARNESS_MM_LIVE=1", "BARNESS_MM_KEY=controlled-mm-credential", "BARNESS_MM_AUTHORIZATION=" + proof, "BARNESS_AI_EVIDENCE_DIR=" + base}
				for i, kv := range env {
					if (fault == "missing-key" && strings.HasPrefix(kv, "BARNESS_MM_KEY=")) || (fault == "missing-proof" && strings.HasPrefix(kv, "BARNESS_MM_AUTHORIZATION=")) || (fault == "switch-off" && strings.HasPrefix(kv, "BARNESS_MM_LIVE=")) {
						env[i] = strings.SplitN(kv, "=", 2)[0] + "="
					}
				}
				if fault == "foreign-key" {
					env = append(env, "OPENAI_API_KEY=foreign-controlled-credential")
				}
				if fault == "foreign-token" {
					env = append(env, "GOOGLE_OAUTH_ACCESS_TOKEN=foreign-controlled-credential")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMultimodalCandidate$", "-test.timeout=25s")
				if fault == "process-deadline" {
					cmd.Args[len(cmd.Args)-1] = "-test.timeout=12s"
				}
				cmd.Env = env
				output, childErr := cmd.CombinedOutput()
				paths, _ := filepath.Glob(filepath.Join(base, "*", "multimodal-report.json"))
				if len(paths) != 1 {
					t.Fatalf("probe report missing: %v\n%s", childErr, output)
				}
				var report struct {
					Outcome, Mode string
					Complete      bool
					Budget        struct{ CallsUsed, HTTPAttempts int }
				}
				data, err := os.ReadFile(paths[0])
				if err != nil {
					t.Fatal(err)
				}
				if err = json.Unmarshal(data, &report); err != nil {
					t.Fatal(err)
				}
				cs.Record("child-report", report)
				want := "FAIL"
				if fault == "success" {
					want = "PASS"
				}
				if fault == "missing-key" || fault == "switch-off" || fault == "missing-proof" {
					want = "NOT_RUN"
				}
				cs.Check("honest outcome and exit", report.Outcome == want && (childErr != nil) == (want == "FAIL"), "outcome=%s exit=%v: %s", report.Outcome, childErr, output)
				cs.Check("controlled is never live evidence", report.Mode == "controlled" && report.Complete == (want == "PASS"), "wrong provenance")
				cs.Check("finite spend even on failures", report.Budget.CallsUsed <= 12 && report.Budget.HTTPAttempts <= 12, "spend cap exceeded")
				if fault == "process-deadline" {
					cs.Check("in-flight deadline retains spend and finalization", report.Budget.CallsUsed == 1 && report.Budget.HTTPAttempts == 1 && ctx.Err() == nil, "deadline evidence lost")
				}
				if want == "NOT_RUN" || strings.HasPrefix(fault, "wrong-") || (fault == "foreign-key" || fault == "foreign-token") {
					cs.Check("configuration stops before spend", report.Budget.CallsUsed == 0 && report.Budget.HTTPAttempts == 0, "configuration spent budget")
				}
				root := filepath.Dir(paths[0])
				_, formalErr := supportmatrix.LoadReport(paths[0])
				cs.Check("formal importer rejects candidate", formalErr != nil, "candidate entered formal matrix")
				_, formalErr = os.Stat(filepath.Join(root, "live-report.json"))
				cs.Check("no formal report published", os.IsNotExist(formalErr), "formal report emitted")
				verify := exec.CommandContext(ctx, "python3", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py", root)
				verified, verifyErr := verify.CombinedOutput()
				cs.Record("verification", map[string]any{"output": string(verified), "error": errText(verifyErr)})
				cs.Check("independent gate accepts only complete controlled pipeline", (verifyErr == nil) == (want == "PASS"), "%s", verified)
				liveVerify := exec.CommandContext(ctx, "python3", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py", "--require-live", root)
				cs.Check("controlled package cannot prove service PASS", liveVerify.Run() != nil, "controlled promoted")
				if fault == "success" {
					verifyMultimodalTampering(t, root, route)
				}
				err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					rel, _ := filepath.Rel(root, path)
					name := strings.ReplaceAll(rel, string(os.PathSeparator), "__")
					if strings.HasSuffix(name, ".json") && !json.Valid(data) {
						name += ".bin"
					}
					cs.Fixture(name, data)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

// Rewrite checksums too: a self-consistent damaged package still needs semantic
// rejection. These are executable bundle tests, not private-function tests.
func verifyMultimodalTampering(t *testing.T, root, route string) {
	t.Helper()
	for _, fault := range []string{"counter", "narrowed", "audit-status", "final-image", "legend-index", "missing-assertions"} {
		if route != "jev" && fault == "legend-index" {
			continue
		}
		t.Run("tamper-"+fault, func(t *testing.T) {
			cs := run.Case(t, "MM11-tamper-"+route+"-"+fault)
			name := "multimodal-report.json"
			switch fault {
			case "final-image":
				name = "MM11-default-context-plain/result.json"
			case "legend-index":
				name = "MM11-visual-criteria-hooked/result.json"
			case "missing-assertions":
				name = "MM11-default-context-plain/assertions.json"
			}
			path := filepath.Join(root, name)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			inventoryPath := filepath.Join(root, "multimodal-integrity.json")
			originalInventory, err := os.ReadFile(inventoryPath)
			if err != nil {
				t.Fatal(err)
			}
			defer os.WriteFile(path, original, 0600)
			defer os.WriteFile(inventoryPath, originalInventory, 0600)
			var doc map[string]any
			if err = json.Unmarshal(original, &doc); err != nil {
				t.Fatal(err)
			}
			switch fault {
			case "counter":
				doc["budget"].(map[string]any)["httpAttempts"] = 0
			case "narrowed":
				doc["expected"] = []any{"default-context-plain"}
			case "audit-status":
				doc["auditPassed"] = false
			case "final-image":
				doc["result"].(map[string]any)["imageInputs"].([]any)[0].(map[string]any)["sha256"] = strings.Repeat("0", 64)
			case "legend-index":
				doc["result"].(map[string]any)["answers"].(map[string]any)["score"].(map[string]any)["legendImageReferences"].([]any)[0].(map[string]any)["imageIndex"] = 99
			case "missing-assertions":
				doc["assertions"] = []any{}
			}
			changed, _ := json.Marshal(doc)
			if err = os.WriteFile(path, changed, 0600); err != nil {
				t.Fatal(err)
			}
			var inventory struct {
				Schema int               `json:"schema"`
				Files  map[string]string `json:"files"`
			}
			if err = json.Unmarshal(originalInventory, &inventory); err != nil {
				t.Fatal(err)
			}
			inventory.Files[name] = sha256Text(changed)
			// Also update the manifest checksum when a case artifact changed.
			manifestPath := filepath.Join(root, "manifest.json")
			originalManifest, _ := os.ReadFile(manifestPath)
			defer os.WriteFile(manifestPath, originalManifest, 0600)
			if strings.Contains(name, "/") {
				var manifest map[string]any
				_ = json.Unmarshal(originalManifest, &manifest)
				parts := strings.SplitN(name, "/", 2)
				for _, v := range manifest["cases"].([]any) {
					c := v.(map[string]any)
					if c["dir"] == parts[0] {
						c["artifacts"].(map[string]any)[parts[1]] = sha256Text(changed)
					}
				}
				data, _ := json.Marshal(manifest)
				_ = os.WriteFile(manifestPath, data, 0600)
				inventory.Files["manifest.json"] = sha256Text(data)
			}
			data, _ := json.Marshal(inventory)
			if err = os.WriteFile(inventoryPath, data, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("python3", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py", root)
			output, err := cmd.CombinedOutput()
			cs.Record("verification", map[string]any{"fault": fault, "output": string(output), "error": errText(err)})
			cs.Check("semantic corruption rejected despite refreshed checksums", err != nil, "%s", output)
		})
	}
}
