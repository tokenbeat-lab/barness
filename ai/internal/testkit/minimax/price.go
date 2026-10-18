package minimax

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
)

// These immutable standard rates are independent of the synthetic six-model
// mechanism catalog and M Plan entitlement. No runtime exchange-rate service.
//
//go:embed priced-m3.json
var pricedM3 []byte

//go:embed price-sources.json
var priceSources []byte

//go:embed pricing-source.md
var priceSource []byte

const PriceSourcePin = "e695d05b6ed05322433702502b1292b33f2984fa1937d961b06213e87b6fea0d"

// PricedM3Catalog is a Host-owned candidate snapshot; inclusion here does not
// promote BuiltinCatalog or claim live acceptance. Cache-write has no verified
// rate; positive writes must block promotion/certification (ADR-0042).
func PricedM3Catalog() ai.Catalog {
	if err := VerifyPriceSources(); err != nil {
		panic(err)
	}
	var catalog ai.Catalog
	if err := json.Unmarshal(pricedM3, &catalog); err != nil {
		panic(err)
	}
	return catalog
}
func VerifyPriceSources() error {
	hash := func(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }
	if hash(priceSources) != PriceSourcePin {
		return fmt.Errorf("MiniMax price source manifest changed")
	}
	var source struct {
		CatalogSHA256 string                          `json:"catalogSha256"`
		Sources       []struct{ File, SHA256 string } `json:"sources"`
	}
	if json.Unmarshal(priceSources, &source) != nil || source.CatalogSHA256 != hash(pricedM3) || len(source.Sources) != 1 {
		return fmt.Errorf("MiniMax priced snapshot changed")
	}
	files := map[string][]byte{"pricing-source.md": priceSource}
	for _, s := range source.Sources {
		raw, ok := files[s.File]
		if !ok || hash(raw) != s.SHA256 {
			return fmt.Errorf("MiniMax fixed USD price source changed")
		}
	}
	return nil
}
