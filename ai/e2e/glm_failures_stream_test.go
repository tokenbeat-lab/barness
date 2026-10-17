package e2e

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMStreamFailures(t *testing.T) {
	raw, err := os.ReadFile("testdata/glm-failures/stream.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Protocols map[string][]struct {
			Name, Code, Business, RawStop string
			Events                        []json.RawMessage
		}
	}
	mustUnmarshal(t, raw, &f)
	for _, protocol := range []string{"responses", "chat", "messages"} {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, sc := range f.Protocols[protocol] {
				t.Run(protocol+"/"+model+"/"+sc.Name, func(t *testing.T) {
					ev := run.Case(t, "GLM08-stream-"+protocol+"-"+model+"-"+sc.Name)
					ev.Fixture("stream.json", raw)
					w, target, bodies := glmFailureWorld(t, protocol, 2)
					target.ModelID = model
					enqueue(ev, w, glmOwnReply(t, protocol, sc.Events))
					o := drain(w.client.Stream(ctxFor(t), textScope("glm08-stream-"+protocol+model+sc.Name), target, glmResponsesRequest(), nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					code := ai.CodeUpstreamError
					if sc.Code != "" {
						code = ai.Code(sc.Code)
					}
					ev.Check("HTTP200 stream failure is never success", errors.Is(o.err, &ai.Error{Code: code, Phase: ai.PhaseStream}) && o.result.Message.StopReason == ai.StopReasonError, "%v", o.err)
					var ae *ai.Error
					if errors.As(o.err, &ae) {
						ev.Check("actual HTTP200/request ID retained", ae.HTTPStatus == 200 && ae.ProviderRequestID == "glm-stream-request", "%+v", ae)
						ev.Check("private echoed content withheld", !strings.Contains(ae.Message, "SYNTHETIC_PRIVATE"), "%s", ae.Message)
						if sc.Business != "" {
							ev.Check("safe diagnostic business code", strings.Contains(ae.Message, sc.Business), "%s", ae.Message)
						}
					}
					if sc.RawStop != "" {
						ev.Check("provider stop retained", o.result.Message.RawStopReason == sc.RawStop, "%+v", o.result.Message)
					}
					ev.Check("stream never replayed and body released", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1 && bodies.Open() == 0, "%+v", o.result.Metadata.Attempts)
				})
			}
		}
	}
}
