---
status: accepted
date: 2026-10-17
---

# GLM Responses 在原始 JSON 边界转换解析副本，单独保留原生 reasoning

GLM 官方 Responses schema 允许 reasoning.content 为对象、message.content 为对象或数组；锁定的 Go SDK 可以无错误解码对象却丢掉正文。因此沿用现有 Responses adapter 的生命周期，在 SSE JSON 进入 SDK 前按结构校验并把已知内容转换为解析用数组，同时按 output_index 保存未经转换的 reasoning 条目供 Native State 使用。解析副本不补造 summary、encrypted_content 或 OpenAI ID；原生签名保留厂商的对象或数组形状。直接认定 Unmarshal 成功意味着内容完整、字符串替换或在 transport 改写响应都会掩盖边界问题，故不采用。

消费端在请求边界独立转换已通过同型号、同账户信任检查的原始 reasoning：对象或数组中的 reasoning_text 转为 InputReasoning 的单个 content 对象，多段沿用可见文本的双换行连接。id/status/summary/encrypted_content 不属于该输入形状，全部省略；保存在公开 Result 与 Native State 中的原始签名不变。InputMessage 同样仅发送 type/role/content，既不发送厂商输出条目 ID，也不生成 OpenAI msg_pi_* ID；function_call 仅以 call_id 与结果配对，公开工具 ID 中的 fc_ 部分保留在原始结果。空原始 content 或只有 text.done 的可见思考以普通 assistant 文本保留，不伪造 reasoning 条目或状态 ID，也不声称这是原生续接。非空畸形原生内容明确失败；未担保、跨型号/Provider/API/账户的状态沿用既有降级，保留可见内容和工具配对。

权威 output_item.done 结束每个内容块；GLM 独立处理 text.done/reasoning_text.done 和仅终态输出，终态不再重复已结束的块。合法 completed/incomplete 不依赖 [DONE]；缺终态 EOF、畸形结构及与事件不符的状态明确失败。保持 OpenAI/DeepSeek 原有解析行为和 pi 契约。

Provider 协议能力只自动发送 store=false 和隔离后的缓存路由键；不带 summary、encrypted include、service tier、缓存保留期或亲和头。GLM max_output_tokens 按官方 schema 限于1..min(型号上限,131072)，不沿用 OpenAI 的16下限。所有采样层的有效预算保留原始 JSON 数字至最终边界，以共享精确整数校验拒绝浮点舍入后看似合法的分数；回调后再次校验。型号级完整档位在授权后、读取凭据前拒绝，simple 按独立候选目录夹取。

工具选择仅默认/none/auto；Provider 能力约束严格字段生成，Host Catalog 不能打开 GLM 的 strict。最终 JSON 在采样与可信回调后统一验证，先检查型号、隔离缓存键和账户引用授权，再检查协议能力、原生输入形状与文本型号的图片限制。未接入的 summary、encrypted include、tier、缓存 TTL、托管/namespace/custom 工具、严格 JSON Schema、服务端保存或 background 不会被回调绕过；这是本模块的接入边界，不等于声称厂商已经逐字段拒绝。合法 text/json_object 回调按原顺序运行一次，全部 Attempt 复用冻结字节。SessionID 存在且 retention 非 none 时，short/long/default 都只发送既有租户/账户派生的路由键，不承诺缓存寿命、亲和或命中率。

5.3 的 Responses generic effort schema 与型号页有冲突，首期按 spec 限制 low/high/max，省略 effort 则保留服务默认；wire low 不等于已测得的 Chat low 计算强度。独立 Responses 候选目录的1,000,000上下文是保守边界，131072是本协议 schema 精确输出整数；价格为明确标记的合成测试值。两个真实基础样本均为数组；对象失败的红绿证据来自官方 schema 的合成公共 Client E2E，不能称为真实对象样本或正式支持。

两型号的真实工具调用和结果续接已在探索中分别捕获；5.2 实际 reasoning 数组经可信实验回调转换成文档对象并续接成功。5.3 工具首轮没有 reasoning，因此其原生 reasoning 历史仅有自身基础数组的离线回放与独立 schema 场景，不能以 5.2 续接成功代替其真实证明。生产适配后的两型号正式 native-history 场景仍由工单11提供。

后续源审计确认官方 InputFunctionCall.call_id 与 OutputFunctionCall.call_id/id 均为普通字符串，没有排除 `|` 的字母表约束。原先直接 `call_id|item_id` 再按首个分隔符切分，会把 `same|a` 与 `same|b` 都回放为 `same`；两组不同原生元组甚至会得到相同公开工具 ID。GLM 因此在组合两个不透明组件前，分别单次转义 `%` 和 `|`；请求边界只逆转义调用组件。Unicode 不改写，普通已捕获 ID 的拼写保持不变，原生条目 ID 仍不发送。公开工具 ID 是 Host 不应解析的不透明关联值，没有新增 DTO 字段或第二套通用 ID 框架。

GLM 自身 Responses 历史在信任降级时保留这份无歧义的关联 ID；账户原生条目、reasoning 与可见内容仍独立遵守既有降级政策。其他来源的整个工具 ID 在 GLM 目标上作为新的调用组件，不使用 OpenAI 的受限字母表，也不猜测来源协议的分隔符；工具结果同步重命名。该取舍由两个型号公共 Client 的实际丢失/碰撞红例驱动，仅改变 GLM 目标，OpenAI/DeepSeek 的既有规范化与冻结差分合同保持。边界字符证据来自官方 schema 的离线合成场景，不冒充已观测的真实厂商 ID；见[专项红绿证据](../../.scratch/barness-ai-glm/evidence/ticket-05-followup-ids/README.md)。

证据与重放见[工单04](../../.scratch/barness-ai-glm/evidence/ticket-04/README.md)和[工单05](../../.scratch/barness-ai-glm/evidence/ticket-05/README.md)；正式目录、价格与独立真实验收由后续票提供。本决策继承 ADR-0014 的无状态完整历史与 Extension Route 边界，保留原有 OpenAI/DeepSeek 字节和差分合同。
