# GLM Messages Host candidate

`MessagesCatalog`, `MessagesBinding` and `MessagesEndpoint` assemble a separate
Host candidate for China Zhipu × Messages. This is basic text support under
verification, not a built-in model or full Messages support claim. Chat and
Responses entries are not used to derive its wire behavior.

The [official Claude compatibility guide](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction.md)
retrieved 2026-10-17 independently shows `glm-5.3` and `glm-5.2` with the base
`https://open.bigmodel.cn/api/anthropic`, x-api-key and basic Messages fields.
Its SHA-256 and limitations are in
`.scratch/barness-ai-glm/implementation-research.md`. Both model pages state
text input/output, 1M context and 128K output. This candidate uses conservative
Host limits of 1000000 and 128000; no Messages schema supplies an exact integer
output cap. A paid request remains bounded to 4096 by its caller. Prices here
are explicitly synthetic fixture USD rates and never a billing assertion.

Own-protocol evidence: the authorized `02-basic-sdk-query` bundle under
`.evidence/barness-ai-glm/protocol-exploration/messages/` captured two successful
public Client calls, one per model. Locked anthropic-sdk-go v1.75.0 appends
`/v1/messages?beta=true` even without a beta header. The endpoint accepted that
query, `anthropic-version:2023-06-01` and SDK metadata. It does not prove any
beta feature. Basic capture used user string content; `03-tools-thinking`
separately accepted the existing adapter's content-block arrays and signed
assistant history, without generated cache/eager/display/beta fields.

`ai/e2e/testdata/glm-messages/*-response.sse` are byte-for-byte copies of each
model's basic capture. `text.json` pins their SHA-256 and separately recorded
content/usage. They prove 5.3 can return signed thinking by default; the 5.2
sample returned text alone. `04-default-unsigned` retained a separate complex
5.2 same-prompt control: omitted thinking again returned only text. Absence of
visible reasoning does not describe internal computation.

Use `Stream`, `Complete`, `StreamSimple` or `CompleteSimple` with omitted
reasoning. Nil/full omitted thinking and simple omitted reasoning preserve
wire omission, including GLM-5.3. Messages ignores model/call SamplingParams
and per-level maps. Full explicit thinking, canonical levels/budgets, display,
effort, cache hints and metadata are not exposed by this basic ticket; they
fail before send instead of being silently mapped to Claude fields. Cache
`none` is a legal omission preference. A single real 5.2 enabled/budget1024
shape exists, but does not establish a budget range or canonical strength map.
Ticket 07 owns the next bounded extension.

No automatic cache/TTL, eager tool streaming, browser header, thinking display,
beta, adaptive/block binding, managed effort, fallback or protocol conversion
is added. A trusted callback may retain/modify the verified basic body and
ordinary headers; final validation rejects unverified fields or beta/browser
headers before any Attempt. Existing authorization still protects model,
credentials and provider-side references.

Actual usage begins with input/output zero in `message_start`, then reports
actual input/output and cache_read=0 in `message_delta`. `cache_creation` is
absent, so UsageReporting remains partial, and reasoning is not fabricated.
The response's X-Request-Id is preserved in Attempts and errors; native
Anthropic retains Request-Id. Missing stop reason/message_stop and in-stream
error fail with one terminal result through the existing lifecycle.

```go
catalog := glm.MessagesCatalog() // explicit trusted Host snapshot
binding := glm.MessagesBinding("local", "local-cn-standard", "glm-standard",
    glm.MessagesEndpoint)
// Store binding in the tenant-aware resolver. The Host loads the standard
// API key from its secret source into a Credential for that tenant/account.
// Construct ai.Client with &catalog, both resolvers and a finite ResourcePolicy.
result, err := client.CompleteSimple(ctx,
    ai.CallScope{TenantID: "local", RequestID: freshRequestID},
    ai.Target{BindingID: binding.BindingID, ModelID: "glm-5.3"},
    ai.Request{Messages: []ai.Message{ai.UserText("Calculate 17 + 25.")}},
    ai.SimpleOptions{MaxTokens: ai.Value(4096)},
)
```

The library never loads `.env`. The Host must select a China standard API key
and fixed account scope. Changing an endpoint or using historical-compatible
keys does not prove plan entitlement, quota or billing. Ticket 11 owns formal
production-path live acceptance; ticket 12 owns any built-in registration.
