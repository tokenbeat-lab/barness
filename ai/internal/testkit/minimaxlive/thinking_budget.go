package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
)

func verifyThinkingBudget(dir string, r Report) error {
	if len(r.Calls) != 1 || r.Calls[0].ID != "thinking-on" || r.Calls[0].Rerun != 0 || r.Calls[0].Failure != "" || r.Budget != (Budget{1, 1, 1, 2048, 2048}) {
		return fmt.Errorf("MiniMax thinking-on finite budget differs")
	}
	g, _, _, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return e
	}
	base := filepath.Join(dir, r.Calls[0].CaseID)
	wire, e := ReadFile(filepath.Join(base, "request.json"), 2048)
	if e != nil {
		return e
	}
	want, e := ThinkingWireBudget(wire, g.Scope)
	if e != nil {
		return e
	}
	var got ThinkingBudget
	raw, e := ReadFile(filepath.Join(base, "thinking-budget.json"), 4096)
	if e != nil || StrictJSON(raw, &got) != nil || !reflect.DeepEqual(got, want) {
		return fmt.Errorf("MiniMax thinking-on known-rate reservation proof differs")
	}
	if e = verifyThinkingResources(dir, r.Calls[0]); e != nil {
		return e
	}
	var result Capture
	raw, e = ReadFile(filepath.Join(base, "result.json"), 2<<20)
	if e != nil || json.Unmarshal(raw, &result) != nil || len(result.Result.Metadata.Attempts) != 1 {
		return fmt.Errorf("MiniMax single known usage missing")
	}
	a := result.Result.Metadata.Attempts[0]
	if _, e := AssessThinkingBilling(dir, r); e != nil {
		return e
	}
	var final map[string]json.RawMessage
	if len(r.Calls[0].UsagePresence) < 2 || json.Unmarshal(r.Calls[0].UsagePresence[len(r.Calls[0].UsagePresence)-1], &final) != nil {
		return fmt.Errorf("MiniMax final output usage missing")
	}
	var output int64
	if v, ok := final["output_tokens"]; !ok || string(v) == "null" || json.Unmarshal(v, &output) != nil || output < 0 {
		return fmt.Errorf("MiniMax final output counter unreported")
	}
	if a.Usage.Output > 2048 || a.Usage.CacheWrite > 0 || a.Usage.Cost.Total > want.ReservedUSD {
		return fmt.Errorf("MiniMax single known usage exceeds reserved USD or has unknown write price")
	}
	return nil
}

func verifyThinkingResources(dir string, c Call) error {
	var metrics struct {
		ActiveBodies int
		HookCalls    int
		Observer     struct{ Dropped, Failed int }
	}
	raw, e := ReadFile(filepath.Join(dir, c.CaseID, "metrics.json"), 4096)
	hooks := 1
	if c.ID == "preflight-error" {
		hooks = 0
	}
	if e != nil || json.Unmarshal(raw, &metrics) != nil || metrics.ActiveBodies != 0 || metrics.Observer.Dropped != 0 || metrics.Observer.Failed != 0 || metrics.HookCalls != hooks {
		return fmt.Errorf("MiniMax captured resource settlement differs: %s", c.ID)
	}
	return nil
}
