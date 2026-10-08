package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"runtime"
)

func classifierPiDifferential(t *testing.T, ledger pioracle.Ledger) {
	f, raw := loadFixture(t, classifierProtocol, "parity.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID+"/classify", func(t *testing.T) {
			ev := run.Case(t, "PIDIFF-P07-"+sc.ID+"-classify")
			ev.ReplayEnv(pioracle.EnableEnv + "=1")
			ev.Fixture("fixture.json", raw)
			o := openOracle(t)
			w := scenarioWorld(t, sc)
			replies := sc.replies(t)
			enqueue(ev, w, replies...)
			res, err := w.client.Classify(ctxFor(t), textScope("req-classifier-pidiff"), sc.target(), classifierInput(t, sc), nil)
			ev.Record("barness-result", res)
			ev.Record("barness-error", errString(err))
			b := pioracle.Observation{Result: mustMarshal(t, piClassifierResult(t, res))}
			bn := observeRequests(ev, "barness-requests", w.provider, &b)
			srv := provider.New(map[string]string{tenantA.secret: tenantA.alias})
			defer srv.Close()
			srv.Enqueue(replies...)
			piRun, err := o.Run(ctxFor(t), pioracle.Case{API: string(ai.APITypeSafeSystemOne), Provider: string(ai.ProviderTypeSafe), Model: "jev-latest", ModelPatch: mustMarshal(t, map[string]any{"id": "jev-1.13.0"}), BaseURL: srv.URL() + "/v1", APIKey: tenantA.secret, Entry: "classify", Context: sc.Classifier, Options: mustMarshal(t, map[string]any{"maxRetries": sc.MaxRetries})})
			if !ev.Check("classifier oracle completed", err == nil, "%v", err) {
				return
			}
			pi, pn := piObservation(ev, srv, piRun)
			// Cost and response model have explicit extension decisions. Only
			// cost is removed from pi, whose other fields must all be compared.
			pi.Result = classifierWithoutCost(t, pi.Result)
			ev.Check("same retry count as binding", pn == sc.Expect.Requests && bn == sc.Expect.Requests, "pi=%d barness=%d", pn, bn)
			// Each retry must carry the same authorized final request.
			for _, r := range append(w.provider.Requests(), srv.Requests()...) {
				ev.Check("version fixed in every attempt", jsonEqual(r.Body, sc.Expect.Request.Body), "got %s", r.Body)
			}
			ev.Record("pi-observation", pi)
			ev.Record("barness-observation", b)
			diffs, err := pioracle.Compare(pi, b)
			if !ev.Check("classifier observations compare", err == nil, "%v", err) {
				return
			}
			verdict := ledger.Classify(string(ai.APITypeSafeSystemOne), ev.ID(), diffs)
			hash, err := w.client.Catalog().Hash()
			if err != nil {
				t.Fatal(err)
			}
			rec := o.Record(ev.ID(), string(ai.APITypeSafeSystemOne), verdict, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version()}, NodeVersion: piRun.Node, ModelCatalogHash: hash, PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), Frames: []byte(replyScripts(replies))})
			ev.Record("pidiff", rec)
			ev.Check("classifier differential zero pending", verdict.Pass, "%+v", verdict.Findings)
		})
	}
}

func piClassifierResult(t *testing.T, res ai.ClassifierResult) map[string]any {
	answers := map[string]any{}
	for key, a := range res.Answers {
		answer := jsonObject(t, a)
		if _, ok := a.(ai.ScoreAnswer); ok {
			delete(answer, "probabilities")
			delete(answer, "legend")
		}
		answers[key] = answer
	}
	out := map[string]any{"api": res.Metadata.API, "provider": res.Metadata.ProviderID, "model": res.Metadata.ModelID, "answers": answers, "stopReason": res.StopReason, "timestamp": 0}
	if res.ErrorMessage != "" {
		out["errorMessage"] = res.ErrorMessage
	}
	if len(res.Metadata.Attempts) > 0 && res.Metadata.Attempts[len(res.Metadata.Attempts)-1].UsageReporting != ai.UsageUnreported {
		u := jsonObject(t, res.Usage)
		delete(u, "cost")
		out["usage"] = u
	}
	return out
}

