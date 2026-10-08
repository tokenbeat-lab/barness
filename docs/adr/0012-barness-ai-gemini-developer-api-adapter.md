---
status: accepted
date: 2026-10-02
---

# barness-ai Gemini Developer API adapter：直接 HTTP、按 @google/genai 解码，回调看到 REST 请求体

工单 19 接入 Google × Gemini Developer API（P03）。spec I5 默认由 Google GenAI Go SDK（研究锁定 v1.71.0）承担 HTTP 与流解码，但要求先核查必需字段能否经公开扩展或原始响应无损保留，否则该组合改用直接 HTTP；spec I7 要求保留 onPayload、不调用 onResponse，且拒绝非默认 fetch 的基线行为不被悄悄改写。spec 没有规定回调看到的请求体形状、Google 路径的重试与超时、Gemini 的"用量完整"含义，以及哪些 pi 模型可以列入目录。本 ADR 记录实现时定下的做法，维护者于 2026-10-02 确认（见文末）。

## 决策一：不用 Go SDK，直接 HTTP

对 `google.golang.org/genai` v1.71.0 的核查结论（依据其源码）：

- `Part.ThoughtSignature` 是 `[]byte`，JSON 解码时按标准 base64 解码：pi 原样保留的签名字符串在这里会被规范化，不是合法 base64 的签名让整个 chunk 解码失败（pi 保留它，只在回放时按 `isValidThoughtSignature` 不发送）。
- `FunctionCall.Args` 是 `map[string]any`：参数对象的键顺序丢失。pi 的 `toolcall_delta` 是 `JSON.stringify(args)`，键序随线上顺序，必需字段无法无损保留。
- 流结果只有类型化的 `GenerateContentResponse`，没有原始 JSON 的公开出口；`HTTPOptions.ExtrasRequestProvider` 只作用于请求体。
- 流迭代器把任何非 `data:` 行（注释心跳、具名事件）当错误，而 pi 使用的 @google/genai 跳过它们。
- `NewClient` 读取 `GOOGLE_API_KEY`、`GEMINI_API_KEY`、`GOOGLE_GEMINI_BASE_URL`、`GOOGLE_GENAI_USE_VERTEXAI`、`GOOGLE_CLOUD_PROJECT` 等环境变量，须逐项覆盖才能保证 spec I3 的"不从环境兜底"。

因此按 spec I5 的例外条款，仅 Google × Gemini Developer API 改为直接 HTTP：adapter 自己编码请求体，经 Client 共享的 `*http.Client`（同一 transport、时限与 body 限额）发送 `POST {endpoint}/models/{model}:streamGenerateContent?alt=sse`，认证只放在 `X-Goog-Api-Key` 头，从不放进 URL。binding 的 endpoint 与 pi 的 `model.baseUrl` 一样带版本路径（`…/v1beta`）。同一组场景（P03 × E01–E09、E11）全部通过，差分 0 待处理。

### Considered Options

- Go SDK 类型化流：签名与参数无法无损保留（见上）。
- Go SDK 只发请求、adapter 自行解码：SDK 没有公开的"发原始请求、交回原始响应"入口，流式方法总是自行解析；绕开它只剩 net/http，却仍要承担它的环境读取。
- 直接 HTTP（采用）。

### Consequences

- 不发送 pi 的 Google SDK 运行时指纹 `x-goog-api-client`，也不发送 undici 默认的 `Accept: */*`（差分账本登记）。
- 被头变换删除的头保持"存在但无值"，net/http 既不写出它，也不补上自己的 `Go-http-client/1.1`。
- spec §5 写的是 Google GenAI SDK 承担 Gemini 的 HTTP 与流解码；这里按其例外条款整体改为直接 HTTP，已经维护者确认并同步到 spec §5（与 ADR-0011 对 Anthropic 的处理相同）。

## 决策二：流按 @google/genai 2.21.0 的 processStreamResponse 解码

pi 通过 @google/genai 读流，所以 adapter 移植它的解码：按缓冲区里最早出现的 `\n\n`、`\r\r`、`\r\n\r\n` 切事件，事件去空白后以 `data:` 开头才是 chunk，其余（注释、具名事件、id 行）跳过；多行 `data:` 不拼接，整段按一个 JSON 解析，因而失败；正文结束时剩余非空白内容按 SDK 原文报 `Incomplete JSON segment at the end`。终态按 pi 判定：只有 `finishReason` 让调用成功；没有就报 `Google stream ended without a finish reason`；错误类结束原因在流结束后报 `Provider stopped with: <reason>`；SDK 枚举之外的结束原因立即报 `Unhandled stop reason: <reason>`。

SDK 还会在"一个网络 chunk 整体是带 HTTP 错误码 `error` 的 JSON 对象"时以 `got status: <status>. <JSON>` 失败。网络分块在这里不可观察，adapter 对每个事件（以及结尾的剩余内容）做同样检查：provider 把错误单独写成一块时（真实服务与本地受控 Provider 都如此）两者一致；错误与其他事件合并在一块时 pi 不会识别，barness 会。

