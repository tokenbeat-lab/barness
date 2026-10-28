//go:build live

package live

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

type minimaxHTTPGate struct {
	next     http.RoundTripper
	caseID   string
	receipts []minimaxlive.Receipt
	active   atomic.Int64
	// A controlled process may synchronize local authorization work without
	// replacing the actual durable Ledger.Attempt below.
	beforeAttempt  func(context.Context)
	thinkingBudget minimaxlive.ThinkingBudget
}

func (g *minimaxHTTPGate) RoundTrip(req *http.Request) (*http.Response, error) {
	b := &minimaxReport.Budget
	if b.Attempts >= 24 || b.Attempts >= b.ReservedAttempts {
		return nil, fmt.Errorf("MiniMax actual attempt budget exhausted")
	}
	if minimaxReport.Kind == minimaxlive.ThinkingKind {
		if minimaxLedger.Scope == nil || b.Attempts >= minimaxLedger.Scope.MaxAttempts || req.Method != "POST" || req.URL.String() != minimaxReport.Identity.Endpoint+"/v1/messages?beta=true" {
			return nil, fmt.Errorf("thinking-on attempt or endpoint differs")
		}
		raw, e := io.ReadAll(io.LimitReader(req.Body, 2049))
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(raw))
		if e != nil || len(raw) > 2048 {
			return nil, fmt.Errorf("thinking-on final wire exceeds reservation")
		}
		g.thinkingBudget, e = minimaxlive.ThinkingWireBudget(raw, *minimaxLedger.Scope)
		if e != nil {
			return nil, e
		}
	}
	if g.beforeAttempt != nil {
		g.beforeAttempt(req.Context())
	}
	r, e := minimaxLedger.Attempt(g.caseID)
	if e != nil {
		return nil, miniMaxBudgetError(e)
	}
	g.receipts = append(g.receipts, r)
	if os.Getenv("BARNESS_MINIMAX_CONTROLLED") == "extra-attempt" {
		if _, err := minimaxLedger.Attempt(g.caseID); err != nil {
			return nil, err
		}
	}
	b.Attempts++
	response, e := g.next.RoundTrip(req)
	if e != nil {
		return response, e
	}
	g.active.Add(1)
	response.Body = &minimaxBody{ReadCloser: response.Body, active: &g.active}
	return response, nil
}

type minimaxBody struct {
	io.ReadCloser
	active *atomic.Int64
	closed atomic.Bool
}

func (b *minimaxBody) Close() error {
	if b.closed.CompareAndSwap(false, true) {
		b.active.Add(-1)
	}
	return b.ReadCloser.Close()
}
