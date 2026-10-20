package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type budgetScriptTransport struct {
	replies []scriptReply
	calls   int
}

func (s *budgetScriptTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := s.replies[s.calls]
	s.calls++
	req.Body.Close()
	return &http.Response{StatusCode: r.Status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(r.Body)), Request: req}, nil
}

// Keep the official Host/final-wire guard while substituting only the wire
// transport. Controlled CLI's loopback endpoint is intentionally not eligible
// for a paid allowance, so it cannot verify this production boundary.
func TestVisualFinalBudget(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		t.Run(route, func(t *testing.T) {
			ev := unsentEvidence.Case(t, "VQ-final-budget-"+route)
			p, err := load(route)
			if err != nil {
				t.Fatal(err)
			}
			if err = p.useFormalCatalog(); err != nil {
				t.Fatal(err)
			}
			a := arguments{mode: "controlled", route: route, entry: "plain", calls: 12, workMS: 225000, callMS: 30000}
			r := initial(p, a)
			b, err := newBundle(filepath.Join(unsentEvidence.Dir(), "budget-"+route), "synthetic-visual-budget-credential")
			if err != nil || b.inputs(p, a, nil, &r) != nil {
				t.Fatal("controlled input assembly")
			}
			raw, _ := data.ReadFile("testdata/" + route + "-responses.json")
			var replies []scriptReply
			if strict(raw, &replies) != nil || len(replies) != 12 {
				t.Fatal("full twelve native replies required")
			}
			budget := filepath.Join(t.TempDir(), "usd.json")
			if err = os.WriteFile(budget, []byte(`{"schema":1,"round":"provider-sites-final-acceptance-2","maxMicrousd":300000,"reservedMicrousd":0,"attempts":0}`), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("BARNESS_FINAL_ROUND", "provider-sites-final-acceptance-2")
			t.Setenv("BARNESS_FINAL_BUDGET", budget)
			transport := &budgetScriptTransport{replies: replies}
			if err = executeCalls(context.Background(), p, a, b, &r, official(route), "synthetic-visual-budget-credential", "synthetic-account", transport); err != nil {
				t.Fatal(err)
			}
			raw, _ = os.ReadFile(budget)
			var ledger struct {
				Attempts int   `json:"attempts"`
				Reserved int64 `json:"reservedMicrousd"`
			}
			json.Unmarshal(raw, &ledger)
			ev.Fixture("usd.json", raw)
			ev.Record("report", r)
			ev.Check("all twelve official-wire samples reserve before synthetic send", transport.calls == 12 && ledger.Attempts == 12 && ledger.Reserved > 0 && ledger.Reserved <= 300000, "calls=%d ledger=%+v", transport.calls, ledger)
			ev.Check("full controlled quality remains distinct from real quality", r.Status == "COMPLETE" && r.RealQuality == "NOT_RUN" && r.Budget.Calls == 12 && r.Budget.Attempts == 12, "status=%s calls=%d attempts=%d", r.Status, r.Budget.Calls, r.Budget.Attempts)
			if err = b.finish(r); err != nil {
				t.Fatal(err)
			}
			// The normal verifier is a separate keyless process with no paid USD
			// environment. Its loopback-only replay must not reserve new spending.
			t.Setenv("BARNESS_FINAL_BUDGET", "")
			t.Setenv("BARNESS_FINAL_ROUND", "")
			err = verify(b.dir, false)
			ev.Check("independent consumer accepts only controlled proof", err == nil && verify(b.dir, true) != nil, "controlled verification: %v", err)
		})
	}
}
