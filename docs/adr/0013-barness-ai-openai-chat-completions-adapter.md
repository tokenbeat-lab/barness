---
status: accepted
date: 2026-10-02
---

# barness-ai OpenAI Chat Completions adapter：SDK 发送、按 openai-node 读取 chunk，标准 OpenAI compat

工单 20 接入 OpenAI × Chat Completions（P04）。spec I5 由 OpenAI Go SDK（v3.66.0）承担 Chat 的 HTTP 与流解码，并要求核查 WithJSONSet、ExtraFields、ExtraBody、受控 RoundTripper 的实际可用性，显式配置 pi 默认字段并以实际请求验证；spec P04 要求以协议终态判定成功、结束信号前不遗漏流末 usage。spec 没有规定 SDK 的流解码在哪一层停下、Chat 的内置模型从哪里来、pi 按厂商探测的 compat 如何对待、Chat 的"用量完整"含义，以及回调可声明的 Chat 专有字段。本 ADR 记录实现时定下的做法，维护者于 2026-10-02 确认（见文末）。

## 决策一：SDK 发送请求并切分 SSE，chunk 由 adapter 按 openai-node 读取

对 `openai-go/v3` v3.66.0 的核查结论（依据其源码，并以 fixture 的实际请求验证）：

- 请求体：与 Responses 相同，adapter 自己编码请求 DTO，经 `option.WithRequestBody` 交给 `ChatCompletionService.NewStreaming`，SDK 的类型化参数为空。类型化参数的 `SetExtraFields`（即 ExtraFields/ExtraBody 机制，v3.66.0 没有名为 `WithExtraBody` 的选项）可以补字段，但 null、零值与未设置的区分、pi 的非标准推理字段（`reasoning_content` 等回放字段）与键序都更难保证；`option.WithJSONSet` 只作用于 `*bytes.Buffer` 请求体，与 `WithRequestBody` 配合可用——SDK 自己就在调用者选项之后追加 `WithJSONSet("stream", true)`，所以即使 payload 回调去掉或改掉 `stream`，发出的请求仍是流式的（pi 不强制，openai-node 会按非流式响应处理并失败）。pi 的默认字段（`stream: true`、`stream_options.include_usage: true`、`store: false`、`max_completion_tokens`、推理模型的 `developer` 角色等）由 DTO 显式写出，各 fixture 断言服务端实际收到的请求体，差分与 pi 逐字段比较。
- 传输：`option.WithHTTPClient` 使用 Client 共享的 `*http.Client`（受控 RoundTripper：时限、loopback 限制），SDK 自身重试关闭（`WithMaxRetries(0)`），认证只在 adapter 给出的头里，不经 `openai.NewClient`（它读取 `OPENAI_*` 环境变量）。
- 流：`NewStreaming` 返回的类型化 `ssestream.Stream[ChatCompletionChunk]` 与 openai-node（pi 所用）的读取方式在可观察行为上不同：遇到 `[DONE]` 立即停止并关闭响应体，而 openai-node 继续读到正文结束；任何带 `error` 字段的 chunk（包括 `"error": null`）都按错误结束，openai-node 只在该字段为真值时失败；非对象 chunk（数字、字符串、数组）解码失败，pi 跳过它们；类型化 chunk 不承载 pi 读取的 `reasoning_content`/`reasoning`/`reasoning_text`/`reasoning_details`（只能经 `ExtraFields` 取原文），也不区分工具调用缺失的 `index` 与 0。因此 adapter 只用 SDK 公开的 `ssestream.NewDecoder` 切分 SSE 事件，事件内容按 openai-node 的 `Stream.fromSSEResponse` 与 pi 的 chunk 循环读取。

### Considered Options

- SDK 类型化流：上述行为差异无法在不读取原文的前提下消除。
- 直接 HTTP（ADR-0012 的做法）：SDK 的发送、错误体读取与 SSE 切分都可用，没有必要。
- SDK 发送与切分、adapter 读取 chunk（采用）：与 ADR-0011（Anthropic）的分层一致。

