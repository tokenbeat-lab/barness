//go:build live

package live

import (
	"context"
	"errors"
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

// Failure ways were recorded before this public regression: local authorization
// charged as connect/header time, cancelled downstream sends, lost receipts,
// skipped authority, widened retries, and unsettled body/Observer resources.
func TestMiniMaxNativeAuthorizationDeadline(t *testing.T) {
	cs := run.Case(t, "MM07-native-authorization-deadline")
	oldReport, oldLedger := minimaxReport, minimaxLedger
	t.Cleanup(func() { minimaxReport, minimaxLedger = oldReport, oldLedger })
	id, err := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	if err != nil {
		t.Fatal(err)
	}
	minimaxReport = &minimaxlive.Report{Identity: id, Budget: minimaxlive.Budget{ReservedAttempts: 1}}
	minimaxLedger = minimaxlive.Ledger{Path: filepath.Join(t.TempDir(), "stage.jsonl"), Stage: "candidate", Session: "authorization-deadline", Identity: id}
	if _, err = minimaxLedger.Reserve("full-stream", 1024, 2048); err != nil {
		t.Fatal(err)
	}
	var vendorArrivals atomic.Int64
	peer := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		// A real http.Transport also refuses an already cancelled request before
		// opening a connection. This controlled peer has no network path.
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		vendorArrivals.Add(1)
		body := "data: {\"id\":\"synthetic\",\"model\":\"MiniMax-M3\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"42\"},\"finish_reason\":null}]}\n\n" +
			"data: {\"id\":\"synthetic\",\"model\":\"MiniMax-M3\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":4,\"completion_tokens\":1,\"total_tokens\":5}}\n\n" + "data: [DONE]\n\n"
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	authorizationStarted, authorizationFinished := time.Time{}, time.Time{}
	delay := 75 * time.Millisecond
	gate := &minimaxHTTPGate{next: peer, caseID: "full-stream", beforeAttempt: func(context.Context) {
		authorizationStarted = time.Now()
		// Deterministic bounded stand-in for descriptor walking and hashes. The
		// real ledger authorization/receipt remains immediately after this stage.
		time.Sleep(delay)
		authorizationFinished = time.Now()
	}}
	rec := newRecorder()
	rec.next.(*http.Transport).CloseIdleConnections()
	rec.next = gate
	policy := miniMaxPolicy()
	// Scale the actual Host factory's durations together. No independent
	// connect override can hide a regression in the real caller's policy.
	call := policy.CallTimeout
	scale := func(d time.Duration) time.Duration { return d / (call / (160 * time.Millisecond)) }
	policy.CallTimeout = scale(policy.CallTimeout)
	policy.AdmissionWait = scale(policy.AdmissionWait)
	policy.ConnectTimeout = scale(policy.ConnectTimeout)
	policy.ResponseHeaderTimeout = scale(policy.ResponseHeaderTimeout)
	policy.ReadIdleTimeout = scale(policy.ReadIdleTimeout)
	selected := &combo{name: "minimax-chat", operation: ai.OperationChat, provider: ai.ProviderMiniMax, api: id.API, endpoint: id.Endpoint, model: id.Model, sdk: openAISDK}
	client, err := newClientWithPolicy(selected, "synthetic-authority-deadline", rec, minimaxlive.CandidateCatalog(id), policy)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream := client.Stream(ctx, newScope(), ai.Target{BindingID: selected.name, ModelID: id.Model}, ai.Request{Messages: []ai.Message{ai.UserText("Reply 42.")}}, ai.ChatOptions{MaxTokens: ai.Value(1024)})
	for stream.Next() {
	}
	result, err := stream.Result()
	stream.Close()
	var failure *ai.Error
	cs.Record("boundary", map[string]any{"authorizationMS": authorizationFinished.Sub(authorizationStarted).Milliseconds(), "vendorArrivals": vendorArrivals.Load(), "attemptReceipts": len(gate.receipts), "activeBodies": gate.active.Load(), "publicError": err, "attempts": result.Metadata.Attempts})
	if err != nil {
		cs.Check("exact original symptom", errors.As(err, &failure) && failure.Code == ai.CodeDeadlineExceeded && failure.Phase == ai.PhaseRequest && strings.Contains(failure.Message, "ConnectTimeout (20ms)"), "error=%v", err)
		cs.Check("vendor was never reached", vendorArrivals.Load() == 0, "vendor arrivals=%d", vendorArrivals.Load())
	}
	cs.Check("local native authorization does not consume vendor connect deadline", err == nil && vendorArrivals.Load() == 1 && result.Message.StopReason == ai.StopReasonStop, "error=%v vendor arrivals=%d stop=%s", err, vendorArrivals.Load(), result.Message.StopReason)
	cs.Check("actual durable authorization remains once only", len(gate.receipts) == 1 && minimaxReport.Budget.Attempts == 1, "receipts=%d attempts=%d", len(gate.receipts), minimaxReport.Budget.Attempts)
	cs.Check("body and observer settle", gate.active.Load() == 0 && client.ObserverStats().Failed == 0 && client.ObserverStats().Dropped == 0, "active=%d observer=%+v", gate.active.Load(), client.ObserverStats())
}
