# 13: 并发准入与超时（E08 准入与时限部分）

**What to build:** 云端宿主使用内置的单租户与全进程并发限制，并可注入按厂商账户聚合或分布式的准入实现；建连、响应头、Provider 读空闲与调用总时限按最早者生效；所有路径释放许可与连接（spec I9 第二至四段、User Stories 29–30、42）。

**Blocked by:** 05

**Status:** resolved

- [x] 每次尝试前获取准入许可，结束、失败、取消、Close 时释放；重试的每次尝试分别获取
- [x] 内置单租户与全进程并发上限；达到上限时有界拒绝（admission_denied）或有限时等待，不建立无限准入队列；等待可取消
- [x] 准入接口参数包含 TenantID 与 AccountScopeID；注入的替身实现验证收到的值，且预检失败的调用不调用准入
- [x] 建连、响应头、读空闲、调用总时限分别可触发；宿主 deadline、策略与协议 timeoutMs 按更早者结束，错误区分 deadline_exceeded；读空闲只计上游等待
- [x] Close 并发/重入、下游断开（宿主取消 context）均终止本次 I/O 并释放许可与等待者
- [x] 一个租户占满自身配额时另一租户仍可获得许可
- [x] 资源探针证明许可计数归零、活动请求归零、goroutine 有界收敛

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/admission_test.go`, `timeouts_test.go`; differential `PIDIFF-P01-E08-protocol-timeout-*`, 3 cases). The design decisions are recorded in ADR-0008.

- **Admission** (`ai/admission.go`). The public `Admission` interface is `Admit(ctx, AdmissionRequest{TenantID, AccountScopeID, AttemptID}) (release func(), error)`, injected as `Config.Admission`. The built-in `limiter` (per tenant and per process, from the policy) always applies; an injected admission is asked after it. Admission is per attempt, inside `initialRequest.send` (the initial-request retry wrapper), after the request body is built. A failed attempt returns its permit before backoff. The attempt that got the initial response keeps its permit until the adapter returns, which has closed the stream (`initial.done`). A refused attempt is not sent, not recorded in `Attempts`, and ends the call with `admission_denied` in the new `PhaseAdmission`. A host's error is kept only via `errors.Is/As`. An ended context during the wait is `canceled`/`deadline_exceeded` in `PhaseAdmission` with StopReason aborted and pi's request-phase texts, because the wait happens inside pi's retry wrapper.
- **Waiting.** `AdmissionWait` is 0 for immediate refusal, otherwise a bounded wait. A freed permit goes to the earliest waiter it fits, so a tenant at its own limit never blocks another tenant's waiters. The number of waiters is capped by the new required policy field `MaxAdmissionWaiters` (see the maintainer decisions below). `AdmissionWait` bounds only the built-in wait; an injected admission must honor ctx itself.
- **Time limits** (`ai/timeouts.go`). `CallTimeout` is the call context's deadline (setup included), so it ends exactly like a host deadline, whichever is earlier. Per-attempt limits are enforced by `watchedTransport`, which wraps the Client's transport and cancels only that attempt's context:
  - connect: from the start of the round trip until the transport reports a connection (httptrace `GotConn`);
  - response headers: from the start of the round trip until the headers arrive, using the earlier of the policy value and the protocol's `timeoutMs` (default 10 minutes, as openai-node);
  - read idle: only the time a body Read waits for upstream data.

  Per-attempt expiries are `deadline_exceeded` with StopReason **error** (pi's SDK timeout ends as error, not aborted). Connect and header expiries are connection failures and are retried under `Binding.Retry`, as pi retries a timed-out request. Read idle is in PhaseStream and is not retried.
- **`timeoutMs`** is new on `ResponsesOptions` and `SimpleOptions`. It is passed through by the simple mapping; a negative value is `invalid_request`, and 0 means unset (in pi, 0 would time out at once). When set, `X-Stainless-Timeout` is sent as openai-node does (whole seconds, truncated). The differential shows no difference apart from issue 15's `usage.cost`/`usage.reasoning`, the same as every other case.
- **Release**:
  - The probe has new `Permits` and `AdmissionWaiters` gauges.
  - Every scenario checks that no permit is held (built-in or host), no attempt is waiting, no body is open and no call is active.
  - `release-under-close-and-cancel` fills the process across both tenants with 4 waiters, then ends the holders by concurrent and repeated Close and by host context cancel. Every waiter is then admitted and completes, and goroutines settle within baseline+10 after idle pooled connections are closed.
- **Mutation checks.** Each of these breaks the suite:
  - a failed attempt keeping its permit;
  - `done` not releasing;
  - no tenant limit, or no process limit;
  - a host denial leaking the built-in permit;
  - an uncancelable wait;
  - attempt timeouts ending as aborted;
  - the connect timer outliving the connection;
  - the later of timeoutMs and the policy winning;
  - no CallTimeout;
  - timeouts not retried;
  - `timeoutMs` ignored;
  - a negative `timeoutMs` accepted;
  - the account not passed to admission.

  One mutation survives: the idle timer not being stopped after a Read. It cannot be observed today, because the producer never pauses between reads and the response callback runs before the first Read.
- **Review follow-ups (code-review 2026-10-02).** The connect/header timers and the read-idle timer now settle under a mutex: a timer firing as a round trip or a Read returns does nothing, so a healthy body is never cancelled and an expiry is never misattributed to PhaseStream. An already-ended context is refused before taking a permit. Questions for the maintainer, all in ADR-0008 (decided below):
  - the uncapped waiter count (decision 2), which this ticket's "不建立无限准入队列" box depends on;
  - `timeoutMs` 0 treated as unset;
  - an admission denial on a retry replacing the earlier attempt's classification as the call's error.

  Behaviour of an injected admission while it waits is the host's: the interface requires it to honor ctx, and the test double never blocks.
- **Not done here:** a host example with stress-tested numbers belongs to issue 18; multi-tenant interleaving under concurrency belongs to issue 17.

**2026-10-02 — maintainer decisions; ADR-0008 accepted.**

- **Waiter cap:** accepted the new required policy field `MaxAdmissionWaiters` (process-wide). An attempt that would wait beyond it is refused at once with `admission_denied`, naming the field. It is covered by E2E `TestAdmission/waiters-capped` (the queued waiter is still served) and by construction rejection of 0 and -1. A mutation that removes the cap is caught. A per-tenant waiter cap is deferred until there is evidence it is needed.
- **`timeoutMs: 0`:** keeps meaning unset; this is a recorded intentional difference from pi.
- **Admission denial on a retry:** the call's error stays `admission_denied`; earlier attempts keep their classification in `Attempts`.
