---
status: accepted
date: 2026-10-17
---

# 中国区标准 GLM Chat 复用生命周期，型号约束守住最终请求

按 ADR-0032 的 Provider/Model 边界，中国区标准智谱以 `ProviderZhipu=zhipu` 和既有 `openai-completions` 路由进入同一 Client/adapter/SDK 生命周期。Provider 配置选用 system、max_tokens、thinking 开关，省略 store、OpenAI 缓存与未核验的 stream_options 默认；省略 effort 保留服务默认思考。DeepSeek 给 assistant 补空 reasoning_content 的规则独立保留，不传给智谱。明确提供 stream_options 的有界候选对照仍可经安全回调发送，字段未出现在 schema 不被解释为官方拒绝。

型号以独立候选 Catalog 声明能力；GLM-5.3 的 off/minimal/medium/xhigh 为 null，low/high/max 有映射，GLM-5.2 则独立采用官方标准 Chat 的开关和别名映射。新增可省略的 `ModelCompat.RequiresThinking` 表达不可关闭思考，零值不改变既有模型，值随目录复制和 Hash 固定。使用型号 ID 分支会把型号事实藏入 adapter，使用 Provider 全局强制思考则会错误限制 5.2，因此均未采用。

full 显式 canonical 档位在绑定/型号授权后、凭据读取前校验；simple 仍按既有向上优先规则夹取。模型采样、canonical 分档、调用采样与 OnPayload 后只校验实际最终 wire，不重新选档；回调可以修复中间采样值，但不能使非法 full 具名档位悄悄变成另一个档位。最终授权检查先于型号能力检查；无 hook 的路径同样校验。最终 max_tokens 必须为 1 到 min(型号上限, Chat schema 131072) 的整数，max_completion_tokens 形状拒绝；具名 null/0 仍省略，实际 wire 上的 null/0 则拒绝。Host 目录可以收窄输出预算，不能扩大服务 schema 上限。发送前失败为零 Provider 请求与零 Attempt。原 OpenAI/DeepSeek full wire 与 canonical 采样分离的契约不变。

候选数据位于 `ai/internal/testkit/glm`，Chat 两条身份各自固定 text-only、官方 schema 的 131072 输出上限与明确标为 barness 保守值的 1000000 上下文，价格仅为合成 fixture 价格。它们未进入内置目录。Responses/Messages 必须有自己的数据和证据，不从 Chat 自动派生；完整 live、价格与正式支持仍由后续票验收。候选 Host/controlled live 只共享已有 Client 装配，单调用策略固定在 120 秒，后续隔离入口另行执行 spec 的双开关、单组合凭据和 24/32 预算。

[工单 02 的证据索引](../../.scratch/barness-ai-glm/evidence/ticket-02/README.md) 给出逐片 red/green、回归、race、冻结差分与重放命令。删除候选装配和 Provider 配置即可撤回尚未发布的路线；内置目录和既有路由不变。

工单 03 补齐函数工具和同 turn 推理往返：声明函数工具时自动发送 `tool_stream:true`，最终请求仍必须保留它；默认与 auto 是首期工具选择的完整集合，full required/none/具名/allowed_tools 与 simple none 均明确拒绝。具名限制在授权之后、凭据之前检查；采样和回调在最终 wire 检查，合法回调可以修复中间采样。`thinking.clear_thinking` 省略时保留官方默认 true，显式值只接受 true；false 属于尚未接入的跨 turn Preserved Thinking，不能借同 turn reasoning_content 回放假称支持。最终授权错误仍优先于这些能力错误。

工具 delta 按已有 Chat index/ID 组装；Host 在完整工具终态后使用 ValidateToolCall、ToolResultText 并自行调用下一轮。GLM 同模型受信回放保留原 `reasoning_content`，普通 assistant 不补 DeepSeek 的空字段。持久化反序列化不会恢复原生信任，须 Host TrustNativeState；不可信或跨身份历史沿既有规则保留可见推理文本和工具配对后降级，不把它计作原生回放。公共 E2E 同时覆盖两型号分片、Host 往返、零请求错误、历史与不可变快照，见[工单 03 索引](../../.scratch/barness-ai-glm/evidence/ticket-03/README.md)。这些为合成离线协议证据，不代替真实服务与正式目录验收。

GLM 的最终输出预算按原始 JSON 数字做精确整数校验，避免 `1.00000000000000001` 被 float64 舍入成合法的 1。GLM Chat 仅保留其有效 max_tokens 采样值直到最终回调之后检查；其他采样字段和旧 Provider 继续既有 JavaScript 数字语义。整数值的 1.0/1e0 仍合法，超大指数先有界拒绝，不展开其数值。Chat 与 Responses 共享精确数值约束，字段选择仍由各协议拥有。

工单 09 将相同边界补到 Chat 最终 native 容器：仅接受文本消息、system/user/assistant/tool 角色、流式请求及函数声明。函数 `strict` 即使为 false 也不发送，Host Catalog 的 Compat 不能扩大 Provider 的 endpoint 能力；工具名称、JSON Schema 属性与 arguments 作为业务数据独立保留。已授权 hosted tools（含 Chat `web_search_options`）仍须通过 Provider 子集检查，绑定放行并不表示该能力已接入（继承 ADR-0005）。未接入的 store、缓存、tier、parallel、context-management 与严格 response-format 在最终请求明确失败；合法 text/json_object、do_sample、temperature 及候选 `stream_options.include_usage` 对照继续可用。授权先于能力判断，保留 tenant_denied 与零 Attempt；OpenAI/DeepSeek 的既有请求不变。来源和公共 Client 红绿证据见[09 索引](../../.scratch/barness-ai-glm/evidence/ticket-09/README.md)。
