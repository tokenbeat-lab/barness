// Package livegrant owns one-process admission for the existing generic, GLM
// and international MiniMax routes. It has no route registry, credentials,
// billing estimates, HTTP, retries or authority recovery mechanism.
package livegrant

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Scope is filled from each route's own reviewed native contract. Zero
// aggregate input/output values mean the route has no such native cap.
type Scope struct {
	Route                    string   `json:"route"`
	Mode                     string   `json:"mode"`
	Stage                    string   `json:"stage"`
	Provider                 string   `json:"provider"`
	API                      string   `json:"api"`
	Endpoint                 string   `json:"endpoint"`
	SDK                      string   `json:"sdk"`
	CatalogVersion           string   `json:"catalogVersion"`
	CatalogHash              string   `json:"catalogHash"`
	SourceSHA256             string   `json:"sourceSha256"`
	BinarySHA256             string   `json:"binarySha256"`
	Models                   []string `json:"models"`
	Scenes                   []string `json:"scenes"`
	MaxCalls                 int      `json:"maxCalls"`
	MaxAttempts              int      `json:"maxAttempts"`
	MaxRequestReruns         int      `json:"maxRequestReruns"`
	MaxRetries               int      `json:"maxRetries"`
	MaxOutputTokens          int      `json:"maxOutputTokens"`
	MaxRequestedOutputTokens int      `json:"maxRequestedOutputTokens"`
	MaxReservedInputTokens   int      `json:"maxReservedInputTokens"`
	CallSeconds              int      `json:"callSeconds"`
	ProcessSeconds           int      `json:"processSeconds"`
}

type Grant struct {
	Schema       int       `json:"schema"`
	ID           string    `json:"id"`
	AccountAlias string    `json:"accountAlias"`
	RegionAlias  string    `json:"regionAlias"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Scope        Scope     `json:"scope"`
}

func Hash(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func FileHash(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() > 256<<20 {
		return "", fmt.Errorf("live process binary invalid")
	}
	h := sha256.New()
	if _, e = io.Copy(h, io.LimitReader(f, 256<<20+1)); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Decode(raw []byte) (Grant, error) {
	var g Grant
	if e := grantJSON(raw); e != nil {
		return g, e
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&g); e != nil {
		return g, e
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return g, fmt.Errorf("trailing grant input")
	}
	return g, nil
}

func (g Grant) Check(want Scope, at time.Time) error {
	if !reflect.DeepEqual(g.Scope, want) {
		a, _ := json.Marshal(g.Scope)
		b, _ := json.Marshal(want)
		return fmt.Errorf("fresh process machine scope differs: got=%s want=%s", Hash(a), Hash(b))
	}
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,96}$`)
	hex64 := regexp.MustCompile(`^[a-f0-9]{64}$`)
	if g.Schema != 1 || !alias.MatchString(g.ID) || !alias.MatchString(g.AccountAlias) || !alias.MatchString(g.RegionAlias) || !g.ExpiresAt.After(at) || g.ExpiresAt.Sub(at) > 24*time.Hour || !reflect.DeepEqual(g.Scope, want) || !hex64.MatchString(want.SourceSHA256) || !hex64.MatchString(want.BinarySHA256) || !hex64.MatchString(strings.TrimPrefix(want.CatalogHash, "sha256:")) || want.Mode != "live" && want.Mode != "controlled" || want.Stage != "candidate" && want.Stage != "formal" || want.Mode == "controlled" && want.Stage == "formal" || len(want.Scenes) == 0 || len(want.Models) == 0 || want.MaxCalls < 1 || want.MaxAttempts < 1 || want.ProcessSeconds != 1800 {
		return fmt.Errorf("fresh live process grant scope, aliases, source, binary, expiry or budget differs")
	}
	return nil
}
