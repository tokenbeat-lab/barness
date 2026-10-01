# 10: 可信回调：transformHeaders / onPayload / onResponse（E04 回调部分）

**What to build:** 可信宿主为单次调用提供 header 变换、请求体回调和响应元数据回调，在规定时点观察或修改请求以完成协议扩展；回调与可反序列化的 Request 隔离，不能绕过绑定授权（spec I7 后半部分、User Story 35）。

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] transformHeaders 在认证头与请求头合并之后、交给 adapter 之前执行；onPayload 在 adapter 构建原生请求体之后执行
- [ ] onPayload 支持观察、原位修改、返回替换对象；Go 类型明确区分“不替换”与“替换”，不使用含义不清的 nil；“不替换”不撤销原位修改
- [ ] Responses 的 onPayload 位于初始请求重试包装之外（每逻辑调用一次，重试下的次数断言见 11）；onResponse 仅在成功取得初始响应后、start 前执行一次，读取 HTTP status/headers 与模型信息，不接收或替换最终消息
- [ ] 回调失败或超出 context 时进入对应 Phase 的错误终态
- [ ] 变换后的最终认证与目标仍须满足绑定授权；改写 Authorization 或目标 host 的尝试被拒绝且不发往未授权目标
- [ ] transport 只在 Client 构造时装配，不存在每调用 transport 覆盖入口
- [ ] 每次调用固定本次回调引用并显式传入 scope；两个租户并发调用时回调不串用
