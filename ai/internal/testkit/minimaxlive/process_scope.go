package minimaxlive

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
)

func ProcessScope(id Identity, mode, stage, binary string) (livegrant.Scope, error) {
	s := livegrant.Scope{Route: "minimax-" + id.Route, Mode: mode, Stage: stage, BinarySHA256: binary, Provider: "minimax", API: string(id.API), Endpoint: id.Endpoint, Models: []string{id.Model}, MaxRetries: 1, MaxRequestReruns: MaxRequestReruns, MaxCalls: 20, MaxAttempts: 24, MaxOutputTokens: 4096, MaxRequestedOutputTokens: 40000, MaxReservedInputTokens: 120000, CallSeconds: 120, ProcessSeconds: 1800}
	if id.Model != M3 || id.Purpose != "standard" {
		return s, fmt.Errorf("fresh process scope requires international M3 standard")
	}
	cat := CandidateCatalog(id)
	if stage == "formal" {
		cat = ai.BuiltinCatalog()
	}
	if _, ok := cat.Lookup(ai.ProviderMiniMax, id.API, id.Model); !ok {
		return s, fmt.Errorf("selected process catalog lacks model")
	}
	s.CatalogVersion = cat.Version
	s.CatalogHash, _ = cat.Hash()
	for _, c := range Plan(id) {
		s.Scenes = append(s.Scenes, c.ID)
	}
	source, e := ExecutionSource()
	if e != nil {
		return s, e
	}
	raw, _ := json.Marshal(source)
	s.SourceSHA256 = Hash(raw)
	s.SDK, e = SDKVersion(id)
	return s, e
}
