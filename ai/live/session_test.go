//go:build live

package live

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// scenario is one smoke check of a combination.
type scenario struct {
	id string
	// model replaces the combination's model for this scenario.
	model string
	run   func(s *session)
}

func scenariosOf(c *combo) []scenario {
	if c.api == ai.APIGoogleInteractions {
		return googleImagesScenarios()
	}
	if c.operation == ai.OperationImage {
		return imagesScenarios()
	}
	if c.operation == ai.OperationClassifier {
		return classifierScenarios()
	}
	return append([]scenario{textStream, textComplete, toolRoundTrip, cancelAfterFirstFrame, reasoningHistory, imageInput},
		c.extra...)
}

// retryableCodes are vendor or environment faults: retried within the
// budget to confirm, then FAIL with Environment set (spec §5).
var retryableCodes = []ai.Code{ai.CodeRateLimited, ai.CodeUpstreamError, ai.CodeTransport, ai.CodeDeadlineExceeded}

// session is one attempt at a scenario.
type session struct {
	t     *testing.T
	ctx   context.Context
	cs    *evidence.Case
	env   *liveEnv
	model string
	// fault is a vendor or environment failure that ended the attempt
	// before any assertion; the harness retries it.
	fault *ai.Error
	// category classifies a failed attempt (supportmatrix.ScenarioResult).
	category      string
	unsupported   string
	notes         []string
	requestIDs    []string
	calls         int
	images        int
	imageCalls    []supportmatrix.ImageCall
	questions     int
	httpAttempts  int
	lastExchanges []exchange
	attempt       int
}

func runScenario(t *testing.T, cs *evidence.Case, c *combo, selected bool, env *liveEnv, sc scenario) {
	model := cmp.Or(sc.model, c.model)
	res := supportmatrix.ScenarioResult{ID: sc.id, CaseID: cs.ID(), Model: model}
	switch {
	case !selected:
		t.Skipf("NOT_RUN: this process runs %s", envCombo+"="+nameOf(cfg.combo))
	case cfg.notRun != "":
		res.Outcome, res.Note = supportmatrix.NotRun, cfg.notRun
		record(res)
		t.Skip("NOT_RUN: " + cfg.notRun)
	case cfg.refused != "":
		cs.Check("process configuration", false, "%s", cfg.refused)
		res.Outcome, res.ErrorCategory = supportmatrix.Fail, supportmatrix.CategoryConfig
		record(res)
		return
	}
	if c.api == ai.APIOpenAIImages && sc.id != "json-edit" && !jsonEditPassed() {
		res.Outcome, res.Note = supportmatrix.NotRun, "required JSON edit did not pass in this process"
		record(res)
		t.Skip("NOT_RUN: " + res.Note)
	}
	began := time.Now()
	var s *session
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), scenarioTimeLimit)
		s = &session{t: t, ctx: ctx, cs: cs, env: env, model: model, attempt: attempt}
		sc.run(s)
		cancel()
		res.Calls += s.calls
		res.Images += s.images
		res.ImageCalls = append(res.ImageCalls, s.imageCalls...)
		res.Questions += s.questions
		res.Attempts += s.httpAttempts
		res.ProviderRequestIDs = append(res.ProviderRequestIDs, s.requestIDs...)
		// A fault after a failed assertion is not retried: the assertion
		// already decided the scenario.
		if s.fault == nil || t.Failed() || attempt > maxRetries || !slices.Contains(retryableCodes, s.fault.Code) || budgetLeft() == 0 {
			break
		}
		res.Retries++
		addRetry()
		wait := retryBackoff * time.Duration(attempt)
		t.Logf("attempt %d met a vendor or environment fault (%s); retrying in %s", attempt, s.fault.Code, wait)
		time.Sleep(wait)
	}
	res.DurationMS = time.Since(began).Milliseconds()
	res.Note = strings.Join(s.notes, "; ")
	if s.fault != nil {
		cs.Check("vendor or environment fault", false, "%s (%s, HTTP %d) after %d attempt(s): %s",
			s.fault.Code, s.fault.Phase, s.fault.HTTPStatus, s.attempt, truncate(s.fault.Message, 300))
		s.category = string(s.fault.Code)
	}
	switch {
	case t.Failed():
		res.Outcome, res.ErrorCategory = supportmatrix.Fail, cmp.Or(s.category, supportmatrix.CategoryAssertion)
		res.Environment = s.fault != nil
	case s.unsupported != "":
		cs.Unsupported(s.unsupported)
		res.Outcome, res.Note = supportmatrix.Unsupported, s.unsupported
	default:
		res.Outcome = supportmatrix.Pass
	}
	record(res)
}

func record(r supportmatrix.ScenarioResult) {
	mu.Lock()
	defer mu.Unlock()
	results = append(results, r)
}

func budgetLeft() int {
	mu.Lock()
	defer mu.Unlock()
	return budget.MaxCalls - budget.CallsUsed
}

func addRetry() {
	mu.Lock()
	defer mu.Unlock()
	budget.EnvironmentRetries++
}

func nameOf(c *combo) string {
	if c == nil {
		return "(none)"
	}
	return c.name
}

// spend takes one logical call from the budget; false ends the scenario as
// FAIL with category budget.
func (s *session) spend() bool {
	mu.Lock()
	ok := budget.CallsUsed < budget.MaxCalls
	if ok {
		budget.CallsUsed++
	}
	mu.Unlock()
	if !ok {
		s.category = supportmatrix.CategoryBudget
		s.cs.Check("call budget", false, "the process's %d calls are used up", maxCalls)
		return false
	}
	s.calls++
	return true
}

