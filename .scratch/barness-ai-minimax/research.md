> 历史中国区研究：2026-10-18用户明确改为国际站；当前接入依据为[国际站研究](research-international.md)。本文保留原事实与Hash，不授予国际站支持。

# MiniMax 三种兼容协议研究与 barness-ai 接入建议

研究日期：**2026-10-18（Asia/Shanghai）**。研究对象是 MiniMax 中国站 `platform.minimax.cn/docs`，国际站仅用于补充地域端点证据。范围是 OpenAI **Responses**、OpenAI **Chat Completions**、Anthropic **Messages**；MiniMax 自有生成协议不在范围。这里的 Completions 指 `/v1/chat/completions`，不是旧式文本 `/v1/completions`。

方法：浏览官方文档与全站索引，下载官方 Markdown 和三份 OpenAPI 交叉核对，并对照 barness-ai 现有协议能力边界（仓库 HEAD `ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`）。未读取密钥、未请求收费模型 API；本报告的“官方明确”“文档未说明”“文档冲突”“建议”分别表达，兼容声明不冒充线上验收。

## 1. 结论：三种协议都有明确官方入口

**MiniMax 当前已明确提供三种兼容协议，包括 Responses。** 不需要把 Responses 请求转换为 Chat，也不能沿用“只支持 Chat / Messages”的旧印象。官方索引列出三种生成接口及 Responses input token 估算，三份 OpenAPI 分别公开其路径。[官方索引](https://platform.minimax.cn/docs/llms.txt)、[Responses](https://platform.minimax.cn/docs/api-reference/responses-create)、[Chat Completions](https://platform.minimax.cn/docs/api-reference/text-chat-openai)、[Messages](https://platform.minimax.cn/docs/api-reference/text-chat-anthropic)

| 协议 | 中国站 SDK base URL | 生成接口完整地址 | 明确鉴权 |
| --- | --- | --- | --- |
| OpenAI Responses | `https://api.minimax.cn/v1` | `POST https://api.minimax.cn/v1/responses` | `Authorization: Bearer <API_KEY>`。[创建参考](https://platform.minimax.cn/docs/api-reference/responses-create) |
| OpenAI Chat Completions | `https://api.minimax.cn/v1` | `POST https://api.minimax.cn/v1/chat/completions` | `Authorization: Bearer <API_KEY>`。[SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[参考](https://platform.minimax.cn/docs/api-reference/text-chat-openai) |
| Anthropic Messages | `https://api.minimax.cn/anthropic` | `POST https://api.minimax.cn/anthropic/v1/messages` | 支持 Bearer 与 `x-api-key`；两者同时出现优先 Bearer，官方推荐 Bearer。[OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json) |

Messages 的 SDK 基址不包含末尾 `/v1`；两个 OpenAI 协议共享基址但使用独立的请求、对象和事件模型。官方 Chat OpenAPI 只有 `/v1/chat/completions`，未找到旧式 `/v1/completions` 的官方兼容契约，不能声明支持旧文本 Completions。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

### 1.1 国内、国际与历史域名

国际站对应基址为 `https://api.minimax.io/v1` 与 `https://api.minimax.io/anthropic`，Responses 为 `POST https://api.minimax.io/v1/responses`。这是官方国际站自己的说明；本报告没有验证国内、国际账户或 Key 可以互换，不能自动跨地域故障转移。[国际 OpenAI SDK](https://platform.minimax.io/docs/api-reference/text-openai-api)、[国际 Anthropic SDK](https://platform.minimax.io/docs/api-reference/text-anthropic-api)、[国际 Responses](https://platform.minimax.io/docs/api-reference/responses-create)

请求的中国文档根 `/docs` 会进入 `/docs/guides/models-intro`。中国站 `llms.txt` 内的绝对文档链接仍使用 `platform.minimaxi.com`，但本次在用户指定 `platform.minimax.cn` 上取得页面、Markdown 和 schema；API 示例的当前中国域名是 `api.minimax.cn`。文档站域名与模型 API 域名须分开，不根据旧拼写自行替换或宣称重定向兼容。[中国入口](https://platform.minimax.cn/docs)、[索引](https://platform.minimax.cn/docs/llms.txt)

### 1.2 Coding Plan 已经演进到 M Plan

当前文档产品为 **M Plan**，保留 Token Plan 老用户说明；本次 `/docs/token-plan/quickstart` 与 `/docs/token-plan/faq` 实际重定向至对应 M Plan 页面。订阅 Key 与按量 Key 独立，订阅 Key 使用套餐额度及积分包，不自动扣账户余额；按量 Key 使用账户余额。两种 Key 不能混用。[M Plan 开通](https://platform.minimax.cn/docs/m-plan/quickstart)、[M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)、[旧入口](https://platform.minimax.cn/docs/token-plan/quickstart)

M Plan 的兼容基址仍为上表的 `/v1` 和 `/anthropic`，官方没有在所查配置中要求另一个 `/coding/...` 生成端点。支持自定义兼容协议的 AI 工具可以使用订阅 Key；FAQ 将套餐定位为个人开发者交互式使用，生产环境建议按量付费。barness 本地交互式与云端生产部署应分别选择凭据用途，不能照搬其他厂商的 Coding URL 或套餐适用名单。[其他工具](https://platform.minimax.cn/docs/m-plan/other-tools)、[M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)

**模型可用性有现实限制**：三协议页面均注明 `MiniMax-M3.1-Flash-Preview` 暂时仅通过 M Plan 和 MiniMax Code 提供。它有官方协议契约，但不能据此承诺普通按量 Key 可用。[Responses](https://platform.minimax.cn/docs/api-reference/responses-create)、[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)

## 2. 型号与思考：必须按“型号 × 协议”区分

### 2.1 版本迁移会影响目录和验收

当前 Chat / Messages 型号列表包括 `MiniMax-M3.1-Flash-Preview`、`MiniMax-M3`、M2.7 / M2.5 / M2.1 及其 highspeed 版本、M2。M3.1 / M3 的上下文标为 1,000,000 tokens，M2.x 标为 204,800。Responses model 是自由字符串，说明明确讨论 M3.1、M3、M2.x，未像 Chat / Messages 一样列封闭 enum；不要把自由字符串当作平台任意模型均支持。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

官方概览已公告：**M2.5 / M2.7 系列于 2026-10-19 停止向新用户开放，2026-11-18 正式下线**；按量老用户可调用至下线，M Plan / Token Plan 请求自 10 月 19 日统一路由至 M3.1 Flash Preview，旧名保留。因此新接入不宜把 M2.7 作为长期默认；套餐旧 ID 也不再可靠代表其旧模型语义。[模型概览与下线通知](https://platform.minimax.cn/docs/guides/models-intro)

Chat / Messages 的输出字段参考建议 M3.1 / M3 使用 131072，上限 524288；M2.x 建议 65536，上限 204800。这些是输出预算，不能直接拿上下文窗口减法当作平台全部可接受组合；Responses 的 `max_output_tokens` 本次 schema 没有给同样的数字上下限，不把另一协议字段说明悄悄移植过去。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)、[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

### 2.2 三种协议的思考默认值不同

| 型号 | Chat Completions | Responses | Messages |
| --- | --- | --- | --- |
| M3.1 Flash Preview | 默认开启且不可关闭；`thinking.type=disabled` / `reasoning_effort=none` 返回 400 | 默认开启且不可关闭；`reasoning.effort=none` 返回 400 | 默认开启且不可关闭；`thinking.type=disabled` 返回 400 |
| M3 | 默认开启；`thinking.type=disabled` 可关闭 | 默认关闭；非 none 的 `reasoning.effort` 开启，但不调节实际深度 | 默认关闭；`thinking.type=adaptive` 开启 |
| M2.x | 不可关闭；disabled 被接收但不生效 | 不可关闭；effort none 被接收但不生效 | 不可关闭；disabled 被接收但不生效 |

上表来自各兼容协议自己的字段说明，不是从自有协议推断。[Chat 思考控制](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Responses 思考控制](https://platform.minimax.cn/docs/api-reference/responses-create)、[Messages 思考控制](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)

M3.1 真正调节深度的档位均为 **low / medium / high / xhigh / max，默认 max**，但编码位置不同：Chat 用 `reasoning_effort`，Responses 用 `reasoning.effort`，Messages 用 `output_config.effort`。Chat / Messages 说明其他型号忽略深度字段。Responses schema 虽把 minimal / none 放在全体 enum 内，M3.1 的文字硬约束只接受上面五档；不能因 schema 总 enum 存在 minimal 就承诺 M3.1 可用。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)、[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

三协议均明确说明推理 tokens 计入输出预算。预算过小可能只有思考而无最终正文：Chat `finish_reason=length` 且 content 为空，Messages `stop_reason=max_tokens` 且无 text，Responses `status=incomplete` 且 output 无 message；这些应返回截断结果，而非伪造成功答案。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[Responses](https://platform.minimax.cn/docs/api-reference/responses-create)

## 3. OpenAI Chat Completions 的实际兼容边界

### 3.1 请求、工具和采样

| 主题 | 官方证据与边界 |
| --- | --- |
| 角色 | schema 列 system / user / assistant / tool；没有 developer。工具结果携带 tool_call_id。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json) |
| 生成预算 | `max_completion_tokens` 是新接入推荐字段；`max_tokens` 作为弃用旧字段仍列出。两者不能同时随意生成。[参考](https://platform.minimax.cn/docs/api-reference/text-chat-openai) |
| 函数工具 | tools 为 OpenAI Chat 的嵌套 function 结构，返回 tool_calls[].function.arguments JSON 字符串；旧 function_call 不再推荐。[SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json) |
| 工具选择 | 当前 Chat 请求 schema 没有 tool_choice / parallel_tool_calls 字段。文档缺失不是已证明拒绝或忽略；auto 可作为首期默认，强制调用/禁止并行须独立验证。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json) |
| 采样 | temperature `[0,2]` 默认 1；top_p `[0,1]`，M3.1 / M3 默认 0.95，M2.x 默认 0.9。n 仅支持 1。[SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api) |
| 明确忽略 | presence_penalty、frequency_penalty、logit_bias 被忽略。[SDK 注意事项](https://platform.minimax.cn/docs/api-reference/text-openai-api) |
| 未获字段证据 | 本次 schema 未列 store、response_format、strict、seed、prompt_cache_key、prompt_cache_retention；不能默认继承 OpenAI 全套参数。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json) |
| 服务层级 | standard / priority，默认 standard；priority 按标准价格 1.5 倍计费，不是 OpenAI 原厂固定倍率。[SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[定价](https://platform.minimax.cn/docs/pricing/overview#models/text) |

### 3.2 推理输出与历史回放有文档冲突

新 Chat SDK 页与 schema 说明 `reasoning_split=true` 将思考拆到 `reasoning_content`，false 则在 content 内使用 think 标签；M3.1 暂不允许 false，流式例子也从 delta.reasoning_content 读取。该开关只控制格式，不决定是否思考。工具往返须完整保留 assistant 返回，包括工具与推理内容。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)

但是专门的 **M3 Tool Use 文档**仍展示 `reasoning_split=true` 返回 `reasoning_details[]`，包含 type、id、format、index、text，并要求完整回传；原始格式则保留整个带 think 标签的 content。Prompt 缓存页的 M3 代码也使用 reasoning_details。**这是当前官方资料的差异，不应自行裁定所有型号都改成 reasoning_content，也不能把 reasoning_details 数组压平成文本后称为原样回放。**[M3 工具与交错思考](https://platform.minimax.cn/docs/guides/text-m3-function-call)、[Prompt 缓存](https://platform.minimax.cn/docs/api-reference/text-prompt-caching)

M3.1 使用 reasoning_content 的证据较完整；M3 的分离字段实际形状须录制真实 wire，或选择完整保留原始 content 的路径并验证往返。SDK 主页的泛化 think 标签说明与 M3.1 专门说明也不完全一致，应以型号化示例和后续 Live 证据收敛，不能通过猜测抹平。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)

### 3.3 流式、终态与用量

官方支持 `stream=true`，增量正文走 choices[].delta.content，M3.1 思考走 delta.reasoning_content；`stream_options.include_usage=true` 可要求流中用量。Chunk schema 说 usage 仅最后一个 chunk 返回；本次示例未给出完整的原始 `data: [DONE]` 序列，因此结束哨兵和“finish 后还有 usage-only chunk”应保留为 Live 验证点，不要把 SDK 循环示例当作完整传输契约。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)

非流式 Choice enum 包括 stop / length / content_filter / tool_calls，流式 Chunk enum 却只列 stop / length；schema 的 Usage 只声明 total_tokens，但同一份官方响应样例已有 prompt_tokens、completion_tokens、prompt_tokens_details.cached_tokens。实现应解析已给出的明细，保留未知停止原因，不因 enum 不齐把工具或过滤误判为自然完成。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)

官方 Chat 流式 schema 没有完整表达 tool_calls 参数 delta，工具使用页主要展示非流式往返。因此函数调用已获明确支持，**参数增量的具体分片、index/ID 首帧与终帧行为还需要协议 Live 证据**；不要照搬 GLM 的 tool_stream 开关或 DeepSeek 的帧特点。[Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json)、[M3 工具指南](https://platform.minimax.cn/docs/guides/text-m3-function-call)

## 4. OpenAI Responses 的实际兼容边界

### 4.1 已明确支持的输入、输出与工具

官方 OpenAPI 当前仅有 `POST /v1/responses` 与 `POST /v1/responses/input_tokens`。input 可为字符串或完整历史数组，条目包括 message / function_call / function_call_output / reasoning；message 角色列 user / assistant / system / developer / tool。可传 instructions、metadata、max_output_tokens、stream。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)、[创建](https://platform.minimax.cn/docs/api-reference/responses-create)

函数工具为 Responses 的顶层 type=function、name、parameters，返回 function_call 条目，以 call_id 关联 function_call_output；arguments 是 JSON 字符串。工具结果可为字符串或内容数组。公开工具类型只有 function；tool_choice 只有 none / auto，不承诺 required、具名强制函数、OpenAI 托管工具、namespace 或 custom。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

输出是 message / reasoning / function_call 列表。message.content 是 output_text 数组；reasoning.content 是 reasoning_text 数组，summary 示例为空。usage 提供 input_tokens、input_tokens_details.cached_tokens、output_tokens、output_tokens_details.reasoning_tokens、total_tokens。[创建响应示例](https://platform.minimax.cn/docs/api-reference/responses-create)

### 4.2 有意缩小的兼容子集

| 字段或能力 | 当前官方契约 |
| --- | --- |
| 保存 | 请求 store 被忽略，响应固定 false；不能承诺服务端保存。[创建](https://platform.minimax.cn/docs/api-reference/responses-create) |
| 并行工具控制 | 请求 parallel_tool_calls 被忽略，响应固定 true；显式禁并行不能靠此字段实现。[创建](https://platform.minimax.cn/docs/api-reference/responses-create) |
| 截断 | 请求 truncation 被忽略，响应固定 disabled。[创建](https://platform.minimax.cn/docs/api-reference/responses-create) |
| 缓存 | 公开 prompt_cache_key 路由标识；未列 prompt_cache_retention 或 OpenAI 亲和头。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json) |
| 输出格式 | text.format.type 仅 text；没有 json_object / json_schema / strict 契约。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json) |
| 推理选项 | reasoning 公开字段只有 effort；未列 summary 请求参数、encrypted_content include。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json) |
| 服务端续接 | 未列 previous_response_id、conversation、background，也没有 retrieve/delete/cancel 路径；表示本次未获官方契约证据，不等于实测拒绝每个字段。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json) |
| 采样 | temperature 与 top_p 的文字范围为 `(0,1]`，schema minimum 却是 0。零值有效性有文档内冲突，且不能继承 Chat 的 temperature 最大 2。[创建](https://platform.minimax.cn/docs/api-reference/responses-create) |

### 4.3 SSE 已支持，但事件契约还不完整

创建参考明确 `stream=true` 使用 SSE，官方也有 stream 请求示例；本次下载的 Responses OpenAPI 没有 StreamEvent schema、逐事件序列、流末 usage 或 DONE 规则。**“Responses 流式支持”已经确认，“与现有 OpenAI reader 所需的事件终态完全吻合”尚未实测。** 应用仍以协议终态完成调用；断线不自动当作完成，也不凭 HTTP 200 推断完整成功。[创建参考](https://platform.minimax.cn/docs/api-reference/responses-create)、[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

最终对象 status 可为 completed / incomplete / failed；incomplete_details.reason 列 max_output_tokens / content_filter，failed 的 error 为 code/message。它们与 HTTP 鉴权、限流错误是不同层级，parser 应保留状态与原因。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

### 4.4 推理回放还需精确核验

输出 reasoning.content 是 reasoning_text 数组，但公开 InputItem 中 reasoning 只说明 summary 数组（summary_text），没有与输出同样完整的 reasoning.content 字段。该差异不能推导“随意回传 output 任何对象必然被接收”，也不能先行把完整推理改写为摘要。首期使用无状态完整历史工具往返，并用真实响应验证 reasoning 条目输入所需形状；原生状态应保留原始协议数据及来源。[Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

`POST /v1/responses/input_tokens` 可在不生成输出的情况下估算输入，输入支持模型、历史、指令和工具等；返回 object=response.input_tokens、input_tokens。本次不调用它，也不把估算结果当作最终计费 usage。库若未来引入预检，应作为明确操作，不在每次 Chat 内隐藏新增远程请求。[Token 估算](https://platform.minimax.cn/docs/api-reference/responses-input-tokens)

## 5. Anthropic Messages 的实际兼容边界

### 5.1 参数、工具和思考块

官方 SDK 指南覆盖 messages、system、max_tokens、stream、temperature、top_p、tools、tool_choice、thinking、output_config.effort、metadata、service_tier；消息文本、tool_use、tool_result、thinking 得到明确说明。工具定义使用 input_schema，工具返回通过 user 消息的 tool_result.tool_use_id 关联。[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

SDK 页对 tool_choice 笼统写“完全支持”，但详细 OpenAPI ToolChoice **明确只支持 auto / none**；没有 any、具名 tool 或 disable_parallel_tool_use。接入时采用详细契约的保守子集，不把泛化描述提升为 Anthropic 全部策略均支持。[SDK 兼容表](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[ToolChoice schema](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

thinking 只列 adaptive / disabled，没有 Claude 预算型 enabled/budget_tokens、display、output_config.format 或专有 beta 的对应声明。M3.1 的 effort 五档与 M3 开关应分别编码，不能把“支持 adaptive”视为支持原厂整组扩展。[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

官方明确忽略 top_k、stop_sequences、mcp_servers、context_management、container。temperature `[0,2]`，top_p `[0,1]`；忽略 stop_sequences 意味着调用者不能依赖其保证停止边界。[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)

工具往返必须完整回带 response.content 的 thinking/text/tool_use 全部内容块。Messages OpenAPI 的 thinking 输出示例**带 signature**，请求、响应 schema 均说签名须原样保留；因此现有有签名可信回放是有依据的路径。文档没有保证每种型号/流式都一定签名，缺签名情况需要 Live 验证，不能预设必须放宽所有 Provider 的回放边界。[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

### 5.2 SSE、用量和错误

Messages 明确列出 message_start、ping、content_block_start/delta/stop、message_delta、message_stop；delta 类型列 text_delta、thinking_delta、signature_delta，SDK 示例分别读取 thinking/text。message_delta 的 stop_reason 列 end_turn / max_tokens / tool_use，usage 可在 message_start 的 message 和 message_delta 中返回。[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)、[SDK 流式示例](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)

公开 StreamEvent schema 没完整列工具参数 input_json_delta / partial_json，不能把 enum 缺项直接理解成工具流式一定不支持；函数工具已确认，参数逐段输出与开始块 input 的实际组合仍需验收。[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

用量字段为 input_tokens、output_tokens、cache_creation_input_tokens、cache_read_input_tokens。显式缓存的总输入应计算为 input + creation + read，不能像 Chat 那样把 cache_read 再视为已含于 input 的子集；流开始的输入缓存用量与后续输出用量须合并，缺失字段不能伪装成已知零。[主动缓存用量说明](https://platform.minimax.cn/docs/api-reference/anthropic-api-compatible-cache)、[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

错误 schema 明确为 HTTP 状态 + JSON，包含 type=error、request_id、error.type/message；流中异常用 `event: error`，收到后停止读取并清理状态。类型包含 invalid_request_error、authentication_error、permission_error、not_found_error、request_too_large、rate_limit_error、api_error、overloaded_error。无需从 MiniMax 自有 base_resp 错误结构反推 Messages。[Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

### 5.3 Prompt 缓存：支持明确，但不要照搬原厂 TTL

MiniMax 有自动重复前缀缓存，官方分别展示 Chat 的 prompt_tokens_details.cached_tokens 与 Messages 的 cache_read_input_tokens；不要求应用为了自动缓存生成原厂 prompt cache key。[Prompt 缓存](https://platform.minimax.cn/docs/api-reference/text-prompt-caching)

Messages 另外明确提供 `cache_control:{type:"ephemeral"}` 主动缓存，可放 tools、system、messages 的相应块；文档生命周期为 **5 分钟，命中刷新**。一次最多使用最近的 4 个断点，每个断点向前检查 20 块。当前 schema 的 CacheControl 只有 type，没有 ttl；不能把 Anthropic 的 `ttl:"1h"` 直接当成 MiniMax 能力。[主动缓存](https://platform.minimax.cn/docs/api-reference/anthropic-api-compatible-cache)、[CacheControl schema](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json)

主动缓存价格表与例子主要是 M2.x，未给 M3.1 / M3 的完整主动写入价格。M3 的自动缓存已有例子，新型号显式缓存的命中与计价仍须分别确认；不能把平台一般缓存声明当作每款型号每种模式同价。[主动缓存](https://platform.minimax.cn/docs/api-reference/anthropic-api-compatible-cache)、[Prompt 缓存](https://platform.minimax.cn/docs/api-reference/text-prompt-caching)

## 6. 跨协议注意事项：模态、错误、限流与计费

### 6.1 多模态兼容范围

Chat / Messages SDK 明确 M3.1 / M3 可通过 image_url / video_url 或 image / video 内容块输入图片与视频，M2.x 不支持这些输入；Chat SDK 明确当前不支持音频输入。Responses schema 另列 input_image / input_video，但没有同样完整的逐型号模态清单，应分别验证，不能从另两协议直接外推。平台语音、图像生成等能力也不等于三种兼容协议可用。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)、[Responses ContentPart](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

Chat / Messages SDK 页给出图片 10 MB、URL/base64 视频 50 MB、请求体 64 MB；detail 为 low/default/high。`video_url`、`input_video` 和 `mm_file://...` 属于兼容路径扩展，barness 当前领域模型是否表达、Files 上传是否在范围必须另行设计；本次不扩展自有文件 API。支持文本后不能顺势声称视频已接通。[OpenAI SDK](https://platform.minimax.cn/docs/api-reference/text-openai-api)、[Anthropic SDK](https://platform.minimax.cn/docs/api-reference/text-anthropic-api)

### 6.2 错误诊断与资源边界

平台错误页建议保留响应 Header 中的 **trace_id** 供排障，列出 1002 限流、1004/2049 鉴权、1008 余额不足、1039 token 限制、2013 参数错误、2056 M Plan 资源限制。该页是平台级说明，不保证三个兼容协议都原样使用数字业务错误；应结合每个协议的实际 HTTP/body/SSE 记录处理。[错误码](https://platform.minimax.cn/docs/api-reference/errorcode)

速率限制包括 RPM、TPM，且主/子账户共享账户限额。M Plan 还有独立的 5 小时与周额度窗口；瞬时限流、套餐资源耗尽、余额不足不能统一无限重试，更不能自动换 Key、改协议、改区域。Host 仍要提供账户级并发、预算、超时与取消边界。[速率限制](https://platform.minimax.cn/docs/guides/rate-limits)、[M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)

### 6.3 价格不是简单的一个 token 单价

中国站按人民币列价。M3 标准价格在输入 ≤512k / >512k 两档不同，输入、输出、缓存读价都随档位变化；priority 是对应标准档的 1.5 倍。主动缓存另有写入价格，套餐额度/积分又不等同于账户账单。不能只把某个小上下文标准价填入模型目录就宣称准确计费。[定价](https://platform.minimax.cn/docs/pricing/overview#models/text)、[主动缓存价格](https://platform.minimax.cn/docs/api-reference/anthropic-api-compatible-cache)、[M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)

## 7. barness-ai 建议方案：借鉴 DeepSeek 的边界，不照抄它的规则

本节是设计建议，尚未实现；协议事实来自上文，本地代码位置用相对链接方便复核。总方向是：**一个明确的地域 Provider，三个独立协议 Binding，复用现有三套 adapter；差异收敛在 Provider 协议能力边界，型号硬约束留在型号目录。** 不增加 MiniMax 专有 API 名，不建立全新 SDK/请求生命周期，也不做三协议互转。

### 7.1 身份、绑定与交付范围

建议先支持中国区 `ProviderMiniMaxCN = "minimax-cn"`，与冻结 pi 的身份保持一致；国际区若正式纳入，再以 `minimax` 独立 Provider、凭据作用域和验收，不用同一个 Provider 根据 URL 猜地域。

| Operation × Provider × API | 建议 BindingID | SDK 基址 | 复用 |
| --- | --- | --- | --- |
| chat × minimax-cn × openai-completions | minimax-cn-chat | `https://api.minimax.cn/v1` | chatAdapter |
| chat × minimax-cn × openai-responses | minimax-cn-responses | `https://api.minimax.cn/v1` | responsesAdapter |
| chat × minimax-cn × anthropic-messages | minimax-cn-messages | `https://api.minimax.cn/anthropic` | anthropicAdapter |

建议以按量 M3 为首期正式接入候选，M3.1 作为具备订阅资格的独立型号候选；每个型号在三个协议下各有自己的目录身份。套餐用途由独立 Binding、账户作用域和凭据表达，不增加一个套餐维度的 Model 身份。M3.1 的文档能力可以研究并实现，但不能把它冒充云端按量默认；不把即将下线/路由的 M2.7 作为新用户默认。库接收 Host 显式提供的凭据，不自行读取 OPENAI_* / ANTHROPIC_* 环境变量。[现有 Binding](../../ai/binding.go)、[模型概览](https://platform.minimax.cn/docs/guides/models-intro)、[M Plan FAQ](https://platform.minimax.cn/docs/m-plan/faq)

三协议目录独立，尤其 M3 的不同思考默认不能从 Chat 一份配置派生其他两份；每项都必须满足其全部硬约束后才纳入内置目录。图片/视频、输出上限、思考必选与可用账户都需要明确表达。[目录纳入 ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)

冻结 pi-ai 1.1.0 的中国数据是 `minimax-cn × anthropic-messages`，国际数据是 `minimax × anthropic-messages`，其基址仍是 `api.minimaxi.com/anthropic` 与 `api.minimax.io/anthropic`。本次中国接入以当前 MiniMax 官方的 `api.minimax.cn` 为准，记录与冻结数据的差异。M3 的冻结条目还带图片 resize 限制，而 barness 当前 Model 尚无 inputLimits；不能只复制 input 模态后宣称完整实现 pi 的图像处理。应区分上游预处理策略与厂商硬限制，明确首期文本/工具子集，并以真实图片验收决定是否开放图片。[冻结来源与重建证明](../../ai/internal/testkit/pioracle/node/PROVENANCE.md)、[逐文件哈希](../../ai/internal/testkit/pioracle/node/model-data.sha256)、[当前 Model](../../ai/catalog_models.go)

### 7.2 Provider 能力必须显式声明

当前 [chatCompatOf](../../ai/chat_compat.go)、[responsesCapabilitiesOf](../../ai/responses_capabilities.go)、[anthropicCapabilitiesOf](../../ai/anthropic_capabilities.go) 对未识别 Provider 会回到原厂能力，新增 Provider 常量后直接调用会生成未经 MiniMax 确认的字段。按 DeepSeek 现有做法为 MiniMax 添加显式分支，按真实第二用例细拆已有小能力结构；不要先造通用策略注册中心。

Provider × API 固有的字段、语义、错误和 wire 差异放在 adapter 能力代码，不能下放 ModelCompat 或让 Host Catalog 承担协议事实；型号目录保存型号硬约束，Host 保存账户、凭据与部署策略。这延续 DeepSeek 的边界，也遵守第二个兼容厂商 GLM 接入后细化的能力分组。[DeepSeek Responses ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)、[DeepSeek Chat ADR-0015](../../docs/adr/0015-barness-ai-deepseek-chat-completions.md)、[Provider 自动字段 ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md)

| 关注点 | Chat | Responses | Messages |
| --- | --- | --- | --- |
| 自动字段 | 开 stream usage，system 指令，max_completion_tokens；不派生 store/developer/cache retention | 仅已确认字段，禁 encrypted include/summary:auto；无状态全历史 | 只生成已确认的 cache/thinking/effort 形状；不附原厂 display/betas/eager |
| 思考 | adaptive/disabled 与 effort 分开；reasoning_split 格式单独管理 | effort 编码与型号强制思考分开 | adaptive/disabled 与 output_config.effort 分开；不使用预算型 enabled |
| 工具 | 默认 auto，强制策略暂不承诺 | 仅 auto / none；禁并行请求明确拒绝 | 仅 auto / none；any / 具名强制不静默降级 |
| 缓存 | 自动缓存，不派生未文档化的 key/TTL | 可用隔离后的 prompt_cache_key，不自动加 retention/亲和头 | ephemeral 的 5m；1h 未确认，不能原样发送 |
| 服务层级 | standard 优先；priority 需完整计费实现 | standard 优先；不能复用原厂倍率 | 同左 |

表中“暂不承诺”是接入产品子集，不等于官方已证明拒绝。显式请求无法保持语义时，应在网络前返回可诊断错误；尤其停止序列、强制工具、禁并行、关闭不可关闭的思考，不能先删除字段再称成功。最终校验须覆盖 SamplingParams 与可信 payload 回调产生的最终请求，保持既有授权、目标模型和原生引用验证。

### 7.3 Chat 的关键实现点

1. **不要继承 DeepSeek 的思考关闭默认。** M3 Chat 默认开，M3.1 强制开，M2.x 关闭无效；simple/full 两条入口都应满足硬约束。M3 effort 不等于深度可调，不能呈现不存在的控制效果。
2. **推理保存按实际 wire 形状。** M3.1 reasoning_content 可复用现有增量读取；[chat_reasoning.go](../../ai/chat_reasoning.go) 也已处理 reasoning_details 数组的 reasoning.text/summary/encrypted，M3 若与该形状吻合直接复用。M3 的实际数组字段与 content 内 think 标签仍须有 fixture 与 Live 证据；保留完整结构，不将展示文本冒充原生回放。
3. **继续完整工具往返。** [chatAssistant](../../ai/chat_request.go) 与 replay 保留可信来源，不继承 DeepSeek“每条 assistant 补空 reasoning_content”的特例；确认 MiniMax 要求的 assistant 形状后再启用。
4. **正确吃完流末用量。** 发送 include_usage=true，验证 finish chunk、usage-only chunk 与 DONE 的组合，不因先看到 finish 就丢用量；usage 的实际明细优先，缺失保留 unreported/partial。

上述行为应沿现有 [Chat adapter](../../ai/chat.go)、[request](../../ai/chat_request.go)、[usage](../../ai/chat_usage.go) 扩展，而不是建立 MiniMax 专属重复 reader。

### 7.4 Responses 的关键实现点

1. **已有解析可以复用。** [responses.go](../../ai/responses.go) 已处理 reasoning_text 增量及 itemDone 的 content 数组，MiniMax 输出示例也为数组，不需再造 reasoning reader；重点验收真实 SSE 名称、权威终态、incomplete/failed 和工具参数分片。
2. **关闭未确认的派生字段。** [ResponsesOptions](../../ai/responses_options.go) 的 summary:auto、encrypted include、缓存保留期与亲和头必须逐项受能力控制；不能仅禁一个 encrypted 字段就认为已兼容。store 可省略，服务端固定 false 的事实仍应写入 differences。
3. **推理回放单列验收。** 输出 content 与输入 summary 的 schema 不对称，需要先记录 fixture/Live 失败场景，确定原生条目能否原样回传；不偷偷转摘要、也不借 previous_response_id 绕过库的历史职责。
4. **缓存键继续租户隔离。** prompt_cache_key 可复用 [cacheScope.key](../../ai/cache_key.go) 的账户、租户和 session 派生，不发送原始 session 标识；仅支持路由键不能连带开启 retention 或亲和头。
5. **服务层级计费单独处理。** [responses_usage.go](../../ai/responses_usage.go) 的原厂 priority 倍率不能直接用于 MiniMax；需要型号长度档位与 1.5 倍规则，未实现前只支持 standard，显式 priority 应拒绝并解释。

首期使用完整历史、无状态、函数工具子集。Retrieve/delete/cancel、OpenAI 托管工具与自动 token 估算不纳入当前 Chat 调用生命周期；取消代表本地 context/连接取消，不能承诺远端没有继续生成或扣费。

### 7.5 Messages 的关键实现点

1. **拆开 nativeThinking 现有组合。** 当前能力把原厂 adaptive/effort 与 display/beta 绑定，MiniMax 支持前者不能自动开启后者；预算型 thinking 也不应从共享逻辑漏出。已有五档 effort enum 可复用。
2. **M3.1 的 mandatory thinking 进入所有入口。** simple 没指定 Reasoning 时不能一律设置 disabled；M3 Messages 默认关闭则按其自身目录规则处理。
3. **有签名回放继续保持严格。** 官方 signed 样例支持复用现有路径；缺签名作为 Live 项，若将来确需支持，只按 MiniMax 同模型可信规则局部扩展，不放宽所有 Provider。[现有 replay](../../ai/replay.go)、[assistant blocks](../../ai/anthropic_request.go)
4. **缓存先实现明确子集。** 首期使用自动缓存；主动缓存的 ephemeral 5m 与最多 4 个断点已有契约，但须补齐目标型号的支持、用量与写入价格证据后再默认启用。长 TTL 不发送，不将 long 请求悄悄解释成已保证一小时；M3/M3.1 主动缓存计价不足时，不虚构费率或保证命中。
5. **鉴权二选一，核验最终请求。** 可复用 SDK 的 x-api-key，或按官方推荐 Bearer；避免无意同时放两把不同 Key。SDK 自动 version/header、路径、beta 与重试均以捕获请求检查，重试仍只由 Binding 控制。

复用 [Anthropic adapter](../../ai/anthropic.go) 的 SSE 生命周期、内容块、资源限制与用量聚合，并以 MiniMax 自身响应确认签名、工具、错误和缓存语义，不能拿 Claude 的通过代替 MiniMax 通过。

### 7.6 计费、观测与回滚

当前 [CostRates / UsageCost](../../ai/usage.go) 使用美元；中国人民币价格应有注明日期、来源、汇率的固定快照，不能直接填成美元。M3 长上下文档位、priority 1.5 倍、主动写缓存均需覆盖；若现有费用模型不能表达，先明确成本未知/未支持，不能假填 0。M Plan token usage 与套餐最终扣减分开展示。

当前只有 ResponsesOptions 直接提供 ServiceTier 与服务层级计价。Chat 的 SamplingParams 和 Messages 的 payload 回调虽可能带入 service_tier，现有两条用量路径不会据此调整价格；只修 Responses 的倍率仍会漏计另外两种协议。首期建议三协议统一只支持 standard，并在全部回调之后验证最终请求、显式拒绝 priority；后续同时补齐三协议的实际层级归属和 1.5 倍计价。[Responses 定价](../../ai/responses_usage.go)、[Chat 计价](../../ai/chat_usage.go)、[Messages 计价](../../ai/anthropic_stream.go)

用量归一化按各协议证据分别执行：Messages 主动缓存文档明确 input 是未缓存部分、总输入要加 read/write，与当前 [anthropic_stream.go](../../ai/anthropic_stream.go) 的加总方向一致；Chat/Responses 的 cached 是输入明细子集，不重复加到总输入；reasoning 是输出明细，不重复加入输出。自动缓存和每个实际型号的 wire 仍纳入 Live，发现与文档不同应记录冲突并按 Provider 归一化，不能以现有 reader 的公式当作厂商证据。[主动缓存用量](https://platform.minimax.cn/docs/api-reference/anthropic-api-compatible-cache)、[Chat 缓存样例](https://platform.minimax.cn/docs/api-reference/text-prompt-caching)、[Responses Usage](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json)

每次调用记录实际模型、协议、地域/端点类别、凭据用途别名、HTTP 状态、原始停止原因、trace_id 与协议 request_id（有则记录）；请求体、思考文本与密钥不进入日志。对 M2.x 套餐别名路由，记录响应 model 与配置 model 的差异，便于定位模型迁移。[模型路由公告](https://platform.minimax.cn/docs/guides/models-intro)、[错误码](https://platform.minimax.cn/docs/api-reference/errorcode)

重试、取消、资源上限沿现有 Binding 约定；流已经开始后不自动重放；任何自动换模型、换协议、跨地域或把套餐切到按量都必须排除在默认失败恢复之外。回滚用关闭 MiniMax 目录/Binding 即可，内部重构删除旧字段，不留永久双路径。

### 7.7 实施与验收顺序

1. **先写失败场景，再实现。** 覆盖错误基址/鉴权、跨租户回放、型号思考默认差异、M3.1 off/none、未知 minimal、强制工具/禁并行、stop 被忽略、payload 回调绕过、流缺终态、取消和资源限额。遵守仓库测试原则，不在实现后补镜像单测。
2. **Provider/型号能力与三绑定先行。** 按量 M3、订阅 M3.1 明确分开；先形成官方契约驱动的最小子集。Messages 可作为优先落地路径，Chat 与 Responses 并列支持目标，不做跨协议降级。
3. **离线公共入口 E2E 留可重放证据包。** 每协议记录合成请求、SSE fixture、事件序列、Result、Attempt/Usage、错误和资源断言；不只有文本首轮，要有工具回传、思考原生状态恢复、截断、缓存与真实失败。
4. **Live Smoke 独立执行并保存自身证据。** 后续实施时，三协议分组合，有界输出、固定型号/区域/Key 类别，录制脱敏证据；补齐本报告的字段冲突、流末 usage、SSE 终态、工具增量、签名、回放与 request id。本次研究没有读取 Key 或发付费请求。
5. **差分基线准确登记。** 冻结 pi 的 MiniMax Provider 已有 Messages，但当前 [oracle runner](../../ai/internal/testkit/pioracle/node/runner.mjs) 的 Messages catalog 只有 Anthropic；需增加 MiniMax catalog 装配并保持冻结包本身不变，才能做该组合差分。只对冻结基线已有型号与行为作差分；M3.1 不在该目录中，作为型号扩展验收，不能假造新基线。Chat / Responses 在该冻结 Provider 数据中未列路由，按扩展组合登记自身 fixture 与 Live 证明。不能因为上游数据存在或共享 reader 测试通过就标 MiniMax 已 PASS。
6. **发布支持矩阵单列三条。** 每种型号/协议/区域/账户用途独立标 PASS/FAIL/NOT_RUN/UNSUPPORTED；同步 contract、differences、目录来源和 ADR。没有真实协议验收时维持 NOT_RUN，不把研究报告等同于上线支持。

## 8. 待验证问题与证据边界

| 问题 | 已有证据 | 还缺什么 |
| --- | --- | --- |
| Responses 是否存在 | 官方路径、schema、示例均明确 | 不需要再以“未确认”表述；还缺 barness Live |
| Responses 流末事件 | 官方声明 SSE | 精确事件序列、终态、usage、错误帧 |
| Responses 推理回放 | 输出 content / 输入 summary schema | 同模型工具往返真实可接受形状 |
| M3 Chat 分离推理 | reasoning_content 与 reasoning_details 两组官方资料 | 每型号/端点真实 wire、原样 replay |
| 工具参数逐段输出 | 三协议均有工具支持 | Chat/Messages delta 与 Responses 事件细节 |
| Messages signature | 官方 signed 输出样例、原样回传说明 | 流式签名、各型号有无缺签名情况 |
| 工具强制与禁并行 | Responses/Messages 只列 auto/none；Responses禁并行被忽略 | Chat 的 tool_choice 实际处理，其他策略不承诺 |
| 采样零值 | Responses 文字范围和 minimum 冲突 | 零值实际接受、效果与错误 |
| 新型号主动缓存 | 一般字段契约、M2.x 主动缓存与M3自动缓存例子 | M3/M3.1 主动缓存支持、用量与写入价 |
| M3.1 普通按量 | 官方目前仅 M Plan/MiniMax Code | 开放公告前不承诺可用 |

## 9. 官方资料快照与复核

主要结论均在邻近位置引用原始官方页面；网页随时间变动，下列是 **2026-10-18** 实际下载版本的 SHA-256。中国站 HTML 与 Markdown 可能呈现不同折叠层级，本报告以下载的公开 schema 交叉核对。临时资料位于 `/tmp/barness-minimax-docs`，不作为永久仓库附件，也不包含凭据。

| 资料 | SHA-256 |
| --- | --- |
| [官方索引](https://platform.minimax.cn/docs/llms.txt) | `a552de02bd2e68aefbae31f0556046fe5d5d67e5e2c3c3b13ba4d59573e410a3` |
| [Responses OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-responses.json) | `045a5d322183cc5114706c1cfeb249cc5a5576bc30947ea68aab8e14a4169c30` |
| [Chat OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-openai.json) | `9b01666f1df386776cb4d9f49a8dd2832f94ab7bbb446d1c72fa3e76af4348fb` |
| [Messages OpenAPI](https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json) | `57111d5dddd2ada7f481e5c373d88c745eba0b852b60bf1ca20096c61f28b77c` |
| [M3 工具指南 Markdown](https://platform.minimax.cn/docs/guides/text-m3-function-call.md) | `42193c24a679a9bc170cf1ee30b15a1ff5adae9a6f5c1a7cd222470f1fecc8a7` |
| [OpenAI SDK Markdown](https://platform.minimax.cn/docs/api-reference/text-openai-api.md) | `2b1b748877fe8f1c5712f9073dfaf3f4b83dfde89bf97cd7219300216fb3f552` |
| [M Plan FAQ Markdown](https://platform.minimax.cn/docs/m-plan/faq.md) | `4c2948a181ba4a52439a44b539df872d80dde6d55c74bbcd7dd4b02768961a09` |

复核可使用 `curl --fail --location --silent --show-error <上表URL> | shasum -a 256`。hash 变化表示文档内容变化，不能单独证明线上行为改变。此 Markdown 为本次研究交付物；没有修改 barness-ai 代码，也没有执行模型付费测试。