### Consequences

- SDK 的 SSE 切分按 LF 结束行（CR 只在 LF 前被去掉），单行超过 32 MiB 时它自己报错（按连接中断报告）；只有 `event:` 没有 `data:` 的事件被它跳过，openai-node 会以 JSON 解析错误失败——真实服务不发这种事件，未纳入差分。
- 读取中途取消时 openai-node 静默结束流，pi 随后关闭所有块再报 `Request was aborted`；barness 同样先发块的 end 事件再以 aborted 结束。连接中断、解码错误与流内错误则不关闭块。
- JSON 解析失败与连接中断的文本是 barness 固定文本（pi 报 V8 与 undici 的运行时文本），登记于差分账本。

## 决策二：内置目录列出 pi OpenAI 数据中 Chat Completions 能服务的模型

pi 的 `openai.json` 只把 OpenAI 模型登记在 `openai-responses` 上；pi 用户经 Chat Completions 调用 OpenAI 时，是把同一条模型数据改成 `openai-completions` 的自定义模型。barness 的 `builtinOpenAIChatModels` 照此列出已列入目录的 OpenAI 模型（字段与 compat 与 Responses 条目一致），只去掉只在 Responses 上提供的 gpt-5-pro 与 gpt-5.5-pro，目录升为 `2026-10-02.4`。差分运行器同样取 pi 的 OpenAI 数据并把 API 换成 `openai-completions`。

## 决策三：只实现标准 OpenAI compat

pi 的 `detectCompat` 按 provider 名与 baseUrl 子串识别十余家兼容服务（DeepSeek、Z.ai、OpenRouter、Together……），切换 `max_tokens` 字段、`developer` 角色、思考格式、session 亲和头、缓存控制等。首期只有 OpenAI 的 Chat 模型在目录中，barness 实现 OpenAI 被识别出的标准 compat（`store: false`、`developer` 角色、`reasoning_effort`、`max_completion_tokens`、流末 usage、必须有 finish_reason），模型 compat 中 barness 已承载的 `supportsStrictMode`、`supportsLongCacheRetention`、`supportsMidConvoSystemMessages` 照常生效；按厂商探测的其余分支不移植，DeepSeek × Chat（P06）在工单 22 中按其 compat 增加。binding 的 endpoint 是经验收的厂商地址（spec I3），不会把 OpenAI binding 指向别家服务。

`prompt_cache_key` 的条件按 pi 原样移植：endpoint 含 `api.openai.com` 且 retention 不是 none，或 retention 为 long 且模型支持长保留时才发送；发送的是按租户与账户派生的键（spec §9，与 Responses 相同的扩展）。OpenAI 的 Chat 路径不发送 session 亲和头（pi 只对 OpenRouter 发送）。pi 的 Chat 路径不发送 metadata，`thinkingBudgets` 只用于 OpenAI 没有的预算字段，因而 `ChatOptions` 没有这两项；simple 入口的同名字段同样不产生请求字段。

## 决策四：终态与流末 usage

finish_reason 是协议终态：`stop`/`end` 为 stop，`length` 为 length，`tool_calls`/`function_call` 为 toolUse，其他（`content_filter`、`network_error`、未知值）在流结束后以 `Provider finish_reason: <reason>` 失败，原样记入 rawStopReason；整个流没有 finish_reason 时以 `Stream ended without finish_reason` 失败（CodeProtocol）。`data: [DONE]` 只是分帧结束标记：有 finish_reason 而没有 `[DONE]` 的正文照常成功，`[DONE]` 之后的事件读到正文结束但不再处理。`include_usage` 的 usage chunk 在 finish chunk 之后、`[DONE]` 之前到达，任何带 usage 的 chunk 都整体替换用量（choice 内的 usage 是 Moonshot 的位置，同样读取），所以结束前的 usage 不会遗漏。

## 决策五：用量完整性

