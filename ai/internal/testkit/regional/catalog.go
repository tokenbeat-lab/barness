// Package regional contains independent source-backed qualification candidates.
// Neither candidate availability nor offline success promotes BuiltinCatalog.
package regional

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
)

// Each literal candidate owns its metadata and references its own market's
// sources. Repeated facts are deliberate: these identities evolve independently
// and never inherit sibling provider/protocol models at runtime.
//
//go:embed *.json sources/*
var files embed.FS

const snapshotPin = "1e244e76976d9eb2814c4c3f2f1b6dbd725baf895f16a1eceb9466681bf07a41"

// Price keeps missing token prices distinct from zero and freezes native
// currency/tier/FX facts beside the USD model estimate. Storage promotions do
// not establish a token-write rate. Positive unknown writes must fail pricing.
type Price struct {
	Currency    string      `json:"currency"`
	Unit        string      `json:"unit"`
	ServiceTier string      `json:"serviceTier"`
	Input       float64     `json:"input"`
	Output      float64     `json:"output"`
	CacheRead   float64     `json:"cacheRead"`
	CacheWrite  *float64    `json:"cacheWrite"`
	Tiers       []PriceTier `json:"tiers,omitempty"`
	Source      string      `json:"source"`
	FX          *FX         `json:"fx"`
	Note        string      `json:"note"`
}

type PriceTier struct {
	InputTokensAbove int64    `json:"inputTokensAbove"`
	Input            float64  `json:"input"`
	Output           float64  `json:"output"`
	CacheRead        float64  `json:"cacheRead"`
	CacheWrite       *float64 `json:"cacheWrite"`
}

type FX struct {
	Source    string  `json:"source"`
	Date      string  `json:"date"`
	Direction string  `json:"direction"`
	Rate      float64 `json:"rate"`
	Formula   string  `json:"formula"`
}

// Candidate is one full identity's target and current qualification state.
// INCOMPLETE is retained until its own live and protocol evidence exists.
type Candidate struct {
	Model   ai.Model `json:"model"`
	Status  string   `json:"status"`
	OwnLive string   `json:"ownLive"`
	Missing []string `json:"missing"`
	Sources []string `json:"sources"`
	Price   Price    `json:"price"`
}

type candidateFile struct {
	Version    string      `json:"version"`
	Candidates []Candidate `json:"candidates"`
}

// Candidates returns fresh independent records, including explicit incomplete
// price/support facts. Raw source originals are gzip-compressed losslessly;
// source-files.json pins both compressed and original bytes and retrieval time.
func Candidates() []Candidate {
	_, candidates, err := load()
	if err != nil {
		panic(err)
	}
	return candidates
}

// CandidateCatalog supplies all nine new identities for explicitly authorized
// qualification. Known USD input/output/cache-read rates are usable; a zero
// write placeholder is never evidence that positive cache writes are free.
// Callers must retain Candidates' price/status metadata with their artifacts.
func CandidateCatalog() (ai.Catalog, error) {
	version, candidates, err := load()
	if err != nil {
		return ai.Catalog{}, err
	}
	cat := ai.Catalog{Version: version}
	for _, c := range candidates {
		cat.Models = append(cat.Models, c.Model)
	}
	if _, err := cat.Hash(); err != nil {
		return ai.Catalog{}, err
	}
	return cat, nil
}

func load() (string, []Candidate, error) {
	if err := VerifySources(); err != nil {
		return "", nil, err
	}
	var version string
	var candidates []Candidate
	for _, name := range []string{"zai-candidates.json", "minimax-cn-candidates.json"} {
		raw, err := files.ReadFile(name)
		if err != nil {
			return "", nil, err
		}
		var part candidateFile
		if err := json.Unmarshal(raw, &part); err != nil {
			return "", nil, err
		}
		if version != "" && part.Version != version {
			return "", nil, errors.New("regional: inconsistent candidate versions")
		}
		version = part.Version
		candidates = append(candidates, part.Candidates...)
	}
	return version, candidates, nil
}

func hash(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }
