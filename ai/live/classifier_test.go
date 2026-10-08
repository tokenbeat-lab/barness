//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func classifierBudget() supportmatrix.Budget {
	return supportmatrix.Budget{MaxCalls: 6, MaxRetries: 1, MaxQuestions: 10, MaxStateBytes: 131072}
}

// Called under mu: reserve the entire public request before any network work.
func classifierBudgetAllows(questions, stateBytes int) bool {
	return questions >= 0 && stateBytes >= 0 && budget.CallsUsed < budget.MaxCalls &&
		questions <= budget.MaxQuestions-budget.QuestionsUsed && stateBytes <= budget.MaxStateBytes
}
func (s *session) spendQuestions(questions, stateBytes int) bool {
	mu.Lock()
	ok := classifierBudgetAllows(questions, stateBytes)
	if ok {
		budget.CallsUsed++
		budget.QuestionsUsed += questions
		budget.StateBytesSubmitted += stateBytes
	}
	mu.Unlock()
	if !ok {
		s.category = supportmatrix.CategoryBudget
		s.check("classifier budget", false, "call, question or state budget exhausted")
		return false
	}
	s.calls++
	s.questions += questions
	return true
}
func (s *session) classify(label string, req ai.ClassifierRequest) (ai.ClassifierResult, error) {
	if !s.spendQuestions(len(req.Questions), len(req.State)) {
		return ai.ClassifierResult{}, errors.New("classifier smoke budget exhausted")
	}
	scope := newScope()
	res, err := s.env.client.Classify(s.ctx, scope, s.target(), req, ai.TypeSafeOptions{})
	exchanges := s.env.rec.take()
	s.lastExchanges = exchanges
	s.httpAttempts += len(exchanges)
	mu.Lock()
	budget.HTTPAttempts += len(exchanges)
	for _, a := range res.Metadata.Attempts {
		budget.InputTokensUsed += int(a.Usage.Input)
		budget.OutputTokensUsed += int(a.Usage.Output)
	}
	mu.Unlock()
	for _, a := range res.Metadata.Attempts {
		if a.ProviderRequestID != "" {
			s.requestIDs = append(s.requestIDs, a.ProviderRequestID)
		}
	}
	name := fmt.Sprintf("attempt-%d-%s", s.attempt, label)
	s.cs.Record(name+"-result", map[string]any{"result": res, "error": errText(err)})
	s.cs.Record(name+"-exchanges", exchanges)
	observations := s.env.rec.observationsOf(scope.RequestID)
	s.cs.Record("observations-"+name, observations)
	s.check("Observer reaches terminal metadata", len(observations) > 0 && observations[len(observations)-1].Kind == ai.ObservationCallFinished, "missing terminal observation")
	if len(exchanges) > 0 {
		s.check("wire uses fixed official target", len(exchanges) == 1 && exchanges[0].URL == s.env.combo.endpoint+"/systemone" && !exchanges[0].Truncated, "unexpected or truncated wire capture")
	}
	return res, err
}
func singleChoiceRequest() ai.ClassifierRequest {
	return ai.ClassifierRequest{State: json.RawMessage(`{"ticket":"My payment failed. Please fix my invoice."}`), Questions: map[string]ai.ClassifierQuestion{
		"intent": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Which department handles this ticket?"`), Criteria: map[string]json.RawMessage{"billing": json.RawMessage(`"Payments and invoices"`), "technical": json.RawMessage(`"Software bugs"`)}},
	}}
}
func mixedQuestionsRequest() ai.ClassifierRequest {
	req := singleChoiceRequest()
	req.Questions["severity"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate the urgency"`), Criteria: []json.RawMessage{json.RawMessage(`"Low"`), json.RawMessage(`"Medium"`), json.RawMessage(`"High"`)}}
	req.Questions["urgent"] = ai.BoolQuestion{Instructions: json.RawMessage(`"Is an immediate response needed?"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`"Immediate"`), False: json.RawMessage(`"Can wait"`)}}
	return req
}
func classifierScenarios() []scenario {
	return []scenario{
		{id: "mixed-questions", run: func(s *session) {
			req := mixedQuestionsRequest()
			res, err := s.classify("mixed", req)
			if s.ok("mixed", err) {
				s.classifierTurn(res, req)
			}
		}},
		{id: "single-choice", run: func(s *session) {
			req := singleChoiceRequest()
			res, err := s.classify("choice", req)
			if s.ok("choice", err) {
				s.classifierTurn(res, req)
			}
		}},
		{id: "context-422", run: func(s *session) {
			req := singleChoiceRequest()
			// 40,000 separate synthetic words fit the host's byte policy while
			// exceeding the documented 32k state-plus-longest-question token budget.
			// No token estimator or bypass of the public Client is introduced.
			req.State, _ = json.Marshal(strings.Repeat("x ", 40000))
			res, err := s.classify("context-422", req)
			if isContext422(res, err) {
				s.note("confirmed official HTTP 422 via Client; %d state bytes, 1 question; usage %s", len(req.State), res.Metadata.Attempts[0].UsageReporting)
				return
			}
			if context400(res, err, s.lastExchanges) {
				s.note("422 shape UNCONFIRMED; bounded public Client probe observed HTTP 400 detail.error_type=max_tokens_exceeded, classified invalid_request/request; %d state bytes, 1 question; usage %s", len(req.State), res.Metadata.Attempts[0].UsageReporting)
				return
			}
			if err != nil && !s.ok("context-422", err) {
				return
			}
			s.check("official 422 shape confirmed", false, "not confirmed within bounded probe; no further input expansion")
		}},
	}
}
func isContext422(res ai.ClassifierResult, err error) bool {
	var e *ai.Error
	return errors.As(err, &e) && e.Code == ai.CodeInvalidRequest && e.Phase == ai.PhaseRequest && e.HTTPStatus == 422 &&
		len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].HTTPStatus == 422 && res.Metadata.Attempts[0].ProviderRequestID != "" && len(res.Answers) == 0
}
func (s *session) classifierTurn(res ai.ClassifierResult, req ai.ClassifierRequest) {
	model, ok := res.ResponseModel.Get()
	s.check("fixed response and authorized identity", ok && model == "jev-1.13.0" && res.Metadata.ModelID == s.model && res.Metadata.Operation == ai.OperationClassifier &&
		res.Metadata.ProviderID == ai.ProviderTypeSafe && res.Metadata.API == ai.APITypeSafeSystemOne && res.StopReason == ai.StopReasonStop, "wrong identity or terminal state")
	s.check("complete questions", len(res.Answers) == len(req.Questions), "answers=%d questions=%d", len(res.Answers), len(req.Questions))
	if !s.check("one official attempt and vendor ID", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].HTTPStatus == 200 && res.Metadata.Attempts[0].ProviderRequestID != "", "missing successful attempt or ID") {
		return
	}
	reporting := res.Metadata.Attempts[0].UsageReporting
	s.note("response model %s; usage %s (absence means unknown consumption)", model, reporting)
	s.check("free output and input pricing", res.Usage.Cost.Output == 0 && math.Abs(res.Usage.Cost.Input-float64(res.Usage.Input)*0.042/1e6) < 1e-12, "invalid price")
	for id, answer := range res.Answers {
		switch a := answer.(type) {
		case ai.ChoiceAnswer:
			sum := 0.0
			for _, p := range a.Probabilities {
				sum += p
			}
			s.check("choice distribution tolerance", math.Abs(sum-1) <= 1e-6, "invalid probability sum")
			s.note("%s choice distribution deviation %.9g", id, math.Abs(sum-1))
		case ai.ScoreAnswer:
			if len(a.Probabilities) > 0 {
				sum, expected := 0.0, 0.0
				for i, p := range a.Probabilities {
					sum += p
					expected += float64(i) * p
				}
				s.check("score distribution tolerance", math.Abs(sum-1) <= 1e-6, "invalid distribution")
				s.note("%s score distribution deviation %.9g / expectation %.9g (Client validates bounded hundredth rounding)", id, math.Abs(sum-1), math.Abs(expected-a.Score))
			}
		case ai.BoolAnswer:
			s.check("typed bool probability", a.Probability >= 0 && a.Probability <= 1, "invalid bool probability")
		}
	}
}
func readOfficial(t *testing.T, v any) {
	t.Helper()
	data, err := os.ReadFile("../e2e/testdata/typesafe/official-capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

// The actual official context guard reports 400, while the API documents
// 422 validation errors. Accept only that concrete context shape and keep
// the 422 conclusion explicitly unconfirmed; unrelated 400s still fail.
func context400(res ai.ClassifierResult, err error, ex []exchange) bool {
	var e *ai.Error
	if !errors.As(err, &e) || e.Code != ai.CodeInvalidRequest || e.Phase != ai.PhaseRequest || e.HTTPStatus != 400 ||
		len(res.Metadata.Attempts) != 1 || res.Metadata.Attempts[0].HTTPStatus != 400 || res.Metadata.Attempts[0].ProviderRequestID == "" || len(res.Answers) != 0 || len(ex) != 1 || ex[0].Truncated {
		return false
	}
	var body struct {
		Detail struct {
			ErrorType string `json:"error_type"`
		} `json:"detail"`
	}
	return json.Unmarshal([]byte(ex[0].ResponseBody), &body) == nil && body.Detail.ErrorType == "max_tokens_exceeded"
}