func (s *session) target() ai.Target { return ai.Target{BindingID: s.env.combo.name, ModelID: s.model} }

func (s *session) note(format string, args ...any) {
	s.notes = append(s.notes, fmt.Sprintf(format, args...))
}

func (s *session) check(name string, ok bool, format string, args ...any) bool {
	s.t.Helper()
	return s.cs.Check(name, ok, format, args...)
}

// observe records a finished call: its result, request ids, usage spent and
// the wire exchanges captured since the previous call.
func (s *session) observe(label string, res ai.Result, err error) []exchange {
	ex := s.env.rec.take()
	for _, a := range res.Metadata.Attempts {
		if a.ProviderRequestID != "" {
			s.requestIDs = append(s.requestIDs, a.ProviderRequestID)
		}
	}
	if len(res.Metadata.Attempts) > 0 && res.Metadata.Attempts[len(res.Metadata.Attempts)-1].ProviderRequestID == "" {
		// Not in the header barness reads; keep whatever id-like header
		// the vendor sent, so the record still names the request.
		for _, e := range ex {
			for name, v := range e.ResponseHeaders {
				if strings.Contains(name, "id") {
					s.requestIDs = append(s.requestIDs, name+"="+v)
				}
			}
		}
	}
	mu.Lock()
	budget.HTTPAttempts += len(ex)
	s.httpAttempts += len(ex)
	budget.InputTokensUsed += int(res.Message.Usage.Input + res.Message.Usage.CacheRead + res.Message.Usage.CacheWrite)
	budget.OutputTokensUsed += int(res.Message.Usage.Output)
	mu.Unlock()
	name := fmt.Sprintf("attempt-%d-%s", s.attempt, label)
	s.cs.Record(name+"-result", map[string]any{"result": res, "error": errText(err)})
	s.cs.Record(name+"-exchanges", ex)
	return ex
}

// ok reports whether a call that should succeed did. A vendor or
// environment fault becomes the attempt's fault, without an assertion, so
// the harness can retry; any other error is an assertion failure.
func (s *session) ok(label string, err error) bool {
	s.t.Helper()
	if err == nil {
		return true
	}
	var e *ai.Error
	if errors.As(err, &e) && s.env.combo.operation == ai.OperationImage && e.HTTPStatus == 400 {
		s.category = string(e.Code)
		s.check(label+": call succeeds", false, "HTTP 400: %s", errText(err))
		return false
	}
	if errors.As(err, &e) && (slices.Contains(retryableCodes, e.Code) || e.Code == ai.CodeUpstreamAuth) {
		s.fault = e
		return false
	}
	if errors.As(err, &e) {
		s.category = string(e.Code)
	}
	s.check(label+": call succeeds", false, "%s", errText(err))
	return false
}

// turn asserts what every successful turn shows (spec §5: structure,
// attribution, terminal state, non-empty content, usage).
func (s *session) turn(label string, res ai.Result, want ai.StopReason) bool {
	s.t.Helper()
	m, meta := res.Message, res.Metadata
	c := s.env.combo
	ok := s.check(label+": stop reason", m.StopReason == want, "got %q, want %q (%s)", m.StopReason, want, m.ErrorMessage)
	ok = s.check(label+": attribution", meta.Resolved && meta.ProviderID == c.provider && meta.API == c.api &&
		meta.ModelID == s.model && m.Provider == c.provider && m.API == c.api && m.Model == s.model,
		"meta %+v, message %s × %s × %s", meta.CallAttribution, m.Provider, m.API, m.Model) && ok
	ok = s.check(label+": one HTTP 200 attempt", len(meta.Attempts) == 1 && meta.Attempts[0].HTTPStatus == 200,
		"attempts %+v", meta.Attempts) && ok
	u := m.Usage
	// Usage must be reported; how completely is the vendor's shape, kept as
	// a note (a missing cache field makes it partial, not absent).
	reporting := reportingOf(meta)
	ok = s.check(label+": usage reported", u.Input+u.CacheRead+u.CacheWrite > 0 && u.Output > 0 &&
		len(reporting) == 1 && reporting[0] != ai.UsageUnreported, "usage %+v, reporting %v", u, reporting) && ok
	if len(reporting) == 1 && reporting[0] != ai.UsageComplete {
		s.note("%s: usage reporting %s", label, reporting[0])
	}
	if want == ai.StopReasonStop {
		ok = s.check(label+": non-empty answer", strings.TrimSpace(textOf(m)) != "", "content %v", m.Content) && ok
	}
	return ok
}

func reportingOf(meta ai.CallMetadata) []ai.UsageReporting {
	var out []ai.UsageReporting
	for _, a := range meta.Attempts {
		out = append(out, a.UsageReporting)
	}
	return out
}

// textOf joins the message's text blocks; like drain, a small copy of an
// ai/e2e helper kept local to this test program.
func textOf(m ai.AssistantMessage) string {
	var b strings.Builder
	for _, c := range m.Content {
		if t, ok := c.(ai.Text); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return truncate(err.Error(), 2000)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// A narrowed run cannot bypass the required first edit. Results are local to
// this process, so another combination or an earlier bundle cannot satisfy it.
func jsonEditPassed() bool {
	mu.Lock()
	defer mu.Unlock()
	return slices.ContainsFunc(results, func(r supportmatrix.ScenarioResult) bool {
		return r.ID == "json-edit" && r.Outcome == supportmatrix.Pass
	})
}
