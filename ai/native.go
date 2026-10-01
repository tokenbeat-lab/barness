package ai

import (
	"errors"
	"fmt"
)

// NativeStateEnvelope is the provenance of an assistant message's native
// state — reasoning signatures and encrypted content, redacted blocks, and
// provider item ids: the tenant, vendor account and provider/API/model that
// produced it. A host stores it beside the message. On its own it proves
// nothing: stored or client-supplied fields are claims until the host vouches
// for them with TrustNativeState.
type NativeStateEnvelope struct {
	TenantID       string     `json:"tenantId"`
	AccountScopeID string     `json:"accountScopeId"`
	ProviderID     ProviderID `json:"providerId"`
	API            API        `json:"api"`
	ModelID        string     `json:"modelId"`
}

// complete reports whether every field is set, so an envelope can only match
// a call field by field, never by two fields both being empty.
func (e NativeStateEnvelope) complete() bool {
	return e.TenantID != "" && e.AccountScopeID != "" && e.ProviderID != "" && e.API != "" && e.ModelID != ""
}

// TrustedNativeState is an envelope a trusted party vouched for. The zero
// value is untrusted, and decoding JSON never produces anything else: the
// envelope is unexported and AssistantMessage does not serialize the field.
// Only two paths make one:
//
//   - barness-ai itself, on every message of a call that resolved its
//     binding: the message was produced in this process for that tenant,
//     account and model, so a host passing a Result's message straight into
//     its next call replays it natively.
//   - TrustNativeState, after the host restored message and envelope from its
//     own storage and checked their tenant and integrity there.
//
// Go code in the host process can copy the value between messages; in-process
// untrusted code is outside what the library can guard (spec I3).
type TrustedNativeState struct {
	env NativeStateEnvelope
}

// ErrInvalidNativeStateEnvelope matches every TrustNativeState refusal.
var ErrInvalidNativeStateEnvelope = errors.New("ai: invalid native state envelope")

// TrustNativeState is the host's statement that env is the authentic
// provenance of a message it restored for scope's tenant, after checking that
// in its own store. It refuses an incomplete envelope or one naming another
// tenant than the scope; it does not check accounts or models, because the
// library compares those with the call's binding and target when the message
// is replayed, and downgrades rather than fails on a mismatch.
func TrustNativeState(scope CallScope, env NativeStateEnvelope) (TrustedNativeState, error) {
	switch {
	case scope.TenantID == "":
		return TrustedNativeState{}, fmt.Errorf("%w: the scope has no TenantID", ErrInvalidNativeStateEnvelope)
	case !env.complete():
		return TrustedNativeState{}, fmt.Errorf("%w: tenant, account, provider, API and model are all required", ErrInvalidNativeStateEnvelope)
	case env.TenantID != scope.TenantID:
		return TrustedNativeState{}, fmt.Errorf("%w: the envelope belongs to another tenant than the scope", ErrInvalidNativeStateEnvelope)
	}
	return TrustedNativeState{env: env}, nil
}

// Envelope returns the vouched-for provenance, for the host to store beside
// the message; ok is false for untrusted state.
func (t TrustedNativeState) Envelope() (env NativeStateEnvelope, ok bool) {
	return t.env, t.env.complete()
}

// NativeStateDowngrades counts the history's assistant messages whose native
// state was not replayed, by reason, in one call. Each such message was
// converted by the cross-model rules instead and the call went ahead. It is
// call metadata, never part of the pi-compatible message.
type NativeStateDowngrades struct {
	// NoEnvelope: the message's state was not vouched for the calling tenant
	// (never trusted, decoded from storage without TrustNativeState, or
	// trusted for another tenant).
	NoEnvelope int `json:"noEnvelope,omitempty"`
	// AccountMismatch: same provider, API and model, but the state belongs to
	// another vendor account than the binding's.
	AccountMismatch int `json:"accountMismatch,omitempty"`
	// CrossModel: the message, or its trusted envelope, is from another
	// provider, API or model than the target.
	CrossModel int `json:"crossModel,omitempty"`
}

// replayOrigin is what a call compares a history message's provenance with:
// the calling tenant, the binding's account and the target model.
type replayOrigin struct {
	tenant  string
	account string
	model   Model
}

// isTarget reports whether provider, API and model name the call's target.
func (o replayOrigin) isTarget(provider ProviderID, api API, model string) bool {
	return provider == o.model.Provider && api == o.model.API && model == o.model.ID
}

func (o replayOrigin) envelope() NativeStateEnvelope {
	return NativeStateEnvelope{TenantID: o.tenant, AccountScopeID: o.account, ProviderID: o.model.Provider, API: o.model.API, ModelID: o.model.ID}
}

// downgradeReason is why a message's native state is not replayed natively.
type downgradeReason int

const (
	noDowngrade downgradeReason = iota
	downgradeNoEnvelope
	downgradeAccountMismatch
	downgradeCrossModel
)

// add counts one downgraded message.
func (d *NativeStateDowngrades) add(r downgradeReason) {
	switch r {
	case downgradeNoEnvelope:
		d.NoEnvelope++
	case downgradeAccountMismatch:
		d.AccountMismatch++
	case downgradeCrossModel:
		d.CrossModel++
	}
}

// nativeReplay decides whether m's native state may be replayed to the
// call's target. The message's own fields decide same-model as in pi; native
// replay additionally needs a trusted envelope for the same tenant, account
// and model. Otherwise it reports why the message is downgraded. Credential
// versions are deliberately not compared: rotating the key of the same
// account keeps the state valid.
func (o replayOrigin) nativeReplay(m AssistantMessage) downgradeReason {
	env, trusted := m.NativeState.Envelope()
	switch {
	case !o.isTarget(m.Provider, m.API, m.Model):
		return downgradeCrossModel
	case !trusted || env.TenantID != o.tenant:
		return downgradeNoEnvelope
	case !o.isTarget(env.ProviderID, env.API, env.ModelID):
		return downgradeCrossModel
	case env.AccountScopeID != o.account:
		return downgradeAccountMismatch
	}
	return noDowngrade
}
