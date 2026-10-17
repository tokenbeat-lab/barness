//go:build live

package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
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
	if b.HTTPAttempts >= b.MaxAttempts || b.HTTPAttempts >= b.AttemptsReserved {
		g.mu.Unlock()
		return nil, fmt.Errorf("GLM actual HTTP attempt budget exhausted")
	}
	b.HTTPAttempts++
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
func recordGLMModelReports() error {
	r := glmReportState
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		var cases []glmOutcome
		calls, attempts := 0, 0
		for _, sc := range r.Scenarios {
			if sc.Model == model {
				cases = append(cases, sc)
				calls += sc.Calls
				attempts += sc.HTTPAttempts
			}
		}
		detail := map[string]any{"schema": 1, "kind": "glm-model-candidate-only", "route": r.Route, "provider": "zhipu", "api": r.API, "endpoint": r.Endpoint, "model": model, "mode": r.Mode, "outcome": r.Models[model], "complete": r.Complete && r.Models[model] == "PASS", "auditPassed": r.AuditPassed, "catalogVersion": r.CatalogVersion, "catalogHash": r.CatalogHash, "sourceManifestHash": r.SourceManifestHash, "gitCommit": r.GitCommit, "startedAt": r.StartedAt, "finishedAt": r.FinishedAt, "sdk": r.SDK, "accountAlias": r.AccountAlias, "regionAlias": r.RegionAlias, "unknown": r.Unknown, "calls": calls, "httpAttempts": attempts, "scenarios": cases}
		if err := run.Record("candidate-"+model+"-report", detail); err != nil {
			return err
		}
	}
	return nil
}
