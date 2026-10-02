---
status: accepted
date: 2026-10-02
---

# barness-ai DeepSeek × Chat Completions：复用 Chat adapter，按 Provider 移植 pi 的 DeepSeek compat，作为冻结 pi 路由进入差分

工单 22 接入 DeepSeek × Chat Completions（P06）。spec I4 要求它复用 Chat adapter，与 DeepSeek Responses 使用不同 BindingID、可经授权引用同一账户凭据，模型配置保持协议维度独立；spec P06 要求 reasoning_content 增量与带工具调用/结果的 assistant 回放、`[DONE]` 前 usage 不遗漏、不混用 Responses 配置，thinking 与工具选择的限制单独验收。ADR-0013 决策三只实现了标准 OpenAI compat，把 DeepSeek 的 compat 留给本工单。spec 没有规定这些 compat 放在哪里、目录取哪些模型、thinking 模式下的强制工具选择如何对待，以及 DeepSeek 的提示缓存字段。本 ADR 记录实现时的做法，维护者于 2026-10-02 确认（见文末）。

## 背景：pi 的 DeepSeek 路由

与 Responses 不同，DeepSeek × Chat 是冻结 pi 自带的路由：pi 0.87.1 的 `providers/data/deepseek.json` 只在 `openai-completions` 上登记 `deepseek-flash` 与 `deepseek-v4-pro`。`detectCompat` 以 provider `deepseek` 或 baseUrl 含 `deepseek.com` 识别 DeepSeek，得到：不发 `store`，不用 `developer` 角色（推理模型的指令仍为 `system`），输出预算字段为 `max_tokens`，`thinkingFormat: "deepseek"`（有推理等级时发 `thinking: {type: "enabled"}` 与经 level map 的 `reasoning_effort`；没有时，除非 level map 的 off 为 null，发 `thinking: {type: "disabled"}`），`requiresReasoningContentOnAssistantMessages`（推理模型的每条回放 assistant 消息都带 `reasoning_content`，没有推理时为空串）；`supportsReasoningEffort` 与 `supportsLongCacheRetention` 为真，不发 session 亲和头。deepseek.json 中两个模型的 compat 逐项重复了这些探测值，另加 `supportsStrictMode`（两者）与 `supportsMidConvoSystemMessages`（v4-pro）。

研究记录（harness `docs/research/pi-ai-go-sdk-validation.md`）在真实服务上验证过：`deepseek-flash` 的文本、关闭 thinking 的强制工具调用、首帧后取消、`reasoning_content` 流、流末 usage，以及带 `reasoning_content`、真实工具调用与工具结果的下一轮回放被服务端接受；并记录 DeepSeek Chat 在 thinking 模式下不能强制 `tool_choice=required`。

## 决策一：DeepSeek compat 按 Provider 派生，不进入模型 compat

`chatCompat`（`ai/chat_compat.go`）是 pi `detectCompat` 中随厂商变化、且 barness 已有对应行为的五项：`store`、`developer` 角色、`max_tokens` 字段、deepseek thinking 格式、assistant 消息上的 `reasoning_content`。`chatCompatOf` 对 DeepSeek 给出 pi 的探测值，其余 Provider 是 ADR-0013 的标准 OpenAI compat；Chat adapter 从调用模型的 Provider（即 binding 的 Provider）取得它。模型 compat 中 barness 已承载的 `supportsStrictMode`、`supportsMidConvoSystemMessages`、`supportsLongCacheRetention` 照旧来自目录。

与 ADR-0014 的 Responses 能力一样按 Provider 而不是模型确定：这些是厂商端点的事实，pi 的 DeepSeek 数据也只是重复探测值；放进 `ModelCompat` 会让宿主目录可以给 DeepSeek 模型换上 OpenAI 的请求形状（或反过来），而且五个标志只有 DeepSeek 一个真实用例。pi 识别的其他十余家厂商仍不移植（它们的 Chat 模型不在目录中）；出现第二个非标准厂商时再评估是否把结构扩为可配置。

### Considered Options

- 在 `ModelCompat` 中加入 pi 的 `supportsStore`、`supportsDeveloperRole`、`maxTokensField`、`thinkingFormat`、`requiresReasoningContentOnAssistantMessages`：与 pi 数据同形，但宿主目录可以关掉 DeepSeek 必需的 `reasoning_content`，或在 OpenAI 自定义模型上意外打开 deepseek thinking 格式；且 pi 自己也先按厂商探测。
- 独立的 DeepSeek Chat adapter：违反 spec "不复制整套生命周期"，差异只在请求字段。
- 按 Provider 的 compat 结构（采用）。

## 决策二：内置目录

