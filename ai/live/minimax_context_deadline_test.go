//go:build live

package live

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// The agreed seam is the actual harness deadline factory driving the public
// Client, synthetic Ledger and controlled HTTP boundary, retaining the bounded
// process tail and the earlier test deadline.
func TestMiniMaxHarnessContextDeadline(t *testing.T) {
	oldReport, oldLedger := minimaxReport, minimaxLedger
	t.Cleanup(func() { minimaxReport, minimaxLedger = oldReport, oldLedger })
	id, e := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name, mode, stage, kind string
		tail                    time.Duration
		testTail                time.Duration
	}{
		{"process", "live", "formal", "minimax-formal", 20 * time.Second, 0},
		{"test-deadline", "controlled", "candidate", "minimax-candidate-only", 20 * time.Second, time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cs := run.Case(t, "MM07-context-deadline-"+tc.name)
			now := time.Now()
			end := now.Add(tc.tail + 40*time.Millisecond)
			r := &minimaxlive.Report{Mode: tc.mode, Stage: tc.stage, Kind: tc.kind, Identity: id, StartedAt: end.Add(-3600 * time.Second), Deadline: end, Budget: minimaxlive.Budget{ReservedAttempts: 1}}
			testDeadline := time.Time{}
			want := now.Add(40 * time.Millisecond)
			if tc.testTail > 0 {
				r.Deadline = now.Add(3600 * time.Second)
				r.StartedAt = now
				testDeadline = now.Add(tc.testTail + 40*time.Millisecond)
			}
			deadline := miniMaxContextDeadline(r, testDeadline)
			cs.Record("bounds", map[string]any{"reportStartedAt": r.StartedAt, "reportDeadline": r.Deadline, "testDeadline": testDeadline, "actualContextDeadline": deadline, "expectedContextDeadline": want, "expectedTailSeconds": tc.tail.Seconds()})
			if !cs.Check("actual harness deadline reserves its fixed tail and takes the test minimum", deadline.Equal(want), "got=%s want=%s", deadline, want) {
				return
			}
			minimaxReport = r
			minimaxLedger = minimaxlive.Ledger{Path: filepath.Join(t.TempDir(), "stage.jsonl"), Stage: "candidate", Session: tc.name, Identity: id}
			if _, e := minimaxLedger.Reserve("full-stream", 1024, 2048); e != nil {
				t.Fatal(e)
			}
			var arrivals atomic.Int64
			var bodyClosed atomic.Bool
			gate := &minimaxHTTPGate{caseID: "full-stream", next: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				req.Body.Close()
				arrivals.Add(1)
				body := &miniMaxDeadlineBody{ctx: req.Context(), first: strings.NewReader("data: {\"id\":\"synthetic\",\"model\":\"MiniMax-M3\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"42\"},\"finish_reason\":null}]}\n\n"), closed: &bodyClosed}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: body, Request: req}, nil
			})}
			rec := newRecorder()
			rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = gate
			selected := &combo{name: "minimax-chat", operation: ai.OperationChat, provider: ai.ProviderMiniMax, api: id.API, endpoint: id.Endpoint, model: id.Model, sdk: openAISDK}
			client, e := newClientWithPolicy(selected, "synthetic-deadline-client", rec, minimaxlive.CandidateCatalog(id), miniMaxPolicy())
			if e != nil {
				t.Fatal(e)
			}
			ctx, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()
			stream := client.Stream(ctx, newScope(), ai.Target{BindingID: selected.name, ModelID: id.Model}, ai.Request{Messages: []ai.Message{ai.UserText("Reply 42.")}}, ai.ChatOptions{MaxTokens: ai.Value(1024)})
			for stream.Next() {
			}
			stream.Close()
			cs.Check("context cancellation reaches the public stream", ctx.Err() == context.DeadlineExceeded && stream.Err() != nil, "ctx=%v stream=%v", ctx.Err(), stream.Err())
			cs.Check("single attempt and all body/observer resources settle", arrivals.Load() == 1 && len(gate.receipts) == 1 && gate.active.Load() == 0 && bodyClosed.Load() && client.ObserverStats().Failed == 0 && client.ObserverStats().Dropped == 0, "arrivals=%d receipts=%d active=%d bodyClosed=%v observer=%+v", arrivals.Load(), len(gate.receipts), gate.active.Load(), bodyClosed.Load(), client.ObserverStats())
		})
	}
}

type miniMaxDeadlineBody struct {
	ctx    context.Context
	first  io.Reader
	closed *atomic.Bool
}

func (b *miniMaxDeadlineBody) Read(p []byte) (int, error) {
	n, e := b.first.Read(p)
	if n > 0 || e != io.EOF {
		return n, e
	}
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *miniMaxDeadlineBody) Close() error { b.closed.Store(true); return nil }
