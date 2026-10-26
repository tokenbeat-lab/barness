//go:build live

package live

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
)

var grokHistoryScenario = scenario{id: grokhistory.Scenario, run: func(s *session) {
	if !s.spend() {
		return
	}
	scope := newScope()
	req, err := grokhistory.Request(scope)
	if !s.check("fixed Host history builds", err == nil, "%v", err) {
		return
	}
	s.cs.Record("public-request", req)
	d := drain(s.env.client.Stream(s.ctx, scope, s.target(), req, s.env.combo.full(1024, toolsDefault)), nil)
	ex := s.observe("history", d.res, d.err)
	s.cs.Record("immutable-input", req)
	if !s.ok("history", d.err) {
		return
	}
	s.events("history", d)
	s.turn("history", d.res, ai.StopReasonStop)
	after, _ := json.Marshal(req)
	s.check("fixed synthetic Host source remains immutable", grokhistory.VerifyHost(after), "history changed")
	s.check("two foreign turns downgraded", d.res.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{CrossModel: 2}), "%+v", d.res.Metadata.NativeStateDowngrades)
	s.check("actual destination wire retains collision/reuse pairing", len(ex) == 1 && grokhistory.Verify(s.env.combo.api, []byte(ex[0].RequestBody)) == nil, "actual wire proof")
}}
