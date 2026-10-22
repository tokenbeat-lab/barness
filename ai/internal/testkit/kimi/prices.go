package kimi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Price preserves original-market facts independently for every protocol.
// FX is historical and fixed; the retained originals, not online conversion,
// explain the model's USD rates. Candidate availability grants no qualification.
type Price struct {
	Provider     string  `json:"provider"`
	API          string  `json:"api"`
	Model        string  `json:"model"`
	Currency     string  `json:"currency"`
	Unit         string  `json:"unit"`
	Input        float64 `json:"input"`
	Output       float64 `json:"output"`
	CacheRead    float64 `json:"cacheRead"`
	CacheWrite5m float64 `json:"cacheWrite5m"`
	CacheWrite1h float64 `json:"cacheWrite1h"`
	SourceID     string  `json:"sourceID"`
	SourceSHA256 string  `json:"sourceSHA256"`
	Extraction   string  `json:"extraction"`
	FXID         string  `json:"fxID"`
	FXUSDCNY     float64 `json:"fxUSDCNY"`
	Conversion   string  `json:"conversion"`
}

// Prices returns a fresh snapshot of native prices and conversion provenance.
func Prices() []Price {
	var snapshot struct {
		Prices []Price `json:"prices"`
	}
	raw, _ := files.ReadFile("prices.json")
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		panic(err)
	}
	return snapshot.Prices
}

// SnapshotHash binds candidate metadata, native-price extraction and all
// original-source hashes; provenance changes require a versioned new snapshot.
func SnapshotHash() string {
	sum := sha256.New()
	for _, name := range []string{"candidate.json", "prices.json", "source-files.json"} {
		raw, err := files.ReadFile(name)
		if err != nil {
			panic(err)
		}
		_, _ = sum.Write([]byte(name + "\n"))
		_, _ = sum.Write(raw)
	}
	return fmt.Sprintf("%x", sum.Sum(nil))
}
