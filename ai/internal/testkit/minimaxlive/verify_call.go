package minimaxlive

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"path/filepath"
	"reflect"
)

type Exchange struct {
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

func verifyCall(dir string, r Report, c Call, p Case) error {
	base := filepath.Join(dir, c.CaseID)
	read := func(name string, v any) error {
		raw, e := ReadFile(filepath.Join(base, name+".json"), 2<<20)
		if e != nil {
			return e
		}
		return json.Unmarshal(raw, v)
	}
	var result Capture
	var ex []Exchange
	var obs []ai.Observation
	var events []json.RawMessage
	if read("result", &result) != nil || read("exchanges", &ex) != nil || read("observations", &obs) != nil || read("events", &events) != nil {
		return fmt.Errorf("MiniMax captured call incomplete")
	}
	var assertions struct {
		Status     string `json:"status"`
		Assertions []struct {
			Pass bool `json:"pass"`
		} `json:"assertions"`
	}
	if read("assertions", &assertions) != nil || assertions.Status != "PASS" || len(assertions.Assertions) == 0 {
		return fmt.Errorf("MiniMax assertions missing")
	}
	for _, a := range assertions.Assertions {
		if !a.Pass {
			return fmt.Errorf("MiniMax assertion failure")
		}
	}
	if c.Reservation.Kind != "reserve" || c.Reservation.Case != c.ID || c.Reservation.Identity != r.Identity || c.Reservation.Output != p.Output || c.Reservation.Input < c.InputBytes+1024 || c.Reservation.Hash != receiptHash(c.Reservation) {
		return fmt.Errorf("MiniMax call reservation mismatch")
	}
	meta := result.Result.Metadata
	if c.ID == "preflight-error" {
		if len(ex) != 0 || len(meta.Attempts) != 0 || len(c.Attempts) != 0 || result.ErrorCode == "" {
			return fmt.Errorf("MiniMax preflight attempted network")
		}
		return nil
	}
	if len(ex) != 1 || len(meta.Attempts) != 1 || len(c.Attempts) != 1 || len(obs) != 4 || obs[3].Kind != ai.ObservationCallFinished {
		return fmt.Errorf("MiniMax omitted or extra actual Attempt")
	}
	if meta.ProviderID != ai.ProviderMiniMax || meta.ModelID != r.Identity.Model || meta.API != r.Identity.API || meta.CatalogHash != r.CatalogHash || meta.RequestID != c.RequestID {
		return fmt.Errorf("MiniMax actual attribution differs")
	}
	if c.Attempts[0].Kind != "attempt" || c.Attempts[0].Case != c.ID || c.Attempts[0].Identity != r.Identity || c.Attempts[0].Hash != receiptHash(c.Attempts[0]) {
		return fmt.Errorf("MiniMax attempt receipt differs")
	}
	for _, o := range obs {
		if o.Call.RequestID != c.RequestID {
			return fmt.Errorf("MiniMax Observer attribution differs")
		}
	}
	if !reflect.DeepEqual(obs[3].Call.Attempts, meta.Attempts) || !reflect.DeepEqual(obs[2].Attempt, &meta.Attempts[0]) {
		return fmt.Errorf("MiniMax Observer/Attempt mismatch")
	}
	if !reflect.DeepEqual(result.Result.Message.Usage, meta.Attempts[0].Usage) || !reflect.DeepEqual(obs[3].Usage, meta.Attempts[0].Usage) {
		return fmt.Errorf("MiniMax Result/Attempt/Observer usage differs")
	}
	if e := verifyEvents(events, c, result); e != nil {
		return e
	}
	models, usage := RawFacts(ex[0].ResponseBody)
	if !reflect.DeepEqual(models, c.RawModels) || !JSONEqual(usage, c.UsagePresence) || len(models) == 0 && c.ID != "cancel" && c.Failure != "transient-request" {
		return fmt.Errorf("MiniMax raw model/usage proof differs")
	}
	for _, model := range models {
		if model != r.Identity.Model {
			return fmt.Errorf("MiniMax actual model differs")
		}
	}
	if e := verifyUsage(dir, r, result, usage); e != nil {
		return e
	}
	a := meta.Attempts[0]
	if a.Usage.CacheWrite > 0 {
		return fmt.Errorf("MiniMax cache-write pricing unverified; promotion/certification refused")
	}
	if a.Usage.Input+a.Usage.CacheRead+a.Usage.CacheWrite > int64(c.Reservation.Input) {
		return fmt.Errorf("MiniMax actual known input exceeds reservation")
	}
	for _, name := range []string{"request.json", "response.sse"} {
		raw, e := ReadFile(filepath.Join(base, name), 2<<20)
		want := ex[0].RequestBody
		if name == "response.sse" {
			want = ex[0].ResponseBody
		}
		if e != nil || string(raw) != want || ex[0].Truncated {
			return fmt.Errorf("MiniMax raw capture mismatch")
		}
	}
	if a.ProviderRequestID != "" {
		found := false
		for _, v := range ex[0].ResponseHeaders {
			found = found || a.ProviderRequestID == v
		}
		if !found {
			return fmt.Errorf("MiniMax actual diagnostic ID has no captured source")
		}
	}
	var req map[string]any
	if json.Unmarshal([]byte(ex[0].RequestBody), &req) != nil || req["model"] != r.Identity.Model {
		return fmt.Errorf("MiniMax wire model differs")
	}
	suffix := "/chat/completions"
	if r.Identity.Route == "responses" {
		suffix = "/responses"
	}
	if r.Identity.Route == "messages" {
		// SDK BetaMessageService transport query is anticipated by R02/R15;
		// it grants no feature beta header/body capability.
		suffix = "/v1/messages?beta=true"
	}
	if ex[0].URL != r.Identity.Endpoint+suffix || ex[0].Method != "POST" {
		return fmt.Errorf("MiniMax actual endpoint differs")
	}
	if c.ID == grokhistory.Scenario {
		before, e1 := ReadFile(filepath.Join(base, "public-request.json"), 2<<20)
		after, e2 := ReadFile(filepath.Join(base, "immutable-input.json"), 2<<20)
		proof, e3 := ReadFile(filepath.Join(base, "history-provenance.json"), 4096)
		if e1 != nil || e2 != nil || e3 != nil || grokhistory.VerifyProvenance(proof, meta.TenantID) != nil || !grokhistory.VerifyHost(before) || !grokhistory.VerifyHost(after) || grokhistory.Verify(r.Identity.API, []byte(ex[0].RequestBody)) != nil || meta.NativeStateDowngrades != (ai.NativeStateDowngrades{CrossModel: 2}) {
			return fmt.Errorf("MiniMax fixed xAI input, wire or downgrades differ")
		}
	}
	if c.ID == "tool-result-native" {
		if e := verifyNativeContinuation(dir, r, req, result); e != nil {
			return e
		}
	}
	field := "max_tokens"
	if r.Identity.Route == "chat" {
		field = "max_completion_tokens"
	}
	if r.Identity.Route == "responses" {
		field = "max_output_tokens"
	}
	if req[field] != float64(p.Output) {
		return fmt.Errorf("MiniMax actual requested output differs")
	}
	if c.Failure == "transient-request" {
		if c.Outcome != "FAIL" || c.Rerun != 0 || !RequestRerunnable(result, events, ex[0].Status, ex[0].ResponseBody) {
			return fmt.Errorf("MiniMax invalid transient request rerun")
		}
		return nil
	}
	stop := result.Result.Message.StopReason
	if c.ID == "cancel" {
		if result.ErrorCode != ai.CodeCanceled || stop != ai.StopReasonAborted || len(events) == 0 {
			return fmt.Errorf("MiniMax cancel proof absent")
		}
	} else if c.ID == "truncate" {
		if result.ErrorCode != "" || stop != ai.StopReasonLength {
			return fmt.Errorf("MiniMax truncation proof absent")
		}
	} else if result.ErrorCode != "" {
		return fmt.Errorf("MiniMax captured call failed")
	}
	if c.ID != "tool-call" && c.ID != "cancel" && c.ID != "truncate" {
		text := ""
		thoughts := 0
		for _, part := range result.Result.Message.Content {
			if v, ok := part.(ai.Text); ok {
				text += v.Text
			}
			if v, ok := part.(ai.Thinking); ok && v.Thinking != "" {
				thoughts++
			}
		}
		if text == "" || stop != ai.StopReasonStop {
			return fmt.Errorf("MiniMax required text/terminal absent")
		}
		if c.ID == "thinking-on" && thoughts == 0 {
			return fmt.Errorf("MiniMax own reasoning absent")
		}
		if c.ID == "thinking-control" && r.Identity.Model == M3 && thoughts != 0 {
			return fmt.Errorf("MiniMax binary off was ignored")
		}
	}
	if c.ID == "tool-call" && c.Outcome == "PASS" {
		tools, thoughts := 0, 0
		for _, part := range result.Result.Message.Content {
			switch v := part.(type) {
			case ai.ToolCall:
				tools++
			case ai.Thinking:
				if v.Thinking != "" {
					thoughts++
				}
			}
		}
		if tools == 0 || thoughts == 0 || stop != ai.StopReasonToolUse {
			return fmt.Errorf("MiniMax own tool/reasoning proof absent")
		}
	}
	return nil
}
