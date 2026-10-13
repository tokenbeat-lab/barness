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

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Public executable seam: each child owns its configuration, budget and bundle.
// Intentional child failures are verified here, never converted to live PASS.
func TestDecisionsCandidateProcess(t *testing.T) {
	if os.Getenv("BARNESS_DECISIONS_CONTROLLED") != "" {
		return
	}
	cases := []string{"not-run", "switch-off", "foreign-key", "missing-alias", "missing-permissions", "endpoint-override", "remote-proxy", "source-missing", "source-changed", "catalog-changed", "success", "refusal", "usage-absent", "usage-partial", "usage-zero", "model-alias", "model-missing", "429", "504", "malformed-2xx", "precision-2xx", "narrowed", "calls-budget", "questions-budget", "state-budget", "images-budget", "image-bytes-budget", "permissions-trailing", "permissions-duplicate", "permissions-broad", "permissions-not-dedicated", "missing-region", "no-key-foreign", "mixed-refusal", "price-threshold", "price-long-tier", "permissions-case-alias", "process-deadline", "integrity-read-failure", "integrity-write-failure", "unknown-account", "unknown-account-invalid-proof", "unknown-account-foreign-key", "unknown-account-remote-proxy", "unknown-account-no-key"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "P10-OD51-process-"+name)
			base := t.TempDir()
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			permissions := filepath.Join(base, "permissions.json")
			proof := `{"accountAlias":"controlled-test","regionAlias":"offline","dedicatedTestKey":true,"model":"gpt-6-luna","permissions":["POST /v1/decisions"],"spendLimitUSD":1,"verifiedAt":"2026-10-12T00:00:00Z","verificationReference":"controlled host permission fixture"}`
			switch name {
			case "unknown-account-invalid-proof":
				proof = strings.Replace(proof, `"dedicatedTestKey":true`, `"dedicatedTestKey":false,"DedicatedTestKey":true`, 1)
			case "permissions-case-alias":
				proof = strings.Replace(proof, `"dedicatedTestKey":true`, `"dedicatedTestKey":false,"DedicatedTestKey":true`, 1)
			case "permissions-trailing":
				proof += `{}`
			case "permissions-duplicate":
				proof = strings.Replace(proof, `"dedicatedTestKey":true`, `"dedicatedTestKey":false,"dedicatedTestKey":true`, 1)
			case "permissions-broad":
				proof = strings.Replace(proof, `"POST /v1/decisions"`, `"*"`, 1)
			case "permissions-not-dedicated":
				proof = strings.Replace(proof, `"dedicatedTestKey":true`, `"dedicatedTestKey":false`, 1)
			}
			if err := os.WriteFile(permissions, []byte(proof), 0600); err != nil {
				t.Fatal(err)
			}
			env := []string{"BARNESS_DECISIONS_CONTROLLED=" + name, envLive + "=1", envCombo + "=openai-decisions-candidate", envAlias + "=controlled-test", "BARNESS_AI_LIVE_REGION_ALIAS=offline", keyVar("openai-decisions-candidate") + "=key:controlled-decisions", "BARNESS_AI_LIVE_PERMISSIONS_FILE=" + permissions, "BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890", "BARNESS_AI_EVIDENCE_DIR=" + base}
			for i, kv := range env {
				switch {
				case (name == "not-run" || name == "no-key-foreign" || name == "unknown-account-no-key") && strings.HasPrefix(kv, keyVar("openai-decisions-candidate")+"="):
					env[i] = keyVar("openai-decisions-candidate") + "="
				case name == "switch-off" && strings.HasPrefix(kv, envLive+"="):
					env[i] = envLive + "=0"
				case name == "missing-region" && strings.HasPrefix(kv, "BARNESS_AI_LIVE_REGION_ALIAS="):
					env[i] = "BARNESS_AI_LIVE_REGION_ALIAS="
				case name == "missing-alias" && strings.HasPrefix(kv, envAlias+"="):
					env[i] = envAlias + "="
				case name == "missing-permissions" && strings.HasPrefix(kv, "BARNESS_AI_LIVE_PERMISSIONS_FILE="):
					env[i] = "BARNESS_AI_LIVE_PERMISSIONS_FILE="
				case name == "remote-proxy" && strings.HasPrefix(kv, "BARNESS_AI_LIVE_PROXY="):
					env[i] = "BARNESS_AI_LIVE_PROXY=http://proxy.invalid:7890"
				}
			}
			if strings.HasPrefix(name, "unknown-account") {
				env = append(env, "BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1")
				if name != "unknown-account-invalid-proof" {
					for i, kv := range env {
						for _, key := range []string{envAlias, "BARNESS_AI_LIVE_REGION_ALIAS", "BARNESS_AI_LIVE_PERMISSIONS_FILE"} {
							if strings.HasPrefix(kv, key+"=") {
								env[i] = key + "="
							}
						}
					}
				}
				if name == "unknown-account-remote-proxy" {
					for i, kv := range env {
						if strings.HasPrefix(kv, "BARNESS_AI_LIVE_PROXY=") {
							env[i] = "BARNESS_AI_LIVE_PROXY=http://proxy.invalid:7890"
						}
					}
				}
			}
			if name == "foreign-key" || name == "no-key-foreign" || name == "unknown-account-foreign-key" {
				env = append(env, keyVar("openai-chat")+"=key:foreign")
			}
			if name == "endpoint-override" {
				env = append(env, "BARNESS_AI_LIVE_ENDPOINT=https://proxy.invalid")
			}
			if strings.HasPrefix(name, "source-") || name == "catalog-changed" {
				dir = filepath.Join(base, "inputs")
				if err := os.MkdirAll(filepath.Join(dir, "testdata", "decisions"), 0755); err != nil {
					t.Fatal(err)
				}
				files, err := os.ReadDir("testdata/decisions")
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					raw, err := os.ReadFile(filepath.Join("testdata/decisions", f.Name()))
					if err != nil {
						t.Fatal(err)
					}
					if err = os.WriteFile(filepath.Join(dir, "testdata/decisions", f.Name()), raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
				file := "guide.md"
				if name == "catalog-changed" {
					file = "candidate.json"
				}
				path := filepath.Join(dir, "testdata/decisions", file)
				if name == "source-missing" {
					err = os.Remove(path)
				} else {
					err = os.WriteFile(path, []byte(`{"changed":true}`), 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDecisionsCandidateControlledChild$", "-test.v")
			if name == "process-deadline" {
				cmd.Args = append(cmd.Args, "-test.timeout=12s")
			}
			cmd.Dir = dir
			// Do not inherit credentials, endpoint or proxy settings into a child.
			cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TMPDIR=" + os.TempDir()}, env...)
			output, childErr := cmd.CombinedOutput()
			paths, _ := filepath.Glob(filepath.Join(base, "*", "candidate-decisions-report.json"))
			if len(paths) != 1 {
				t.Fatalf("candidate report missing: %v\n%s", childErr, output)
			}
			raw, err := os.ReadFile(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			var report candidateDecisionsReport
			if err = json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			cs.Record("child-report", report)
			// Preserve the complete child evidence, including intentional FAILs.
			root := filepath.Dir(paths[0])
			err = filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if e.IsDir() {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				cs.Fixture(strings.ReplaceAll(rel, string(os.PathSeparator), "__"), data)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := supportmatrix.LoadReport(paths[0])
			cs.Check("formal loader rejects candidate schema", loadErr != nil, "candidate report entered formal schema")
			integrity, err := os.ReadFile(filepath.Join(root, "candidate-integrity.json"))
			if err != nil && !strings.HasPrefix(name, "integrity-") {
				t.Fatal(err)
			}
			if err == nil {
				var inventory struct {
					Files map[string]string `json:"files"`
				}
				if err := json.Unmarshal(integrity, &inventory); err != nil {
					t.Fatal(err)
				}
				cs.Check("root artifacts inventoried", inventory.Files["manifest.json"] != "" && inventory.Files["audit.json"] != "" && inventory.Files["candidate-decisions-report.json"] != "", "missing root checksum")
				for file, hash := range inventory.Files {
					data, err := os.ReadFile(filepath.Join(root, file))
					cs.Check("inventory matches exact captured bytes", err == nil && sha256Text(data) == hash, "changed artifact %s", file)
				}
				cs.Fixture("candidate-integrity-observed.json", integrity)

			}
			_, err = os.Stat(filepath.Join(root, "live-report.json"))
			cs.Check("candidate cannot be merged as formal support", os.IsNotExist(err), "formal report emitted")
			var audit struct {
				Findings []any `json:"findings"`
			}
			a, err := os.ReadFile(filepath.Join(root, "audit.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(a, &audit); err != nil {
				t.Fatal(err)
			}
			cs.Check("child audit clean", len(audit.Findings) == 0, "audit findings")
			want := "FAIL"
			switch name {
			case "not-run", "switch-off", "narrowed", "unknown-account-no-key":
				want = "NOT_RUN"
			case "success", "refusal", "usage-absent", "usage-partial", "usage-zero", "model-alias", "mixed-refusal", "price-threshold", "price-long-tier", "unknown-account":
				want = "PASS"
			}
			cs.Check("honest process status", string(report.Outcome) == want && (childErr != nil) == (want == "FAIL"), "outcome %s exit %v: %s", report.Outcome, childErr, output)
			cs.Check("complete only own full run", report.Complete == (want == "PASS") && report.Mode == "controlled", "invalid complete/mode")
			if name == "unknown-account" {
				var fields map[string]any
				if err := json.Unmarshal(raw, &fields); err != nil {
					t.Fatal(err)
				}
				cs.Check("permission proof remains absent", fields["hostPermissionVerification"] == nil, "fabricated permission proof")
				cs.Check("metadata is explicitly unknown", fields["accountVerificationStatus"] == "unknown" && fields["regionVerificationStatus"] == "unknown" && fields["permissionVerificationStatus"] == "unknown", "missing unknown statuses")
				unknown := []string{"provider account identity", "account region", "credential privilege scope", "dedicated test credential status", "provider-side spend limit"}
				for _, item := range unknown {
					cs.Check("unverified claims remain unknown", strings.Contains(string(raw), item), "missing unknown %s", item)
				}
			}
			b := report.Budget
			cs.Check("bounded independent budget", b.CallsUsed <= 12 && b.AttemptsReserved <= 12 && b.HTTPAttempts <= b.AttemptsReserved && b.QuestionsUsed <= 24 && b.StateBytesSubmitted <= 4<<20 && b.ImagesUsed <= 8 && b.ImageBytesSubmitted <= 4<<20 && b.Retries == 0 && b.ProcessSeconds == 180 && b.CallSeconds == 30, "budget %+v", b)
			if want == "PASS" {
				resultRaw, err := os.ReadFile(filepath.Join(root, "P10-candidate-mixed-questions", "result.json"))
				if err != nil {
					t.Fatal(err)
				}
				var result struct {
					Result struct {
						ResponseModel string `json:"responseModel"`
						Metadata      struct {
							ModelID  string `json:"modelId"`
							Attempts []struct {
								UsageReporting ai.UsageReporting `json:"usageReporting"`
							} `json:"attempts"`
						} `json:"metadata"`
						Usage ai.Usage `json:"usage"`
					} `json:"result"`
				}
				if err = json.Unmarshal(resultRaw, &result); err != nil {
					t.Fatal(err)
				}
				reporting := ai.UsageComplete
				if name == "usage-absent" {
					reporting = ai.UsageUnreported
				}
				if name == "usage-partial" {
					reporting = ai.UsagePartial
				}
				cs.Check("usage completeness independent of pipeline completeness", len(result.Result.Metadata.Attempts) == 1 && result.Result.Metadata.Attempts[0].UsageReporting == reporting, "usage presence altered")
				if name == "usage-zero" || name == "usage-absent" {
					cs.Check("zero is not fabricated consumption", result.Result.Usage.Input == 0 && result.Result.Usage.Output == 0, "invented tokens")
				}
				if name == "price-threshold" {
					cs.Check("candidate threshold stays short tier", result.Result.Usage.Input == 272000 && costClose(result.Result.Usage.Cost.Input, 0.0272), "wrong threshold price")
				}
				if name == "price-long-tier" {
					cs.Check("candidate whole request long tier", result.Result.Usage.Input == 272001 && costClose(result.Result.Usage.Cost.Input, 0.0544002) && result.Result.Usage.CacheRead == 0 && result.Result.Usage.Cost.Output == 0, "wrong long tier price")
				}
				if name == "model-alias" {
					cs.Check("actual alias cannot replace authorized model", result.Result.ResponseModel == "controlled-actual-luna-alias" && result.Result.Metadata.ModelID == "gpt-6-luna", "identity conflation")
				}
				cs.Check("entire mandatory pipeline captured", len(report.Scenarios) == 9 && b.CallsUsed == 9 && b.HTTPAttempts == 9 && b.ImagesUsed == 3, "pipeline %+v", b)
			}
			if name == "process-deadline" {
				cs.Check("deadline preserves terminal evidence", b.CallsUsed == 1 && b.HTTPAttempts == 1 && !report.Complete, "deadline lost or exceeded budget")
			}
			if strings.HasPrefix(name, "integrity-") {
				cs.Check("artifact failure revokes acceptance", !report.Complete && report.Outcome == supportmatrix.Fail && b.CallsUsed == 9, "stale successful report after finalization failure")
			}
			if strings.Contains(name, "budget") {
				cs.Check("overspend refused", report.Scenarios[len(report.Scenarios)-1].ErrorCategory == "budget", "missing budget failure")
			}
			if name == "429" || name == "504" || strings.HasSuffix(name, "2xx") {
				cs.Check("failures and all reservations counted without retry", b.CallsUsed == 9 && b.AttemptsReserved == 9 && b.HTTPAttempts == 9, "failure consumption %+v", b)
			}
			if want == "FAIL" && (strings.HasPrefix(name, "source-") || name == "catalog-changed" || name == "foreign-key" || name == "missing-alias" || name == "missing-permissions" || name == "endpoint-override" || name == "remote-proxy" || strings.HasPrefix(name, "permissions-") || name == "missing-region" || name == "no-key-foreign" || strings.HasPrefix(name, "unknown-account-")) {
				cs.Check("config refuses before spend", b.CallsUsed == 0 && b.HTTPAttempts == 0, "config escaped")
			}
			_, included := ai.BuiltinCatalog().LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "gpt-6-luna")
			cs.Check("candidate absent from builtins", !included, "premature inclusion")
		})
	}
}

func TestDecisionsCandidateControlledChild(t *testing.T) {
	name := os.Getenv("BARNESS_DECISIONS_CONTROLLED")
	if name == "" {
		t.Skip("isolated controlled process only")
	}
	driveCandidateDecisions(t, controlledDecisionsRecorder(t, name), func(t *testing.T) []decisionsScenario { return controlledCandidateBudgetScenarios(t, name) }, "controlled")
	if name == "integrity-write-failure" {
		if err := os.Mkdir(filepath.Join(run.Dir(), "candidate-integrity.json"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if name == "integrity-read-failure" {
		// A nonregular file exercises the same inventory-read rejection as the
		// former 16 MiB+1 sparse file. Auditing that huge synthetic file exceeded
		// the 20 s child supervisor under race (live-authorized-20261013 receipt).
		// DevNull is empty for the privacy audit; descriptor validation still
		// rejects it before inventory reads. No production budget is changed.
		if err := os.Symlink(os.DevNull, filepath.Join(run.Dir(), "nonregular-controlled-fixture.bin")); err != nil {
			t.Fatal(err)
		}
	}
}
