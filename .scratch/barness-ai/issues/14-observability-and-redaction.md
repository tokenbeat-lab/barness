# 14: 观测与秘密脱敏（E09）

**What to build:** 运维人员通过 Observer 按租户、逻辑调用和请求尝试追踪状态、重试、取消与用量归属；观测异步有界，拥塞或失败不改变生成结果；key、正文、工具内容与原生密文不进入默认日志、观测和错误（spec I9 后半、User Stories 37–39）。

**Blocked by:** 11

**Status:** ready-for-agent

- [ ] 事件 CallStarted / AttemptStarted / AttemptFinished / CallFinished；调用级关联 TenantID/RequestID，尝试级另含 AttemptID 与厂商 request ID
- [ ] 记录已解析的 BindingID、AccountScopeID、配置/凭据版本、真实 Provider/API/model、时间、错误类别、已知 usage；ActorID/JobID 提供时保留
- [ ] 预检失败不伪造 AttemptStarted；无效身份的拒绝事件不归属到已授权租户
- [ ] 07 的原生状态降级计数与原因同时出现在 Observer
- [ ] 慢或失败的 Observer 不阻塞、不改变结果；丢失计数可查询
- [ ] 成功、预检拒绝、失败、重试、中断、Close 各路径的 Call/Attempt 关系准确
- [ ] 以含 key、Authorization、正文、工具参数、原生密文的场景（含敏感错误体）扫描日志、观测、错误文本：无泄漏；授权 Result 中原生状态仍完整
- [ ] 原始 TenantID 不进入厂商 payload，也不作为默认无界指标标签
