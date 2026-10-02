---
status: accepted
date: 2026-10-02
---

# barness-ai 的总输出按读取的响应流字节计量，超限终态不重试

ADR-0002 要求宿主为请求、图片、帧、工具 JSON、错误体和总输出等提供有限限额，并规定超限产生 resource_limit 终态，但没有说明“总输出”按什么计量，也没有说明超限终态与重试（ADR-0006）的关系。工单 12 实现限额时需要定下这两点，维护者于 2026-10-02 确认。

## 决策一：`MaxOutputBytes` 计量一次调用读取的响应流字节

`MaxOutputBytes` 计量一次调用成功响应的流式响应体在传输解码后的字节数，每一帧都计入，而不只是助手消息的文本。

### Considered Options

- 消息内容字节（文本、thinking、工具参数、签名）：数值直观，配置多少就允许多少输出。但协议开销和不增长内容的事件（大量空块、未知事件、重复的 done 内容）不受它约束，只能靠单帧限额和调用时限（工单 13）兜底；按块累计还要在每个 adapter 的合并路径上分别记账。
- 响应流字节（采用）：在读取阶段由同一个包装器检测，与协议和 adapter 无关，能约束调用读取并可能保留的一切，异常 Provider 无法绕过。

### Consequences

- OpenAI Responses 的终态帧（`response.completed`）会重复整个响应，所以同一策略实际允许的文本约为配置值的一半；`MaxFrameBytes` 也必须容纳宿主接受的最大响应。两点都写在 `ResourcePolicy` 的字段文档里，宿主示例（工单 18）给出数值时须一并说明。
- 校验规则 `MaxToolJSONBytes ≤ MaxOutputBytes` 比较的是工具参数字节和流字节，作为宽松上界保留：单个工具参数不可能超过承载它的流。
- 错误状态的响应体不计入 `MaxOutputBytes`，而由 `MaxErrorBodyBytes` 按每次尝试单独限制。
- 以后若需要“消息文本”层面的上限，可以另增字段，不改变本字段的含义。

## 决策二：达到资源限额的尝试不重试

达到任一资源限额的失败都是 barness 自己的终态，不按 ADR-0006 的规则重试。实际只涉及初始请求阶段的 `MaxErrorBodyBytes`：429/5xx 等可重试状态的错误体超限时，调用以 resource_limit（PhaseRequest）结束，HTTP status、厂商 request ID 和 Retry-After 均保留。

### Considered Options

- 按状态码照常重试（与 pi 一致）：读取有上限，重试在内存上是安全的。但每次重试都要再读一个达到上限的错误体，且这次失败的诊断信息（错误体内容）已经丢失。
- 不重试（采用）：符合 spec I9“超限以 resource_limit 终态结束”；会发超大错误体的服务很可能再次这样做，重试只会多读几遍。

### Consequences

- 与 pi 不同：pi 不限制错误体，会按状态码重试。该差异作为扩展登记在 pi 差分 ledger（`PIDIFF-P01-E08-error-body-over-limit-stream`），由离线 E2E `TestByteLimits/error-body` 断言（配置 `MaxRetries: 2` 时仍只发出一次请求）。
- 限额以内（错误体恰好等于上限）的行为与 pi 一致，照常分类和重试。
- 流开始后的限额（帧、总输出、工具 JSON、事件队列）本来就不重试（ADR-0006 不重放已开始的流），不受本决策影响。

行为与验收见 [spec §9](../../.scratch/barness-ai/spec.md#9-资源准入错误与观测)、工单 12 与 E08（`TestByteLimits`、`TestEventQueueLimits`、`TestResourceRelease`、`PIDIFF-P01-E08-*`）。
