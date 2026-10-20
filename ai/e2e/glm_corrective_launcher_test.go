package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The corrective experiment has a separate finite budget and committed-script
// freeze, unlike the older minimal guard fixture. Exercise its real CLI before
// any credential source or paid child can be reached.
func TestGLMCorrectiveLauncherGuards(t *testing.T) {
	for _, mode := range []string{"review-denied", "script-drift", "binary-drift", "already-started", "gate-report-absent"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "P0-GLM13-corrective-"+mode)
			root := t.TempDir()
			scripts := filepath.Join(root, ".scratch/barness-ai-glm/evidence/ticket-13")
			os.MkdirAll(scripts, 0700)
			for _, name := range []string{"commands.py", "run.py", "glm_launcher.py", "rollback.py", "rollback-probe.txt", "corrective.py"} {
				raw, err := os.ReadFile(filepath.Join("../../.scratch/barness-ai-glm/evidence/ticket-13", name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(scripts, name), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			os.MkdirAll(filepath.Join(root, "ai/release"), 0700)
			for name, text := range map[string]string{
				"ai/fixture.go": "package fixture\n", "go.mod": "module fixture\n", "go.sum": "",
				"ai/release/traceability.json": `{"liveCombos":[]}`, "ai/release/catalog-snapshot.json": `{"hash":"fixture-catalog"}`,
			} {
				os.WriteFile(filepath.Join(root, name), []byte(text), 0600)
			}
			bin := filepath.Join(root, "bin")
			os.Mkdir(bin, 0700)
			fakeGo := "#!/usr/bin/env python3\nimport sys,pathlib\na=sys.argv[1:]\nif '-o' in a:\n p=pathlib.Path(a[a.index('-o')+1]);p.write_bytes(b'frozen fixture binary');p.chmod(0o700)\nif any('/releasegate' in x for x in a):sys.exit(17)\n"
			os.WriteFile(filepath.Join(bin, "go"), []byte(fakeGo), 0700)
			command := func(name string, args ...string) ([]byte, error) {
				cmd := exec.CommandContext(ctxFor(t), name, args...)
				cmd.Dir = root
				for _, key := range []string{"HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
					if v, ok := os.LookupEnv(key); ok {
						cmd.Env = append(cmd.Env, key+"="+v)
					}
				}
				cmd.Env = append(cmd.Env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "PYTHONDONTWRITEBYTECODE=1")
				return cmd.CombinedOutput()
			}
			for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=fixture", "-c", "user.email=fixture@invalid", "commit", "-qm", "fixture"}} {
				if raw, err := command("git", args...); err != nil {
					t.Fatalf("git fixture: %v %s", err, raw)
				}
			}
			script := filepath.Join(scripts, "corrective.py")
			raw, err := command("python3", script, "plan", ".evidence/round")
			var plan struct {
				MaxCalls        int            `json:"maxCalls"`
				MaxAttempts     int            `json:"maxAttempts"`
				ProviderCaps    map[string]int `json:"providerCaps"`
				AutomaticReruns *bool          `json:"automaticReruns"`
				Components      []struct {
					Name     string `json:"name"`
					MaxCalls int    `json:"maxCalls"`
				} `json:"components"`
			}
			decodeErr := json.Unmarshal(raw, &plan)
			names, calls := []string{}, 0
			for _, c := range plan.Components {
				names = append(names, c.Name)
				calls += c.MaxCalls
			}
			ev.Check("only three independently reserved components and exact provider caps", err == nil && decodeErr == nil && plan.MaxCalls == 36 && plan.MaxAttempts == 36 && calls == 36 && plan.AutomaticReruns != nil && !*plan.AutomaticReruns && len(plan.ProviderCaps) == 2 && plan.ProviderCaps["typesafe"] == 24 && plan.ProviderCaps["openai"] == 12 && slices.Equal(names, []string{"protocol114", "visual-decisions", "visual-jev"}), "plan=%+v error=%v %v", plan, err, decodeErr)
			ev.Record("plan", plan)
			for _, action := range []string{"prepare", "check"} {
				if raw, err := command("python3", script, action, ".evidence/round"); err != nil {
					t.Fatalf("%s: %v %s", action, err, raw)
				}
			}
			out := filepath.Join(root, ".evidence/round")
			commit, err := command("git", "rev-parse", "HEAD")
			if err != nil {
				t.Fatal(err)
			}
			decision := "APPROVED"
			if mode == "review-denied" {
				decision = "DENIED"
			}
			proof, _ := json.Marshal(map[string]string{"decision": decision, "sourceCommit": strings.TrimSpace(string(commit))})
			os.WriteFile(filepath.Join(out, "prerelease-source-review.json"), proof, 0600)
			marker := filepath.Join(out, "paid-corrective-round-started.json")
			want := "prerelease source review"
			switch mode {
			case "script-drift":
				f, _ := os.OpenFile(script, os.O_APPEND|os.O_WRONLY, 0600)
				f.WriteString("\n# source mutation\n")
				f.Close()
				want = "source/config changed"
			case "binary-drift":
				os.WriteFile(filepath.Join(out, "live.test"), []byte("changed binary"), 0600)
				want = "binary changed"
			case "already-started":
				os.WriteFile(marker, []byte("original closed intent"), 0600)
				want = "FileExistsError"
			}
			action := "live"
			if mode == "gate-report-absent" {
				action = "gate"
				original := filepath.Join(root, ".evidence/glm13-legacy-current-r3")
				os.MkdirAll(original, 0700)
				os.WriteFile(filepath.Join(original, "glm-support-inputs.json"), []byte(`{}`), 0600)
				for _, route := range []string{"text-jev", "text-luna"} {
					os.MkdirAll(filepath.Join(original, route, "one"), 0700)
					os.WriteFile(filepath.Join(original, route, "one/manifest.json"), []byte(`{}`), 0600)
				}
			}
			raw, err = command("python3", script, action, ".evidence/round")
			_, child := os.Stat(filepath.Join(out, "protocol114-command.json"))
			_, intent := os.Stat(marker)
			ev.Record("process", map[string]any{"failed": err != nil, "output": string(raw), "credentialsInherited": false})
			if mode == "gate-report-absent" {
				var summary struct {
					ExitCode int  `json:"exitCode"`
					Pass     bool `json:"passVerdict"`
				}
				data, readErr := os.ReadFile(filepath.Join(out, "gate-summary.json"))
				decodeErr := json.Unmarshal(data, &summary)
				code := 0
				if e, ok := err.(*exec.ExitError); ok {
					code = e.ExitCode()
				}
				ev.Check("reused text inputs reach child and retain failure without a report", readErr == nil && decodeErr == nil && code == 17 && summary.ExitCode == 17 && !summary.Pass && !strings.Contains(string(raw), "Traceback"), "read=%v decode=%v exit=%d output=%s", readErr, decodeErr, code, raw)
			} else {
				ev.Check("source, approval, or exclusive round stops before credential or paid child", err != nil && strings.Contains(string(raw), want) && os.IsNotExist(child) && (mode == "already-started" || os.IsNotExist(intent)), "mode=%s error=%v output=%s", mode, err, raw)
			}
		})
	}
}
