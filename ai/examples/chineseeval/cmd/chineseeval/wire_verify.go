package main

import (
	"encoding/json"
	"errors"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

func verifyWire(p *chineseeval.Plan, r chineseeval.Report, raw []byte) error {
	bad := errors.New("actual capture mismatch")
	var records []wireCapture
	if json.Unmarshal(raw, &records) != nil || len(records) != r.Budget.Attempts {
		return bad
	}
	scripts := []fixtureCapture{}
	next := 0
	for _, s := range r.Samples {
		if len(s.Metadata.Attempts) == 0 {
			continue
		}
		c := records[next]
		next++
		a := s.Metadata.Attempts[0]
		if c.Source != r.Kind || c.Status != a.HTTPStatus || c.ProviderRequestID != a.ProviderRequestID || len(c.Request) > 16384 || len(c.Response) > 16385 || (c.Status == 0) != c.NetworkFailure {
			return bad
		}
		if (c.Status < 200 || c.Status >= 300) && len(c.Response) > 4097 {
			return bad
		}
		if c.NetworkFailure && (c.Response != "" || c.ProviderRequestID != "" || c.ReadFailure) {
			return bad
		}
		interrupted := s.Code == ai.CodeCanceled || s.Code == ai.CodeDeadlineExceeded
		if interrupted && s.Answer != nil {
			return bad
		}
		scripts = append(scripts, fixtureCapture{Path: c.Path, Request: c.Request, Response: json.RawMessage(c.Response), Status: c.Status, NetworkFailure: c.NetworkFailure, ReadFailure: c.ReadFailure, Canceled: false})
	}
	// Reuse public-client replay for complete bytes. Interruption timing cannot
	// be recreated from a prefix; its status/phase/usage remains reconciled with
	// the recorded Attempt and Observer, without inventing a complete response.
	return verifyCapturedScripts(p, r, scripts)
}
