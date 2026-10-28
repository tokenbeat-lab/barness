package minimaxlive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
)

// A read-only transport has no dialer. It compares the real current encoder's
// wire before parsing the retained response; it never replaces a payload hook.
type thinkingReplay struct {
	binding  ai.Binding
	exchange Exchange
	calls    int
	failure  string
}

func (h *thinkingReplay) ResolveBinding(context.Context, ai.CallScope, string) (ai.Binding, error) {
	return h.binding, nil
}
func (h *thinkingReplay) ResolveCredential(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
	return ai.Credential{OwnerTenantID: h.binding.TenantID, CredentialID: h.binding.CredentialRef, Version: "thinking-replay", BindingVersion: h.binding.Version, AccountScopeID: h.binding.AccountScopeID, Active: true, APIKey: ai.NewSecret("synthetic-offline-replay")}, nil
}
func (h *thinkingReplay) RoundTrip(req *http.Request) (*http.Response, error) {
	h.calls++
	raw, e := io.ReadAll(req.Body)
	req.Body.Close()
	var current, original any
	sameWire := json.Unmarshal(raw, &current) == nil && json.Unmarshal([]byte(h.exchange.RequestBody), &original) == nil && reflect.DeepEqual(current, original)
	if e != nil || h.calls != 1 || !sameWire || req.URL.String() != h.exchange.URL || req.Method != h.exchange.Method {
		h.failure = fmt.Sprintf("wire bytes=%d/%d sha=%s/%s method=%s/%s target=%s/%s", len(raw), len(h.exchange.RequestBody), Hash(raw), Hash([]byte(h.exchange.RequestBody)), req.Method, h.exchange.Method, req.URL.String(), h.exchange.URL)
		return nil, fmt.Errorf("MiniMax current encoded request differs from original wire")
	}
	if h.exchange.Status == 0 {
		return nil, errors.New(h.exchange.TransportError)
	}
	header := http.Header{}
	for k, v := range h.exchange.ResponseHeaders {
		header.Set(k, v)
	}
	return &http.Response{StatusCode: h.exchange.Status, Header: header, Body: io.NopCloser(strings.NewReader(h.exchange.ResponseBody)), Request: req}, nil
}

