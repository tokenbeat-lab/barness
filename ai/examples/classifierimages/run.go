package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type assertion struct {
	Name string `json:"name"`
	Pass bool   `json:"pass"`
}
type failure struct {
	Code  ai.Code  `json:"code"`
	Phase ai.Phase `json:"phase"`
}
type report struct {
	Route      string                `json:"route"`
	Scenario   string                `json:"scenario"`
	Entry      string                `json:"entry"`
	Status     string                `json:"status"`
	Result     ai.ClassifierResult   `json:"result"`
	Error      *failure              `json:"error"`
	View       map[string]answerView `json:"view"`
	Assertions []assertion           `json:"assertions"`
}

func (r *report) check(name string, pass bool) {
	r.Assertions = append(r.Assertions, assertion{name, pass})
	if !pass {
		r.Status = "FAIL"
	}
}

func runScenario(ctx context.Context, b *bundle, s scenario, files []fixture, live *hostConfig, key string, liveTransport http.RoundTripper) (bool, error) {
	var server *provider.Server
	var scriptCapture *responseCapture
	var transport http.RoundTripper
	var cfg hostConfig
	var liveCapture *captureTransport
	if live == nil {
		server = provider.New(map[string]string{syntheticKey: "key:synthetic-example"})
		defer server.Close()
		reply, capture := reply(s)
		scriptCapture = capture
		server.Enqueue(reply)
		cfg = controlledConfig(s.Route, server.URL())
		local := provider.LoopbackTransport()
		defer local.CloseIdleConnections()
		transport = local
	} else {
		cfg = *live
		liveCapture = &captureTransport{inner: liveTransport, requestLimit: cfg.Policy.MaxRequestBytes, responseLimit: cfg.Policy.MaxOutputBytes}
		transport = liveCapture
	}
	if s.Name == "budget" {
		cfg.Policy.Classifier.Images.References.MaxReferences = 2
	}
	store := &resolver{binding: cfg.Binding, key: key}
	obs := &observer{finished: make(chan struct{})}
	gauges := probe.New()
	bodies := provider.TrackBodies(transport)
	client, err := newClient(cfg, store, obs, bodies, live == nil, gauges)
	if err != nil {
		return false, fmt.Errorf("Host Client assembly failed: %w", err)
	}
	// Revocation is an external Host state update, observed on the next call.
	if s.Name == "image-revoked" {
		store.revoke(true)
	}
	if s.Name == "binding-revoked" {
		store.revoke(false)
	}
	req := input(s.Name, files)
	var options ai.Options = ai.OpenAIDecisionsOptions{}
	if s.Route == "jev" {
		options = ai.TypeSafeOptions{ExperimentalImages: s.Name != "experiment-off"}
	}
	scope := ai.CallScope{TenantID: cfg.Binding.TenantID, RequestID: requestid.New()}
	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if s.Name == "canceled" {
		cancel()
	}
	hooks := ai.Hooks{}
	search := files[0].Identity
	if s.Name == "hook" || s.Name == "box" && s.Entry == "hooked" {
		hooks = swapImages(files)
		// The Host explicitly updates search-image attribution with its own Hook.
		if s.Name == "box" {
			search = files[1].Identity
		}
	}
	var res ai.ClassifierResult
	if s.Entry == "hooked" {
		res, err = client.WithHooks(hooks).Classify(callCtx, scope, cfg.Target, req, options)
	} else {
		res, err = client.Classify(callCtx, scope, cfg.Target, req, options)
	}
	r := report{Route: s.Route, Scenario: s.Name, Entry: s.Entry, Status: "PASS", Result: res}
	var ae *ai.Error
	if errors.As(err, &ae) {
		r.Error = &failure{ae.Code, ae.Phase}
	}
	waitCtx, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()
	observations, waitErr := obs.await(waitCtx)
	if waitErr != nil {
		return false, waitErr
	}
	requests := []provider.Request{}
	var response []byte
	meta := responseMetadata{}
	if server != nil {
		requests = server.Requests()
		var fixtureErr error
		response, fixtureErr = scriptCapture.snapshot()
		if len(requests) > 0 {
			meta = responseMetadata{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "synthetic-example-response"}}
		}
		if fixtureErr != nil {
			return false, fixtureErr
		}
	} else {
		requests, response, meta = liveCapture.snapshot()
	}
	r.check("one terminal observation", len(observations) >= 2 && observations[len(observations)-1].Kind == ai.ObservationCallFinished)
	r.check("observer matches final metadata", reflect.DeepEqual(observations[len(observations)-1].Call, res.Metadata))
	if err == nil {
		var viewErr error
		r.View, viewErr = interpret(res, files, search)
		r.check("Host resolves final identities and geometry", viewErr == nil)
	} else {
		r.check("typed atomic error", r.Error != nil && res.Answers == nil && res.ImageInputs == nil)
	}
	if live != nil {
		r.check("live Classify returned validated output", err == nil)
		if r.Status != "FAIL" {
			r.Status = "OBSERVED"
		}
	} else {
		checkControlled(&r, s, err, files, requests, store.reads())
	}
	resources := map[string]int64{"bodies": bodies.Open(), "calls": gauges.ActiveCalls(), "permits": gauges.Permits(), "waiters": gauges.AdmissionWaiters()}
	for name, value := range resources {
		r.check("released "+name, value == 0)
	}
	prefix := "cases/" + s.id() + "/"
	effective, _ := store.ResolveBinding(ctx, scope, cfg.Binding.BindingID)
	config := map[string]any{"host": cfg, "effectiveBinding": effective, "target": cfg.Target, "options": options, "scope": scope, "credentialSource": "explicit Host injection", "credentialReads": store.reads(), "transport": b.mode, "retry": "binding budget; controlled example zero retries", "replay": "go run ./ai/examples/classifierimages -route " + s.Route + " -scenario " + s.Name + " -entry " + s.Entry}
	if live != nil {
		config["replay"] = "repeat with the same explicit -live -config -key-env -route -scenario -entry; never reuse an output directory"
	}
	artifacts := []struct {
		name  string
		value any
	}{{"input.json", req}, {"config.json", config}, {"requests.json", requests}, {"observations.json", observations}, {"observer-stats.json", client.ObserverStats()}, {"report.json", r}, {"resources.json", resources}, {"response-meta.json", meta}}
	for _, a := range artifacts {
		if err := b.record(prefix+a.name, a.value); err != nil {
			return false, err
		}
	}
	if len(response) == 0 {
		response = []byte("null")
	}
	// Retain even a malformed vendor response without manufacturing valid JSON.
	name := "response.json"
	if !json.Valid(response) {
		name = "response.bin"
	}
	if err := b.file(prefix+name, response); err != nil {
		return false, err
	}
	b.cases = append(b.cases, s.id())
	return r.Status != "FAIL", nil
}

