# 05: 完成 GLM Responses 无状态工具往返与缓存提示

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 以完整历史完成 GLM Responses 函数工具往返，并安全使用按租户和账户隔离的缓存路由提示；未接入的服务端能力在发送前明确拒绝。

Blocked by: 04

依据：GLM 接入 spec R08、R09、R10、R15、R16；验收组 A04、A07、A13；User Stories 24、25、26、27、28、30、34、35、40、41。

- [x] 先写两型号的完整文本/原生 reasoning/函数调用→工具结果→下一轮回答及失败 E2E，再实现历史转换。
- [x] 每轮完整回放 input，不使用 store=true、previous_response_id 或 conversation；GLM InputReasoning 形状正确，内部数组规范化不污染原生回放。
- [x] 函数参数增量、权威参数 done、call_id/条目 ID 与 function_call_output 正确配对；不因跨型号或来源降级制造孤立工具结果。
- [x] 仅默认、none、auto 工具选择可用；required、具名强制、allowed_tools、strict 及本期未接入托管工具不能通过具名选项、sampling 或 payload hook 偷渡。
- [x] 有 SessionID 且 retention 非 none 才发送按既有租户/账户/session 派生的 prompt_cache_key；不发送原始身份/session、TTL、retention 或 OpenAI 亲和头。
- [x] short/long 都只作为路由提示；无 session 或 none 不派生 key，不承诺命中率/期限；非法自报缓存键仍按授权契约拒绝。
- [x] 显式 summary、ServiceTier、encrypted include、服务端状态引用等按 spec 拒绝；合法回调保留原生 payload 与既有顺序，发送前失败零 Attempt。
- [x] 产出两型号完整无状态往返、缓存隔离和非法字段拒绝的可重放证据；同模型可信回放与未担保/跨账户降级均有公共入口示例。

**独立验收：** 不保存服务端会话地完成两型号工具往返，捕获下一轮 GLM 原生 reasoning 与正确隔离的缓存 key，并验证非法强制工具/状态引用没有请求。

**依赖理由：** 需要 04 的 Responses 调用、content 解析和原始 reasoning 状态。



## Answer

2026-10-17：两型号的 Responses 无状态完整历史、原生 reasoning 请求转换、Host 工具往返、默认/none/auto 工具选择、按租户/账户隔离的缓存路由提示与最终请求约束完成。GLM 输入消息、reasoning 和函数调用不发送输出/合成状态 ID；原始公开 Native State 不变。只有可见思考时按普通文本保留，非空畸形原生内容明确失败；未担保和跨身份历史保留可见内容及有效工具配对。

实现提交 `688847c`，合并最新集成至 `e107074` 后的已提交重放源 `d0d935f`：3,716 公共回归 PASS，471 Responses race PASS，335 冻结 OpenAI Responses 差分 PASS、零 pending，vet/diff checks 通过。具名能力校验早于凭据；最终采样/回调约束晚于授权检查，非法请求零网络与零 Attempt。合法回调和显式重试复用同一冻结请求。

上下文：[ADR-0034](../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md)、[失败计划与逐包证据](../evidence/ticket-05/README.md)。原始包已持久保存在根 `.evidence/barness-ai-glm/ticket-05/`。本票未读取 `.env` 或执行收费调用。5.3 自身真实 native-history 仍由工单11验证；5.2 取证不能替代该项，正式目录与矩阵保持原交付边界。


### Scoped follow-up: opaque tool IDs before ticket11 freeze

2026-10-17：官方 schema 的 plain-string call_id/id 未排除分隔符，公共 Client 已在两型号复现原生 `same|a`/`same|b` 回放碰撞、不同原生元组的公开 ID 碰撞，以及跨协议标点/长度降级碰撞。另已从公共 GLM Chat 复现单个 Unicode ID 在旧 Responses 目标上变空。均先记录失败场景与公开红例，再作最小来源限定修复。

GLM Responses 组合 ID 按组件单次转义 percent/分隔符，并在自身请求边界准确恢复 call_id；普通捕获 ID 不变。GLM 来源跨协议转换仅在最终 ID 冲突或为空时采用短关联 ID，完整历史保留一对一工具结果；其余旧历史与无 GLM 来源的规则保持不变。没有新增 DTO、环境凭据读取或收费调用。

实现 `709e6d0`、`c29cb4a`、`0f33a39`；已包含当时最新集成 `8839138`。已提交源 `0f33a39`：90 新边界场景通过，288 相关历史/工具/跨路线回归 PASS，124 race PASS，335 冻结 Responses 差分 PASS、零 pending，vet/diff check 通过。[失败计划、来源范围与哈希索引](../evidence/ticket-05-followup-ids/README.md)，[ADR-0034](../../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md)。原始 bundle 已持久保存到根 `.evidence/barness-ai-glm/ticket-05-followup-ids/`；边界 ID 为文档导出的离线合成数据，不冒充真实厂商观察。
