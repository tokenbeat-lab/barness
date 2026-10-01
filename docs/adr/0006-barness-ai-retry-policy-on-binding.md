---
status: accepted
date: 2026-10-02
---

# barness-ai 的重试策略随绑定配置，不作为请求选项

pi 把 `maxRetries`、`maxRetryDelayMs` 放在每次调用的 StreamOptions 里，调用方可以随意设置。spec §8 和 User Story 31 要求由运维人员显式配置重试次数与延迟，默认不重试；重试会重复计费，花的是厂商账户的钱，云端请求不能自行调高。所以 barness-ai 不把这两项放进 `ResponsesOptions`/`SimpleOptions`，而是作为 `Binding.Retry`（`RetryPolicy`）随绑定配置。

## Considered Options

- 请求选项（与 pi 一致）：最直接，但不可信请求可以把重试次数调高，增加未知费用，违背 spec §3“完整选项不得越过授权边界”的要求。
- Client 级配置：可信，但所有租户与账户只能共用一个值；不同厂商账户的限流和费用承受能力不同，做不到分别配置。
- 绑定配置（采用）：绑定由可信宿主按租户解析，本来就固定厂商账户和凭据。重试策略随同一份快照固定（ADR-0003），在途调用的所有尝试使用解析时的值；宿主可在自己的解析器中套用全局默认值。

## Consequences

- 零值即 pi 的默认值：`MaxRetries` 为 0，不重试；`MaxRetryDelay` 为 nil 时上限 60 秒，显式设为 0 只取消这一项上限，等待仍受调用的 context 截止时间约束（ADR-0002）。负数是宿主配置错误，调用在 PhaseBinding 以 `invalid_request` 失败，不发出请求。
- 判定和延迟逐条移植 pi 的 `provider-retry`：x-should-retry 优先；没有响应的连接失败，以及 408/409/429/5xx 才重试；延迟依次取 retry-after-ms、retry-after（秒数或 HTTP 日期）、带抖动的指数退避。SDK 自身的重试始终关闭。只在各 adapter 的初始请求处重试，流开始后从不重放；onPayload 每个逻辑调用执行一次，onResponse 只在取得初始响应后执行一次。
- “没有响应的连接失败”只认 HTTP 客户端发送失败的错误（Responses adapter 用 SDK middleware 标记），不把任意 Go error 当作可重试。
- 每次尝试记录在 `CallMetadata.Attempts`，AttemptID 为 `RequestID#序号`。Observer（工单 14）以同样的粒度上报尝试。
- 时钟与抖动源是内部类型 `internal/clock.Clock`，仅供验收测试替换（`Config.Clock`），宿主无法构造。
- pi 的 retry 包装在请求被取消时总是抛出自己的 AbortError，所以初始请求阶段取消的 errorMessage 为 `Request aborted`，与是否配置重试无关；barness 随之修正了工单 05 采用的 SDK 文本 `Request was aborted.`。截止时间到期仍用 `Request timed out.`，这是 barness 的扩展（pi 没有调用截止时间）。
- 已知的两处细微差异，均不进入 pi 差分场景：一是 `MaxRetryDelay` 为 0 时，超过 2^31-1 毫秒（约 24.8 天）的请求延迟在 pi 中被 Node 的 setTimeout 当作立即触发，barness 则按规范一直等到调用截止时间；二是 `retry-after` 日期只按 HTTP 日期格式（RFC 9110）解析，JavaScript `Date.parse` 能接受的其他自由格式日期在 barness 中按无法解析处理，立即重试。
- 初始请求在 context 结束时，无论该次尝试拿到什么（包括同时到达的 HTTP 错误），都按中断分类，与 pi 先检查 signal 的顺序一致。
- pi 差分中重试场景两侧使用相同设置：barness 用绑定，pi 用调用选项。这是配置位置的差异，不是行为差异，所以不登记差分条目。

行为与验收见 [spec §8](../../.scratch/barness-ai/spec.md#8-流终态取消与重试) 和 E05（`TestExplicitRetry`、`PIDIFF-P01-E05-*`）。
