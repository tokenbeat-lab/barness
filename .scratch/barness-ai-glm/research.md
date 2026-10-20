# GLM 三种兼容协议研究与 barness-ai 接入建议

研究日期：2026-10-17（Asia/Shanghai）。范围：智谱中国站 `docs.bigmodel.cn` 文档中的 OpenAI Responses、OpenAI Chat Completions、Anthropic Messages，以及 barness-ai 对这些协议的接入。GLM 自有协议、模型托管、Z Agent Builder、图像生成等不在范围。

方法：实际浏览官方兼容指南，并下载官方 Markdown、[全站索引](https://docs.bigmodel.cn/llms.txt)、[Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json) 与 [Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json) 交叉核对。部分 HTML 页面在浏览工具中超时，其官方 `.md` 地址可正常获取。未使用真实 API Key 发起收费请求；“官方已声明”“文档未说明”“本地验证”“建议”分别表述，不把兼容宣传等同于逐字段实测。

## 1. 核心结论与端点

**GLM 已有三种协议的明确官方接入说明，包括 Responses；无需把 Responses 降级转换成 Chat。** Responses 是独立端点与对象/事件模型；Chat 和 Messages 也各有兼容指南。这里的 Completions 指 `client.chat.completions`、即 `/chat/completions`，不是旧式文本 `/completions`；本次检查的官方兼容指南与 OpenAPI 未提供旧式文本 Completions 的契约，不能宣称支持它。[Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[OpenAI 兼容指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[Claude 兼容指南](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)、[Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json)

| API 协议 | 官方 SDK base URL | 生成请求完整地址 | 鉴权证据 | 官方 SDK 示例 |
| --- | --- | --- | --- | --- |
| OpenAI Chat Completions，标准 API | `https://open.bigmodel.cn/api/paas/v4` | `POST https://open.bigmodel.cn/api/paas/v4/chat/completions` | `Authorization: Bearer <API_KEY>` | OpenAI Python `client.chat.completions.create`。[指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json) |
| OpenAI Chat Completions，Coding 端点 | `https://open.bigmodel.cn/api/coding/paas/v4` | `POST https://open.bigmodel.cn/api/coding/paas/v4/chat/completions` | 官方按同一 OpenAI Chat 协议配置；计费须结合 Key 类型 | 官方 Coding Plan 端点表。[快速开始](https://docs.bigmodel.cn/cn/coding-plan/quick-start)、[Key 公告](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management) |
| OpenAI Responses，通用 | `https://open.bigmodel.cn/api/v1` | `POST https://open.bigmodel.cn/api/v1/responses` | `Authorization: Bearer <API_KEY>` | OpenAI Python、TypeScript `client.responses.create`。[指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction) |
| Anthropic Messages，通用 | `https://open.bigmodel.cn/api/anthropic` | `POST https://open.bigmodel.cn/api/anthropic/v1/messages` | cURL 明确为 `x-api-key: <API_KEY>` | Anthropic Python、TypeScript、Java `messages.create`。[指南](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction) |

路径不是同一个 OpenAI base URL 下任选资源：Responses 不使用 `/api/paas/v4`，Messages 的 SDK base URL 不含末尾 `/v1`。协议兼容不代表所有 OpenAI/Anthropic 原生功能都被承诺；三份指南都应结合字段参考判断。[Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[OpenAI 指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[Claude 指南](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)

## 2. API Key、计费与 Coding Plan 的新规则

这是直接影响默认配置的近期变更：官方于 **2026-10-16** 发布 Key 分类公告。新建标准 API Key、个人套餐 Key、团队套餐 Key 分开管理；在其支持的端点中，标准 Key 使用标准 API 额度，套餐 Key 使用对应套餐额度，修改 URL 不会把新 Key 改为另一计费用途。不能再笼统以“URL 中是否含 coding”判断费用。[API Key 分类管理说明](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)

历史 Key 的处理不同：历史兼容 Key 的额度取决于调用端点和账号是否曾拥有 Coding Plan。公告明确把 Messages 与 Responses 列为通用地址；曾有套餐的历史兼容 Key 调用它们会走套餐，套餐过期会中断服务，不自动退回余额。标准 Chat 端点则走标准 API。需要按量时可换标准 Key，公告说原通用地址可以保留。[API Key 分类管理说明](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)

Coding Plan 接入页也明确列出 Messages、Coding Chat、Responses 三个 base URL，但使用资格与协议能力是两件事：套餐仅用于官方指定工具/产品环境，指定工具名单包含 Pi Coding Agent；本次列表未列 barness，不能从“复刻 pi”推导获得同等套餐资格。这会影响 barness 本地与云端的默认产品方案。[Coding 快速开始](https://docs.bigmodel.cn/cn/coding-plan/quick-start)、[适用工具](https://docs.bigmodel.cn/cn/coding-plan/tool/others)、[使用须知](https://docs.bigmodel.cn/cn/coding-plan/usage-notes)

存在文档时差：GLM-5.3 模型页仍说“有过 Coding Plan（含过期）的账号，暂只能通过 Chat 调用模型 API”；10 月 16 日公告则提供标准 Key 在原通用地址继续按量的路径。最新公告能解释凭据分类后的变化，但本研究没有真实账号验证，不能保证所有历史账号已完成迁移。[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)、[Key 公告](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)

## 3. OpenAI Chat Completions 的兼容边界

### 3.1 请求与响应

| 主题 | 官方证据所支持的行为 |
| --- | --- |
| 基础消息 | 官方 OpenAI SDK 示例使用 `system`、`user`、`assistant`，工具回传采用 `role: tool` 与 `tool_call_id`。不能把 Responses 的 `developer` 支持外推给 Chat。[OpenAI 指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全) |
| 输出预算 | 公开字段是 `max_tokens`，GLM-5.3/5.2 上限为 131072；不是把 OpenAI 默认 `max_completion_tokens` 原样视为已获支持。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全) |
| 普通流式 | `stream=true` 使用 SSE，`choices[].delta.content` 为正文；Chat 文档明确以 `data: [DONE]` 结束。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全) |
| 思考输出 | 兼容 OpenAI SDK 的官方示例从 `delta.reasoning_content` 读取推理文本，从 `delta.content` 读取可见回答。`reasoning_content` 是该兼容路径中的扩展字段，不是 OpenAI 标准 Chat 推理字段。[OpenAI 指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction) |
| 函数调用 | `tools` 使用 OpenAI Chat 的嵌套 `function: {name, description, parameters}`；返回 `tool_calls[].function.arguments` 为 JSON 字符串。公开 `tool_choice` enum 只有 `auto`，最多 128 个函数；`required`、具名强制函数不能按 OpenAI 原生功能默认承诺。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全) |
| 工具参数增量 | `tool_stream` 默认 false；明确要求同时 `stream=true`、`tool_stream=true` 才启用工具参数逐段输出。平台不替应用执行自定义函数，应用收到完整参数后执行并回传结果。[工具流式输出](https://docs.bigmodel.cn/cn/guide/capabilities/stream-tool) |
| 结构化输出 | 公开 `response_format.type` 为 `text` / `json_object`；未给出 OpenAI `json_schema` 和函数 `strict` 的同等保证。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全)、[Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json) |
| 停止词 | `stop` 在参考 schema 中是字符串数组，最多 4 项。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全) |
| 采样范围 | 详细参考对 `temperature` 写 `[0,1]`、`top_p` 写 `[0.01,1]`，均限两位小数；OpenAI 兼容指南却提示 temperature 为 `(0,1)` 且不适用零值关闭采样。`temperature=0` 在兼容 SDK 路径的实际行为应单列验证，不能仅按一页推定。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全)、[OpenAI 指南](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction) |
| 未文档化的原生字段 | 本次 Chat schema 未列 `store`、`stream_options.include_usage`、`parallel_tool_calls`、`service_tier`、`prompt_cache_key`、`prompt_cache_retention`。这表示未获逐字段契约证据，不表示已证实全部拒绝或忽略。[Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json) |

### 3.2 思考开关、强度与历史回放

必须按型号区分，不能直接复制 DeepSeek 的“不指定等级就发送 disabled”规则：

| 模型范围 | 开关与强度的官方限定 |
| --- | --- |
| GLM-5.3 / GLM-5.3-Flash 系列 | `thinking.type` 只能为 `enabled`；禁用会报错。`reasoning_effort` 为 `low` / `high` / `max`，默认 `max`。[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)、[深度思考](https://docs.bigmodel.cn/cn/guide/capabilities/thinking)、[Chat schema](https://docs.bigmodel.cn/openapi/openapi.json) |
| GLM-5.2 | Chat schema 明确 `none` / `minimal` 放弃思考，`low` / `medium` 映射 `high`，`xhigh` 映射 `max`。不要把这组映射应用到 5.3 的 Chat 型号。[Chat schema](https://docs.bigmodel.cn/openapi/openapi.json) |
| GLM-4.5 起的交错思考 | 工具调用之间、工具结果后可继续思考；使用“交错思考 + 工具”必须保留并在回传工具结果时携带原 `reasoning_content`。[思考模式](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode) |
| 保留式思考 | Coding Plan 端点默认开启，标准端点默认关闭；标准端点通过 `thinking.clear_thinking=false` 启用。要求历史推理内容完整、未修改、按原顺序返回；不能裁剪、改写或重排。[思考模式](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode)、[Chat schema](https://docs.bigmodel.cn/openapi/openapi.json) |

上述 replay 证据来自 **Chat** 示例，不能据此写出 Anthropic `thinking` 块或 Responses `reasoning` 条目的相同字段与签名规则。保留历史推理也不等于让模型本轮产生思考；`clear_thinking` 仅控制跨轮历史。[思考模式](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode)、[Chat schema](https://docs.bigmodel.cn/openapi/openapi.json)

### 3.3 用量、缓存与终态

Chat 的公开用量为 `prompt_tokens`、`completion_tokens`、`total_tokens` 与 `prompt_tokens_details.cached_tokens`；缓存示例还出现 `completion_tokens_details.reasoning_tokens`。隐式缓存自动识别重复前缀，无须 Anthropic 的显式缓存控制；官方建议重复前缀足够长（示例建议 500 Token 以上），不是稳定保证的命中阈值。缓存折扣说明只适用于标准 API 计费，不应机械套给资源包和 Coding Plan。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全)、[上下文缓存](https://docs.bigmodel.cn/cn/guide/capabilities/cache)

`finish_reason` 除 `stop`、`tool_calls`、`length` 外，还说明 `sensitive`、`network_error`、`model_context_window_exceeded`。后者在部分 enum 列表中漏列但文字说明明确存在；这些不应无条件映射成正常停止。错误码页特别说明 SSE 推理中途异常使用 `finish_reason`，不是一定得到非 2xx 错误 JSON。[Chat 参考](https://docs.bigmodel.cn/api-reference/模型-api/对话补全)、[错误码](https://docs.bigmodel.cn/cn/api/api-code)

## 4. OpenAI Responses 的兼容边界

### 4.1 已明确支持的契约

官方不仅有兼容声明，还公开了独立 [Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json)；本次下载的 `paths` 为 `/v1/responses`、`/v1/responses/{response_id}`、`/v1/responses/{response_id}/input_items`，覆盖创建、查询、删除、列举输入项。[Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)

| 主题 | 官方已说明的行为 |
| --- | --- |
| SDK 与输入模型 | 官方示例直接使用 OpenAI Python、TypeScript `responses.create`。`input` 可为字符串，或 `message` / `function_call` / `function_call_output` / `reasoning` 列表；消息角色支持 user/system/assistant/developer。[指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response) |
| 保存与续接 | 默认 `store=false`；显式 `store=true` 后可检索、删除、列举输入项并用 `previous_response_id` 服务端续接，ID 有效期 7 天。当前没有 cancel 接口。[指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction) |
| 生成预算 | `max_output_tokens` 包含回答和思维链，默认 65536，最大 131072；支持 `instructions`、`temperature`、`top_p`、`stop`。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response) |
| 工具 | 公开 `function` / `namespace` / `custom` / `web_search`。Function 采用 Responses 的顶层 `name`、`parameters`；结果回传 `function_call_output`、`call_id`、`output`。`tool_choice` enum 只有 `none` / `auto`。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response) |
| 缓存 | 明确提供 `prompt_cache_key`，用途是集群路由以提高缓存命中；并未提供 `prompt_cache_retention` 的同等契约。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response)、[OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json) |
| 文本格式 | `text.format.type` 仅列 `text` / `json_object`；函数工具 schema 未列 `strict`。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response)、[OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json) |
| 用量 | `input_tokens`、`input_tokens_details.cached_tokens`、`output_tokens`、`output_tokens_details.reasoning_tokens`；schema 未列 `total_tokens`。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response) |
| 错误与非完整结果 | Response 有 `status`、`error.{code,message}`、`incomplete_details.reason`，成功不能只看 HTTP 200。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response) |
| 客户端超时 | 指南建议客户端 timeout 7200 秒。它是厂商示例建议，不能等同 barness 宿主必须采用的资源策略。[指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction) |

Responses 的公开 `reasoning.effort` 默认 `max`，文档写 `none`/`minimal` 放弃思考、`low`/`medium`→`high`、`xhigh`→`max`；没有像 Chat schema 那样把映射限定 GLM-5.2。它与 5.3 模型页“始终思考、low/high/max 三档”存在协议层说明差异；**不能用 Chat 的模型表替 Responses 声称 low 一定保留 low，也不能宣称 Responses 的 none 在 5.3 已实测可关闭。** 该组合需要真实服务专项验证。[Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json)、[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)

### 4.2 SSE 终态与形状差异

Responses SSE **不发送 `[DONE]`**，终态看 `response.completed`、`response.failed`、`response.incomplete` 或 `error`。公开事件包括正文 `response.output_text.delta/done`、工具参数 `response.function_call_arguments.delta/done`、推理正文 `response.reasoning_text.delta/done`、output item/content part 事件与搜索事件；事件带 `sequence_number`。不要只监听 OpenAI 的推理摘要事件名。[创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response)、[Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json)

对 SDK/DTO 复用最重要的公开形状差异如下，不能仅因官方 Python 示例可运行就假定当前 Go SDK 的类型完全匹配：

| schema | GLM 官方定义 | 对复用的约束 |
| --- | --- | --- |
| `OutputReasoning.content` | **单个** `ReasoningTextPart` 对象 | 不能默认按内容数组解码；推理输出与输入回放均应核对。 |
| `OutputMessage.content` | `OutputTextPart` 单对象 **或** 数组 | 两种形状都应覆盖。 |
| `InputReasoning` | 必填 `type: reasoning` 和单对象 `content` | 官方未列 `encrypted_content`、`summary`；不能自动按 OpenAI 加密推理条目回放。 |
| `InputMessage.content` | 字符串、内容对象或内容数组 | assistant 回放可用 `output_text`。 |
| `InputContentPart` 文件 | `file_data` 或 `file_url` | 未列 OpenAI 账户文件 `file_id`；不能从协议名称推定该资源引用可用。 |

来源为 [Responses OpenAPI 中的 OutputReasoning / OutputMessage / InputReasoning / InputMessage / InputContentPart](https://docs.bigmodel.cn/openapi/openapi-responses.json) 与 [创建 Response](https://docs.bigmodel.cn/api-reference/response/创建-response)。这些是 schema 事实；服务实际返回对象或数组的频率尚未实测。

公开 schema 未列 `reasoning.summary`、`include`、`service_tier`、`parallel_tool_calls`、`prompt_cache_retention`、`conversation`、`background`、加密推理内容，以及工具强制选择和 JSON Schema strict。**未列不是证实不支持，也不是可以默认照发。** 指南只明确当前无 cancel；其他字段逐项接受、忽略或拒绝的行为尚无证据。[Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json)

### 4.3 协议模态与模型模态分开

Responses 协议公开图片与文件输入类型，但 GLM-5.3 模型页明确 **仅文本**。Responses 指南却以 `glm-5.3` 展示图片输入，这是官方文档内部的不一致；不能据这段示例把 GLM-5.3 的目录能力标为图像输入。协议具有某种容器，不证明所有模型支持该模态。[Responses 指南](https://docs.bigmodel.cn/cn/guide/develop/responses/introduction)、[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)

## 5. Anthropic Messages 的兼容边界

### 5.1 能够直接证实的部分

官方 Claude 兼容指南明确支持现有 Anthropic SDK：替换 Key、base URL 和 GLM 型号；cURL 示例使用 `/api/anthropic/v1/messages`、`x-api-key`、`model: glm-5.3`、`max_tokens`、`messages`、`stream: true`。Python/Java 示例也是 `glm-5.3`，TypeScript 示例使用 `glm-5.2`。[Claude API 兼容](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)

官方 Claude Code 接入示例另使用 `ANTHROPIC_AUTH_TOKEN` 和同一 base URL，证明官方支持这类客户端配置；这不是可以任意把它与 SDK 的 `api_key` 头方式混用的逐字段说明。对 barness 当前 Anthropic SDK 路径，最直接证据仍是指南 cURL 的 `x-api-key`。[Claude Code 接入](https://docs.bigmodel.cn/cn/coding-plan/tool/claude)、[Claude API 兼容](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)

### 5.2 文档证据的明显缺口

本次读取的官方完整索引、兼容指南和公开 OpenAPI 没有与 Responses 同等详细的 Messages 请求/响应 schema。指南明确保留接口差异，却未列下表字段的接受/忽略/拒绝行为。因此可以说“Messages 接口受到官方支持”，**不能说 Anthropic 的全部高级功能已被 GLM 承诺**。[全站索引](https://docs.bigmodel.cn/llms.txt)、[Claude API 兼容](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)、[公开 OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json)

| 需要确认的契约 | 本次官方证据状态 |
| --- | --- |
| `system`、`tools.input_schema`、`tool_use`→`tool_result` 完整往返 | 指南仅展示基本 user 文本；GLM 模型页的工具能力和 Claude Code 支持是背景证据，不能替代 Messages wire 字段验收。 |
| `tool_choice=auto/any/tool/none`、并行工具 | 未有 GLM Messages 字段兼容表，不能沿用 Chat 仅 auto 的限制，也不能默认全部 Anthropic 取值有效。 |
| Anthropic `thinking` 参数、budget/adaptive、`output_config.effort` | 未说明 GLM Messages 的转换规则；Chat `thinking.type`、Responses `reasoning.effort` 都不能直接移植。 |
| 推理 `thinking`/`redacted_thinking` 块、`signature_delta`、历史签名回放 | 未有 GLM wire 示例或签名信任说明。需要检查流与下一轮请求，不可凭原生 Anthropic 行为构造假签名。 |
| `anthropic-beta: interleaved-thinking-*`、`eager_input_streaming` | 兼容指南未列。模型支持交错思考不等于必须发送 Anthropic beta，也不等于接受该工具字段。 |
| 显式缓存 `cache_control`、TTL、`cache_creation_input_tokens`、`cache_read_input_tokens` | GLM 的缓存指南说明的是 Chat 形状的隐式缓存，不能证明 Messages 的 Anthropic 缓存写入/读取计数或缓存控制有效。 |
| `anthropic-version`、额外 Stainless headers | GLM cURL 未列版本头，SDK 示例隐含 SDK 默认行为；文档没有明确“必须”或“禁止”，不得据省略断言不支持。 |
| SSE 事件序列、usage 位置、stop reason、流内错误 | `stream: true` 有官方示例，但未给逐事件契约；成功终态、工具 JSON 分片、思考分片、usage 缺项都需留真实录制证据。 |
| `/messages/count_tokens`、批处理、容器、文件资源、MCP、服务等级 | 未找到 GLM 对应端点或逐字段说明；不纳入本次已确认能力。 |

上表的“不明确”均是对所查官方资料范围的审计结论，不是断言 GLM 服务拒绝这些能力。依据：[Claude 兼容指南](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)、[全站索引](https://docs.bigmodel.cn/llms.txt)、[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)、[思考模式](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode)、[上下文缓存](https://docs.bigmodel.cn/cn/guide/capabilities/cache)。

## 6. 型号选择与运维事实

| 型号 | 可直接采用的官方事实 | 尚需按协议确认 |
| --- | --- | --- |
| `glm-5.3` | 文本输入/输出；1M 上下文；128K 最大输出；始终思考；模型页列出三种协议端点。 | Responses effort 映射、Messages 思考参数、实际服务 quota/Key 资格。 |
| `glm-5.2` | 模型页为文本输入/输出、1M 上下文、128K 最大输出；OpenAI thinking 示例和 Anthropic TypeScript 示例使用该型号。 | 三种协议独立的流、回放、用量与模型硬约束，不能仅据 SDK 示例视为生产验收通过。 |

来源：[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)、[GLM-5.2](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.2)、[OpenAI 兼容](https://docs.bigmodel.cn/cn/guide/develop/openai/introduction)、[Claude 兼容](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction)。这里的“1M/128K”保留官方单位，不自行猜十进制换算成精确目录整数；实现时应与实际错误边界核对。

官方错误由 HTTP 状态与业务码组成：例如 401 对应认证失败、400 对应参数/型号问题、403 对应无权访问。**429 不全是可重试限流**：1113 是欠费，1302 是用户速率限制，1305 是模型过载，1308 是用量上限，1309 是套餐到期，1311 是套餐未开放型号；仅依赖 HTTP 429 作统一重试会丢失操作含义。Responses 对象或 SSE 内错误还要按各协议读取；不能把 Chat `finish_reason` 的规则套给 Messages。[错误码](https://docs.bigmodel.cn/cn/api/api-code)、[Responses 创建参考](https://docs.bigmodel.cn/api-reference/response/创建-response)

限额主要按账户、模型、权益/套餐控制并发，且有动态平台过载保护；Coding Plan 并发随套餐等级与时段变化，没有适合硬编码的统一每分钟数字。协议切换也不能被当作绕过同账户限额的方法。[速率限制](https://docs.bigmodel.cn/cn/api/rate-limit)

## 7. barness-ai 建议方案：一个 Provider，复用三套协议 adapter

本节是**建议，尚未实施或作为 ADR 接受**。仓库核查基线为 `f5cf536`，当前冻结 pi oracle 为 `1.1.0`；不是早期 DeepSeek ADR 中的 `0.87.1`。以下设计延续现有 [Provider / API / Binding 定义](../../GLOSSARY.md)、[协议边界 ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)、[DeepSeek Responses ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)、[DeepSeek Chat ADR-0015](../../docs/adr/0015-barness-ai-deepseek-chat-completions.md)。

### 7.1 领域边界与接入形状

建议新增 `ProviderZhipu = "zhipu"`，代表本报告研究的 `open.bigmodel.cn` 服务。GLM 是模型系列；协议仍用现有三个 API 常量，不新增 `glm-*` 协议名，也不把调用伪装成 OpenAI 或 Anthropic。

| Operation × Provider × API | 建议 BindingID 示例 | Endpoint（SDK 基址） | 复用实现 |
| --- | --- | --- | --- |
| chat × zhipu × openai-responses | `glm-responses` | `https://open.bigmodel.cn/api/v1` | `responsesAdapter` |
| chat × zhipu × openai-completions | `glm-chat` | `https://open.bigmodel.cn/api/paas/v4` | `chatAdapter` |
| chat × zhipu × anthropic-messages | `glm-messages` | `https://open.bigmodel.cn/api/anthropic` | `anthropicAdapter` |

端点来自第 1–5 节官方证据；Binding 示例与 Provider ID 是设计建议。[现有 registry](../../ai/adapter.go) 已按 API 选 adapter。继续使用不可变的调用快照、租户授权、账户作用域、有限资源策略和显式重试；凭据在 Host 装配，库不读取 `ZAI_*`、`OPENAI_*`、`ANTHROPIC_*` 环境变量。

默认选择标准 API Key 与上表标准/通用端点。标准/个人套餐/团队套餐/历史兼容 Key 的分类由 Host 保存；配置界面应明确凭据用途和目标账户，不能根据 URL 或密钥字符串猜计费方式。Coding Plan 如有真实需求，先确认 barness 对应使用环境的适用资格，再使用独立 Binding；同账户、实际授权允许时才共享 CredentialRef。不要增加按失败自动改 URL、改协议、改 Key 的路径。[Binding 契约](../../ai/binding.go)、[10/16 官方 Key 公告](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management)、[套餐适用工具](https://docs.bigmodel.cn/cn/coding-plan/tool/others)

### 7.2 借鉴 DeepSeek 的能力分层，避免照抄其值

目前 [chatCompatOf](../../ai/chat_compat.go) 仅对 DeepSeek 特判，其他 Provider 使用 OpenAI 默认；[responsesCapabilitiesOf](../../ai/responses_capabilities.go) 同样如此。新增 Provider 常量和目录后直接接入，会错误沿用这些默认值。

新增 GLM 的显式 Provider 分支即可；只有出现实际不同能力时扩充小结构，不先建立可配置策略注册中心。Provider 协议能力管理端点请求形状；型号是否允许关闭思考、输入模态、输出上限等硬约束留在独立的 Model 条目。GLM 是第二个兼容厂商，适合在已有边界上去掉 DeepSeek 专名，例如将 `deepseekThinking` 拆成实际需要的 thinking 开关与 effort 编码职责；删除旧字段并迁移现有分支。[现有模型能力](../../ai/catalog_models.go)、[目录纳入原则 ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)

| 关注点 | Responses 建议 | Chat Completions 建议 | Messages 建议 |
| --- | --- | --- | --- |
| 默认生成字段 | 显式 `store:false`；只生成已获证实的核心字段 | `system` 指令、`max_tokens`；不生成 `store`、`developer`、`max_completion_tokens` | 先使用指南已展示的基础形状 |
| 缓存控制 | 可使用隔离后的 `prompt_cache_key`；保留期与亲和头分别关闭 | 先不派生 OpenAI 缓存键/保留期 | 未证实前不自动生成 `cache_control` 或 TTL |
| 思考控制 | effort 映射独立；不自动附 `summary:"auto"` 或 encrypted include | 开关、effort、不可关闭的模型分别表达 | 不套用 Claude adaptive、display、budget、managed effort 或 beta |
| 工具选择 | 允许 `none` / `auto`；其他模式请求前明确拒绝 | 文档只有 `auto`；其他模式请求前明确拒绝 | 对 `any`、强制函数、禁并行等分别真机核验 |
| 响应解析 | raw 边界适配对象/数组及 reasoning，然后进入共用 assembler | 复用文本、reasoning_content、工具参数增量解析 | 复用 SSE 生命周期；以 GLM 自身 wire 证据确认块与终态 |

上表是根据第 2–4 节文档作出的保守产品子集，**未列字段不等于官方已证明拒绝**。显式请求没有接入的能力应返回可诊断的 `invalid_request`，不能静默删除强制工具选择后仍称成功。对现有缓存提示无派生的约定可延续，但文档要说明没有缓存收益保证。能力校验应覆盖具名选项、SamplingParams 和可信 payload 回调的最终请求，继续保留原有模型、目标与原生引用授权检查。

### 7.3 Responses：重点补齐字段能力和 wire 边界

1. **拆开缓存能力**。当前 `promptCache` 同时控制 key、retention 和亲和头。GLM 文档只确认路由键；应分别控制这些行为。key 继续复用 [cacheScope.key](../../ai/cache_key.go) 的租户、账户、session 派生，不能发送原始 session 标识。
2. **增加 summary 能力判断**。[ResponsesOptions.encode](../../ai/responses_options.go) 当前在设置 effort 时自动写 `summary:auto`；GLM 的 ReasoningConfig 只有 effort，不能仅关闭 encryptedReasoning 就认为请求已适配。ServiceTier、encrypted include、缓存 TTL 等先不生成，显式请求按已接入能力拒绝。[创建 Response schema](https://docs.bigmodel.cn/api-reference/response/创建-response)
3. **修复实际可复现的 shape 风险**。官方 schema 的 reasoning `content` 为单对象，message `content` 可对象或数组；现有 SDK union 与 [itemDone](../../ai/responses.go) 以数组读取。第 8 节证明单对象会静默变成零条内容。应在 raw JSON 边界转换成内部解析形状，沿用共用流/assembler；不能用字符串替换 JSON。保留原始 reasoning 条目供 Native State 回放，避免把为解析而改成数组的内容当成 GLM 原生状态发回。
4. **无状态全历史工具往返仍作为首期模式**。即使 GLM 支持服务端保存，继续不发 `previous_response_id` / `conversation`，不增加查询/删除会话职责；保留同模型可信 reasoning、call_id 与工具结果。核对 GLM 的 InputReasoning、assistant message、函数结果是否接受现有编码；不能只验证首轮文本。[当前 history 编码](../../ai/responses_request.go)、[原生状态边界 ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)
5. **以协议终态判成功**。当前 [readResponsesStream](../../ai/responses_stream.go) 已不依赖 `[DONE]`，可复用。取消是本地 context/连接取消，不承诺远端任务已停止或没有计费。当前代码还有 OpenAI 的 16-token 输出下限，GLM schema 最小为 1；实现时按 Provider 约束处理，避免隐藏改写调用者预算。

首期工具范围是 barness 已有的函数工具。`namespace` / `custom` / `web_search` 不因 GLM 文档支持就自动承诺；以后有真实需求时再完善其领域结果、parser 与授权，不能仅靠 payload 回调透传后忽略输出。[现有 Tool 契约](../../ai/tool.go)

### 7.4 Chat：把 GLM-5.2 与 GLM-5.3 的硬约束分开

首批可评估 `glm-5.3` 与 `glm-5.2`，每款在三个协议下分别建目录条目；不根据 Chat 配置自动派生另外两份。**GLM-5.3 Chat** 先只暴露 low/high/max：off/minimal/medium/xhigh 要明确映射或拒绝，不能产生 `thinking:disabled`；simple 的夹取与 full 的直传都必须满足硬约束。GLM-5.2 再按官方语义映射开关和 effort。[Chat schema](https://docs.bigmodel.cn/api-reference/模型-api/对话补全)、[GLM-5.3](https://docs.bigmodel.cn/cn/guide/models/text/glm-5.3)、[当前档位处理](../../ai/chat_options.go)

不要把 Responses 指南的 effort 规则直接搬到 GLM-5.3 Chat。对 Responses 自身的同型号冲突，在真机确认前采用 low/high/max 的保守子集；其他档位属于待核验范围，不等同于已证明不支持。

[chatAssistant](../../ai/chat_request.go) 已能按来源字段回放 `reasoning_content`；GLM 不需要继承 DeepSeek“每条 assistant 强制补空 reasoning_content”的规则。默认按官方 `clear_thinking:true` 处理跨 turn 思考，工具往返保留同一 turn 的必要推理；只有新增明确需求且验证完整性后才开放 Preserved Thinking。它要求顺序和内容完整，不能在任意裁剪历史后仍宣称启用。[思考模式](https://docs.bigmodel.cn/cn/guide/capabilities/thinking-mode)

`stream_options.include_usage` 没有在所查 GLM Chat schema 中列出；现有 [chatBody](../../ai/chat_request.go) 却默认生成。为 GLM 先省略并验证其流末 usage 行为；文档缺字段不代表拒绝，应将带/不带该字段列入探针。工具参数增量还需按官方要求发送 `tool_stream:true`，不能仅靠已有 `stream:true` 宣称逐段工具输出已接通。[工具流式输出](https://docs.bigmodel.cn/cn/guide/capabilities/stream-tool)；[chatUsage](../../ai/chat_usage.go) 已支持 `prompt_tokens_details.cached_tokens`，缺失明细仍保持 partial/unreported，不把未报告补成已知零。

### 7.5 Messages：共享协议实现，先建立 GLM 自己的能力证据

这是 [anthropicAdapter](../../ai/anthropic.go) 第一次服务非 Anthropic Provider，建议新增小型 `anthropicCapabilitiesOf(provider)`，控制自动生成的可选厂商字段。当前 [请求构造](../../ai/anthropic_request.go) 会生成默认缓存标记、`eager_input_streaming:false`、thinking display/budget、某些 beta；这些不是 GLM 兼容指南已经承诺的全集。

初始保留官方示例确认的 `x-api-key` 与基本 messages 请求；沿用显式 SDK service、关闭 SDK 重试、独立 SSE 解码和资源限制。未证实的 cache/beta/display/eager 等不自动发出；SDK 自带 `anthropic-version`、服务路径与最终 headers 也需请求捕获核对。thinking、signature、tool_use/tool_result、usage 的 GLM 形状以自己的真实证据完善适配，不能从 Chat 的 `reasoning_content` 推导。

特别检查无 signature 的 thinking：当前 [anthropicAssistantBlocks](../../ai/anthropic_request.go) 会将无签名思考回放为 text。如果 GLM 返回无签名 thinking，应按 GLM 规则决定同模型回放，不能伪造 Claude 签名或默默改变 block 类型后承诺语义一致。持久化后的原生状态仍须 Host 重新担保，跨租户/账户不放宽。[原生状态约定](../../docs/barness-ai/contract.md)

缺少字段级公开契约意味着需要真机确认，并不意味着停在仅文本支持。三协议的最终目标均应覆盖文本流、思考、工具往返、取消、错误与用量；某项确实不存在时，用来源与本协议证据标注能力限制。

### 7.6 模型、计费与观测

- 按 `(Operation, Provider, API, ModelID)` 独立保存目录、档位、模态、上下文、输出上限、价格来源与版本。以适配器遵守型号全部硬约束为纳入前提，不把共享 adapter 的通过当成 GLM 通过。[目录 ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)
- GLM 官网价格以人民币表达，而现有 [CostRates / UsageCost](../../ai/usage.go) 明确采用美元。首期保留该外部契约：以注明日期、来源和换算率的固定价格快照转为美元估算，不引入运行时汇率抓取；没有经确认的换算快照就不发布内置价格条目。不能直接把人民币数填进美元费率。Coding Plan 的 token usage 不等于套餐最终扣费；估算与真实额度扣减分别展示，未知价格不能冒充免费。[官方定价](https://docs.bigmodel.cn/cn/guide/start/pricing)
- Responses 的字符串业务错误码与 Chat 的数字业务码分开解释；Messages 再按其实际错误形状映射。Chat 的欠费、套餐到期、用量耗尽等业务错误即使以 HTTP 429 返回，也不能按暂时限流统一重试；保留现有错误类别，并让重试判定识别这些不可恢复原因。继续保留 HTTP status、协议类型、原始 stop reason 与可取得的 request id，未确认响应头名就不照抄 DeepSeek 的 `x-ds-trace-id`。[官方错误码](https://docs.bigmodel.cn/cn/api/api-code)、[Responses 错误 schema](https://docs.bigmodel.cn/api-reference/response/创建-response)、[现有 requestIDHeaderOf](../../ai/openai_sdk.go)
- retry 仍由 Binding 固定、默认关闭；流开始后不重放。GLM 指南的 7200 秒是厂商建议，不能成为库无条件等待的默认值；超时与并发预算仍由 Host 的 ResourcePolicy 决定。[绑定重试 ADR-0006](../../docs/adr/0006-barness-ai-retry-policy-on-binding.md)、[资源时限 ADR-0008](../../docs/adr/0008-barness-ai-admission-and-time-limits.md)

### 7.7 实施顺序与验收产物

1. **先写失败场景和公共入口 E2E**：错误基址/路径、鉴权、跨租户与账户回放、非法档位、未实现工具选择、回调绕过能力校验、流缺终态、帧/队列限额、取消、用量缺失，以及对象型 content 静默丢失；然后实现。沿用仓库测试原则，不在实现后补镜像单测。
2. **按三个协议交付最小适配**：先建 Provider、Bindings、独立目录和字段能力；Chat 可作为较低改动的首条验证路径；Responses 处理 raw shape 后接入；Messages 补全厂商能力边界。实现时新增 ADR 记录采用的设计，更新 contract/differences。
3. **离线 E2E 每个协议留下可重放证据包**：合成请求、脚本化 SSE、事件序列、Result、Attempt/Usage、错误/断言与资源结果。Responses 的对象/数组、纯终态、工具往返与 reasoning 回放，Chat 的 5.3 禁关闭和 5.2 映射，Messages 的 thinking/签名/工具块分别验证。现有 OpenAI、Anthropic、DeepSeek 受影响路线跑回归。
4. **独立 Live Smoke**：新增 `zhipu-chat`、`zhipu-responses`、`zhipu-messages` 三组合；每进程只持该组合低权限 Key，固定型号与 Key 类别/区域别名，有界调用及输出预算。Messages 的未知字段、Responses 的档位冲突、全历史工具往返、缓存用量与真实 request id 必须留下证据，不能只用 SDK 示例成功替代公共 Client 验收。[现有 Live 契约](../../ai/live/doc.go)
5. **协议与账户变体分清**：如果正式支持 coding Chat 和不同 Key 类别，支持证据记录 endpoint 类别、Key 类别与账号条件；同协议的标准 Key 冒烟不能证明旧兼容 Key 的额度行为。只执行用户授权的计费测试，不在本次研究中读 Key 或发付费请求。
6. **支持矩阵与发布门禁同步**：先核查冻结 pi 是否已有对应 Provider × API 路由；有则进入该组合差分，无则按 ADR-0014 登记 extension、自身 fixture 与 live 证明，不能先验把三条都算 pi 差分。新组合分别生成 PASS/FAIL/NOT_RUN/UNSUPPORTED 和脱敏审计产物，合并自己的矩阵行，并在承诺支持时纳入相应发布门禁。

**推荐的首个交付范围**：中国区智谱一个 Provider、三份明确的协议绑定、官方文档支持且硬约束已实现的少量型号；复用三套 adapter 和现有调用生命周期。首期不加入服务端会话管理，也不加入本报告范围外的 GLM 自有协议。

## 8. 已执行的本地 SDK shape 探针（可复现）

本次只做文档核查和本地解析实验，**未调用智谱模型 API**。探针运行于当前仓库锁定的 `github.com/openai/openai-go/v3 v3.66.0`，不需要 API Key。SDK 的 [ResponseOutputItemUnion.Content](https://github.com/openai/openai-go/blob/v3.66.0/responses/response.go) 是数组；输入按 GLM 创建 Response 文档构造，不冒充抓取到的线上响应。

在仓库根目录执行下面代码。临时文件仅用于实验，不修改 barness-ai 实现：

```sh
cat > /tmp/barness-glm-responses-shape-probe.go <<'GO'
package main
import (
    "encoding/json"
    "fmt"
    "github.com/openai/openai-go/v3/responses"
)
func main() {
    for _, s := range []struct{ name, raw string }{
        {"reasoning-object", `{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"rs_probe","content":{"type":"reasoning_text","text":"thinking"}}}`},
        {"reasoning-array", `{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"rs_probe","content":[{"type":"reasoning_text","text":"thinking"}]}}`},
        {"message-object", `{"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"msg_probe","content":{"type":"output_text","text":"answer"}}}`},
    } {
        var e responses.ResponseStreamEventUnion
        err := json.Unmarshal([]byte(s.raw), &e)
        fmt.Printf("%s: error=%v type=%s content_count=%d\n", s.name, err, e.Item.Type, len(e.Item.Content))
    }
}
GO
go run /tmp/barness-glm-responses-shape-probe.go
```

实际输出：

```text
reasoning-object: error=<nil> type=reasoning content_count=0
reasoning-array: error=<nil> type=reasoning content_count=1
message-object: error=<nil> type=message content_count=0
```

结论只限于本地 SDK：按官方允许的单对象形状输入时，解析不报错但内容数组为空；因此当前 adapter 存在丢失推理或最终正文的路径。线上服务实际选择哪种形状、是否始终补全 delta，仍需对应 Live Smoke 证明。正式验收应同时覆盖“只有权威 done 内容”和“delta 后 done 校正”两类，防止终帧覆盖已流出的正文。

## 9. 可复核的官方资料入口

主要资料均在正文逐条引用。以下是本次下载版本的复核入口与关键 SHA-256；官方文档会变动，后续 hash 不同只说明内容变化，不代表接口行为已经改变。

| 官方资料 | SHA-256（2026-10-17 下载） |
| --- | --- |
| [llms.txt](https://docs.bigmodel.cn/llms.txt) | `3a07b47f4fe3c7c8dcd536fe24a5bbadaf04cd1700e3ecec3e29eaedcd9b5080` |
| [Responses OpenAPI](https://docs.bigmodel.cn/openapi/openapi-responses.json) | `0914af477ff4c1fae0372f9c44ace23033d0bed0cc486796eaff2c2024e5f18c` |
| [Chat OpenAPI](https://docs.bigmodel.cn/openapi/openapi.json) | `e1867cd1a7e31b15cc9088b03d8c1361e275153a1334b9e9e3bea95f4c623dd9` |
| [Claude 兼容指南 Markdown](https://docs.bigmodel.cn/cn/guide/develop/claude/introduction.md) | `becda63d7e3a6e71b4155b2a966794ea84a33dba67beebcaa6abafb823cd2cdc` |
| [2026-10-16 Key 公告 Markdown](https://docs.bigmodel.cn/cn/coding-plan/notice/api-key-management.md) | `3ccf060b12bb79de363d620bfbc17bbc463612765a77ace8a3d74639f4dd1a04` |

复核示例：`curl --fail --location --silent --show-error https://docs.bigmodel.cn/openapi/openapi-responses.json | shasum -a 256`。本研究文件是交付物；临时下载位于 `/tmp`，不属于长期仓库证据，也没有保存任何凭据。
