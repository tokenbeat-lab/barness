//go:build live

package live

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
)

// This assembly is deliberately outside legacy combos/scenarios/budget. The
// isolated ticket-11 runner must reserve its 24 calls/32 attempts before using
// this Client and enforce its own double opt-in, key isolation and process cap.
func newGLMCandidateClient(c *combo, key string, rec *recorder, cat ai.Catalog) (*ai.Client, error) {
	if c == nil || c.provider != ai.ProviderZhipu || c.operation != ai.OperationChat {
		return nil, fmt.Errorf("GLM candidate assembly requires an explicit Zhipu chat route")
	}
	policy := livePolicy()
	policy.CallTimeout = 120 * time.Second
	policy.ResponseHeaderTimeout = 120 * time.Second
	var models []string
	for _, model := range cat.Models {
		if model.Provider == c.provider && model.API == c.api {
			models = append(models, model.ID)
		}
	}
	// This assembly is independently source-pinned by GLM. Keeping the capture
	// policy here avoids changing the shared recorder or other formal proofs.
	r := &resolver{
		binding: ai.Binding{TenantID: liveTenant, BindingID: c.name, Version: "live", Enabled: true,
			Operation: c.operation, ProviderID: c.provider, API: c.api, Endpoint: c.endpoint, AuthKind: ai.AuthAPIKey,
			AccountScopeID: "live-" + c.name, CredentialRef: "live-key", AllowedModels: models},
		credential: ai.Credential{OwnerTenantID: liveTenant, CredentialID: "live-key", Version: "live", BindingVersion: "live",
			AccountScopeID: "live-" + c.name, Active: true, APIKey: ai.NewSecret(key)},
	}
	return ai.NewClient(ai.Config{Policy: policy, Bindings: r, Credentials: r, Transport: glmCaptureTransport{rec}, Observer: rec, Catalog: &cat})
}

