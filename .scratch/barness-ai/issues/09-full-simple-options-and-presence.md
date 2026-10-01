# 09: full/simple 选项、reasoning 映射与字段 presence（E04 选项部分）

**What to build:** 高级开发者通过按 API 区分的完整选项使用 Responses 全部能力；普通开发者用 simple 入口的统一 reasoning 等级与预算；未设置、显式 null、零值在请求中得到与冻结 pi 一致的编码（spec I7 前五条、User Stories 21–23）。

**Blocked by:** 01, 02

**Status:** ready-for-agent

- [x] full 入口校验选项与绑定协议匹配，不建立按厂商名索引的选项命名空间
- [x] simple 支持 minimal/low/medium/high/xhigh/max、thinkingBudgets、toolChoice 与公共参数；`off` 不加入 simple 输入枚举
- [x] 移植支持等级、clamp、thinkingLevelMap 禁用/重映射与未设置分支；默认预算 minimal=1024、low=2048、medium=8192、high=16384 及自定义规则；maxTokens、回答空间、输入估算、4096 安全余量与无 contextWindow 例外；估算不裁剪历史
- [x] presence-aware 字段表达未设置 / null / 零值三态
- [x] samplingParams 与调用参数逐键合并、调用值优先；OpenAI-compatible 路径保留最后覆盖已命名字段的语义（其他 API 的忽略行为在对应协议票验证）
- [x] cacheRetention、metadata、toolChoice 完整保留并按 fixture 验证；TenantID 不自动进入厂商 metadata；缓存/亲和标识按租户与账户作用域派生
- [x] 发送前复核模型与资源授权，采样参数不能越过认证/目标边界
- [x] 全部 reasoning 等级、预算边界与 presence 用例接入 pi 差分，无待处理差异

## Comments

**2026-10-01 — implemented** (E2E `ai/e2e/options_test.go` with fixture `testdata/responses/options.json`, 58 scenarios; differential `PIDIFF-P01-E04-options-*`; isolated `ai/reasoning_budget_test.go`).

- Public contract:
  - `Nullable[T]` (`ai.Value`, `ai.Null`, zero value = unset) carries the fields where unset / null / value encode differently: temperature, maxTokens, serviceTier, toolChoice, thinking budgets, and model `thinkingLevelMap` entries. Its JSON keeps all three states.
  - `ResponsesOptions` and `SimpleOptions` use pi's option names in JSON, so one fixture `options` object drives both barness and frozen pi.
  - `ResponsesOptions`: temperature, maxTokens, samplingParams, cacheRetention, sessionId, reasoningEffort, reasoningSummary, serviceTier, and toolChoice (mode / function / `allowed_tools`). It has no metadata field, because pi's Responses adapter ignores metadata.
  - `SimpleOptions` adds reasoning, thinkingBudgets and metadata. Responses ignores metadata (scenario `simple-metadata-ignored`).
  - The `off` reasoning level is rejected as input.
  - `Model` gains `ThinkingLevelMap` and `SamplingParams`. `ModelCompat` gains `SupportsLongCacheRetention` (unset means true) and `SupportsExplicitPromptCacheMode`.
- Mapping (pi parity, every case checked against pi):
  - Supported levels, clamp up then down, and the level map's null, remap and unset branches (`map[x] ?? x` on the full entry).
  - The off mapping when no effort is given. A summary given alone gives effort `medium`.
  - Simple maxTokens: unset or null takes the model maximum; 0 is kept. The value is clamped to contextWindow − estimate − 4096, with a floor of 1. A model without a contextWindow is not clamped. Responses enforces a floor of 16.
  - The estimate uses UTF-16 length, counts each image as 4800 characters, and includes tool JSON and system sections and removals. It uses the last successful usage and skips failed turns. It never trims history.
  - samplingParams merge per key, with the model's first (simple entry only, as in pi), and are applied last over named fields.
  - cacheRetention covers the 24h retention, explicit mode and long-retention compat paths.
- Deliberate differences from pi:
  - **Cache key** (spec §9): `prompt_cache_key` and the `session_id` / `x-client-request-id` headers carry a SHA-256 key over tenant, account and session instead of the raw sessionId (`cache_key.go`; ledger extension; `TestCacheKeyScope`). An empty sessionId counts as unset. pi would send `prompt_cache_key: ""` in that case.
  - **Session affinity header format:** pi switches to `x-session-id` when the base URL contains `openrouter.ai`. Only the OpenAI format is implemented, because OpenRouter is not a phase-1 provider.
  - **Reserved samplingParams keys:** samplingParams may not set `model`, `input`, `stream`, `tools`, `previous_response_id`, `conversation`, `prompt`, `prompt_cache_key`, `store` or `background`. This covers both call and model samplingParams, after the merge. Such keys are rejected as `invalid_request` in the capability phase, before credentials are read. This is how "采样参数不能越过认证/目标边界" is enforced. samplingParams only reach the body, so headers, endpoint and key cannot change through them. A re-check of the final payload after build belongs with ticket 10, the first point where the payload can change after it is built.
  - **Stricter validation than pi:** negative maxTokens, non-finite temperature, null or negative thinking budgets, unknown enums and malformed tool choices are rejected rather than forwarded (`TestOptionRejects`). They never reach the wire, so they do not appear in the differential.
  - **Estimate timestamps:** only assistant messages carry timestamps, and every other message counts as 0, as in the differential. If user or tool messages ever gain timestamps, `estimateContextTokens` must follow pi's every-message rule.
- **Thinking budgets:** the defaults (minimal 1024 / low 2048 / medium 8192 / high 16384), custom budgets and the 1024-token answer room are ported in `reasoning.go`. They have no Responses wire effect and no production caller yet. Tickets 16 (Anthropic) and 19 (Gemini) are their consumers; delete them if neither ends up using them. They are verified by an isolated test listing failure modes B1–B4. With `BARNESS_AI_PIDIFF=1` it checks every expectation against pi's own `adjustMaxTokensForThinking` through a new runner mode. That is a direct comparison, not a ledger-classified request diff. The functions were written before that test, contrary to the isolation-test rule; the expectations were then derived independently from pi's source and confirmed against pi.
- **Catalog:** version 2026-10-01.4 adds gpt-5, gpt-5-mini, gpt-5-nano, gpt-5-pro, gpt-5.1, gpt-5.2, o3, o3-mini and o4-mini, with pi's level maps. Models whose compat needs additional_tools or tool search are still not listed.
- **Differential:** the ledger's issue-09 pending entries are resolved. The simple-entry `max_output_tokens` entry is fixed. Content-Length is an extension limited to the six cache-key cases. The only pending findings left in every case are ticket 15's usage cost and reasoning items. The runner also gained `modelPatch` for custom model fields.
- **Existing fixtures** record pi's simple-entry `simpleMaxOutputTokens` (32768, or 16 for the gpt-4 image history), and their checks add it on simple entries.