### Consequences

- `streamBody` 只在 LF 处结束 frame，以 `\r\r` 分隔的事件会与后一个事件计入同一 frame：frame 限额只会提前触发，不会漏判。
- JSON 解析失败的文本是 barness 固定文本（pi 报 V8 的 JSON.parse 消息），连接中断报 `Connection lost while reading the Gemini stream`（pi 报 undici 的 `terminated`），流中取消一律报 pi 自己的 `Request was aborted`；均登记于差分账本。

## 决策三：重试与超时按 pi 的 Google 路径

pi 用 `retryGoogleRequest` 包装请求：只有 SDK 的 `ApiError`（带 HTTP 状态）可重试，按状态 408/409/429/5xx 判定；该错误没有响应头（`error.headers` 被设为 undefined），所以 `x-should-retry` 与 `retry-after(-ms)` 都不起作用，退避一律指数退避；没有拿到响应的请求（fetch 失败）不重试。barness 的 Google adapter 照此执行：初始请求失败时不向重试规则交出响应头（`Error.RetryAfter` 仍作为错误元数据报告），没有响应的尝试（含策略的建连/响应头时限到期）不重试。这与 Responses、Anthropic 的重试规则不同，后者按 Stainless SDK 重试连接失败与超时。

pi 创建 Google client 时不设请求超时，所以 Gemini 没有协议 timeoutMs：`GeminiOptions` 没有该字段，simple 入口的 `TimeoutMs` 不起作用；资源策略的各项时限照常生效（spec I9）。

## 决策四：onPayload 看到 REST 请求体；不调用 onResponse

pi 的 onPayload 看到的是 @google/genai 的 `generateContentStream` 参数（`{model, contents, config}`），SDK 随后把它转换为 REST 请求体并丢弃它不认识的字段。barness 没有这层 SDK 参数，回调看到的是 adapter 构建的 REST 请求体（`contents`、`systemInstruction`、`tools`、`toolConfig`、`generationConfig`），即 spec I7 所说"adapter 构建原生请求体后"的那份，也是授权检查实际检查、实际发送的那份。为 pi 写的 Google onPayload 回调移植时需改用 REST 字段名（如 `generationConfig.thinkingConfig` 而非 `config.thinkingConfig`）。

授权（ADR-0005 的 Gemini 版）：模型在 adapter 构造的 URL 里，请求体只能不写 `model` 或写授权模型本身；工具条目以字段名表示种类，`functionDeclarations` 总是允许，其他种类（如 `googleSearch`、`codeExecution`、`urlContext`）须在 binding 的 `AllowedHostedTools` 中；`cachedContent`、`mcpServers`、`fileSearch` 以及 system instruction 与 contents 中任何位置的 `fileData`（Files API 引用）一律拒绝——它们指向账户上的缓存、带自身地址与凭据的 MCP 服务器、文件检索库或文件，库无法替多租户共享的账户担保。

pi 的 Google 路径从不调用 onResponse，barness 同样不调用（设置了也不运行，返回的错误也不影响调用）。pi 拒绝非默认 fetch；barness 没有每调用的 transport/fetch 入口（transport 只在 Client 构造时装配，spec I7），没有可拒绝的对象，这一基线不变。将来若增加每调用 fetch 支持，须作为扩展登记。

## 决策五：错误文本

非 2xx 响应的消息是 SDK `ApiError` 的消息，pi 原样报告：Content-Type 为 JSON 时是错误体的 `JSON.stringify`（barness 以紧凑化近似，转义与数字写法保持原样），否则是 `{"error":{"message":<正文>,"code":<状态>,"status":<状态文本>}}`。声明为 JSON 却无法解析的错误体，pi 报 JavaScript 的解析错误且不重试；barness 按非 JSON 正文描述、按状态重试——真实服务不会发送这种响应，未纳入差分。没有响应时报 undici 的 `fetch failed`（pi 原样）。Gemini 响应不带厂商 request id 头，`ProviderRequestID` 为空。key 形态脱敏扩展到 Google API key（`AIza` 加 35 个字符），适用于所有协议的错误文本。

## 决策六：用量完整性

pi 每收到一个 `usageMetadata` 就整体替换用量（缺失或 null 计为 0）：input = prompt − cached，output = candidates + thoughts，cacheRead = cached，reasoning = thoughts，totalTokens 取 Gemini 的总数；Gemini 不报告缓存写入。barness 同样在每次替换后写入消息 Usage 与当前尝试（ADR-0010）。完整性按最近一次 `usageMetadata` 判定：`promptTokenCount`、`candidatesTokenCount`、`totalTokenCount` 都报告（0 也算）为 `complete`，缺任一为 `partial`，从未收到为 `unreported`；`cachedContentTokenCount` 与 `thoughtsTokenCount` 只在有值时出现，缺失读作没有，不影响 `complete`。

## 决策七：thoughtSignature、生成的工具调用 ID

