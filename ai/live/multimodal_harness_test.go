//go:build live

package live

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var multimodalCatalogPins = map[string]string{"jev": "9ac5655199e22d45db2fd5b72af79ef597068a89f4298fbc47c539ed4f6426a4", "decisions": "d5cc54e4189c8075d031180fc6f9456e06306fa0ca7e8ae99d6dab776450712a"}

func TestMultimodalCandidate(t *testing.T) {
	if os.Getenv("BARNESS_MM_FORMAL") != "" {
		t.Skip("candidate entry only")
	}
	driveMultimodal(t)
}

func TestMultimodalFormal(t *testing.T) {
	if os.Getenv("BARNESS_MM_FORMAL") != "1" {
		t.Skip("explicit formal entry required")
	}
	driveMultimodal(t)
}

func driveMultimodal(t *testing.T) {
	if multimodalReport == nil {
		t.Skip("explicit isolated candidate route required")
	}
	r := multimodalReport
	fail := func(reason string) { r.Outcome, r.Failure = "FAIL", reason; t.Error(reason) }
	route, err := selectedMultimodalRoute()
	if err != nil {
		fail(err.Error())
		return
	}
	cs := run.Case(t, "MM11-configuration")
	cs.ReplayEnv("BARNESS_MM_ROUTE=" + route.name)
	cs.ReplayRun("^TestMultimodalCandidate$")
	catRaw, err := readDecisionsFile("testdata/multimodal/"+route.name+".json", 65536)
	formal := os.Getenv("BARNESS_MM_FORMAL") == "1"
	if formal {
		if route.name != "jev" {
			fail("formal multimodal evidence requires fixed Jev route")
			return
		}
		catRaw, err = json.MarshalIndent(ai.BuiltinCatalog(), "", "  ")
	}
	if os.Getenv("BARNESS_MM_CONTROLLED") == "wrong-catalog" {
		catRaw = append(catRaw, ' ')
	}
	if err != nil || (!formal && sha256Text(catRaw) != multimodalCatalogPins[route.name]) {
		fail("candidate file differs from fixed route pin")
		return
	}
	var cat ai.Catalog
	if json.Unmarshal(catRaw, &cat) != nil {
		fail("candidate catalog invalid")
		return
	}
	r.CatalogVersion, r.CatalogFileHash = cat.Version, sha256Text(catRaw)
	r.CatalogHash, err = cat.Hash()
	if err != nil {
		fail("candidate hash invalid")
		return
	}
	catalogFile := "candidate.json"
	if formal {
		catalogFile = "formal-catalog.json"
	}
	cs.Fixture(catalogFile, catRaw)
	run.SetVersion("model_catalog_version", r.CatalogVersion)
	run.SetVersion("model_catalog_hash", r.CatalogHash)
	// Seal actual working source bytes as well as the commit. A dirty source run
	// cannot silently borrow the identity of an earlier candidate package.
	r.SourceHashes = map[string]string{}
	paths := []string{"../../go.mod", "../../go.sum", "../../.scratch/jev-1.14-multimodal-box/verify-11-evidence.py"}
	if formal {
		path := "../live/testdata/jev-formal/sources.json"
		sources, e := os.ReadFile(path)
		if e != nil {
			fail("formal primary sources unavailable")
			return
		}
		cs.Fixture("formal-sources.json", sources)
		paths = append(paths, path)
	}
	err = filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		fail("source enumeration failed")
		return
	}
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			fail("nonregular source in seal")
			return
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			fail("source seal failed")
			return
		}
		r.SourceHashes[filepath.ToSlash(path)] = sha256Text(raw)
	}
	scenarios := multimodalScenarios(t, route.name)
	for _, sc := range scenarios {
		r.Expected = append(r.Expected, sc.id)
	}
	rec := newRecorder()
	defer rec.next.(*http.Transport).CloseIdleConnections()
	rec.requestLimit = 4 << 20
	if r.Mode == "controlled" {
		rec = controlledMultimodalRecorder(t, route, os.Getenv("BARNESS_MM_CONTROLLED"))
	}
	proof, notRun, err := multimodalConfiguration(route, rec)
	if err != nil {
		fail(err.Error())
		return
	}
	if notRun != "" {
		r.Failure = notRun
		t.Skip("NOT_RUN: " + notRun)
	}
	r.Authorization = &proof
	proofRaw, _ := json.Marshal(proof)
	r.AuthorizationHash = sha256Text(proofRaw)
	cs.Fixture("authorization.json", proofRaw)
	policy := multimodalPolicy()
	b := ai.Binding{TenantID: liveTenant, BindingID: "multimodal-" + route.name, Version: "candidate-v1", Enabled: true, Operation: ai.OperationClassifier, ProviderID: route.provider, API: route.api, Endpoint: route.endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: proof.AccountAlias + "@" + proof.RegionAlias, CredentialRef: "approved-candidate", AllowedModels: []string{route.model}, AllowClassifierImages: true}
	resolver := &resolver{binding: b, credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: b.CredentialRef, Version: "candidate-v1", BindingVersion: b.Version, AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(os.Getenv("BARNESS_MM_KEY"))}}
	bodies, gauges := provider.TrackBodies(rec), probe.New()
	client, err := ai.NewClient(ai.Config{Catalog: &cat, Policy: policy, Bindings: resolver, Credentials: resolver, Transport: &multimodalGate{route: route, next: bodies, report: r}, Observer: rec, Probe: gauges})
	if err != nil {
		fail("candidate Client assembly failed: " + errText(err))
		return
	}
	cs.Record("host-configuration", map[string]any{"binding": b, "policy": policy, "target": ai.Target{BindingID: b.BindingID, ModelID: route.model}, "catalog": cat})
	deadline := r.ExecutionDeadline
	if outer, ok := t.Deadline(); ok && outer.Add(-10*time.Second).Before(deadline) {
		deadline = outer.Add(-10 * time.Second)
	}
	r.ExecutionDeadline = deadline
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	fault := os.Getenv("BARNESS_MM_CONTROLLED")
	if fault == "deadline" {
		cancel()
	}
	applyMultimodalBudgetFault(r, fault)
	r.Outcome = "PASS"
	for _, sc := range scenarios {
		if fault == "body-budget" {
			sc.request.State = json.RawMessage(`"` + strings.Repeat("x", (4<<20)-1026) + `"`)
		}
		ok := t.Run(sc.id, func(t *testing.T) { executeMultimodal(t, ctx, client, rec, route, b, sc) })
		if !ok {
			r.Outcome = "FAIL"
			break
		}
	}
	resources := map[string]int64{"bodies": bodies.Open(), "calls": gauges.ActiveCalls(), "permits": gauges.Permits(), "waiters": gauges.AdmissionWaiters(), "events": gauges.QueuedEvents()}
	cs.Record("resources", resources)
	for name, value := range resources {
		cs.Check("released "+name, value == 0, "resource retained")
	}
	if fault == "audit" {
		cs.Record("observations-invalid", map[string]any{"imageData": "synthetic-audit-violation"})
	}
}

