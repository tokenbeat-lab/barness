package e2e

import (
	"context"
	"encoding/base64"
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
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

type imageQualificationFixture struct {
	root, binary, attempt, bundle string
	ctx                           context.Context
}

func newImageQualificationFixture(t *testing.T) imageQualificationFixture {
	t.Helper()
	privateGrantStore(t)
	root, _ := filepath.Abs("../..")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	f := imageQualificationFixture{root: root, ctx: ctx}
	f.binary = filepath.Join(t.TempDir(), "releasegate")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", f.binary, "./ai/release/cmd/releasegate")
	cmd.Dir, cmd.Env = root, regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("consumer build %v %s", err, raw)
	}
	return f
}

// Synthetic authority is confined to a private common-Git store. These test
// packets exercise the original live contract with retained real response
// bytes; they are never vendor measurements or publishable live receipts.
func (f imageQualificationFixture) packet(t *testing.T) imageQualificationFixture {
	t.Helper()
	var err error
	f.attempt, err = os.MkdirTemp(run.Dir(), "xai-images-proof-fixture-")
	if err != nil {
		t.Fatal(err)
	}
	session := fmt.Sprintf("SYNTHETIC-%d", time.Now().UnixNano())
	f.bundle = filepath.Join(f.attempt, "bundles", session)
	if err = os.MkdirAll(f.bundle, 0700); err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC()
	cat := xai.ImagesCandidateCatalog()
	catHash, _ := cat.Hash()
	var source map[string]string
	// Retain the actual 545-entry execution map from the pre-consumer
	// 724e9881 measurement, independent of current source and consumer rules.
	sourceRaw, err := os.ReadFile("testdata/xai-images/measured-source.json")
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, sourceRaw, &source)
	binary := filepath.Join(f.attempt, "synthetic-producer")
	if err = os.WriteFile(binary, []byte("#!/bin/sh\nprintf 'PASS\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	binaryHash, _ := livegrant.FileHash(binary)
	grant := livegrant.Grant{Schema: 1, ID: session, AccountAlias: "synthetic", RegionAlias: "global", ExpiresAt: started.Add(time.Hour), Scope: livegrant.Scope{Route: "xai-images", Mode: "live", Stage: "candidate", Provider: "xai", API: "xai-images", Endpoint: xaiimages.Endpoint, SDK: xaiimages.SDK, CatalogVersion: cat.Version, CatalogHash: catHash, SourceSHA256: "7a77c0cfb9f7abe0a838ebedd5d9d72074758abfff41e1bcdf9e89a35f014113", BinarySHA256: binaryHash, Models: []string{xaiimages.Model}, Scenes: xaiimages.Expected(), MaxCalls: 6, MaxAttempts: 6, CallSeconds: 180, ProcessSeconds: 1200, MaxUSD: .60}}
	grantRaw := mustMarshal(t, grant)
	claim, err := livegrant.Claim(grant, grantRaw, session, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(f.bundle, "execution-source.json"), sourceRaw)
	mustWriteXAINative(t, filepath.Join(f.bundle, "process-grant.json"), grantRaw)
	writeXAINativeJSON(t, filepath.Join(f.bundle, "process-authorization-receipt.json"), claim)
	writeXAINativeJSON(t, filepath.Join(f.bundle, "model-catalog.json"), cat)
	var captures imagineLiveContract
	raw, err := os.ReadFile("testdata/xai-images/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &captures)
	cases := []any{}
	scenes := []any{}
	ev := run.Case(t, "XAIIMG-qualification-fixture-"+session)
	rec := newRecorder()
	w := newImagineWorld(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Policy.MaxOutputBytes = 32 << 20
		c.Policy.Image.MaxOutputImageBytes = 8 << 20
		c.Policy.Image.MaxTotalOutputImageBytes = 16 << 20
		c.Observer = rec
		c.Policy.MaxQueuedObservations = 64
	})
	b := imagineBinding()
	b.TenantID = "live-smoke"
	b.BindingID = "xai-images"
	b.Version = "live"
	b.AccountScopeID = "live-xai-images"
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: b.TenantID, CredentialID: b.CredentialRef, Version: "live", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(xaiKey.secret)})
	for _, tc := range captures.Cases {
		var req struct {
			Prompt              string
			N                   int
			AspectRatio         string `json:"aspect_ratio"`
			Resolution, Quality string
			Image               *struct{ URL string }
			Images              []struct{ URL string }
		}
		mustUnmarshal(t, []byte(tc.RequestBody), &req)
		input := ai.ImagesRequest{Prompt: req.Prompt}
		if req.Image != nil {
			input.ReferenceImages = append(input.ReferenceImages, imagineLiveReference(t, req.Image.URL))
		}
		for _, ref := range req.Images {
			input.ReferenceImages = append(input.ReferenceImages, imagineLiveReference(t, ref.URL))
		}
		opts := ai.XAIImagesOptions{N: ai.Value(req.N), Resolution: ai.Value(req.Resolution), Quality: ai.Value(req.Quality)}
		if req.AspectRatio != "" {
			opts.AspectRatio = ai.Value(req.AspectRatio)
		}
		reply := imagineReply(tc.ResponseBody)
		reply.Header = tc.ResponseHeaders
		enqueue(ev, w, reply)
		requestID := "qualification-" + tc.ID
		result, err := w.client.GenerateImages(ctxFor(t), ai.CallScope{TenantID: "live-smoke", RequestID: requestID, JobID: "live-smoke"}, ai.Target{BindingID: "xai-images", ModelID: imagineModel}, input, opts)
		if err != nil {
			t.Fatal(err)
		}
		// Preserve the captured request ID through the public result and Observer
		// boundary; the server stays controlled and sends no vendor traffic.
		observations := rec.awaitCall(ev, requestID)

		dir := "XAIIMG-live-" + tc.ID
		path := filepath.Join(f.bundle, dir)
		if err = os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		writeXAINativeJSON(t, filepath.Join(path, "result.json"), result)
		writeXAINativeJSON(t, filepath.Join(path, "error.json"), "")
		writeXAINativeJSON(t, filepath.Join(path, "observations.json"), observations)
		writeXAINativeJSON(t, filepath.Join(path, "http.json"), []any{map[string]any{"method": "POST", "url": xaiimages.Endpoint + map[bool]string{true: "/images/edits", false: "/images/generations"}[len(input.ReferenceImages) > 0], "requestHeaders": map[string]string{"authorization": "[REDACTED]", "content-type": "application/json"}, "requestBody": tc.RequestBody, "status": 200, "responseHeaders": tc.ResponseHeaders, "responseHeaderNames": []string{"content-type", "x-request-id"}, "responseBody": tc.ResponseBody, "elapsedMs": 1}})
		fixtureHashes := map[string]string{}
		for i, block := range result.Content {
			image := block.(ai.ImageOutputImage)
			bytes, _ := base64.StdEncoding.Strict().DecodeString(image.Data)
			name := fmt.Sprintf("output-%d.image", i)
			mustWriteXAINative(t, filepath.Join(path, name), bytes)
			fixtureHashes[name] = evidence.Digest(bytes)
		}
		writeXAINativeJSON(t, filepath.Join(path, "assertions.json"), map[string]any{"case_id": dir, "status": "PASS", "test": "TestXAIImagesCandidate/" + tc.ID, "assertions": []any{map[string]any{"name": "synthetic protocol contract", "pass": true}}})
		cases = append(cases, imageQualificationCase(t, path, dir, fixtureHashes))
		scenes = append(scenes, map[string]string{"id": tc.ID, "outcome": "PASS", "caseId": dir})
	}
	gate := "XAIIMG-live-gate"
	if err = os.Mkdir(filepath.Join(f.bundle, gate), 0700); err != nil {
		t.Fatal(err)
	}
	writeXAINativeJSON(t, filepath.Join(f.bundle, gate, "assertions.json"), map[string]any{"case_id": gate, "status": "PASS", "test": "TestXAIImagesCandidate", "assertions": []any{map[string]any{"name": "synthetic gate", "pass": true}}})
	cases = append(cases, imageQualificationCase(t, filepath.Join(f.bundle, gate), gate, map[string]string{}))
	finished := time.Now().UTC()
	commit := "724e988132ac5bf6ac393db156e07081eb0e2427"
	report := map[string]any{"schema": 1, "kind": "xai-images-candidate-only", "mode": "live", "stage": "candidate", "outcome": "PASS", "complete": true, "formalQualified": false, "auditPassed": true, "startedAt": started, "finishedAt": finished, "gitCommit": commit, "scope": grant.Scope, "buildSourceSha256": grant.Scope.SourceSHA256, "plan": xaiimages.Plan(), "scenarios": scenes, "callsUsed": 6, "httpAttempts": 6, "vendorHTTPAttempts": 6, "inputImages": 8, "outputImages": 7, "reservedUSD": .53, "reportedUSD": .51}
	writeXAINativeJSON(t, filepath.Join(f.bundle, "xai-images-candidate-report.json"), report)
	writeXAINativeJSON(t, filepath.Join(f.bundle, "audit.json"), map[string]any{"findings": []any{}})
	manifest := map[string]any{"module": "barness-ai", "package": "-tags live -timeout 20m0s ./ai/live", "git_commit": commit, "started_at": started, "finished_at": time.Now().UTC(), "versions": map[string]string{"process_build_source": grant.Scope.SourceSHA256, "xai_images_source": grant.Scope.SourceSHA256, "live_combo": "xai-images", "live_sdk": xaiimages.SDK, "model_catalog_version": cat.Version, "model_catalog_hash": catHash}, "cases": cases}
	writeXAINativeJSON(t, filepath.Join(f.bundle, "manifest.json"), manifest)
	binaryRelative, _ := filepath.Rel(f.root, binary)
	argv := []string{filepath.ToSlash(binaryRelative), "-test.run=^TestXAIImagesCandidate$", "-test.count=1", "-test.timeout=1200s"}
	process := exec.CommandContext(f.ctx, binary, argv[1:]...)
	process.Env = regionalChildEnvironment()
	out, err := process.Output()
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(f.attempt, "stdout.jsonl"), out)
	mustWriteXAINative(t, filepath.Join(f.attempt, "stderr.log"), nil)
	mustWriteXAINative(t, filepath.Join(f.attempt, "exit-code"), []byte("0\n"))
	command := map[string]any{"schema": 1, "argv": argv, "stage": "live", "route": "xai-images", "timeoutSeconds": 1260, "environmentNames": []string{"BARNESS_AI_LIVE", "BARNESS_AI_LIVE_KEY_XAI_IMAGES", "BARNESS_AI_LIVE_PERMISSIONS_FILE", "GOENV", "GOFLAGS"}, "toolchain": map[string]any{"executableSHA256": binaryHash, "goEnv": "off", "goFlags": ""}}
	writeXAINativeJSON(t, filepath.Join(f.attempt, "command.json"), command)
	files := map[string]string{}
	for _, n := range []string{"command.json", "stdout.jsonl", "stderr.log", "exit-code"} {
		b, _ := os.ReadFile(filepath.Join(f.attempt, n))
		files[n] = evidence.Digest(b)
	}
	selected := map[string]string{filepath.ToSlash(binaryRelative): binaryHash}
	auth := filepath.Join(f.attempt, "authority")
	if err = os.Mkdir(auth, 0700); err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string][]byte{"grant.json": grantRaw, "user-approval.md": []byte("Synthetic private-authority parser fixture; no real execution permission.\n"), "approved-retry-plan.md": []byte("Synthetic finite six-scene image capture plan.\n"), "proxy-selection.json": []byte(`{"proxy":"","explicitUserSelection":false,"target":"https://api.x.ai/v1","ambientProxyUsed":false}`)} {
		path := filepath.Join(auth, name)
		mustWriteXAINative(t, path, raw)
		rel, _ := filepath.Rel(f.root, path)
		selected[filepath.ToSlash(rel)] = evidence.Digest(raw)
	}
	receipt := map[string]any{"schema": 1, "feature": "barness-ai-xai-imagine", "stage": "live", "route": "xai-images", "status": "PASS", "exitCode": 0, "argv": argv, "startedAt": started, "finishedAt": time.Now().UTC(), "source": map[string]any{"commit": commit, "files": source}, "selectedInputs": selected, "files": files, "commandSHA256": files["command.json"], "toolchain": map[string]any{"executableSHA256": binaryHash, "goEnv": "off", "goFlags": ""}, "bundles": []any{map[string]string{"path": "bundles/" + session, "manifestSHA256": imageQualificationHash(t, filepath.Join(f.bundle, "manifest.json")), "auditSHA256": imageQualificationHash(t, filepath.Join(f.bundle, "audit.json")), "package": "-tags live -timeout 20m0s ./ai/live"}}}
	writeXAINativeJSON(t, filepath.Join(f.attempt, "receipt.json"), receipt)
	return f
}
func imageQualificationHash(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return evidence.Digest(raw)
}
func imageQualificationCase(t *testing.T, path, id string, fixtures map[string]string) map[string]any {
	t.Helper()
	files, _ := os.ReadDir(path)
	hashes := map[string]string{}
	for _, f := range files {
		hashes[f.Name()] = imageQualificationHash(t, filepath.Join(path, f.Name()))
	}
	test := "TestXAIImagesCandidate"
	if id != "XAIIMG-live-gate" {
		test += "/" + strings.TrimPrefix(id, "XAIIMG-live-")
	}
	return map[string]any{"id": id, "dir": id, "status": "PASS", "artifacts": hashes, "fixtures": fixtures, "test": test}
}
func (f imageQualificationFixture) consume(t *testing.T, name string) (map[string]any, int) {
	t.Helper()
	ev := run.Case(t, "XAIIMG-qualification-"+name)
	out := filepath.Join(f.attempt, "consumer-"+name)
	relative, _ := filepath.Rel(f.root, f.bundle)
	cmd := exec.CommandContext(f.ctx, f.binary, "-live-bundle", filepath.ToSlash(relative), "-out", out)
	cmd.Dir, cmd.Env = f.root, regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	code := imagineProcessExit(err)
	if code != 0 {
		diagnostic := exec.CommandContext(f.ctx, f.binary, "-inspect-bundle", f.bundle)
		diagnostic.Dir, diagnostic.Env = f.root, regionalChildEnvironment()
		text, diagErr := diagnostic.CombinedOutput()
		ev.Record("native-diagnostic", map[string]any{"argv": diagnostic.Args, "exitCode": imagineProcessExit(diagErr), "output": string(text)})
	}
	ev.Record("process", map[string]any{"argv": cmd.Args, "exitCode": code, "output": string(raw), "vendorCalls": 0})
	var result map[string]any
	b, _ := os.ReadFile(filepath.Join(out, "live-consumer.json"))
	json.Unmarshal(b, &result)
	return result, code
}
func imageQualificationMutate(t *testing.T, path string, mutate func(map[string]any)) {
	t.Helper()
	v := readXAINativeObject(t, path)
	mutate(v)
	writeXAINativeJSON(t, path, v)
}
func imageQualificationRefresh(t *testing.T, f imageQualificationFixture) {
	t.Helper()
	m := readXAINativeObject(t, filepath.Join(f.bundle, "manifest.json"))
	for _, raw := range m["cases"].([]any) {
		c := raw.(map[string]any)
		dir := filepath.Join(f.bundle, c["dir"].(string))
		entries, _ := os.ReadDir(dir)
		art := map[string]string{}
		for _, entry := range entries {
			art[entry.Name()] = imageQualificationHash(t, filepath.Join(dir, entry.Name()))
		}
		c["artifacts"] = art
	}
	writeXAINativeJSON(t, filepath.Join(f.bundle, "manifest.json"), m)
	imageQualificationMutate(t, filepath.Join(f.attempt, "receipt.json"), func(r map[string]any) {
		r["bundles"].([]any)[0].(map[string]any)["manifestSHA256"] = imageQualificationHash(t, filepath.Join(f.bundle, "manifest.json"))
	})
}