func verifyThinkingReplay(dir string, r Report, c Call) error {
	base := filepath.Join(dir, c.CaseID)
	read := func(n string, v any) error {
		b, e := ReadFile(filepath.Join(base, n+".json"), 2<<20)
		if e != nil {
			return e
		}
		return json.Unmarshal(b, v)
	}
	var saved Capture
	var ex []Exchange
	var opts ai.AnthropicOptions
	if read("result", &saved) != nil || read("exchanges", &ex) != nil || read("options", &opts) != nil {
		return fmt.Errorf("MiniMax current replay inputs missing")
	}
	req, e := thinkingReplayRequest(dir, r, c, saved.Result.Metadata)
	if e != nil {
		return e
	}
	var cat ai.Catalog
	raw, e := ReadFile(filepath.Join(dir, "minimax-catalog.json"), 2<<20)
	if e != nil || json.Unmarshal(raw, &cat) != nil {
		return fmt.Errorf("MiniMax replay catalog missing")
	}
	meta := saved.Result.Metadata
	host := &thinkingReplay{binding: ai.Binding{TenantID: meta.TenantID, BindingID: meta.BindingID, Operation: ai.OperationChat, Version: "thinking-replay", Enabled: true, ProviderID: ai.ProviderMiniMax, API: r.Identity.API, Endpoint: r.Identity.Endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: meta.AccountScopeID, CredentialRef: "thinking-replay", AllowedModels: []string{r.Identity.Model}}}
	if c.ID != "preflight-error" {
		if len(ex) != 1 {
			return fmt.Errorf("MiniMax replay exchange differs")
		}
		host.exchange = ex[0]
	}
	policy := &ai.ResourcePolicy{MaxRequestBytes: 256 << 10, MaxImageBytes: 1 << 10, MaxFrameBytes: 512 << 10, MaxToolJSONBytes: 256 << 10, MaxErrorBodyBytes: 16 << 10, MaxOutputBytes: 1 << 20, MaxQueuedEvents: 8192, MaxQueuedEventBytes: 2 << 20, MaxConcurrentPerTenant: 1, MaxConcurrentProcess: 1, MaxAdmissionWaiters: 1, AdmissionWait: time.Second, CallTimeout: 5 * time.Second, ConnectTimeout: time.Second, ResponseHeaderTimeout: time.Second, ReadIdleTimeout: time.Second}
	client, e := ai.NewClient(ai.Config{Catalog: &cat, Policy: policy, Bindings: host, Credentials: host, Transport: host})
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	scope := ai.CallScope{TenantID: meta.TenantID, RequestID: "minimax-offline-replay"}
	target := ai.Target{BindingID: meta.BindingID, ModelID: r.Identity.Model}
	var result ai.Result
	switch c.ID {
	case "full-complete":
		result, e = client.Complete(ctx, scope, target, req, opts)
	case "simple-complete":
		result, e = client.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{MaxTokens: opts.MaxTokens})
	default:
		var stream *ai.Stream
		if c.ID == "simple-stream" {
			stream = client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{MaxTokens: opts.MaxTokens})
		} else {
			stream = client.Stream(ctx, scope, target, req, opts)
		}
		for stream.Next() {
			v := stream.Event()
			if c.ID == "cancel" && (v.Type() == ai.EventTextDelta || v.Type() == ai.EventThinkingDelta) {
				cancel()
			}
		}
		result, e = stream.Result()
		stream.Close()
	}
	if c.ID == "preflight-error" {
		if e == nil || host.calls != 0 || len(result.Metadata.Attempts) != 0 {
			return fmt.Errorf("MiniMax replay preflight attempted HTTP")
		}
		return nil
	}
	if host.calls != 1 || len(result.Metadata.Attempts) != 1 {
		return fmt.Errorf("MiniMax replay did not encode one captured request: %s", c.ID)
	}
	if c.ID == "cancel" {
		// Cancellation timing is local. The captured prefix, not an invented full
		// response, must justify the saved content and known usage.
		if e == nil || !thinkingContentPrefix(saved.Result.Message, result.Message) {
			return fmt.Errorf("MiniMax replay cancellation prefix differs")
		}
		return nil
	}
	if e != nil {
		return fmt.Errorf("MiniMax current raw response replay failed (%s): %w", c.ID, fmt.Errorf("%w; %s", e, host.failure))
	}
	a, b := saved.Result.Message, result.Message
	a.Timestamp, b.Timestamp = 0, 0
	a.DurationMs, b.DurationMs = ai.Nullable[int64]{}, ai.Nullable[int64]{}
	a.NativeState, b.NativeState = ai.TrustedNativeState{}, ai.TrustedNativeState{}
	if !JSONEqual(a, b) || !reflect.DeepEqual(saved.Result.Metadata.Attempts[0].UsageReporting, result.Metadata.Attempts[0].UsageReporting) || !reflect.DeepEqual(saved.Result.Metadata.Attempts[0].CostEstimate, result.Metadata.Attempts[0].CostEstimate) {
		return fmt.Errorf("MiniMax current response interpretation differs: %s", c.ID)
	}
	return nil
}
func thinkingContentPrefix(saved, got ai.AssistantMessage) bool {
	if len(saved.Content) > len(got.Content) {
		return false
	}
	for i, p := range saved.Content {
		switch v := p.(type) {
		case ai.Text:
			w, ok := got.Content[i].(ai.Text)
			if !ok || !strings.HasPrefix(w.Text, v.Text) {
				return false
			}
		case ai.Thinking:
			w, ok := got.Content[i].(ai.Thinking)
			if !ok || !strings.HasPrefix(w.Thinking, v.Thinking) {
				return false
			}
		default:
			return false
		}
	}
	return reflect.DeepEqual(saved.Usage, got.Usage)
}

