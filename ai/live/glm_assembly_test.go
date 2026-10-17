//go:build live

package live

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	return newClientWithPolicy(c, key, rec, cat, policy)
}

// This is an offline public Client acceptance test despite the package's live
// build tag: the transport synthesizes responses and never opens a connection.
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
}