用量换算按 pi 的 `parseChunkUsage`：缓存读取依次取 `prompt_tokens_details.cached_tokens`、DeepSeek 的 `prompt_cache_hit_tokens`、Kimi 的 `cached_tokens`；OpenRouter 的 `cache_write_tokens` 单独计为缓存写入；input 扣除两者；reasoning 是 output 的子集；totalTokens 是各部分之和而不是厂商的 `total_tokens`。完整性按最近一次 usage 判定：`prompt_tokens`、`completion_tokens`、任一缓存读取字段与 `completion_tokens_details.reasoning_tokens` 都报告（0 也算）为 `complete`，缺任一为 `partial`，从未收到为 `unreported`；`total_tokens` 不参与换算，`cache_write_tokens` 只有 OpenRouter 报告，二者缺失不影响 `complete`。

## 决策六：回调授权的 Chat 专有字段

onPayload 看到的就是 adapter 构建并实际发送的请求体；onResponse 与 Responses 相同，只在取得初始响应后、start 前执行一次。授权（ADR-0005 的 Chat 版）：模型必须是授权模型，`prompt_cache_key` 只能是派生值或不出现，`store` 只能为 false 或不出现；`tools` 只能声明 `function`，其他类型须在 binding 的 `AllowedHostedTools` 中；托管网页搜索在 Chat 上以字段 `web_search_options` 开启，须由 `AllowedHostedTools` 列出 `web_search_options`；消息中任何位置的 `file_id`（Files API 引用）与 assistant 消息带 `id` 的 `audio`（厂商保存的音频回复）一律拒绝。samplingParams 不得设置 `model`、`messages`、`stream`、`tools`、`functions`、`prompt_cache_key`、`store`、`web_search_options`。

## 决策七：错误文本、自定义工具调用

pi 的 Chat 路径报告错误时不加 "API error" 前缀：SDK 消息没有引用的非空错误对象显示为 `<status>: <错误对象>`（截到 4000 个 UTF-16 单位），否则为 SDK 消息（`<status> <message>` 或 `<status> status code (no body)`）；错误对象 `metadata.raw`（OpenRouter 的上游错误）是字符串且消息中没有时另起一行附加。流内错误 chunk 报其 message。key 形态文本照常脱敏。重试与超时与 Responses 相同（openai-node 的 Stainless 规则、`X-Stainless-Timeout` 只在显式 timeoutMs 时发送）。

自定义（grammar）工具调用只由 barness 无法声明的工具产生，与 Responses 一样被跳过；pi 会为它建块。pi 在出现两个 finish_reason 且先错后成功时把旧的 errorMessage 留在成功消息上，barness 只在失败终态写 errorMessage——真实服务每个 choice 只发一个 finish_reason，未纳入差分。畸形 chunk 中只读取 JSON 类型合 pi 用法的值：非字符串的 chunk id、工具调用 id 与参数被忽略（pi 的 `||=` 与字符串拼接会接受任何真值），`function` 不是对象的工具调用条目被跳过（pi 仍会为它建块）；同样不纳入差分。

## 维护者决定（2026-10-02）

1. 决策一：采纳。SDK 承担 Chat 的请求发送与 SSE 事件切分，chunk 由 adapter 按 openai-node 读取；SDK 强制 `stream: true`。spec §5 已同步。
2. 决策二：采纳。Chat 内置模型取 pi 的 OpenAI 数据换为 `openai-completions`，不列入 gpt-5-pro 与 gpt-5.5-pro。
3. 决策三：采纳。只实现标准 OpenAI compat，其余厂商探测留待各自工单（DeepSeek 见工单 22）。
4. 决策五：采纳。`prompt_tokens`、`completion_tokens`、任一缓存读取字段与 `reasoning_tokens` 定义完整性。
5. 决策六：采纳。`web_search_options` 以字段名作为托管工具类型，须由 binding 的 `AllowedHostedTools` 放行；`file_id` 与已存音频引用一律拒绝。

差分中只在 Chat 出现的差异（运行时错误文本、未担保原生状态降级、派生缓存键）登记在 `ai/e2e/testdata/pidiff/ledger.json`。
