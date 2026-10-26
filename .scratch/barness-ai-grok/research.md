# Grok 的 Responses / Messages 兼容协议研究与 barness-ai 接入建议

研究日期：2026-10-26（Asia/Shanghai）。范围：xAI 官方 API 对 OpenAI Responses 与 Anthropic Messages 的支持，以及 barness-ai 的接入方向。排除 Chat Completions 和 xAI 自有协议；官方页面混有其他协议示例时，仅取可明确归属本范围的事实。

证据等级：**文档确认**表示官方文字、示例或页面所带 schema 已确认；**设计建议**表示本报告的工程判断；**待确认**表示尚无充分公开契约或尚需真实请求证据。本次未调用付费推理 API，不把文档兼容声明等同于已通过 Live Smoke。

## 1. 结论与直接影响

1. **Responses 是当前正式主路径。** 官方把它作为文本、推理和工具的主要接口，并直接展示 OpenAI Python / JavaScript SDK 调用 `https://api.x.ai/v1/responses`。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R2](https://docs.x.ai/developers/model-capabilities/text/generate-text)
2. **Messages 已被官方完整弃用。** 当前 `Legacy & Deprecated` 页面在 `/v1/messages` 前明确标注 “The Anthropic SDK compatibility is fully deprecated”，同时仍列出端点和 schema。可确认弃用；不能据此声称端点已经关闭或必然返回 404。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy)
3. **不能靠更换 Base URL 完成生产接入。** 存储默认值、推理状态、输出预算、usage、模型能力及静默忽略字段都有 xAI 自己的语义，需要独立 Provider 协议能力配置。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning)[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)
4. **优先使用 `store:false` 和完整 Transcript 回放。** Grok 4.7 默认返回加密推理，但客户端仍须原样带回下一轮；`store:false` 不等于团队级 Zero Data Retention。[R2](https://docs.x.ai/developers/model-capabilities/text/generate-text)[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning)[R6](https://docs.x.ai/developers/faq/security)
5. **成本不能直接按 OpenAI 假设处理。** 官方 Responses 示例存在 `output_tokens` 与 `reasoning_tokens` 分列相加的迹象；还存在长上下文价格、priority 倍率与区域溢价，均需明确归属。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R7](https://docs.x.ai/developers/pricing)[R8](https://docs.x.ai/developers/cost-tracking)

**建议的支持范围**：新增 `Provider=xai × API=openai-responses`；对 `xai × anthropic-messages` 保留研究结论，默认不晋升为正式支持。若 Host 有必须兼容 Messages 的真实需求，可另建受限路径，并以该组合自己的证据决定能力；不自动从 Responses 回退到已弃用端点。

## 2. 官方资料与版本取舍

从 [官方 llms.txt](https://docs.x.ai/llms.txt) 定位当前文档树，再核对 REST Reference、能力指南和定价专题。官方支持在页面 URL 后加 `.md` 读取 Markdown；REST 页的折叠 union 类型需要检查同一官方 HTML 内的 schema。[R0](https://docs.x.ai/llms.txt)

当前文档主推 `grok-4.7`，不能继续把旧的 Grok 3 / Grok 4 文档印象当作现行模型能力。Grok 4.7 专页只在 API 栏列 Responses；它没有承诺新的旗舰模型仍完整支持 Messages。[R9](https://docs.x.ai/developers/grok-4-7)

文档存在以下不一致，接入时不得静默任选其一：

| 主题 | 官方材料的差异 | 本报告取舍 |
| --- | --- | --- |
| 工具数量 | Responses schema 写 1000；Function Calling 指南的表格写 350 | 记录冲突；Host 首期采取更小上限，不宣称已证明 1000。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R10](https://docs.x.ai/developers/tools/function-calling) |
| 指定 function | Function Calling 通用表使用嵌套 `function.name`；Responses schema 使用平铺 `type/name` | 按 Responses schema 和现有协议编码，指定工具形状仍须独立确认。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R10](https://docs.x.ai/developers/tools/function-calling) |
| 结构化输出参数 | 指南概述写 `response_format`；Responses 示例和 schema 使用 `text.format` | 本研究采用 Responses 的 `text.format`，不把另一协议字段移植进来。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R11](https://docs.x.ai/developers/model-capabilities/text/structured-outputs) |
| SSE 示例 | 通用 Streaming 页混有另一协议；Responses 最佳实践给出 `response.*` 事件 | 以 Responses 专属示例为依据，不复制混合页面的 chunks。[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)[R12](https://docs.x.ai/developers/model-capabilities/text/streaming) |
| `service_tier:fast` | REST enum 包含 `fast`；定价页明确 Grok 4.7 Fast 不向公共 API 开放 | 公共 API 首期禁止 `fast`，不能仅由 enum 放行。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R7](https://docs.x.ai/developers/pricing) |
| 图片类型 | Models 页列 JPEG/PNG；Messages image schema 还列 WebP | 按协议分别判断；未验证前以交集 JPEG/PNG 为受限范围。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy)[R13](https://docs.x.ai/developers/models) |

## 3. Responses：传输与调用契约

### 3.1 端点、认证、SDK

| 项目 | 文档确认 |
| --- | --- |
| REST 服务 | `https://api.x.ai`；OpenAI SDK 的 Base URL 为 `https://api.x.ai/v1`。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R14](https://docs.x.ai/developers/rest-api-reference/inference) |
| 创建 | `POST /v1/responses`，JSON 请求体。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses) |
| 认证 | `Authorization: Bearer <XAI_API_KEY>`，凭据属于 xAI，不能复用其他 Provider 的 key。[R14](https://docs.x.ai/developers/rest-api-reference/inference) |
| SDK 复用 | 官方示例直接调用 OpenAI SDK 的 `responses.create`，不需要 xAI 专用 SDK 才能使用此协议。[R2](https://docs.x.ai/developers/model-capabilities/text/generate-text)[R9](https://docs.x.ai/developers/grok-4-7) |
| 其他资源操作 | 文档列 `GET/DELETE /v1/responses/{response_id}` 与 `POST /v1/responses/compact`；它们的存在不要求 barness-ai 首期开放这些资源能力。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R15](https://docs.x.ai/developers/advanced-api-usage/context-compaction) |

Base URL 包含 `/v1`，实际请求必须核对只有一个版本前缀。服务绑定应固定 Provider、API Protocol、目标账户和 endpoint；协议兼容不授予跨租户或跨账户使用原生资源的权限。

### 3.2 请求与模型控制

| 能力 | 当前事实与边界 |
| --- | --- |
| 输入 / 指令 | `input` 支持字符串或条目数组；可使用 `instructions`。官方规定 `instructions` 不能和 `previous_response_id` 同时使用，续接沿用之前的系统提示。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses) |
| 角色 | 官方模型说明允许 `system/user/assistant` 自由顺序；不能由此推断另一 Provider 的 `developer` 角色规则或全部 SDK union 都兼容。[R13](https://docs.x.ai/developers/models) |
| 输出预算 | `max_output_tokens` 仅限制可见回答，不限制推理或 function-call tokens；默认 128000，可以设更大值。[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices) |
| 推理强度 | 4.7 / 4.6 支持 `low/medium/high/xhigh`，默认 `high`；4.5 只实质支持前三档，`xhigh` 会按 `high` 处理。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning) |
| 关闭推理 | 4.7 / 4.6 / 4.5 不能关闭推理；不能把统一接口的 off/none 默认翻译成有效关闭。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning) |
| 多 Agent 型号 | `grok-4.20-multi-agent` 的 effort 控制 agent 数量，不能和上述型号共用“推理深度”解释。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning) |
| `reasoning.summary` | schema 称只为兼容而保留，实际固定 detailed；`generate_summary` 同样仅兼容。**摘要可接收**与**摘要粒度可配置**是不同能力。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning) |
| 惩罚 / 停止词 | 推理型号不允许 presence/frequency penalty 或 stop；不要由 SDK 字段存在推断可用。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning) |
| Temperature / TopP | 可提交 temperature 0–2、top_p≤1，但实际 temperature 上限 1，top_p 上限为模型默认值（当前通常 0.95）；响应可能仍回显提交值。[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices) |

