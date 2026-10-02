# 26: Anthropic 对话中途原生工具变更（claude-opus-4-8）

**What to build:** 支持中途 system 消息与原生工具变更的 Anthropic 模型（pi compat `supportsMidConvoSystemMessages` + `supportsMidConvoToolChanges`），按冻结 pi 在对话中原地声明工具的增删，使前缀缓存不因工具变化失效；完成后把 claude-opus-4-8 列入内置目录（ADR-0011 决策五，维护者决定 2026-10-02）。

**Blocked by:** 16

**Status:** ready-for-agent

**Context:** 基线为 pi-ai 0.87.1 `anthropic-messages.ts` 的 `buildParams`/`convertMessages`/`getBetaFeatures` 与 `utils/transcript.ts` 的 `hasToolRedefinitions`、`getDeclaredTools`。目前 adapter 只把中途 system 消息的文本按 pi 规则延后到下一条 assistant 之前，不渲染工具变更；`ModelCompat` 没有 `supportsMidConvoToolChanges`。注意：只要请求带工具，这类模型的请求就已不同（占位工具与 beta），不只在工具中途变化时才不同。

- [ ] `ModelCompat.SupportsMidConvoToolChanges`；启用条件与 pi 一致：两个开关都开、初始 system 消息声明了工具、历史中没有同名工具的重定义
- [ ] 启用时：初始工具照常声明并在最后一个上加 cache_control，随后声明 pi 的 `__pi_deferred_placeholder__`（`defer_loading: true`），再声明之后加入的工具（`defer_loading: true`、无 cache_control）；中途 system 消息带 `tool_removal`/`tool_addition` 块（`tool_reference` 名称），与文本一起延后到下一条 assistant 之前；最后一条 system 消息的末块（含 tool_addition/removal）可带 cache 标记
- [ ] 请求头加 `mid-conversation-tool-changes-2026-07-01` beta；未启用时（有重定义或无初始工具）照旧发送当前工具列表
- [ ] payload 授权放行 adapter 自己声明的 `defer_loading` 工具，不放宽托管工具与厂商侧引用的拒绝（ADR-0005）
- [ ] claude-opus-4-8 列入目录（目录版本与 pin 同步升级）
- [ ] 离线 E2E：无工具、工具不变、中途加/删工具、重定义回退、与图片/工具结果混排；pi 差分无待处理差异
