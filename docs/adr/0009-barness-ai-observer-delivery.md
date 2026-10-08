---
status: accepted
date: 2026-10-02
---

# barness-ai Observer 异步有界投递，记录只含标识与分类

spec §9 要求 Observer 记录 CallStarted、AttemptStarted、AttemptFinished、CallFinished，失败不改变生成，且 key、正文、工具内容与原生密文不进入观测。spec 没有规定投递方式、拥塞时的处理、Observer 的失败形态，以及哪些字段算作“可观测”。工单 14 实现时定下以下几点，维护者于 2026-10-02 确认（含下文决策四至六）。

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

## 决策四：预检被拒的调用保留可信 TenantID，以 `Resolved=false` 区分

预检被拒（actor 被拒、绑定不属于该租户等）的调用，记录仍携带可信 scope 中的 TenantID，但 `Resolved=false`，不出现账户、Provider、模型与版本。scope 没有 TenantID 时记录中也为空。

### Considered Options

- 被拒调用不填 TenantID，或另设"声称的租户"字段：彻底避免误归属，但运维看不到"某租户被拒了多少次"这一安全信号。
- 保留 TenantID（采用，维护者 2026-10-02）：TenantID 是可信宿主的判断而非请求内容，拒绝原因在 actor 或绑定，不在租户；`Resolved=false` 已表明未获授权。

### Consequences

- 宿主统计用量、成功率与资源归属时只取 `Resolved=true` 的记录（`Observation` 文档已写明）。

## 决策五：Provider 回显的正文留在 ErrorMessage，登记为批准的安全差异

HTTP 错误的 ErrorMessage 与 `Error.Message` 按 pi 原样引用 Provider 错误体，只脱敏 key 形态的文本，所以 Provider 回显的提示词会保留在其中（E2E `P01-E09-redaction` 的 `echoed-error-message` 有记录）；它不进入观测，库也不写日志。spec I9 的“正文不进入错误”对此作出例外。

### Considered Options

- 不引用错误体：最安全，但开发者只剩笼统分类，排查困难，且与 pi 不一致。
- 消息保留原文、`err.Error()` 只含分类：堵住"顺手打日志"，但 `Error()` 文本与现有契约不同。
- 原样保留（采用，维护者 2026-10-02）：错误只返回给发起调用的租户，内容本属于该租户，不构成跨租户泄露；Observer 提供保证干净的日志渠道。

### Consequences

- `ErrorMessage`、`Error.Message` 与 `err.Error()` 属于租户内容，宿主不得写入共享日志；需要记录时使用 Observer 记录（`Error` 文档已写明）。
- 若日后出现宿主误记日志的事故证据，可改为上一条折中方案。

## 决策六：记录时间使用系统时钟

`Observation.Time` 与 `Duration` 使用系统时钟，与消息时间戳一致；重试的可替换时钟只覆盖重试所读的时间、抖动与退避等待（见 `internal/clock`）。在假时钟测试中，记录时长与 `Attempt.RetryDelay` 不对应；生产环境二者都是真实时间。维护者 2026-10-02 确认。

## 工单 05 已交付扩展（2026-10-08）

Observer 接收与操作无关的终态摘要；Classify 的 CallStarted、尝试与 CallFinished 均记录
classifier，始终复用原有白名单元数据和 Usage。状态、问题、答案、ResponseModel 和错误正文不进入观测。
拥塞仍只丢弃并计数，不改变分类结果；P07-E09 验证记录与证据脱敏审计覆盖这一点。

详见 [ADR-0021](0021-barness-ai-typesafe-unary-classification.md) 与 P07 离线证据。

## 工单 07：尝试的用量完整性轴（2026-10-08）

AttemptStarted 的 UsageReporting 明确为 unreported：尚未读取厂商用量，不能以空字符串
引入完整性枚举外的第四种状态。AttemptFinished 与 CallFinished.Attempts 使用相同的
unreported/partial/complete 轴，保持各自独立的公共快照。

P07 的新 E09 证据覆盖慢、拥塞、错误和 panic；队列故障只改变 ObserverStats，不改变
Classify 结果。观测不包含分类状态、问题、答案、认证头或正文，调用/尝试记录按可信入口和
固定快照关联；详见 [工单 07](../../.scratch/barness-ai-pi-1.0/unary-failures-evidence/README.md)。

## 工单 09：原生图像生成（2026-10-08）

图像操作记录 image；可选模态 token 分项是审计允许的元数据。结果、尝试与异步 Observer 各自拥有用量副本，不共享可变分项。提示、base64、revised prompt 和错误正文不进入观测。

详见 [ADR-0022](0022-barness-ai-openai-images-unary.md)。

## 工单 15：混合 Observer（2026-10-08）

同一个 Observer 接收三种入口的批准身份、版本、尝试和用量元数据，按调用不可变快照
归属。E09-mixed-observer 在慢、拥塞、返回错误和 panic 下同时运行四条协议路线及
一次模型失败，模型成功/失败保持，丢弃/失败由既有 ObserverStats 识别。未增加正文
白名单：提示、base64、状态、问题、答案、概率/置信度和 Provider 错误正文均不得进入
观测。每次写出后的 evidence audit 复核所有 observations 文件。

[混合操作证据](../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/README.md)。
