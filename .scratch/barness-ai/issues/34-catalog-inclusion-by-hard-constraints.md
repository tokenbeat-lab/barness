# 34: 按硬约束标准列入被可选特性挡住的模型（claude-opus-4-8、claude-fable-5 与十个 OpenAI 模型）

**What to build:** 按 ADR-0018 的列入标准，把此前仅因可选协议特性未实现而排除的模型列入内置目录：Anthropic 的 claude-opus-4-8、claude-fable-5，OpenAI 中需要 additional_tools/tool search 的十个模型。这些模型按 pi 中对应特性关闭时的行为发送请求；与 pi 的差异登记为扩展；用真实冒烟确认厂商接受这些请求。

**Blocked by:** 16, 20

**Status:** ready-for-human

**Context:** 基线为 pi-ai 0.87.1 的 `anthropic.json`、`openai.json`（哈希见 `ai/catalog.go` 的 `BuiltinCatalog` 注释）。不需要改 adapter：`supportsMidConvoToolChanges` 关闭时 pi 发送当前工具列表（`buildParams`）；`allowedFallbackModels` 不发送时请求有效；`supportsAdditionalTools`/`supportsToolSearch` 关闭时 pi 同样发送当前工具列表（`resolveTranscriptTools`）。barness 现有路径与这些回退一致。只携带 `ModelCompat` 已有的字段，其余开关（`supportsMidConvoToolChanges`、`allowedFallbackModels`、`supportsAdditionalTools`、`supportsToolSearch`、`supportsOpenAIGrammarTools`、`supportsStrictTools`）不携带，由工单 26、28 或以后的工单引入。托管强度模型（claude-fable-5-1、claude-opus-5、claude-opus-5-5）不在本工单，见工单 27。

- [x] Anthropic × Messages 列入 claude-opus-4-8（adaptive、`SupportsTemperature=false`、level map `{xhigh, max}`、`SupportsMidConvoSystemMessages`）与 claude-fable-5（adaptive、level map `{off: null, xhigh, max}`、`SupportsMidConvoSystemMessages`）；名称、上下文、输出上限、价格逐字段与 pi 数据核对
- [x] OpenAI × Responses 列入 gpt-5.4、gpt-5.4-mini、gpt-5.4-pro、gpt-5.5、gpt-5.6-luna、gpt-5.6-sol、gpt-5.6-terra、gpt-6-astra、gpt-6-luna、gpt-6-sol：含 `SupportsMidConvoSystemMessages`、`SupportsExplicitPromptCacheMode`（gpt-5.6/gpt-6 系列）、level map 与阶梯价格，逐字段核对
- [x] OpenAI × Chat Completions 列入上一项中除 gpt-5.4-pro 外的模型（`builtinOpenAIChatModels` 的 pro 排除规则扩展为按 pi 数据判定，不再硬编码两个 ID）
- [x] 目录版本升级，`ai/release/catalog-snapshot.json` 与 pin 同步
- [x] 删除 `BuiltinCatalog` 注释与 `ModelCompat.SupportsMidConvoSystemMessages` 注释中"compat 需要未实现行为的模型不列入"的表述，改述 ADR-0018 的标准；`docs/barness-ai/differences.md` §6"未纳入目录的模型"一行改为"未实现的可选特性"（中途工具变更、服务端备用模型、additional_tools/tool search），指向工单 26、28 与 ADR-0018，托管强度模型指向工单 27
- [x] 离线 E2E 与 pi 差分：每个新增 Anthropic 模型覆盖无工具、带工具、中途增删工具、fable-5 的普通调用；OpenAI 每族至少一个模型覆盖带工具与中途 system/工具变更。pi 多出的占位工具与 tool-changes beta、`fallbacks` 与 server-side-fallback beta、`additional_tools`/`tool_search_*` 条目按 ADR-0018 决策四登记为 `extension`，`cases` 限于这些用例；无待处理差异
- [ ] 真实冒烟（工单 23 的工具）：claude-opus-4-8 与 claude-fable-5 各跑带工具的多轮与中途工具变更；至少一个 gpt-5.4/5.5 与一个 gpt-5.6/gpt-6 模型跑 Responses 与 Chat 两组合；gpt-5.4-pro 在 Chat 上不可用的假设一并确认。厂商拒绝任何一项时，该模型不列入，结论记入 ADR-0018 与支持矩阵

## Comments

**2026-10-03 — implemented; live smoke not run.** Status is ready-for-human. Everything except the real-API smoke is done. This environment has no test keys, so the last checkbox waits for the maintainer.

