package supportmatrix

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

type xaiExchange struct {
	Method              string            `json:"method"`
	URL                 string            `json:"url"`
	RequestHeaders      map[string]string `json:"requestHeaders"`
	RequestBody         string            `json:"requestBody,omitempty"`
	Status              int               `json:"status,omitempty"`
	ResponseHeaderNames []string          `json:"responseHeaderNames,omitempty"`
	ResponseHeaders     map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody        string            `json:"responseBody,omitempty"`
	Truncated           bool              `json:"truncated,omitempty"`
	TransportError      string            `json:"transportError,omitempty"`
	ElapsedMS           int64             `json:"elapsedMs"`
}
type xaiCall struct {
	ex                          xaiExchange
	result                      regionalResult
	message                     ai.AssistantMessage
	request                     map[string]json.RawMessage
	terminal                    map[string]json.RawMessage
	output                      []map[string]json.RawMessage
	input, outputTokens, tokens int64
	usd                         float64
}

func verifyXAICases(root *os.Root, r xaiEvidence, inventory map[string]string) (xaiTotals, error) {
	var total xaiTotals
	seen := map[string]bool{}
	for _, scene := range r.Scenarios {
		var calls []xaiCall
		labels := xaiLabels(scene.ID)
		allowed := map[string]bool{"assertions.json": true}
		if scene.ID == "encrypted-continuation" {
			allowed["persisted-turn.json"] = true
		}
		for i, label := range labels {
			for _, suffix := range []string{"http", "result", "observer"} {
				allowed[label+"-"+suffix+".json"] = true
			}
			stream := label == "calls" || label == "continued" || scene.ID == "entry-stream" || scene.ID == "entry-stream-simple"
			if stream {
				allowed[label+"-events.json"] = true
			}
			call, err := readXAICall(root, r, scene, label, stream, seen)
			if err != nil {
				return total, err
			}
			calls = append(calls, call)
			if len(scene.ProviderRequestIDs) != len(labels) || scene.ProviderRequestIDs[i] != call.result.Metadata.Attempts[0].ProviderRequestID {
				return total, refuse("xAI scenario request IDs differ from actual attempts")
			}
			if call.input > math.MaxInt64-total.input || call.outputTokens > math.MaxInt64-total.output || call.tokens > math.MaxInt64-total.tokens {
				return total, refuse("xAI aggregate token arithmetic overflow")
			}
			total.calls++
			total.input += call.input
			total.output += call.outputTokens
			total.tokens += call.tokens
			total.usd += call.usd
		}
		for name := range inventory {
			prefix := scene.CaseID + "/"
			if strings.HasPrefix(name, prefix) && !allowed[strings.TrimPrefix(name, prefix)] {
				return total, refuse("xAI case contains an undeclared attempt or artifact")
			}
		}
		for name := range allowed {
			if inventory[scene.CaseID+"/"+name] == "" {
				return total, refuse("xAI required triplet or stream proof missing")
			}
		}
		if err := verifyXAIScenario(root, scene, calls); err != nil {
			return total, err
		}
	}
	return total, nil
}
func readXAICall(root *os.Root, r xaiEvidence, s ScenarioResult, label string, stream bool, seen map[string]bool) (xaiCall, error) {
	var c xaiCall
	prefix := s.CaseID + "/" + label
	var ex []xaiExchange
	if xaiReadJSON(root, prefix+"-http.json", &ex) != nil || len(ex) != 1 {
		return c, refuse("xAI exactly one complete own HTTP capture required")
	}
	c.ex = ex[0]
	if c.ex.Method != "POST" || c.ex.URL != xaiEndpoint+"/responses" || c.ex.Status != 200 || c.ex.Truncated || c.ex.TransportError != "" || len(c.ex.ResponseBody) == 0 || len(c.ex.ResponseBody) > 512<<10 || len(c.ex.RequestBody) == 0 || len(c.ex.RequestBody) > 128<<10 || c.ex.ElapsedMS < 0 || c.ex.ElapsedMS > 65000 || c.ex.RequestHeaders["authorization"] != "[REDACTED]" || c.ex.RequestHeaders["x-api-key"] != "" {
		return c, refuse("xAI endpoint, authentication or complete response differs")
	}
	if err := regionalDecode([]byte(c.ex.RequestBody), &c.request); err != nil {
		return c, err
	}
	if err := verifyXAIRequest(c.request); err != nil {
		return c, err
	}
	response, output, err := xaiTerminal(c.ex.ResponseBody)
	if err != nil {
		return c, err
	}
	c.terminal, c.output = response, output
	if xaiReadJSON(root, prefix+"-result.json", &c.result) != nil {
		return c, refuse("xAI actual Result missing")
	}
	c.message, err = c.result.Message.assistant()
	if err != nil {
		return c, err
	}
	m := c.result.Metadata
	if !m.Resolved || m.Operation != ai.OperationChat || m.ProviderID != ai.ProviderXAI || m.API != ai.APIOpenAIResponses || m.ModelID != xaiModel || m.BindingID != xaiCombo || m.AccountScopeID != "live-"+xaiCombo || m.TenantID != "live-smoke" || m.BindingVersion != "live" || m.CredentialVersion != "live" || m.CatalogVersion != r.CatalogVersion || m.CatalogHash != r.CatalogHash || m.RequestID == "" || seen[m.RequestID] || len(m.Attempts) != 1 || m.NativeStateDowngrades != (ai.NativeStateDowngrades{}) || c.message.Provider != ai.ProviderXAI || c.message.API != ai.APIOpenAIResponses || c.message.Model != xaiModel {
		return c, refuse("xAI Result attribution, attempt or native source differs")
	}
	seen[m.RequestID] = true
	a := m.Attempts[0]
	header := ""
	matches := 0
	for key, v := range c.ex.ResponseHeaders {
		if strings.EqualFold(key, "x-request-id") {
			header = v
			matches++
		}
	}
	if matches > 1 || a.AttemptID != m.RequestID+"#1" || a.HTTPStatus != 200 || a.ProviderRequestID != header || a.Code != "" || a.RetryDelay != 0 || a.UsageReporting != ai.UsageComplete || !reflect.DeepEqual(a.Usage, c.message.Usage) {
		return c, refuse("xAI actual Attempt, optional header ID or usage differs")
	}
	if err := verifyXAIUsage(&c); err != nil {
		return c, err
	}
	if err := verifyXAIOutput(c); err != nil {
		return c, err
	}
	var obs []ai.Observation
	if xaiReadJSON(root, prefix+"-observer.json", &obs) != nil || len(obs) != 4 {
		return c, refuse("xAI Observer lifecycle missing")
	}
	scope := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: m.TenantID, RequestID: m.RequestID, ActorID: m.ActorID, JobID: m.JobID, BindingID: m.BindingID, Operation: m.Operation}}
	resolved := m
	resolved.Attempts = nil
	startAttempt := ai.Attempt{AttemptID: a.AttemptID, UsageReporting: ai.UsageUnreported, CostEstimate: ai.CostEstimate{Status: ai.CostUnreported}}
	if obs[0].Kind != ai.ObservationCallStarted || !reflect.DeepEqual(obs[0].Call, scope) || obs[1].Kind != ai.ObservationAttemptStarted || !reflect.DeepEqual(obs[1].Call, resolved) || obs[1].Attempt == nil || !reflect.DeepEqual(*obs[1].Attempt, startAttempt) || obs[2].Kind != ai.ObservationAttemptFinished || !reflect.DeepEqual(obs[2].Call, resolved) || obs[2].Attempt == nil || !reflect.DeepEqual(*obs[2].Attempt, a) || obs[3].Kind != ai.ObservationCallFinished || !reflect.DeepEqual(obs[3].Call, m) || !reflect.DeepEqual(obs[3].Usage, c.message.Usage) || obs[3].StopReason != c.message.StopReason {
		return c, refuse("xAI terminal Observer or attempt lifecycle differs")
	}
	for i, o := range obs {
		if o.Error != nil || o.Duration < 0 || o.Duration > 65*time.Second || o.Time.Before(r.StartedAt) || o.Time.After(r.FinishedAt) || i > 0 && o.Time.Before(obs[i-1].Time) {
			return c, refuse("xAI observation error or timestamps differ")
		}
	}
	if stream {
		if err := verifyXAIEvents(root, prefix+"-events.json", c); err != nil {
			return c, err
		}
	}
	return c, nil
}
