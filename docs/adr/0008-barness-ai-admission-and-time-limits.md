---
status: proposed
date: 2026-10-02
---

# barness-ai 按尝试准入，时限分调用级与尝试级执行

spec §9 要求每次尝试前获取准入许可、内置单租户与全进程并发、可注入按账户或分布式的准入，并要求建连、响应头、Provider 读空闲与调用总时限按最早者生效。ADR-0002 只规定这些数值由宿主显式给出，没有说明它们在调用中的执行位置、与重试（ADR-0006）的关系，以及超时在 pi 兼容终态中的分类。工单 13 实现时需要定下这些点，待维护者确认。

## 决策一：准入按尝试，位于初始请求的重试包装之内

每次尝试（含重试）在发送前先取内置许可，再取宿主注入的 `Admission` 许可；尝试失败时在退避等待之前归还，取得初始响应的尝试持有许可直到其响应流关闭、adapter 返回。准入发生在请求体构造（含 onPayload）之后，所以 onPayload 失败的调用不占许可；预检失败的调用从不调用准入。被拒绝的尝试不发送、不记入 `Attempts`，并以 `admission_denied`（`PhaseAdmission`）结束调用，不再重试。

### Considered Options

- 每次逻辑调用取一次许可、跨重试持有：实现简单，但退避期间空占许可，且与 spec“重试的每次尝试分别获取”不符。
- 在 adapter 之前取许可：符合 spec I3 的字面顺序，但重试无法逐次获取。
- 在重试包装内逐次获取（采用）：退避不占许可，宿主准入能按尝试计量与关联（`AttemptID`）。

### Consequences

- 等待准入时调用被取消或到期，按 pi 重试包装内的取消处理：`canceled`/`deadline_exceeded`、`PhaseAdmission`、StopReason aborted，文本为 “Request aborted” / “Request timed out.”。
- 重试的后续尝试被准入拒绝时，调用以 `admission_denied` 结束；前一次尝试的上游分类（如 `rate_limited` 与 Retry-After）仍记录在 `Attempts` 中，但不再是调用的错误类别。
- 宿主 `Admission` 的错误分类为 `admission_denied`，原错误只经 `errors.Is/As` 供宿主诊断，不进入消息；宿主授予许可后调用若已结束，许可立即归还。

## 决策二：内置准入只限本进程，等待有时限而非有条数上限

内置准入按 `MaxConcurrentPerTenant` 与 `MaxConcurrentProcess` 计数。满额时 `AdmissionWait` 为 0 立即拒绝，否则最多等待该时长；释放的许可按先来先得交给能放得下的最早等待者，因此一个租户占满自身配额时，其等待者不会挡住其他租户。宿主注入的 `Admission` 在内置许可之后叠加，`AdmissionWait` 只约束内置等待，宿主实现需自行限时并遵守 context。

### Considered Options

- 另设等待者条数上限：需要新的策略字段或隐式默认值（违反 D1），而每个等待者本就是宿主已发起、在有限时间内结束的调用。
- 仅按时限约束等待（采用）：不维护独立的准入队列，不预分配，等待者随调用结束而移除。

### Consequences

- 等待者数量受宿主自身发起的调用数约束；宿主若需要限制排队条数，应在进入 barness-ai 之前限流。
- 跨实例配额、按厂商账户聚合与费用预算由注入的 `Admission` 实现；`AdmissionRequest` 只携带 `TenantID`、`AccountScopeID` 与 `AttemptID`，均来自可信 scope 与已解析的绑定快照。

## 决策三：调用总时限走 context，尝试时限走传输包装

`CallTimeout` 作为调用 context 的截止时间，覆盖解析与整个流；它与宿主 deadline 谁早谁生效，结果与宿主 deadline 相同（`deadline_exceeded`，进入 adapter 后 StopReason aborted）。`ConnectTimeout`、`ResponseHeaderTimeout`、`ReadIdleTimeout` 与协议 `timeoutMs` 由包装 Client 传输的 `watchedTransport` 对每次 HTTP 往返执行，通过取消该次尝试自己的 context 立即终止 I/O，与 adapter 和 SDK 无关：

- 建连：从往返开始到传输经 `net/http/httptrace` 报告取得连接（GotConn）。
- 响应头：从往返开始到收到响应头；协议 `timeoutMs`（未设置时为 openai-node 默认 10 分钟）与策略值取较早者。
- 读空闲：只计每次 Read 等待上游数据的时间，回调、事件分发等读与读之间的时间不计入。

### Considered Options

- 用 openai-go 的 `WithRequestTimeout`：它的 context 也覆盖响应体读取，会在流中途截断，且只适用于该 SDK。
- 在 `http.Transport` 上设置拨号与响应头超时：宿主可注入自己的传输，barness-ai 不能改写。
- 包装传输并按尝试取消（采用）：对任何传输和 SDK 成立，超时原因可区分。

### Consequences

- 尝试级时限是 barness 对 pi SDK 超时的对应：以 `deadline_exceeded` 结束但 StopReason 为 error（与 openai-node 的 `APIConnectionTimeoutError` 在 pi 中的终态一致），不是 aborted。协议 `timeoutMs` 保留 openai-node 文本 “Request timed out.”；策略时限文本为 barness 自己的，指明字段与数值。
- 建连与响应头超时没有响应，按 pi 的连接错误形状可被绑定的重试策略重试（与 pi 对超时请求的重试一致，`PIDIFF-P01-E08-protocol-timeout-retried-stream`）；读空闲发生在流开始后，不重试（ADR-0006）。
- `timeoutMs` 为 0 视为未设置、取协议默认值，负数为 `invalid_request`；pi 中 0 会立即超时。显式 `timeoutMs` 时按 openai-node 发送 `X-Stainless-Timeout`（整秒截断）。
- 若宿主注入的传输不通过 httptrace 报告连接，`ConnectTimeout` 会一直计到响应头，此约束写在 `Config.Transport` 文档中。

行为与验收见 [spec §9](../../.scratch/barness-ai/spec.md#9-资源准入错误与观测)、工单 13 与 E08（`TestAdmission`、`TestTimeouts`、`PIDIFF-P01-E08-protocol-timeout-*`）。
