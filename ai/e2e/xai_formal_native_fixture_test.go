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

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type xaiFormalNativeFixture struct {
	root, binary, matrixBinary, source string
	ctx                                context.Context
}

func newXAIFormalNativeFixture(t *testing.T) xaiFormalNativeFixture {
	t.Helper()
	privateGrantStore(t)
	probe := exec.CommandContext(ctxFor(t), "git")
	probe.Dir, probe.Env = "../..", xaiProcessEnv(t)
	if !privateGrantProcess(t, run.Case(t, "PD-isolation-"+t.Name()), probe) {
		t.FailNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	f := xaiFormalNativeFixture{root: root, ctx: ctx}
	live := filepath.Join(t.TempDir(), "xai-live.test")
	f.binary = filepath.Join(t.TempDir(), "releasegate")
	f.matrixBinary = filepath.Join(t.TempDir(), "supportmatrix")
	for _, args := range [][]string{{"test", "-c", "-tags=live", "-o", live, "./ai/live"}, {"build", "-o", f.binary, "./ai/release/cmd/releasegate"}, {"build", "-o", f.matrixBinary, "./ai/live/cmd/supportmatrix"}} {
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated native fixture build: %v %s", err, out)
		}
	}
	out, err := os.MkdirTemp(run.Dir(), "xai-native-controlled-")
	if err != nil {
		t.Fatal(err)
	}
	runChild := func(extra ...string) string {
		cmd := exec.CommandContext(ctx, live, "-test.run=^TestXAIFormal$", "-test.timeout=30s")
		cmd.Dir = root
		cmd.Env = append(xaiProcessEnv(t), "BARNESS_AI_LIVE_COMBO=xai-responses", evidence.DirEnv+"="+out)
		cmd.Env = append(cmd.Env, extra...)
		if raw, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("controlled source fixture failed: %v %s", err, raw)
		}
		paths, _ := filepath.Glob(filepath.Join(out, "*", "xai-formal-report.json"))
		if len(paths) == 0 {
			t.Fatal("controlled source report missing")
		}
		return paths[len(paths)-1]
	}
	metadata := readXAINativeObject(t, runChild())
	grant := map[string]any{}
	for _, k := range []string{"combo", "provider", "api", "endpoint", "model", "sourceSha256", "binarySha256", "catalogVersion", "catalogHash", "sdk", "expected"} {
		grant[k] = metadata[k]
	}
	grant["schema"], grant["id"], grant["mode"], grant["stage"] = 1, fmt.Sprintf("SYNTHETIC_control_%d", time.Now().UnixNano()), "controlled", "formal"
	grant["accountAlias"], grant["regionAlias"], grant["expiresAt"] = "synthetic", "global", time.Now().UTC().Add(time.Hour)
	grant["maxCalls"], grant["maxAttempts"], grant["maxRequestedOutputTokens"], grant["maxReservedInputTokens"], grant["maxUSD"] = 24, 24, 24576, 196608, 1
	gp := filepath.Join(t.TempDir(), "grant.json")
	writeXAINativeJSON(t, gp, grant)
	path := runChild("BARNESS_XAI_CONTROLLED=fixture", "BARNESS_AI_LIVE_ACCOUNT_ALIAS=synthetic", "BARNESS_AI_LIVE_REGION_ALIAS=global", "BARNESS_AI_LIVE_PERMISSIONS_FILE="+gp)
	f.source = filepath.Dir(path)
	return f
}

