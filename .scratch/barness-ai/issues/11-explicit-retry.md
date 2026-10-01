# 11: 显式重试（E05）

**What to build:** 运维人员显式配置重试次数、延迟与截止时间后，可重试的初始请求失败按冻结 pi 规则有界重试；默认不重试，流开始后从不重放，所有尝试固定同一配置快照并分别记录（spec I8 后三条、User Stories 31–32）。

**Blocked by:** 05

**Status:** resolved

- [x] 默认 maxRetries=0；SDK 自身不产生额外重试（服务端计数证明）
- [x] 判定顺序：x-should-retry 优先，再按基线网络错误形状及 408/409/429/5xx；不把任意 Go error 当可重试；401/403 不重试也不换身份
- [x] 延迟依次采用 retry-after-ms、retry-after 秒数/HTTP 日期、指数退避（500ms 起、封顶 8s、最多下浮 25%）；maxRetryDelayMs 默认 60s，设 0 只关闭单项上限，仍受调用截止时间约束
- [x] 退避等待可被取消/截止时间打断；时钟与退避随机源在最小控制点可替换，用例不依赖真实 sleep
- [x] 只在初始请求位置重试；首帧后断流以错误终态结束，不重放、不产生重复内容或工具调用
- [x] 每次尝试有独立 AttemptID，沿用同一 RequestID、租户、绑定版本、凭据与账户；中途更新凭据不影响剩余尝试
- [x] onPayload 在重试下仍每逻辑调用一次；onResponse 只在最终成功取得的初始响应后执行一次
- [x] 重试次数与延迟序列可重复，接入 pi 差分无待处理差异

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/retry_test.go`, fixture `testdata/responses/retry.json`: 21 scenarios on all four entries, plus repeatability, cancellation and deadline during backoff, a K1→K2 rotation between attempts, and invalid policies; differential `PIDIFF-P01-E05-*`, 18 cases).

- Configuration: `Binding.Retry` (`RetryPolicy{MaxRetries, MaxRetryDelay *time.Duration}`), not a request option. Accepted by the maintainer on 2026-10-02 (ADR-0006). The zero value never retries; a nil MaxRetryDelay means 60s and 0 lifts the cap, with the wait still bounded by the context. Negative values fail in PhaseBinding as invalid_request.
- Core: `ai/retry.go` ports pi's `provider-retry` (decision order, delay sources, the "Server requested Ns retry delay (max: Ms). <SDK message>" refusal, and NaN → no wait). Adapters call `adapterCall.initial.send` at their initial-request point only. A connection failure is retryable only when the SDK middleware marked an HTTP client error; other Go errors are not.
- Attempts: `CallMetadata.Attempts`, with AttemptID `RequestID#n`, status, vendor request id, code and planned delay. Observer events remain ticket 14's.
- Clock: internal `ai/internal/clock` (`Config.Clock`, which hosts cannot construct), so no case sleeps for real. `Error.RetryAfter` now converts an HTTP-date retry-after, the item ticket 05 had deferred.
- Parity fix from verification against frozen pi: a cancel during the initial request is `Request aborted`, pi's retry-wrapper AbortError, not the SDK's `Request was aborted.` (this updates TestFailureTerminals). When the context ends, the attempt is classified as an interruption whatever response arrived.
- Ledger: the "Connection lost" runtime-text extension now also lists `PIDIFF-P01-E05-stream-cut-not-replayed-stream`. The remaining pending items are ticket 15's usage.cost and usage.reasoning. Known small divergences (setTimeout overflow, Date.parse leniency) are recorded in ADR-0006.
- Mutation-checked: jitter, 409 retryability, x-should-retry false, the cap, off-by-one in MaxRetries, the abort text, connection marking, date parsing, and attempt recording each fail the suite.
