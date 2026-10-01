# 05: 失败终态、取消与错误分类（E02 / P01 终态）

**What to build:** 设置失败、鉴权失败、限流、服务端错误、流内协议错误、异常断流、截断、取消和超时都返回完整的最终 assistant 消息，StopReason 与带 Code/Phase 的分类错误区分正常停止、截断、错误和取消，已收到的部分内容和 usage 不丢失（spec I8、I9 错误部分、E02、P01 终态）。

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] start 前设置失败可直接 error 事件；建立 HTTP 失败产生错误 assistant 消息
- [ ] 401/403 → upstream_auth、429 → rate_limited、5xx、transport、protocol 各自分类；保留可安全公开的 HTTP status、厂商 request ID、Retry-After；401/403 不触发换 key 或身份兜底
- [ ] Responses 分别处理 response.completed / incomplete / failed；流内 error、半个 JSON、无协议终态的 EOF 均为 StopReason=error，SDK 无错误 EOF 不得视为成功
- [ ] 正常 length 为明确截断终态，未完成的工具参数不被标记为可执行
- [ ] StopReason 为 error/aborted 时 Complete / Result 返回非 nil error 且 Result 完整有效，Err() 与之相同；成功终态 error 为 nil
- [ ] 取消阶段语义保留基线：lazy setup 期间取消形成 error，进入 adapter 后形成 aborted；附加分类区分 canceled 与 deadline_exceeded
- [ ] Close 未完成时取消本次设置或 I/O，已完成时不改写结果；所有路径关闭响应体并释放等待者，不依赖再次 Next
- [ ] 流中断保留部分内容、usage 与 errorMessage；涉及秘密的 errorMessage 脱敏并登记差异
- [ ] 适用用例接入 02 的 pi 差分，无待处理差异
