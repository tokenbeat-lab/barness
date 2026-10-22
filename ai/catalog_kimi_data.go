package ai

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
)

// Kimi activation is a separate reviewed data payload. Availability of a
// candidate or a controlled capture never puts any Kimi model in production.
//
//go:embed catalog_kimi.json
var kimiCatalogData []byte

func builtinKimiCatalog() Catalog {
	var payload struct {
		Schema           int     `json:"schema"`
		Active           bool    `json:"active"`
		Version          string  `json:"version"`
		ProvenanceSHA256 string  `json:"provenanceSha256"`
		Models           []Model `json:"models"`
	}
	d := json.NewDecoder(bytes.NewReader(kimiCatalogData))
	d.DisallowUnknownFields()
	if d.Decode(&payload) != nil || d.Decode(new(any)) != io.EOF || payload.Schema != 1 {
		panic("invalid embedded Kimi activation")
	}
	if !payload.Active {
		if payload.Version != "" || payload.ProvenanceSHA256 != "" || len(payload.Models) != 0 {
			panic("inactive Kimi catalog contains promotion data")
		}
		return Catalog{}
	}
	if payload.Version != "kimi-k3-formal-2026-10-22.1" || payload.ProvenanceSHA256 != "3a615910c7e849dbc9d11e4ea7cf39b4759b9dc6e7fdd64402ba716f1fa7420d" || len(payload.Models) != 6 {
		panic("Kimi activation differs from reviewed source-backed six identities")
	}
	cat := Catalog{Version: payload.Version, Models: payload.Models}
	if _, err := cat.Hash(); err != nil {
		panic("invalid embedded Kimi models")
	}
	for _, provider := range []ProviderID{ProviderMoonshot, ProviderMoonshotCN} {
		for _, api := range []API{APIOpenAICompletions, APIOpenAIResponses, APIAnthropicMessages} {
			if _, ok := cat.Lookup(provider, api, "kimi-k3"); !ok {
				panic("Kimi activation identity missing or substituted")
			}
		}
	}
	return cat
}
