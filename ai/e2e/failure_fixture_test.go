package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// failureFixture is testdata/responses/failures.json: one logical input and
// the E02 failure scenarios run against it. A scenario whose expectation is
// "partial" starts its reply with the shared partial events, and its expected
// events start with partialEvents.
type failureFixture struct {
	Model         string            `json:"model"`
	SystemPrompt  string            `json:"systemPrompt"`
	User          string            `json:"user"`
	Partial       []json.RawMessage `json:"partial"`
	PartialEvents []string          `json:"partialEvents"`
	Scenarios     []failureScenario `json:"scenarios"`
}

type failureScenario struct {
	ID string `json:"id"`
	// Unreachable points the binding at an endpoint nothing listens on.
	Unreachable bool         `json:"unreachable"`
	Reply       failureReply `json:"reply"`
	// CancelAfterEvents > 0 makes the caller Close the stream once it has read
	// that many events. Such a scenario runs on the Stream entry points only.
	CancelAfterEvents int           `json:"cancelAfterEvents"`
	Expect            failureExpect `json:"expect"`
}

type failureReply struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	// Body is a non-SSE reply's body; Events and Tail (raw bytes appended
	// after the events) make up an SSE reply.
	Body   string            `json:"body"`
	Events []json.RawMessage `json:"events"`
	Tail   string            `json:"tail"`
	End    provider.End      `json:"end"`
}

type failureExpect struct {
	StopReason        ai.StopReason `json:"stopReason"`
	Code              ai.Code       `json:"code"`
	Phase             ai.Phase      `json:"phase"`
	HTTPStatus        int           `json:"httpStatus"`
	ProviderRequestID string        `json:"providerRequestId"`
	RetryAfterMs      int64         `json:"retryAfterMs"`
	ErrorMessage      string        `json:"errorMessage"`
	RawStopReason     string        `json:"rawStopReason"`
	Partial           bool          `json:"partial"`
	Content           []string      `json:"content"`
	ResponseID        string        `json:"responseId"`
	Usage             ai.Usage      `json:"usage"`
	Events            []string      `json:"events"`
	// Requests is the number of inference requests the provider must
	// receive; nil means exactly one.
	Requests *int `json:"requests"`
}

func loadFailureFixture(t *testing.T) (failureFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "failures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f failureFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("failures.json: %v", err)
	}
	return f, raw
}

func (f failureFixture) target() ai.Target { return ai.Target{BindingID: "primary", ModelID: f.Model} }

func (f failureFixture) request() ai.Request {
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User)}}
}

// reply scripts sc's provider response.
func (f failureFixture) reply(t *testing.T, sc failureScenario) provider.Reply {
	t.Helper()
	r := sc.Reply
	if r.Status != 200 {
		return provider.Reply{Status: r.Status, Header: r.Header, Chunks: [][]byte{[]byte(r.Body)}, Script: r.Body, End: r.End}
	}
	var events []json.RawMessage
	if sc.Expect.Partial {
		events = append(events, f.Partial...)
	}
	events = append(events, r.Events...)
	chunks, err := provider.EncodeSSE(events, provider.FramingLF)
	if err != nil {
		t.Fatal(err)
	}
	if r.Tail != "" {
		chunks = append(chunks, []byte(r.Tail))
	}
	reply := provider.SSE(chunks)
	for k, v := range r.Header {
		reply.Header[k] = v
	}
	reply.End = r.End
	return reply
}

// wantEvents is the full expected event summary of sc on a stream.
func (f failureFixture) wantEvents(sc failureScenario) []string {
	var out []string
	if sc.Expect.Partial {
		out = append(out, f.PartialEvents...)
	}
	return append(out, sc.Expect.Events...)
}

// contentSummary renders blocks as "kind:text" plus "|signature" when signed;
// a tool call as "toolCall:<id> <name> <arguments> raw=<raw arguments>".
func contentSummary(content []ai.AssistantContent) []string {
	out := []string{}
	for _, c := range content {
		var s string
		switch c := c.(type) {
		case ai.Text:
			s = "text:" + c.Text
			if c.Signature != "" {
				s += "|" + c.Signature
			}
		case ai.Thinking:
			s = "thinking:" + c.Thinking
			if c.Signature != "" {
				s += "|" + c.Signature
			}
		case ai.ToolCall:
			s = fmt.Sprintf("toolCall:%s %s %s raw=%s", c.ID, c.Name, c.Arguments, c.RawArguments)
		default:
			s = fmt.Sprintf("unexpected:%T", c)
		}
		out = append(out, s)
	}
	return out
}