目录 `2026-10-02.6` 新增 DeepSeek × Chat 的 `deepseek-flash` 与 `deepseek-v4-pro`，取 pi `deepseek.json`（sha256 `549a7ddb…4d0d`）中两者的名称、reasoning、level map（都没有 off 项，所以不带等级的调用关闭 thinking）、输入（flash 文本与图片，v4-pro 只有文本，图片走占位）、价格、上下文与输出上限，compat 只保留 barness 承载的 `supportsStrictMode` 与 v4-pro 的 `supportsMidConvoSystemMessages`。v4-pro 不锚定新增工具（pi 的 `supportsMidConvoToolAdditions` 未设），Chat 上不发送追加的工具声明，barness 可以完整服务它，`ModelCompat` 的注释相应更新。DeepSeek Responses 的 `deepseek-flash` 条目保持独立字面量（内容与 Chat 条目相同），不由 Chat 条目派生：ADR-0014 决策三可能按冒烟结果单独修改 Responses 的 level map。两个协议的条目是目录中各自的 (Provider, API, ID)，宿主改动其一不影响另一个（P06-E10 用例证明）。

## 决策三：thinking 模式下的强制工具选择不在发送前拒绝

barness 与 pi 一样照发 `reasoningEffort` 与 `toolChoice: required` 的组合，DeepSeek 拒绝时（fixture 中为 400）以 `invalid_request` 报告，不重试。用例按 spec P06 分开：强制工具（`required`、具名函数）在关闭 thinking 时运行，thinking + 工具的自动选择、thinking 下的工具往返与推理历史回放另跑，两者互不证明；被拒的组合单列一个 fixture。

ADR-0014 决策二拒绝显式 `ServiceTier`，是因为 DeepSeek 会静默忽略它而计价会出错；这里厂商自己报错，调用者看到的是明确失败，发送前拒绝反而把一项研究时点的厂商限制写死在 adapter 中，并与 pi 产生差分。fixture 中的错误文本是示意；若工单 23 的冒烟发现 DeepSeek 静默忽略该组合（返回非强制的结果），改为发送前拒绝并更新本 ADR。

## 决策四：提示缓存字段沿用 pi

pi 对 DeepSeek 探测出 `supportsLongCacheRetention: true`，因此 session id 加 long 保留期时发送 `prompt_cache_key` 与 `prompt_cache_retention: "24h"`（short 保留期在 DeepSeek 端点上不发，DeepSeek 不是 `api.openai.com`）。barness 照此发送，缓存键与其他协议一样按租户和账户派生（spec §9）。DeepSeek 的 Chat 文档没有列出这两个字段（它的上下文缓存是自动的），预计被忽略；这与 ADR-0014 中 DeepSeek Responses 不派生缓存字段不同——那里 DeepSeek 的指南明确列为忽略，且该路由没有 pi 可对照。不发送的替代做法是登记一项差分扩展；在冒烟确认 DeepSeek 拒绝它们之前，按 pi。

## 决策五：作为冻结 pi 路由进入差分

P06 不登记为扩展路由（ledger `routes` 中没有 DeepSeek × openai-completions，`P06-E10-pi-route` 断言），六个 fixture 文件（`ai/e2e/testdata/deepseek-chat`）的全部场景都以 `PIDIFF-P06-*` 运行。差分装置按 (Provider, API) 选择 binding 与 base URL，runner 的 `openai-completions` 同时查 pi 的 OpenAI 数据（换 API）与 DeepSeek 数据（原样）。差异与 P04 同类，按协议 `openai-completions` 归入账本：连接中断的运行时文本、派生缓存键（及随之变化的 Content-Length）、未担保原生状态降级沿用 P04 的决定并加入 P06 用例；新增一项：降级后的回合在 DeepSeek compat 下仍带空的 `reasoning_content`，而 pi（信任任何历史）回放原推理文本。

## Consequences

- OpenAI 的 Chat 请求不变（P04 差分 0 pending）；`chatBody.Store` 改为可省略的指针，并新增 `max_tokens` 与 `thinking`。
- DeepSeek 文档或 pi 的探测更新时，改 `chatCompatOf`、目录与 P06 fixture，并记录差异。
- DeepSeek 的错误体（401 的掩码 key 后缀、402 余额不足、503 过载）按其错误码文档以 OpenAI 错误对象形状编写，文本示意；402 按通用 4xx 归为 `invalid_request`。usage 同时覆盖在 finish chunk 内与其后独立 chunk 两种位置；实际位置、错误体与是否有厂商请求 id 头由工单 23 的冒烟确认。
- 中途断流的错误文本仍写 "OpenAI Chat Completions"（协议名），与 P04 一致。

## 维护者决定（2026-10-02）

1. 决策一：采纳。DeepSeek 的 Chat compat 按 Provider 派生（`chatCompatOf`），不进入模型 compat。
2. 决策二：采纳。目录列入 `deepseek-flash` 与 `deepseek-v4-pro`（含 v4-pro 的 mid-conversation system），版本 `2026-10-02.6`；v4-pro 的 `high`/`max` 由工单 23 的真实冒烟确认。
3. 决策三：采纳，附条件。thinking + 强制工具选择照发，由 DeepSeek 拒绝；若工单 23 的冒烟显示 DeepSeek 静默忽略强制选择，改为发送前以 `invalid_request` 拒绝并更新本 ADR 与 P06 fixture。
4. 决策四：采纳，附条件。long 保留期下照 pi 发送派生的 `prompt_cache_key` 与 `24h` 保留期；若工单 23 的冒烟显示 DeepSeek 拒绝这两个字段，改为不发送，并在差分账本登记为扩展。
