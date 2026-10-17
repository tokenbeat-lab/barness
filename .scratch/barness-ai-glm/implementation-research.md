# GLM implementation evidence supplement

Read with `.scratch/barness-ai-glm/spec.md` and `research.md`; this file records new implementation-critical checks, not completed service verification.

Retrieved 2026-10-17, approximately 12:20–12:32 UTC (Asia/Shanghai 20:20–20:32). No credentials were inspected and no model API request was made. Repo files were read only. Copies are `/tmp/barness-glm-*.txt` plus the FX JSON. Source facts and proposed implementation choices are distinguished below.

## Messages: what is actually established

Primary China-region source: [Claude API compatibility guide](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction.md). Its SHA-256 remains `becda63d7e3a6e71b4155b2a966794ea84a33dba67beebcaa6abafb823cd2cdc`, identical to the existing research snapshot.

Established by examples:

- Base URL is `https://open.bigmodel.cn/api/anthropic`; curl addresses `POST /api/anthropic/v1/messages`.
- curl sends `x-api-key` and `content-type: application/json`.
- Body contains `model`, `max_tokens`, `messages` and, for streaming curl, `stream:true`. A user message may have string content.
- Python/Java/curl examples use `glm-5.3`; TypeScript example uses `glm-5.2`. Thus both model IDs have independent Messages example evidence, but this is not evidence for each advanced field or successful real usage.
- The guide expressly acknowledges protocol differences; general compatibility language does not specify which advanced fields work.
- The published index (`https://docs.bigmodel.cn/llms.txt`) lists this one Claude compatibility page. The current China Chat OpenAPI document has no Messages/Anthropic path. No official China-region per-field Messages schema or response examples were located.

The guide supplies **no field-level evidence** for these: `thinking`, `output_config.effort`, `budget_tokens`, `display`, `block_binding`, signature generation/presence/requirements, redacted thinking, tool definition/choice options, cache markers/TTL, eager input streaming, betas, SSE thinking deltas, usage positions, request-ID response headers, unsigned-thinking replay. These remain unresolved, not officially unsupported.

### SDK facts, independently established locally

Locked SDK is `github.com/anthropics/anthropic-sdk-go@v1.75.0`. Its `internal/requestconfig/requestconfig.go:155–170` initializes Content-Type when a body exists, `Accept: application/json`, `X-Stainless-Retry-Count: 0`, platform metadata, and `anthropic-version: 2023-06-01`. `X-Stainless-Timeout` is removed if no request timeout is set, otherwise contains timeout seconds. These are SDK behavior, **not GLM field acceptance evidence**. The official GLM guide's Anthropic SDK examples make this a reasonable candidate request, but a local final-request capture and real smoke must establish actual wire and acceptance. Keep SDK retries disabled as required by spec.

### Minimal supported Messages request candidate

```json
{
  "model":"glm-5.3",
  "max_tokens":1024,
  "stream":true,
  "messages":[{"role":"user","content":"Hello"}]
}
```

The same minimal shape is documented for `glm-5.2` without streaming. The candidate is narrow **documentation-supported shape**, not a claim that basic text establishes spec completion. Omit default Claude cache/beta/eager/thinking fields. Do not emit disabled thinking merely because simple reasoning is omitted. Tool auto/default and thinking/signature replay still need a bounded, authorized service probe followed by public Client verification (R13/R14). Do not copy Chat's `thinking.type` or `reasoning_effort` mapping into Messages.

## Chat: GLM-5.2 and GLM-5.3 have separately explicit rules

Primary sources:

