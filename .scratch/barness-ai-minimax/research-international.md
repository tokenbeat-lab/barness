# MiniMax international integration research

Researched: **2026-10-18, 08:56–08:57 Asia/Shanghai** (2026-10-18T00:56–00:57Z). Scope: official `platform.minimax.io` documentation, compared with existing China snapshots at `/Users/cyber/RestoX/barness/ai/live/testdata/minimax` and `.scratch/barness-ai-minimax/research.md`. No credential or `.env` access, no model/entitlement/quota HTTP requests, no repository writes. The existing `MINIMAX_KEY` is authorized by the user for the international integration; its account/purpose/entitlement remains uninspected in this research.

Retained raw snapshots: [official source archive](evidence/ticket-11/international-official-sources-20261018.tar.gz), [archive provenance](evidence/ticket-11/international-source-archive.json); original research workspace `/tmp/barness-minimax-international-sources/`; exact URLs, fetch timestamps, content types, Last-Modified headers and SHA-256 are in `sources.json`. Structured facts are in `facts.json`. China comparison details are in `china-schema-comparison.json`.

## 1. Region, account and operator identity

Use a distinct **international `minimax` Provider identity**, international account scope, and international endpoint bindings. This is an implementation recommendation matching the existing spec's requirement to distinguish regional service operators; it is not a claim that a URL identifies a user's account or physical processing location.

Official international prerequisites register/login at `platform.minimax.io`, create ordinary pay-as-you-go API Keys in that platform's API Keys page, and obtain Subscription Keys separately through M Plan. The international Terms of Service (effective March 30, 2026) identify **Nanonoble Pte Ltd.** as the service provider; the registered address is in Singapore. This identifies the international contract operator, not a region guarantee for every request. [Prerequisites](https://platform.minimax.io/docs/guides/quickstart-preparation), [API key FAQ](https://platform.minimax.io/docs/faq/about-apis), [Terms](https://platform.minimax.io/protocol/terms-of-service)

No fetched official source states China and international accounts, balances, keys or M Plan entitlements are interchangeable. Existing China authentication failure does not prove international failure, and changing URL must not silently retain a China provider/account identity. The user-directed international route is explicit; errors must not cause automatic region, credential, model, or billing-purpose fallback. Account and credential purpose should be asserted by the Host and subsequently supported by its own live evidence.

## 2. Exact generation endpoints and authentication

| API | SDK base URL | Final generation request | Official authentication |
| --- | --- | --- | --- |
| `openai-responses` | `https://api.minimax.io/v1` | `POST https://api.minimax.io/v1/responses` | `Authorization: Bearer <API_KEY>` |
| `openai-completions` | `https://api.minimax.io/v1` | `POST https://api.minimax.io/v1/chat/completions` | `Authorization: Bearer <API_KEY>` |
| `anthropic-messages` | `https://api.minimax.io/anthropic` | `POST https://api.minimax.io/anthropic/v1/messages` | Either Bearer or `x-api-key: <API_KEY>` accepted. Bearer recommended; if both sent, Bearer takes precedence. |

