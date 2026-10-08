package ai

import (
	"context"
	"errors"
	"net"
	"net/url"
)

// bindingUnresolved is the one refusal for a missing binding and for a record
// that belongs to another tenant or binding, so existence is not revealed.
const bindingUnresolved = "binding could not be resolved for this tenant"

// resolveBinding is the first preflight read; with callRuntime.resolve,
// resolveCredential and checkSnapshot it turns a call's trusted scope and
// target into the configuration snapshot the call is pinned to, or refuses
// before any Provider request. Every refusal message is fixed text: resolver
// errors are classified, never echoed, so secret-backend internals cannot
// reach messages or errors.
//
// resolveBinding reads the binding for (scope.TenantID, bindingID) and checks
// it is the tenant's, enabled and of a supported auth kind.
func (c *Client) resolveBinding(ctx context.Context, scope CallScope, bindingID string) (Binding, *Error) {
	if failure := contextError(ctx, PhaseBinding); failure != nil {
		return Binding{}, failure
	}
	b, err := c.bindings.ResolveBinding(ctx, scope, bindingID)
	if failure := contextError(ctx, PhaseBinding); failure != nil {
		return Binding{}, failure
	}
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			return Binding{}, newError(CodeTenantDenied, PhaseBinding, "the caller may not use this binding")
		}
		return Binding{}, newError(CodeBindingNotFound, PhaseBinding, bindingUnresolved)
	}
	// A record for another tenant or binding is a host fault; refuse it exactly
	// like a missing binding so its existence is not revealed.
	if b.TenantID != scope.TenantID || b.BindingID != bindingID {
		return Binding{}, newError(CodeBindingNotFound, PhaseBinding, bindingUnresolved)
	}
	if !b.Enabled {
		return Binding{}, newError(CodeTenantDenied, PhaseBinding, "binding is disabled")
	}
	if b.AuthKind != AuthAPIKey {
		return Binding{}, newError(CodeInvalidRequest, PhaseBinding, "binding auth kind is not supported")
	}
	switch b.Operation {
	case "":
		b.Operation = OperationChat
	case OperationChat, OperationImage, OperationClassifier:
	default:
		return Binding{}, newError(CodeInvalidRequest, PhaseBinding, "binding operation is not supported")
	}
	if !endpointAllowed(b.Endpoint, c.loopback) {
		return Binding{}, newError(CodeInvalidRequest, PhaseBinding, "binding endpoint must be https (http only for loopback)")
	}
	for _, typ := range b.AllowedHostedTools {
		if typ == "" || typ == "function" {
			return Binding{}, newError(CodeInvalidRequest, PhaseBinding, "binding hosted tool allowance names an invalid tool type")
		}
	}
	if problem := b.Retry.validate(); problem != "" {
		return Binding{}, newError(CodeInvalidRequest, PhaseBinding, problem)
	}
	return b.clone(), nil
}

// resolveCredential reads the credential the binding version references and
// checks it is usable.
func (c *Client) resolveCredential(ctx context.Context, scope CallScope, b Binding) (Credential, *Error) {
	if failure := contextError(ctx, PhaseCredential); failure != nil {
		return Credential{}, failure
	}
	cred, err := c.credentials.ResolveCredential(ctx, scope, b.clone())
	if failure := contextError(ctx, PhaseCredential); failure != nil {
		return Credential{}, failure
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrSnapshotConflict):
			return Credential{}, snapshotConflict("configuration changed while the call was resolving it")
		case errors.Is(err, ErrAccessDenied):
			return Credential{}, newError(CodeTenantDenied, PhaseCredential, "the caller may not use this credential")
		}
		return Credential{}, newError(CodeCredentialUnavailable, PhaseCredential, "credential could not be resolved for this binding")
	}
	if !cred.Active {
		return Credential{}, newError(CodeCredentialUnavailable, PhaseCredential, "credential is not active")
	}
	if cred.APIKey.reveal() == "" {
		return Credential{}, newError(CodeCredentialUnavailable, PhaseCredential, "credential has no API key")
	}
	return cred, nil
}

// checkSnapshot verifies the separately resolved binding and credential form
// one consistent snapshot: same tenant, same account, the referenced
// credential, both versioned, and paired with the same binding version. It
// fails rather than re-resolving (D2, ADR-0003); a consistent new credential
// version is not a conflict.
//
// A foreign-tenant credential is reported as
// tenant_denied rather than hidden like a foreign binding: the binding was
// already authorized for this tenant, so nothing about another tenant's
// records is revealed, and the host fault should be visible.
func checkSnapshot(scope CallScope, b Binding, cred Credential) *Error {
	switch {
	case cred.OwnerTenantID != scope.TenantID:
		return newError(CodeTenantDenied, PhaseConsistency, "credential does not belong to the calling tenant")
	case cred.CredentialID != b.CredentialRef:
		return snapshotConflict("credential is not the one the binding references")
	case b.AccountScopeID == "" || cred.AccountScopeID != b.AccountScopeID:
		return snapshotConflict("credential account does not match the binding account")
	case b.Version == "" || cred.Version == "":
		return snapshotConflict("binding and credential snapshots must both be versioned")
	case cred.BindingVersion != b.Version:
		return snapshotConflict("binding changed between the binding and credential reads")
	}
	return nil
}

func snapshotConflict(msg string) *Error {
	return newError(CodeCredentialUnavailable, PhaseConsistency,
		"no consistent configuration snapshot: "+msg+"; retry as a new call")
}

// endpointAllowed accepts https endpoints, and plain http to loopback hosts
// only when the Client was test-assembled with AllowLoopbackHTTP.
func endpointAllowed(endpoint string, allowLoopbackHTTP bool) bool {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		if !allowLoopbackHTTP {
			return false
		}
		ip := net.ParseIP(u.Hostname())
		return ip != nil && ip.IsLoopback()
	}
	return false
}
