//go:build live

package live

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// These are filesystem attacks on the public verify/merge commands. Refresh
// the outer checksum so semantic rejections cannot rely on an accidental seal.
func formalTamperCases(t *testing.T, cs *evidence.Case, bundle string, env []string) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "jevformal")
	cmd := exec.Command("go", "build", "-o", bin, "./ai/live/cmd/jevformal")
	cmd.Dir = "../.."
	cmd.Env = env
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	for _, kind := range []string{"model", "case-alias", "invalid-box", "missing-state-scope", "route", "hash", "expected", "scenario", "budget", "capture", "manifest", "audit", "outer-inventory", "outer-manifest", "duplicate-report", "trailing-report", "symlink", "unlisted", "source", "protocol-hash", "effect-hash", "unknown-field", "old-identity"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "UC13-tamper-"+kind)
			dest := filepath.Join(run.Dir(), "formal-corrupt-"+kind)
			if kind == "trailing-report" {
				dest = filepath.Join(t.TempDir(), "corrupt")
			}
			if err := os.Mkdir(dest, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.CopyFS(dest, os.DirFS(bundle)); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dest, "live-report.json")
			raw, _ := os.ReadFile(path)
			var report map[string]any
			if json.Unmarshal(raw, &report) != nil {
				t.Fatal("invalid fixture")
			}
			switch kind {
			case "missing-state-scope":
				corruptFormalStateScope(t, dest)
				cmd := exec.Command(bin, "-protocol", filepath.Join(dest, "protocol"), "-effects", filepath.Join(dest, "effects"), "-out", filepath.Join(run.Dir(), "scope-relabelled-package"))
				cmd.Dir = "../.."
				cmd.Env = env
				_, err := cmd.CombinedOutput()
				ev.Record("producer", map[string]any{"rejected": err != nil})
				ev.Check("producer requires actual State image scope", err != nil, "relabeled local scenario accepted without State image")
				return
			case "invalid-box":
				corruptFormalBox(t, dest)
				seal, _ := os.ReadFile(filepath.Join(dest, "protocol/multimodal-integrity.json"))
				sum := sha256.Sum256(seal)
				report["jev"].(map[string]any)["protocolHash"] = hex.EncodeToString(sum[:])
			case "model":
				report["model"] = "jev-1.13.0"
			case "route":
				report["endpoint"] = "https://unapproved.invalid/v1"
			case "hash":
				report["catalogHash"] = "candidate"
			case "expected":
				report["expected"] = []string{"box-existence-plain"}
			case "scenario":
				report["scenarios"] = report["scenarios"].([]any)[1:]
			case "budget":
				report["budget"].(map[string]any)["callsUsed"] = 13
			case "capture":
				os.Remove(filepath.Join(dest, "protocol/MM11-box-existence-plain/exchanges.json"))
			case "manifest":
				os.WriteFile(filepath.Join(dest, "protocol/manifest.json"), []byte(`{}`), 0600)
			case "audit":
				os.WriteFile(filepath.Join(dest, "audit.json"), []byte(`{"findings":[{"rule":"synthetic"}]}`), 0600)
			case "outer-manifest":
				os.WriteFile(filepath.Join(dest, "manifest.json"), []byte(`{}`), 0600)
			case "unlisted":
				os.WriteFile(filepath.Join(dest, "extra.txt"), []byte("unlisted"), 0600)
			case "source":
				report["gitCommit"] = "borrowed-source"
			case "protocol-hash":
				report["jev"].(map[string]any)["protocolHash"] = "wrong"
			case "effect-hash":
				report["jev"].(map[string]any)["effectsHash"] = "wrong"
			case "unknown-field":
				report["unexpected"] = "payload"
			case "old-identity":
				report["combo"] = "typesafe-classifier"
			}
			raw, _ = json.Marshal(report)
			if kind == "duplicate-report" {
				raw = append([]byte(`{"model":"jev-1.14.0",`), raw[1:]...)
			}
			if kind == "case-alias" {
				raw = append([]byte(`{"MODEL":"jev-1.13.0",`), raw[1:]...)
			}
			if kind == "trailing-report" {
				raw = append(raw, []byte(`{}`)...)
				ev.Fixture("attack-report.txt", raw)
			}
			os.WriteFile(path, raw, 0600)
			files := map[string]string{}
			err := filepath.WalkDir(dest, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(dest, path)
				if rel == "formal-integrity.json" {
					return nil
				}
				raw, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				sum := sha256.Sum256(raw)
				files[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			raw, _ = json.Marshal(files)
			if kind == "outer-inventory" {
				raw = []byte(`{}`)
			}
			os.WriteFile(filepath.Join(dest, "formal-integrity.json"), raw, 0600)
			if kind == "symlink" {
				target := filepath.Join(dest, "seal.json")
				os.Rename(filepath.Join(dest, "formal-integrity.json"), target)
				os.Symlink(target, filepath.Join(dest, "formal-integrity.json"))
			}
			if kind == "invalid-box" {
				check := exec.Command("python3", "../internal/testkit/jevformal/verify_protocol.py", "--formal-catalog", filepath.Join(dest, "protocol/MM11-configuration/formal-catalog.json"), filepath.Join(dest, "protocol"))
				check.Env = env
				_, directErr := check.CombinedOutput()
				ev.Check("independent structural verifier rejects invalid Box", directErr != nil, "invalid normalized Box accepted")
			}
			cmd := exec.Command(bin, "-verify", dest)
			cmd.Dir = "../.."
			cmd.Env = env
			_, err = cmd.CombinedOutput()
			ev.Record("verification", map[string]any{"rejected": err != nil})
			ev.Check("resealed corruption rejected at public command", err != nil, "accepted %s", kind)
		})
	}
}

