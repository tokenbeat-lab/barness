package ai

import (
	_ "embed"
	"encoding/json"
)

// Activation is generated only after three own actual M3 candidate processes
// pass. Keeping this data separate permits a reviewed data-only promotion after
// runtime source freeze; production reads no evidence directory or credential.
//
//go:embed catalog_minimax.json
var miniMaxCatalogData []byte

func builtinMiniMaxCatalog() Catalog {
	var c Catalog
	if err := json.Unmarshal(miniMaxCatalogData, &c); err != nil {
		panic("invalid embedded MiniMax catalog activation")
	}
	if len(c.Models) != 0 && len(c.Models) != 3 {
		panic("MiniMax activation requires three own M3 identities")
	}
	for _, m := range c.Models {
		if m.Provider != ProviderMiniMaxCN || m.ID != "MiniMax-M3" {
			panic("invalid embedded MiniMax formal identity")
		}
	}
	return c
}
