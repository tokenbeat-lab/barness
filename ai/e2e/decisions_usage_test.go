package e2e

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const decisionsFullUsage = `"usage":{"input_tokens":296,"output_tokens":20,"total_tokens":316,"input_tokens_details":{"cached_tokens":100,"cache_write_tokens":50},"output_tokens_details":{"reasoning_tokens":7}}`

type decisionsUsageCase struct {
	name      string
	fields    string
	want      ai.Usage
	reporting ai.UsageReporting
	invalid   bool
	badAnswer bool
	noModel   bool
}

func decisionsKnownUsage() ai.Usage {
	return ai.Usage{Input: 296, Output: 20, TotalTokens: 316, Reasoning: ai.Value[int64](7), Cost: ai.UsageCost{Input: .0000296, Total: .0000296}}
}

// Only the public unary seam is used. Raw response fixtures/captures retain
// omitted/null/zero detail presence without adding raw details to Observer.
func checkDecisionsUsage(t *testing.T, cases []decisionsUsageCase) {
	t.Helper()
	for _, sc := range cases {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD35-38-usage-"+sc.name+"-"+entry)
				var catalog ai.Catalog
				u := newDecisionsWorld(t, func(c *ai.Config) {
					c.Catalog.Version = "synthetic-decisions-input-prices-v1"
					m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
					m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: .1, Output: 80, CacheRead: 40, CacheWrite: 40}, Tiers: []ai.CostTier{{InputTokensAbove: 272000, CostRates: ai.CostRates{Input: .2, Output: 160, CacheRead: 80, CacheWrite: 80}}}}
					// Two chat identities deliberately share the request/response IDs
					// with incompatible rates. Classification must use its own price.
					for _, id := range []string{"host-decisions", "response-version"} {
						chat := c.Catalog.Models[0]
						chat.ID, chat.Provider, chat.API = id, ai.ProviderOpenAI, ai.APIOpenAIResponses
						chat.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: 999, Output: 999}}
						c.Catalog.Models = append(c.Catalog.Models, chat)
					}
					catalog = *c.Catalog
				})
				hash, hashErr := catalog.Hash()
				if hashErr != nil {
					t.Fatal(hashErr)
				}
				answer := decisionsAnswer
				if sc.badAnswer {
					answer = strings.Replace(answer, `"choice":"track"`, `"choice":"foreign"`, 1)
				}
				body := `{"model":"response-version","answers":[` + answer + `]`
				if sc.noModel {
					body = `{"answers":[` + answer + `]`
				}
				if sc.fields != "" {
					body += "," + sc.fields
				}
				body += "}"
				ev.Fixture("response.json", []byte(body))
				enqueue(ev, u.world, fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-usage"}, Body: body}.script(t, decisionsProtocol, ""))
				// Catalog is caller-owned and was copied at NewClient. Mutating
				// rates now cannot rewrite the pinned estimate or snapshot.
				catalog.ClassifierModels[len(catalog.ClassifierModels)-1].Cost.Input = 900
				id := "req-usage-" + sc.name + "-" + entry
				res, err := decisionsPublicCall(t, u, entry, id, decisionsInput(), ai.Hooks{})
				ev.Record("result", res)
				ev.Record("expected", map[string]any{"usage": sc.want, "reporting": sc.reporting, "protocolError": sc.invalid || sc.badAnswer, "catalogVersion": catalog.Version, "catalogHash": hash, "billingReconciled": false, "regionalPremiumIncluded": false})
				if sc.invalid || sc.badAnswer {
					checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
				} else {
					ev.Check("valid answer published", err == nil && res.StopReason == ai.StopReasonStop && len(res.Answers) == 1, "got %v / %+v", err, res)
				}
				actualCounts, expectedCounts := res.Usage, sc.want
				actualCounts.Cost, expectedCounts.Cost = ai.UsageCost{}, ai.UsageCost{}
				ev.Check("exact known counts and reasoning presence", reflect.DeepEqual(actualCounts, expectedCounts), "got %+v want %+v", res.Usage, sc.want)
				// Decimal USD examples are independent worked totals. ModelCost
				// deliberately retains pi's IEEE operation order; tolerate only
				// rounding here while all free cost buckets remain exactly zero.
				cost := res.Usage.Cost
				ev.Check("input-only estimate", math.Abs(cost.Input-sc.want.Cost.Input) < 1e-15 && cost.Total == cost.Input && cost.Output == 0 && cost.CacheRead == 0 && cost.CacheWrite == 0, "got %+v want %+v", cost, sc.want.Cost)
				ev.Check("authorized identity and price snapshot", res.Metadata.ModelID == "host-decisions" && res.Metadata.Operation == ai.OperationClassifier && res.Metadata.CatalogVersion == catalog.Version && res.Metadata.CatalogHash == hash, "got %+v", res.Metadata)
				responseModel, reportedModel := res.ResponseModel.Get()
				ev.Check("response identity presence stays separate", (sc.invalid || reportedModel == !sc.noModel) && (!reportedModel || responseModel == "response-version"), "got %+v", res.ResponseModel)
				if ev.Check("2xx has one attempt and no replay", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata) {
					a := res.Metadata.Attempts[0]
					ev.Check("attempt presence and estimate agree", a.UsageReporting == sc.reporting && reflect.DeepEqual(a.Usage, res.Usage) && a.HTTPStatus == 200 && a.ProviderRequestID == "req-usage" && a.RetryDelay == 0, "got %+v", a)
				}
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsUsagePresence(t *testing.T) {
	known := decisionsKnownUsage()
	basic := known
	basic.Reasoning = ai.Nullable[int64]{}
	cases := []decisionsUsageCase{
		{name: "missing", reporting: ai.UsageUnreported},
		{name: "null", fields: `"usage":null`, reporting: ai.UsageUnreported},
		{name: "empty", fields: `"usage":{}`, reporting: ai.UsagePartial},
		{name: "basic", fields: decisionsUsage, want: basic, reporting: ai.UsagePartial},
		{name: "complete", fields: decisionsFullUsage, want: known, reporting: ai.UsageComplete},
		{name: "no-model", fields: decisionsFullUsage, want: known, reporting: ai.UsageComplete, noModel: true},
		{name: "cache-reasoning-at-limit", fields: strings.ReplaceAll(strings.ReplaceAll(decisionsFullUsage, `"cached_tokens":100`, `"cached_tokens":246`), `"reasoning_tokens":7`, `"reasoning_tokens":20`), want: func() ai.Usage { u := known; u.Reasoning = ai.Value[int64](20); return u }(), reporting: ai.UsageComplete},
		{name: "complete-zero", fields: `"usage":{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, want: ai.Usage{Reasoning: ai.Value[int64](0)}, reporting: ai.UsageComplete},
		{name: "reasoning-zero", fields: strings.Replace(decisionsFullUsage, `"reasoning_tokens":7`, `"reasoning_tokens":0`, 1), want: func() ai.Usage { u := known; u.Reasoning = ai.Value[int64](0); return u }(), reporting: ai.UsageComplete},
		{name: "total-only", fields: `"usage":{"total_tokens":999}`, want: ai.Usage{TotalTokens: 999}, reporting: ai.UsagePartial},
		{name: "input-only", fields: `"usage":{"input_tokens":296}`, want: ai.Usage{Input: 296, TotalTokens: 296, Cost: known.Cost}, reporting: ai.UsagePartial},
		{name: "output-only", fields: `"usage":{"output_tokens":20}`, want: ai.Usage{Output: 20, TotalTokens: 20}, reporting: ai.UsagePartial},
		{name: "cache-without-input", fields: `"usage":{"input_tokens_details":{"cached_tokens":100,"cache_write_tokens":50}}`, reporting: ai.UsagePartial},
		{name: "reasoning-without-output", fields: `"usage":{"output_tokens_details":{"reasoning_tokens":7}}`, want: ai.Usage{Reasoning: ai.Value[int64](7)}, reporting: ai.UsagePartial},
		{name: "vendor-metadata", fields: strings.TrimSuffix(decisionsFullUsage, "}") + `,"vendor_metadata":{"version":1}}`, want: known, reporting: ai.UsageComplete},
		{name: "broken-answer", fields: decisionsFullUsage, want: known, reporting: ai.UsageComplete, badAnswer: true},
	}
	for _, field := range []struct{ key, value string }{{"input_tokens", "296"}, {"output_tokens", "20"}, {"total_tokens", "316"}, {"cached_tokens", "100"}, {"cache_write_tokens", "50"}, {"reasoning_tokens", "7"}} {
		for _, mode := range []string{"missing", "null"} {
			var usage map[string]any
			mustUnmarshal(t, []byte(strings.TrimPrefix(decisionsFullUsage, `"usage":`)), &usage)
			parent := usage
			if field.key == "cached_tokens" || field.key == "cache_write_tokens" {
				parent = usage["input_tokens_details"].(map[string]any)
			} else if field.key == "reasoning_tokens" {
				parent = usage["output_tokens_details"].(map[string]any)
			}
			delete(parent, field.key)
			if mode == "null" {
				parent[field.key] = nil
			}
			want := known
			switch field.key {
			case "input_tokens":
				want.Input, want.Cost = 0, ai.UsageCost{}
			case "output_tokens":
				want.Output = 0
			case "reasoning_tokens":
				want.Reasoning = ai.Nullable[int64]{}
			}
			cases = append(cases, decisionsUsageCase{name: mode + "-" + field.key, fields: `"usage":` + string(mustMarshal(t, usage)), want: want, reporting: ai.UsagePartial})
		}
	}
	for _, details := range []string{"input_tokens_details", "output_tokens_details"} {
		for _, mode := range []string{"missing", "null"} {
			var usage map[string]json.RawMessage
			mustUnmarshal(t, []byte(strings.TrimPrefix(decisionsFullUsage, `"usage":`)), &usage)
			delete(usage, details)
			if mode == "null" {
				usage[details] = json.RawMessage(`null`)
			}
			want := known
			if details == "output_tokens_details" {
				want.Reasoning = ai.Nullable[int64]{}
			}
			cases = append(cases, decisionsUsageCase{name: mode + "-" + details, fields: `"usage":` + string(mustMarshal(t, usage)), want: want, reporting: ai.UsagePartial})
		}
	}
	checkDecisionsUsage(t, cases)
}
