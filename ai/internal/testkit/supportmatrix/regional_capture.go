package supportmatrix

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

type regionalExchange struct {
	Method, URL, RequestBody, ResponseBody, TransportError string
	RequestHeaders, ResponseHeaders                        map[string]string
	Status                                                 int
	Truncated                                              bool
}

func verifyRegionalCase(root *os.Root, r Report, s ScenarioResult) error {
	read := func(name string, v any) error {
		raw, e := regionalRead(root, s.CaseID+"/"+name)
		if e != nil {
			return e
		}
		return regionalDecode(raw, v)
	}
	var checks struct {
		ID         string `json:"case_id"`
		Status     string
		Assertions []struct{ Pass bool }
	}
	if read("assertions.json", &checks) != nil || checks.ID != s.CaseID || checks.Status != "PASS" || len(checks.Assertions) < 3 {
		return refuse("regional native assertions missing")
	}
	for _, check := range checks.Assertions {
		if !check.Pass {
			return refuse("regional PASS hides failed assertion")
		}
	}
	var exchanges []regionalExchange
	if read("exchanges.json", &exchanges) != nil || len(exchanges) != 1 {
		return refuse("regional exactly one own HTTP attempt required")
	}
	x := exchanges[0]
	if err := verifyRegionalStream(r, s, x.ResponseBody); err != nil {
		return err
	}
	wantURL := r.Endpoint + "/chat/completions"
	if r.API == "openai-responses" {
		wantURL = r.Endpoint + "/responses"
	}
	if r.API == "anthropic-messages" {
		wantURL = r.Endpoint + "/v1/messages?beta=true"
	}
	if x.Method != "POST" || x.URL != wantURL || x.Status != 200 || x.Truncated || x.TransportError != "" || x.ResponseBody == "" {
		return refuse("regional own endpoint, actual response or complete capture differs")
	}
	if r.API == "anthropic-messages" {
		if x.RequestHeaders["x-api-key"] != "[REDACTED]" || x.RequestHeaders["authorization"] != "" || x.RequestHeaders["anthropic-version"] != "2023-06-01" {
			return refuse("regional Messages authentication/version differs")
		}
	} else if x.RequestHeaders["authorization"] != "[REDACTED]" || x.RequestHeaders["x-api-key"] != "" {
		return refuse("regional Bearer authentication differs")
	}
	var request map[string]json.RawMessage
	if json.Unmarshal([]byte(x.RequestBody), &request) != nil || string(request["model"]) != string(mustRegionalJSON(s.Model)) || string(request["stream"]) != "true" {
		return refuse("regional own wire model differs")
	}
	var result regionalResult
	if read("result.json", &result) != nil {
		return refuse("regional actual Result missing")
	}
	message, err := result.Message.assistant()
	if err != nil {
		return err
	}
	m := result.Metadata
	if !m.Resolved || string(m.ProviderID) != r.Provider || string(m.API) != r.API || m.ModelID != s.Model || m.BindingID != r.Combo || m.AccountScopeID != "regional:"+r.AccountAlias || m.CatalogVersion != r.CatalogVersion || m.CatalogHash != r.CatalogHash || result.Message.Provider != m.ProviderID || result.Message.API != m.API || result.Message.Model != m.ModelID || result.Message.ResponseModel != "" && result.Message.ResponseModel != s.Model || len(m.Attempts) != 1 {
		return refuse("regional actual attribution or model substituted")
	}
	a := m.Attempts[0]
	header := "x-request-id"
	if r.API == "anthropic-messages" && r.Provider == "minimax-cn" {
		header = "request-id"
	}
	if a.HTTPStatus != x.Status || a.ProviderRequestID != x.ResponseHeaders[header] || !reflect.DeepEqual(a.Usage, result.Message.Usage) || a.RetryDelay != 0 {
		return refuse("regional actual attempt, request ID or usage differs")
	}
	_, kind, _ := strings.Cut(s.ID, "/")
	if kind == "cancel" {
		if result.Message.StopReason != ai.StopReasonAborted {
			return refuse("regional cancellation lacks aborted Result")
		}
	} else if result.Message.StopReason != ai.StopReasonStop && !(kind == "tool-call" && result.Message.StopReason == ai.StopReasonToolUse) {
		return refuse("regional own legal terminal missing")
	}
	if kind != "cancel" && a.UsageReporting == ai.UsageUnreported {
		return refuse("regional required usage not reported")
	}
	var observations []ai.Observation
	if read("observations.json", &observations) != nil || len(observations) != 4 || observations[0].Kind != ai.ObservationCallStarted || observations[0].Call.Resolved || observations[3].Kind != ai.ObservationCallFinished || !reflect.DeepEqual(observations[3].Call, m) || !reflect.DeepEqual(observations[3].Usage, result.Message.Usage) {
		return refuse("regional observed lifecycle differs")
	}
	var events []struct{ Type string }
	if read("events.json", &events) != nil {
		return refuse("regional event capture missing")
	}
	if kind != "full-complete" && kind != "simple-complete" {
		terminal := "done"
		if kind == "cancel" {
			terminal = "error"
		}
		if len(events) < 2 || events[0].Type != "start" || events[len(events)-1].Type != terminal {
			return refuse("regional event terminal missing")
		}
	}
	if kind == "tool-result" || kind == "native-history" {
		if err := verifyRegionalToolResult(root, r, s, x.RequestBody, message); err != nil {
			return err
		}
	}
	if kind == "native-history" {
		if m.NativeStateDowngrades != (ai.NativeStateDowngrades{}) {
			return refuse("regional native history was downgraded")
		}
		if err := verifyRegionalNative(root, r, s, x.RequestBody, m); err != nil {
			return err
		}
	}
	if kind == "tool-call" {
		var tools []ai.Tool
		if read("tools.json", &tools) != nil {
			return refuse("regional Host tool definitions missing")
		}
		valid := false
		for i, b := range message.Content {
			if _, ok := b.(ai.ToolCall); ok {
				_, e := ai.ValidateToolCall(message, i, tools)
				valid = e == nil
				break
			}
		}
		if !valid {
			return refuse("regional actual tool cannot be executed by Host")
		}
	}
	var resources map[string]int64
	if read("resources.json", &resources) != nil || len(resources) != 5 {
		return refuse("regional resource proof missing")
	}
	for _, name := range []string{"calls", "events", "permits", "waiters", "bodies"} {
		if n, ok := resources[name]; !ok || n != 0 {
			return refuse("regional resources not released")
		}
	}
	return nil
}

