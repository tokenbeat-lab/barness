---
status: accepted
date: 2026-10-02
---

# barness-ai 请求体回调中的托管工具由绑定显式放行

可信宿主（上层 harness）通过 onPayload 改写 adapter 已构建的原生请求体，用于协议扩展。pi 对改写内容不设限制；barness-ai 是多租户的，同一厂商账户可能由多个租户共享，库无法判断托管资源（向量库、容器）或外部网络目标（MCP server）归属于谁。spec §7 只要求回调后的最终模型和原生引用必须通过授权复核，没有规定托管工具应如何处理。维护者于 2026-10-02 决定：默认拒绝，由绑定显式放行。

## Considered Options

- 一律拒绝：最安全，但会挡住 harness 常用的 web_search 等扩展，与 User Story 35 的目的冲突。
- 一律放行：与 pi 一致，但回调能借托管工具引用他人的厂商资源或访问未授权的网络目标，等于绕过了绑定授权。
- 由绑定显式放行（采用）：`Binding.AllowedHostedTools` 列出允许的托管工具类型，以协议原生的名称表示（如 Responses 的 `web_search`）。绑定由可信宿主按租户解析，列入某个类型即表示宿主担保该账户的托管资源和该工具访问的网络目标归这个租户使用。

## Consequences

- 默认值为空，即拒绝所有托管工具；barness-ai 自身从不声明托管工具。允许列表只放宽 `tools` 的类型检查。模型、缓存键、`store: false`，以及 `previous_response_id`、`conversation`、`prompt`、`background`、`item_reference`、`file_id` 等厂商侧状态引用仍一律拒绝，因为它们与无状态调用和原生状态担保（ADR-0001）冲突。
- 放行只按类型判断，不检查工具的参数（如 `vector_store_ids`、`server_url`），参数的归属由宿主担保。如果以后需要按参数细分授权，另立决策。
- 列表中出现空字符串或 `function` 属于宿主配置错误，调用在 PhaseBinding 以 `invalid_request` 失败，回调不会执行。
- 允许列表随绑定快照固定，在调用开始时复制，之后绑定更新只影响新的调用。
- 托管工具输出的流事件和条目与 pi 一样被忽略，不进入消息，也不进入回放历史。如果以后需要呈现这些输出，另作扩展登记。
- 这是 barness-ai 的租户授权扩展，不属于 pi 兼容行为；pi 差分不覆盖这部分。Anthropic、Chat、Gemini 等 adapter 接入 onPayload 时沿用同一字段，按各自协议的工具类型名称判断。

行为与验收见 [spec §7](../../.scratch/barness-ai/spec.md#7-完整选项统一选项与可信回调) 和 E04（`TestTrustedCallbackHostedTools`）。

## 工单 05 已交付扩展（2026-10-08）

Classify 沿用三种可信回调与 headers → payload → response 顺序，均带入口 Operation。
TypeSafe 没有托管工具授权：最终 body 只接受模型、状态和问题集合。改变授权模型/认证或添加授权字段为
tenant_denied；无法解码、非法问题或超出分类容量为 callback_failed。最终问题集合独立冻结，答案按它校验。

详见 [ADR-0021](0021-barness-ai-typesafe-unary-classification.md) 与 P07 离线证据。

## 工单 09：原生图像生成（2026-10-08）

图像生成回调只可改变允许的生成字段，不能改授权型号、端点或操作，也不能加入流、partial images、user、外部资源或续接状态。以最终冻结请求校验数量和格式；无效回调为 callback_failed，扩大授权为 tenant_denied。

详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。
