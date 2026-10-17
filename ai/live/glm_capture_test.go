//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// A second, independent ceiling before the real transport. Reservations do
// not count as network traffic; SDK retries cannot spend unreserved attempts.
// Only one call is active, but the SDK invokes transport on a producer goroutine.
type glmHTTPGate struct {
	mu   sync.Mutex
	next http.RoundTripper
}

func (g *glmHTTPGate) RoundTrip(req *http.Request) (*http.Response, error) {
	g.mu.Lock()
	b := &glmReportState.Budget
	if glmAttemptCount() >= b.MaxAttempts || glmAttemptCount() >= b.AttemptsReserved {
		g.mu.Unlock()
		return nil, fmt.Errorf("GLM actual HTTP attempt budget exhausted")
	}
	if glmReportState.Mode == "replay" {
		b.ReplayAttempts++
	} else {
		b.HTTPAttempts++
	}
	g.mu.Unlock()
	return g.next.RoundTrip(req)
}
func glmUsagePresence(body string) []json.RawMessage {
	var usages []json.RawMessage
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var obj map[string]json.RawMessage
		if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &obj) != nil {
			continue
		}
		if u, ok := obj["usage"]; ok {
			usages = append(usages, u)
		}
		for _, field := range []string{"response", "message"} {
			var nested map[string]json.RawMessage
			if json.Unmarshal(obj[field], &nested) == nil {
				if u, ok := nested["usage"]; ok {
					usages = append(usages, u)
				}
			}
		}
	}
	return usages
}
func glmRequestIDSources(res ai.Result, e exchange) []string {
	var sources []string
	for _, a := range res.Metadata.Attempts {
		source := "unreported"
		if a.ProviderRequestID != "" {
			source = "unmatched actual provider request ID"
			for k, v := range e.ResponseHeaders {
				if v == a.ProviderRequestID {
					source = "response header:" + k
					break
				}
			}
		}
		sources = append(sources, source)
	}
	return sources
}

func glmSafeFailure(err error) string {
	var e *ai.Error
	if errors.As(err, &e) {
		return string(e.Code) + "/" + string(e.Phase)
	}
	if err != nil {
		return "unclassified failure"
	}
	return "none"
}
func glmModelReports() map[string]any {
	reports := map[string]any{}
	r := glmReportState
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		var cases []glmOutcome
		calls, attempts, replayAttempts := 0, 0, 0
		var knownInput, knownOutput, knownCacheRead, knownCacheWrite int64
		knownEstimate := 0.0
		unreported := 0
		for _, sc := range r.Scenarios {
			if sc.Model == model {
				cases = append(cases, sc)
				calls += sc.Calls
				attempts += sc.HTTPAttempts
				replayAttempts += sc.ReplayAttempts
				knownInput += sc.Usage.Input
				knownOutput += sc.Usage.Output
				knownCacheRead += sc.Usage.CacheRead
				knownCacheWrite += sc.Usage.CacheWrite
				knownEstimate += sc.Usage.Cost.Total
				for _, reporting := range sc.UsageReporting {
					if reporting != ai.UsageComplete {
						unreported++
					}
				}
			}
		}
		kind := "glm-model-candidate-only"
		if r.Mode == "replay" {
			kind = "glm-model-capture-replay-only"
		}
		detail := map[string]any{"schema": 1, "kind": kind, "route": r.Route, "provider": "zhipu", "api": r.API, "endpoint": r.Endpoint, "model": model, "mode": r.Mode, "outcome": r.Models[model], "complete": r.Complete && r.Models[model] == "PASS", "auditPassed": r.AuditPassed, "catalogVersion": r.CatalogVersion, "catalogHash": r.CatalogHash, "sourceManifestHash": r.SourceManifestHash, "gitCommit": r.GitCommit, "startedAt": r.StartedAt, "finishedAt": r.FinishedAt, "sdk": r.SDK, "accountAlias": r.AccountAlias, "regionAlias": r.RegionAlias, "unknown": r.Unknown, "reportedKnownUsage": map[string]any{"input": knownInput, "output": knownOutput, "cacheRead": knownCacheRead, "cacheWrite": knownCacheWrite, "estimatedUSD": knownEstimate, "incompleteAttempts": unreported, "note": "reported known counters only; incomplete or cancelled usage remains unknown, estimate not invoice"}, "replayAttempts": replayAttempts, "calls": calls, "httpAttempts": attempts, "scenarios": cases}
		name := "candidate-" + model + "-report"
		if r.Mode == "replay" {
			name = "replay-" + model + "-report"
		}
		reports[name] = detail
	}
	return reports
}

func setGLMReplay(cs *evidence.Case) {
	if glmReportState != nil && glmReportState.Mode == "replay" {
		if name := os.Getenv("BARNESS_GLM_REPLAY_TESTCASE"); name != "" {
			cs.ReplayEnv("BARNESS_GLM_ROUTE= BARNESS_GLM_KEY= BARNESS_GLM_REPLAY_ARCHIVE= BARNESS_AI_LIVE=0")
			cs.ReplayRun("^TestGLMCaptureReplayProcess$/^" + regexp.QuoteMeta(name) + "$")
			return
		}
		cs.ReplayEnv("BARNESS_AI_LIVE=0 BARNESS_GLM_ROUTE=" + glmReportState.Route + " BARNESS_GLM_REPLAY_ARCHIVE=../../" + glmReplayArchiveRelative())
		cs.ReplayRun("^TestGLMCaptureReplay$")
		return
	}
	if name := os.Getenv("BARNESS_GLM_CONTROLLED"); name != "" {
		cs.ReplayEnv("BARNESS_GLM_ROUTE= BARNESS_GLM_KEY= BARNESS_GLM_CONTROLLED= BARNESS_AI_LIVE=0")
		cs.ReplayRun("^TestGLMCandidateProcess$/^" + regexp.QuoteMeta(name) + "$")
		return
	}
	cs.ReplayEnv("BARNESS_AI_LIVE=1 BARNESS_GLM_ROUTE=" + glmReportState.Route)
	cs.ReplayRun(glmEntry)
}

func glmAttemptCount() int {
	if glmReportState.Mode == "replay" {
		return glmReportState.Budget.ReplayAttempts
	}
	return glmReportState.Budget.HTTPAttempts
}
func glmSelectedEntry() string {
	if glmReportState.Mode == "replay" {
		return "^TestGLMCaptureReplay$"
	}
	return glmEntry
}
