# 21: DeepSeek × Responses（P05）

**What to build:** 租户可通过独立的 DeepSeek Responses binding 完成文本、工具往返、推理与完整历史回放；复用 Responses adapter，但 Provider 身份、能力配置、endpoint 与 key 和 OpenAI 完全独立（spec I4 末两段、P05）。

**Blocked by:** 18

**Status:** resolved

- [x] 独立能力配置（endpoint `https://api.deepseek.com`）；Result 中 ProviderID 为 DeepSeek，与 OpenAI Responses 共享 adapter 时不串配置
- [x] 处理 response.reasoning_text.delta / done、incomplete、failed
- [x] 请求不自动注入 previous_response_id、conversation、store；历史通过完整输入回放；不支持字段的行为单列 fixture
- [x] 工具往返与推理历史回放
- [x] 本组合属于冻结 pi 路由之外的扩展：使用研究记录的官方协议 fixture 证明，登记为扩展，不计为 pi 差分通过
- [x] 复用 E01–E09、E11 适用用例全部通过

## Comments

**2026-10-02 — implemented** (capabilities `ai/responses_capabilities.go`; E2E `ai/e2e/deepseek_{responses,routing,limits}_test.go` with fixtures `testdata/deepseek-responses/{text,failures,history,unsupported,usage,retry}.json` and a `deepseek-responses` protocol in `testdata/isolation/interleave.json`; decisions in ADR-0014, status proposed).

- **API:** `ProviderDeepSeek` (`deepseek`). Built-in catalog `2026-10-02.5` adds `deepseek-flash` on `openai-responses` — pi's `deepseek.json` data with the API swapped, compat reduced to `supportsStrictMode` (ADR-0014 决策三). DeepSeek's guide lists only deepseek-flash on Responses, so deepseek-v4-pro is Chat only.
- **Capabilities (ADR-0014 决策一、二):** the Responses adapter derives store, include, the prompt cache fields with the affinity headers, and service tier pricing only when the binding's provider serves them; DeepSeek serves none, so its requests carry none of them and a session id derives nothing. An explicit `ServiceTier` on DeepSeek is refused (`invalid_request`, before any request). `previous_response_id`/`conversation` were never sent; samplingParams and payload callbacks still cannot add them, nor `store: true` or a non-derived cache key. History is always complete input: same-model reasoning items go back verbatim, as DeepSeek merges them into the assistant turn. OpenAI's requests are unchanged.
- **Events:** `response.reasoning_text.delta` was already handled; `reasoning_text.done` is framing (the item's `reasoning_text` content closes the block), as pi. `incomplete` (max_output_tokens → length, other reasons → `Response incomplete: <reason>`), `failed` (with/without details), a missing terminal, HTTP 400/401/429, cancel and retry each have a fixture.
- **Extension registration (ADR-0014 决策四):** the differential ledger gained `routes` (Provider × API without a frozen pi route, classification `extension` only); DeepSeek × Responses is registered there. Every P05 scenario carries `pidiffSkip`; no `PIDIFF-P05-*` case exists; `P05-E10-registered-as-extension` asserts both.
- **Coverage (offline):** P05-E01 text (full/simple × Stream/Result/Complete), reasoning (full high, simple medium clamped to high), reasoning + tool call; P05-E02 ten terminals; P05-E03 full-history replay, same-model tool round trip with reasoning replay (also as a live round trip through `ValidateToolCall`), OpenAI gpt-5 history (cross-model downgrade), untrusted state, image; P05-E04 unsupported fields (session/retention on both entries, reasoning summary, a response's service tier left unpriced) and eight refusals with zero requests; P05-E11 usage (cache reads and reasoning priced at DeepSeek rates, details missing → partial, null → unreported); P05-E05 retry; P05-E07 each binding resolves only its own provider's catalog; P05-E08 request body, frame and output bounds; P05-E10 OpenAI and DeepSeek concurrently on one Client through the shared adapter (own path, key, account, attribution per event, OpenAI keeps store/cache, DeepSeek has none) plus the envelope and local/host example suites; E06 isolation for every scenario and entry (DeepSeek derives no cache key for either tenant). `go test ./...`, `go test -race ./ai/...`, `go vet ./...` and `BARNESS_AI_PIDIFF=1` (0 pending) pass.
- **Not covered here:** the real-API smoke (issue 23) — in particular effort `high`/`max`, the reasoning and function call item shapes, DeepSeek's error bodies and request id header are inferred from OpenAI's schema and need live confirmation. The E09 observer/admission records are asserted through the E06 isolation suite only; the tool JSON and error body bounds are the shared adapter's (P01).
- **Open for the maintainer:** ADR-0014 "待维护者确认" — capabilities by provider, the ServiceTier refusal, the catalog entry, the ledger `routes`.
