# MiniMax China Host example

[简体中文](README.zh-CN.md). Status on 2026-10-18: the three adapters and all six
candidate identities have offline tests. **M3 formal support is pending ticket10**;
M3.1 has no confirmed M Plan entitlement or sourced production price here. Neither
is promoted by this example. The current [candidate catalog](../../internal/testkit/minimax/candidate.json)
uses explicitly synthetic test rates and must not be presented as a real estimate.
The [formal work](../../../.scratch/barness-ai-minimax/issues/10-formal-catalog-support.md)
must supply each M3 protocol's own real evidence, sourced USD prices, current
source/catalog hashes and matrix/release checks. M3.1 remains separate candidates.

## Authorize account purpose before assembly

`minimaxhost.Bindings` accepts trusted Host configuration, not a client claim or
an entitlement discovery result. It constructs three separately revocable bindings.
This example grants standard M3, or explicitly qualified M Plan M3.1, one purpose
at a time. A different entitled model set requires an explicit reviewed Host grant.

```go
bindings, err := minimaxhost.Bindings(minimaxhost.Authorization{
    TenantID: "tenant-from-authentication", AccountScopeID: "actual-provider-account",
    CredentialRef: "minimax-standard-vault-reference", Version: "config-17",
    Purpose: minimaxhost.Standard,
})
// Check err, then store each binding in the tenant-scoped resolver.
```

| Binding suffix | Protocol | Base endpoint | Final resource | Authentication |
| --- | --- | --- | --- | --- |
| `chat` | `openai-completions` | `https://api.minimax.cn/v1` | `/v1/chat/completions` | Bearer |
| `responses` | `openai-responses` | `https://api.minimax.cn/v1` | `/v1/responses` | Bearer |
| `messages` | `anthropic-messages` | `https://api.minimax.cn/anthropic` | `/anthropic/v1/messages` | x-api-key |

Binding IDs include `minimax-standard-` or `minimax-m-plan-`. The Host keeps the two
purposes in separate credential references and grants; it may share one reference
among the same authorized tenant/account/purpose's three protocols. AccountScopeID
represents the real provider account, not a key version or arbitrary user claim.
No key-format guessing, Coding endpoint, region change, key/model/protocol fallback
or conversion of subscription use into standard balance use occurs. M Plan access
for a local interactive client does not establish cloud production authorization.

Messages uses the SDK's documented x-api-key route. SDK `?beta=true` and version
headers are transport facts; they enable no Claude beta header/body feature. Never
add a second Authorization/x-api-key credential through a header hook.

## Local configuration

Reuse [localassembly](../localassembly/localassembly.go): the local program chooses
one key source and passes an explicit reviewed catalog. The library/SDK never reads
vendor environment keys, dotenv, profiles or a default endpoint. Repeat the same
assembly for another explicitly selected binding; each client owns its key snapshot.

```go
// reviewedCatalog is loaded and validated by trusted Host configuration.
// Until formal promotion, it must be clearly labelled a controlled candidate.
binding := bindings[0] // This example's chat route; select API explicitly.
local, err := localassembly.Open(localassembly.Config{
    TenantID: binding.TenantID, Binding: binding, Catalog: &reviewedCatalog,
    Key: localassembly.KeySource{EnvVar: "HOST_MINIMAX_STANDARD_KEY"},
    Policy: localassembly.LocalPolicy(),
})
if err != nil { return err }
result, err := local.Client.Complete(ctx, local.Scope(),
    local.Target("MiniMax-M3"), ai.Request{Messages: []ai.Message{
        ai.UserText("Reply with one short sentence."),
    }}, ai.ChatOptions{MaxTokens: ai.Value(512)})
```

For an actually qualified M Plan candidate, construct separate bindings with
`Purpose: minimaxhost.MPlan`, a separate account/credential reference, its reviewed
candidate catalog and an explicitly selected source such as `HOST_MINIMAX_MPLAN_KEY`.
The core has no credential-purpose enum; the Host owns that grant. `Catalog:nil`
retains the built-in catalog and does not make an unpromoted MiniMax model available.
`AllowLoopbackHTTP` and controlled transports belong only in offline test assembly.
Local Open intentionally creates an immutable enabled assembly and assigns its local
credential/version. To revoke a local route, stop admitting new turns and retire that
assembly. For mutable Enabled/version revocation, use the versioned resolvers below;
passing Enabled=false into local Open does not configure a disabled binding.

## Cloud configuration