func corruptFormalBox(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "protocol/MM11-box-existence-plain")
	resultPath := filepath.Join(dir, "result.json")
	raw, _ := os.ReadFile(resultPath)
	var result map[string]any
	json.Unmarshal(raw, &result)
	result["result"].(map[string]any)["answers"].(map[string]any)["location"].(map[string]any)["box"].(map[string]any)["left"] = 2
	raw, _ = json.Marshal(result)
	os.WriteFile(resultPath, raw, 0600)
	exchangePath := filepath.Join(dir, "exchanges.json")
	raw, _ = os.ReadFile(exchangePath)
	var exchanges []map[string]any
	json.Unmarshal(raw, &exchanges)
	var response map[string]any
	json.Unmarshal([]byte(exchanges[0]["responseBody"].(string)), &response)
	response["answers"].(map[string]any)["location"].(map[string]any)["box"].(map[string]any)["left"] = 2
	raw, _ = json.Marshal(response)
	exchanges[0]["responseBody"] = string(raw)
	raw, _ = json.Marshal(exchanges)
	os.WriteFile(exchangePath, raw, 0600)
	refreshFormalProtocol(t, root)
}

// Move the State image to a second Instructions slot. The payload remains
// valid and all counters/captures agree, but the mandatory State scope is gone.
func corruptFormalStateScope(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "protocol/MM11-local-state-instructions-plain")
	path := filepath.Join(dir, "public-request.json")
	raw, _ := os.ReadFile(path)
	var request map[string]any
	json.Unmarshal(raw, &request)
	request["imageReferences"].([]any)[0].(map[string]any)["pointer"] = "/questions/example/instructions/a"
	request["questions"].(map[string]any)["example"].(map[string]any)["instructions"].(map[string]any)["a"] = nil
	raw, _ = json.Marshal(request)
	os.WriteFile(path, raw, 0600)
	path = filepath.Join(dir, "exchanges.json")
	raw, _ = os.ReadFile(path)
	var exchanges []map[string]any
	json.Unmarshal(raw, &exchanges)
	var body map[string]any
	json.Unmarshal([]byte(exchanges[0]["requestBody"].(string)), &body)
	oldState, _ := json.Marshal(body["state"])
	state := body["state"].(map[string]any)
	body["questions"].(map[string]any)["example"].(map[string]any)["instructions"].(map[string]any)["a"] = state["photo"]
	state["photo"] = nil
	newState, _ := json.Marshal(state)
	raw, _ = json.Marshal(body)
	exchanges[0]["requestBody"] = string(raw)
	raw, _ = json.Marshal(exchanges)
	os.WriteFile(path, raw, 0600)
	path = filepath.Join(root, "protocol/multimodal-report.json")
	raw, _ = os.ReadFile(path)
	var report map[string]any
	json.Unmarshal(raw, &report)
	actual := report["budget"].(map[string]any)["actual"].(map[string]any)
	actual["stateBytes"] = actual["stateBytes"].(float64) + float64(len(newState)-len(oldState))
	raw, _ = json.Marshal(report)
	os.WriteFile(path, raw, 0600)
	refreshFormalProtocol(t, root)
}

func refreshFormalProtocol(t *testing.T, root string) {
	t.Helper()
	manifestPath := filepath.Join(root, "protocol/manifest.json")
	raw, _ := os.ReadFile(manifestPath)
	var manifest map[string]any
	json.Unmarshal(raw, &manifest)
	for _, c := range manifest["cases"].([]any) {
		caseMap := c.(map[string]any)
		for name := range caseMap["artifacts"].(map[string]any) {
			raw, _ := os.ReadFile(filepath.Join(root, "protocol", caseMap["id"].(string), name))
			sum := sha256.Sum256(raw)
			caseMap["artifacts"].(map[string]any)[name] = hex.EncodeToString(sum[:])
		}
	}
	raw, _ = json.Marshal(manifest)
	os.WriteFile(manifestPath, raw, 0600)
	sealPath := filepath.Join(root, "protocol/multimodal-integrity.json")
	raw, _ = os.ReadFile(sealPath)
	var seal map[string]any
	json.Unmarshal(raw, &seal)
	for name := range seal["files"].(map[string]any) {
		raw, _ := os.ReadFile(filepath.Join(root, "protocol", name))
		sum := sha256.Sum256(raw)
		seal["files"].(map[string]any)[name] = hex.EncodeToString(sum[:])
	}
	raw, _ = json.Marshal(seal)
	os.WriteFile(sealPath, raw, 0600)
}
