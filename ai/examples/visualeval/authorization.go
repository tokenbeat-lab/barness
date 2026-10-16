package main

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
)

type authorization struct {
	Route               string `json:"route"`
	Model               string `json:"model"`
	Service             string `json:"service"`
	CatalogHash         string `json:"catalogHash"`
	AccountAlias        string `json:"accountAlias"`
	RegionAlias         string `json:"regionAlias"`
	Reference           string `json:"reference"`
	DedicatedCredential bool   `json:"dedicatedCredential"`
	MaxCalls            int    `json:"maxCalls"`
	MaxAttempts         int    `json:"maxAttempts"`
}

func official(route string) string {
	if route == "jev" {
		return "https://api.typesafe.ai/v1"
	}
	return "https://api.openai.com/v1"
}
func service(route string) string {
	if route == "jev" {
		return official(route) + "/systemone"
	}
	return official(route) + "/decisions"
}
func authorize(p *plan, a arguments) (*authorization, string, string, error) {
	credentialName := regexp.MustCompile(`(?i)(key|token|secret|password|passwd|credential|auth)`)
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value != "" && name != "BARNESS_VISUAL_KEY" && (credentialName.MatchString(name) || strings.Contains(name, "BASE_URL")) {
			return nil, "", "", errors.New("foreign credential or endpoint environment is set")
		}
	}
	if a.authorization == "" {
		return nil, "", "explicit isolated credential and Host authorization are absent", nil
	}
	raw, err := readBounded(a.authorization, 16384)
	if err != nil {
		return nil, "", "Host authorization is unavailable", nil
	}
	var proof authorization
	var fields map[string]json.RawMessage
	if strict(raw, &fields) != nil || len(fields) != 10 {
		return nil, "", "", errors.New("invalid authorization document")
	}
	for _, key := range []string{"route", "model", "service", "catalogHash", "accountAlias", "regionAlias", "reference", "dedicatedCredential", "maxCalls", "maxAttempts"} {
		if fields[key] == nil {
			return nil, "", "", errors.New("noncanonical authorization field")
		}
	}
	if strict(raw, &proof) != nil || validAuthorization(p, &proof) != nil {
		return nil, "", "", errors.New("authorization does not match pinned route/model/catalog/budget")
	}
	key := os.Getenv("BARNESS_VISUAL_KEY")
	if strings.TrimSpace(key) == "" {
		return nil, "", "the sole explicit evaluation credential is absent", nil
	}
	return &proof, key, "", nil
}

func validAuthorization(p *plan, proof *authorization) error {
	hash, _ := p.catalog.Hash()
	alias := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{1,79}$`)
	if proof == nil || proof.Route != p.route || proof.Model != p.model.ID || proof.Service != service(p.route) || proof.CatalogHash != hash || !proof.DedicatedCredential || proof.MaxCalls != 12 || proof.MaxAttempts != 12 || !alias.MatchString(proof.AccountAlias) || !alias.MatchString(proof.RegionAlias) || !alias.MatchString(proof.Reference) {
		return errors.New("invalid pinned authorization")
	}
	return nil
}
