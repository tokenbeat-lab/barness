package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMHTTPFailures(t *testing.T) {
	fixtures, raw := glmFailureFixtures(t)
	for _, protocol := range []string{"responses", "chat", "messages"} {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, sc := range fixtures[protocol] {
				t.Run(protocol+"/"+model+"/"+sc.Name, func(t *testing.T) {
					ev := run.Case(t, "GLM08-http-"+protocol+"-"+model+"-"+sc.Name)
					ev.Fixture("http.json", raw)
					w, target, bodies := glmFailureWorld(t, protocol, sc.Retries)
					target.ModelID = model
					reply := provider.Reply{Status: sc.Status, Header: map[string]string{"Content-Type": "application/json", "X-Request-Id": "glm-own-request-id", "X-Ds-Trace-Id": "wrong-provider-id", "retry-after-ms": "1", "x-should-retry": sc.ShouldRetry}, Chunks: [][]byte{sc.Body}, Script: string(sc.Body)}
					replies := make([]provider.Reply, sc.Retries+1)
					for i := range replies {
						replies[i] = reply
					}
					enqueue(ev, w, replies...)
					o := drain(w.client.Stream(ctxFor(t), textScope("glm08-http-"+protocol+model+sc.Name), target, glmResponsesRequest(), nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					want := ai.CodeInvalidRequest
					switch {
					case sc.Status == 401 || sc.Status == 403:
						want = ai.CodeUpstreamAuth
					case sc.Status == 429:
						want = ai.CodeRateLimited
					case sc.Status >= 500:
						want = ai.CodeUpstreamError
					}
					ev.Check("stable HTTP classification", errors.Is(o.err, &ai.Error{Code: want, Phase: ai.PhaseRequest}), "%v", o.err)
					var ae *ai.Error
					if ev.Check("classified error", errors.As(o.err, &ae), "%v", o.err) {
						ev.Check("status, provider ID and requested delay preserved", ae.HTTPStatus == sc.Status && ae.ProviderRequestID == "glm-own-request-id" && ae.RetryAfter > 0, "%+v", ae)
						ev.Check("upstream echo withheld", !strings.Contains(ae.Message, "SYNTHETIC_PRIVATE") && !strings.Contains(o.result.Message.ErrorMessage, "SYNTHETIC_PRIVATE"), "%q", ae.Message)
						if sc.Code != "" {
							ev.Check("safe business code retained", strings.Contains(ae.Message, sc.Code), "%q", ae.Message)
						}
					}
					ev.Check("bounded attempts exclude permanent failure", len(w.provider.Requests()) == sc.Requests && len(o.result.Metadata.Attempts) == sc.Requests, "requests=%d attempts=%+v", len(w.provider.Requests()), o.result.Metadata.Attempts)
					ev.Check("all bodies released", bodies.Open() == 0, "open=%d", bodies.Open())
					for _, a := range o.result.Metadata.Attempts {
						ev.Check("HTTP failures report no known consumption", a.UsageReporting == ai.UsageUnreported, "%+v", a)
					}
				})
			}
		}
	}
}