// This is an offline public Client acceptance test despite the package's live
// build tag: responses are synthetic or served by a controlled loopback TLS peer.
func TestGLMCandidateAssemblyControlled(t *testing.T) {
	c := &combo{name: "zhipu-chat-candidate", operation: ai.OperationChat, provider: ai.ProviderZhipu,
		api: ai.APIOpenAICompletions, endpoint: glm.ChatEndpoint}
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM02-controlled-candidate-assembly-"+model)
			cat := glm.ChatCatalog()
			ev.Record("candidate-catalog", cat)
			rec := newRecorder()
			rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				deadline, ok := req.Context().Deadline()
				ev.Check("candidate call bounded at 120 seconds", ok && time.Until(deadline) <= 120*time.Second && time.Until(deadline) > 110*time.Second, "deadline=%v", deadline)
				ev.Check("fixed standard resource and explicit credential", req.URL.String() == glm.ChatEndpoint+"/chat/completions" && req.Header.Get("Authorization") == "Bearer synthetic-candidate-key", "url=%s", req.URL)
				body := fmt.Sprintf("data: {\"model\":%q,\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Controlled GLM.\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n", model)
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})
			client, err := newGLMCandidateClient(c, "synthetic-candidate-key", rec, cat)
			if !ev.Check("candidate client assembles", err == nil, "%v", err) {
				return
			}
			res, err := client.CompleteSimple(context.Background(), newScope(), ai.Target{BindingID: c.name, ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Controlled fixture.")}}, ai.SimpleOptions{MaxTokens: ai.Value(4096)})
			ev.Record("result", res)
			ev.Record("error", errText(err))
			ev.Record("exchanges", rec.take())
			ev.Check("candidate generation succeeded with one attempt", err == nil && len(res.Metadata.Attempts) == 1 && res.Metadata.API == c.api && res.Metadata.ProviderID == c.provider && res.Metadata.ModelID == model, "%v", err)
		})
	}
	t.Run("response-headers-after-90s", func(t *testing.T) {
		// The public header timer has no fake-time seam. One shared-factory
		// Responses fixture crosses the actual old90s/new120s boundary; other
		// routes keep their existing controlled/early-Host/cancel coverage.
		ev := run.Case(t, "GLM02-controlled-headers-after-90s")
		var requests atomic.Int32
		headers := make(chan time.Duration, 1)
		peer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requests.Add(1)
			io.Copy(io.Discard, req.Body)
			began := time.Now()
			timer := time.NewTimer(91 * time.Second)
			defer timer.Stop()
			select {
			case <-req.Context().Done():
				return
			case <-timer.C:
			}
			headers <- time.Since(began)
			w.Header().Set("Content-Type", "text/event-stream")
			io.WriteString(w, "data: {\"type\":\"response.output_item.done\",\"output_index\":0,\"item\":{\"type\":\"message\",\"id\":\"msg_headers\",\"role\":\"assistant\",\"content\":{\"type\":\"output_text\",\"text\":\"Controlled GLM.\"}}}\n\n")
			io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"response_headers\",\"model\":\"glm-5.3\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"id\":\"msg_headers\",\"role\":\"assistant\",\"content\":{\"type\":\"output_text\",\"text\":\"Controlled GLM.\"}}],\"usage\":{\"input_tokens\":2,\"output_tokens\":3,\"total_tokens\":5}}}\n\n")
		}))
		defer peer.Close()
		local, err := url.Parse(peer.URL)
		if err != nil {
			t.Fatal(err)
		}
		rec := newRecorder()
		rec.next.(*http.Transport).CloseIdleConnections()
		transport := peer.Client().Transport.(*http.Transport)
		defer transport.CloseIdleConnections()
		rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			ev.Check("whole GLM call remains bounded at120s", ok && time.Until(deadline) <= 120*time.Second && time.Until(deadline) > 110*time.Second, "deadline missing or changed")
			ev.Check("binding remains official Responses with synthetic credential", req.URL.String() == glm.ResponsesEndpoint+"/responses" && req.Header.Get("Authorization") == "Bearer synthetic-candidate-key", "official binding changed")
			// Only this controlled transport maps the request to loopback TLS;
			// the production Client binding and captured URL stay unchanged.
			copy := req.Clone(req.Context())
			mapped := *req.URL
			mapped.Scheme, mapped.Host = local.Scheme, local.Host
			copy.URL, copy.Host = &mapped, local.Host
			return transport.RoundTrip(copy)
		})
		c := &combo{name: "zhipu-responses-candidate", operation: ai.OperationChat, provider: ai.ProviderZhipu, api: ai.APIOpenAIResponses, endpoint: glm.ResponsesEndpoint}
		client, err := newGLMCandidateClient(c, "synthetic-candidate-key", rec, glm.ResponsesCatalog())
		if !ev.Check("controlled Responses client assembles", err == nil, "%v", err) {
			return
		}
		began := time.Now()
		res, err := client.CompleteSimple(context.Background(), newScope(), ai.Target{BindingID: c.name, ModelID: "glm-5.3"}, ai.Request{Messages: []ai.Message{ai.UserText("Controlled fixture.")}}, ai.SimpleOptions{MaxTokens: ai.Value(4096)})
		var delay time.Duration
		select {
		case delay = <-headers:
		default:
		}
		ev.Record("result", res)
		ev.Record("error", errText(err))
		ev.Record("exchanges", rec.take())
		ev.Record("timing", map[string]any{"headersAfterMs": delay.Milliseconds(), "callElapsedMs": time.Since(began).Milliseconds(), "transportRequests": requests.Load(), "localTLSPeer": true, "vendorCalls": 0})
		ev.Check("headers beyond90s complete before120s with one own attempt", err == nil && delay > 90*time.Second && delay < 120*time.Second && time.Since(began) < 120*time.Second && requests.Load() == 1 && len(res.Metadata.Attempts) == 1, "%v", err)
		ev.Check("complete controlled text and terminal preserved", err == nil && res.Message.StopReason == ai.StopReasonStop && textOf(res.Message) == "Controlled GLM.", "stop=%s text=%q", res.Message.StopReason, textOf(res.Message))
	})
}