func mustRegionalJSON(v any) []byte { raw, _ := json.Marshal(v); return raw }

func verifyRegionalNative(root *os.Root, r Report, s ScenarioResult, wire string, current ai.CallMetadata) error {
	source := RegionalNativeSourceCase(r.Combo, s.Model)
	raw, err := regionalRead(root, source+"/result.json")
	if err != nil {
		return err
	}
	var first regionalResult
	if json.Unmarshal(raw, &first) != nil || first.Message.Provider != ai.ProviderID(r.Provider) || first.Message.API != ai.API(r.API) || first.Message.Model != s.Model {
		return refuse("regional native source is another service/model")
	}
	m := first.Metadata
	if !m.Resolved || m.Operation != ai.OperationChat || m.Operation != current.Operation || m.TenantID == "" || m.TenantID != current.TenantID || m.AccountScopeID != current.AccountScopeID || m.BindingID != r.Combo || string(m.ProviderID) != r.Provider || string(m.API) != r.API || m.ModelID != s.Model || m.CatalogVersion != r.CatalogVersion || m.CatalogHash != r.CatalogHash || m.RequestID == "" || m.RequestID == current.RequestID {
		return refuse("regional native source scope differs")
	}
	if err := verifyRegionalNativeRequest(root, source, s.CaseID, raw); err != nil {
		return err
	}

	firstMessage, err := first.Message.assistant()
	if err != nil {
		return err
	}
	var body map[string]any
	if json.Unmarshal([]byte(wire), &body) != nil {
		return refuse("regional native request invalid")
	}
	var native []map[string]any
	if r.API == "openai-responses" {
		for _, raw := range regionalArray(body["input"]) {
			item, _ := raw.(map[string]any)
			if item["type"] == "reasoning" {
				native = append(native, item)
			}
		}
	} else {
		for i, raw := range regionalArray(body["messages"]) {
			message, _ := raw.(map[string]any)
			if message["role"] != "assistant" {
				continue
			}
			if r.Combo == "zai-chat" {
				var text strings.Builder
				for _, block := range firstMessage.Content {
					if b, ok := block.(ai.Text); ok {
						text.WriteString(b.Text)
					}
				}
				if i != 1 || message["content"] != text.String() {
					continue
				}
			}
			if r.API == "openai-completions" {
				native = append(native, message)
				continue
			}
			for _, rawBlock := range regionalArray(message["content"]) {
				block, _ := rawBlock.(map[string]any)
				if block["type"] == "thinking" {
					native = append(native, block)
				}
			}
		}
	}
	for _, block := range firstMessage.Content {
		b, ok := block.(ai.Thinking)
		if !ok || b.Thinking == "" {
			continue
		}
		for _, item := range native {
			switch r.API {
			case "openai-completions":
				if item["reasoning_content"] == b.Thinking || regionalPartText(item["reasoning_details"]) == b.Thinking {
					return nil
				}
			case "openai-responses":
				if regionalPartText(item["content"]) == b.Thinking || regionalPartText(item["summary"]) == b.Thinking {
					return nil
				}
			case "anthropic-messages":
				if item["thinking"] == b.Thinking && (b.Signature != "" && item["signature"] == b.Signature || b.Signature == "" && item["signature"] == nil && r.Provider == "zai") {
					return nil
				}
			}
		}
	}
	return refuse("regional own reasoning/signature was not replayed in a native wire container")
}