func checkControlled(r *report, s scenario, err error, files []fixture, requests []provider.Request, reads int) {
	success := s.Name == "context" || s.Name == "reuse" || s.Name == "refusal" || s.Route == "jev" && (s.Name == "state" || s.Name == "instructions" || s.Name == "criteria" || s.Name == "box" || s.Name == "hook")
	sent := success || s.Name == "bad-box" || s.Name == "bad-legend"
	r.check("expected success or rejection", (err == nil) == success)
	want := 0
	if sent {
		want = 1
	}
	r.check("HTTP and Attempt counts", len(requests) == want && len(r.Result.Metadata.Attempts) == want)
	if !sent {
		r.check("rejected before credential lookup", reads == 0)
	}
	if !success {
		var code ai.Code
		var phase ai.Phase
		switch s.Name {
		case "image-revoked":
			code, phase = ai.CodeTenantDenied, ai.PhaseCapability
		case "binding-revoked":
			code, phase = ai.CodeTenantDenied, ai.PhaseBinding
		case "budget":
			code, phase = ai.CodeResourceLimit, ai.PhaseScope
		case "canceled":
			code, phase = ai.CodeCanceled, ai.PhaseScope
		case "bad-box", "bad-legend":
			code, phase = ai.CodeProtocol, ai.PhaseResponse
		default:
			code, phase = ai.CodeInvalidRequest, ai.PhaseCapability
		}
		r.check("expected error code and phase", r.Error != nil && r.Error.Code == code && r.Error.Phase == phase)
		if sent {
			r.check("usage retained after invalid response without retry", r.Result.Usage.Input == 100 && len(r.Result.Metadata.Attempts) == 1)
		}
		return
	}
	expected := []ai.ClassifierImageInput{files[0].Identity, files[1].Identity}
	switch s.Name {
	case "reuse":
		expected = []ai.ClassifierImageInput{files[1].Identity, files[0].Identity}
		if s.Route == "decisions" {
			expected = append(expected, files[1].Identity)
		}
	case "criteria", "bad-legend":
		expected = []ai.ClassifierImageInput{files[1].Identity, files[0].Identity}
	case "box":
		if s.Entry == "hooked" {
			expected = []ai.ClassifierImageInput{files[1].Identity, files[0].Identity}
		}
	}
	r.check("final MIME and file hashes", reflect.DeepEqual(r.Result.ImageInputs, expected))
	if s.Name == "criteria" || s.Name == "hook" {
		r.check("two resolved visual rubric slots", len(r.View["score"].LegendImages) == 2)
	}
	if s.Name == "box" {
		a := r.View["location"]
		r.check("orientation and crop map to original edges", a.OriginalPixels != nil && *a.OriginalPixels == (pixels{200, 200, 440, 360}))
		p := r.View["exists"].Probability
		r.check("zero existence does not suppress Box", p != nil && *p == 0 && a.Kind == "box")
		confidence, ok := a.Confidence.Get()
		r.check("null differs from numeric zero", s.Entry == "plain" && a.Confidence.IsNull() || s.Entry == "hooked" && ok && confidence == 0)
	}
	if s.Name == "refusal" {
		r.check("explicit refusal view", r.View["exists"].Kind == "refusal")
	}
}
