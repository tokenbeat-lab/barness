package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type glmUsageCase struct {
	Name                            string
	Events                          []json.RawMessage
	Reporting                       ai.UsageReporting
	Input, Output, CacheRead, Total int64
	CacheWrite                      int64
	CacheWriteCost                  float64
	Reasoning                       *int64
	Failure                         bool
	End                             provider.End
}

func glmOwnReply(t *testing.T, protocol string, events []json.RawMessage) provider.Reply {
	t.Helper()
	var chunks [][]byte
	for _, e := range events {
		var field struct{ Type string }
		mustUnmarshal(t, e, &field)
		prefix := ""
		if protocol != "chat" && field.Type != "" {
			prefix = "event: " + field.Type + "\n"
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, e); err != nil {
			t.Fatal(err)
		}
		chunks = append(chunks, []byte(prefix+"data: "+compact.String()+"\n\n"))
	}
	r := provider.SSE(chunks)
	r.Header["X-Request-Id"] = "glm-stream-request"
	return r
}

func TestGLMUsagePresence(t *testing.T) {
	raw, err := os.ReadFile("testdata/glm-failures/usage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Protocols map[string][]glmUsageCase }
	mustUnmarshal(t, raw, &fixture)
	for _, protocol := range []string{"responses", "chat", "messages"} {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, sc := range fixture.Protocols[protocol] {
				t.Run(protocol+"/"+model+"/"+sc.Name, func(t *testing.T) {
					ev := run.Case(t, "GLM08-usage-"+protocol+"-"+model+"-"+sc.Name)
					ev.Fixture("usage.json", raw)
					w, target, bodies := glmFailureWorld(t, protocol, 2)
					target.ModelID = model
					reply := glmOwnReply(t, protocol, sc.Events)
					reply.End = sc.End
					enqueue(ev, w, reply)
					o := drain(w.client.Stream(ctxFor(t), textScope("glm08-usage-"+protocol+model+sc.Name), target, glmResponsesRequest(), nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("expected terminal", (o.err != nil) == sc.Failure, "%v", o.err)
					u := o.result.Message.Usage
					ev.Check("counts do not duplicate cache or reasoning", u.Input == sc.Input && u.Output == sc.Output && u.CacheRead == sc.CacheRead && u.TotalTokens == sc.Total, "%+v", u)
					if sc.Reasoning != nil {
						n, ok := u.Reasoning.Get()
						ev.Check("reported reasoning retained inside output", ok && n == *sc.Reasoning, "%+v", u)
					}
					if protocol == "messages" {
						ev.Check("no invented reasoning count", u.Reasoning.IsZero(), "%+v", u)
						ev.Check("unverified one-hour cache semantics withheld", u.CacheWrite1h.IsZero(), "%+v", u)
						ev.Check("aggregate cache writes use configured rate without Anthropic premium", u.CacheWrite == sc.CacheWrite && math.Abs(u.Cost.CacheWrite-sc.CacheWriteCost) < 1e-15, "%+v", u)
					}
					a := o.result.Metadata.Attempts
					if ev.Check("stream never replayed", len(a) == 1 && len(w.provider.Requests()) == 1, "%+v", a) {
						ev.Check("usage presence retained on Attempt", a[0].UsageReporting == sc.Reporting && a[0].Usage == u, "%+v", a[0])
					}
					ev.Check("body released", bodies.Open() == 0, "%d", bodies.Open())
					if sc.Failure {
						ev.Check("failure remains caller-visible", errors.Is(o.err, &ai.Error{Phase: ai.PhaseStream, Code: ai.CodeTransport}) || errors.Is(o.err, &ai.Error{Phase: ai.PhaseStream, Code: ai.CodeUpstreamError}), "%v", o.err)
					}
				})
			}
		}
	}
}
