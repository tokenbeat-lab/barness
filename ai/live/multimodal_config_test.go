//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

type multimodalAuthorization struct {
	Route                  string `json:"route"`
	Model                  string `json:"model"`
	Service                string `json:"service"`
	AccountAlias           string `json:"accountAlias"`
	RegionAlias            string `json:"regionAlias"`
	AuthorizationReference string `json:"authorizationReference"`
	DedicatedCredential    bool   `json:"dedicatedCredential"`
	MaxCalls               int    `json:"maxCalls"`
	MaxAttempts            int    `json:"maxAttempts"`
}

type multimodalRoute struct {
	name, model, endpoint, path string
	provider                    ai.ProviderID
	api                         ai.API
}

var multimodalCredentialVariable = regexp.MustCompile(`(?i)(key|token|secret|password|passwd|credential|auth)`)

func selectedMultimodalRoute() (multimodalRoute, error) {
	switch os.Getenv("BARNESS_MM_ROUTE") {
	case "jev":
		return multimodalRoute{"jev", "jev-1.14.0", "https://api.typesafe.ai/v1", "/systemone", ai.ProviderTypeSafe, ai.APITypeSafeSystemOne}, nil
	case "decisions":
		return multimodalRoute{"decisions", "gpt-6-luna", "https://api.openai.com/v1", "/decisions", ai.ProviderOpenAI, ai.APIOpenAIDecisions}, nil
	default:
		return multimodalRoute{}, fmt.Errorf("select exactly one candidate route")
	}
}

func multimodalConfiguration(route multimodalRoute, rec *recorder) (multimodalAuthorization, string, error) {
	var proof multimodalAuthorization
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value == "" {
			continue
		}
		foreignCredential := multimodalCredentialVariable.MatchString(name) && name != "BARNESS_MM_KEY" && name != "BARNESS_MM_AUTHORIZATION"
		if foreignCredential || strings.Contains(name, "BASE_URL") || name == "BARNESS_AI_LIVE_ENDPOINT" || name == envCombo {
			return proof, "", fmt.Errorf("foreign credential or endpoint override is set")
		}
	}
	if os.Getenv("BARNESS_MM_LIVE") != "1" || strings.TrimSpace(os.Getenv("BARNESS_MM_KEY")) == "" || os.Getenv("BARNESS_MM_AUTHORIZATION") == "" {
		return proof, "explicit switch, isolated approved credential and authorization record are required", nil
	}
	raw, err := readDecisionsFile(os.Getenv("BARNESS_MM_AUTHORIZATION"), 16384)
	if err != nil {
		return proof, "authorization record is unavailable", nil
	}
	// Reject aliases, duplicate names and unknown fields before typed decoding.
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || len(fields) != 9 {
		return proof, "", fmt.Errorf("invalid authorization record")
	}
	for _, key := range []string{"route", "model", "service", "accountAlias", "regionAlias", "authorizationReference", "dedicatedCredential", "maxCalls", "maxAttempts"} {
		if fields[key] == nil || strings.Count(string(raw), `"`+key+`"`) != 1 {
			return proof, "", fmt.Errorf("noncanonical authorization record")
		}
	}
	if json.Unmarshal(raw, &proof) != nil || proof.Route != route.name || proof.Model != route.model || proof.Service != route.endpoint+route.path || !proof.DedicatedCredential || proof.MaxCalls != 12 || proof.MaxAttempts != 12 || !decisionsAlias.MatchString(proof.AccountAlias) || !decisionsAlias.MatchString(proof.RegionAlias) || !decisionsAlias.MatchString(proof.AuthorizationReference) {
		return proof, "", fmt.Errorf("authorization does not match selected route, model or finite budget")
	}
	if proxy := os.Getenv("BARNESS_AI_LIVE_PROXY"); proxy != "" {
		u, err := smokeProxy(proxy)
		if err != nil {
			return proof, "", err
		}
		if transport, ok := rec.next.(*http.Transport); ok {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return proof, "", nil
}

func multimodalPolicy() *ai.ResourcePolicy {
	p := livePolicy()
	p.Image = nil
	p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 32, MaxStateBytes: 4 << 20, MaxQuestionBytes: 16384, Images: &ai.ClassifierImagePolicy{MaxImages: 24, MaxTotalImageBytes: 4 << 20, References: &ai.ClassifierImageReferencePolicy{MaxReferences: 32, MaxExpandedImageBytes: 16 << 20}}}
	p.MaxImageBytes, p.MaxRequestBytes, p.MaxOutputBytes = 4<<20, 4<<20, 512<<10
	p.CallTimeout, p.ConnectTimeout, p.ResponseHeaderTimeout, p.ReadIdleTimeout = 30*time.Second, 5*time.Second, 25*time.Second, 25*time.Second
	p.MaxConcurrentPerTenant, p.MaxConcurrentProcess = 1, 1
	return p
}
