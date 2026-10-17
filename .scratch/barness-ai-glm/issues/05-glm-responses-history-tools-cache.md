# 05: 完成 GLM Responses 无状态工具往返与缓存提示

Status: ready-for-agent
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 以完整历史完成 GLM Responses 函数工具往返，并安全使用按租户和账户隔离的缓存路由提示；未接入的服务端能力在发送前明确拒绝。

Blocked by: 04

依据：GLM 接入 spec R08、R09、R10、R15、R16；验收组 A04、A07、A13；User Stories 24、25、26、27、28、30、34、35、40、41。

- [ ] 先写两型号的完整文本/原生 reasoning/函数调用→工具结果→下一轮回答及失败 E2E，再实现历史转换。
- [ ] 每轮完整回放 input，不使用 store=true、previous_response_id 或 conversation；GLM InputReasoning 形状正确，内部数组规范化不污染原生回放。
- [ ] 函数参数增量、权威参数 done、call_id/条目 ID 与 function_call_output 正确配对；不因跨型号或来源降级制造孤立工具结果。
- [ ] 仅默认、none、auto 工具选择可用；required、具名强制、allowed_tools、strict 及本期未接入托管工具不能通过具名选项、sampling 或 payload hook 偷渡。
- [ ] 有 SessionID 且 retention 非 none 才发送按既有租户/账户/session 派生的 prompt_cache_key；不发送原始身份/session、TTL、retention 或 OpenAI 亲和头。
- [ ] short/long 都只作为路由提示；无 session 或 none 不派生 key，不承诺命中率/期限；非法自报缓存键仍按授权契约拒绝。
- [ ] 显式 summary、ServiceTier、encrypted include、服务端状态引用等按 spec 拒绝；合法回调保留原生 payload 与既有顺序，发送前失败零 Attempt。
- [ ] 产出两型号完整无状态往返、缓存隔离和非法字段拒绝的可重放证据；同模型可信回放与未担保/跨账户降级均有公共入口示例。

**独立验收：** 不保存服务端会话地完成两型号工具往返，捕获下一轮 GLM 原生 reasoning 与正确隔离的缓存 key，并验证非法强制工具/状态引用没有请求。

**依赖理由：** 需要 04 的 Responses 调用、content 解析和原始 reasoning 状态。

