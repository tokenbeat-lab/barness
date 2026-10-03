---
status: accepted
date: 2026-10-02
---

# barness-ai Anthropic Messages adapter：SDK 只发请求、流自行解码，用量随事件记录

工单 16 接入 Anthropic × Messages（P02）。spec I5 要求用 Anthropic Go SDK（研究锁定 v1.75.0）承担 HTTP，关闭默认重试，并由模块自行判定协议终态；spec I7 要求替换 payload 后仍强制 `stream=true`、onResponse 时点与 Responses 一致，非 OpenAI-compatible 路径忽略 samplingParams。spec 没有规定流由谁解码、beta 特性如何与回调交互、Anthropic 的"用量完整"含义，以及哪些 pi 模型可以列入目录。本 ADR 记录实现时定下的做法，维护者于 2026-10-02 确认（见文末）。

## 决策一：SDK 发请求与读错误体，SSE 由 adapter 按 pi 自己的解码器解码

请求经 `anthropic.NewBetaMessageService(...).NewStreaming` 发出：`WithMaxRetries(0)`，请求体是 adapter 自己编码的字节（`WithRequestBody`），认证放在 `X-Api-Key` 头里，不调用会读取 `ANTHROPIC_*` 环境变量与凭据文件的 `anthropic.NewClient`。SDK 负责 Stainless 头、`anthropic-version` 与非 2xx 错误体读取；响应体经与 Responses 相同的 body 限额和时限中间件。

事件流不用 SDK 的 `ssestream`，而是移植 pi 的 `iterateSseMessages`/`iterateAnthropicEvents`：按 SSE 事件名过滤、`parseJsonWithRepair` 修复 JSON、`error` 事件以其 data 失败、`message_start` 之后缺 `message_stop` 为错误、正文末尾没有空行的最后一个事件仍被派发。pi 自己也不用其 SDK 的解码器。

### Considered Options

- 用 SDK 的 typed `Stream`：只认识固定事件名，`error` 事件被转成 SDK 错误，未知事件与修复 JSON 无法对齐 pi。
- 用 SDK 的 `ssestream.Decoder`：丢弃无结尾空行的最后一个事件，并给 data 追加换行，与 pi 不同（pi 测试 "ignores unknown SSE events after message_stop" 依赖前者）。
- 直接 HTTP：spec I5 只允许在 SDK 无法满足时采用；SDK 发送请求没有缺口。
- SDK 发请求 + 自行解码（采用）。

### Consequences

- 行以 LF 结束（CR 在 LF 前被去掉），与 `streamBody` 的帧计数一致；pi 还把单独的 CR 当作行尾，Anthropic 不会发送这种流。
- spec I5 写的是 SDK 承担"HTTP 与流解码"，SDK 不能满足时才改直接 HTTP；这里 HTTP 仍由 SDK 承担，只有流解码改由 adapter 完成，理由是 SDK 的两种解码方式都与 pi 的可观察行为不同。这是对 spec 措辞的收窄，已经维护者确认并同步到 spec §5。
- 成功终态按 pi 判定：收到 stop reason 即可，`message_start` 之后才要求 `message_stop`。没有 `message_start` 却有 stop reason、也没有 `message_stop` 的流与 pi 一样算成功；工单"以 message_stop 判定成功终态"应理解为"按 pi 的 message_stop 规则"。
- `NewStreaming` 总是把 `stream` 设为 true。pi 只在 onPayload 返回替换体时强制，原位改成 `false` 时会照发；barness 在两种情况下都发 `true`（adapter 只读事件流，与 Responses 把 `stream` 当作自身契约一致）。

## 决策二：beta 特性以请求头为准，payload 回调看到 pi 的 `betas` 参数

`headers` 按 pi 的 `getBetaFeatures` 计算 `anthropic-beta`（仅预算型 thinking 需要 interleaved-thinking）；可信的 header 变换改写它时，改写值去空格、去空项、去重后取代计算值，等同 pi 中配置的 `anthropic-beta` 头。配置了 onPayload 时，请求体额外带 `betas` 数组供回调查看或修改，回调结束后从请求体移回请求头，与 pi 的 SDK 行为一致；没有回调时请求体从不含 `betas`。回调给出的 `betas` 与 pi 的 SDK 一样按 `toString()` 原样拼接（配置的头才去重）。`anthropic-beta` 不在受保护头列表中：它开启的特性若引用厂商侧状态，仍要经过请求体授权（决策四）；但不以请求体字段出现、只改变厂商侧行为的 beta 不受检查。回调属于可信宿主，所以没有设 beta 白名单。

## 决策三：Anthropic 的用量在 message_start 与每个 message_delta 时记入消息与尝试

