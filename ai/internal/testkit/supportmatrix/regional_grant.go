package supportmatrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
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
	CallIntervalMs           int       `json:"callIntervalMs,omitempty"`
	ScenarioIDs              []string  `json:"scenarioIds,omitempty"`
	BaselineManifestSHA256   string    `json:"baselineManifestSha256,omitempty"`
}

func (g RegionalGrant) Check(combo string, at time.Time) error {
	want, err := RegionalGrantSnapshot(combo, g.Stage)
	if g.Schema == 2 {
		want, err = KimiHistoryGrantSnapshot(combo, g.BaselineManifestSHA256)
	}
	return g.check(combo, at, want, err)
}

// checkMeasurement validates historical authority at execution time, without
// granting new execution or substituting current source/catalog into originals.
func (g RegionalGrant) checkMeasurement(r Report, at time.Time) error {
	return g.checkSelectedMeasurement(r, at, regionalFull)
}
func (g RegionalGrant) check(combo string, at time.Time, want RegionalGrant, err error) error {
	r, ok := RegionalRouteOf(combo)
	if !slices.Equal(g.ScenarioIDs, want.ScenarioIDs) || g.BaselineManifestSHA256 != want.BaselineManifestSHA256 {
		return refuse("regional selected evidence scope differs")
	}
	if g.Schema == 2 && (g.Stage != "candidate" || g.Purpose != KimiHistoryPurpose || g.MaxUSD != 0.10 || g.Mode == "live" && g.CallIntervalMs != 60000) {
		return refuse("Kimi supplement spend, stage or pacing differs")
	}
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	if g.CallIntervalMs < 0 || g.CallIntervalMs > 60000 || g.CallIntervalMs != 0 && !RegionalIsKimi(combo) || !slices.Contains([]string{"live", "controlled"}, g.Mode) || err != nil || g.Endpoint != want.Endpoint || g.SourceSHA256 != want.SourceSHA256 || g.CatalogVersion != want.CatalogVersion || g.CatalogHash != want.CatalogHash || g.SDK != want.SDK || !ok || g.Schema != want.Schema || !alias.MatchString(g.ID) || g.Contract != want.Contract || g.Combo != combo || g.Provider != r.Provider || g.Purpose != want.Purpose || !alias.MatchString(g.AccountAlias) || !alias.MatchString(g.RegionAlias) || !slices.Equal(g.Models, RegionalModels(combo)) || !g.ExpiresAt.After(at) || g.ExpiresAt.Sub(at) > 24*time.Hour || g.MaxCalls != want.MaxCalls || g.MaxAttempts != want.MaxAttempts || g.MaxRequestedOutputTokens != want.MaxRequestedOutputTokens || g.MaxReservedInputTokens != want.MaxReservedInputTokens || RegionalIsKimi(combo) && (!regionalValidUSD(g.MaxUSD) || len(g.BinarySHA256) != 64) {
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

// Optional means omitted or an integer, never JSON null. Validate before the
// typed zero value can erase that distinction at the native grant boundary.
func (g *RegionalGrant) UnmarshalJSON(raw []byte) error {
	// encoding/json matches struct fields case-insensitively and retains the
	// last duplicate. Inspect keys first so this new authorization control has
	// one exact spelling and value; unrelated legacy grant fields keep their
	// existing decoding contract.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("regional grant must be an object")
	}
	seen := false
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return fmt.Errorf("regional grant key invalid")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
		if strings.EqualFold(key, "callIntervalMs") {
			if key != "callIntervalMs" || seen || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("regional call interval must be one canonical integer field")
			}
			seen = true
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	type wireGrant RegionalGrant
	var meta struct{ Schema int }
	if json.Unmarshal(raw, &meta) != nil {
		return fmt.Errorf("invalid grant JSON")
	}
	if meta.Schema == 2 {
		guard := json.NewDecoder(bytes.NewReader(raw))
		guard.UseNumber()
		if err := uniqueFields(guard, 0); err != nil {
			return err
		}
		var keys map[string]json.RawMessage
		if json.Unmarshal(raw, &keys) != nil {
			return fmt.Errorf("invalid selected grant")
		}
		allowed := map[string]bool{}
		typ := reflect.TypeFor[wireGrant]()
		for i := 0; i < typ.NumField(); i++ {
			allowed[strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]] = true
		}
		for key := range keys {
			if !allowed[key] {
				return fmt.Errorf("unknown or noncanonical selected grant field")
			}
		}
	}
	var value wireGrant
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	*g = RegionalGrant(value)
	return nil
}