- **Catalog** `2026-10-03.1` (`ai/catalog.go`). It adds claude-opus-4-8 and claude-fable-5 on Messages, and the ten OpenAI models on Responses. The nine without gpt-5.4-pro are on Chat.
  - Chat now leaves out OpenAI models whose pi id ends in `-pro`, instead of two hard-coded ids.
  - The models only use compat fields that already exist, so `ModelCompat` is unchanged. The `BuiltinCatalog` and `SupportsMidConvoSystemMessages` comments now state ADR-0018's criterion.
  - The snapshot (`ai/release/catalog-snapshot.json`) and the pin (`testdata/responses/usage.json`) are regenerated.
- **Field check against pi.** With `BARNESS_AI_PIDIFF=1`, `TestCatalogInclusion/builtin-models-are-pi's` compares every built-in model with frozen pi's own model data. That covers name, reasoning, input, limits, level map, carried compat and prices. The data comes from a new runner entry, `models` (`pioracle.Oracle.Models`). Mutating a name, a level map entry or a limit fails it.
- **Offline E2E** (`TestCatalogInclusion`, `testdata/{anthropic,responses,chat}/catalog.json`, 23 scenarios). Responses joins the shared scenario format for this file only (`responsesProtocol`).
  - Anthropic, both models: no tools, tools, and a mid-conversation removal plus addition. Also opus-4-8 simple `xhigh` and fable-5 simple with no level (off is null, so no thinking switch-off is sent). opus-4-8 drops the temperature. No `Anthropic-Beta` header is sent.
  - Responses (gpt-5.4, gpt-6-sol) and Chat (gpt-5.5, gpt-5.6-terra): tools with input above the 272000-token tier, an addition-only change, and a removal. Responses also covers gpt-6-sol's long cache retention as `prompt_cache_options`, gpt-5.4-pro's level clamping and gpt-6-astra's null off level.
- **Differential.** The new scenarios join the P01, P02 and P04 gates. Chat and every case without an optional feature match pi.
  - Nine `extension` decisions are added, each limited to the cases that show it (ADR-0018 决策四).
  - Anthropic: the beta header and `Content-Length`; `tools` (placeholder and deferred tools); `messages[3].content` (`tool_removal`/`tool_addition`); `fallbacks`.
  - Responses: `Content-Length`; `input[4]` and `input[5]` (pi's `additional_tools` item ahead of the system message); `tools[1]`.
  - Issues 26 and 28 already say to delete their entries.
- **Docs.** `differences.md` §1 counts are updated (P01 61, P02 43). §6 now has an "未实现的可选特性" row and a narrower "未纳入目录的模型" row (issue 27). ADR-0018 gains an implementation section.
- **Other test changes.** Three Anthropic preflight tests used claude-fable-5 as a model the binding does not allow; they now use the made-up `claude-legacy-x`. The test `primary` binding allows the four new Responses models that the scenarios use.
- **Live smoke (to run).** New scenarios in `ai/live/catalog_test.go`, registered in `combos_test.go`:
  - `tool-changes-<model>`: a multi-turn tool conversation with a mid-conversation removal and addition. It checks that the current tools are declared, the system message is sent and no optional-feature marker appears. It also checks that no `Anthropic-Beta` header is sent. It runs on opus-4-8 and fable-5 (anthropic-messages), and on gpt-5.4 and gpt-6-sol on both openai-responses and openai-chat.
  - `pro-model-unavailable` (openai-chat): a probe client whose catalog lists gpt-5.4-pro on Chat must get a refusal of the model (404, or 400 naming the model). Any other failure fails the check.
  - The largest combination now makes up to 16 calls, so `maxCalls` is raised from 24 to 32.
  - Run `anthropic-messages`, `openai-responses` and `openai-chat`, then merge the reports. If a vendor refuses one of these models, take it out of the catalog under a new version and record the result in ADR-0018 and the matrix.
- **Verified.** `go vet ./...` (with and without `-tags live`), `go test ./...`, and `BARNESS_AI_PIDIFF=1 go test ./...` all pass. The release gate was not run: it still needs live bundles for the three combinations above.
- **Review (2026-10-03).** Standards and spec reviews found no hard violations or missing offline requirements. Changes made after the review:
  - The pro probe only counts a model refusal.
  - The live leak check covers the `Anthropic-Beta` header.
  - The same two OpenAI models run live on both protocols.
  - The test's pro check uses the `-pro` rule instead of a list.
  - Small helper cleanups.

  Left as known limits:
  - The `-pro` rule is a naming convention in pi's ids, not a pi data field.
  - The pi field check only covers the flags `ModelCompat` carries, so it cannot catch a hard constraint barness does not model.
  - The first live turn relies on automatic tool choice.

  The catalog lists the models before the live smoke has confirmed them. ADR-0018 决策三 takes a model out again if a vendor refuses it.
