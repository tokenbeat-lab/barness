package chineseeval

import (
	"encoding/json"
	"errors"

	"github.com/tokenbeat-lab/barness/ai"
)

// Route is the complete authorized model identity, independent of ResponseModel.
// The two fixed host routes choose their own endpoint and credential source.
type Route struct {
	Operation ai.Operation  `json:"operation"`
	Provider  ai.ProviderID `json:"provider"`
	API       ai.API        `json:"api"`
	Model     string        `json:"model"`
}

func RouteNamed(name string) (Route, error) {
	switch name {
	case "jev":
		return Route{ai.OperationClassifier, ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.13.0"}, nil
	case "luna":
		return Route{ai.OperationClassifier, ai.ProviderOpenAI, ai.APIOpenAIDecisions, "gpt-6-luna"}, nil
	default:
		return Route{}, errors.New("evaluation: unknown fixed route")
	}
}
func (r Route) Valid() bool {
	for _, name := range []string{"jev", "luna"} {
		want, _ := RouteNamed(name)
		if r == want {
			return true
		}
	}
	return false
}
func (r Route) Hash() string { raw, _ := json.Marshal(r); return Hash(raw) }
func (r Route) KeyEnv() string {
	if r.Provider == ai.ProviderOpenAI {
		return "BARNESS_AI_CHINESE_EVAL_OPENAI_KEY"
	}
	return "BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY"
}
func (r Route) Endpoint() string {
	if r.Provider == ai.ProviderOpenAI {
		return "https://api.openai.com/v1"
	}
	return "https://api.typesafe.ai/v1"
}
func (r Route) SDK() string {
	if r.Provider == ai.ProviderOpenAI {
		return "direct-net-http-openai-decisions"
	}
	return "direct-net-http-fixed-jev-1.13.0"
}
func (p *Plan) Route() Route {
	if p.config.Route != nil {
		return *p.config.Route
	}
	r, _ := RouteNamed("jev")
	return r
}

// Catalog returns a separately decoded snapshot for the host's evidence and Client.
func (p *Plan) Catalog() ai.Catalog {
	raw, _ := json.Marshal(p.catalog)
	var c ai.Catalog
	json.Unmarshal(raw, &c)
	return c
}
