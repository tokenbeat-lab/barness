---
status: accepted
date: 2026-10-02
---

# barness-ai 按尝试准入，时限分调用级与尝试级执行

spec §9 要求每次尝试前获取准入许可、内置单租户与全进程并发、可注入按账户或分布式的准入，并要求建连、响应头、Provider 读空闲与调用总时限按最早者生效。ADR-0002 只规定这些数值由宿主显式给出，没有说明它们在调用中的执行位置、与重试（ADR-0006）的关系，以及超时在 pi 兼容终态中的分类。工单 13 实现时定下这些点，维护者于 2026-10-02 确认（含决策二的等待人数上限、`timeoutMs` 为 0 视为未设置、重试被拒时以 `admission_denied` 为调用错误）。

## 决策一：准入按尝试，位于初始请求的重试包装之内

每次尝试（含重试）在发送前先取内置许可，再取宿主注入的 `Admission` 许可；尝试失败时在退避等待之前归还，取得初始响应的尝试持有许可直到其响应流关闭、adapter 返回。准入发生在请求体构造（含 onPayload）之后，所以 onPayload 失败的调用不占许可；预检失败的调用从不调用准入。被拒绝的尝试不发送、不记入 `Attempts`，并以 `admission_denied`（`PhaseAdmission`）结束调用，不再重试。

### Considered Options

- 每次逻辑调用取一次许可、跨重试持有：实现简单，但退避期间空占许可，且与 spec“重试的每次尝试分别获取”不符。
- 在 adapter 之前取许可：符合 spec I3 的字面顺序，但重试无法逐次获取。
- 在重试包装内逐次获取（采用）：退避不占许可，宿主准入能按尝试计量与关联（`AttemptID`）。

### Consequences

- 等待准入时调用被取消或到期，按 pi 重试包装内的取消处理：`canceled`/`deadline_exceeded`、`PhaseAdmission`、StopReason aborted，文本为 “Request aborted” / “Request timed out.”。
- 重试的后续尝试被准入拒绝时，调用以 `admission_denied` 结束：调用的错误表达“最后为什么停下”，前一次尝试的上游分类（如 `rate_limited` 与 Retry-After）保留在 `Attempts` 中表达“经过了什么”（维护者 2026-10-02 确认；不改报前一次错误，也不在同一错误上混入两个来源的字段）。
- 宿主 `Admission` 的错误分类为 `admission_denied`，原错误只经 `errors.Is/As` 供宿主诊断，不进入消息；宿主授予许可后调用若已结束，许可立即归还。

## 决策二：内置准入只限本进程，等待在时长与人数上都有界

内置准入按 `MaxConcurrentPerTenant` 与 `MaxConcurrentProcess` 计数。满额时 `AdmissionWait` 为 0 立即拒绝，否则最多等待该时长；同时等待的尝试数达到 `MaxAdmissionWaiters`（必填正数）时，新尝试不排队、立即以 `admission_denied` 拒绝，错误指明该字段。释放的许可按先来先得交给能放得下的最早等待者，因此一个租户占满自身配额时，其等待者不会挡住其他租户。宿主注入的 `Admission` 在内置许可之后叠加，`AdmissionWait` 与 `MaxAdmissionWaiters` 只约束内置等待，宿主实现需自行限时并遵守 context。

### Considered Options

- 仅按时限约束等待：不增加配置，但突发请求时等待者数量随请求量增长，与 spec“不建立无限准入队列”的字面要求不符，且依赖宿主在外层限流。
- 增设全进程等待人数上限（采用，维护者 2026-10-02）：显式、有限的宿主策略，与 D1 一致；宿主漏配外层限流时仍有兜底。
- 按租户的等待人数上限：能防止单个租户占满所有等待位，但首期没有该场景的证据，暂不引入；出现后再以新字段扩展。

### Consequences

- 新增策略字段 `MaxAdmissionWaiters`，零或负数构造失败；宿主示例（工单 18）需给出数值依据。
- 一个租户的突发请求可能占满全部等待位，使其他租户在满额时无法排队（仍可在有空闲许可时直接获得）；若成为问题，按上一条增加按租户上限。
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
- `timeoutMs` 为 0 视为未设置、取协议默认值，负数为 `invalid_request`；pi 中 0 会立即超时。这是有意的差异（维护者 2026-10-02 确认）：该输入没有实际用途，不为区分“未填”与“填 0”改用可空类型；不进入 pi 差分场景。显式 `timeoutMs` 时按 openai-node 发送 `X-Stainless-Timeout`（整秒截断）。
- 若宿主注入的传输不通过 httptrace 报告连接，`ConnectTimeout` 会一直计到响应头，此约束写在 `Config.Transport` 文档中。

行为与验收见 [spec §9](../../.scratch/barness-ai/spec.md#9-资源准入错误与观测)、工单 13 与 E08（`TestAdmission`、`TestTimeouts`、`PIDIFF-P01-E08-protocol-timeout-*`）。

## 工单 05 已交付扩展（2026-10-08）

聊天与 Classify 共用调用运行时、唯一 HTTP 客户端和传输时限。unary 许可持有到成功 body 关闭、
读取与整体校验完成；响应回调只读 HTTP 元数据，在读取前执行一次。调用取消/总时限在 response 阶段形成
aborted，读空闲等尝试时限仍形成 error；所有失败路径释放 body 与许可。

详见 [ADR-0021](0021-barness-ai-typesafe-unary-classification.md) 与 P07 离线证据。

## 工单 07：unary 生命周期验收（2026-10-08）

Classify 在 scope 验证后、任何可变输入复制或配置读取前检查 context：已取消或已到期的
调用在 scope 阶段以 error 结束，只保留可信 scope、请求绑定和入口操作。
解析期间 context 结束（包括后端成功返回）在对应 binding/credential 阶段以 error 结束。
准入等待/退避/成功体读取期间的调用取消或总截止时间分别在 admission/request/response
以 aborted 结束；连接、响应头和读空闲尝试时限仍以 error 结束。
响应回调属于既有 request 回调契约，失败或 context 结束保留该阶段与可信宿主错误链。

TestUnaryAdmissionWaitIsolation 证明取消等待调用只回收自己的等待者，已持有响应的调用
和其他租户继续执行；注入准入在 context 结束后返回的许可立即归还、不伪造 Attempt。
读取、整体解析与校验、body 关闭均在尝试许可下完成，后续调用可继续获得许可。
详见 [工单 07](../../.scratch/barness-ai-pi-1.0/unary-failures-evidence/README.md)。

## 工单 09：原生图像生成（2026-10-08）

GenerateImages 复用共同 unary 发送与绑定重试、逐次准入及取消/时限；许可持有到图片解析结束和 body 关闭，回调 panic 解栈也释放。P08 E08 记录 body/call/permit/waiter/event 全部归零。

详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。

## 工单 15：混合准入（2026-10-08）

所有操作共用租户/进程许可和等待者上限，不按入口各自开池。E08-mixed-admission
以不同操作占有者/等待者验证租户上限、进程上限、人数和等待时限；取消等待者不
归还占有者许可，有剩余进程容量时其他租户仍可用。unary 完整 JSON 读到 EOF、
解析并全部校验后才结束尝试；已读完整 body 但 Provider 未关闭时仍持有许可。
E08-mixed 另验证各入口取消/截止/HTTP 失败只回收自己资源，后续调用可重获许可。
混合策略的 4/8 并发与两个等待位针对声明的图像负载，跨实例/账户限额仍由宿主注入。

[混合操作证据](../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/README.md)。
