package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
)

// K3 effort is standalone: Claude adaptive, budget and per-turn controls
// describe different contracts. Model metadata may narrow this fixed subset.
func validateKimiMessagesOptions(m Model, options Options) *Error {
	if m.API != APIAnthropicMessages || anthropicCapabilitiesOf(m.Provider).implementation != anthropicKimi {
		return nil
	}
	o, _ := options.(AnthropicOptions)
	invalid := func() *Error {
		return newError(CodeInvalidRequest, PhaseCapability, "Kimi Messages option is outside the supported subset")
	}
	if !o.ThinkingEnabled.IsZero() || o.ThinkingBudgetTokens != 0 || o.ThinkingDisplay != "" || !o.InterleavedThinking.IsZero() || !o.Temperature.IsZero() {
		return invalid()
	}
	if o.Effort != "" && !kimiMessagesEffort(m, string(o.Effort)) {
		return invalid()
	}
	if o.MaxTokens.IsNull() {
		return invalid()
	}
	if n, ok := o.MaxTokens.Get(); ok && (n < 1 || n > min(m.MaxTokens, 131072)) {
		return invalid()
	}
	if o.CacheRetention == CacheRetentionNone || o.CacheRetention == CacheRetentionLong && !m.supportsLongCacheRetention() {
		return invalid()
	}
	if choice, ok := o.ToolChoice.Get(); ok && (choice.Tool != "" || !slices.Contains([]string{"auto", "any", "none"}, choice.Mode)) {
		return invalid()
	}
	return nil
}

func kimiMessagesEffort(m Model, effort string) bool {
	return m.Reasoning && slices.Contains([]string{"low", "high", "max"}, effort) && slices.Contains(supportedThinkingLevels(m), ThinkingLevel(effort))
}

func validateKimiMessagesBody(m Model, raw []byte, header http.Header, beta string, authorize func(map[string]any) string) *Error {
	if anthropicCapabilitiesOf(m.Provider).implementation != anthropicKimi {
		return nil
	}
	invalid := func() *Error {
		return newError(CodeInvalidRequest, PhaseRequest, "Kimi Messages request is outside the supported subset")
	}
	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if dec.Decode(&body) != nil || body == nil {
		return invalid()
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if beta != "" || header.Get("Anthropic-Dangerous-Direct-Browser-Access") != "" {
		return invalid()
	}
	if v := header.Get("Anthropic-Version"); v != "" && v != "2023-06-01" {
		return invalid()
	}
	if !kimiMessagesFields(body, []string{"model", "messages", "system", "max_tokens", "stream", "tools", "tool_choice", "output_config", "metadata", "cache_control"}) || body["stream"] != true {
		return invalid()
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if !outputBudgetNumber(fields["max_tokens"], min(m.MaxTokens, 131072)) {
		return invalid()
	}
	if value, present := body["output_config"]; present {
		config, ok := value.(map[string]any)
		effort, _ := config["effort"].(string)
		if !ok || len(config) != 1 || !kimiMessagesEffort(m, effort) {
			return invalid()
		}
	}
	if value, present := body["tool_choice"]; present {
		choice, ok := value.(map[string]any)
		tools, _ := body["tools"].([]any)
		if !ok || len(choice) != 1 || !slices.Contains([]any{"auto", "any", "none"}, choice["type"]) || choice["type"] == "any" && len(tools) == 0 {
			return invalid()
		}
	}
	cc, ok := body["cache_control"].(map[string]any)
	if !ok || len(cc) != 2 || cc["type"] != "ephemeral" || !slices.Contains([]any{"5m", "1h"}, cc["ttl"]) || cc["ttl"] == "1h" && !m.supportsLongCacheRetention() {
		return invalid()
	}
	if value, present := body["metadata"]; present {
		meta, ok := value.(map[string]any)
		_, str := meta["user_id"].(string)
		if !ok || len(meta) != 1 || !str {
			return invalid()
		}
	}
	if value, present := body["system"]; present && !kimiMessagesBlocks(value, "system") {
		return invalid()
	}
	if !kimiMessagesHistory(body["messages"]) {
		return invalid()
	}
	if value, present := body["tools"]; present && !kimiMessagesTools(value) {
		return invalid()
	}
	return nil
}

func kimiMessagesFields(obj map[string]any, allowed []string) bool {
	for key := range obj {
		if !slices.Contains(allowed, key) {
			return false
		}
	}
	return true
}

func kimiMessagesTools(value any) bool {
	tools, ok := value.([]any)
	if !ok {
		return false
	}
	for _, value := range tools {
		tool, ok := value.(map[string]any)
		if !ok || !kimiMessagesFields(tool, []string{"name", "description", "input_schema"}) || !miniMaxMessagesString(tool["name"]) {
			return false
		}
		if description, present := tool["description"]; present {
			if _, ok := description.(string); !ok {
				return false
			}
		}
		schema, ok := tool["input_schema"].(map[string]any)
		if !ok || schema["type"] != "object" {
			return false
		}
		// JSON schema properties and tool input remain business JSON. Never walk
		// their property names as protocol fields or vendor-side references.
	}
	return true
}

// The final payload is authoritative: a hook may replace the original choice.
func kimiMessagesRequiresTool(m Model, raw []byte) bool {
	if !isKimi(m.Provider) {
		return false
	}
	var body struct {
		Choice struct {
			Type string `json:"type"`
		} `json:"tool_choice"`
	}
	_ = json.Unmarshal(raw, &body)
	return body.Choice.Type == "any"
}
