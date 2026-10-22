package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func regionalChildEnvironment(extra ...string) []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+v)
		}
	}
	return append(env, extra...)
}

// The real isolated process drives all nine new model identities through the
// public Client/SDK into a controlled server. Captures stay in their original
// native directory; controlled PASS never authorizes formal service support.
func TestRegionalProviderReleaseCapturePipeline(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "regional-live.test")
	source, err := supportmatrix.RegionalExecutionSource()
	if err != nil {
		t.Fatal(err)
	}
	build := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource="+supportmatrix.RegionalSourceHash(source), "-o", binary, "./ai/live")
	build.Dir = "../.."
	build.Env = regionalChildEnvironment()
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("regional live binary build: %v %s", err, regionalSafeOutput(raw))
	}
	var promotionInputs []string
	for _, combo := range []string{"zai-chat", "zai-responses", "zai-messages", "minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages"} {
		t.Run(combo, func(t *testing.T) {
			ev := run.Case(t, "RP18-capture-pipeline-"+combo)
			if combo == "minimax-cn-responses" {
				regionalMissingNativeSource(t, binary, combo)
			}
			route, _ := supportmatrix.RegionalRouteOf(combo)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:regional-controlled"})
			defer srv.Close()
			for _, model := range supportmatrix.RegionalModels(combo) {
				for _, kind := range []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "tool-call", "tool-result", "native-history", "cancel", "recovery"} {
					var reply provider.Reply
					if combo == "zai-chat" && (kind == "full-stream" || kind == "tool-call") {
						reply, _, _, _ = zaiChatSource(t, model, kind)
					} else if combo == "zai-responses" && (kind == "full-stream" || kind == "tool-call") {
						reply = regionalZAIResponsesSource(t, model, kind)
					} else if combo == "minimax-cn-responses" && (kind == "full-stream" || kind == "tool-call") {
						reply = regionalCNResponsesSource(t, kind)
					} else if kind == "tool-call" {
						reply = regionalPipelineToolReply(t, ai.ProviderID(route.Provider), ai.API(route.API), model)
					} else if kind == "cancel" {
						reply = regionalPipelinePrefix(t, ai.API(route.API))
						reply.End = provider.EndHold
					} else {
						reply = regionalTextReply(t, ai.ProviderID(route.Provider), ai.API(route.API), model)
					}
					if kind == "tool-result" || kind == "native-history" {
						for i, c := range reply.Chunks {
							reply.Chunks[i] = []byte(strings.NewReplacer("Regional text.", "REGIONAL42", "Hello MiniMax.", "REGIONAL42", "Synthetic reply.", "REGIONAL42").Replace(string(c)))
						}
					}
					if kind == "full-complete" || kind == "simple-complete" {
						mutation := "raw-usage-zero"
						if kind == "simple-complete" {
							mutation = "raw-usage-partial"
						}
						var body strings.Builder
						for _, chunk := range reply.Chunks {
							body.Write(chunk)
						}
						mutated := regionalMutateRaw(t, body.String(), mutation)
						reply.Chunks, reply.Script = [][]byte{[]byte(mutated)}, mutated
					}
					reply.Header["X-Request-Id"], reply.Header["Request-Id"] = "regional-controlled", "regional-controlled"
					srv.Enqueue(reply)
				}
			}
			grant, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
			if err != nil {
				t.Fatal(err)
			}
			grant.Mode = "controlled"
			grant.ID, grant.AccountAlias, grant.RegionAlias = "synthetic-"+combo+"-"+evidence.Digest([]byte(run.Dir()))[:12], "synthetic-account", "synthetic-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			grantPath := filepath.Join(t.TempDir(), "grant.json")
			if err := os.WriteFile(grantPath, mustMarshal(t, grant), 0600); err != nil {
				t.Fatal(err)
			}
			// Native originals are produced directly beside the parent bundle,
			// so disposing this worktree never destroys or refreshes the capture.
			base := filepath.Join(filepath.Dir(run.Dir()), "regional-producer-native", filepath.Base(run.Dir()), combo)
			cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+grantPath, "BARNESS_AI_EVIDENCE_DIR="+base)
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"failed": err != nil, "output": regionalSafeOutput(raw), "argv": []string{"regional-live.test", "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s"}, "realCredentialsInherited": false, "vendorCalls": 0})
			ev.Record("actual-controlled-requests", srv.Requests())
			if !ev.Check("complete controlled qualification process", err == nil, "%v %s", err, regionalSafeOutput(raw)) {
				return
			}
			if combo == "zai-messages" {
				for i, request := range srv.Requests() {
					fields := bodyFields(t, request.Body)
					full52 := i >= 9 && i%9 != 2 && i%9 != 3
					if full52 {
						ev.Check("GLM-5.2 full uses independently observed exact shape", jsonEqual(fields["thinking"], []byte(`{"type":"enabled","budget_tokens":1024}`)) && string(fields["max_tokens"]) == "4096", "scenario index=%d", i)
					} else {
						ev.Check("GLM-5.3 and canonical simple keep thinking omitted", fields["thinking"] == nil, "scenario index=%d", i)
					}
				}
			}
			nativePaths, err := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
			if !ev.Check("one original native bundle", err == nil && len(nativePaths) == 1, "count=%d", len(nativePaths)) {
				return
			}
			dir := filepath.Dir(nativePaths[0])
			promotionInputs = append(promotionInputs, dir)
			var report supportmatrix.Report
			if raw, err := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile)); err != nil || json.Unmarshal(raw, &report) != nil {
				t.Fatal("native regional report unavailable")
			}
			if combo == "zai-messages" {
				for _, scene := range report.Scenarios {
					var opts ai.AnthropicOptions
					raw, err := os.ReadFile(filepath.Join(dir, scene.CaseID, "options.json"))
					if err != nil || json.Unmarshal(raw, &opts) != nil {
						t.Fatal("native request options unavailable")
					}
					_, kind, _ := strings.Cut(scene.ID, "/")
					full52 := scene.Model == "glm-5.2" && kind != "simple-stream" && kind != "simple-complete"
					enabled, _ := opts.ThinkingEnabled.Get()
					ev.Check("saved replay options preserve actual request shape", full52 && enabled && opts.ThinkingBudgetTokens == 1024 || !full52 && opts.ThinkingEnabled.IsZero() && opts.ThinkingBudgetTokens == 0, "scenario=%s", scene.ID)
				}
			}
			ev.Record("native-report", report)
			if combo == "zai-chat" || combo == "zai-responses" || combo == "minimax-cn-responses" {
				var source struct {
					Regional struct{ NativeSourceCases map[string]string }
				}
				mustUnmarshal(t, mustMarshal(t, report), &source)
				for _, model := range supportmatrix.RegionalModels(combo) {
					ev.Check("native source predeclared independently of tool output", source.Regional.NativeSourceCases[model] == supportmatrix.RegionalCaseID(combo, model+"/full-stream"), "model=%s", model)
				}
			}
			replay := exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/regionalverify", "-controlled", dir)
			replay.Dir, replay.Env = "../..", regionalChildEnvironment()
			replayRaw, replayErr := replay.CombinedOutput()
			ev.Check("independent consumer verifies controlled native captures", replayErr == nil, "%s", regionalSafeOutput(replayRaw))
			inputReserved := 0
			for _, request := range srv.Requests() {
				inputReserved += max(2048, len(request.Body)+1024)
			}
			ev.Check("reservation covers actual encoded wire", report.Budget.ReservedInputTokens == inputReserved, "reserved=%d encoded=%d", report.Budget.ReservedInputTokens, inputReserved)
			regionalCaptureRefusals(t, dir, combo)
			regionalCaptureCancellationPrefixes(t, dir, combo)

			manifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			rel, _ := filepath.Rel(run.Dir(), filepath.Join(dir, "manifest.json"))
			ev.Record("original-native-pointer", map[string]string{"manifest": rel, "sha256": evidence.Digest(manifest)})
			ev.Check("all own models and capabilities captured", report.Regional != nil && report.Regional.Complete && report.Regional.Mode == "controlled" && len(report.Scenarios) == len(supportmatrix.RegionalExpected(combo)) && len(srv.Requests()) == len(report.Scenarios), "%+v", report.Budget)
			_, err = supportmatrix.VerifyRegionalQualification(dir)
			ev.Check("controlled capture never becomes own live proof", err != nil, "controlled qualification accepted")
			_, err = supportmatrix.NewRegionalReport(dir)
			ev.Check("candidate never becomes formal support", err != nil, "candidate certified")
			if combo == "zai-responses" {
				regionalResponsesNativeOriginals(t, dir)
			}
			if combo == "minimax-cn-responses" {
				regionalCNResponsesNativeOriginals(t, dir)
				regionalNativeSourceRefusals(t, dir, combo)
			}
			if combo == "zai-chat" || combo == "zai-responses" {
				regionalMissingNativeSource(t, binary, combo)
				regionalNativeSourceRefusals(t, dir, combo)
			}
			// The pre-send durable reservation is consumed even in a successful
			// controlled run. Reusing it must fail before any additional HTTP.
			again := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			again.Dir, again.Env = cmd.Dir, cmd.Env
			raw, err = again.CombinedOutput()
			ev.Check("grant cannot be reused", err != nil && len(srv.Requests()) == len(report.Scenarios), "%s", regionalSafeOutput(raw))
		})
	}
	if len(promotionInputs) == 6 {
		regionalControlledPromotion(t, promotionInputs)
	}
}