func classifierWithoutCost(t *testing.T, raw json.RawMessage) json.RawMessage {
	var value map[string]json.RawMessage
	mustUnmarshal(t, raw, &value)
	if u, ok := value["usage"]; ok {
		var usage map[string]json.RawMessage
		mustUnmarshal(t, u, &usage)
		delete(usage, "cost")
		value["usage"] = mustMarshal(t, usage)
	}
	return mustMarshal(t, value)
}

func TestClassifierCatalogParity(t *testing.T) {
	ev := run.Case(t, "PIDIFF-P07-E11-catalog")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	o := openOracle(t)
	models, err := o.Models(ctxFor(t), []pioracle.ModelKey{{Provider: "typesafe", API: "typesafe-system-one", ID: "jev-latest"}})
	if !ev.Check("pi classifier catalog entry found", err == nil && len(models) == 1 && string(models[0]) != "null", "%v / %s", err, fmt.Sprint(models)) {
		return
	}
	ev.Record("pi-catalog", models)
	var fields map[string]json.RawMessage
	mustUnmarshal(t, models[0], &fields)
	// The synthetic host catalog is the same public boundary used by calls.
	// Jev's ID and cost are explicit extensions; pi has no capabilities fields.
	f, _ := loadFixture(t, classifierProtocol, "parity.json")
	w := scenarioWorld(t, f.Scenarios[0])
	m := w.client.Catalog().ClassifierModels[0]
	want := map[string]any{"type": "classifier", "api": m.API, "provider": m.Provider, "contextWindow": m.ContextWindow, "name": m.Name}
	for key, value := range want {
		ev.Check("shared catalog field "+key, jsonEqual(fields[key], mustMarshal(t, value)), "got %s", fields[key])
	}
	for _, key := range []string{"type", "api", "provider", "contextWindow", "name", "id", "cost", "baseUrl", "input"} {
		delete(fields, key)
	}
	ev.Check("no unaccounted pi catalog fields", len(fields) == 0, "uncompared %v", fields)
	// Re-read fields for the explicitly registered alias, cost and host-owned
	// endpoint / implicit text modality checks rather than dropping them silently.
	mustUnmarshal(t, models[0], &fields)
	ev.Check("pi endpoint is binding configuration", string(fields["baseUrl"]) == `"https://api.typesafe.ai/v1/"`, "got %s", fields["baseUrl"])
	ev.Check("classifier implicit modality is text", string(fields["input"]) == `["text"]`, "got %s", fields["input"])
	ev.Check("pi zero catalog cost is registered", jsonEqual(fields["cost"], []byte(`{"input":0,"output":0,"cacheRead":0,"cacheWrite":0}`)), "got %s", fields["cost"])
	ev.Check("frozen alias remains pi's", string(fields["id"]) == `"jev-latest"`, "got %s", fields["id"])
	// Catalog comparisons also need a differential verdict for the release
	// gate. Keep pi's whole entry; host-owned endpoint and implicit modality
	// are named ledger differences rather than silently deleted fields.
	b := pioracle.Observation{Result: mustMarshal(t, map[string]any{
		"type": "classifier", "provider": m.Provider, "api": m.API, "id": m.ID,
		"name": m.Name, "contextWindow": m.ContextWindow, "cost": m.Cost,
	})}
	pi := pioracle.Observation{Result: models[0]}
	diffs, err := pioracle.Compare(pi, b)
	if !ev.Check("classifier catalog compares", err == nil, "%v", err) {
		return
	}
	v := loadLedger(t).Classify(string(ai.APITypeSafeSystemOne), ev.ID(), diffs)
	hash, err := w.client.Catalog().Hash()
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("pidiff", o.Record(ev.ID(), string(ai.APITypeSafeSystemOne), v, pioracle.Sides{
		BarnessVersions: map[string]string{"go": runtime.Version()}, ModelCatalogHash: hash,
		PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), Frames: models[0],
	}))
	ev.Check("classifier catalog zero pending", v.Pass, "%+v", v.Findings)
}
