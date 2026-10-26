package e2e

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIRouteRejectedBeforeCredentials(t *testing.T) {
	for _, kind := range []string{"messages", "chat-completions", "images", "classifier", "model", "us-endpoint", "foreign-endpoint", "http", "port", "userinfo", "query", "fragment", "root", "double-slash", "empty-port", "empty-fragment", "empty-query", "encoded-path", "loopback"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-E01-precredential-reject-"+kind)
			reads := 0
			w := newXAIWorld(t, nil, func(c *ai.Config) {
				c.Credentials = credentialResolverFunc(func(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
					reads++
					return ai.Credential{}, ai.ErrAccessDenied
				})
			})
			b := xaiBinding()
			target := xaiTarget()
			switch kind {
			case "messages":
				b.API = ai.APIAnthropicMessages
			case "chat-completions":
				b.API = ai.APIOpenAICompletions
			case "images":
				b.Operation = ai.OperationImage
			case "classifier":
				b.Operation = ai.OperationClassifier
			case "model":
				target.ModelID = "grok-4.7-fast"
				b.AllowedModels = append(b.AllowedModels, target.ModelID)
			case "us-endpoint":
				b.Endpoint = "https://us-east-1.api.x.ai/v1"
			case "foreign-endpoint":
				b.Endpoint = "https://example.invalid/v1"
			case "http":
				b.Endpoint = "http://api.x.ai/v1"
			case "port":
				b.Endpoint = "https://api.x.ai:8443/v1"
			case "userinfo":
				b.Endpoint = "https://user@api.x.ai/v1"
			case "query":
				b.Endpoint = "https://api.x.ai/v1?q=1"
			case "fragment":
				b.Endpoint = "https://api.x.ai/v1#f"
			case "empty-port":
				b.Endpoint = "https://api.x.ai:/v1"
			case "empty-fragment":
				b.Endpoint = "https://api.x.ai/v1#"
			case "empty-query":
				b.Endpoint = "https://api.x.ai/v1?"
			case "encoded-path":
				b.Endpoint = "https://api.x.ai/%76%31"
			case "loopback":
				b.Endpoint = w.provider.URL() + "/v1"
			case "root":
				b.Endpoint = "https://api.x.ai"
			case "double-slash":
				b.Endpoint = "https://api.x.ai/v1//"
			}
			w.host.PutBinding(b)
			result, err := w.client.Complete(ctxFor(t), textScope("xai-route-"+kind), target, xaiRequest(), nil)
			ev.Record("binding", b)
			ev.Record("result", result)
			var failure *ai.Error
			ev.Check("route is refused before secrets", errors.As(err, &failure) && reads == 0, "err=%v reads=%d", err, reads)
			ev.Check("refusal stays unresolved and sends no HTTP", !result.Metadata.Resolved && len(w.provider.Requests()) == 0 && len(result.Metadata.Attempts) == 0, "metadata=%+v HTTP=%d", result.Metadata, len(w.provider.Requests()))
		})
	}
}

func TestXAIOfficialTarget(t *testing.T) {
	for i, endpoint := range []string{"https://api.x.ai/v1", "https://api.x.ai/v1/", "https://api.x.ai:443/v1", "https://API.X.AI/v1"} {
		t.Run(endpoint, func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("XAI-E02-official-target-%d", i))
			w := newXAIWorld(t, nil)
			b := xaiBinding()
			b.Endpoint = endpoint
			w.host.PutBinding(b)
			reply := xaiReply(t)
			w.provider.Enqueue(reply)
			result, err := w.client.Complete(ctxFor(t), textScope("xai-official-target"), xaiTarget(), xaiRequest(), nil)
			ev.Record("script", reply)
			ev.Record("requests", w.provider.Requests())
			ev.Record("result", result)
			ev.Check("official target succeeds through injected transport", err == nil, "err=%v", err)
			requests := w.provider.Requests()
			ev.Check("single version prefix and bound Bearer identity", len(requests) == 1 && requests[0].Path == "/v1/responses" && requests[0].KeyAlias == xaiKey.alias, "requests=%+v", requests)
			ev.Check("resolved metadata retains independent identity", result.Metadata.Resolved && result.Metadata.ProviderID == ai.ProviderXAI && result.Metadata.API == ai.APIOpenAIResponses && result.Metadata.ModelID == "grok-4.7", "metadata=%+v", result.Metadata)
		})
	}
}
