# 08: 让三协议的失败、重试和用量可解释

Status: claimed
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 在 GLM 认证、额度、暂时限流、断流或取消时得到准确错误和已知用量；显式重试只用于可恢复的初始请求，缺失用量不会被当成免费。

Blocked by: 04, 06

依据：GLM 接入 spec R17、R18、R19；验收组 A09、A10；User Stories 36、37、43、44、45、48、49。

- [x] 先为三协议、两型号写公共入口 HTTP/流内错误、usage presence 与重试失败场景，再实现厂商错误解释；只依赖基础文本调用，不等待工具功能。
- [x] 分别读取 Responses 字符串业务码、Chat 数字业务码与 Messages 实际错误形状，保持既有稳定 Code/Phase 与 HTTPStatus，不跨协议套用错误体。
- [x] 已知欠费、额度耗尽、套餐到期、型号权益不足即使 HTTP 429 也不重试；暂时限流/过载只按 Binding 的显式有界策略重试；未知 429 不被猜成具体账务错误。
- [x] 默认不重试，SDK 重试关闭；全部 Attempt 使用同一凭据和最终 payload，流开始后零重放；不靠换账号/端点/型号完成重试。
- [x] Chat 的 sensitive/network_error/model_context_window_exceeded、HTTP 200 流内 error 和缺终态 EOF 不能变成普通成功 stop。
- [x] 正常、零、缺项、缓存、推理明细和中断用量按各协议 presence 处理；Reasoning 属于 Output 不重复计数，Input 不重复计算缓存，缺失是 partial/unreported。
- [x] 保留既有 Responses failed-response 用量约定，不借 GLM 接入改变已有 Provider 的取值；流中断保留实际已知用量。
- [x] 从真实 header/错误体证据确认 request id 来源；缺失不伪造，不把本地 RequestID 或 DeepSeek trace 头当作 GLM 已返回 ID。
- [x] 不可安全主动触发的账务/限流错误使用有来源的离线 fixture；产出各协议错误、Attempt、用量、重试计数和脱敏断言证据。

**独立验收：** 对三协议重放同类初始 429/断流/缺失 usage，能区分不可恢复额度错误与有限重试，并展示实际 Attempt 和已知成本输入。

**依赖理由：** 错误和用量需要三个基础路线已经能通过公共 Client 发起/读取；工具与历史增强不构成本票前置条件。

