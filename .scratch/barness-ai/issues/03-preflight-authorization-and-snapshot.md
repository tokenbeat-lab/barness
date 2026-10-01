# 03: 授权预检、凭据解析与配置快照（E07 / D2）

**What to build:** 缺身份、越权、凭据不可用或配置不一致的调用在发出任何 Provider 推理请求前以错误终态结束；新调用取得更新或撤销后的凭据状态，在途调用固定使用原快照。覆盖 spec I3、T01–T05、T10–T11 与 ADR-0003（D2）。

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] 调用顺序按 spec I3：scope/结构/大小检查 → 绑定解析 → 能力/选项/历史来源检查 → 凭据解析 → 租户/账户/引用/版本一致性校验 → 准入 → 发送
- [ ] 缺 TenantID 或 RequestID、binding 不存在或不属于该租户、调用主体无权、模型不在 AllowedModels ∩ 目录、binding 禁用、缺 key、秘密后端故障：均以错误 assistant 消息结束，Code 分别可用 errors.Is/As 区分（invalid_request / tenant_denied / binding_not_found / credential_unavailable 等），本地 Provider 收到零推理请求
- [ ] 以上拒绝在 full/simple × Stream/Complete 四个入口均验证，不存在绕过解析器的网络入口
- [ ] 两次解析间发生更新或撤销且无法取得一致快照时直接失败，Phase 指明配置一致性阶段，无内部重解析，错误不泄漏秘密；宿主用新 RequestID 再调用可见新的一致快照
- [ ] K1→K2 更新后新调用使用 K2；撤销阻止后续新调用；在途调用继续使用原快照
- [ ] 普通 Request 不能携带 key、凭据引用、endpoint、Authorization、代理或可执行函数（类型层面不提供该字段，或在边界拒绝）
- [ ] 测试环境设置污染的 OPENAI_API_KEY / ANTHROPIC_API_KEY / GOOGLE_API_KEY 及 endpoint 变量，证明无隐式回退或误出网
- [ ] 预检失败的 Result 中未解析出的 Provider/模型/账户标记为未解析，不使用请求自报值
- [ ] 秘密值不出现在消息、事件、Result、错误文本中
