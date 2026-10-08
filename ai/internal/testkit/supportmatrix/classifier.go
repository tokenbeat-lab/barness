package supportmatrix

import (
	"slices"
	"strings"
)

// ClassifierExpected is the complete required smoke for the first fixed
// classifier release. A report cannot invent a smaller capability set.
func ClassifierExpected() []string {
	return []string{"mixed-questions", "single-choice", "context-422"}
}

func checkClassifier(r Report) error {
	if r.Combo != "typesafe-classifier" {
		return nil
	}
	if r.Operation != "classifier" || r.Provider != "typesafe" || r.API != "typesafe-system-one" ||
		r.Model != "jev-1.13.0" || r.Endpoint != "https://api.typesafe.ai/v1" {
		return refuse("classifier report does not name the fixed official route and model")
	}
	expected := ClassifierExpected()
	if len(r.Expected) != len(expected) {
		return refuse("classifier report lacks its complete capability set")
	}
	for _, id := range expected {
		if !slices.Contains(r.Expected, id) {
			return refuse("classifier report lacks %q", id)
		}
	}
	b := r.Budget
	if b.MaxCalls != 6 || b.MaxRetries != 1 || b.MaxQuestions != 10 || b.MaxStateBytes != 131072 ||
		b.CallsUsed < 0 || b.CallsUsed > b.MaxCalls || b.QuestionsUsed < 0 || b.QuestionsUsed > b.MaxQuestions ||
		b.HTTPAttempts < 0 || b.HTTPAttempts > b.CallsUsed || b.StateBytesSubmitted < 0 || (b.HTTPAttempts > 0 && b.StateBytesSubmitted == 0) || b.StateBytesSubmitted > b.CallsUsed*b.MaxStateBytes ||
		b.EnvironmentRetries < 0 || b.EnvironmentRetries > len(expected) {
		return refuse("classifier report has an invalid or exceeded budget")
	}
	calls, questions, attempts, retries := 0, 0, 0, 0
	for _, s := range r.Scenarios {
		if s.Model != r.Model || s.Outcome == Unsupported {
			return refuse("classifier capability %q lacks its required fixed model", s.ID)
		}
		if s.Calls < 0 || s.Attempts < 0 || s.Questions < 0 || s.Retries < 0 || s.Retries > b.MaxRetries || s.Attempts > s.Calls ||
			(s.Outcome == Pass && (s.Calls == 0 || s.Attempts == 0 || s.Questions == 0 || len(s.ProviderRequestIDs) == 0 || slices.ContainsFunc(s.ProviderRequestIDs, func(id string) bool { return strings.TrimSpace(id) == "" }))) {
			return refuse("classifier capability %q lacks consistent actual call evidence", s.ID)
		}
		calls += s.Calls
		questions += s.Questions
		attempts += s.Attempts
		retries += s.Retries
	}
	if calls != b.CallsUsed || questions != b.QuestionsUsed || attempts != b.HTTPAttempts || retries != b.EnvironmentRetries {
		return refuse("classifier actual counters disagree with its scenarios")
	}
	return nil
}
