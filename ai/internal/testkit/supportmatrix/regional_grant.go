package supportmatrix

import (
	"fmt"
	"regexp"
	"slices"
	"time"
)

// RegionalGrant is a new, explicit one-process authorization. The harness
// atomically consumes the trusted repository reservation before any request;
// historical providers' grants, ledgers and remaining budgets never qualify.
type RegionalGrant struct {
	Mode                     string    `json:"mode"`
	Stage                    string    `json:"stage"`
	Endpoint                 string    `json:"endpoint"`
	SourceSHA256             string    `json:"sourceSha256"`
	CatalogVersion           string    `json:"catalogVersion"`
	CatalogHash              string    `json:"catalogHash"`
	SDK                      string    `json:"sdk"`
	Schema                   int       `json:"schema"`
	ID                       string    `json:"id"`
	Contract                 string    `json:"contract"`
	Combo                    string    `json:"combo"`
	Provider                 string    `json:"provider"`
	Purpose                  string    `json:"purpose"`
	AccountAlias             string    `json:"accountAlias"`
	RegionAlias              string    `json:"regionAlias"`
	Models                   []string  `json:"models"`
	ExpiresAt                time.Time `json:"expiresAt"`
	MaxCalls                 int       `json:"maxCalls"`
	MaxAttempts              int       `json:"maxAttempts"`
	MaxRequestedOutputTokens int       `json:"maxRequestedOutputTokens"`
	MaxReservedInputTokens   int       `json:"maxReservedInputTokens"`
	MaxUSD                   float64   `json:"maxUSD,omitempty"`
	BinarySHA256             string    `json:"binarySha256,omitempty"`
}

func (g RegionalGrant) Check(combo string, at time.Time) error {
	want, err := RegionalGrantSnapshot(combo, g.Stage)
	r, ok := RegionalRouteOf(combo)
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	if !slices.Contains([]string{"live", "controlled"}, g.Mode) || err != nil || g.Endpoint != want.Endpoint || g.SourceSHA256 != want.SourceSHA256 || g.CatalogVersion != want.CatalogVersion || g.CatalogHash != want.CatalogHash || g.SDK != want.SDK || !ok || g.Schema != 1 || !alias.MatchString(g.ID) || g.Contract != RegionalAuthorization(combo) || g.Combo != combo || g.Provider != r.Provider || g.Purpose != "standard" || !alias.MatchString(g.AccountAlias) || !alias.MatchString(g.RegionAlias) || !slices.Equal(g.Models, RegionalModels(combo)) || !g.ExpiresAt.After(at) || g.ExpiresAt.Sub(at) > 24*time.Hour || g.MaxCalls != want.MaxCalls || g.MaxAttempts != want.MaxAttempts || g.MaxRequestedOutputTokens != want.MaxRequestedOutputTokens || g.MaxReservedInputTokens != want.MaxReservedInputTokens || RegionalIsKimi(combo) && (!regionalValidUSD(g.MaxUSD) || len(g.BinarySHA256) != 64) {
		return fmt.Errorf("regional explicit own identity, account, purpose, expiry or budget authorization invalid")
	}
	return nil
}

// RegionalGrantSnapshot fills only the current reviewed machine facts. The
// aliases, new grant ID and near-term expiry require explicit human scope.
func RegionalGrantSnapshot(combo, stage string) (RegionalGrant, error) {
	r, ok := RegionalRouteOf(combo)
	if !ok || stage != "candidate" && stage != "formal" {
		return RegionalGrant{}, fmt.Errorf("unknown regional grant target or stage")
	}
	cat, err := RegionalCandidateCatalog(combo)
	if stage == "formal" {
		cat, err = RegionalFormalCatalog(combo)
	}
	if err != nil {
		return RegionalGrant{}, err
	}
	hash, err := cat.Hash()
	if err != nil {
		return RegionalGrant{}, err
	}
	source, err := RegionalExecutionSource()
	if err != nil {
		return RegionalGrant{}, err
	}
	sdk, err := RegionalSDK(combo)
	if err != nil {
		return RegionalGrant{}, err
	}
	limits := RegionalBudgetFor(combo)
	return RegionalGrant{Schema: 1, Mode: "live", Stage: stage, Endpoint: r.Endpoint, SourceSHA256: RegionalSourceHash(source), CatalogVersion: cat.Version, CatalogHash: hash, SDK: sdk, Contract: RegionalAuthorization(combo), Combo: combo, Provider: r.Provider, Purpose: "standard", Models: RegionalModels(combo), MaxCalls: limits.MaxCalls, MaxAttempts: limits.MaxAttempts, MaxRequestedOutputTokens: limits.MaxRequestedOutputTokens, MaxReservedInputTokens: limits.MaxReservedInputTokens}, nil
}
