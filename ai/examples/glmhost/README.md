# Explicit GLM market Host example

`glmhost.Bindings` requires an explicit `ProviderID` (`ProviderZhipu` China or
`ProviderZAI` international), standard purpose and account authorization. It returns
three independent `OperationChat` allowlists for both `glm-5.3` and `glm-5.2`. Each SDK base is complete: Responses
`https://open.bigmodel.cn/api/v1` adds `/responses`; Chat
`https://open.bigmodel.cn/api/paas/v4` adds `/chat/completions`; Messages
`https://open.bigmodel.cn/api/anthropic` adds `/v1/messages`. Responses/Chat use
Bearer; Messages uses `x-api-key`. The SDK's observed Messages `?beta=true`
query is retained; no Claude beta/browser headers are claimed.

This example uses a **standard API Key** whose purpose the Host has checked.
After the2026-10-16 key classification change, new standard/personal/team keys
have their declared purposes. Historical compatible-key behavior depends on
endpoint/account history. An endpoint is not proof of a key's billing source,
Coding Plan eligibility, permissions or invoice. barness-ai never infers a key
category from its bytes. See [captured primary sources](../../../.scratch/barness-ai-glm/research.md).

## Local program

The program chooses exactly one source; the library never reads conventional
provider environment variables. `localassembly` reads this explicitly named
source once during assembly. `Operations.Scope()` generates a fresh RequestID.

```go
bindings, err := glmhost.Bindings(glmhost.Authorization{
    ProviderID: ai.ProviderZhipu, Purpose: glmhost.Standard,
    TenantID: "local-tenant", AccountScopeID: "approved-account",
    CredentialRef: "standard-glm", Version: "local",
})
if err != nil { return err }
local, err := localassembly.OpenOperations(localassembly.OperationsConfig{
    TenantID: "local-tenant", Bindings: bindings,
    Keys: map[string]localassembly.KeySource{
        "standard-glm": {EnvVar: "MY_CHINA_STANDARD_GLM_KEY"}, // or File, exactly one
    },
    Policy: localassembly.LocalPolicy(), // review its documented local design load
})
if err != nil { return err }
request := ai.Request{Messages: []ai.Message{ai.UserText("Explain this code.")}}
turn, err := glmhost.Turn(ctx, local.Client, local.Scope(),
    ai.Target{BindingID: "glm-responses", ModelID: "glm-5.3"}, request,
    ai.APIOpenAIResponses, glmhost.FullStream, ai.Hooks{})
// Handle typed ai.Error and turn.Result, even on failure; keep content private.
```

Use `glm-chat` with `APIOpenAICompletions`, or `glm-messages` with
`APIAnthropicMessages`. `Turn` demonstrates `FullStream`, `FullComplete`,
`SimpleStream`, `SimpleComplete`, and `WithHooks`. Full calls use the actual
`ChatOptions`/`ResponsesOptions`/`AnthropicOptions`; simple calls use
`SimpleOptions`. All cap output at4096 and leave thinking omitted. Streaming
consumes envelopes, awaits Result and closes the stream on every path. This
small demonstration retains envelopes; a production UI should render them as
they arrive under the same bounded queue policy.

The immutable local resolver stamps Binding/Credential versions `local` and
preserves `Enabled`. Reassemble after local configuration/key changes; use
`OpenOperations` for disablement. The single-binding `localassembly.Open`
convenience example always enables its binding.

## Cloud Host

Authenticate the principal outside this package. Never accept tenant, endpoint,
credential reference, or hooks as authority from a public DTO. Persist each
returned binding under `(TenantID, BindingID)` with an actual backend version:

```go
authorization := glmhost.Authorization{
    ProviderID: ai.ProviderZhipu, Purpose: glmhost.Standard,
    TenantID: principal.TenantID, AccountScopeID: approvedAccount,
    CredentialRef: "standard-glm", Version: "binding-v7",
}
bindings, err := glmhost.Bindings(authorization)
if err != nil { return err }
// Load this secret-free record from the same backend transaction as the
// authorization. Its BindingVersion is stored independently of incoming b.
record := glmhost.CredentialRecord{
    ProviderID: ai.ProviderZhipu, TenantID: principal.TenantID,
    CredentialRef: "standard-glm", AccountScopeID: approvedAccount,
    Purpose: glmhost.Standard, Active: true,
    Version: "secret-v3", BindingVersion: "binding-v7",
    AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses,
        ai.APIAnthropicMessages},
}
secretResolver := glmhost.Credentials{
    Authorization: authorization, Record: record,
    ReadSecret: func(ctx context.Context, ref string) (ai.Secret, error) {
        return tenantSecretStore.Read(ctx, principal.TenantID, ref)
    },
}
// Ownership, purpose, API and version checks precede tenantSecretStore.Read.
client, err := ai.NewClient(ai.Config{
    Policy: hostintegration.CloudInteractivePolicy(),
    Bindings: tenantBindingResolver, Credentials: secretResolver,
    Observer: metadataOnlyObserver, // no prompt, reasoning or credential logging
})
if err != nil { return err }
scope := ai.CallScope{TenantID: principal.TenantID, ActorID: principal.ActorID, RequestID: requestid.New()}
turn, err := glmhost.Turn(ctx, client, scope,
    ai.Target{BindingID: "glm-chat", ModelID: "glm-5.2"}, request,
    ai.APIOpenAICompletions, glmhost.SimpleComplete,
    ai.Hooks{TransformHeaders: func(ctx context.Context, scope ai.CallScope, h http.Header) error {
        h.Set("X-Application-Trace", safeTraceID) // trusted, secret-free metadata
        return nil
    }})
```

