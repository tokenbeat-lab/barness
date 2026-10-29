package supportmatrix

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// ProcessScope fills machine facts only. Human identity, aliases and expiry
// remain blank in templates, and a template confers no authority.
func ProcessScope(route, mode, stage, binary string) (livegrant.Scope, error) {
	if route == "xai-images" {
		return xaiImagesProcessScope(mode, stage, binary)
	}
	s := livegrant.Scope{Route: route, Mode: mode, Stage: stage, BinarySHA256: binary, MaxOutputTokens: 4096, CallSeconds: 180, ProcessSeconds: 1800}
	cat := ai.BuiltinCatalog()
	var source map[string]string
	var err error
	s.Scenes = GenericExpected(route)
	if len(s.Scenes) > 0 {
		s.MaxCalls, s.MaxAttempts, s.MaxRetries = 32, 32, 1
		s.Provider, s.API, s.Endpoint = "openai", "openai-responses", "https://api.openai.com/v1"
		s.Models = []string{"gpt-5-mini", "gpt-5.4", "gpt-6-sol"}
		switch route {
		case "openai-chat":
			s.API = "openai-completions"
			s.Models = append(s.Models, "gpt-5.4-pro")
		case "deepseek-responses":
			s.Provider, s.Endpoint, s.Models = "deepseek", "https://api.deepseek.com", []string{"deepseek-flash"}
		case "deepseek-chat":
			s.Provider, s.API, s.Endpoint, s.Models = "deepseek", "openai-completions", "https://api.deepseek.com", []string{"deepseek-flash", "deepseek-v4-pro"}
		case "anthropic-messages":
			s.Provider, s.API, s.Endpoint = "anthropic", "anthropic-messages", "https://api.anthropic.com"
			s.Models = []string{"claude-haiku-4-5", "claude-opus-4-8", "claude-fable-5", "claude-fable-5-1", "claude-opus-5", "claude-opus-5-5"}
		case "google-gemini":
			s.Provider, s.API, s.Endpoint, s.Models = "google", "google-generative-ai", "https://generativelanguage.googleapis.com/v1beta", []string{"gemini-3.8-flash"}
		}
		source, err = RegionalExecutionSource()
	} else if route == "zhipu-chat" || route == "zhipu-messages" {
		s.Provider, s.Models, s.Scenes, s.CallSeconds = "zhipu", []string{"glm-5.3", "glm-5.2"}, GLMExpected(route), 120
		s.MaxCalls, s.MaxAttempts = 24, 32
		s.API, s.Endpoint = "openai-completions", glm.ChatEndpoint
		if route == "zhipu-messages" {
			s.API, s.Endpoint = "anthropic-messages", glm.MessagesEndpoint
		}
		if stage == "candidate" {
			root, e := minimaxlive.Root()
			if e != nil {
				return s, e
			}
			raw, e := minimaxlive.ReadFile(filepath.Join(root, "ai/live/testdata/glm-candidate/candidate.json"), 2<<20)
			if e != nil {
				return s, e
			}
			if e = json.Unmarshal(raw, &cat); e != nil {
				return s, e
			}
		}
		source, err = GLMExecutionSource()
	} else if route == "minimax-chat" || route == "minimax-responses" || route == "minimax-messages" {
		id, e := minimaxlive.Select(strings.TrimPrefix(route, "minimax-"), minimaxlive.M3, "standard")
		if e != nil {
			return s, e
		}
		return minimaxlive.ProcessScope(id, mode, stage, binary)
	} else {
		return s, fmt.Errorf("unknown bounded process grant route")
	}
	if err != nil {
		return s, err
	}
	s.CatalogVersion = cat.Version
	s.CatalogHash, err = cat.Hash()
	if err != nil {
		return s, err
	}
	if stage == "formal" {
		for _, model := range s.Models {
			if _, ok := cat.Lookup(ai.ProviderID(s.Provider), ai.API(s.API), model); !ok && model != "gpt-5.4-pro" {
				return s, fmt.Errorf("formal process model absent from published catalog")
			}
		}
	}
	s.SourceSHA256 = RegionalSourceHash(source)
	s.SDK, err = RegionalSDK(route)
	if s.Provider == "google" {
		s.SDK, err = "direct HTTP (ADR-0012)", nil
	}
	if len(s.Scenes) == 0 || !slices.Contains([]string{"candidate", "formal"}, stage) {
		return s, fmt.Errorf("invalid bounded process stage")
	}
	return s, err
}
