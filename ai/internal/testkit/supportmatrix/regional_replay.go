package supportmatrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/tokenbeat-lab/barness/ai"
)

// Response evidence is interpreted by the public Client's real SDK/codec.
// This transport cannot dial: captured bytes are its only response source.
// Request authority is checked separately against the actual captured wire;
// replay's synthetic prompt has no bearing on response parsing.
type regionalReplay struct {
	binding  ai.Binding
	body     string
	exchange regionalExchange
}

func (h regionalReplay) ResolveBinding(context.Context, ai.CallScope, string) (ai.Binding, error) {
	return h.binding, nil
}
func (h regionalReplay) ResolveCredential(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
	return ai.Credential{OwnerTenantID: h.binding.TenantID, CredentialID: h.binding.CredentialRef, Version: "regional-replay", BindingVersion: h.binding.Version, AccountScopeID: h.binding.AccountScopeID, Active: true, APIKey: ai.NewSecret("synthetic-regional-replay")}, nil
}
func (h regionalReplay) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		req.Body.Close()
	}
	if req.URL.String() != h.exchange.URL {
		return nil, refuse("regional replay target differs")
	}
	header := http.Header{}
	for k, v := range h.exchange.ResponseHeaders {
		header.Set(k, v)
	}
	return &http.Response{StatusCode: h.exchange.Status, Header: header, Body: io.NopCloser(strings.NewReader(h.body)), Request: req}, nil
}

func verifyRegionalResponse(root *os.Root, r Report, s ScenarioResult, x regionalExchange, saved ai.AssistantMessage, attempt ai.Attempt) (ai.Usage, error) {
	cat, err := RegionalCandidateCatalog(r.Combo)
	if r.Regional.Stage == "formal" {
		cat, err = RegionalFormalCatalog(r.Combo)
	}
	if err != nil {
		return ai.Usage{}, err
	}
	raw, err := regionalRead(root, s.CaseID+"/tools.json")
	if err != nil {
		return ai.Usage{}, err
	}
	var tools []ai.Tool
	if regionalDecode(raw, &tools) != nil {
		return ai.Usage{}, refuse("regional replay tools invalid")
	}
	raw, err = regionalRead(root, s.CaseID+"/options.json")
	if err != nil {
		return ai.Usage{}, err
	}
	var opts ai.Options
	switch r.API {
	case "openai-completions":
		var o ai.ChatOptions
		if regionalDecode(raw, &o) != nil {
			return ai.Usage{}, refuse("regional replay options invalid")
		}
		opts = o
	case "openai-responses":
		var o ai.ResponsesOptions
		if regionalDecode(raw, &o) != nil {
			return ai.Usage{}, refuse("regional replay options invalid")
		}
		opts = o
	default:
		var o ai.AnthropicOptions
		if regionalDecode(raw, &o) != nil {
			return ai.Usage{}, refuse("regional replay options invalid")
		}
		opts = o
	}
	replay := func(body string) (ai.Result, error) {
		host := regionalReplay{binding: ai.Binding{TenantID: "regional-replay", BindingID: r.Combo, Operation: ai.OperationChat, Version: "regional-replay", Enabled: true, ProviderID: ai.ProviderID(r.Provider), API: ai.API(r.API), Endpoint: r.Endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: "regional:" + r.AccountAlias, CredentialRef: "regional-replay", AllowedModels: []string{s.Model}}, body: body, exchange: x}
		policy := &ai.ResourcePolicy{
			MaxRequestBytes: 256 << 10, MaxImageBytes: 1 << 10, MaxFrameBytes: 512 << 10,
			MaxToolJSONBytes: 256 << 10, MaxErrorBodyBytes: 16 << 10, MaxOutputBytes: 1 << 20,
			MaxQueuedEvents: 8192, MaxQueuedEventBytes: 2 << 20, MaxConcurrentPerTenant: 1,
			MaxConcurrentProcess: 1, MaxAdmissionWaiters: 1, AdmissionWait: time.Second,
			CallTimeout: 5 * time.Second, ConnectTimeout: time.Second,
			ResponseHeaderTimeout: time.Second, ReadIdleTimeout: time.Second,
		}
		client, e := ai.NewClient(ai.Config{Catalog: &cat, Policy: policy, Bindings: host, Credentials: host, Transport: host})
		if e != nil {
			return ai.Result{}, e
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return client.Complete(ctx, ai.CallScope{TenantID: host.binding.TenantID, RequestID: "regional-replay"}, ai.Target{BindingID: r.Combo, ModelID: s.Model}, ai.Request{Messages: []ai.Message{ai.UserText("Replay captured response.")}, Tools: tools}, opts)
	}
	_, kind, _ := strings.Cut(s.ID, "/")
	if kind != "cancel" {
		got, e := replay(x.ResponseBody)
		if e != nil || len(got.Metadata.Attempts) != 1 || got.Message.StopReason != ai.StopReasonStop && !(kind == "tool-call" && got.Message.StopReason == ai.StopReasonToolUse) {
			code, phase := ai.Code(""), ai.Phase("")
			if failure, ok := e.(*ai.Error); ok {
				code, phase = failure.Code, failure.Phase
			}
			return ai.Usage{}, fmt.Errorf("%w: regional raw response %s unsuccessful: stop=%s attempts=%d code=%s phase=%s", ErrInvalidReport, s.ID, got.Message.StopReason, len(got.Metadata.Attempts), code, phase)
		}
		if !regionalSameMessage(got.Message, saved) || got.Metadata.Attempts[0].UsageReporting != attempt.UsageReporting || RegionalIsKimi(r.Combo) && !reflect.DeepEqual(got.Metadata.Attempts[0].CostEstimate, attempt.CostEstimate) {
			return ai.Usage{}, refuse("regional raw response contradicts Result usage, content or terminal")
		}
		return got.Message.Usage, nil
	}
	// Cancellation may cut a frame or leave captured read-ahead beyond the
	// emitted Result. Only complete observed frames can justify its usage; the
	// emitted content/native fields must be a prefix of that decoded evidence.
	decoder := ssestream.NewDecoder(&http.Response{Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(x.ResponseBody))})
	defer decoder.Close()
	var prefix strings.Builder
	for n := 0; decoder.Next() && n < 4096; n++ {
		event := decoder.Event()
		if !json.Valid(event.Data) {
			// The real codec cannot reach later events after malformed JSON.
			// A valid earlier cancellation prefix has already been accepted;
			// an incomplete trailing frame cannot justify later content.
			break
		}
		prefix.WriteString("event: " + event.Type + "\ndata: " + string(event.Data) + "\n\n")
		got, _ := replay(prefix.String())
		if len(got.Metadata.Attempts) == 1 && got.Metadata.Attempts[0].UsageReporting == attempt.UsageReporting && (!RegionalIsKimi(r.Combo) || reflect.DeepEqual(got.Metadata.Attempts[0].CostEstimate, attempt.CostEstimate)) && reflect.DeepEqual(got.Message.Usage, saved.Usage) && regionalMessagePrefix(saved, got.Message) {
			return got.Message.Usage, nil
		}
	}
	return ai.Usage{}, refuse("regional canceled Result lacks observed raw prefix")
}

