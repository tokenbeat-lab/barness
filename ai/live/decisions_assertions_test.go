//go:build live

package live

import (
	"encoding/json"
	"math"
	"reflect"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func costClose(a, b float64) bool { return math.Abs(a-b) < 1e-12 }
func validateCandidateMetadata(cs *evidence.Case, r *candidateDecisionsReport, scope ai.CallScope, res ai.ClassifierResult, ex []exchange, obs []ai.Observation) {
	m := res.Metadata
	cs.Check("own immutable attribution", m.Resolved && m.TenantID == scope.TenantID && m.RequestID == scope.RequestID && m.ActorID == scope.ActorID && m.JobID == scope.JobID && m.BindingID == decisionsCandidate.name && m.Operation == ai.OperationClassifier && m.ProviderID == ai.ProviderOpenAI && m.API == ai.APIOpenAIDecisions && m.ModelID == "gpt-6-luna" && m.AccountScopeID == r.AccountScopeID, "wrong call attribution")
	cs.Check("own candidate price identity", m.CatalogVersion == r.CatalogVersion && m.CatalogHash == r.CatalogHash && m.BindingVersion == "candidate" && m.CredentialVersion == "candidate", "wrong configuration snapshot")
	if !cs.Check("one reserved official attempt", len(m.Attempts) == 1 && len(ex) == 1, "attempts %d exchanges %d", len(m.Attempts), len(ex)) {
		return
	}
	a := m.Attempts[0]
	cs.Check("official route and complete bounded capture", ex[0].Method == "POST" && ex[0].URL == "https://api.openai.com/v1/decisions" && !ex[0].Truncated, "wrong or truncated request")
	cs.Check("attempt identity and observed provider ID", a.AttemptID == scope.RequestID+"#1" && a.RetryDelay == 0 && a.HTTPStatus == ex[0].Status && (ex[0].Status == 0 || a.ProviderRequestID != ""), "missing attempt evidence")
	cs.Check("usage presence never inferred from zero", slices.Contains([]ai.UsageReporting{ai.UsageComplete, ai.UsagePartial, ai.UsageUnreported}, a.UsageReporting) && reflect.DeepEqual(a.Usage, res.Usage), "inconsistent attempt usage")
	if cs.Check("complete own Observer sequence", len(obs) == 4, "observations %d", len(obs)) {
		kinds := []ai.ObservationKind{ai.ObservationCallStarted, ai.ObservationAttemptStarted, ai.ObservationAttemptFinished, ai.ObservationCallFinished}
		for i, o := range obs {
			cs.Check("Observer kind and attribution", o.Kind == kinds[i] && o.Call.RequestID == scope.RequestID && o.Call.TenantID == scope.TenantID && o.Call.Operation == ai.OperationClassifier, "wrong observation")
		}
		cs.Check("Observer terminal agrees with result", reflect.DeepEqual(obs[3].Call, m) && obs[3].StopReason == res.StopReason && reflect.DeepEqual(obs[3].Usage, res.Usage), "wrong terminal metadata")
	}
	// Captures of a damaged 2xx response remain intact for attribution. No
	// tolerance is silently relaxed to turn a counterexample into a PASS.
	var body map[string]json.RawMessage
	cs.Check("minimal native body", json.Unmarshal([]byte(ex[0].RequestBody), &body) == nil && len(body) == 3 && body["model"] != nil && body["input"] != nil && body["questions"] != nil, "unexpected native fields")
}
func validateCandidateAnswers(cs *evidence.Case, req ai.ClassifierRequest, res ai.ClassifierResult) {
	actual, reported := res.ResponseModel.Get()
	cs.Check("actual response model is reported separately", reported && strings.TrimSpace(actual) != "" && res.Metadata.ModelID == "gpt-6-luna", "actual model missing; never fabricate from requested ID")
	cs.Check("successful full typed questions", res.StopReason == ai.StopReasonStop && len(res.Answers) == len(req.Questions), "missing question answers")
	for id, q := range req.Questions {
		answer, ok := res.Answers[id]
		if !cs.Check("own question name", ok, "missing answer %s", id) {
			continue
		}
		if _, refusal := answer.(ai.RefusalAnswer); refusal {
			cs.Record("refusal-"+id, answer)
			continue
		}
		switch question := q.(type) {
		case ai.ChoiceQuestion:
			a, typed := answer.(ai.ChoiceAnswer)
			cs.Check("typed choice and own candidate set", typed && len(a.Probabilities) == len(question.Criteria), "wrong choice %s", id)
			sum := 0.0
			highest := 0.0
			for key, p := range a.Probabilities {
				_, known := question.Criteria[key]
				cs.Check("choice probability and key", known && p >= 0 && p <= 1, "invalid distribution")
				sum += p
				highest = max(highest, p)
			}
			_, known := question.Criteria[a.Choice]
			cs.Check("choice argmax and probability total", known && a.Probabilities[a.Choice] >= highest && math.Abs(sum-1) <= 1e-6, "invalid choice distribution")
		case ai.ScoreQuestion:
			a, typed := answer.(ai.ScoreAnswer)
			cs.Check("typed score and complete ordered rubric", typed && len(a.Probabilities) == len(question.Criteria) && len(a.Legend) == len(question.Criteria), "wrong score %s", id)
			sum, expected := 0.0, 0.0
			for i, label := range question.Criteria {
				p, exists := a.Probabilities[i]
				cs.Check("score label and probability", exists && p >= 0 && p <= 1 && jsonEquivalent(a.Legend[i], label), "wrong level %d", i)
				sum += p
				expected += float64(i) * p
			}
			cs.Check("Decisions own score precision", math.Abs(sum-1) <= 1e-6 && math.Abs(expected-a.Score) <= 1e-6, "distribution deviation %.9g score deviation %.9g", math.Abs(sum-1), math.Abs(expected-a.Score))
		case ai.BoolQuestion:
			a, typed := answer.(ai.BoolAnswer)
			cs.Check("typed predicate probability", typed && a.Probability >= 0 && a.Probability <= 1, "wrong bool %s", id)
		}
	}
}
func jsonEquivalent(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