// Restore trust only from this bundle's successful tool-call result, with the
// exact same tenant/account/provider/model and independent request attribution.
func thinkingReplayRequest(dir string, r Report, c Call, meta ai.CallMetadata) (ai.Request, error) {
	raw, e := ReadFile(filepath.Join(dir, c.CaseID, "public-request.json"), 2<<20)
	if e != nil {
		return ai.Request{}, e
	}
	if c.ID == grokhistory.Scenario {
		if !grokhistory.VerifyHost(raw) {
			return ai.Request{}, fmt.Errorf("MiniMax foreign Host history differs")
		}
		return grokhistory.Request(ai.CallScope{TenantID: meta.TenantID})
	}
	var input struct {
		SystemPrompt string
		Messages     []json.RawMessage
		Tools        []ai.Tool
	}
	if json.Unmarshal(raw, &input) != nil {
		return ai.Request{}, fmt.Errorf("MiniMax public input invalid")
	}
	out := ai.Request{SystemPrompt: input.SystemPrompt, Tools: input.Tools}
	for _, m := range input.Messages {
		var fields map[string]json.RawMessage
		if json.Unmarshal(m, &fields) != nil {
			return out, fmt.Errorf("MiniMax public message invalid")
		}
		switch {
		case fields["api"] != nil:
			if c.ID != "tool-result-native" {
				return out, fmt.Errorf("unowned MiniMax assistant history")
			}
			stored, e := DecodeMessage(m)
			if e != nil {
				return out, e
			}
			var prior Capture
			found := false
			for _, p := range r.Calls {
				if p.ID == "tool-call" && p.Outcome == "PASS" {
					b, e := ReadFile(filepath.Join(dir, p.CaseID, "result.json"), 2<<20)
					if e != nil || json.Unmarshal(b, &prior) != nil {
						return out, fmt.Errorf("MiniMax own tool predecessor absent")
					}
					found = true
				}
			}
			p := prior.Result.Metadata
			if !found || !JSONEqual(stored, prior.Result.Message) || p.TenantID != meta.TenantID || p.AccountScopeID != meta.AccountScopeID || p.ProviderID != meta.ProviderID || p.API != meta.API || p.ModelID != meta.ModelID || p.BindingID != meta.BindingID || p.RequestID == meta.RequestID {
				return out, fmt.Errorf("MiniMax native predecessor trust differs")
			}
			stored.NativeState, e = ai.TrustNativeState(ai.CallScope{TenantID: meta.TenantID}, ai.NativeStateEnvelope{TenantID: p.TenantID, AccountScopeID: p.AccountScopeID, ProviderID: p.ProviderID, API: p.API, ModelID: p.ModelID})
			if e != nil {
				return out, e
			}
			out.Messages = append(out.Messages, stored)
		case fields["toolCallId"] != nil:
			var v struct {
				ToolCallID, ToolName string
				IsError              bool
				Content              []ai.Text
			}
			if c.ID != "tool-result-native" || json.Unmarshal(m, &v) != nil {
				return out, fmt.Errorf("MiniMax tool result invalid")
			}
			tool := ai.ToolResultMessage{ToolCallID: v.ToolCallID, ToolName: v.ToolName, IsError: v.IsError}
			for _, p := range v.Content {
				tool.Content = append(tool.Content, p)
			}
			out.Messages = append(out.Messages, tool)
		default:
			var v struct{ Content []ai.Text }
			if json.Unmarshal(m, &v) != nil {
				return out, fmt.Errorf("MiniMax user input invalid")
			}
			user := ai.UserMessage{}
			for _, p := range v.Content {
				user.Content = append(user.Content, p)
			}
			out.Messages = append(out.Messages, user)
		}
	}
	return out, nil
}
