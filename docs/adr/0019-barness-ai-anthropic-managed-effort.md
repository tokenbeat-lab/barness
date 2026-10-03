---
status: accepted
date: 2026-10-03
---

# barness-ai Anthropic 托管推理强度：每轮强度随消息记录，按 Provider 回放，不属于原生状态

pi-ai 0.87.1 对 compat 开启 `supportsMidConvoEffort` 的 Anthropic 模型（claude-fable-5-1、claude-opus-5、claude-opus-5-5）改用"托管强度"：每轮的 effort 写进对话，而不是只作为请求级参数。ADR-0018 决策二把它归为硬约束，这三款模型要等它实现后才能列入目录。工单 27 实现它。实现时需要判定新字段 `providerThinkingLevel` 是否属于 spec I6 的原生状态。

## 决策一：请求形状与 pi 一致

基线是 `anthropic-messages.ts` 中的 `stream`、`buildParams`、`convertMessages`、`insertThinkingLevelMessages` 和 `getBetaFeatures`。

- `ModelCompat.SupportsMidConvoEffort` 携带该开关。
- 本轮强度是 `AnthropicOptions.Effort`，缺省为 high。adapter 在 `stream` 开始时把它写进 `AssistantMessage.ProviderThinkingLevel`，所以失败消息也带有该字段，与 pi 在创建 output 时写入一致。
- thinking 恒为 `{type: adaptive, display, block_binding: {prefix_mismatch_behavior: drop_block}}`，请求级 `output_config` 恒为 `{effort: high}`。`thinkingEnabled` 被忽略：false 或 simple 入口没有等级时，thinking 也不会关闭。temperature 从不发送，即使模型数据没有 `supportsTemperature: false`（fable-5-1）。
- 对于每条回放的 assistant 消息，如果它来自同一 Provider 的 Anthropic Messages，并且记录的强度是五个 Anthropic effort 之一，就在它前面插入 `{role: system, content: [], output_config: {effort}}`。对话末尾插入一条带本轮强度的同类消息。插入发生在放置缓存标记之后，所以缓存断点仍在最后一条真实消息上。中途 system 消息照旧延后到下一条 assistant 之前，所以强度标记不会把 tool_use 和 tool_result 隔开。
- 请求头加 `mid-conversation-output-config-2026-07-01` 与 `thinking-binding-controls-2026-08-01`。
- simple 入口沿用 adaptive 的 effort 映射：level map 中为 null 的等级不算映射值，按 pi 的默认映射处理。例如 opus-5-5 的 minimal 映射为 low。

## 决策二：`providerThinkingLevel` 不属于原生状态，与 pi 一样按 Provider 回放

spec I6 只让可信封套下的原生状态按同模型回放，例如推理签名与密文、redacted 块、厂商条目 ID。`providerThinkingLevel` 不在此列：

- 它不是厂商签发的值，也不绑定账户。它只是宿主自己为这一轮选择的 effort，五个枚举值之一，不含密文，也不引用厂商侧状态。
- 伪造它得不到任何权限。宿主自报一个强度，只影响这个租户自己的请求怎样描述它自己的历史。前缀与强度不一致时，厂商按 `drop_block` 丢弃对应的 thinking 块，请求不会失败。
- pi 按 `api === "anthropic-messages"` 且 Provider 相同来回放它，不比较模型。降级跨模型或无封套的消息时，它的签名 thinking 转成文本，但强度标记仍要保留：这条消息确实是以那个强度生成的，去掉标记反而会让厂商把它当作本轮强度的输出。

所以：

- 它随消息序列化（JSON `providerThinkingLevel`）。宿主读回时不需要 `TrustNativeState`，也不计入 `NativeStateDowngrades`。只带这个字段的消息不算携带原生状态。
- 回放规则与 pi 相同：同 Provider 的 Anthropic Messages 消息，不论模型，也不论原生状态是否降级，都插入它的标记。其他 Provider 或 API 的消息不插入。
- 外部输入的校验与 pi 相同：不是五个 effort 之一的值被忽略，不插入标记，调用照常进行。被忽略的值从不发送，所以不会把未校验的输入带上线。

## 决策三：本轮强度不受可选特性影响

三款模型也开启了 `supportsMidConvoToolChanges`，那是可选特性（ADR-0018 决策二，工单 26）。带工具的请求中，pi 多出的占位工具和 tool-changes beta 按 ADR-0018 决策四登记为扩展，限于本工单带工具的用例。工单 26 实现后删除这些登记。

## Considered Options

- 把 `providerThinkingLevel` 当作原生状态，要求可信封套才回放：被降级的同 Provider 历史会丢失强度标记，导致与 pi 不一致的请求；这个值又不带任何需要担保的内容，担保要求没有安全收益。
- 按模型而非 Provider 回放：比 pi 更严格，换模型时会丢失历史强度。厂商按 Provider 理解这个标记，没有理由收紧。
- 与 pi 一致，按 Provider 回放，不需要封套（采用）。
- 拒绝非法强度值（invalid_request）：pi 会忽略这种值并发出请求，拒绝会形成待处理差异。忽略是安全的，因为这个值从不上线。

## Consequences

- claude-fable-5-1、claude-opus-5、claude-opus-5-5 列入内置目录 `2026-10-03.2`。字段由 `TestCatalogInclusion/builtin-models-are-pi's` 与冻结 pi 的数据逐字段比较，`SupportsMidConvoEffort` 也在比较范围内。
- 宿主存储 assistant 消息时应保留 `providerThinkingLevel`。缺失时与 pi 的旧消息一样不插入标记，请求仍然有效。
- 离线 E2E 是 `TestManagedEffort`（`ai/e2e/testdata/anthropic/managed-effort.json`），场景全部进入 P02 差分。真实冒烟场景 `effort-changes-<model>` 在 anthropic-messages 组合中。厂商拒绝强度变更时，从目录中撤下该模型，结论记入本 ADR 与支持矩阵。

## 真实冒烟（2026-10-03，账户 prod-anthropic）

anthropic-messages 组合的 11 个场景全部通过，共 20 次调用，没有环境重试。

- `effort-changes-<model>` 在 claude-fable-5-1、claude-opus-5、claude-opus-5-5 上都通过。第一轮以 high 运行并返回签名 thinking；第二轮回放第一轮，换成 low。两轮都是 HTTP 200，请求头带两个 beta，第二轮请求中的强度标记依次为 `high,low`，消息分别记录 high 与 low。厂商接受了对话中途的强度变更，三款模型留在目录中。
- 首次运行时，冒烟检查按 `Anthropic-Beta` 读取录制的请求头，但录制器以小写保存头名，所以 beta 检查读到的总是空值。三款模型实际都发送了两个 beta。工单 34 的 `tool-changes-*` 用同样的写法检查"不发送 beta"，那项检查因此一直是空检查。两处都已改为 `anthropic-beta`，重跑后 opus-4-8 与 fable-5 确实没有发送 beta。
- 报告已合并进支持矩阵。
