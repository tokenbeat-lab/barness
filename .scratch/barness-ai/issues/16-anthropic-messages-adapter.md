# 16: Anthropic × Messages（P02）

**What to build:** 租户可通过 Anthropic Messages binding 完成文本、工具往返、thinking 签名回放、图片输入与缓存计费，全部经同一 Client 入口，并通过已在 Responses 上建立的所有适用横向场景（spec I4、I5、P02）。

**Blocked by:** 05, 06, 08, 09, 10, 11, 15

**Status:** resolved

- [x] 使用 Anthropic Go SDK（研究锁定 v1.75.0 为起点，纳入时重新核实），关闭默认重试，transport 约束与 01 一致
- [x] 以 message_stop 判定成功终态；缺终态 EOF、SSE error 事件为错误终态
- [x] signed empty thinking、redacted thinking、交错块、input_json_delta / signature_delta 正确归一与回放
- [x] 工具往返、图片与不支持图片占位分别覆盖
- [x] Anthropic adaptive effort / token budget 映射；非 OpenAI-compatible 路径忽略 samplingParams
- [x] 替换 payload 后仍强制 stream=true；onResponse 时点与 Responses 一致
- [x] 缓存读写、1h 写入与计价
- [x] 复用 E01–E05、E11 适用用例（full/simple × Stream/Complete、不读事件的 Result、成功与失败终态）全部通过
- [x] pi 差分无待处理差异

## Comments

**2026-10-02 — implemented** (adapter `ai/anthropic.go`, `anthropic_options.go`, `anthropic_request.go`, `anthropic_sse.go`, `anthropic_stream.go`; E2E `ai/e2e/anthropic_*_test.go` with fixtures `testdata/anthropic/{text,failures,history,options,usage,retry}.json`; decisions in ADR-0011, status proposed).

- **API:** `ProviderAnthropic`, `APIAnthropicMessages`, `AnthropicOptions` (+ `AnthropicEffort`, `AnthropicToolChoice`), `AssistantMessage.ResponseModel`, `ModelCompat.ForceAdaptiveThinking`/`SupportsTemperature`. Built-in catalog `2026-10-02.2` adds ten Anthropic models (ADR-0011 决策五 lists the five left out).
- **SDK:** `anthropic-sdk-go v1.75.0` (the research pin; re-verified for this use). It sends the request with `WithMaxRetries(0)` through the Client's shared transport, never via `anthropic.NewClient` (which reads `ANTHROPIC_*` and credential files). The event stream is decoded by barness as pi decodes it (ADR-0011 决策一).
- **Shared code:** the initial-request failure classification moved to `ai/http_failures.go` and the SDK middleware to `ai/sdk_middleware.go`; Responses uses both. `sanitizeToolCallID` is shared with Responses' `normalizeIDPart`. The adapter interface's `simpleOptions` now receives the normalized history (Anthropic clamps the output budget again after adding the thinking budget).
- **Bug fixed on the way:** `ModelCost.estimate` converted only the operands of each product, so Go could fuse a product with the later sum (it did on arm64) and the total differed from pi in the last bit. Each product is now rounded on its own.
- **Coverage (offline, 102 cases):** full/simple × Stream/Complete and Result without Next on text and every failure terminal; all six SSE framings; interleaved thinking/text/tool blocks with signature_delta and input_json_delta; JSON repair and trailing events; signed empty thinking and redacted replay; untrusted and cross-provider downgrade; a live tool round trip; images and placeholders; adaptive effort and budget mapping; samplingParams ignored; presence (null temperature, thinkingEnabled false); cache retention none/short/long (1h ttl); payload replacement still streaming; betas; callback authorization; onResponse before start; header guard; retry and no replay after start; cache read/write/1h pricing, thinking tokens, partial reporting; no environment fallback.
- **Differential:** 45 cases `PIDIFF-P02-*` reuse the same fixtures; the full run (`BARNESS_AI_PIDIFF=1 go test ./ai/...`, both protocols) has **0 pending**. New `anthropic-messages` ledger entries mirror the approved runtime-header, live-partial and rawArguments decisions, plus case-scoped ones for the redacted key, V8/undici/DOM runtime texts and the untrusted-envelope downgrade. `TestCostEstimate` drops its two one-hour-write cases now covered end to end (ADR-0010 决策五), keeps the tier cases, and checks every built-in Anthropic price against pi's data.
- **Code review (2026-10-02):** fixed callback `betas` to go out as pi's `toString()` (they were deduplicated); authorization refusals now also run through the simple entry; shared helpers moved out of Anthropic files and payload authorization split into `ai/anthropic_payload.go`.
- **Open for the maintainer:** ADR-0011 "待维护者确认": SSE decoded by the adapter (narrows spec I5's wording) with pi's message_stop rule and stream always true; trusted callbacks may enable any beta; Anthropic usage completeness and lenient missing usage; excluded models.
