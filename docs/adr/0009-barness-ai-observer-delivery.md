---
status: proposed
date: 2026-10-02
---

# barness-ai Observer 异步有界投递，记录只含标识与分类

spec §9 要求 Observer 记录 CallStarted、AttemptStarted、AttemptFinished、CallFinished，失败不改变生成，且 key、正文、工具内容与原生密文不进入观测。spec 没有规定投递方式、拥塞时的处理、Observer 的失败形态，以及哪些字段算作“可观测”。工单 14 实现时定下以下几点，待维护者确认。

## 决策一：每个 Client 一个有界队列，按需启动单个投递 goroutine，满则丢弃并计数

记录在调用的生产 goroutine 中入队，从不阻塞；队列长度由新的策略字段 `ResourcePolicy.MaxQueuedObservations` 约束（配置 Observer 时必须为正，否则构造失败；未配置时 0 合法）。有记录等待且无人投递时启动一个 goroutine，按入队顺序逐条调用 `Observe`，队列清空即退出，所以 Client 不需要 Close，也不常驻 goroutine。队列满时新记录丢弃，`Client.ObserverStats()` 分别给出已投递、失败与丢弃的计数。

### Considered Options

- 同步调用 Observer：最简单，但慢 Observer 直接拖慢或阻塞调用，违反 spec。
- 每条记录一个 goroutine：不阻塞，但并发数无界、顺序丢失。
- 常驻 goroutine + 有界 channel：需要 Client.Close 才能回收，现有 API 没有生命周期。
- 按需启动的单个投递者（采用）：有界、保序、无常驻资源。

### Consequences

- 一个永不返回的 Observer 会让该 goroutine 一直存在，之后的记录全部丢弃并计数；这是宿主的缺陷，库只保证有界。
- 丢弃会使某次调用的记录不完整（例如有 CallFinished 而无 CallStarted）；宿主应依据 `Dropped` 判断观测是否完整，而不是依据记录本身。
- 记录在 Result 返回前入队，但在之后投递；宿主不能假设 `Complete` 返回时 Observer 已经收到 CallFinished。

## 决策二：`Observe` 返回错误或 panic 均计为失败，panic 被拦截

`Observe(Observation) error` 的错误只计入 `Failed`。Observer 在 Client 的投递 goroutine 上运行，未拦截的 panic 会终止整个宿主进程，等于“Observer 失败改变了生成”，因此投递时 recover 并计为失败。

### Consequences

- panic 不会再以崩溃暴露；宿主需要监控 `Failed` 计数。

## 决策三：记录只携带标识、分类、时间与用量

记录复用 `CallMetadata`（租户、RequestID、Actor/Job、绑定 ID 与版本、凭据版本、账户、真实 Provider/API/model、原生状态降级计数、尝试列表）与 `Attempt`，CallFinished 另含 StopReason、`ObservedError`（Code、Phase、HTTP status、厂商 request ID、Retry-After）与 Usage。错误消息文本不进入记录：它按 pi 兼容规则可能引用 Provider 的错误体（库只对 key 形态的内容脱敏），可能含有正文。`CallMetadata` 因此新增 `BindingVersion` 与 `CredentialVersion`，在快照一致后设置，Result 同样可见。

- CallStarted 在任何解析之前发出，只含可信 scope 与请求的 BindingID，`Resolved=false`；预检被拒的调用其后所有记录仍为未解析，不出现账户、Provider、模型或版本，因此不会被归属到已授权租户的快照。
- AttemptStarted 在尝试取得准入许可之后、发送之前发出；准入被拒或预检失败的调用没有尝试记录。
- 取得初始响应的尝试在其响应流关闭（adapter 返回、许可归还之前）时发出 AttemptFinished，其 `Code` 仍按 `Attempt` 的定义为空；流中途的失败由 CallFinished 的 `Error` 表达。

### Consequences

- TenantID、RequestID 等为无界取值，Observer 文档提示宿主默认不要用作指标标签；库本身不产生指标，也不写日志。

## 待维护者确认

1. **被拒身份的租户归属。** 预检被拒（含 actor 被拒、绑定不属于该租户）的调用，记录仍携带可信 scope 中的 TenantID，但 `Resolved=false`，且不出现账户、模型和版本。宿主若按 TenantID 汇总，被拒调用会计入该租户名下。备选方案是被拒调用不填 TenantID，或另设“声称的租户”字段。
2. **Provider 回显正文进入 ErrorMessage。** 按 pi 兼容规则，HTTP 错误的 ErrorMessage 会引用 Provider 的错误体，库只脱敏 key 形态的文本；Provider 回显的提示词会留在 ErrorMessage 与 `Error.Message` 中（E2E `P01-E09-redaction` 的 `echoed-error-message` 有记录），但不进入观测与日志。spec 要求正文不进入错误，因此这里需要决定：登记为批准的安全差异，还是改为不引用错误体（与 pi 不一致）。
3. **记录时间使用系统时钟。** 与消息时间戳一致，不使用重试的替换时钟；在假时钟场景下，记录的 Duration 与 `RetryDelay` 不对应。
