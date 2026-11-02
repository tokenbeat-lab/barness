package ai

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
)

// Catalog is a versioned model catalog and price snapshot. It is never
// updated online. Version names a content: a catalog whose models or prices
// change gets a new Version, and Hash tells two contents apart even when a
// host reuses one.
type Catalog struct {
	Version          string            `json:"version"`
	Models           []Model           `json:"models"`
	ImageModels      []ImageModel      `json:"imageModels,omitempty"`
	ClassifierModels []ClassifierModel `json:"classifierModels,omitempty"`
}

//go:generate go run ../cmd/cataloggen --check --out catalog_builtin.json

// builtinCatalogJSON is generated from catalog/data/providers. The ai package
// consumes only runtime definitions and has no dependency on maintenance.
//
//go:embed catalog_builtin.json
var builtinCatalogJSON []byte

// BuiltinCatalog returns a fresh copy of the generated default catalog.
// The Provider maintenance source retains each definition's provenance and
// human decisions; historical measurements and catalog pins remain immutable.
func BuiltinCatalog() Catalog {
	var snapshot struct {
		Hash    string  `json:"hash"`
		Catalog Catalog `json:"catalog"`
	}
	d := json.NewDecoder(bytes.NewReader(builtinCatalogJSON))
	d.DisallowUnknownFields()
	if d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF {
		panic("invalid generated builtin catalog JSON")
	}
	hash, err := snapshot.Catalog.Hash()
	if err != nil || hash != snapshot.Hash {
		panic("invalid generated builtin catalog content")
	}
	return snapshot.Catalog
}
