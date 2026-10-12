package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const decisionsScore = `{"type":"score","name":"urgency","score":1.1,"confidence":0.6,"probabilities":[{"value":2,"label":"高","probability":0.3},{"value":0,"label":"低","probability":0.2},{"value":1,"label":"低","probability":0.5}]}`
const decisionsPredicate = `{"type":"predicate","name":"approved","probability":0.37}`
const decisionsMixedWire = `{"model":"host-decisions","input":"track order","questions":[{"type":"predicate","name":"approved","instructions":"Approved?\n\nTrue means: 同意\nFalse means: 不同意"},{"type":"choice","name":"intent","instructions":"Classify","choices":[{"value":"other"},{"value":"track","description":"Tracking"}]},{"type":"score","name":"urgency","instructions":"Urgency","levels":[{"label":"低"},{"label":"低"},{"label":"高"}]}]}`
const decisionsMixedAnswers = `{"approved":{"type":"bool","probability":0.37},"intent":{"type":"choice","choice":"track","confidence":0.7,"probabilities":{"track":0.9,"other":0.1}},"urgency":{"type":"score","score":1.1,"confidence":0.6,"probabilities":{"0":0.2,"1":0.5,"2":0.3},"legend":{"0":"低","1":"低","2":"高"}}}`

func decisionsMixedInput() ai.ClassifierRequest {
	req := decisionsInput()
	req.Questions["urgency"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"Urgency"`), Criteria: []json.RawMessage{json.RawMessage(`"低"`), json.RawMessage(`"低"`), json.RawMessage(`"高"`)}}
	req.Questions["approved"] = ai.BoolQuestion{Instructions: json.RawMessage(`"Approved?"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`"同意"`), False: json.RawMessage(`"不同意"`)}}
	return req
}

func newDecisionsMixedWorld(t *testing.T, configure func(*ai.Config)) *unaryWorld {
	return newDecisionsWorld(t, func(c *ai.Config) {
		m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
		m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}
		m.Capabilities.MinScoreLevels, m.Capabilities.MaxScoreLevels = 2, 10
		if configure != nil {
			configure(c)
		}
	})
}

func decisionsPublicCall(t *testing.T, u *unaryWorld, entry, id string, req ai.ClassifierRequest, hooks ai.Hooks) (ai.ClassifierResult, error) {
	t.Helper()
	target := ai.Target{BindingID: "classifier", ModelID: "host-decisions"}
	if entry == "plain" {
		return u.client.Classify(ctxFor(t), textScope(id), target, req, nil)
	}
	return u.client.WithHooks(hooks).Classify(ctxFor(t), textScope(id), target, req, nil)
}

// Mixed fixtures intentionally retain their own three-kind catalog; the issue
// 02 choice-only fixture independently verifies a model's narrower capability.
func decisionsNoopHooks() ai.Hooks {
	return ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
		return ai.KeepPayload(), nil
	}}
}