Use one Client with tenant-scoped, versioned resolvers; pair Credential.BindingVersion
with the exact binding snapshot. Store/retrieve credentials under tenant and reference,
check ownership/account/active status, and return a new value. Do not use a global key.

```go
client, err := ai.NewClient(ai.Config{
    Policy: hostintegration.CloudInteractivePolicy(),
    Bindings: trustedBindings, Credentials: trustedCredentials,
    Catalog: &reviewedCatalog, Observer: metadataObserver,
})
// Check err. Both Client and HookedClient satisfy hostintegration.ModelClient.
bounded := client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context,
    _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
    if p.ProviderID == ai.ProviderMiniMaxCN {
        field := map[ai.API]string{
            ai.APIOpenAICompletions: "max_completion_tokens",
            ai.APIOpenAIResponses: "max_output_tokens",
            ai.APIAnthropicMessages: "max_tokens",
        }[p.API]
        if field != "" { p.Body[field] = 512 }
    }
    return ai.KeepPayload(), nil
}})
cloud := hostintegration.New(bounded, tenantSessionStore, requestid.New)
// Principal comes from verified authentication, never request JSON.
who := hostintegration.Principal{TenantID: authenticatedTenant, ActorID: authenticatedActor}
session, err := cloud.OpenSession(ctx, who)
// Check err; each handler also checks its own app permissions and downstream sink.
result, err := cloud.Turn(ctx, who, hostintegration.TurnRequest{
    SessionID: session, BindingID: bindings[0].BindingID, ModelID: "MiniMax-M3",
    Prompt: "Reply briefly.",
}, downstreamSink)
```

The sample MemoryStore is for demonstration. A durable store must validate record
integrity and tenant/session ownership before restoring native state. The cloud Host
rejects forged tenant claims and other tenants' sessions before Client calls, forwards
cancellation, and sends attributed EventEnvelopes. A real UI handles loading, partial
output, empty/error/length outcomes, retry feedback and accessible controls.

## Thinking and entry styles

| Model × protocol | Omitted reasoning | Explicit support |
| --- | --- | --- |
| M3 Chat | on | binary off/high; high only means enabled |
| M3 Responses | off | `high` enables; native `none` disables |
| M3 Messages | off | ThinkingEnabled adaptive/disabled; Effort is rejected |
| M3.1 Chat | on, default max | low/medium/high/xhigh/max; cannot disable |
| M3.1 Responses | on, default max | low/medium/high/xhigh/max; cannot disable |
| M3.1 Messages | on, default max | same five Efforts, including effort-only full options; cannot disable |

`Model.DefaultThinkingLevel` chooses omitted sampling defaults, not a wire effort or
an explicit user choice. ThinkingSwitchOnly distinguishes binary M3 independently
of RequiresThinking. Narrowing M3.1 to high-only still retains effort; requiring M3
thinking still does not enable depth. Model declarations may only narrow implemented
provider shapes. All six identities are independently declared; this is not a promise
of measured reasoning intensity or automatic cross-protocol migration.

Choose full options matching the binding: `ChatOptions`, `ResponsesOptions`, or
`AnthropicOptions`; another protocol's options fail. Use MaxTokens=512 in examples.
For explicit M3 on, use Chat/Responses ReasoningEffort=ThinkingHigh or Messages
ThinkingEnabled=Value(true). M3 off uses Chat SamplingParams `thinking:{type:disabled}`,
Responses SamplingParams `reasoning:{effort:none}`, or Messages ThinkingEnabled=false.
M3.1 rejects off/none and full minimal; full Messages EffortLow without ThinkingEnabled
still sends the effort. SimpleOptions Reasoning empty keeps the service default;
minimal/low clamp upward to high for M3 and minimal clamps to low for M3.1. It has no
new off enum. Messages never applies Claude token-budget arithmetic.

`Complete`/`Stream` use full options; `CompleteSimple`/`StreamSimple` use SimpleOptions.
Every new logical call needs a globally unique RequestID (requestid.New), even tool
continuations and manual retries. Streaming consumers read Next/Event/Envelope, then
Result, and always close the Stream; Complete consumes the same streaming protocol.
WithHooks supports all four entries. `toolloop.Loop` demonstrates a bounded Host tool
round trip: validate only complete tool calls, execute them in the Host, append the
entire assistant message and paired results, then make a fresh call. Do not discard
thinking or native state or execute a length/error/aborted turn's partial calls.

## Request subset and hooks