func regionalPipelineToolReply(t *testing.T, p ai.ProviderID, api ai.API, model string) provider.Reply {
	if api != ai.APIAnthropicMessages {
		r := regionalToolReply(t, p, api, model)
		for i, c := range r.Chunks {
			r.Chunks[i] = []byte(strings.ReplaceAll(string(c), `"name":"inspect"`, `"name":"inspect_regional"`))
		}
		return r
	}
	frames := []json.RawMessage{
		json.RawMessage(`{"type":"message_start","message":{"id":"regional-tool","model":"` + model + `","usage":{"input_tokens":8,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`),
		json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":"regional-signature"}}`),
		json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Regional reasoning."}}`),
		json.RawMessage(`{"type":"content_block_stop","index":0}`),
		json.RawMessage(`{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"regional-call","name":"inspect_regional","input":{}}}`),
		json.RawMessage(`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"label\":"}}`),
		json.RawMessage(`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"\"alpha\"}"}}`),
		json.RawMessage(`{"type":"content_block_stop","index":1}`),
		json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":9}}`),
		json.RawMessage(`{"type":"message_stop"}`),
	}
	return sseEvents(t, frames, provider.FramingLF)
}

func regionalSafeOutput(raw []byte) string {
	text := string(raw)
	if home := os.Getenv("HOME"); home != "" {
		text = strings.ReplaceAll(text, home, "[HOME]")
	}
	return text
}

func regionalPipelinePrefix(t *testing.T, api ai.API) provider.Reply {
	reply := regionalPrefix(t, api)
	if api == ai.APIAnthropicMessages {
		for i, c := range reply.Chunks {
			reply.Chunks[i] = []byte(strings.ReplaceAll(string(c), `"text":"known-content"`, `"text":""`))
		}
		reply.Chunks = append(reply.Chunks, []byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"known-content\"}}\n\n"))
	}
	return reply
}
