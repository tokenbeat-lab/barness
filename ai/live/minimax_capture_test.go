//go:build live

package live

import (
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
}

func (g *minimaxHTTPGate) RoundTrip(req *http.Request) (*http.Response, error) {
	b := &minimaxReport.Budget
	if b.Attempts >= 24 || b.Attempts >= b.ReservedAttempts {
		return nil, fmt.Errorf("MiniMax actual attempt budget exhausted")
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