This release connects the text/tool subset, not the vendor's whole multimodal model.
Unaccepted images follow existing placeholder/history conversion rules; payload hooks
cannot inject multimodal containers. Only default/auto/none custom tools are supported;
required/any/named forcing, disabling parallel tools, strict/hosted tools and ignored
stop/resource controls are rejected. Only omitted/standard service tier is supported.
Messages rejects explicit thinking budget/display/interleaved/beta/binding controls,
eager streaming, browser access, top_k and stop_sequences. Explicit Temperature is
retained during adaptive thinking. Chat/Messages temperature is [0,2], top_p [0,1];
Responses uses (0,1] for both pending conflicting zero documentation. Established named
null/omission rules still apply; unknown null is not a valid number. Final output limits
are exact integers 1..model maximum, with no 16-token minimum. Default and call sampling
merge before payload hooks; no second profile selection occurs afterward. Messages
ignores all SamplingParams/profile maps.

Hooks run headers → payload → response once per logical call, outside initial-request
retries. They cannot change authorized identity, authentication/target/model or native
references. Final fields are revalidated after payload changes; an invalid named control
is rejected before credentials and cannot be repaired by a later hook. Keep hook traces
non-secret. Binding retries default to zero; explicit finite policies never replay a
started stream or switch account, purpose, model or protocol.

## Caching, cost, trust and diagnostics

Chat/Messages use automatic provider caching. Messages short/long do not add cache_control
or promise TTL; explicit cache_control/ttl is rejected. Responses with SessionID and
retention other than none derives a tenant/account/session route key. No SessionID or
none emits no key. Short/long do not promise retention or cache hits; no raw session
or OpenAI affinity headers are sent, and a routing key is not a claim of isolated
provider cache storage. Do not log native/cache state.

Usage fields retain presence: missing/partial/cancelled usage is not reported zero.
Messages input excludes separately reported cache read/write; total input is their sum.
Chat/Responses cached input is split out of input totals. Reasoning is an output subset
and is never billed twice by the estimator. CostRates use USD per million tokens;
UsageCost is USD, with CatalogVersion/Hash identifying the fixed Price Snapshot.
Formal M3 rates need dated CNY→USD sources and the inclusive 512000-input tier boundary;
unknown cache-write or M3.1 rates must not be replaced with zero or copied prices.
Current synthetic candidate rates are mechanism fixtures, not an M Plan/standard quote.
Cancellation ends local work, not a guarantee of free or remotely cancelled generation.

Store NativeStateEnvelope beside the full message. In-process Results already carry
trusted provenance; JSON does not. After authenticating storage integrity and ownership,
call TrustNativeState with the trusted scope and saved envelope. Same tenant/account/
provider/API/model can replay native state; mismatches downgrade with explicit metadata.
Chat keeps its own reasoning entries, Responses uses its documented lossless candidate
input projection, and Messages currently permits trusted signed thinking. Unverified
opaque/redacted replay fails rather than becoming ordinary text; no GLM unsigned exception
is inherited. Each live continuation shape still needs its own ticket10 evidence.

Use errors.Is/As with Error.Code/Phase and actual HTTPStatus/ProviderRequestID when
reported; retain RequestID/AttemptID attribution separately. Missing provider IDs stay
missing. Never fabricate a trace ID from Host RequestID or another provider's header.
Observer records contain metadata only and are appropriate for shared operational logs;
Result, prompts, tool arguments, reasoning, native state, raw error text and secrets are
not. Metrics should avoid unbounded actor/job/request IDs as labels.

## Revoke and restore

Set just the selected Binding.Enabled=false and publish a new version atomically. New
calls have zero credential reads/Attempts/HTTP requests; other tenants, purposes and
routes remain usable. Already sent calls retain pinned snapshots and known consumption.
Restoring permission is another version and fresh call, with no fallback or double write.

Whole software rollback is a separate isolated restoration of baseline
`ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`, tree
`b9798a4e96dfb08cbfcce45c182915589b42eed1`: all 7039 paths/blobs/modes, runtime, catalog,
prices, matrix, diff/release contracts and dependencies. It verifies the old public
provider route and all six absent MiniMax identities before deleting its temporary tree.
Caller/user root, modified files and untracked bytes are preserved. Run into a fresh path:

```sh
python3 .scratch/barness-ai-minimax/evidence/ticket-08/rollback.py .evidence/minimax-rollback-review
go test ./ai/e2e -run '^TestMiniMax(Host|FullRollback)' -count=1
```

[Ticket08 evidence](../../../.scratch/barness-ai-minimax/evidence/ticket-08/README.md)
contains replayable configuration/software proofs and redaction audits. It does not
renew historical real-provider evidence or bypass the final ticket09/10 release gates.
