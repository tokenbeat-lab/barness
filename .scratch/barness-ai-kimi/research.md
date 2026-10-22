# Kimi 三种兼容协议与 barness-ai 接入建议

研究日期：2026-10-22（Asia/Shanghai）。仓库基线：`68b7da0a3400f909dc535d7f585c93771df1569c`。范围：Kimi API 开放平台中国站、国际站的 **OpenAI Responses、OpenAI Chat Completions、Anthropic Messages**。不研究 GLM、Kimi 自有协议、托管智能体或 Kimi Code 服务接入；兼容接口中的厂商字段只在影响互操作时说明。

## 1. 结论与建议

**两站均已明确提供三种兼容协议。建议仿照 DeepSeek 的接入结构：复用三套现有协议 adapter，新增精确的 Kimi Provider profile；按地区建立两个独立 Provider，首批以 `kimi-k3` 完成六条独立路线。** 官方概述明确列出三种协议及各自地址，不能仅凭“兼容 OpenAI”推导能力，也不能继续把 Responses 当成未支持。这里的 Completions 指 Chat `/v1/chat/completions`；本次未在官方接口目录找到 legacy `/v1/completions` 的支持声明。[中国 API 概述](https://platform.kimi.com/docs/api/overview)、[国际 API 概述](https://platform.kimi.ai/docs/api/overview)

最影响实现的差异有四项：

1. K3 的 Chat、Responses、Messages 使用三种不同的 effort 字段；工具选择范围也不同，不能共享一份厂商参数配置。
2. K3 多轮调用需要保留推理与工具历史；Responses 提供明文 summary，不能请求 OpenAI 的 encrypted reasoning；Messages thinking signature 需按实际返回保留。
3. 自动缓存默认开启，而且缓存写入已经单独计费。Chat/Responses 的总输入包含读、写；Messages 的 `input_tokens` 排除了这两者。现有读缓存兼容不足以保证新计费正确。
4. 两站账户、余额、Key 完全独立，支持证据、价格和原生状态必须独立，错误不能触发跨站 fallback。

以上分别来自[模型参数](https://platform.kimi.com/docs/api/models-overview)、[推理强度](https://platform.kimi.com/docs/guide/use-reasoning-effort)、[Responses](https://platform.kimi.com/docs/api/responses)、[Messages](https://platform.kimi.com/docs/api/messages)、[缓存指南](https://platform.kimi.com/docs/guide/context-caching)、[错误码的平台隔离说明](https://platform.kimi.com/docs/api/errors)。具体契约和代码落点见下文。

本研究只读取公开文档与仓库代码，**没有读取凭据、没有调用模型、没有产生真实支持资格**。下文把“官方文档事实”“仓库事实”和“建议/待核实”分开。

## 2. 地区、地址与认证

### 2.1 六条生成路线

| 站点 | API Protocol | SDK base URL | 最终生成 endpoint |
| --- | --- | --- | --- |
| 中国 | `openai-completions` | `https://api.moonshot.cn/v1` | `POST /v1/chat/completions` |
| 中国 | `openai-responses` | `https://api.moonshot.cn/v1` | `POST /v1/responses` |
| 中国 | `anthropic-messages` | `https://api.moonshot.cn/anthropic` | `POST /anthropic/v1/messages` |
| 国际 | `openai-completions` | `https://api.moonshot.ai/v1` | `POST /v1/chat/completions` |
| 国际 | `openai-responses` | `https://api.moonshot.ai/v1` | `POST /v1/responses` |
| 国际 | `anthropic-messages` | `https://api.moonshot.ai/anthropic` | `POST /anthropic/v1/messages` |

地址直接来自两站[中国概述](https://platform.kimi.com/docs/api/overview)、[国际概述](https://platform.kimi.ai/docs/api/overview)。控制台/新文档域名是 `platform.kimi.com` 和 `platform.kimi.ai`，API host 仍然是 `api.moonshot.cn` 和 `api.moonshot.ai`；不要将控制台域名替换进生成 URL。旧国际文档 `platform.moonshot.ai` 在本次直接读取中返回 403，新国际文档可以正常读取；这只记录文档访问结果，不推断 API 可用性。

### 2.2 认证与地区隔离

三种 endpoint 的参考页均声明 `Authorization: Bearer <MOONSHOT_API_KEY>`，JSON 请求使用 `Content-Type: application/json`。Messages 的 OpenAPI 也使用 `bearerAuth`，其 curl 示例没有 `x-api-key` 或 beta query。[中国 Chat](https://platform.kimi.com/docs/api/chat)、[中国 Responses](https://platform.kimi.com/docs/api/responses)、[中国 Messages](https://platform.kimi.com/docs/api/messages)、[国际 Messages](https://platform.kimi.ai/docs/api/messages)

官方明确：中国站和国际站的**账户、余额和 API Key 完全独立**；混用会导致 401。Kimi 开放平台、Kimi Code、Kimi 会员又是独立产品，Key 与权益也不能混用。因此不是“同一个 Provider 换域名”，也不应以某站成功推导另一站可用。[中国错误码](https://platform.kimi.com/docs/api/errors)、[官方国际排障](https://www.kimi.ai/help/kimi-api/api-troubleshooting)

**认证文档歧义**：两站概述同时展示 `Anthropic(api_key=...)` 的通用示例，而 endpoint OpenAPI 与 Claude Code 教程使用 Bearer/`ANTHROPIC_AUTH_TOKEN`。不能据此断言 `x-api-key` 必然拒绝或两种头等价。建议直接采用明确的 Bearer 契约，只发送一种认证头，真实接入时核实最终 wire。[中国概述](https://platform.kimi.com/docs/api/overview)、[国际概述](https://platform.kimi.ai/docs/api/overview)、[Claude Code 教程](https://platform.kimi.ai/docs/guide/claude-code-kimi)

## 3. 当前模型与协议覆盖

两站模型目录一致列出以下当前型号；这证明目录存在，不证明指定账户有调用权限。`kimi-k2.5`、`moonshot-v1` 系列已于 2026-08-31 下线，`kimi-k2` 系列已于 2026-05-25 下线，不建议作为新接入基线。[中国模型列表](https://platform.kimi.com/docs/models)、[国际模型列表](https://platform.kimi.ai/docs/models)

| 型号 | 上下文文档值 | Chat 契约 | Responses 契约 | Messages 契约 |
| --- | --- | --- | --- | --- |
| `kimi-k3` | 1,048,576 | 明确支持 | 参考页明确支持 | 参考页 model enum 明确支持 |
| `kimi-k2.7-code` | 262,144 | 明确支持 | 当前参考页未列出 | Claude Code 教程展示，但 Messages model enum 未列入 |
| `kimi-k2.7-code-highspeed` | 262,144 | 明确支持 | 当前参考页未列出 | 当前参考页未列出 |
| `kimi-k2.6` | 262,144 | 明确支持 | 当前参考页未列出 | Claude Code 教程描述其行为，但 Messages model enum 未列入 |

协议覆盖依据：[Chat 的模型 discriminator](https://platform.kimi.com/docs/api/chat)、[Responses model 字段](https://platform.kimi.com/docs/api/responses)、[Messages model 字段](https://platform.kimi.com/docs/api/messages)、[国际 Claude Code 教程](https://platform.kimi.ai/docs/guide/claude-code-kimi)；精确上下文值来自两站[中国价格表](https://platform.kimi.com/docs/pricing/chat)、[国际价格表](https://platform.kimi.ai/docs/pricing/chat)内嵌表格。Messages 的文档差异属于待对齐契约，不应武断写成所有 K2.x 不支持，也不应直接扩张正式目录。

建议首批每站每协议只放 `kimi-k3` 候选，用原始 ID。Claude Code 的 `kimi-k3[1m]` 是该工具教程使用的兼容名称，不作为 barness-ai 的通用型号 ID。K2.x 留到其各自协议文档与真实行为明确后单独扩展。[官方排障的名称说明](https://www.kimi.ai/help/kimi-api/api-troubleshooting)

## 4. OpenAI Chat Completions

### 4.1 请求、推理与参数

兼容的基本形状是 `messages`、`tools[].function`、assistant `tool_calls`、`role:"tool"` 的结果及 `tool_call_id` 配对。标准消息使用 `system/user/assistant/tool`；应按该 schema 发送系统指令，不能自动沿用 OpenAI 的 `developer` 角色。新接入使用 `max_completion_tokens`，`max_tokens` 已被标为弃用。[Chat API](https://platform.kimi.com/docs/api/chat)

| 模型 | 推理控制 | 重要限制 |
| --- | --- | --- |
| K3 | 顶层 `reasoning_effort: low/high/max`；默认 max；始终进行推理并保留历史推理 | 移除 K2.x 的 `thinking`，不承诺 off/none 或 medium/xhigh |
| K2.7 Code / highspeed | 可省略 `thinking`；显式时只接受 `{"type":"enabled","keep":"all"}` | 不能关闭思考；不支持 `reasoning_effort` |
| K2.6 | `thinking.type=enabled/disabled`，默认 enabled；`keep:"all"` 可保留历史推理 | 不支持 `reasoning_effort` |

来源：[模型参数参考](https://platform.kimi.com/docs/api/models-overview)、[推理强度指南](https://platform.kimi.com/docs/guide/use-reasoning-effort)。这些是模型差异，应由精确型号规则描述；不能把所有 Kimi 模型塞进 DeepSeek 的 thinking toggle + effort 编码。

K3/Code 的 `temperature=1`、`top_p=.95`、`n=1`、presence/frequency penalty=0 固定；K2.6 非思考模式 temperature=.6，思考模式=1。其他值会报错，官方建议省略这些固定参数。K3 输出上限默认 131072，可设至 1048576；输入与所请求输出超过上下文剩余空间时，服务端会调低输出预算。这不是保证每次能生成请求上限。[模型参数参考](https://platform.kimi.com/docs/api/models-overview)、[Chat 的输出预算字段](https://platform.kimi.com/docs/api/chat)

### 4.2 工具、历史与流

K3 支持 `tool_choice=auto/none/required`；K2.6/Code 不支持 required。虽然 schema 也列具名函数对象，专门指南明确它与开启 thinking 不兼容并返回 400；K3 始终思考，不能把“schema 有该对象”当成 K3 具名强制调用可用。[工具选择](https://platform.kimi.com/docs/guide/use-tool-choice)、[模型参数](https://platform.kimi.com/docs/api/models-overview)

多轮和工具往返应回传完整 assistant message，包括 `reasoning_content` 与 `tool_calls`，随后追加相应 tool result；K3/Code 的 Preserved Thinking 始终开启。K2.6 的历史保留开关与同一工具生成轮次内的推理回传要求应区分，不应丢弃厂商返回的推理。[思考模型](https://platform.kimi.com/docs/guide/use-thinking-models)、[推理强度](https://platform.kimi.com/docs/guide/use-reasoning-effort)、[工具调用指南](https://platform.kimi.com/docs/guide/use-kimi-api-to-complete-tool-calls)

流式使用 SSE 的 `chat.completion.chunk`，分别处理 `delta.reasoning_content`、`delta.content` 和 `delta.tool_calls`。应请求 `stream_options.include_usage=true`，最终 usage 在 `[DONE]` 前；官方示例出现带 finish_reason 的 chunk 上 usage，schema 又描述独立 `choices:[]` usage chunk，接收器应兼容两种位置，不能见 finish_reason 就退出。[Chat API](https://platform.kimi.com/docs/api/chat)、[流式输出指南](https://platform.kimi.com/docs/guide/utilize-the-streaming-output-feature-of-kimi-api)

JSON Mode 与 Structured Output 都有文档，后者 schema 受 MFJS 约束；严格输出不代表支持 JSON Schema 的任意关键字。初批如果只交付现有文本/函数工具合同，可以独立登记结构化输出为后续能力，不新增厂商私有 API。[输出格式指南](https://platform.kimi.com/docs/guide/response_format)

## 5. OpenAI Responses

### 5.1 明确支持的兼容子集

两站都是实际 `POST /v1/responses`，当前参考页明确支持 `kimi-k3`。输入可以是字符串或有序 item 数组，顶层 `instructions` 提供系统指令；message 的 `developer` 按系统指令处理。使用 `max_output_tokens`，K3 默认 131072、最大 1048576；推理强度使用 **`reasoning.effort=low/high/max`**，默认 max。[中国 Responses](https://platform.kimi.com/docs/api/responses)、[国际 Responses](https://platform.kimi.ai/docs/api/responses)

函数工具使用平铺 `type/name/parameters/strict`，调用输出是 `function_call`，结果用 `function_call_output` 并配对 `call_id`。官方要求下一轮把先前 `output` 原样追加到完整 `input`。schema 另列 namespace、仅名为 `apply_patch` 且使用 `grammar`/`lark` 格式的 custom、服务端 web_search；首批复用 barness 已有 function tool 语义即可，其余需要自身授权和解析能力。[Responses 工具说明](https://platform.kimi.com/docs/api/responses)

**`tool_choice` 的公开枚举只有 `auto`。** Chat 的 required/none、Messages 的 any 不能映射进来承诺支持；显式超出当前契约的请求建议发送前拒绝。[国际 Responses 的 tool_choice](https://platform.kimi.ai/docs/api/responses)

### 5.2 状态、推理和不应推导的字段

输出 reasoning 是 `summary:[{type:"summary_text",text:...}]`；`encrypted_content` 固定 null。输入 reasoning 支持 summary，亦支持 `content:[{type:"reasoning_text",text:...}]` 且 content 优先。使用真实返回的 summary 形成完整历史，不请求 `include:["reasoning.encrypted_content"]`。[Responses schema](https://platform.kimi.com/docs/api/responses)

响应中 `store/background=false`、`previous_response_id/conversation=null`。这些字段没有出现在请求 schema，因此采用完整历史回放；不能据此声称传入它们一定被忽略、接受或拒绝。`service_tier`、`reasoning.summary`、`prompt_cache_retention` 等也没有可用的请求契约。响应出现某字段不是该字段可写的证明。请求 `include` 只明确列 web_search sources/results。[中国 Responses](https://platform.kimi.com/docs/api/responses)、[国际 Responses](https://platform.kimi.ai/docs/api/responses)

SSE 包含 response created/in_progress、item/content part、output_text delta/done、**reasoning_summary_text** delta/done、function_call_arguments delta/done，以及 completed/incomplete/failed/error。应保留 terminal usage；incomplete 可由 max_output_tokens 或 content_filter 导致，不能统一当作成功或仅凭 HTTP 200 判定完成。[Responses SSE 与终态 schema](https://platform.kimi.com/docs/api/responses)

web_search 的部分约束确有明确处理方式：search_context_size/blocked_domains 会报 invalid_request_error，user_location/external_web_access/indexed_web_access 被忽略。该结论只适用于这些字段，不能扩张为“未知 OpenAI 字段全被忽略”；首批建议不启用托管搜索。[Responses web_search 工具 schema](https://platform.kimi.com/docs/api/responses)

## 6. Anthropic Messages

### 6.1 基本形状与专有差异

Messages 使用顶层 `system`、必填 `max_tokens`、user/assistant `messages`；工具定义是 `input_schema`，模型生成 `tool_use`，Host 在 user 消息中提供配对 `tool_result.tool_use_id`。工具选择公开为 `{type:auto|any|none}`，未列具名 tool 选择。[中国 Messages](https://platform.kimi.com/docs/api/messages)、[国际 Messages](https://platform.kimi.ai/docs/api/messages)

当前参考页用 **`output_config.effort=low/high/max`**，默认 max；没有声明 Anthropic 的 thinking budget、adaptive、display、interleaved beta、eager streaming 或浏览器访问字段。K3 的首批接入应单独编码 effort，显式超出文档能力的选项应拒绝，不能为了生成 output_config 而附带 `thinking:{type:adaptive}`。[Messages request schema](https://platform.kimi.com/docs/api/messages)

响应内容顺序是 thinking → text → tool_use。thinking 块 schema 包含可选 signature，回放说明要求存在时原样保留；可选不等于实际必返，也不能伪造一个空签名来补齐。跨账户、跨地区、跨协议的签名不应获得原生回放权限。[Messages 内容及输入块 schema](https://platform.kimi.com/docs/api/messages)

图片可用 base64；其 URL 类型指 `ms://<file_id>` 引用。Responses 的 input_image 则明确 data URL，不支持公网 HTTP(S) URL。首批直接复用 barness 的自含图片编码，不凭“兼容 Anthropic/OpenAI”放行外部 URL 或上传文件资源。[Messages 图片块](https://platform.kimi.com/docs/api/messages)、[Responses input_image](https://platform.kimi.com/docs/api/responses)

### 6.2 流式、终态与文档冲突

SSE 顺序为 message_start → content block start/delta/stop → message_delta → message_stop，delta 分 text/thinking/signature/input_json；需要分块汇总工具 JSON，并跨 message_start/message_delta 合并 usage。stop_reason 列 end_turn/max_tokens/tool_use/refusal，命中 stop_sequences 也可能是 end_turn。[Messages stream schema](https://platform.kimi.com/docs/api/messages)

**待核实的官方内部差异**：Claude Code 教程展示 K2.7/2.6 和 `kimi-k3[1m]`，还称 K3 thinking 默认开启且可关闭；Messages API schema 却只列 K3、没有 thinking 字段，而 Chat 参数参考写 K3 始终推理。它可能涉及兼容层、工具特有行为或文档版本差异，本次没有真实请求可以消除歧义。首批按照 Messages 参考页的 K3 + effort 合同接入，不承诺 off，不能据教程直接写出未声明的 disabled 请求。[Claude Code 教程](https://platform.kimi.ai/docs/guide/claude-code-kimi)、[Messages](https://platform.kimi.ai/docs/api/messages)、[Chat 模型参考](https://platform.kimi.ai/docs/api/models-overview)

## 7. 缓存、用量与价格

### 7.1 缓存是默认服务行为

K3 默认自动写入 5m 缓存；Chat/Responses 通过 `prompt_cache_options:{mode:"implicit",ttl:"5m"|"1h"}`，Messages 通过**顶层** `cache_control:{type:"ephemeral",ttl:...}` 控制。Messages 消息内部 cache_control 被忽略；Responses 的显式 prompt_cache_breakpoint 会导致 400。这与原生 Anthropic 的逐块断点并不等价。[中国缓存指南](https://platform.kimi.com/docs/guide/context-caching)、[国际缓存指南](https://platform.kimi.ai/docs/guide/context-caching)、[Messages](https://platform.kimi.com/docs/api/messages)、[Responses](https://platform.kimi.com/docs/api/responses)

省略配置不等于关闭，也未找到关闭写入或手动删除的契约。缓存按 org 隔离，同 org 共享；两档 TTL 不互通。指南又指出 TTL 在首次写入后锁定，命中及新增部分沿用原 TTL，过期后才能切换。Responses 回显实际应用的 options，Chat 不回显，Messages 提供实际写入 TTL 明细。后续实现不能仅用本次请求 ttl 给全部写入计价。[缓存指南](https://platform.kimi.com/docs/guide/context-caching)、[Chat 缓存说明](https://platform.kimi.com/docs/api/chat)、[Responses 缓存说明](https://platform.kimi.com/docs/api/responses)

安全含义：派生缓存 key/metadata 可以避免泄露租户标识，但不能把厂商 org 级缓存升级为租户级隔离保证。Host 如需厂商侧硬隔离，应分配各自账户/org，按实际 Account Scope 绑定；不把缓存命中作为授权机制。这是对官方隔离粒度与本仓库[Binding/Account Scope 合同](../../GLOSSARY.md)的设计推论。

### 7.2 用量转换不能重复计费

| 协议 | 厂商总输入 | CacheRead | CacheWrite | 转为 barness `Usage.Input` |
| --- | --- | --- | --- | --- |
| Chat | `prompt_tokens` | `prompt_tokens_details.cached_tokens`（旧顶层 cached_tokens 是兼容来源） | `prompt_tokens_details.cache_write_tokens` | 总输入 − 读 − 写 |
| Responses | `input_tokens` | `input_tokens_details.cached_tokens` | `input_tokens_details.cache_write_tokens` | 总输入 − 读 − 写 |
| Messages | `input_tokens + cache_read_input_tokens + cache_creation_input_tokens` | `cache_read_input_tokens` | `cache_creation_input_tokens` | 厂商 input_tokens 本身 |

来源：[用量迁移表](https://platform.kimi.com/docs/guide/context-caching)。Messages 的 `cache_creation.ephemeral_5m_input_tokens/ephemeral_1h_input_tokens` 是写量分档，不能再加进总量；Responses `output_tokens_details.reasoning_tokens`、Messages `output_tokens_details.thinking_tokens` 是输出内部分项，也不应叠加到输出成本。[Responses usage](https://platform.kimi.com/docs/api/responses)、[Messages usage](https://platform.kimi.com/docs/api/messages)

仓库 Chat 与 Responses 已经解析 `cache_write_tokens` 并扣除读写后形成 Input，可直接复用；主要缺口是缺失/零与计价可靠性、实际 TTL、异常计数的边界校验（非负、读写之和不大于总输入）。用量完整性按各协议现有合同判定：Chat 请求推理而缺 reasoning count 会 partial，Responses 也要求 reasoning count；Messages 的 thinking_tokens 当前是可选细分，缺失本身不降 partial。流中断没有最终 usage 不表示免费；`UsageReporting` 表示 token 报告完整度，不能代替成本可靠性标识。[Chat 转换](../../ai/chat_usage.go)、[Responses 转换](../../ai/responses_usage.go)、[Messages 转换](../../ai/anthropic_stream.go)、[用量合同](../../ai/usage.go)

### 7.3 两站独立价格快照

当前完整价格表给出以下 K3 费率，单位为每 1,000,000 tokens，**两站分别定价，非换算关系**；本次表格对 K3 只有一行，未列长上下文阶梯：

| 站点/币种 | 未缓存输入 | 5m 写入 | 1h 写入 | 缓存读取 | 输出（含推理） |
| --- | ---: | ---: | ---: | ---: | ---: |
| 中国 CNY | ¥20 | ¥20 | ¥40 | ¥2 | ¥100 |
| 国际 USD | $3.00 | $3.00 | $6.00 | $0.30 | $15.00 |

来源：[中国价格表](https://platform.kimi.com/docs/pricing/chat)、[国际价格表](https://platform.kimi.ai/docs/pricing/chat)的 HTML 内嵌 `DocTable` rows（普通网页提取和 `.md` 中表格为空，本次额外读取原站 HTML，未执行其脚本）；缓存项与[中国缓存指南](https://platform.kimi.com/docs/guide/context-caching)、[国际缓存指南](https://platform.kimi.ai/docs/guide/context-caching)一致。国际定价页注明不含适用税费。K3 1h 写价恰好是输入的两倍，与当前算式一致，仍需实际写入 TTL，不能仅凭倍率正确就声称成本已完整。

作为后续 Chat 扩展的价格参考，两站 K2.x 表另列：

| 型号 | 中国：缓存读 / 输入 / 输出（CNY/M） | 国际：缓存读 / 输入 / 输出（USD/M） |
| --- | --- | --- |
| `kimi-k2.7-code` | 1.30 / 6.50 / 27.00 | 0.19 / 0.95 / 4.00 |
| `kimi-k2.7-code-highspeed` | 2.60 / 13.00 / 54.00 | 0.38 / 1.90 / 8.00 |
| `kimi-k2.6` | 1.10 / 6.50 / 27.00 | 0.16 / 0.95 / 4.00 |

来源同上。K2.x 表没有缓存写价；缓存指南只明确 K3 支持 Cache Write，FAQ 使用 `kimi-k2.7`/`kimi-k2.7-highspeed` 的简称，不据此发明可用 model ID。[中国缓存 FAQ](https://platform.kimi.com/docs/guide/context-caching)

仓库 `CostRates` 与 `UsageCost` 明确以 USD 为单位，因此中国人民币费率不能直接填入并继续标 USD。建议为中国站保留原币来源，由 Host 提供带来源/日期的汇率快照形成独立 USD Price Snapshot；或先明确扩展估算币种合同。两者属于需要采纳的设计决定，不能暗用实时汇率、国际站费率或零值补齐。[当前成本类型](../../ai/usage.go)

## 8. 错误与故障处理

官方错误码包括 400 参数/内容审查、401 认证、403 权限、404 模型/资源、429 限速/余额/过载，以及 499/500/503/504。特别是 429 的 `error.type` 必须区分 engine_overloaded、rate_limit_reached、exceeded_current_quota；不能统一认为可重试的瞬时限速。504 可能是 HTML 超时页，错误处理需要有界读取并兼容非 JSON。[错误码](https://platform.kimi.com/docs/api/errors)

Messages 参考定义 Anthropic 风格 `type:error/error/request_id`，其他 endpoint 使用 `error.type/message/code`，不能依靠错误正文暴露用户输入或 Key。建议保留 HTTP 状态、脱敏分类、厂商 request ID（存在时）与 Retry-After；确切响应头名需实测后加入白名单，不能把请求签名头当 request ID。[Messages 错误 schema](https://platform.kimi.com/docs/api/messages)、[API 概述](https://platform.kimi.com/docs/api/overview)

重试继续由 Binding 显式配置，流开始后不自动重放；不采用 Kimi Partial Mode 的自动断线续写作为 SDK 隐式恢复。取消、超时、有界资源与多租户归属均沿用现有公共 Client 生命周期。这是仓库[重试 ADR](../../docs/adr/0006-barness-ai-retry-policy-on-binding.md)与[协议边界 ADR](../../docs/adr/0001-barness-ai-protocol-boundary.md)要求，不是对 Kimi 新增限制。

## 9. barness-ai 的具体接入设计

### 9.1 复用 DeepSeek 的结构，按 Kimi 契约编码

[ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)、[ADR-0015](../../docs/adr/0015-barness-ai-deepseek-chat-completions.md)确立了 Provider 与 API 分离、复用协议 adapter、按 Provider 派生能力及协议独立型号数据的结构。Kimi 应复用这套边界，不能复制 DeepSeek 的 max_tokens、thinking 或缓存请求。

| 代码落点 | 仓库当前事实 | 建议改动 |
| --- | --- | --- |
| [binding.go](../../ai/binding.go)、[adapter.go](../../ai/adapter.go) | Provider 与 API 分离；registry 按协议复用；openai-completions 就是 Chat | 增加 `moonshot-cn` 与 `moonshot`（命名提议）两个 Provider；仍用三种已有 API，不增加 Kimi 协议或 Site 维度 |
| [chat_compat.go](../../ai/chat_compat.go)、[chat_options.go](../../ai/chat_options.go) | default 会沿用 OpenAI 自动字段；DeepSeek toggle 还会连发 effort | Kimi 精确 profile：stream usage 开、store/developer 关、使用 max_completion_tokens；K3 effort 直接编码，不发 thinking；缓存选项单独编码 |
| [responses_capabilities.go](../../ai/responses_capabilities.go) | default 会发送 OpenAI store/encrypted/cache/tier 相关字段 | Kimi profile：不开 encrypted/store/serviceTier/summary 请求；保留支持的 cache key；按已公开的 summary SSE/历史接收推理，不和请求 summary 开关混淆 |
| [anthropic_capabilities.go](../../ai/anthropic_capabilities.go)、[anthropic.go](../../ai/anthropic.go) | 默认开启原生 Anthropic 扩展；当前用 BetaMessageService、X-Api-Key | Kimi 精确 profile、Bearer、稳定 Messages URL；单独 effort 编码与顶层缓存，复用生命周期/parser；不附带 beta=true、native adaptive/budget/display 等未确认字段 |
| [catalog_models.go](../../ai/catalog_models.go)、[catalog_validate.go](../../ai/catalog_validate.go) | 型号完整身份含操作/Provider/API/ID；已有 RequiresThinking、ThinkingLevelMap | 六份 K3 独立元数据；low/high/max 按协议映射、默认 max；不把 K2.x 的二态混进 K3；模型 compat 只能收窄 profile |
| [chat_usage.go](../../ai/chat_usage.go)、[responses_usage.go](../../ai/responses_usage.go)、Messages usage 转换 | 已支持 Moonshot 顶层 cached_tokens、新 cache_write_tokens 归一化与多个推理文本 key | 复用读写转换，补实际 TTL、计数边界和缺失状态；Kimi profile 打开 Messages 已有 TTL 分档接收，不把 unknown 计价混同 token 报告完整度 |
| [usage.go](../../ai/usage.go) | 已有 CacheWrite1h，1h 算式是 Input×2；价格类型是 USD | K3 两站倍率匹配，不必为此新增价格抽象；实际 TTL 可证明才精确套费率；中国币种需要明确快照策略 |
| [native.go](../../ai/native.go)、[hooks_run.go](../../ai/hooks_run.go) | 原生封套精确比对租户/账户/Provider/API/型号，认证头受保护 | 保持跨站/跨协议不互授原生状态；profile 校验必须覆盖最终 SamplingParams 与 OnPayload，不能让 hook 扩大能力 |

当前底层依赖已经有 OpenAI Go SDK v3 与 Anthropic Go SDK，adapter registry 已存在，研究没有发现需要新 SDK 的理由。[go.mod](../../go.mod)；现有地区 Provider 决策要求身份分离，[ADR-0049](../../docs/adr/0049-barness-ai-regional-provider-identities.md)可直接作为地区设计依据。

目录还要区分厂商的默认输出预算与可请求上限：K3 文档分别为 131072 与 1048576，而现有 `Model.MaxTokens` 同时参与能力上限与 Simple/Messages 默认预算。不能只填厂商最大值就声称保留了默认行为；实施时应明确初批支持的输出上限与 Host 预算，若要同时完整表达两者，再记录最小的数据模型调整。[模型字段](../../ai/catalog_models.go)、[Simple 默认预算](../../ai/options.go)、[Messages 请求](../../ai/anthropic_request.go)、[K3 参数](https://platform.kimi.com/docs/api/models-overview)

### 9.2 几项不能省略的行为

- **最终 wire 校验**：Responses 仅 auto；Messages auto/any/none；Chat K3 auto/none/required。命名选项、型号限制在读取凭据前校验；OnPayload 是凭据 pin 后的 hook，其产物在 HTTP 前再验。Host 自定义目录和 hook 均不能扩大 profile，显式错误字段也不能悄悄丢掉。[当前 resolve](../../ai/call_runtime.go)、[payload 边界](../../ai/hooks_run.go)
- **K3 effort 与保留式思考**：分别编码 reasoning_effort / reasoning.effort / output_config.effort。Chat/Responses Simple 入口保留现有规范档位 clamp 合同（如 medium→high、xhigh→max），精确 wire 字段只接受 low/high/max；Messages 的显式 Effort 也只接受这三档，不借 Anthropic adaptive 间接生成。显式关闭推理不应静默转为 low；默认 max 用现有 DefaultThinkingLevel 表达。[现有档位规则](../../ai/reasoning.go)、[现有选项](../../ai/options.go)、[模型数据合同](../../ai/catalog_models.go)
- **缓存 TTL 与计价**：不能直接打开 SupportsExplicitPromptCacheMode：当前 encoder 会发 `mode:explicit` 或 `ttl:30m`，不符合 Kimi；可复用 DTO，但要按 Provider 编码 implicit/5m/1h，Messages 放顶层。首批可先固定 5m，再扩展 1h；这不能保证厂商 org 中没有别的客户端已创建 1h 缓存。Chat 无实际 TTL 时保留用量，并决定显示估算假设/区间还是增加成本可靠性状态，不把请求值当实际证明，不把 `CacheRetentionNone` 承诺为关闭收费缓存。[现有 encoder](../../ai/responses_options.go)、[缓存事实](https://platform.kimi.com/docs/guide/context-caching)
- **Host 装配与隔离**：每站三种 BindingID 独立；同站同账户经 Host 授权可引用同一凭据，不跨站复用。借现有 [glmhost](../../ai/examples/glmhost)、[minimaxhost](../../ai/examples/minimaxhost) 的账户装配结构，但不复制厂商语义；错误无自动换地区/模型/Key。
- **候选与正式支持分开**：先登记候选和官方能力表，完成各自证据后再进 BuiltinCatalog；共享 adapter 的成功不连带激活另一协议/站点。研究不修改 ADR、不生成已批准实施决定。[目录准入 ADR](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)、[地区身份 ADR](../../docs/adr/0049-barness-ai-regional-provider-identities.md)

### 9.3 建议实施顺序

1. 先采纳两站 Provider 命名、K3 范围、价格币种/快照策略，记录新的 Kimi ADR 和本地 spec；这份研究不是批准记录。
2. 先写公共 Client 边界失败场景，再接入六条候选路线：精确 URL/单一认证、字段限制、推理与工具往返、计费及跨作用域拒绝。实现三套内部 profile，不复制 adapter。
3. 完成每站三协议的自身真实验证，再按本仓库测试 SOP 推进人工阶段；K2.x、结构化输出、动态工具/namespace/custom/托管工具分别扩展，不让其阻塞三种基础协议。

## 10. 验证建议、未决事项与本次边界

本节是未来接入的验证设计，**没有执行这些测试**。遵循[测试流程 SOP](../../docs/agents/verification.md)：先失败后实现；公共 E2E 为主；live 必须另有该次明确范围、账户、来源和付费预算；race/压力按 SOP 的人工选择；pi 只作架构参考，不续建历史差分门禁。

| 必交范围 | 可重复核验的断言/产物 |
| --- | --- |
| 六条精确身份 | loopback 请求原件：method、host、path、无重复 v1、无 beta=true、只有正确 Bearer；错误站点/Key/目录不触发另一站请求 |
| 字段与档位 | 三种 effort 编码与默认值；Simple 保留 clamp，Full 精确 wire/Messages Effort 拒绝未支持值、显式禁用推理拒绝；固定采样冲突；tool_choice 各协议限制；SamplingParams/OnPayload 绕过被拦 |
| SSE 与终态 | reasoning/text/tool 增量、UTF-8 切分、JSON 参数分片、各终态、error、截断、取消和资源释放；usage 不被 finish_reason 提前截掉 |
| 工具往返/原生历史 | 原始 assistant reasoning/summary/thinking signature 与工具结果的下一轮 wire；相同完整身份可回放，跨租户/账户/站点/协议无授权回放 |
| 用量与成本 | 默认缓存写量、读/写/未缓存互斥、负数/超总量、缺失/零、5m/1h实际TTL、混合和阶梯边界；CNY/USD快照不可混用 |
| 自身真实支持 | 每站每协议 K3 的完整有限场景；模型与权限、实际认证、SSE、usage、request ID及历史接受；脱敏 capture + 哈希 + 命令/退出 + 消费者结果 |

待落地前解决：Messages 的 API reference 与 Claude Code thinking/型号差异；x-api-key 是否另行接受及稳定 SDK 的实际请求头/URL；Chat 写入 TTL 的可证明方法和中国价格币种策略；六条路线真实 request ID、错误体、signature 与原生历史接受性。文档支持不替代真实支持，低预算样例也不能证明全上下文容量或任意 JSON Schema 支持。

本次为纯研究文档：运行代码、型号目录、ADR 与支持矩阵均未修改；没有 E2E/live/race/压力运行，不宣称发布通过。后续真实调用、账户 Key 或付费预算均不由本研究授权。

文档检查范围为 Markdown 格式、本地引用存在性、仓库代码与建议的一致性，以及关键官方来源交叉核对；程序 E2E/live/race/压力对本次纯文档变更不适用。未来接入的验证范围仍按本节和 SOP 另行确定。

### 主要一手资料入口与读取说明

- 官方完整目录：[中国 llms.txt](https://platform.kimi.com/docs/llms.txt)、[国际 llms.txt](https://platform.kimi.ai/docs/llms.txt)。正文已就近引用，未采用搜索到的第三方转述作为结论依据。
- 三协议 API reference：中国 [Chat](https://platform.kimi.com/docs/api/chat) / [Responses](https://platform.kimi.com/docs/api/responses) / [Messages](https://platform.kimi.com/docs/api/messages)；国际 [Chat](https://platform.kimi.ai/docs/api/chat) / [Responses](https://platform.kimi.ai/docs/api/responses) / [Messages](https://platform.kimi.ai/docs/api/messages)。中国页面的官方 `.md` 附带完整 OpenAPI，可核嵌套字段；国际对应参考页交叉核对。
- 定价原站 HTML 在本次读取时的 SHA-256：中国 `85bc6d4c807e92130b9718b84723d6cf6c38ff30b6953cea92a59eb80023ed98`；国际 `f2f9bd654b785bba57f8bd6dbdf8129aa3f6ec00822c5e13be0e0aad5616441a`。这是本次页面快照标识，页面更新后需重新取价，不视为永久费率。