func applyMultimodalBudgetFault(r *multimodalExecution, fault string) {
	b, l := &r.Budget.Reserved, r.Budget.Limits
	switch fault {
	case "calls-budget":
		b.Calls = l.Calls
	case "attempts-budget":
		b.Attempts = l.Attempts
	case "questions-budget":
		b.Questions = l.Questions
	case "sources-budget":
		b.Sources = l.Sources
	case "references-budget":
		b.References = l.References
	case "source-bytes-budget":
		b.SourceBytes = l.SourceBytes
	case "expanded-bytes-budget":
		b.ExpandedBytes = l.ExpandedBytes
	case "state-budget":
		b.StateBytes = l.StateBytes
	}
}

func executeMultimodal(t *testing.T, ctx context.Context, client *ai.Client, rec *recorder, route multimodalRoute, b ai.Binding, sc multimodalScenario) {
	r := multimodalReport
	cs := run.Case(t, "MM11-"+sc.id)
	if r.Mode == "controlled" {
		cs.ReplayRun("^TestMultimodalCandidateProcess$/^" + route.name + "$/^" + os.Getenv("BARNESS_MM_CONTROLLED") + "$")
	} else {
		cs.ReplayEnv("BARNESS_MM_ROUTE=" + route.name + " BARNESS_MM_LIVE=1")
		cs.ReplayRun("^TestMultimodalCandidate$")
	}
	cs.Record("public-request", sc.request)
	recordCandidateImages(cs, sc.request.Images)
	refs := sc.request.ImageReferences
	if len(refs) == 0 {
		for i := range sc.request.Images {
			refs = append(refs, ai.ImageReference{ImageIndex: i, Placement: ai.ImagePlacementContext})
		}
	}
	size := 0
	for _, img := range multimodalImages(t) {
		raw, _ := decodeCandidateImage(img)
		size = max(size, len(raw))
	}
	reservation := multimodalResources{1, 1, len(sc.request.Questions), len(sc.request.Images), len(refs), len(sc.request.Images) * size, len(refs) * size, len(sc.request.State) + 1024*len(refs)}
	next := r.Budget.Reserved.add(reservation)
	if !cs.Check("reserve before call", ctx.Err() == nil && reservation.fits(r.Budget.Limits) && next.fits(r.Budget.Limits), "finite cumulative budget exhausted") {
		return
	}
	r.Budget.Reserved = next
	r.Budget.CallsUsed++
	r.Budget.Actual.Calls++
	cs.Record("reservation", reservation)
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	scope := newScope()
	scope.JobID = "multimodal-" + route.name
	scope.ActorID = "authorized-maintainer"
	target := ai.Target{BindingID: b.BindingID, ModelID: route.model}
	var opts ai.Options = ai.OpenAIDecisionsOptions{}
	if route.name == "jev" {
		opts = ai.TypeSafeOptions{ExperimentalImages: true}
	}
	var res ai.ClassifierResult
	var err error
	began := time.Now()
	if sc.entry == "hooked" {
		res, err = client.WithHooks(multimodalHooks(multimodalImages(t))).Classify(callCtx, scope, target, sc.request, opts)
	} else {
		res, err = client.Classify(callCtx, scope, target, sc.request, opts)
	}
	duration := time.Since(began)
	cs.Record("duration", map[string]int64{"milliseconds": duration.Milliseconds()})
	ex := rec.take()
	obs := rec.observationsOf(scope.RequestID)
	r.Budget.HTTPAttempts += len(ex)
	r.Budget.ReportedInputTokens += int(res.Usage.Input)
	r.Budget.ReportedOutputTokens += int(res.Usage.Output)
	cs.Record("result", map[string]any{"result": res, "error": errText(err)})
	cs.Record("exchanges", ex)
	cs.Record("observations", obs)
	cs.Record("observer-stats", client.ObserverStats())
	if !cs.Check("own protocol call succeeds", err == nil, "%s", errText(err)) {
		return
	}
	cs.Check("call within time budget", duration <= 30*time.Second, "call exceeded 30 seconds")
	cs.Check("fixed attribution", res.Metadata.ModelID == route.model && res.Metadata.API == route.api && res.Metadata.ProviderID == route.provider && res.Metadata.CatalogHash == r.CatalogHash && res.Metadata.AccountScopeID == b.AccountScopeID, "wrong identity")
	if cs.Check("one final bounded attempt", len(ex) == 1 && len(res.Metadata.Attempts) == 1, "unexpected attempts") {
		cs.Check("official selected service", ex[0].URL == route.endpoint+route.path && ex[0].Method == "POST" && !ex[0].Truncated && len(ex[0].RequestBody) <= 4<<20, "wrong route/body")
		cs.Check("attempt correlation", res.Metadata.Attempts[0].AttemptID == scope.RequestID+"#1" && res.Metadata.Attempts[0].RetryDelay == 0 && res.Metadata.Attempts[0].HTTPStatus == 200, "attempt mismatch")
		actual, parseErr := multimodalNativeResources(route, ex[0].RequestBody)
		cs.Check("final resources fit reservation", parseErr == nil && actual.fits(reservation), "final native input exceeded reservation")
		validateMultimodal(cs, route, sc, res, ex[0], obs)
	}
	r.Scenarios = append(r.Scenarios, sc.id)
	cs.Record("budget", r.Budget)
}
