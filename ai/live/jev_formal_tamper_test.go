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
	for _, kind := range []string{"model", "route", "hash", "expected", "scenario", "budget", "capture", "manifest", "audit", "outer-inventory", "outer-manifest", "duplicate-report", "trailing-report", "symlink", "unlisted", "source", "protocol-hash", "effect-hash", "unknown-field", "old-identity"} {
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
			cmd := exec.Command(bin, "-verify", dest)
			cmd.Dir = "../.."
			cmd.Env = env
			_, err = cmd.CombinedOutput()
			ev.Record("verification", map[string]any{"rejected": err != nil})
			ev.Check("resealed corruption rejected at public command", err != nil, "accepted %s", kind)
		})
	}
}