因此统一参数可以保持原值，校验与文档须说明真实采样上限；不可把回显值误标成厂商实际采样参数，也不应默默改变 Host 要求的推理能力。

下列字段必须从“有效支持”中剔除或单独处理：

| 字段 | 官方说明 |
| --- | --- |
| `metadata`、`truncation` | 保留兼容形状，未提供对应功能。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses) |
| `background` | 不支持异步 background；不能因响应有同名字段就开放后台任务。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses) |
| `context_management` | 可以解析指令，但尚不执行；独立 `/compact` 端点不改变这点。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R15](https://docs.x.ai/developers/advanced-api-usage/context-compaction) |
| `include:message.output_text.logprobs` | 接受但静默忽略。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses) |
| `logprobs/top_logprobs` | `grok-4.20` 及更新型号静默忽略。[R13](https://docs.x.ai/developers/models) |
| `prompt_cache_retention` | 当前 Responses schema 与缓存专题未给出支持承诺；不应继承 OpenAI 控制。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R16](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/maximizing-cache-hits) |

### 3.3 会话与加密原生续接状态

默认服务端保存输入和输出，Responses 资源保留 30 天，可通过 response ID 获取或续接；`store:false` 关闭这类会话资源存储。超过保留期仍要继续时，客户端需持有历史与加密推理。[R2](https://docs.x.ai/developers/model-capabilities/text/generate-text)

官方推荐两种模式：

1. `store:true`（默认）加 `previous_response_id`，由服务端恢复历史。
2. `store:false` 加完整本地历史与加密 reasoning items，每次请求自包含。[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)

Grok 4.7 总会返回 `reasoning.encrypted_content`；其他适用型号可通过 `include:["reasoning.encrypted_content"]` 请求。默认返回密文和显式 include 在内部 thinking trace 留存行为上还有差别，但 `store` 仍独立控制是否建立可用 ID 续接的资源。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning)

