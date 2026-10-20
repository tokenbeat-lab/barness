package ai

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
)

// The independent regional payload remains inactive until all six own live
// candidates pass. Activation changes data only; production never reads test
// evidence, source candidates, credentials or another region's model records.
//
//go:embed catalog_regional.json
var regionalCatalogData []byte

func builtinRegionalCatalog() Catalog {
	var payload struct {
		Schema  int     `json:"schema"`
		Active  bool    `json:"active"`
		Version string  `json:"version"`
		Models  []Model `json:"models"`
	}
	decoder := json.NewDecoder(bytes.NewReader(regionalCatalogData))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(new(any)) != io.EOF || payload.Schema != 1 {
		panic("invalid embedded regional catalog activation")
	}
	if !payload.Active {
		if payload.Version != "" || len(payload.Models) != 0 {
			panic("inactive regional catalog contains models or version")
		}
		return Catalog{}
	}
	if payload.Version != "2026-10-20.1-regional" || len(payload.Models) != 9 {
		panic("regional activation requires nine own identities and version")
	}
	cat := Catalog{Version: payload.Version, Models: payload.Models}
	if _, err := cat.Hash(); err != nil {
		panic("invalid embedded regional catalog models")
	}
	for _, api := range []API{APIOpenAICompletions, APIOpenAIResponses, APIAnthropicMessages} {
		for _, identity := range []struct {
			provider ProviderID
			id       string
		}{{ProviderZAI, "glm-5.3"}, {ProviderZAI, "glm-5.2"}, {ProviderMiniMaxCN, "MiniMax-M3"}} {
			if _, ok := cat.Lookup(identity.provider, api, identity.id); !ok {
				panic("regional activation identity missing or substituted")
			}
		}
	}
	return cat
}