// The recorded domain request must extend that exact predeclared turn. Wire
// comparison below separately proves native data was actually sent; a copied
// reasoning string or another successful response cannot become the source.
func verifyRegionalNativeRequest(root *os.Root, source, next string, firstResult []byte) error {
	type request struct {
		SystemPrompt string
		Messages     []json.RawMessage
	}
	var before, after request
	raw, err := regionalRead(root, source+"/request.json")
	if err != nil || json.Unmarshal(raw, &before) != nil {
		return refuse("regional native source request missing")
	}
	raw, err = regionalRead(root, next+"/request.json")
	if err != nil || json.Unmarshal(raw, &after) != nil || before.SystemPrompt != after.SystemPrompt || len(before.Messages) == 0 || len(after.Messages) <= len(before.Messages) {
		return refuse("regional native continuation request missing")
	}
	var sourceMessage struct{ Message json.RawMessage }
	if json.Unmarshal(firstResult, &sourceMessage) != nil {
		return refuse("regional native source message missing")
	}
	want := append(before.Messages, sourceMessage.Message)
	for i, raw := range want {
		var a, b any
		if json.Unmarshal(raw, &a) != nil || json.Unmarshal(after.Messages[i], &b) != nil || !reflect.DeepEqual(a, b) {
			return refuse("regional native continuation does not extend fixed source")
		}
	}
	return nil
}

func regionalArray(value any) []any { out, _ := value.([]any); return out }
func regionalPartText(value any) string {
	if item, ok := value.(map[string]any); ok {
		text, _ := item["text"].(string)
		return text
	}
	var parts []string
	for _, raw := range regionalArray(value) {
		item, _ := raw.(map[string]any)
		if text, ok := item["text"].(string); ok {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}