func regionalSameMessage(a, b ai.AssistantMessage) bool {
	// Wall clock and local trust handles cannot be reproduced from wire bytes.
	a.Timestamp, b.Timestamp = 0, 0
	a.DurationMs, b.DurationMs = ai.Nullable[int64]{}, ai.Nullable[int64]{}
	a.NativeState, b.NativeState = ai.TrustedNativeState{}, ai.TrustedNativeState{}
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

func regionalMessagePrefix(saved, observed ai.AssistantMessage) bool {
	if saved.ResponseID != "" && saved.ResponseID != observed.ResponseID || saved.RawStopReason != "" && saved.RawStopReason != observed.RawStopReason || len(saved.Content) > len(observed.Content) {
		return false
	}
	for i, b := range saved.Content {
		switch b := b.(type) {
		case ai.Text:
			v, ok := observed.Content[i].(ai.Text)
			if !ok || !strings.HasPrefix(v.Text, b.Text) || b.Signature != "" && b.Signature != v.Signature {
				return false
			}
		case ai.Thinking:
			v, ok := observed.Content[i].(ai.Thinking)
			if !ok || !strings.HasPrefix(v.Thinking, b.Thinking) || !strings.HasPrefix(v.Signature, b.Signature) || b.Redacted != v.Redacted {
				return false
			}
		case ai.ToolCall:
			v, ok := observed.Content[i].(ai.ToolCall)
			if !ok || !strings.HasPrefix(v.ID, b.ID) || !strings.HasPrefix(v.Name, b.Name) || !strings.HasPrefix(v.RawArguments, b.RawArguments) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
