package e2e

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Issue 07's agreed seams are public Classify/HookedClient.Classify, Catalog,
// and the controlled HTTP provider. Native scripts are shared with frozen pi;
// expected wires/answers are literals, independent of either adapter.
type decisionsDiffCase struct {
	ID               string               `json:"id"`
	Input            ai.ClassifierRequest `json:"input"`
	Payload          json.RawMessage      `json:"payload,omitempty"`
	Replies          []fixtureReply       `json:"replies"`
	Retries          int                  `json:"retries"`
	BarnessRequests  int                  `json:"barnessRequests"`
	PiRequests       int                  `json:"piRequests"`
	BarnessWire      json.RawMessage      `json:"barnessWire"`
	PiWire           json.RawMessage      `json:"piWire"`
	BarnessAnswers   json.RawMessage      `json:"barnessAnswers"`
	PiAnswers        json.RawMessage      `json:"piAnswers"`
	BarnessStop      string               `json:"barnessStop"`
	PiStop           string               `json:"piStop"`
	Reporting        ai.UsageReporting    `json:"reporting"`
	Usage            ai.Usage             `json:"usage"`
	BarnessCallbacks []string             `json:"barnessCallbacks"`
	PiCallbacks      []string             `json:"piCallbacks"`
}

func TestDecisionsPiDifferential(t *testing.T) {
	for _, name := range []string{"parity", "strict", "runtime"} {
		raw, err := os.ReadFile("testdata/decisions/differential-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var cases []decisionsDiffCase
		mustUnmarshal(t, raw, &cases)
		for _, sc := range cases {
			for _, entry := range []string{"plain", "hooked"} {
				if len(sc.Payload) != 0 && entry == "plain" {
					continue
				}
				t.Run(sc.ID+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "PIDIFF-P10-"+sc.ID+"-"+entry+"-classify")
					ev.ReplayEnv(pioracle.EnableEnv + "=1")
					ev.Fixture("fixture.json", raw)
					o := openOracle(t)
					decisionsDifferential(t, ev, o, sc, entry)
				})
			}
		}
	}
}