- [China Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json), `components.schemas.ChatCompletionTextRequest` and `ChatThinking`.
- [Deep thinking](https://docs.bigmodel.cn/cn/guide/capabilities/thinking.md).
- [Thinking modes](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode.md).
- [GLM-5.3 model](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3.md).

Chat OpenAPI explicitly provides:

| Model | Accepted documented API effort inputs | Documented service mapping |
| --- | --- | --- |
| GLM-5.3 | low, high, max | own names; default max; disabled thinking errors |
| GLM-5.2 | none, minimal, low, medium, high, xhigh, max | none/minimal abandon thinking; low/medium→high; high→high; xhigh/max→max |

The deep-thinking guide independently confirms GLM-5.2 `thinking.type:disabled` and GLM-5.3 mandatory thinking. Default enabled is explicit in ChatThinking; defaults in general capability prose are less precise about dynamic versus forced older-model thinking, so preserve the wire default rather than claiming all enabled models spend a fixed amount reasoning.

The same deep-thinking guide separately describes Coding Plan alias coercion for GLM-5.3 (`none/minimal/low→low`, `medium/high→high`, `xhigh/max→max`). **Do not apply those Coding Plan rules to the standard API**; standard API explicitly errors on other levels.

`thinking.clear_thinking` is nested, defaults true, and applies to previous turns. Current-turn reasoning and tool interleaving are separate: return complete `reasoning_content` with assistant tool calls and the matching results. Spec deliberately rejects false (preserved prior-turn thinking) for this initial integration.

Minimal standard Chat candidate:

```json
{"model":"glm-5.3","messages":[{"role":"user","content":"Hello"}],"stream":true,"max_tokens":1024}
```

Optional, documented additions: `thinking:{"type":"enabled"}`, `reasoning_effort:"low"|"high"|"max"` for 5.3; 5.2 switch/effort rules above. `tool_stream:true` is documented for streaming function parameters; `tool_choice` schema supports only `auto`. No `stream_options` property appears in the text request schema: omit it by default but do not describe it as officially rejected. No default store/developer/max_completion_tokens/cache-retention additions.

## Responses: preserve the documented conflict

Primary [Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json):

- `ReasoningConfig.effort` defaults max, accepts all seven strings, and documents generic aliases `none/minimal` abandon, `low/medium→high`, `xhigh→max`.
- This conflicts with model-page GLM-5.3 low/high/max-only, always-thinking description. Spec resolves the initial contract to low/high/max inputs. Preserve the source conflict; `low` is an accepted wire input in both sources but its *actual computation intensity* must not be equated with Chat low without service evidence.
- The Responses document does not independently enumerate 5.2 in the model string schema (it is a free string with 5.3 example); don't overclaim protocol+model verification from the Chat enum.
- `store` defaults false; send it explicitly. `previous_response_id` is a stateful feature, so omit it for the stateless contract.
- `max_output_tokens` schema explicitly has min 1, max **131072**, default **65536**, including reasoning plus answer.
- `reasoning` has effort only: no summary/encrypted fields. `tool_choice` enum none/auto; function tool schema has no strict property.
- `InputReasoning` requires type reasoning and **single object** content of type reasoning_text with text. It does not define id, summary or encrypted fields. Input function call requires name/call_id/arguments but no item id; result pairs by call_id.
- Output reasoning content is an object; output message content object or array. SSE reasoning_text delta/done are schema members. This confirms the existing shape-fix requirement.

Minimal Responses candidate:

```json
{"model":"glm-5.3","input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"Hello"}]}],"stream":true,"store":false,"max_output_tokens":1024}
```

Optional documented conservative reasoning addition: `reasoning:{"effort":"low"|"high"|"max"}`. No generated summary, include, service tier, background, retention, affinity, strict functions. Prompt cache key is a routing key, not a promised lifetime or cache hit.

## Capacities: precise output, conservative context

[5.2 model](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.2.md) and [5.3 model](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3.md) independently state text input/output, 1M context and 128K max output.

New precision: Chat `ChatCompletionTextRequest.max_tokens.maximum` and Responses `CreateResponseRequest.max_output_tokens.maximum` each explicitly say **131072**. This is a real schema integer for those protocol schemas, not a guessed expansion of 128K. Chat min is also 1. No equally precise official Messages capacity/schema integer was found.

No official exact decimal/binary expansion of the 1M context was found. Proposed explicitly labeled barness conservative context bound: **1,000,000**, no higher than either decimal 1M or binary 1M interpretation. A Messages conservative output bound **128000** stays within either usual 128K interpretation; using 131072 for Messages would be a model-level inference shared from Chat/Responses rather than Messages-specific schema fact, so label accordingly or remain conservative. These are candidate metadata limits, not load-tested service capacity.

## Pricing and dated fixed currency conversion

Primary current [China API pricing](https://docs.bigmodel.cn/cn/guide/start/pricing.md), snapshot 2026-10-17:

| Model | RMB input / million | RMB output / million | RMB cache hit / million | Cache storage |
| --- | ---: | ---: | ---: | --- |
| GLM-5.3 | 8 | 28 | 2 | limited-time free |
| GLM-5.2 | 8 | 28 | 2 | limited-time free |

Unlike 5.1/5/Turbo rows, these two have **one 1M row each and no context-length price tier**. Do not copy smaller models' tiers. Cache storage is temporarily free, not a permanent zero-price promise. No separate cache creation token price is shown; do not infer a permanent free cache-write service. Actual billing remains provider usage/account billing; local rates are estimates. Coding Plan cannot inherit these as verified quota deduction.

Dated primary conversion source obtained from the **China Foreign Exchange Trade System / ChinaMoney**, whose page identifies PBOC-authorized central parity:

- [Dated announcement page](https://www.chinamoney.com.cn/chinese/ccprnoticecontent/index.html?searchDate=2026-10-16).
- [Dated public JSON used by that page](https://www.chinamoney.com.cn/r/cms/www/chinamoney/data/fx/ccpr-notice2026-10-16.json).
- JSON `data.vo.showDateForCn` is 2026年10月16日; `data.vo.price` starts with **1 USD = CNY 6.7220**. This is a dated central parity reference, not a retail conversion quote.
- Page source loads `/r/cms/www/chinamoney/data/fx/ccpr-notice` plus query date plus `.json`; downloaded page and `cm-basic.js` establish the JSON provenance. JSON SHA-256 `bb0b4f2575c08eeb09daba9af3e59c1c9d2ddda7c1df09dfbfbac9994c5ae36c`.

For an explicit fixed estimation snapshot: `USD amount = CNY amount / 6.7220`. Decimal calculations:

| Rate | USD / million |
| --- | ---: |
| Input 8 RMB | 1.19012198750371913121094912229 |
| Output 28 RMB | 4.16542695626301695923832192800 |
| Cache hit 2 RMB | 0.297530496875929782802737280571 |

A practical finite float snapshot may round to 1.1901219875 / 4.1654269563 / 0.2975304969 with formula, reference date, provider-price retrieval date and source URL retained. No runtime FX lookup or implicit currency reinterpretation.

Cross-check only: [US Federal Reserve H.10 current release](https://www.federalreserve.gov/releases/h10/current/) dated October 13 reports Oct 9 CNY 6.6920 per USD; this is a different reference/date, not a contradiction. Prefer the explicitly dated ChinaMoney source above rather than mixing sources or picking a favorable rate.

## Frozen pi 1.1.0 identity audit (for ticket 10)

Read-only inspection base: `ai/internal/testkit/pioracle/node/node_modules/@earendil-works/pi-ai/dist`. Existing `PROVENANCE.md` and `provenance.json` fix version 1.1.0, commit `abe508e1b89912adde45528136c3221eb69acdd7`, entire model-data hash `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`. npm gitHead is empty; do not fabricate it.

Provider identities actually present:

| Pi identity | Provider module/base | Model data APIs | Target examples |
| --- | --- | --- | --- |
| zai | providers/zai.js; `https://api.z.ai/api/coding/paas/v4` | openai-completions only | 5.2 and 5.3 |
| zai-coding-cn | providers/zai-coding-cn.js; `https://open.bigmodel.cn/api/coding/paas/v4` | openai-completions only | 5.3; **5.2 absent** |

Provider data has no identity `zhipu` and no China **standard** base URL `https://open.bigmodel.cn/api/paas/v4`, `https://open.bigmodel.cn/api/v1`, or `https://open.bigmodel.cn/api/anthropic`. Scanning all provider data found open.bigmodel.cn only in zai-coding-cn Chat entries (glm-4.6v, glm-5.3, glm-5.3-flash, glm-5.3-highspeed), all coding endpoint. This is the important identity evidence: same GLM model spelling in international/Coding Plan catalogs does not establish the spec's China standard route.

Therefore **all three spec combinations are barness extensions relative to this frozen native routing**. They can reuse public Client fixtures and explicit ledger/extension coverage, but cannot be counted as native same-route Pi differential PASS. Explicit model+endpoint changes to make Pi target China standard would test a custom route, not prove it shipped that native route.

Interesting Pi implementation similarities are evidence of code behavior only:

- api/openai-completions.js `getCompat` detects `zai`, `zai-coding-cn`, api.z.ai or open.bigmodel.cn as Zai-style thinking.
- buildParams sends `thinking:{type:"enabled",clear_thinking:false}` when reasoningEffort is truthy, otherwise disabled. This default is incompatible with standard GLM-5.3 mandatory thinking; do not copy it into barness merely for superficial diff parity.
- With declared tools and `compat.zaiToolStream`, Pi sends tool_stream true. It emits stream_options by default unless compatibility says false.
- Pi Anthropic adapter comment mentions grouping consecutive tool results for a z.ai endpoint, but no native China standard Messages provider/model route exists. Generic adapter knowledge is not a native routing claim or official GLM field contract.

Useful exact file hashes (already covered by released package provenance):

| Dist file | SHA-256 |
| --- | --- |
| providers/data/zai.json | 58e6642fdfb736f0b9a4a277c760c536a3cea66a030abbfb2615f7b2851d5c6f |
| providers/data/zai-coding-cn.json | ee5ee705fb6e413a6088db5e3616c27fce802a8ece6c355f1db35490d787f610 |
| providers/zai.js | d897597f9bd9749ccc674c34e5c3bb738d9b90911b04c7f7fc7bc66150d6ff40 |
| providers/zai-coding-cn.js | 3820dea6091183bafa6b0c37e2cca684fba6d50674f190db259b2765c70efa85 |
| api/openai-completions.js | 5a79cc5aa41dee180692695e65771720d1ea556e9a978649dc2040278be51753 |

## Download manifest

| Primary source file in /tmp | SHA-256 |
| --- | --- |
| barness-glm-claude.txt | becda63d7e3a6e71b4155b2a966794ea84a33dba67beebcaa6abafb823cd2cdc |
| barness-glm-glm52.txt | ad435d6334a6c221a7004ada5cb6a0c936968c445b1648f7158623ed1f2967d1 |
| barness-glm-glm53.txt | 2ae550297e839b51498db0f487d9ba5cb0c44f996236546d0eb2e07ce43ef8d8 |
| barness-glm-pricing.txt | e47e492fffffdd7d0cdb3fb6b61f0ed33b2ab829dd3973c9d55e8a1f1c537e4a |
| barness-glm-index.txt | 3a07b47f4fe3c7c8dcd536fe24a5bbadaf04cd1700e3ecec3e29eaedcd9b5080 |
| barness-glm-responses.txt | 0914af477ff4c1fae0372f9c44ace23033d0bed0cc486796eaff2c2024e5f18c |
| barness-glm-chat.txt | e1867cd1a7e31b15cc9088b03d8c1361e275153a1334b9e9e3bea95f4c623dd9 |
| barness-glm-thinking.txt | 280bf00451b2ad210439f95d2c6bee3ee7c9145db6da85ba23348aa0a98f12a2 |
| barness-glm-thinking-mode.txt | 0c3ce28a685cc23085b3f2d3cba39f000eae1403a720dd0fbcbc8a47f297524a |

## Remaining indispensable evidence

1. Messages true response/block/usage/signature/tool shapes and same-model continuation, under standard China key; no unsupported claims from absent docs.
2. Per-model/per-protocol real smoke, especially 5.2 Responses and 5.3 low effort conflict.
3. Final SDK headers/path capture and GLM acceptance, without accepting inherited beta/caching assumptions.
4. Provider request-ID response headers and per-protocol error shapes under actual standard account.
5. Any future price/free-period changes require a new snapshot; this fixed snapshot is reproducible but does not verify an invoice or account entitlement.

## Authorized early public Client Messages exploration (later update)

Actual two-request synthetic evidence is now available at `/tmp/barness-glm-messages-probe-sdk-query/`, especially `finding.md`, `index.json`, each model's `wire.json`, `response.sse`, `frames.json`, `events.json`, `result.json`, `usage.json`, and `secret-audit.json`. First local guard failure is retained separately at `/tmp/barness-glm-messages-probe/`; there were zero external requests in that first run because locked SDK beta service adds `?beta=true`. Parent then explicitly authorized the second run with only that query accepted.

The two model basic requests both returned200 and legal Messages termination. 5.3 produced signed thinking (one24-character signature fragment) preserved by the public Client; 5.2 produced text only in its simple arithmetic sample. Usage begins as input/output0 in message_start, actual counts arrive in message_delta (5.3 input26/output29;5.2 input20/output1;cache_read0;cache_creation absent), and public partial reporting matches. Header X-Request-Id is present in both samples; native Anthropic Request-Id lookup misses it. No advanced controls or tools/replay have yet been tested. This update resolves basic shape and available default5.3 thinking evidence only.

Cumulative early budget:4 Logical Calls /4 conservative local attempt reservations /2 external HTTP requests; both secret audits passed. Actual live credential never appears in notes/artifacts.

## Cache creation versus storage pricing (official source follow-up)

Source retrieved2026-10-17: https://docs.bigmodel.cn/cn/guide/capabilities/cache.md (`/tmp/barness-glm-cache.txt`, SHA256 `f6a428c0c63ae10e9423306205c012b40d2b8c9a26eb277f16f79a52cfc7cb84`). Its billing section is expressly for standard API, excluding resource packs and Coding Plan. It separates new-content tokens (standard price), cache hits (discounted price), and output tokens (standard price). The source's generic example uses a 50% hit discount, but that is illustrative; use model-specific current2/8 cache/input rates for5.2/5.3, not the generic percentage.

The pricing page separately defines cache storage in RMB per million tokens per hour and calls it limited-time free. That time-based storage charge is different from new-content input computation. Therefore a zero CacheWrite token rate cannot be justified by temporary free storage.

Proposed normalized mapping, explicitly an implementation inference: if a Messages response separately reports cache_creation_input_tokens, those are newly computed/new-content input tokens. Since normalized Usage.Input excludes CacheWrite, price CacheWrite at the same sourced standard input rate (8RMB/M→1.190121987503719 USD/M) to count them once, without a separate Anthropic premium. This is derived from the source's new-content rule, not a documented GLM Messages cache_creation field contract or observed bill. If the meaning of that future field is not established, preserve partial reporting/unknown status rather than treating it as free. No rate for a future storage-free-period ending or one-hour Anthropic premium is established.

All early live Messages samples so far report cache_read_input_tokens0 but omit cache_creation_input_tokens; they do not establish any cache-creation counter semantics or charge. No additional service calls were made for pricing research.

## Authorized tools/thinking/native replay evidence (later update)

Six fixed public Client calls are recorded in `/tmp/barness-glm-messages-probe-tools-thinking/` (`finding.md`, `index.json`, SHA256SUMS and per-scenario raw wire/public snapshots). Both models accepted function schema+explicit auto, emitted tool_use/input_json_delta, passed public ValidateToolCall, and accepted original public Message+ToolResultText continuation. GLM5.3 first tool turn produced signed thinking; its one original thinking/signature block was preserved unchanged in public conversion and service accepted same-turn replay. GLM5.2 tool sample had no thinking.

One exact GLM5.2 enabled/budget1024 candidate succeeded with signed thinking; original public signed history plus a new user message was accepted with the same settings. This establishes exact wire acceptance and legal replay, not arbitrary budgets, enforced computation cap, effort mappings or guaranteed semantic preserved thinking. GLM5.3 tool continuation emitted text only, so do not require new observable thinking on every Messages response. All6 calls were HTTP200 and untruncated, with late usage fields matching public counts and partial reporting. No advanced/alternate/retry requests were made. Both secret audits passed.

Cumulative all early probes:10 logical calls /10 local attempt reservations /8 actual external HTTP requests. All three bundles remain separate and secret-free; no production edits.

## Final authorized Messages default/unsigned evidence

Bundle `/tmp/barness-glm-messages-probe-default-unsigned/` records3 more fixed public calls, allHTTP200/untruncated and audited. GLM5.2 equal challenging prompt with only thinking omitted emitted text only (no observable thinking); compared to the prior accepted enabled/budget1024 request, exact outgoing JSON equality minus thinking is proven. Both models accepted trusted restored thinking history with ONLY signature field omitted after public conversion, full JSON compared to pinned prior successful signed continuations. This establishes absent-signature replay acceptance for these exact cases, not empty-string signature or provider-emitted unsigned output. Trusted Host restoration/identity checks and public native-state path remained active.

Final early Messages cumulative13 Logical Calls/13 conservative local attempt reservations/11 actual external requests. No further Messages exploration is authorized or pending. All4 bundles remain separately hashed/audited outside repo.
