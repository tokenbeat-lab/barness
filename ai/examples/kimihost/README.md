# Explicit Kimi market Host example

`kimihost.Bindings` requires an explicit `ProviderMoonshot` (international) or
`ProviderMoonshotCN` (China), tenant, account, credential reference, version and
`Standard` API purpose. It creates three independently revocable K3-only routes:

| Market | Binding IDs | Chat / Responses SDK base | Messages SDK base |
| --- | --- | --- | --- |
| International | `kimi-intl-chat`, `kimi-intl-responses`, `kimi-intl-messages` | `https://api.moonshot.ai/v1` | `https://api.moonshot.ai/anthropic` |
| China | `kimi-cn-chat`, `kimi-cn-responses`, `kimi-cn-messages` | `https://api.moonshot.cn/v1` | `https://api.moonshot.cn/anthropic` |

The adapters append `/chat/completions`, `/responses` and `/v1/messages`.
All three protocols use one Bearer credential. The market is never inferred from
an endpoint or key; both markets have independent accounts, authorization and
price/evidence records. A key may be shared across protocols only when its
same-market, same-tenant account record explicitly allows those APIs.

`Endpoints` overrides must match the exact official service/protocol base.
`TrustedGateway: true` permits a reviewed nonofficial HTTPS target or controlled
loopback HTTP (also requiring Client opt-in). Another official service/market,
wrong official path, URL credentials, query and fragment reject even with gateway
trust. A gateway does not grant another service identity or account purpose.

## Local assembly

Choose a named key source at assembly time; the library reads no conventional
provider environment variable. Load a reviewed `ai.Catalog` as trusted deployment
configuration. The six source-backed candidates under
`ai/internal/testkit/kimi/` are available for qualification and do not activate
`BuiltinCatalog` or confer formal support. Never rewrite their market identity.

```go
bindings, err := kimihost.Bindings(kimihost.Authorization{
    ProviderID: ai.ProviderMoonshotCN, Purpose: kimihost.Standard,
    TenantID: "local-tenant", AccountScopeID: "approved-cn-account",
    CredentialRef: "cn-standard-kimi", Version: "local",
})
if err != nil { return err }
local, err := localassembly.OpenOperations(localassembly.OperationsConfig{
    TenantID: "local-tenant", Bindings: bindings, Catalog: &reviewedCatalog,
    Keys: map[string]localassembly.KeySource{
        "cn-standard-kimi": {EnvVar: "MY_EXPLICIT_CN_KIMI_KEY"}, // or File
    },
    Policy: localassembly.LocalPolicy(),
})
if err != nil { return err }
turn, err := kimihost.Turn(ctx, local.Client, local.Scope(),
    ai.Target{BindingID: "kimi-cn-chat", ModelID: "kimi-k3"},
    ai.Request{Messages: []ai.Message{ai.UserText("Explain this code.")}},
    ai.APIOpenAICompletions, kimihost.FullStream, ai.Hooks{})
// Handle typed ai.Error and turn.Result even on failure; keep content private.
```

`Turn` demonstrates FullStream, FullComplete, SimpleStream and SimpleComplete,
plus trusted Hooks. Every entry caps output at4096 and leaves effort omitted,
retaining K3's max default. The supported model subset is text, context1048576,
output131072, low/high/max effort. Simple canonical levels clamp to that subset.
Streaming drains attributed envelopes, awaits Result and closes the stream.
This small example retains events; a production UI should render incrementally
under bounded queue limits. Reassemble immutable local state after key changes.

## Cloud assembly

Authenticate principals outside this package; never accept a public DTO as
authority for tenant, endpoint, credential reference or hooks. Store returned
bindings under `(TenantID, BindingID)` in a versioned backend. Keep the following
credential metadata independently in the same consistent backend snapshot:

```go
record := kimihost.CredentialRecord{
    ProviderID: authorization.ProviderID, TenantID: principal.TenantID,
    AccountScopeID: approvedAccount, CredentialRef: selectedRecordRef,
    Purpose: kimihost.Standard, Active: true,
    Version: "secret-v3", BindingVersion: "binding-v7",
    AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses,
        ai.APIAnthropicMessages},
}
credentials := kimihost.Credentials{
    Authorization: authorization, Record: record,
    ReadSecret: func(ctx context.Context, ref string) (ai.Secret, error) {
        return tenantSecretStore.Read(ctx, principal.TenantID, ref)
    },
}
client, err := ai.NewClient(ai.Config{
    Policy: hostintegration.CloudInteractivePolicy(), Catalog: &reviewedCatalog,
    Bindings: tenantBindingResolver, Credentials: credentials,
    Observer: metadataOnlyObserver,
})
```

`Credentials` checks Provider, Tenant, AccountScope, purpose, allowed APIs,
Active and both independent versions before invoking ReadSecret. It checks the
selected Binding against the authorized route. Do not derive record versions
from the Binding received by ResolveCredential. Resolve current state atomically;
these example values are immutable while used. The BindingResolver still owns
actor authorization and revocation. Same labels at two markets grant no access.

The Host supplies fresh RequestIDs and optional bounded distributed Admission,
executes tools after `ValidateToolCall`, appends each original assistant and
matching ToolResult, and starts a fresh logical call. Preserve native history
with its trusted envelope; persisted/untrusted JSON alone grants no replay trust.
No model, protocol, key, market or endpoint fallback occurs on failure.

## Cache, cost and withdrawal

Default/short requests5m cache writes, long requests1h; none is unsupported.
The preference does not prove actual TTL. Caches are org-scoped at the vendor;
derived session keys do not provide vendor-side tenant isolation. Use separate
org/accounts when hard isolation is required.

Read `Attempt.CostEstimate` with `UsageReporting`. Resolved amounts agree;
bounded reports both finite USD bounds and uses the lower bound as Usage.Cost;
reserve budget using Maximum. Chat/Responses optional cached_tokens defaults to0 before completeness checks.
Missing write counts keep UsageReporting partial; when known total input/output
constrain every allocation, CostEstimate may still be bounded or resolved.
Partial/unreported cost estimates omit amount bounds and never
mean free service. Reasoning tokens are already included in output. Chat may
have complete counters yet an unknown paid write tier. Reported Messages
TTL counters or applied Responses options resolve those tiers. Request TTL is
never observed billing evidence.

International per-million input/5m-write/1h-write/read/output rates are
USD3/3/6/.30/15; China retains CNY20/20/40/2/100. China USD estimates use the
fixed retained2026-10-16 CFETS1USD=6.7220CNY snapshot, ten-decimal half-even
base rates, with1h=rounded Input×2. This is a historical estimate, not today's FX
or an invoice. Candidate native facts, original hashes and conversion provenance
are retained under `ai/internal/testkit/kimi/`; price changes require new versions.

Disable a route by atomically setting Enabled=false with a new binding version.
New calls fail before secret reads or Attempts; already sent calls finish under
their pinned snapshot and retain usage. Other markets/protocols stay independent.
Before software rollback, disable Kimi bindings, remove Kimi identities from the
loaded catalog, then restore the prior software snapshot. Preserve all history
and evidence. Restoring permission requires a new consistent authorization and
credential snapshot and a fresh call. Offline success alone grants no formal
catalog/support-matrix promotion; live/race/load each need their own SOP scope.

No-key public proof: `go test ./ai/e2e -run '^TestKimiHost' -count=1` (run via
repository `docs/agents/verification.md` for retained receipts).
