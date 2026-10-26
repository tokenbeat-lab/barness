package grokhistory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"io"
	"slices"
)

// Provenance contains only the fixed synthetic source identities and actual
// tenant. TrustedNativeState is omitted by Message JSON, so capture its public
// values independently before and after the call.
type Provenance struct {
	Before []ai.NativeStateEnvelope `json:"before"`
	After  []ai.NativeStateEnvelope `json:"after"`
}

func Envelopes(req ai.Request) []ai.NativeStateEnvelope {
	var out []ai.NativeStateEnvelope
	for _, message := range req.Messages {
		if m, ok := message.(ai.AssistantMessage); ok {
			env, _ := m.NativeState.Envelope()
			out = append(out, env)
		}
	}
	return out
}

func VerifyProvenance(raw []byte, tenant string) error {
	var p Provenance
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if len(raw) > 4096 || tenant == "" || decoder.Decode(&p) != nil {
		return fmt.Errorf("xAI history: bounded source provenance absent")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return fmt.Errorf("xAI history: trailing provenance input")
	}
	env := ai.NativeStateEnvelope{TenantID: tenant, AccountScopeID: "synthetic-xai-source-account", ProviderID: ai.ProviderXAI, API: ai.APIOpenAIResponses, ModelID: "grok-4.7"}
	want := []ai.NativeStateEnvelope{env, env}
	if !slices.Equal(p.Before, want) || !slices.Equal(p.After, want) {
		return fmt.Errorf("xAI history: original trusted source changed")
	}
	return nil
}
