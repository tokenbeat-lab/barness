package chineseeval

import (
	_ "embed"
	"errors"

	"github.com/tokenbeat-lab/barness/ai"
)

// This immutable historical snapshot was exported from the report's original
// commit a12c037356751301f292f770188bd44bb629dabf, without rewriting its pins.
// Historical verification says nothing about current release applicability.
//
//go:embed testdata/recorded-jev-20261008.json
var recordedJev []byte

func LoadRecorded(data, config, snapshot []byte) (*Plan, error) {
	if len(data) > 1<<20 || len(config) > 16384 {
		return nil, errors.New("evaluation: oversized recorded source")
	}
	var c Config
	if decode(config, &c) != nil {
		return nil, errors.New("evaluation: invalid recorded config")
	}
	var catalog ai.Catalog
	if len(snapshot) > 0 {
		if len(snapshot) > 2<<20 || decode(snapshot, &catalog) != nil {
			return nil, errors.New("evaluation: invalid recorded catalog")
		}
	} else {
		// Original schema 1 bundles predate the catalog file. Resolve only their
		// exact archived content-addressed snapshot, never a rewritten config/hash.
		if c.Route != nil || decode(recordedJev, &catalog) != nil {
			return nil, errors.New("evaluation: catalog snapshot missing")
		}
	}
	return Load(data, config, catalog)
}