pi 在 `message_start` 取初始用量，之后每个 `message_delta` 替换其报告的非 null 计数并重算总量与成本，所以中断的流仍保留已报告的输入用量。barness 在每次更新后同时写入消息 Usage 与当前尝试的 `Attempt.Usage`，保持 ADR-0010 的"消息用量等于读取流那次尝试的用量"。

完整性：`input_tokens`、`output_tokens`、`cache_read_input_tokens`、`cache_creation_input_tokens` 都至少报告过一次（0 也算）为 `complete`；收到过 usage 但缺其中之一为 `partial`；从未收到 usage 为 `unreported`。`cache_creation.ephemeral_1h_input_tokens` 与 `output_tokens_details.thinking_tokens` 是较晚加入的明细，缺失读作没有，不影响 `complete`（与 ADR-0010 对 `cache_write_tokens` 的处理一致）。

### Consequences

- pi 对缺少 `message.usage` 的 `message_start` 抛出 JavaScript TypeError；barness 按 0 计并保持 `unreported`，调用照常进行。真实 API 总是发送 usage，差分未覆盖这一形态。

## 决策四：payload 回调的授权

回调产出的请求体必须仍指向授权模型；只能声明无 `type` 或 `type: "custom"` 的工具，其他工具类型须由绑定的 `AllowedHostedTools` 放行（ADR-0005）；不能出现 `container`、`mcp_servers`、`fallbacks`，也不能在 system/messages 任何位置出现 `file_id` 或 `type: "file"` 的来源——它们指向执行容器、带自身地址与令牌的 MCP 服务器、服务端回退模型或厂商账户上的文件，库无法替多租户共享的账户担保。

## 决策五：内置目录只列 adapter 能完整服务的 Anthropic 模型

列入 pi 数据中的 claude-haiku-4-5(-20251001)、claude-opus-4-5(-20251101)、claude-opus-4-6、claude-opus-4-7、claude-sonnet-4-5(-20250929)、claude-sonnet-4-6、claude-sonnet-5，目录升为 `2026-10-02.2`。`ModelCompat` 增加 `ForceAdaptiveThinking` 与 `SupportsTemperature`（未设置或 null 为 true，与 pi 一致）。不列入需要原生中途工具变更、托管中途 effort 或服务端回退模型的 claude-fable-5、claude-fable-5-1、claude-opus-4-8、claude-opus-5、claude-opus-5-5；`supportsStrictTools` 只影响 barness 无法声明的约束采样工具，不携带。

> 已被 ADR-0018 部分取代（2026-10-03）：列入标准改为"遵守模型硬约束即可列入"。claude-opus-4-8 与 claude-fable-5 由工单 34 列入；托管强度是硬约束，claude-fable-5-1、claude-opus-5、claude-opus-5-5 已由工单 27 按 ADR-0019 实现后列入。

## 决策六：选项形状

`AnthropicOptions` 对应 pi 的 `AnthropicOptions`，去掉传输、凭据、头、回调、重试与注入 client。pi 的 `sessionId` 只用于 Anthropic 官方 API 不接受的会话亲和头，`samplingParams` 在该协议被忽略（spec I7），两者都没有字段；simple 入口的同名字段同样不产生请求字段。`thinkingEnabled`、`interleavedThinking` 用 `Nullable` 区分未设置、false 与 true，null 与未设置相同（pi 的 `?.`/`??` 语义）。`X-Stainless-Timeout` 与 pi 一样总是声明（默认 600 秒），而 Responses 只在显式 timeoutMs 时声明，这是两个 TypeScript SDK 在 pi 中的实际差别。

## 维护者决定（2026-10-02）

1. 决策一：采纳。SSE 由 adapter 按 pi 解码，spec §5 已同步；`stream` 一律发送 true。
2. 成功终态与 pi 一致：`message_start` 之后才要求 `message_stop`。
3. 决策二：信任可信回调，不设 beta 白名单。
4. 决策三：按现有实现（四个计数定义完整性，缺少 `message.usage` 时按 0 计并标为未上报）。
5. 决策五：不列入的五个模型按所需能力拆为三张工单，依次纳入：工单 26（原生中途工具变更，claude-opus-4-8）、工单 27（托管推理强度，claude-fable-5-1/opus-5/opus-5-5，依赖 26）、工单 28（服务端备用模型，claude-fable-5，依赖 26）。备用模型只发送 binding 同样允许的那些（spec §4 授权交集）。工单依赖顺序已被 ADR-0018 取代。

差分中只在 Anthropic 出现的差异（运行时错误文本、pi 流中块的 `index` 暂存字段、未担保原生状态降级）登记在 `ai/e2e/testdata/pidiff/ledger.json`。
