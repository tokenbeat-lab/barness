package e2e

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Failure scripts alter one terminal element of each actual Messages capture.
// They are explicitly synthetic failures, not claims of real upstream faults.
func TestGLMMessagesTerminals(t *testing.T) {
	scenarios, prompt, raw := loadGLMMessages(t)
	for _, sc := range scenarios {
		for _, failure := range []string{"missing-message-stop", "missing-stop-reason", "stream-error", "http-error"} {
			t.Run(sc.Model+"/"+failure, func(t *testing.T) {
				ev := run.Case(t, "GLM06-terminal-"+sc.Model+"-"+failure)
				w := newGLMMessagesWorld(t, nil)
				ev.Fixture("text.json", raw)
				r := glmMessagesReply(t, sc)
				code, phase := ai.CodeProtocol, ai.PhaseStream
				script := r.Script
				switch failure {
				case "missing-message-stop":
					script = script[:strings.LastIndex(script, "event: message_stop")]
				case "missing-stop-reason":
					script = regexp.MustCompile(`"stop_reason"\s*:\s*"end_turn"`).ReplaceAllString(script, `"stop_reason":null`)
				case "stream-error":
					script = script[:strings.LastIndex(script, "event: message_stop")] + "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"api_error\",\"message\":\"synthetic fault\"}}\n\n"
					code = ai.CodeUpstreamError
				case "http-error":
					r = provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "X-Request-Id": sc.RequestID[0]}, Script: `{"message":"synthetic service failure"}`, Chunks: [][]byte{[]byte(`{"message":"synthetic service failure"}`)}}
					code, phase = ai.CodeUpstreamError, ai.PhaseRequest
				}
				if failure != "http-error" {
					r.Script, r.Chunks = script, [][]byte{[]byte(script)}
				}
				enqueue(ev, w, r)
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-messages-failure-"+failure), ai.Target{BindingID: "glm-messages", ModelID: sc.Model}, ai.Request{Messages: []ai.Message{ai.UserText(prompt)}}, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("incomplete/error stream fails with expected classification", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}), "%v", o.err)
				terminals, done := 0, 0
				for _, event := range o.events {
					if event.Type() == ai.EventError {
						terminals++
					}
					if event.Type() == ai.EventDone {
						done++
					}
				}
				ev.Check("exactly one failure terminal and no retry", terminals == 1 && done == 0 && o.streamErr == o.err && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "terminals=%d done=%d metadata=%+v", terminals, done, o.result.Metadata)
				if err, ok := o.err.(*ai.Error); ok {
					ev.Check("GLM error retains X-Request-Id", err.ProviderRequestID == sc.RequestID[0], "%+v", err)
				}
				if len(o.result.Metadata.Attempts) == 1 {
					a := o.result.Metadata.Attempts[0]
					ev.Check("attempt retains X-Request-Id", a.ProviderRequestID == sc.RequestID[0], "%+v", a)
				}
			})
		}
	}
}
