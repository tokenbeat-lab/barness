//go:build live

package live

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Issue 17: successful and failed chat consumption must retain the terminal
// metadata Observer record just as the unary routes do. Use the public Client
// and the existing controlled transport, never a synthetic live PASS report.
func TestChatConsumptionEvidence(t *testing.T) {
	for _, status := range []int{200, 401} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			c := &combos[slices.IndexFunc(combos, func(c combo) bool { return c.name == "openai-responses" })]
			rec := newRecorder()
			defer rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body := "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"synthetic\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n"
				if status != 200 {
					body = `{"error":{"message":"synthetic denial"}}`
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})
			client, err := newClient(c, "key:chat-test", rec, ai.BuiltinCatalog())
			if err != nil {
				t.Fatal(err)
			}
			cs := run.Case(t, "EVIDENCE-chat-"+http.StatusText(status))
			s := &session{t: t, ctx: context.Background(), cs: cs, env: &liveEnv{combo: c, client: client, rec: rec}, model: c.model, attempt: 1}
			res, err := client.Complete(s.ctx, newScope(), s.target(), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic fixture")}}, c.full(32, toolsDefault))
			cs.Check("controlled public outcome", (err == nil) == (status == 200), "status %d error %v", status, err)
			s.observe("controlled", res, err)
			raw, err := os.ReadFile(filepath.Join(run.Dir(), cs.ID(), "observations-attempt-1-controlled.json"))
			if err != nil {
				t.Fatal("consumed chat has no Observer artifact:", err)
			}
			var observations []ai.Observation
			if err := json.Unmarshal(raw, &observations); err != nil {
				t.Fatal(err)
			}
			cs.Check("terminal belongs to consumed request", len(observations) > 0 && observations[len(observations)-1].Kind == ai.ObservationCallFinished && observations[len(observations)-1].Call.RequestID == res.Metadata.RequestID, "terminal absent or wrong call")
		})
	}
}