A shared CredentialRef is deliberate only for the same authorized tenant,
Provider, account and purpose, with each API explicitly present in the credential
record. China uses the formal `BuiltinCatalog`; Z.AI requires its independently
reviewed Host candidate catalog and has no new live qualification in this feature. A bounded distributed Admission can be supplied by
the Host. Neither resolver nor hooks select another endpoint on failure.
`ResourcePolicy` values must fit the deployment's documented load; the cloud
and local examples have different measured load assumptions.

The Host executes tools. Reuse `toolloop.Loop{Client: client, Target: target,
Scope: freshAuthenticatedScope, MaxTurns: 4}` with explicit declarations and
Host functions. It calls `ValidateToolCall`, runs functions only after a
successful `StopReasonToolUse`, appends the original assistant and each
`ToolResultText`, and starts a fresh Logical Call. No model function executes
inside barness-ai. Persist native history only with the separate trusted
Host envelope; untrusted JSON signatures never grant replay authority.

## Capabilities and recovery

All six identities are text-only. Chat supports ordinary function tools and
interleaved `reasoning_content` history. GLM-5.3 cannot disable thinking;
full supports low/high/max, simple clamps to its declared subset. GLM-5.2
has its independent maps. Responses is stateless (`store:false`), preserves
object/array reasoning separately, and supports the verified low/high/max
wire subset; accepted low is not a measurement of thinking intensity.
Messages preserves actual signed thinking and the verified signature-field
omission case. Canonical Messages thinking maps are unsupported; only5.2's
explicit enabled1024/max4096 shape is verified, while5.3 keeps defaults.
Strict/required/forced tools, hosted tools, Claude cache/betas/effort,
OpenAI saved sessions/encrypted reasoning/summary/retention guarantees and
vision are outside this implemented subset. See [contract](../../../docs/barness-ai/contract.md).

Usage presence distinguishes unreported from reported zero. Reasoning is
already part of output. Current prices are fixed **USD per million tokens**
from each model's CNY8 input/cache-write,28 output,2 cache-read rates and
CFETS2026-10-16 FX6.7220:1.1901219875 /4.1654269563 /0.2975304969 respectively.
Cache-write input pricing is an explicit source-based inference; no Anthropic
one-hour premium is invented. Partial cancellation may lack usage, so known
estimated cost is a lower bound, never proof of free cancellation or an invoice.

To disable a route, atomically set its binding `Enabled=false` with a new
version. New calls return `tenant_denied/binding` with zero credential reads
and Attempts. Already sent calls retain their pinned version and consumption;
other tenants/routes remain independent. Restoring authorization is another
explicit version and fresh call. Whole-software rollback restores the complete
previous tree/catalog/capabilities/matrix/evidence, not just six model entries.
See [ticket13 proof](../../../.scratch/barness-ai-glm/evidence/ticket-13/README.md).

Repeat the no-key public example proof with:

```sh
go test ./ai/e2e -run '^TestGLMHost' -count=1
```

The retained artifact includes both models/three APIs/four entries/hooks,
Host tools, disable/in-flight/other-tenant/restore behavior, own final captures,
formal catalog/price identity, usage, Observer lifecycle and resource cleanup.


## New market configuration and secret ownership

Z.AI uses `ProviderZAI` and its own `zai-chat`, `zai-responses`, `zai-messages`
bindings, credential records and catalog identities. Its three official bases use
`https://api.z.ai` with the same documented protocol paths. No identity is inferred
from a URL. `Endpoints` may override exact official bases; an explicit
`TrustedGateway: true` permits reviewed nonofficial HTTPS or loopback test targets.
URL credentials, query parameters, fragments, another official market, and wrong
protocol paths fail assembly. HTTP loopback also requires Client opt-in. Gateways
never change service identity or authorize a fallback.

`glmhost.Credentials` implements the public CredentialResolver with separate
`Authorization`, secret-free `CredentialRecord`, and `ReadSecret(ctx, ref)` callback.
The record carries Provider, Tenant, AccountScope, Purpose, allowed APIs, Active,
credential Version and independently stored BindingVersion. All ownership and
purpose checks precede the callback; mismatch returns ErrAccessDenied, while
version conflict returns ErrSnapshotConflict before reading the secret. Never
construct record metadata from the Binding handed to the resolver. Resolve current
backend snapshots atomically; these example values are immutable during use.
The callback must read exactly the selected tenant-owned record, never search keys.

The candidate data and source gaps are documented in
[regional candidates](../../internal/testkit/regional/README.md). Before software
rollback disable new-market bindings and remove their catalog identities from the
snapshot loaded by the old process. Retain original history without rewriting its
Provider. New-market ordinary proofs: `go test ./ai/e2e -count=1 -run '^TestRegionalProvider'`.
