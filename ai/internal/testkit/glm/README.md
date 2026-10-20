# GLM Host candidates

`ChatCatalog` and `ChatBinding` are explicit Host assembly for offline acceptance and later bounded live evidence. They do not register built-in models. Every call uses the Host's Credential Snapshot; the core never reads `.env` or SDK profiles. Use China standard API keys at `ChatEndpoint`; Coding Plan acceptance and account billing are not implied.

Retrieved 2026-10-17: [Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json), [thinking guide](https://docs.bigmodel.cn/cn/guide/capabilities/thinking.md), [GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3.md), [GLM-5.2](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.2.md). Source hashes and exact schema locations are in `.scratch/barness-ai-glm/implementation-research.md`.

Both models independently state text input/output, 1M context and 128K output. `1000000` is a barness conservative context limit, not an exact vendor integer. Chat schema explicitly caps `max_tokens` at `131072`. GLM-5.3 accepts low/high/max with mandatory thinking; GLM-5.2 separately documents none/minimal as abandoning thinking, low/medium/high as high and xhigh/max as max. Omitted levels retain the wire default, independently of canonical sampling profile selection.

All rates in `chat-catalog.json` are **synthetic fixture rates**, not provider prices, currency conversions or billing assertions. Paid evidence must replace them with its own sourced, fixed price snapshot. Other protocols require independent candidate metadata, never a copy with the API field changed.

## Standard key Host configuration

The Host loads the key from its named secret source, authorizes the tenant and
returns these snapshots from its resolvers. The core does not open `.env`.
Keep a new standard API key on the standard endpoint; new personal/team
Coding Plan keys have a different purpose under the 2026-10-16 key rules.
Historical compatible keys can have account/endpoint-dependent behavior;
neither their string shape nor this example proves plan entitlement or billing.

```go
binding := ai.Binding{
    TenantID: "local", BindingID: "glm-chat", Version: "host-v1",
    Operation: ai.OperationChat, Enabled: true,
    ProviderID: ai.ProviderZhipu, API: ai.APIOpenAICompletions,
    Endpoint: "https://open.bigmodel.cn/api/paas/v4", // SDK appends /chat/completions
    AuthKind: ai.AuthAPIKey, AccountScopeID: "local-cn-standard",
    CredentialRef: "glm-standard", AllowedModels: []string{"glm-5.3", "glm-5.2"},
    // Retry zero: no implicit retry or account/endpoint fallback.
}
credential := ai.Credential{
    OwnerTenantID: "local", CredentialID: "glm-standard", Version: "key-v1",
    BindingVersion: "host-v1", AccountScopeID: "local-cn-standard",
    Active: true, APIKey: ai.NewSecret(standardKeyLoadedByHost),
}
// Store binding and credential in the Host's tenant-aware resolvers.
// catalog is a trusted, versioned ai.Catalog with independent Chat entries;
// policy is the Host's finite ai.ResourcePolicy (candidate CallTimeout <= 120s).
client, err := ai.NewClient(ai.Config{
    Catalog: &catalog, Policy: policy, Bindings: hostBindings, Credentials: hostCredentials,
})
// Handle err before calling. Each logical call gets a fresh RequestID.
result, err := client.CompleteSimple(ctx,
    ai.CallScope{TenantID: "local", RequestID: freshRequestID},
    ai.Target{BindingID: binding.BindingID, ModelID: "glm-5.3"},
    ai.Request{Messages: []ai.Message{ai.UserText("Say hello.")}},
    ai.SimpleOptions{MaxTokens: ai.Value(4096), Reasoning: ai.ThinkingLow},
)
```

`Stream`, `Complete`, `StreamSimple`, `CompleteSimple` and `WithHooks` use the
same binding. Nil/full omitted effort retains the service default; it does
not disable 5.3. Explicit full unsupported levels fail before credentials.
Simple minimal/medium/xhigh clamp to low/high/max on the 5.3 candidate.
Final `max_tokens` must be an integer from 1 to the smaller of the model cap
and Chat schema cap 131072. `max_completion_tokens` is refused. Named null/0
retain omission semantics; actual wire null/0 is invalid.
Tool round trips and formal built-in registration belong to tickets 03/12.

The later paid candidate entry must select its route explicitly under the live
build tag and opt-in, inject only that route's key, reserve 24 Logical Calls /
32 Attempts, cap each call at 4096 output tokens / 120 seconds and the process
at 1800 seconds. `ai/live/glm_assembly_test.go` shares Client assembly with the
existing harness while keeping its policy separate from the legacy runner.
Its controlled test never makes a paid call; ticket 11 owns the isolated entry,
switches, budget reservation and report. Missing credentials remain NOT_RUN.

## Independent Responses candidate

`ResponsesCatalog`, `ResponsesBinding` and `ResponsesEndpoint` are independent
Responses definitions, sourced from the China Responses schema, model pages
and the separately authorized basic/tool exploration. They are not generated
from Chat entries. The base is `https://open.bigmodel.cn/api/v1`; the SDK appends
`/responses`, using Bearer authentication. Output schema accepts 1..131072;
context 1000000 is an explicitly conservative limit. Each entry is text only.

The generic Responses effort schema describes all seven wire strings and maps
low/medium to high; the 5.3 model page separately says low/high/max only. The
candidate follows the spec's conservative 5.3 low/high/max contract and keeps
omitted effort omitted. Actual low computation intensity is not established by
Chat's semantics. The 5.2 map is independently based on the Responses generic
schema (none/minimal abandon, low/medium/high→high, xhigh/max→max), not copied
from Chat. Candidate costs are synthetic fixture rates, never RMB/USD pricing.
Both early real basic streams used arrays; object-shape regression fixtures
are official-schema-derived synthetic evidence. Native reasoning keeps its
original shape, separately from the SDK parsing copy. Ticket05 owns history
request conversion and tools; ticket11 owns paid production acceptance.