`ToolCall.ThoughtSignature` 用 `Nullable[string]` 区分缺失、null、空串与值，随消息持久化。跨模型回放按 pi 的 truthy 条件只删除非空签名，null 与空串原样保留（工单 08 移交的规则，实现在共享的 `replayContent`）；非空签名计入原生状态，只在可信封套下同模型回放。文本块的签名放在 `Text.Signature`、思考块的放在 `Thinking.Signature`，与 pi 的 `textSignature`/`thinkingSignature` 对应；`thought: true` 才是思考，签名本身不是。

Gemini 2 不发送调用 id。pi 为缺 id 或与已有 id 重复的调用生成 `<name>_<Date.now()>_<进程内计数>`；barness 生成同样形态，时间取 Client 时钟（clock 包的控制点扩展到此，测试可固定），计数在每次调用内从 1 开始。同一会话的调用依次进行，毫秒时间与计数一起保证会话内唯一。差分把这种形态的工具调用 id 与时间戳一样一一映射。

## 决策八：内置目录只列 generateContent 能完整服务的 Google 模型

列入 pi 数据中的 gemini-2.5-flash、gemini-2.5-flash-lite、gemini-2.5-pro、gemini-3-flash-preview、gemini-3.1-flash-lite、gemini-3.1-flash-lite-preview、gemini-3.1-pro-preview、gemini-3.1-pro-preview-customtools、gemini-3.5-flash、gemini-3.5-flash-lite、gemini-3.6-flash、gemini-3.7-flash、gemini-3.8-flash、gemini-flash-latest、gemini-flash-lite-latest、gemma-4-26b-a4b-it、gemma-4-31b-it（逐字段与 pi 数据核对一致），目录升为 `2026-10-02.3`。不列入：deep-research-preview-04-2026 与 deep-research-max-preview-04-2026（Deep Research 代理经 Interactions API 运行）、gemini-2.5-computer-use-preview-10-2025（需要托管的 computer use 工具）、gemini-3.1-flash-live-preview（Live API 模型）、gemini-3.1-flash-lite-image（输出图片，adapter 只读文本、思考与函数调用部分）。

约束采样工具（pi 工具的 `constrainedSampling`，Gemini 3 上发送 `VALIDATED` 调用模式与严格化 schema）在 barness 的 `Tool` 中无法声明，与 Responses、Anthropic 路径相同；Gemini 请求因此从不发送 `VALIDATED`，工具 schema 原样发送。

## 决策九：选项形状

`GeminiOptions` 对应 pi 的 `GoogleOptions`，去掉传输、凭据、头、回调、重试与 fetch：`Temperature`、`MaxTokens`（null 与未设置一样不发送，SDK 丢弃 null 配置字段；0 照发）、`ToolChoice`（auto/none/any，仅在声明了工具时发送）、`Thinking`（`Enabled`、`BudgetTokens`、`Level`）。`Level` 一旦出现（null 也算）就走等级分支，null 时既不发等级也不发预算；`BudgetTokens` 为 null 时按 pi 原样发送 `null`。pi 会静默丢弃不认识的等级，barness 拒绝（invalid_request）；预算允许 -1（动态）。pi 的 Google 路径忽略的 cacheRetention、sessionId、metadata、timeoutMs、samplingParams 没有字段，simple 入口的同名字段同样不产生请求字段。simple 入口遇到模型等级映射到非 Gemini 等级时，pi 在 `streamSimple` 中同步抛错，barness 以 invalid_request（能力阶段）结束。

## 维护者决定（2026-10-02）

1. 决策一：采纳。Google × Gemini Developer API 整体直接 HTTP，spec §5 已同步。
2. 决策三：采纳。Google 路径只按 HTTP 状态重试，不看 `x-should-retry` 与 `retry-after`，不重试无响应的尝试（含策略时限到期）；没有协议 timeoutMs。
3. 决策四：采纳。onPayload 看到 REST 请求体；拒绝 `cachedContent`、`mcpServers`、`fileSearch` 与 `fileData`，其他托管工具种类须由 binding 的 `AllowedHostedTools` 放行。spec §7 已同步。
4. 决策六：按现有实现（`promptTokenCount`、`candidatesTokenCount`、`totalTokenCount` 定义完整性）。
5. 决策八：不列入的五个模型维持不列入。

差分中只在 Gemini 出现的差异（`x-goog-api-client` 与 `Accept` 头、运行时错误文本、未担保原生状态降级）登记在 `ai/e2e/testdata/pidiff/ledger.json`。

## 工单 12：图像操作并行接入（2026-10-08）

Google × google-interactions × image 使用独立同步 unary 协议，与本 ADR 的
generateContent 聊天操作并存。聊天选项、目录、重试语义及不调用 OnResponse 的决定保持不变。
新路线固定 v1beta、无状态与内联图片，使用共享 Client HTTP；图像型号继续排除出聊天目录。
详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。

## 工单 14：Google 图像真实证据（2026-10-08）

Google 图像首批型号已由工单 14 自己的生成/参考图编辑真实 PASS 纳入 image 目录；不进入聊天目录，generateContent 路线不变。显式 delivery 拒绝及授权修订见 ADR-0023。
