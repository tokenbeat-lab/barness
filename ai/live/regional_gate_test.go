//go:build live

package live

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type regionalHTTPGate struct {
	next     http.RoundTripper
	process  *regionalProcess
	attempts atomic.Int64
	bodies   atomic.Int64
}

func (g *regionalHTTPGate) RoundTrip(req *http.Request) (*http.Response, error) {
	original := req.Body
	defer original.Close()
	raw, err := io.ReadAll(io.LimitReader(original, (256<<10)+1))
	if err != nil || len(raw) > 256<<10 {
		return nil, fmt.Errorf("regional encoded request capture bound exceeded")
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(raw))
	input := max(2048, len(raw)+1024)
	if g.process.report.Budget.ReservedInputTokens+input > 120000 {
		return nil, fmt.Errorf("regional encoded input reservation exhausted")
	}
	g.process.report.Budget.ReservedInputTokens += input
	if g.attempts.Add(1) > 24 {
		return nil, fmt.Errorf("regional own attempt budget exhausted")
	}
	res, err := g.next.RoundTrip(req)
	if res != nil && res.Body != nil {
		g.bodies.Add(1)
		res.Body = &regionalTrackedBody{ReadCloser: res.Body, gate: g}
	}
	return res, err
}

type regionalTrackedBody struct {
	io.ReadCloser
	gate *regionalHTTPGate
	once sync.Once
}

func (b *regionalTrackedBody) Close() error {
	b.once.Do(func() { b.gate.bodies.Add(-1) })
	return b.ReadCloser.Close()
}

func (p *regionalProcess) reserve(ctx context.Context, outputBudget int) error {
	b := &p.report.Budget
	if ctx.Err() != nil || b.CallsUsed >= 24 || b.RequestedOutputTokens+outputBudget > 98304 {
		return fmt.Errorf("regional own call/token/time reservation exhausted")
	}
	b.CallsUsed++
	b.AttemptsReserved++
	b.RequestedOutputTokens += outputBudget
	return nil
}
