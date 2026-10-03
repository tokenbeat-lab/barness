# 27: Anthropic 托管推理强度（claude-fable-5-1、claude-opus-5、claude-opus-5-5）

**What to build:** 支持每轮推理强度可变的 Anthropic 模型（pi compat `supportsMidConvoEffort`）：记录每条 assistant 消息当时的强度，回放历史时按 pi 插入强度标记，使对话中途调强度不破坏 thinking 绑定；完成后把三款模型列入内置目录。按 ADR-0018 托管强度是硬约束（pi 恒用 adaptive thinking 加 `drop_block`，以免前缀不匹配时持续 400，并且从不发送 temperature），所以本工单仍是这三款模型的列入条件。

**Blocked by:** 16

**Status:** ready-for-agent

**Context:** 基线为 pi-ai 0.87.1 `anthropic-messages.ts`：`stream` 中 `providerThinkingLevel`、`buildParams` 的 managed effort 分支、`insertThinkingLevelMessages`、`convertMessages` 的 `assistantLevels`、`getBetaFeatures`。三款模型同时开启了工单 26 的工具变更开关。按 ADR-0018 那是可选特性，不阻塞本工单：26 未完成时按 pi 的回退发送当前工具列表，差异按工单 34 的方式登记为扩展。

- [ ] `ModelCompat.SupportsMidConvoEffort`；`AssistantMessage` 增加 pi 的 `providerThinkingLevel`（本轮强度：`effort` 选项，缺省 high），随消息序列化与回放
- [ ] 请求：thinking 恒为 `{type: adaptive, display, block_binding: {prefix_mismatch_behavior: drop_block}}`，`output_config: {effort: "high"}`；每条同 Provider 的 Anthropic 历史 assistant 消息前插入 `{role: system, content: [], output_config: {effort: <其强度>}}`，末尾插入本轮强度；不发送 temperature
- [ ] 请求头加 `mid-conversation-output-config-2026-07-01` 与 `thinking-binding-controls-2026-08-01` beta
- [ ] 判定 `providerThinkingLevel` 是否属于原生状态（spec I6：需可信封套才回放，还是与 pi 一样按 Provider 回放），结论记入 ADR
- [ ] simple 入口沿用现有 adaptive effort 映射（含 opus-5-5 的 minimal 为 null）
- [ ] 三款模型列入目录（目录版本与 pin 同步升级）
- [ ] 离线 E2E：首轮、带不同强度历史的多轮、跨模型历史、full/simple；pi 差分无待处理差异
