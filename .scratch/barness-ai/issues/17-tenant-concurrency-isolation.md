# 17: 租户并发隔离（E06）

**What to build:** 两个租户在同一 Client、同一 transport 上使用同名 binding/model/session 并发调用；一方鉴权失败、取消或超时不影响另一方，实际发送的 key、endpoint、结果和费用归属互不串用（spec T03、T05、T06、T08、User Stories 5、30）。

**Blocked by:** 03, 05, 11, 16

**Status:** ready-for-agent

- [ ] 服务端屏障强制 A/B 交错（不依赖概率碰撞或真实 sleep），分别在 Responses 与 Anthropic 上运行
- [ ] 捕获每个请求实际收到的 key 别名与 endpoint，与各自租户绑定一一对应
- [ ] A 发生 401、取消、超时、重试时 B 正常完成；B 的事件、Result、usage 不含 A 的内容
- [ ] 不从其他租户或环境变量兜底
- [ ] 结果元数据与观测的 TenantID/RequestID/AccountScopeID 归属正确
- [ ] 在 `-race` 下重复运行稳定
