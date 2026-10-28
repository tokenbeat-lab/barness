package supportmatrix

import (
	"fmt"
	"slices"
	"strings"
)

// VerifyGenericReport checks the complete fixed native smoke report. Historical
// matrix merges may retain partial outcomes; formal qualification requires every
// capability to pass except the one declared zero-request native limitation.
// Artifact integrity and process authority remain separate consumer boundaries.
func VerifyGenericReport(r Report) error {
	var provider, api, endpoint, model string
	switch r.Combo {
	case "openai-responses", "openai-chat":
		provider, endpoint, model = "openai", "https://api.openai.com/v1", "gpt-5-mini"
		api = "openai-responses"
		if r.Combo == "openai-chat" {
			api = "openai-completions"
		}
	case "anthropic-messages":
		provider, api, endpoint, model = "anthropic", "anthropic-messages", "https://api.anthropic.com", "claude-haiku-4-5"
	case "google-gemini":
		provider, api, endpoint, model = "google", "google-generative-ai", "https://generativelanguage.googleapis.com/v1beta", "gemini-3.8-flash"
	case "deepseek-responses", "deepseek-chat":
		provider, endpoint, model = "deepseek", "https://api.deepseek.com", "deepseek-flash"
		api = "openai-responses"
		if r.Combo == "deepseek-chat" {
			api = "openai-completions"
		}
	default:
		return fmt.Errorf("unknown fixed generic route")
	}
	if r.Operation != "chat" || r.Provider != provider || r.API != api || r.Model != model || r.Endpoint != endpoint ||
		strings.TrimSpace(r.AccountAlias) == "" || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) {
		return fmt.Errorf("chat report differs from the fixed official native route")
	}
	sdk := "github.com/openai/openai-go/v3 "
	if provider == "anthropic" {
		sdk = "github.com/anthropics/anthropic-sdk-go "
	} else if provider == "google" {
		sdk = "direct HTTP"
	}
	if !strings.HasPrefix(r.SDK, sdk) {
		return fmt.Errorf("chat SDK does not serve the native route")
	}
	b := r.Budget
	if b.MaxCalls != 32 || b.MaxRetries != 1 || b.MaxOutputTokens != 4096 || b.CallsUsed <= 0 || b.CallsUsed > b.MaxCalls ||
		b.HTTPAttempts <= 0 || b.HTTPAttempts > b.CallsUsed || b.InputTokensUsed < 0 || b.OutputTokensUsed < 0 ||
		b.OutputTokensUsed > b.CallsUsed*b.MaxOutputTokens || b.EnvironmentRetries < 0 {
		return fmt.Errorf("chat budget missing, invalid or exceeded")
	}
	expected := GenericExpected(r.Combo)
	got, want := slices.Clone(r.Expected), slices.Clone(expected)
	slices.Sort(got)
	slices.Sort(want)
	if len(expected) == 0 || !slices.Equal(got, want) || len(r.Scenarios) != len(expected) {
		return fmt.Errorf("incomplete generic native capability set")
	}
	seen := map[string]bool{}
	calls, attempts, retries := 0, 0, 0
	for _, s := range r.Scenarios {
		if seen[s.ID] || !slices.Contains(expected, s.ID) {
			return fmt.Errorf("duplicate or substituted generic native scenario")
		}
		seen[s.ID] = true
		if r.Combo == "openai-chat" && s.ID == "reasoning-history" {
			if !declaredGenericUnsupported(r, s) {
				return fmt.Errorf("generic native no-replay declaration changed")
			}
			continue
		}
		if s.Outcome != Pass {
			return fmt.Errorf("generic native capability did not pass")
		}
		fixedModel := model
		switch {
		case strings.HasPrefix(s.ID, "tool-changes-"):
			fixedModel = strings.TrimPrefix(s.ID, "tool-changes-")
		case strings.HasPrefix(s.ID, "effort-changes-"):
			fixedModel = strings.TrimPrefix(s.ID, "effort-changes-")
		case s.ID == "pro-model-unavailable":
			fixedModel = "gpt-5.4-pro"
		case r.Combo == "deepseek-chat" && strings.HasPrefix(s.ID, "reasoning-level-"):
			fixedModel = "deepseek-v4-pro"
		}
		if s.Model != fixedModel || s.Calls <= 0 || s.Calls > b.MaxCalls || s.Attempts <= 0 || s.Attempts > s.Calls ||
			s.Retries < 0 || s.Retries > b.MaxRetries ||
			slices.ContainsFunc(s.ProviderRequestIDs, func(id string) bool { return strings.TrimSpace(id) == "" }) {
			return fmt.Errorf("chat scenario lacks fixed model or actual bounded calls")
		}
		calls, attempts, retries = calls+s.Calls, attempts+s.Attempts, retries+s.Retries
	}
	if calls != b.CallsUsed || attempts != b.HTTPAttempts || retries != b.EnvironmentRetries {
		return fmt.Errorf("chat counters disagree with scenarios")
	}
	return nil
}