func (f xaiFormalNativeFixture) copy(t *testing.T, name string) string {
	t.Helper()
	parent, err := os.MkdirTemp(run.Dir(), "xai-synthetic-native-"+name+"-")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(parent, "bundle")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(f.source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(f.source, path)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.Mkdir(target, 0700)
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, raw, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
	cat := ai.BuiltinCatalog()
	hash, err := cat.Hash()
	if err != nil {
		t.Fatal(err)
	}
	// Rewrite only explicit fixture metadata; this is not a paid-report renewal.
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		raw, _ := os.ReadFile(path)
		var v any
		if json.Unmarshal(raw, &v) != nil {
			return nil
		}
		var walk func(any)
		walk = func(v any) {
			switch n := v.(type) {
			case map[string]any:
				for k, x := range n {
					if k == "catalogVersion" {
						n[k] = cat.Version
					} else if k == "catalogHash" {
						n[k] = hash
					} else {
						walk(x)
					}
				}
			case []any:
				for _, x := range n {
					walk(x)
				}
			}
		}
		walk(v)
		writeXAINativeJSON(t, path, v)
		return nil
	})
	detail := readXAINativeObject(t, filepath.Join(dir, "xai-formal-report.json"))
	detail["kind"], detail["stage"], detail["mode"], detail["formalQualified"] = "xai-formal", "formal", "live", true
	detail["buildSourceSha256"] = detail["sourceSha256"]
	detail["budget"].(map[string]any)["vendorHTTPAttempts"] = 24
	grant := readXAINativeObject(t, filepath.Join(dir, "xai-authorization-grant.json"))
	grant["stage"], grant["mode"] = "formal", "live"
	claim := readXAINativeObject(t, filepath.Join(dir, "xai-authorization-claim.json"))
	bindXAINativeGrant(t, dir, detail, grant, claim)
	var rows []supportmatrix.ScenarioResult
	raw, _ := json.Marshal(detail["scenarios"])
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	input, output := int64(0), int64(0)
	for i := range rows {
		rows[i].ProviderRequestIDs = nil
		for _, label := range xaiNativeFixtureLabels(rows[i].ID) {
			result := readXAINativeObject(t, filepath.Join(dir, rows[i].CaseID, label+"-result.json"))
			md := result["metadata"].(map[string]any)
			a := md["attempts"].([]any)[0].(map[string]any)
			rid, _ := a["providerRequestId"].(string)
			rows[i].ProviderRequestIDs = append(rows[i].ProviderRequestIDs, rid)
			u := result["message"].(map[string]any)["usage"].(map[string]any)
			input += int64(u["input"].(float64) + u["cacheRead"].(float64))
			output += int64(u["output"].(float64))
		}
	}
	detail["scenarios"] = rows
	writeXAINativeJSON(t, filepath.Join(dir, "xai-formal-report.json"), detail)
	for _, name := range []string{"xai-candidate-report.json", "xai-candidate-report.unfinalized.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			mustRemoveXAINative(t, filepath.Join(dir, name))
		}
	}
	source, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := os.ReadFile(filepath.Join(f.root, "ai/internal/testkit/xai/candidate.json"))
	if err != nil {
		t.Fatal(err)
	}
	source["ai/internal/testkit/xai/candidate.json"] = evidence.Digest(candidate)
	writeXAINativeJSON(t, filepath.Join(dir, "xai-execution-source.json"), source)
	var base struct {
		GitCommit             string
		SDK                   string
		AccountAlias          string
		StartedAt, FinishedAt time.Time
		Expected              []string
	}
	raw, _ = json.Marshal(detail)
	json.Unmarshal(raw, &base)
	projection := supportmatrix.Report{Schema: supportmatrix.SchemaVersion, Combo: "xai-responses", Operation: "chat", Provider: "xai", API: "openai-responses", Endpoint: "https://api.x.ai/v1", Model: "grok-4.7", SDK: base.SDK, AccountAlias: base.AccountAlias, GitCommit: base.GitCommit, CatalogVersion: cat.Version, CatalogHash: hash, StartedAt: base.StartedAt, FinishedAt: base.FinishedAt, Expected: base.Expected, Scenarios: rows, Budget: supportmatrix.Budget{MaxCalls: 24, MaxAttempts: 24, MaxRetries: 0, MaxOutputTokens: 1024, AttemptsReserved: 24, MaxRequestedOutputTokens: 24576, RequestedOutputTokens: 24576, MaxReservedInputTokens: 196608, ReservedInputTokens: 196608, CallsUsed: 24, HTTPAttempts: 24, InputTokensUsed: int(input), OutputTokensUsed: int(output), ProcessSeconds: 1800, CallSeconds: 65, FinalizationReserveSeconds: 30}}
	writeXAINativeJSON(t, filepath.Join(dir, "live-report.json"), projection)
	manifest := readXAINativeObject(t, filepath.Join(dir, "manifest.json"))
	manifest["package"] = "-tags live -timeout 30m0s ./ai/live"
	versions := manifest["versions"].(map[string]any)
	versions["model_catalog_version"], versions["model_catalog_hash"] = cat.Version, hash
	writeXAINativeJSON(t, filepath.Join(dir, "manifest.json"), manifest)
	resealXAINativeManifest(t, dir)
	return dir
}
func xaiNativeFixtureLabels(id string) []string {
	switch id {
	case "encrypted-continuation":
		return []string{"first", "continued"}
	case "tools-auto", "tools-required", "tools-parallel", "tool-result-image":
		return []string{"calls", "results"}
	case "tools-none":
		return []string{"calls"}
	case "user-png", "user-jpeg":
		return []string{"image"}
	case "cache-usage":
		return []string{"first", "repeat"}
	default:
		return []string{"text"}
	}
}
func (f xaiFormalNativeFixture) consume(t *testing.T, ev *evidence.Case, dir string) (int, []byte) {
	t.Helper()
	parent, makeErr := os.MkdirTemp(run.Dir(), "xai-native-consumer-")
	if makeErr != nil {
		t.Fatal(makeErr)
	}
	out := filepath.Join(parent, "consumer")
	relative, relativeErr := filepath.Rel(f.root, dir)
	if relativeErr != nil {
		t.Fatal(relativeErr)
	}
	cmd := exec.CommandContext(f.ctx, f.binary, "-live-bundle", filepath.ToSlash(relative), "-out", out)
	cmd.Dir, cmd.Env = f.root, xaiProcessEnv(t)
	raw, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		code = -1
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		}
	}
	ev.Record("synthetic-consumer-process", map[string]any{"kind": "synthetic-native-contract-fixture", "notActualLive": true, "argv": cmd.Args, "exitCode": code, "output": string(raw)})
	report, e := os.ReadFile(filepath.Join(out, "live-consumer.json"))
	if e != nil {
		t.Fatal(e)
	}
	ev.Fixture("synthetic-consumer.json", report)
	return code, report
}
func readXAINativeObject(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func mustWriteXAINative(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func writeXAINativeJSON(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, path, append(raw, '\n'))
}
func editXAINativeJSON(t *testing.T, dir, name string, edit func(map[string]any)) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	v := readXAINativeObject(t, path)
	edit(v)
	writeXAINativeJSON(t, path, v)
}
func editXAINativeArray(t *testing.T, dir, name string, edit func([]any)) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v []any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	edit(v)
	writeXAINativeJSON(t, path, v)
}
func editXAINativeHTTPBody(t *testing.T, dir, name string, edit func(map[string]any)) {
	t.Helper()
	editXAINativeArray(t, dir, name, func(v []any) {
		x := v[0].(map[string]any)
		var body map[string]any
		if err := json.Unmarshal([]byte(x["requestBody"].(string)), &body); err != nil {
			t.Fatal(err)
		}
		edit(body)
		raw, _ := json.Marshal(body)
		x["requestBody"] = string(raw)
	})
}
func resealXAINativeManifest(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, "manifest.json")
	m := readXAINativeObject(t, p)
	for _, value := range m["cases"].([]any) {
		c := value.(map[string]any)
		files := map[string]string{}
		entries, err := os.ReadDir(filepath.Join(dir, c["dir"].(string)))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, c["dir"].(string), e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			files[e.Name()] = evidence.Digest(raw)
		}
		c["artifacts"] = files
	}
	writeXAINativeJSON(t, p, m)
}

func addXAINativeFixtureHeaders(t *testing.T, dir string) {
	for _, which := range []string{"xai-formal-report.json", "live-report.json"} {
		editXAINativeJSON(t, dir, which, func(report map[string]any) {
			for _, raw := range report["scenarios"].([]any) {
				scene := raw.(map[string]any)
				var ids []string
				for _, label := range xaiNativeFixtureLabels(scene["id"].(string)) {
					id := "SYNTHETIC_header_" + scene["id"].(string) + "_" + label
					ids = append(ids, id)
					prefix := scene["caseId"].(string) + "/" + label
					editXAINativeArray(t, dir, prefix+"-http.json", func(v []any) { e := v[0].(map[string]any); e["responseHeaders"].(map[string]any)["x-request-id"] = id })
					editXAINativeJSON(t, dir, prefix+"-result.json", func(v map[string]any) {
						v["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["providerRequestId"] = id
					})
					editXAINativeArray(t, dir, prefix+"-observer.json", func(v []any) {
						v[2].(map[string]any)["attempt"].(map[string]any)["providerRequestId"] = id
						v[3].(map[string]any)["call"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["providerRequestId"] = id
					})
				}
				scene["providerRequestIds"] = ids
			}
		})
	}
}
