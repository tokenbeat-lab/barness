# 34: 按硬约束标准列入被可选特性挡住的模型（claude-opus-4-8、claude-fable-5 与十个 OpenAI 模型）

**What to build:** 按 ADR-0018 的列入标准，把此前仅因可选协议特性未实现而排除的模型列入内置目录：Anthropic 的 claude-opus-4-8、claude-fable-5，OpenAI 中需要 additional_tools/tool search 的十个模型。这些模型按 pi 中对应特性关闭时的行为发送请求；与 pi 的差异登记为扩展；用真实冒烟确认厂商接受这些请求。

**Blocked by:** 16, 20

**Status:** ready-for-agent

**Context:** 基线为 pi-ai 0.87.1 的 `anthropic.json`、`openai.json`（哈希见 `ai/catalog.go` 的 `BuiltinCatalog` 注释）。不需要改 adapter：`supportsMidConvoToolChanges` 关闭时 pi 发送当前工具列表（`buildParams`）；`allowedFallbackModels` 不发送时请求有效；`supportsAdditionalTools`/`supportsToolSearch` 关闭时 pi 同样发送当前工具列表（`resolveTranscriptTools`）。barness 现有路径与这些回退一致。只携带 `ModelCompat` 已有的字段，其余开关（`supportsMidConvoToolChanges`、`allowedFallbackModels`、`supportsAdditionalTools`、`supportsToolSearch`、`supportsOpenAIGrammarTools`、`supportsStrictTools`）不携带，由工单 26、28 或以后的工单引入。托管强度模型（claude-fable-5-1、claude-opus-5、claude-opus-5-5）不在本工单，见工单 27。

- [ ] Anthropic × Messages 列入 claude-opus-4-8（adaptive、`SupportsTemperature=false`、level map `{xhigh, max}`、`SupportsMidConvoSystemMessages`）与 claude-fable-5（adaptive、level map `{off: null, xhigh, max}`、`SupportsMidConvoSystemMessages`）；名称、上下文、输出上限、价格逐字段与 pi 数据核对
- [ ] OpenAI × Responses 列入 gpt-5.4、gpt-5.4-mini、gpt-5.4-pro、gpt-5.5、gpt-5.6-luna、gpt-5.6-sol、gpt-5.6-terra、gpt-6-astra、gpt-6-luna、gpt-6-sol：含 `SupportsMidConvoSystemMessages`、`SupportsExplicitPromptCacheMode`（gpt-5.6/gpt-6 系列）、level map 与阶梯价格，逐字段核对
- [ ] OpenAI × Chat Completions 列入上一项中除 gpt-5.4-pro 外的模型（`builtinOpenAIChatModels` 的 pro 排除规则扩展为按 pi 数据判定，不再硬编码两个 ID）
- [ ] 目录版本升级，`ai/release/catalog-snapshot.json` 与 pin 同步
- [ ] 删除 `BuiltinCatalog` 注释与 `ModelCompat.SupportsMidConvoSystemMessages` 注释中"compat 需要未实现行为的模型不列入"的表述，改述 ADR-0018 的标准；`docs/barness-ai/differences.md` §6"未纳入目录的模型"一行改为"未实现的可选特性"（中途工具变更、服务端备用模型、additional_tools/tool search），指向工单 26、28 与 ADR-0018，托管强度模型指向工单 27
- [ ] 离线 E2E 与 pi 差分：每个新增 Anthropic 模型覆盖无工具、带工具、中途增删工具、fable-5 的普通调用；OpenAI 每族至少一个模型覆盖带工具与中途 system/工具变更。pi 多出的占位工具与 tool-changes beta、`fallbacks` 与 server-side-fallback beta、`additional_tools`/`tool_search_*` 条目按 ADR-0018 决策四登记为 `extension`，`cases` 限于这些用例；无待处理差异
- [ ] 真实冒烟（工单 23 的工具）：claude-opus-4-8 与 claude-fable-5 各跑带工具的多轮与中途工具变更；至少一个 gpt-5.4/5.5 与一个 gpt-5.6/gpt-6 模型跑 Responses 与 Chat 两组合；gpt-5.4-pro 在 Chat 上不可用的假设一并确认。厂商拒绝任何一项时，该模型不列入，结论记入 ADR-0018 与支持矩阵