func decisionsDifferential(t *testing.T, ev *evidence.Case, o *pioracle.Oracle, sc decisionsDiffCase, entry string) {
	u := newClassifierImagesWorld(t, func(c *ai.Config) {
		m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
		m.ID, m.Name, m.ContextWindow = "gpt-6-luna", "GPT-6 Luna", 922000
		m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}
		m.Capabilities.MinScoreLevels, m.Capabilities.MaxScoreLevels = 2, 10
		m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: .1}, Tiers: []ai.CostTier{{InputTokensAbove: 272000, CostRates: ai.CostRates{Input: .2}}}}
	})
	u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"gpt-6-luna"}; b.Retry.MaxRetries = sc.Retries })
	replies := make([]provider.Reply, len(sc.Replies))
	for i, r := range sc.Replies {
		replies[i] = r.script(t, decisionsProtocol, "")
	}
	enqueue(ev, u.world, replies...)
	callbacks := []string{}
	hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
		callbacks = append(callbacks, "payload")
		if len(sc.Payload) != 0 {
			return ai.ReplacePayload(jsonObject(t, sc.Payload)), nil
		}
		return ai.KeepPayload(), nil
	}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
		callbacks = append(callbacks, "response")
		return nil
	}}
	target := ai.Target{BindingID: "classifier", ModelID: "gpt-6-luna"}
	client := u.client.WithHooks(hooks)
	var res ai.ClassifierResult
	var callErr error
	if entry == "plain" {
		res, callErr = u.client.Classify(ctxFor(t), textScope(ev.ID()), target, sc.Input, nil)
	} else {
		res, callErr = client.Classify(ctxFor(t), textScope(ev.ID()), target, sc.Input, nil)
	}
	u.records(ev, res, callErr)
	u.released(ev)
	gotUsage, wantUsage := res.Usage, sc.Usage
	gotCost, wantCost := gotUsage.Cost, wantUsage.Cost
	gotUsage.Cost, wantUsage.Cost = ai.UsageCost{}, ai.UsageCost{}
	ev.Check("own exact usage and reasoning presence", jsonEqual(mustMarshal(t, gotUsage), mustMarshal(t, wantUsage)), "got %+v", res.Usage)
	ev.Check("own input-only estimate", math.Abs(gotCost.Input-wantCost.Input) < 1e-15 && gotCost.Total == gotCost.Input && gotCost.Output == 0 && gotCost.CacheRead == 0 && gotCost.CacheWrite == 0, "got %+v", gotCost)
	ev.Check("public terminal", string(res.StopReason) == sc.BarnessStop && (callErr == nil) == (sc.BarnessStop == "stop"), "got %+v / %v", res, callErr)
	ev.Check("public named answers", jsonEqual(mustMarshal(t, decisionsPiResult(t, res)["answers"]), sc.BarnessAnswers), "got %+v", res.Answers)
	if len(res.Metadata.Attempts) > 0 {
		ev.Check("public usage presence", res.Metadata.Attempts[len(res.Metadata.Attempts)-1].UsageReporting == sc.Reporting, "got %+v", res.Metadata.Attempts)
	}
	b := pioracle.Observation{Result: mustMarshal(t, decisionsPiResult(t, res)), Events: callbackEvents(t, callbacks)}
	bn := observeRequests(ev, "barness-requests", u.provider, &b)
	srv := provider.New(map[string]string{tenantA.secret: tenantA.alias})
	defer srv.Close()
	srv.Enqueue(replies...)
	piCase := pioracle.Case{API: string(ai.APIOpenAIDecisions), Provider: "openai", Model: "gpt-6-luna", BaseURL: srv.URL() + "/v1", APIKey: tenantA.secret, Entry: "classify", Context: mustMarshal(t, sc.Input), Options: mustMarshal(t, map[string]any{"maxRetries": sc.Retries})}
	// Hooks are declarative fixture input to the runner, never executable code.
	if entry == "hooked" {
		piCase.TraceHooks = true
		piCase.Payload = sc.Payload
	}
	piRun, err := o.Run(ctxFor(t), piCase)
	if !ev.Check("frozen Decisions completed", err == nil, "%v", err) {
		return
	}
	pi, pn := piObservation(ev, srv, piRun)
	var result struct {
		Answers json.RawMessage `json:"answers"`
		Stop    string          `json:"stopReason"`
	}
	mustUnmarshal(t, pi.Result, &result)
	ev.Check("frozen named answers and terminal", result.Stop == sc.PiStop && jsonEqual(result.Answers, sc.PiAnswers), "got %s", pi.Result)
	ev.Check("each side's exact attempt count", bn == sc.BarnessRequests && pn == sc.PiRequests, "barness=%d pi=%d", bn, pn)
	for _, side := range []struct {
		name     string
		requests []provider.Request
		wire     json.RawMessage
	}{{"barness", u.provider.Requests(), sc.BarnessWire}, {"pi", srv.Requests(), sc.PiWire}} {
		for _, r := range side.requests {
			ev.Check(side.name+" full frozen wire on every attempt", jsonEqual(r.Body, side.wire), "got %s", r.Body)
			ev.Check(side.name+" complete target and credential", r.Method == "POST" && r.Path == "/v1/decisions" && r.Query == "" && r.KeyAlias == tenantA.alias, "got %+v", r)
			ev.Check(side.name+" exact Content-Length", r.Header["Content-Length"] == strconv.Itoa(len(r.Body)), "got %q size=%d", r.Header["Content-Length"], len(r.Body))
			if side.name == "pi" {
				for key, value := range map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip, deflate", "Accept-Language": "*", "Connection": "keep-alive", "Sec-Fetch-Mode": "cors", "User-Agent": "node"} {
					ev.Check("own pi header "+key, r.Header[key] == value, "got %q", r.Header[key])
				}
			} else {
				ev.Check("own barness runtime headers", r.Header["Accept-Encoding"] == "gzip" && r.Header["User-Agent"] == "barness-ai", "got %+v", r.Header)
			}
		}
	}
	if entry == "hooked" {
		ev.Check("barness callback timing", jsonEqual(mustMarshal(t, callbacks), mustMarshal(t, sc.BarnessCallbacks)), "got %v", callbacks)
		ev.Check("pi callback timing", jsonEqual(mustMarshal(t, piRun.Events), mustMarshal(t, callbackEvents(t, sc.PiCallbacks))), "got %s", mustMarshal(t, piRun.Events))
	}
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", b)
	diffs, err := pioracle.Compare(pi, b)
	if !ev.Check("complete observations compare", err == nil, "%v", err) {
		return
	}
	v := loadLedger(t).Classify("openai-decisions", ev.ID(), diffs)
	checkDecisionsLedger(ev, diffs)
	ev.Record("pidiff", o.Record(ev.ID(), "openai-decisions", v, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version()}, NodeVersion: piRun.Node, ModelCatalogHash: catalogHash(t, u.client.Catalog()), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), PiResult: pi.Result, BarnessResult: b.Result, Frames: mustMarshal(t, replies)}))
	ev.Check("zero pending Decisions differences", v.Pass, "%+v", v.Findings)
}

func checkDecisionsLedger(ev *evidence.Case, diffs []pioracle.Diff) {
	ledger := loadLedger(ev.T())
	for _, diff := range diffs {
		v := ledger.Classify("openai-decisions", "PIDIFF-P10-unregistered-classify", []pioracle.Diff{diff})
		ev.Check("extension stays pending outside its finite cases: "+diff.Path, !v.Pass && v.Pending == 1, "got %+v", v)
	}
	v := ledger.Classify("openai-decisions", ev.ID(), []pioracle.Diff{{Path: "request.body.model", Kind: pioracle.Changed, Pi: json.RawMessage(`"gpt-6-luna"`), Barness: json.RawMessage(`"wrong"`)}})
	ev.Check("common authorized model remains compared", !v.Pass && v.Pending == 1, "got %+v", v)
}

func callbackEvents(t *testing.T, names []string) []json.RawMessage {
	events := []json.RawMessage{}
	for _, name := range names {
		events = append(events, mustMarshal(t, map[string]string{"type": "callback", "name": name}))
	}
	return events
}

// P10 retains score details, refusal, cost and usage presence as observable
// fields. No P07 field or existing projection is removed to make it pass.
func decisionsPiResult(t *testing.T, res ai.ClassifierResult) map[string]any {
	answers := map[string]any{}
	for name, answer := range res.Answers {
		answers[name] = jsonObject(t, answer)
	}
	result := map[string]any{"api": res.Metadata.API, "provider": res.Metadata.ProviderID, "model": res.Metadata.ModelID, "answers": answers, "stopReason": res.StopReason, "timestamp": 0}
	if res.ErrorMessage != "" {
		result["errorMessage"] = res.ErrorMessage
	}
	if len(res.Metadata.Attempts) > 0 {
		a := res.Metadata.Attempts[len(res.Metadata.Attempts)-1]
		result["usageReporting"] = a.UsageReporting
		if a.UsageReporting != ai.UsageUnreported {
			result["usage"] = res.Usage
		}
	}
	return result
}