Sources: [OpenAI SDK](https://platform.minimax.io/docs/api-reference/text-openai-api), [Anthropic SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api), [Responses OpenAPI](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json), [Chat OpenAPI](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-openai.json), [Messages OpenAPI](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json).

All JSON bodies use `Content-Type: application/json`. Existing SDK `x-api-key` behavior can remain for Messages because it is explicitly supported; send only the chosen authentication header. Do not append `/v1` to the Messages SDK base URL. No dedicated Coding/M Plan generation base is required: official M Plan “Other Tools” specifies the same `/v1` and `/anthropic` bases. [M Plan tool configuration](https://platform.minimax.io/docs/m-plan/other-tools)

Documented adjacent operations are `POST /v1/responses/input_tokens` and `POST /anthropic/v1/messages/count_tokens`; these are not hidden prerequisites for a generation call. No requests to them were made.

## 3. Models, availability and billing purpose

| Model | Protocol documentation | Context | Availability/entitlement |
| --- | --- | --- | --- |
| `MiniMax-M3` | Listed by Chat and Messages SDKs; Responses documents its behavior | 1,000,000 tokens | Published international pay-as-you-go USD pricing exists. Specific key/account access is untested. |
| `MiniMax-M3.1-Flash-Preview` | Explicitly documented in all three protocols | 1,000,000 tokens | Current docs explicitly limit access to **M Plan and MiniMax Code**. Existing key's entitlement remains unconfirmed. |

[Models overview](https://platform.minimax.io/docs/guides/models-intro), [OpenAI SDK](https://platform.minimax.io/docs/api-reference/text-openai-api), [Anthropic SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api), [Responses](https://platform.minimax.io/docs/api-reference/responses-create)

International M Plan has Go, Explore and Build tiers; all include M3.1 Flash Preview. Subscription Keys consume M Plan usage and eligible credit packs, while ordinary API Keys consume account balance. The two key types cannot be used interchangeably. M Plan is for individual interactive developer use; official FAQ recommends pay-as-you-go for production. Model identity, subscription purpose and user entitlement must remain separate dimensions. No public per-token M3.1 price was found in the international pricing overview or explicit-cache pricing page; do not substitute M3 prices or infer subscription deductions from estimated standard API cost. [M Plan FAQ](https://platform.minimax.io/docs/m-plan/faq), [M Plan quick start](https://platform.minimax.io/docs/m-plan/quickstart)

The international models overview repeats the China migration schedule: M2.5/M2.7 unavailable to new users on **2026-10-19** and retired on **2026-11-18**; plan requests under legacy IDs automatically route to M3.1 from October 19. The provider catalog should not use legacy model aliases as proof of M3/M3.1 semantics. [Models overview](https://platform.minimax.io/docs/guides/models-intro)

## 4. International USD pricing and token accounting

The international pricing overview explicitly states **all prices are USD**. The following are current displayed prices per **1,000,000 tokens**, after the permanent 50% discount shown on that page; strikethrough prices are not current rates. No FX conversion is involved. [Pricing overview](https://platform.minimax.io/docs/pricing/overview#models/text)

| Model/tier | Total input length | Input | Output | Cache read | Separately classified cache write |
| --- | --- | ---: | ---: | ---: | --- |
| M3 standard | ≤512k input tokens | $0.30 | $1.20 | $0.06 | Unpriced for M3 in fetched sources |
| M3 standard | >512k input tokens | $0.60 | $2.40 | $0.12 | Unpriced for M3 in fetched sources |
| M3 priority | ≤512k input tokens | $0.45 | $1.80 | $0.09 | Unpriced for M3 in fetched sources |
| M3 priority | >512k input tokens | $0.90 | $3.60 | $0.18 | Unpriced for M3 in fetched sources |
| M3.1 any | — | Unknown | Unknown | Unknown | Unknown |

Priority is 1.5× the corresponding standard length tier. If first delivery supports only standard, explicit priority should stay rejected until its own pricing and behavior are supported.

**Input tier selection includes cache-hit tokens.** The passive prompt-cache page expressly says long-context M3 pricing applies above 512k input tokens, including cache hits. Thus using only uncached input to select the tier is incorrect. [Passive cache](https://platform.minimax.io/docs/api-reference/text-prompt-caching)

**Exact integer threshold caveat:** fetched official price/cache pages write `512k`, without defining `512000` or `524288` for pricing. Keep the existing spec's `>512000` boundary as the declared conservative barness estimate policy, not quoted as an explicit official integer. Reliable future confirmation of the vendor's integer threshold is the condition for revising this snapshot policy; retain numeric boundary tests. Output/context configuration docs explicitly equate `512K` to `524288`, which does not settle the input-pricing threshold. No source among these fetched docs contains pricing threshold `512000`. This remains an unresolved source ambiguity. Small bounded live samples are below both interpretations; no high-context paid probe is needed or authorized by this research.

**Passive writes do not establish a Messages creation-token rate.** The passive-cache Pricing section assigns discounted rates to cache hits, standard input rates to new input, and standard output rates to output. Its comparison table describes passive caching as having no extra write charge. That is a general pricing description, not a definition of `cache_creation_input_tokens` classification for M3, and does not establish that separately classified creation tokens have unit rate zero or unit rate equal to ordinary input. [Passive cache](https://platform.minimax.io/docs/api-reference/text-prompt-caching)

The explicit-cache guide defines:

```text
total_input_tokens = input_tokens + cache_creation_input_tokens + cache_read_input_tokens
```

Here `input_tokens` excludes both cache reads and creation; creation and read are distinct categories. Its supported-model and write-price table contains M2.x only. The passive/explicit comparison also lists M3 only under passive support, and omits M3.1 from both lists. Consequently, positive M3/M3.1 `cache_creation_input_tokens` remains unpriced until reliable model-specific evidence defines both accounting and rate. Preserve known usage but fail cost estimation instead of fabricating a zero or input-equivalent CacheWrite price. Do not reuse the M2.x $0.375 write rate. [Explicit cache](https://platform.minimax.io/docs/api-reference/anthropic-api-compatible-cache), [Cache comparison](https://platform.minimax.io/docs/api-reference/text-prompt-caching)

**Current official documentation conflict:** passive-cache worked example still uses pre-discount M3 standard rates $0.60/$2.40/$0.12 for ≤512k, while pricing overview crosses those out and presents $0.30/$1.20/$0.06 as permanent 50% off. Treat the current pricing overview as the dated catalog rate source and preserve this discrepancy. The cache example remains useful for token-accounting shape, not current unit rates.

## 5. Reasoning defaults, depth and budgets

| Model | Chat Completions | Responses | Messages |
| --- | --- | --- | --- |
| M3 | On when omitted; `thinking.type=disabled` turns it off; adaptive enables | Off when omitted; non-`none` `reasoning.effort` enables | Off when omitted; `thinking.type=adaptive` enables |
| M3.1 | On when omitted; cannot disable; `reasoning_effort` controls depth | On when omitted; cannot disable; `reasoning.effort` controls depth | On when omitted; cannot disable; `output_config.effort` controls depth |

M3.1 accepts **low, medium, high, xhigh, max**, defaults to **max**, and rejects off/none with HTTP 400. Responses' broad effort enum also contains minimal/none, but its model-specific prose limits M3.1 to the five supported levels. Other models ignore Chat `reasoning_effort` / Messages `output_config.effort`; M3 Responses effort enables reasoning without changing depth. [Chat SDK](https://platform.minimax.io/docs/api-reference/text-openai-api), [Responses schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json), [Messages SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api)

Chat uses `max_completion_tokens` for new integrations; deprecated `max_tokens` is still listed. Messages uses `max_tokens`. For M3/M3.1 the Chat/Messages fields recommend 131072 and state maximum 524288. Responses documents `max_output_tokens` but does not give those numeric limits, so do not mechanically infer a matching Responses contract. Thinking counts against output budget; too-small budgets can produce only reasoning with `length`, `incomplete`, or `max_tokens` and no answer. [Chat schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-openai.json), [Messages schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json), [Responses schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json)

## 6. Tool and native history requirements

| API | Tool wire/selection | History and reasoning |
| --- | --- | --- |
| Chat | Function tool nested in `tools[].function`; result uses `tool_call_id`; schema does not declare `tool_choice` or `parallel_tool_calls` | Append complete returned assistant message, including tools and reasoning. SDK says split output uses `reasoning_content`; M3-specific tool guide says `reasoning_details[]` with type/id/format/index/text. Preserve actual shape; conflict remains. |
| Responses | Flat `type=function`, name, parameters; `function_call.call_id` pairs with `function_call_output.call_id`; choice only `none`/`auto` | Stateless full input history. Output reasoning uses `content:[{type:reasoning_text,text:...}]`, but documented reasoning input exposes `summary:[{type:summary_text,text:...}]`; actual native replay acceptance needs own wire evidence. |
| Messages | Tool uses `input_schema`; return `tool_result` in user message with `tool_use_id`; detailed schema supports only `auto`/`none` | Return the full assistant `content` list including thinking/text/tool_use unchanged. Request/response thinking schemas include a signature and explicitly require preserving it. Actual signature presence needs wire verification. |

[Chat schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-openai.json), [M3 Tool Use](https://platform.minimax.io/docs/guides/text-m3-function-call), [Chat SDK](https://platform.minimax.io/docs/api-reference/text-openai-api), [Responses schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json), [Messages schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json), [Messages SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api)

Chat `reasoning_split` controls output representation rather than enabling reasoning; M3.1 currently rejects `false`. For M3 the documented raw format preserves full `content` including `<think>` tags. Never flatten a native reasoning array into display text and call it unchanged replay. The native object must preserve its actual source/account/protocol/model boundary.

Responses store, parallel_tool_calls and truncation request fields are ignored; returned values are respectively false, true and disabled. Only `text.format.type=text` is documented; no JSON schema, previous_response_id, encrypted reasoning include, prompt cache retention, or hosted tools contract appears. Responses exposes prompt_cache_key routing only. Detailed Messages tool-choice schema is narrower than SDK's generic fully-supported wording. [Responses](https://platform.minimax.io/docs/api-reference/responses-create), [Messages schema](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json)

Messages explicit caching only declares ephemeral, no ttl field; general explicit-cache guide gives 5 minutes refreshed on hit, at most the latest 4 breakpoints and 20-block lookback. This does not prove M3/M3.1 explicit-cache support or a write rate.

SSE evidence gaps are unchanged: Responses JSON lacks StreamEvent schema; Chat/Messages detailed schemas omit some function-argument delta shapes; SDK snippets do not certify every final-usage or terminal-event edge. Integration should keep failure-on-premature-EOF and own protocol/model tool-round-trip acceptance.

## 7. Differences from committed China evidence

Three official international OpenAPI files were compared against committed China raw snapshots. Recursively excluding presentation/prose keys (`description`, `summary`, `title`, `example`, `examples`, `info`, `servers`, `externalDocs`, `x-codeSamples`, `x-code-samples`) yields **zero structural differences for all three**. Model enums, fields, types, defaults, numerical schema constraints, required lists and paths agree. This is not a claim of identical full prose, endpoint identity, live availability or billing eligibility. Source hashes and normalized hashes are in `china-schema-comparison.json`.

| Area | Committed China evidence | International evidence / consequence |
| --- | --- | --- |
| Provider/operator identity | `minimax-cn`, China platform account | Explicit `minimax`, international platform account and operator; no automatic fallback |
| Generation hosts | `api.minimax.cn` | `api.minimax.io` |
| Billing currency | CNY converted through fixed FX snapshot | Native USD rate snapshot; remove FX dependence for international catalog |
| M3 standard current prices | CNY 2.10/8.40/0.42 low tier; 4.20/16.80/0.84 high tier | USD 0.30/1.20/0.06; 0.60/2.40/0.12 directly; no implied conversion contract |
| Subscription details | China M Plan conditions / CNY credits | International M Plan's own entitlements and key-purpose evidence apply; existing key eligibility unconfirmed |
| Model reasoning/tool contracts | Existing defaults/replay conflicts | Same structural schema constraints and documentation conflicts remain |
| Promotion and cache docs | China price snapshot exists | Pin international promotional prices and stale passive-cache worked example separately |
| Live evidence | China auth failure / controlled fixtures | Cannot certify international account, model or protocol; fresh international evidence required |

Do not mechanically relabel or renew China live records as international. Preserve original region/source and record new international identity, endpoint, USD source snapshot and actual model/protocol outcomes.

## 8. Remaining questions and implementation-safe positions

1. **Existing key identity/purpose and M3/M3.1 entitlement:** not inspected. Host must explicitly select the international account/purpose. M3.1 remains candidate/NOT_RUN if no entitled Subscription Key; no guessed rate or fallback.
2. **Exact integer pricing threshold:** official spelling is 512k. The existing 512000 boundary remains the conservative barness estimate policy; only reliable future integer confirmation should revise it. Keep the numeric boundary tests and avoid high-context paid probing.
3. **Positive M3/M3.1 creation tokens:** no model-specific separate write rate/accounting mapping established. Preserve positive usage and report estimate unsupported.
4. **M3 reasoning format and Responses native history asymmetry:** retain own bounded live research/fixtures and failure inventory; document actual shape.
5. **SSE, tools, signature, usage:** same implementation may be reusable, but independent international protocol/model evidence is required to certify support.

## 9. Source snapshot manifest

All rows below were retrieved from official `platform.minimax.io`; full requested and final URLs plus headers are in `sources.json`. `.md` Terms/Privacy routes redirected to their official HTML protocol pages. Extracted text files are derived conveniences, not raw snapshots.

| File | Retrieved UTC | SHA-256 | URL |
| --- | --- | --- | --- |
| `llms.txt` | 2026-10-18T00:56:09.783740+00:00 | `5b98d1c4243506ce4f8767d3c0d629bc03db916a3fba81d03be320d3b63d7ba7` | [source](https://platform.minimax.io/docs/llms.txt) |
| `guides_models-intro.md` | 2026-10-18T00:56:09.400857+00:00 | `0bbe91acbb57a6d328577417a9740cfa9175e93b64f16d1d7dcd765204c648ff` | [source](https://platform.minimax.io/docs/guides/models-intro.md) |
| `api-reference_text-openai-api.md` | 2026-10-18T00:56:09.714304+00:00 | `ee59b46734b303cce80dc60fd7729abce43cb205ad5914356c16beac95934412` | [source](https://platform.minimax.io/docs/api-reference/text-openai-api.md) |
| `api-reference_text-anthropic-api.md` | 2026-10-18T00:56:09.573849+00:00 | `b4a02ef2e755bcd07c897918ff2cbac71340ed12cc01c172a09720fd725aa80f` | [source](https://platform.minimax.io/docs/api-reference/text-anthropic-api.md) |
| `api-reference_responses-create.md` | 2026-10-18T00:56:09.682775+00:00 | `b8ed98eb11ce6639e48801e0e8efd2e1dc6730ba4265e5c28698fbaef24c07d6` | [source](https://platform.minimax.io/docs/api-reference/responses-create.md) |
| `api-reference_responses-input-tokens.md` | 2026-10-18T00:56:09.752054+00:00 | `ea557c7a509ee191f56de6a7504c10d67274f647bcf90c586813c83bd6a08058` | [source](https://platform.minimax.io/docs/api-reference/responses-input-tokens.md) |
| `api-reference_text-chat-openai.md` | 2026-10-18T00:56:10.920348+00:00 | `a61da663f19d9e8c79026164186b7539a999aa93337b3381574ee4dbea25a2b5` | [source](https://platform.minimax.io/docs/api-reference/text-chat-openai.md) |
| `api-reference_text-chat-anthropic.md` | 2026-10-18T00:56:10.926089+00:00 | `158f3cfa0d5954b904e0ee3975f7b0fbf60c2074a7cc62760db47bc7f2faf005` | [source](https://platform.minimax.io/docs/api-reference/text-chat-anthropic.md) |
| `api-reference_text_api_openapi-chat-openai.json` | 2026-10-18T00:56:11.281460+00:00 | `8d26ce6ffaa62d95be4ccaa5926e08bbdb6cb0f7d172067373439207f93a65ce` | [source](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-openai.json) |
| `api-reference_text_api_openapi-chat-anthropic.json` | 2026-10-18T00:56:11.329978+00:00 | `867ac2ff2f6ceb9da629bdeb2c0172c6af95a25864adc3323e643229eddd45b0` | [source](https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json) |
| `api-reference_text_api_openapi-responses.json` | 2026-10-18T00:56:11.169901+00:00 | `63f2b483a9355562d58f63210251ab0b96da893a4be1d682326cf3cd08dd1687` | [source](https://platform.minimax.io/docs/api-reference/text/api/openapi-responses.json) |
| `pricing_overview.md` | 2026-10-18T00:56:11.293290+00:00 | `92fde6068cfd911157e944bd7c2c16887abea7b1ab10be7b89dfe35695350400` | [source](https://platform.minimax.io/docs/pricing/overview.md) |
| `m-plan_faq.md` | 2026-10-18T00:56:12.002555+00:00 | `c23fb3ba9a595cff7d0042dac0df5472e76eb4da4f6b9145715542cc6576a3de` | [source](https://platform.minimax.io/docs/m-plan/faq.md) |
| `m-plan_quickstart.md` | 2026-10-18T00:56:11.926244+00:00 | `6bedf6bbc72dd88f161b7ad3bdc1ed6c5ae7c6f59171c39558ef3ae0b531aeb0` | [source](https://platform.minimax.io/docs/m-plan/quickstart.md) |
| `m-plan_other-tools.md` | 2026-10-18T00:56:12.518691+00:00 | `32f2567d1019bfc368a5ad718d438a9079757ed33bda6f664eb3c7d6edf2a56f` | [source](https://platform.minimax.io/docs/m-plan/other-tools.md) |
| `guides_text-m3-function-call.md` | 2026-10-18T00:56:12.984331+00:00 | `40ab52a616a5321ff841dc6250b0b62a7cdc0384eb0218365f27beef648f25bb` | [source](https://platform.minimax.io/docs/guides/text-m3-function-call.md) |
| `api-reference_text-prompt-caching.md` | 2026-10-18T00:56:12.427405+00:00 | `02431e0ec5f11dae75593ed6a9097b4a06c3b5a8f8c29d09dd49c5818054a3e5` | [source](https://platform.minimax.io/docs/api-reference/text-prompt-caching.md) |
| `api-reference_anthropic-api-compatible-cache.md` | 2026-10-18T00:56:12.522361+00:00 | `2bc621902accdfbe026a919ec487fd2d0466947057cb6fcfa4fcc8a6c877ed6e` | [source](https://platform.minimax.io/docs/api-reference/anthropic-api-compatible-cache.md) |
| `guides_rate-limits.md` | 2026-10-18T00:56:12.975650+00:00 | `6191937e3e86ecbea823f9952ba58c51a3c0f637c8150263f2715f6fc2dcffe6` | [source](https://platform.minimax.io/docs/guides/rate-limits.md) |
| `api-reference_errorcode.md` | 2026-10-18T00:56:13.000327+00:00 | `812319dd99f041c99456ba1a84f93e8f06f0e5e51c255fed30ebea3792c51a6b` | [source](https://platform.minimax.io/docs/api-reference/errorcode.md) |
| `guides_quickstart-preparation.md` | 2026-10-18T00:57:15.834345+00:00 | `17f69d413b6d362e02a29e158d76c8a0a8baa65e846b461d9f306a5326e78290` | [source](https://platform.minimax.io/docs/guides/quickstart-preparation.md) |
| `guides_terms-of-service.md` | 2026-10-18T00:57:17.105417+00:00 | `e291f66b2a112c3f471d8f40ad3da2c0287a94890f89d1dbbeb698627a5d1679` | [source](https://platform.minimax.io/protocol/terms-of-service) |
| `guides_privacy-policy.md` | 2026-10-18T00:57:17.265546+00:00 | `ba579d48a4295d5aaf38744e061f8da580caba8e0cc785297fb70acff1a7fcd3` | [source](https://platform.minimax.io/protocol/privacy-policy) |
| `faq_about-account.md` | 2026-10-18T00:57:15.598842+00:00 | `4170f89e4cfc7fd2bbe1eb95d94511ae4b95edb7cf27b37d03db6c81f23ef3d6` | [source](https://platform.minimax.io/docs/faq/about-account.md) |
| `faq_about-apis.md` | 2026-10-18T00:57:15.757631+00:00 | `a9d25ff241c64388d955724d49aa39143e3282301deb441a65799a8da195c0a4` | [source](https://platform.minimax.io/docs/faq/about-apis.md) |

Derived artifacts:

- `china-schema-comparison.json` SHA-256 `fe0bbfba7e3703868301f370ddfb0a83d3f838091397dcb09043e37c72ea18e0`
- `facts.json` SHA-256 `d9d4a1a8baf345b7d158a8687d13fa1180eaaa187d7dd4d8a8cc5427352963f6`
- `guides_terms-of-service.md.extracted.txt` SHA-256 `3edccad1882a1ead4716dc6f4bc0e62390f76aaf24c79398b3ccae33e1db08fc`
- `guides_privacy-policy.md.extracted.txt` SHA-256 `dec6075dced83cf7eb462ac3a17264848ab3efdd98a4151383431b65410eb4af`

## 2026-10-18 own international research checkpoint

Actual source `38b7098`; [unchanged raw capsule and pins](evidence/ticket-10/international-research-38b7098.json). International authentication works: all attempted generation requests returned HTTP200 and actual modelMiniMax-M3. Chat aborts at the first mirrored reasoning chunk: simultaneous reasoning_content and reasoning_details carry equal text; the existing parser incorrectly treats the pair as conflicting, leaving usage unreported. Native continuation was NOT_RUN. Responses text and tool/thinking succeed with known usage, but persisted trusted native continuation fails during local request encoding before any HTTP attempt. Messages text, tool/thinking and native continuation all PASS. These are bounded research observations, not complete candidate/formal reports.

The three new processes consumed8 Logical Calls and7 Attempts; with the retained CN2/2, candidate stage totals10/120 calls and9/144 Attempts (10 reserved). Chat slot4calls/4attempts, Responses3/2, Messages3/3; formal0/0. Unknown Chat consumption is not a zero bill. Source changes require fresh exact-source preparation/review before the unexecuted one-shot complete candidates; no failed research intent/receipt is overwritten or reset.
