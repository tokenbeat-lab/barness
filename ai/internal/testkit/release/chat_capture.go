package release

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

type chatUsage struct{ input, output int }

// Consume authenticated native triplets, not just their filenames. A call is
// counted once across the entire bundle and its result/terminal attribution,
// wire target/status and scenario counters must all describe the same request.
func verifyChatCaptures(files verifiedCase, r supportmatrix.Report, s supportmatrix.ScenarioResult, seen map[string]bool) (chatUsage, error) {
	var usage chatUsage
	results := files.matches("attempt-*-result.json")
	if len(results) != s.Calls || len(files.matches("attempt-*-exchanges.json")) != s.Calls || len(files.matches("observations-attempt-*.json")) != s.Calls {
		return usage, fmt.Errorf("chat call inventory disagrees with scenario")
	}
	attempts := 0
	for _, name := range results {
		prefix := strings.TrimSuffix(name, "-result.json")
		// Captures are persisted JSON, not domain values. AssistantContent is a
		// closed interface without JSON decoding; read only the fields this
		// consumer owns instead of trying to reconstruct polymorphic content.
		var result struct {
			Result struct {
				Message struct {
					API          ai.API        `json:"api"`
					Provider     ai.ProviderID `json:"provider"`
					Model        string        `json:"model"`
					Usage        ai.Usage      `json:"usage"`
					StopReason   ai.StopReason `json:"stopReason"`
					ErrorMessage string        `json:"errorMessage"`
				} `json:"message"`
				Metadata ai.CallMetadata `json:"metadata"`
			} `json:"result"`
			Error string `json:"error"`
		}
		var exchanges []struct {
			Method, URL, RequestBody string
			Status                   int
			Truncated                bool
		}
		var observations []ai.Observation
		if files.json(name, maxInventoryFile, &result) != nil ||
			files.json(prefix+"-exchanges.json", maxInventoryFile, &exchanges) != nil ||
			files.json("observations-"+prefix+".json", maxInventoryFile, &observations) != nil {
			return usage, fmt.Errorf("chat capture triplet missing or invalid")
		}
		meta := result.Result.Metadata
		if s.ID == grokhistory.Scenario {
			message := result.Result.Message
			if result.Error != "" || message.ErrorMessage != "" || message.StopReason != ai.StopReasonStop || message.Provider != meta.ProviderID || message.API != meta.API || message.Model != s.Model || meta.AccountScopeID != "live-"+r.Combo || len(exchanges) != 1 || exchanges[0].Status != 200 {
				return usage, fmt.Errorf("mandatory xAI history lacks actual successful target Result")
			}
			before, e1 := files.read("public-request.json", maxInventoryFile)
			after, e2 := files.read("immutable-input.json", maxInventoryFile)
			proof, e3 := files.read("history-provenance.json", 4096)
			if e1 != nil || e2 != nil || e3 != nil || grokhistory.VerifyProvenance(proof, meta.TenantID) != nil || !grokhistory.VerifyHost(before) || !grokhistory.VerifyHost(after) || len(exchanges) != 1 || grokhistory.Verify(ai.API(r.API), []byte(exchanges[0].RequestBody)) != nil || meta.NativeStateDowngrades != (ai.NativeStateDowngrades{CrossModel: 2}) {
				return usage, fmt.Errorf("fixed xAI Host input, actual wire or downgrades differ")
			}
		}
		if meta.RequestID == "" || meta.TenantID == "" || seen[meta.RequestID] || !meta.Resolved || meta.BindingID != r.Combo ||
			string(meta.Operation) != r.Operation || string(meta.ProviderID) != r.Provider || string(meta.API) != r.API || meta.ModelID != s.Model ||
			len(meta.Attempts) == 0 || len(meta.Attempts) != len(exchanges) || len(observations) == 0 {
			return usage, fmt.Errorf("chat result attribution or actual attempts disagree")
		}
		if s.ID != "pro-model-unavailable" && (meta.CatalogHash != r.CatalogHash || meta.CatalogVersion != r.CatalogVersion) {
			return usage, fmt.Errorf("chat result catalog disagrees")
		}
		terminal := observations[len(observations)-1]
		if terminal.Kind != ai.ObservationCallFinished || !reflect.DeepEqual(terminal.Call, meta) {
			return usage, fmt.Errorf("chat terminal Observer disagrees with consumed result")
		}
		seen[meta.RequestID] = true
		for i, exchange := range exchanges {
			u, err := url.Parse(exchange.URL)
			target, _ := url.Parse(r.Endpoint)
			if err != nil || u.Scheme != target.Scheme || u.Host != target.Host || u.User != nil ||
				!strings.HasPrefix(u.Path, strings.TrimSuffix(target.Path, "/")+"/") ||
				exchange.Method != "POST" || exchange.Truncated || !json.Valid([]byte(exchange.RequestBody)) ||
				meta.Attempts[i].AttemptID != fmt.Sprintf("%s#%d", meta.RequestID, i+1) || meta.Attempts[i].HTTPStatus != exchange.Status {
				return usage, fmt.Errorf("chat wire target, body or attempt disagrees")
			}
			if r.API == "google-generative-ai" {
				if !strings.HasPrefix(u.Path, target.Path+"/models/"+s.Model+":") {
					return usage, fmt.Errorf("chat wire model disagrees")
				}
			} else {
				var body struct{ Model string }
				if json.Unmarshal([]byte(exchange.RequestBody), &body) != nil || body.Model != s.Model {
					return usage, fmt.Errorf("chat wire model disagrees")
				}
			}
		}
		attempts += len(exchanges)
		u := result.Result.Message.Usage
		if u.Input < 0 || u.Output < 0 || u.CacheRead < 0 || u.CacheWrite < 0 {
			return usage, fmt.Errorf("chat usage negative")
		}
		usage.input += int(u.Input + u.CacheRead + u.CacheWrite)
		usage.output += int(u.Output)
	}
	if attempts != s.Attempts {
		return usage, fmt.Errorf("chat captured attempts disagree with scenario")
	}
	return usage, nil
}