**设计建议**：统一走第二种模式，明确设置 `store:false`，显式 include 加密推理，原样保留 output 中需要回放的 reasoning、message 和 function-call 条目。不要只提取 `output_text`；不要把加密状态改写成普通 thinking 文本。是否、何时丢弃受限状态，应遵循 barness-ai 已有 Native State Envelope 与回放策略。

`store:false` **不代表零留存**：安全 FAQ 说明默认存在 30 天的审计留存；ZDR 由 team 级设置控制并影响全部 key，会禁用依赖持久数据的功能。响应头 `x-zero-data-retention` 可供 Host 核查团队状态。[R6](https://docs.x.ai/developers/faq/security)

### 3.4 文本、图片和文件

| 输入 | 文档确认 / 接入边界 |
| --- | --- |
| 文本 | 使用普通消息或 `input_text`。[R2](https://docs.x.ai/developers/model-capabilities/text/generate-text) |
| 图片 | 使用 `input_image.image_url`，可以是公开 URL 或 base64 data URL；官方示例含 `detail:high`。[R17](https://docs.x.ai/developers/model-capabilities/images/understanding) |
| 图片型号 | 4.7 支持文本、图片输入和文本输出；图片上限 20 MiB，Models 页列 JPEG/PNG。[R9](https://docs.x.ai/developers/grok-4-7)[R13](https://docs.x.ai/developers/models) |
| 图片与存储 | Image Understanding 指南建议关闭服务端会话存储，否则可能失败；同页部分示例没有带 `store:false`，实现按文字限制处理。[R17](https://docs.x.ai/developers/model-capabilities/images/understanding) |
| 文件 | Responses 支持 `input_file` 的公开 `file_url` 或已上传 `file_id`；附文件会自动启用 document search，转成服务端工具流程，并产生额外资源和费用语义。[R18](https://docs.x.ai/developers/model-capabilities/files/chat-with-files) |

**设计建议**：首期文本与内嵌图片即可。文件、公开 URL 抓取、server-side tools 不应因协议字段存在而默认放行；它们需要 Host 的资源授权、工具许可与费用处理，超出基础聊天转换范围。

### 3.5 Function Calling 与结构化输出

工具请求采用 Responses 的平铺 function 定义；响应 `function_call` 含 arguments，工具结果以 `function_call_output` 和对应 `call_id` 交回。自定义工具由 Host 执行；与之混用的内置工具则由厂商执行。[R10](https://docs.x.ai/developers/tools/function-calling)

| 控制 | 文档确认 |
| --- | --- |
| 默认 / 必选 / 禁用 | `auto/required/none`。[R10](https://docs.x.ai/developers/tools/function-calling) |
| 并行 | 默认允许，`parallel_tool_calls:false` 关闭。[R10](https://docs.x.ai/developers/tools/function-calling) |
| 工具 schema | 参数根须为 object，或每个分支均为 object 的 anyOf/oneOf；不合要求会 400。[R10](https://docs.x.ai/developers/tools/function-calling) |
| 严格性 | 工具参数始终按 schema 生成，`strict` 被描述为隐式 true；不能照搬另一 Provider 对 strict 的降级语义。[R11](https://docs.x.ai/developers/model-capabilities/text/structured-outputs) |
| 结构化回答 | Responses 使用 `text.format`；json_schema、json_object 与 text 按端点形状编码。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R11](https://docs.x.ai/developers/model-capabilities/text/structured-outputs) |

JSON Schema 是支持子集：可选字段不要求一律设为 required，`additionalProperties` 默认 false；非循环 `$ref` 可用，部分条件、复杂约束只是尽力满足。不要用 OpenAI 的 schema 重写规则覆盖 xAI 的规则；若业务要求严格，需要独立验证生成结果。[R11](https://docs.x.ai/developers/model-capabilities/text/structured-outputs)

内置 web/x search、code execution、collections/file search、MCP 等可在 Responses 内使用，但不表示 barness-ai 可以把所有名字当作原厂 OpenAI tools。当前 REST 总述甚至仍称只支持 functions/web search，与专题的工具列表不完全同步。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R19](https://docs.x.ai/developers/tools/streaming-and-sync)

### 3.6 Streaming 与终态

Responses 专属官方示例确认以下事件：[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning)[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)

| 事件 | 意义 / 实现要求 |
| --- | --- |
| `response.output_text.delta` | 增量回答文字；按 item/content index 聚合。 |
| `response.reasoning_text.delta` | 推理摘要的一种事件表示。 |
| `response.reasoning_summary_text.delta` | 推理摘要的另一种事件表示；避免重复累计。 |
| `response.completed` | 完整 response，包括续接需要的加密 output items。 |
| `response.incomplete` | 不完整 response；保留结果并标明截断。 |

Function Calling 指南特别说明：流中的 function call 以整块返回，不保证按参数字符串分片。现有 decoder 需要接受最终条目已经带完整 arguments 的情形，不能依赖一定先收到参数 delta。[R10](https://docs.x.ai/developers/tools/function-calling)

REST 字段描述仍写 data-only SSE 和 `[DONE]`，但它并未替代 Responses 对象与终态；**设计建议**是依据协议终态判断成功，不把 EOF 或 `[DONE]` 自身视为完成。完整事件顺序、失败事件、只有 final item 的工具调用和 usage 到达位置，仍应取得 xAI 自己的 stream 证据。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)

官方建议长推理请求留足超时：最佳实践建议至少 900 秒，复杂评测可超过 1800 秒。此为工作负载建议，不是修改 barness-ai 全局默认值的理由；Host 仍决定 deadline、取消与预算。[R5](https://docs.x.ai/developers/model-capabilities/text/best-practices)

### 3.7 Usage、缓存和成本

Responses 至少有 input/output/total、`input_tokens_details.cached_tokens`、`output_tokens_details.reasoning_tokens`；缓存输入包含在 input 总数中，缓存命中与 reasoning 按各自费率计费。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R20](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing)

**用量疑点，不能直接确认为稳定规律**：REST 创建响应示例为 input=32、output=9、reasoning=110、total=151，恰为 `32+9+110`；schema 只笼统描述 output，并未明确说 reasoning 是否包含其中。另一个缓存专题示例 reasoning=0，不能消除该歧义。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R20](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing)

若真实行为与该示例一致，把 OpenAI 的 output 包含 reasoning 这一假设直接套用，会少计 xAI 消耗。应单独明确实际输出与推理的包含关系，再转换为 barness-ai Usage；缺失字段也不能当作零消耗。

缓存由厂商自动完成。Responses 的 `prompt_cache_key` 是固定会话路由提示，功能等同 `x-grok-conv-id`，帮助命中同服务器缓存，不是安全隔离凭据；缓存可能随负载或重启被驱逐，没有保留时长承诺。[R16](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/maximizing-cache-hits)[R21](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/best-practices)

**设计建议**：如果复用现有 CacheKey，沿用当前 tenant/provider/account/session 派生范围；是否另加协议维度须以实际缓存合同决定，不把原生状态的协议隔离规则误称为现有缓存键实现；不发送可识别用户信息，不把缓存 key 当成资源授权。保持消息前缀不变，缓存 miss 也应正常工作。

厂商报告 `usage.cost_in_usd_ticks`，单位为 `1 USD = 10^10 ticks`，官方称它包含此请求所有 token、工具费与折扣，是实际请求计费字段；REST schema 还保留另一个 `cost_in_nano_usd`，两者单位不同，不能混用。[R8](https://docs.x.ai/developers/cost-tracking)[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)

厂商计费字段可以作为独立的 Provider 报告保留，不应无说明地覆盖模块的 Estimated Cost。失败或中断时不一定有最终 cost，用量未知不能标成免费。

截至研究日，4.7 的全球 default 价格如下，单位 USD / 1M tokens：[R7](https://docs.x.ai/developers/pricing)

| prompt 总数 | 输入 | 缓存输入 | 输出 / 推理 |
| --- | ---: | ---: | ---: |
| `<200000` | 2.00 | 0.50 | 6.00 |
| `≥200000` | 4.00 | 1.00 | 12.00 |

长上下文是达到阈值后整个请求使用较高费率，不是只对超出部分收费；prompt 阈值包含缓存 token。priority 对所有 token 费率乘 2，且仅在实际响应确认 priority 时适用。[R7](https://docs.x.ai/developers/pricing)[R20](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing)[R22](https://docs.x.ai/developers/advanced-api-usage/priority-processing)

US endpoint 为 `https://us.api.x.ai/v1`，当前只支持 4.7/4.6，同 key 可用，token 费率另乘 1.1；cache 不保证跨 endpoint 复用。区域请求、工具和文件的处理范围也不同，不能只换 URL 后仍用全球目录价格。[R23](https://docs.x.ai/developers/advanced-api-usage/regions)

### 3.8 错误、配额与资源边界

官方列 400/401/403/404/405/415/422/429；认证失败不能重试成普通参数错误。限流按 team、model 的 RPS 与 TPM 计算，reasoning 和缓存 tokens 同样占 TPM，缓存折扣不降低配额消耗。[R24](https://docs.x.ai/developers/debugging)[R25](https://docs.x.ai/developers/rate-limits)

**设计建议**：继续沿用 Binding 显式 Retry Policy、Admission、取消和有界流队列；不能照抄文档中“持续重试直到有结果”的评测建议。流开始后不自动重放，响应与错误保留可诊断元数据并脱敏。

## 4. Messages：已有兼容面与弃用边界

本节事实来自当前官方 `/v1/messages` 参考及该页嵌入的折叠 schema。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy) **已有 schema 不等于所有当前 Grok 型号均支持，也不等于将来持续维护。**

### 4.1 请求与响应形状

| 项目 | 文档确认 |
| --- | --- |
| 端点 | `POST /v1/messages`；使用 xAI 服务地址。 |
| 基本输入 | model、max_tokens、messages；system 可为字符串或 text blocks；messages.content 可为字符串或 blocks。 |
| 输入 blocks | text、image、tool_use、tool_result、thinking、redacted_thinking。 |
| 图片 source | base64 的 media_type/data，或 URL；schema 列 JPEG/PNG/WebP，但与 Models 页类型说明有差异。 |
| 工具 | 自定义函数，name/description/input_schema；不由此开放厂商托管工具。 |
| tool_choice | `auto`、`any`、指定 `tool`；当前 schema 未列 `none`、`disable_parallel_tool_use`。 |
| 工具往返 | tool_use 使用 id/name/input；tool_result 使用 tool_use_id、content，可带 is_error；content 支持字符串或 text/image blocks。 |
| stop_reason | end_turn、max_tokens、stop_sequence、tool_use，流中未到终态可以为 null。 |
| metadata | 有 user_id；不是任意元数据能力。 |

### 4.2 Thinking、缓存与不支持字段

| 能力 | 当前可确认的范围 |
| --- | --- |
| Thinking 输出 | response union 有 `thinking:{thinking,signature}`，signature 在 response schema 必填。 |
| Thinking 输入 | input thinking 的 signature 可为 null；原生回放仍应使用真实返回值，不能把 schema 的宽松输入当成伪造签名许可。 |
| Redacted thinking | 输入输出都列 `redacted_thinking:{data}`。 |
| 控制推理 | 顶层请求 schema 未列 thinking、budget_tokens、effort、output_config；不能由 block 类型反推这些开关可用。 |
| 显式缓存 | text/image/system/tool/tool_result 中的 cache_control 标注不支持。 |
| 缓存 usage | input_tokens 明确是不由缓存提供的输入；cache_read_input_tokens 单列；cache_creation_input_tokens 不支持。 |
| TopK | 顶层 top_k 明确不支持；不能从 Responses 的同名字段推导可用。 |
| 停止词 | reasoning models 不支持 stop_sequences；其他场景上限四个。 |

该表均为当前 Messages schema 的事实。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy) 若将来实现，Native State Envelope 仍须分别保护 Messages signature/redacted data，不能与 Responses 密文互授回放资格。

### 4.3 认证、版本与 SSE 的证据缺口

Inference 总览统一说明 Bearer 认证，但当前 Messages 段没有给出明确的 Anthropic SDK 初始化或 X-Api-Key 认证例子。不能仅凭“兼容 Anthropic”就断言默认 Anthropic SDK 请求头一定可用。[R14](https://docs.x.ai/developers/rest-api-reference/inference)[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy)

当前 Messages 段未说明 `anthropic-version`、`anthropic-beta` 或 `?beta=true` 的要求与容忍范围；也没有公开标准 Messages 事件序列，仅笼统写 SSE、`[DONE]`。没有证据证明 `message_start → content_block_* → message_delta → message_stop` 及 signature_delta 完整兼容。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy)

**设计建议**：若真实需求要求继续接入，先以 stable `/v1/messages`、最小请求形状建立该路径自己的证据；逐项确认认证与 header，然后再开放 thinking、工具和 streaming。不要通过吞错、忽略签名或省略原生状态来伪造兼容。

## 5. 需要保留的开放问题

下列问题没有被文档研究解决，不能在功能表中标成 PASS：

1. Responses output/reasoning 的实际包含关系，以及 stream 终态与非流 usage 是否一致。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R20](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing)
2. function-call 整块返回具体使用哪些 item / done 事件，参数是否可能完全没有 delta。[R10](https://docs.x.ai/developers/tools/function-calling)
3. reasoning_text 与 reasoning_summary_text 是否可能在同一响应重复表达同一摘要。[R4](https://docs.x.ai/developers/model-capabilities/text/reasoning)
4. 工具数 350/1000、指定 function 的平铺/嵌套形式，以哪个现行契约为准。[R1](https://docs.x.ai/developers/rest-api-reference/inference/responses)[R10](https://docs.x.ai/developers/tools/function-calling)
5. Messages 的当前模型白名单、认证 header、版本 header、SSE 与 thinking signature 回放。[R3](https://docs.x.ai/developers/rest-api-reference/inference/legacy)[R14](https://docs.x.ai/developers/rest-api-reference/inference)
6. 多轮切换 model、account、endpoint 时厂商是否接收旧密文；无论厂商是否宽松，模块仍须遵守自身来源边界。

## 6. barness-ai 接入方案（基于当前仓库的建议，尚未作实施决定）

源码基线：`2f58c68d8fbb30a07e71b9666feed9979c791d01`（2026-10-26 读取）。以下是建议，不代表已实现或已取得厂商支持资格。

### 沿用 DeepSeek 的架构模式，按 xAI 的事实取值

[ADR-0014](../../docs/adr/0014-barness-ai-deepseek-responses.md)确立 Provider 与 API Protocol 独立、复用 Responses adapter、按 Provider 生成字段、完整历史回放的方式；[ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md)已把同一边界扩展到 Messages。当前代码已有多家兼容服务的 profile，不需要重建 adapter 注册框架。

建议新增 `ProviderXAI = "xai"`，产品名称保留 Grok；仅使用现有 `APIOpenAIResponses` 与 `APIAnthropicMessages`，不新增 `APIGrok`。协议内的厂商差异由内部 profile 与边界转换承担，复用现有 Client、assembler、限额、取消、准入、错误归类和显式重试。现有 OpenAI Go v3.66.0 与 Anthropic Go v1.75.0 已在 [go.mod](../../go.mod) 中，首期无需引入 xAI SDK。

| 路线 | 建议 Binding | SDK base URL | 首期定位 |
| --- | --- | --- | --- |
| xAI × Responses | `grok-responses`，Provider=xai，API=openai-responses | `https://api.x.ai/v1` | 正式接入的目标，先以候选装配取得本组合证据 |
| xAI × Messages | `grok-messages`，Provider=xai，API=anthropic-messages | `https://api.x.ai` | 弃用兼容路线；首期不默认启用、不纳入正式内置目录 |

两份 Binding 分别授权型号；只有 Host 已授权时才可引用同一 xAI 账户及凭据。不同 Provider/协议的成功不能互授支持资格；失败不能自动切到另一协议，也不能用本地 Messages→Responses 转译冒充 xAI 的 Messages 原生兼容。授权边界见 [binding.go](../../ai/binding.go)、[adapter.go](../../ai/adapter.go)。

### Responses：需要明确处理的边界

1. **显式 xAI profile。** 在 [responses_capabilities.go](../../ai/responses_capabilities.go) 增加 xAI 分支；当前 default 是 OpenAI 全功能配置，新增 ProviderID 不能直接落入 default。与 DeepSeek 不同，xAI 支持 `store:false` 和 encrypted reasoning，不能复制 DeepSeek 的 false 配置。缓存 key、缓存保留期、亲和头、summary、service tier 分别判断，不能用一个“兼容 OpenAI”开关捆绑。首期建议 store/encryptedReasoning/promptCacheKey 按已证实子集开启，promptCacheRetention/sessionAffinity/reasoningSummary/serviceTier关闭；reasoningSummary=false指不提供摘要风格控制，仍消费4.7自动返回的摘要。strictFunctionTools先关闭以省略strict字段，避免当前builder的strict:false被误解为可关闭xAI的schema约束。
2. **无状态请求保持完整。** 系统指令先使用xAI明确列出的system角色；当前builder对多数推理型号会自动改成developer，需为xAI单独约束，不从SDK union推导支持。每次发送完整 input，必须显式 `store:false`，不用 `previous_response_id`、conversation、后台运行或账户资源引用。[responses_request.go](../../ai/responses_request.go) 现有回调授权仅在 store 存在且不等于 false 时拒绝，允许回调删除 store；xAI 默认保存，因此新增最终请求守卫必须要求回调之后 store **存在且严格为 false**，并再次验证流模式、型号和资源引用。
3. **推理档位与摘要分开。** 首个候选可选当前官方主推的 `grok-4.7`，逐协议保留条目。使用既有 `ThinkingLevelMap`、`DefaultThinkingLevel` 与 `RequiresThinking` 表达真实约束；建议4.7候选明确 off/minimal/max=null，low/medium/high/xhigh映射同名wire值，DefaultThinkingLevel=high、RequiresThinking=true；首期可主动把输出能力收窄为128000以保留厂商默认，明确这是barness候选范围而非厂商硬上限。这些只是候选配置，仍需对应协议验证。simple 保留当前公开的 clamp 语义，full 与最终 sampling/回调路径校验实际 wire 值。可显示的 summary 与用于续接的 encrypted content 分别保存，不能用摘要替换密文。官方把reasoning.summary列为兼容字段且实际总返回detailed，建议不自动附summary:auto。xAI请求应在默认effort和显式effort两种情况下都明确include:[reasoning.encrypted_content]；现有builder只在显式推理控制时附include，必须为xAI补上该缺口，而非凭4.7默认返密文推断存储语义相同。
4. **保留原生状态。** 复用 [responses_signatures.go](../../ai/responses_signatures.go)、[responses_assistant.go](../../ai/responses_assistant.go) 与 [native.go](../../ai/native.go) 保存并回放完整 reasoning item；同一可信 tenant/account/provider/API/model 才能原生续接。工具 call_id 与 item id 需分别保留与配对；当前 [responses_history.go](../../ai/responses_history.go) 对非 OpenAI 的跨来源 ID 会整体净化，不能未经 xAI 证据直接照搬该规则或 `fc_`/64 字符假设。
5. **校验预算和采样实际语义。** xAI 的 max_output_tokens 不能作为总 output/总费用上限；公共选项注释和 Host 文档应明确此差异，实际费用仍按最终 usage。当前 [responses_options.go](../../ai/responses_options.go) 会对多数 Provider 把小于16的预算抬到16，不应无来源继承。对 temperature/top_p 的服务端 cap，应明确记录“请求值”和“实际采样语义”的差别，不能把响应 echo 当作执行证明。
6. **用量与价格独立转换。** 当前 [usage.go](../../ai/usage.go) 约定 Reasoning 已包含于 Output，但官方 Responses 示例对二者包含关系留下疑点；先通过 xAI 自身报告核实，再决定是否合并 reasoning，缺项不能默认为免费。首期可只服务默认等级、全球 endpoint，不开放显式ServiceTier选项（包括显式default），统一保持省略；这是建议的交付范围，不是声称 xAI 不支持 priority。官方 priority 当前为2倍，不能因数值相同就借用 [responses_usage.go](../../ai/responses_usage.go) 的 OpenAI规则；[定价](https://docs.x.ai/developers/pricing)还要求长上下文全请求分档。现有 CostTier 是严格大于阈值，若官方边界确为≥200000，需正确表达恰好200000并覆盖边界，不能照填200000导致少计。US endpoint有1.1倍价格，但当前目录身份不含endpoint，初期限制全球服务可避免同型号误用价格；以后扩地区需设计绑定价格选择，不因同key可用就自动跨区fallback。厂商报告费用与 Estimated Cost 分开，ticks/nano单位要精确，不以新字段覆盖旧成本语义。
7. **工具先交付现有公共合同。** 首期以函数工具、完整参数、工具结果续接为目标；严格结构化输出与托管搜索等另立有证据的能力。当前 Responses builder 在 SupportsStrictMode 为真时发送的是 `strict:false`；xAI文档将工具约束描述为隐式strict:true，应先省略该字段并按xAI的schema子集验证，不能声称可以关闭严格性，也不能把目录开关或 SDK 可编码误报为 barness 已提供可配置的严格工具合同。托管工具仍由 Binding.AllowedHostedTools 明确授权，默认不放行。

### Messages：保留一个可实施但不默认投资的分支

建议当前先保留研究与候选方案，不为弃用路线创建正式型号条目。若确有宿主必须使用 Messages 的业务要求，则复用现有 `anthropicAdapter`，增加独立 xAI profile、选项校验、最终请求守卫及本组合流解析验证，按已证实子集交付。

只交付 Responses 时，xAI × Messages 应明确拒绝为尚未实现的路线，不能让 Host 自带目录落入 Anthropic 原厂默认配置；以后实现该路线时再用独立证据解除此限制。

[anthropic_capabilities.go](../../ai/anthropic_capabilities.go) 默认是 Anthropic 原厂配置：缓存断点、adaptive/native thinking、beta、eager streaming、browser access 等不能随 xAI Provider 自动打开。顶层 thinking 接受或被忽略，也不构成 Claude thinking/signature 回放合同的证明；无法兑现的显式选项应明确失败，不能静默冒充启用。

[anthropic.go](../../ai/anthropic.go) 当前只有 Kimi 分支使用稳定 MessageService，其余经过 BetaMessageService；本地固定 SDK 源码显示后者发送 `v1/messages?beta=true`，前者发送 `v1/messages`。xAI 首次验证应采用官方公开的稳定路径；不得从 Anthropic SDK 自身推导 xAI 支持 beta query、原厂 beta header 或 thinking 预算。认证及版本头仍须按本报告列出的 xAI 文档缺口取得本路线 wire 证据。

SDK 来源（本地已核对的相同版本官方源文件）：[message.go](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/message.go)、[betamessage.go](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/betamessage.go)、[requestconfig.go](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/internal/requestconfig/requestconfig.go)。SDK 会自动添加 `anthropic-version:2023-06-01`，这证明发包行为，不证明 xAI 对该头的语义承诺。

### 目录、验证与实施顺序

| 阶段 | 交付内容 | 完成依据 |
| --- | --- | --- |
| 1. 固定候选契约 | 首个 Responses 型号、profile、请求字段与模型约束、标准费率、明确未知项 | 官方来源及本报告；不改变支持矩阵 |
| 2. 先失败再实现 | 公共 Client E2E：字段presence、回调删除store、文本/图片、两种reasoning增量、密文续接、工具往返、错误/断流、租户与账户隔离 | 按 [测试 SOP](../../docs/agents/verification.md) 留存命令、请求、事件、结果与可重复的脱敏证据 |
| 3. 本组合真实确认 | xAI × Responses 的文本、reasoning、工具续接、图片、用量、缓存；只在固定范围和预算获授权后执行 | 自身 live 捕获及回放，不以 SDK 示例或其他 Provider 通过替代 |
| 4. 晋升与文档 | 固定来源的型号/价格版本、支持矩阵、实施 ADR、Host 装配示例 | 满足 [目录纳入准则](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md) 与当前 SOP 的全部适用条件 |
| 5. 必要时再做 Messages | 先核实当前 endpoint/型号可调用、认证、流事件与工具子集，再决定候选实现 | 独立 Messages 证据；端点弃用的生命周期风险不能由一次文本 PASS 消除 |

建议首期最小改动集中在 `binding.go`、Responses capability profile 与 Messages 未实现守卫、Responses 的选项/最终请求守卫、历史/用量中经证据确认的差异，以及独立候选 catalog；Messages 真正进入实施范围时再增加其 profile。新代码文件继续遵循500行限制，不复制整套生命周期。以上是后续实现需求，本次只交付研究文档。

本次只有文档研究与源码审阅；未读取凭据、未调用推理API、未运行live或实现验收。本报告不建立正式支持资格。纯文档验证限于引用、路径、格式、事实/推断区分及范围核对，未以文档完成替代任何测试结果。

## 7. 官方来源索引

- [R0 — 文档目录与 Markdown 读取约定](https://docs.x.ai/llms.txt)
- [R1 — Responses REST Reference](https://docs.x.ai/developers/rest-api-reference/inference/responses)
- [R2 — Generate Text：存储、续接和加密历史](https://docs.x.ai/developers/model-capabilities/text/generate-text)
- [R3 — Legacy & Deprecated：Messages 及页面嵌入 schema](https://docs.x.ai/developers/rest-api-reference/inference/legacy)
- [R4 — Reasoning：型号、effort、加密内容与摘要事件](https://docs.x.ai/developers/model-capabilities/text/reasoning)
- [R5 — Best Practices：Responses 多轮、stream、预算和采样](https://docs.x.ai/developers/model-capabilities/text/best-practices)
- [R6 — Security FAQ：审计留存、ZDR 与标识](https://docs.x.ai/developers/faq/security)
- [R7 — Pricing：长上下文、priority、Fast 和区域价格](https://docs.x.ai/developers/pricing)
- [R8 — Cost Tracking：实际费用与 ticks](https://docs.x.ai/developers/cost-tracking)
- [R9 — Grok 4.7 专页](https://docs.x.ai/developers/grok-4-7)
- [R10 — Function Calling](https://docs.x.ai/developers/tools/function-calling)
- [R11 — Structured Outputs](https://docs.x.ai/developers/model-capabilities/text/structured-outputs)
- [R12 — Streaming 通用指南](https://docs.x.ai/developers/model-capabilities/text/streaming)
- [R13 — Models：模态、图片限制、角色和型号别名](https://docs.x.ai/developers/models)
- [R14 — Inference API：地址与认证](https://docs.x.ai/developers/rest-api-reference/inference)
- [R15 — Context Compaction](https://docs.x.ai/developers/advanced-api-usage/context-compaction)
- [R16 — Maximizing Cache Hits](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/maximizing-cache-hits)
- [R17 — Image Understanding](https://docs.x.ai/developers/model-capabilities/images/understanding)
- [R18 — Chat with Files](https://docs.x.ai/developers/model-capabilities/files/chat-with-files)
- [R19 — Tools Streaming & Sync](https://docs.x.ai/developers/tools/streaming-and-sync)
- [R20 — Prompt Caching Usage & Pricing](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing)
- [R21 — Prompt Caching Best Practices & FAQ](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/best-practices)
- [R22 — Priority Processing](https://docs.x.ai/developers/advanced-api-usage/priority-processing)
- [R23 — Regional Endpoints](https://docs.x.ai/developers/advanced-api-usage/regions)
- [R24 — Debugging Errors](https://docs.x.ai/developers/debugging)
- [R25 — Rate Limits](https://docs.x.ai/developers/rate-limits)
