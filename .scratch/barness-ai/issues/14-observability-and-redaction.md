# 14: 观测与秘密脱敏（E09）

**What to build:** 运维人员通过 Observer 按租户、逻辑调用和请求尝试追踪状态、重试、取消与用量归属；观测异步有界，拥塞或失败不改变生成结果；key、正文、工具内容与原生密文不进入默认日志、观测和错误（spec I9 后半、User Stories 37–39）。

**Blocked by:** 11

**Status:** resolved

- [x] 事件 CallStarted / AttemptStarted / AttemptFinished / CallFinished；调用级关联 TenantID/RequestID，尝试级另含 AttemptID 与厂商 request ID
- [x] 记录已解析的 BindingID、AccountScopeID、配置/凭据版本、真实 Provider/API/model、时间、错误类别、已知 usage；ActorID/JobID 提供时保留
- [x] 预检失败不伪造 AttemptStarted；无效身份的拒绝事件不归属到已授权租户
- [x] 07 的原生状态降级计数与原因同时出现在 Observer
- [x] 慢或失败的 Observer 不阻塞、不改变结果；丢失计数可查询
- [x] 成功、预检拒绝、失败、重试、中断、Close 各路径的 Call/Attempt 关系准确
- [x] 以含 key、Authorization、正文、工具参数、原生密文的场景（含敏感错误体）扫描日志、观测、错误文本：无泄漏；授权 Result 中原生状态仍完整
- [x] 原始 TenantID 不进入厂商 payload，也不作为默认无界指标标签

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/observability_test.go`; decisions in ADR-0009, status proposed).

- **API** (`ai/observer.go`): `Config.Observer` (`Observe(Observation) error`), `Observation{Kind, Time, Call CallMetadata, Attempt, Duration, StopReason, Error *ObservedError, Usage}`, `Client.ObserverStats()` (Delivered/Failed/Dropped). New required-when-observing policy field `MaxQueuedObservations`; `CallMetadata` gains `BindingVersion`/`CredentialVersion`.
- **Delivery:** bounded per-Client queue drained by one goroutine started on demand, which exits when the queue is empty. When the queue is full, records are dropped and counted. An Observer error or panic is counted as failed and never reaches the call.
- **Relations:**
  - CallStarted comes before any resolution.
  - AttemptStarted comes after the attempt's admission permit and before it is sent. A call refused at preflight or admission has no attempt records.
  - AttemptFinished comes when an attempt fails or when its stream closes, before its permit is returned.
  - CallFinished mirrors the Result: the message's StopReason, Usage and classification, but no error text.
  - Every path is covered by E2E: success (4 entry points), three preflight refusals, admission refusal, an HTTP failure, a retry, cancel, Close, a native-state downgrade (counts on the attempt records and on CallFinished), a slow observer (drops counted, every record accounted for), and an error or panicking observer.
- **Redaction:** scenario `P01-E09-redaction` puts a key, an Authorization header, the prompt, tool arguments and reasoning ciphertext into the request and the response, plus a 401 body that echoes the key, the header and the prompt. Observations (as JSON and as fmt output) and captured `log`/`slog` output contain none of them. The error text contains no key and no content the library added. The round-1 Result keeps its encrypted reasoning and trusted envelope. The raw TenantID appears in no provider request body or header. The library writes no logs and exposes no metrics.
- **Mutation checks:** each of these breaks the suite:
  - AttemptStarted emitted before admission;
  - the streamed attempt's AttemptFinished missing;
  - a blocking queue instead of drop;
  - no panic recovery;
  - snapshot versions not set.
- **Open for the maintainer** (ADR-0009 "待维护者确认"):
  - refused calls still carry the scope's TenantID (with `Resolved=false`);
  - a provider-echoed prompt stays in ErrorMessage per pi (only key-shaped text is redacted), which needs approval as a security difference or a change;
  - record times use the system clock.
- **Deferred:** telling unknown usage from zero is issue 15's (records carry the message's Usage as is).

**2026-10-02 — maintainer decisions; ADR-0009 accepted.**

- **Refused calls:** keep the scope's TenantID with `Resolved=false`. Usage, success rates and attribution count resolved records only; the `Observer` doc says so.
- **Provider-echoed content in ErrorMessage:** kept as pi does (only key-shaped text redacted). This is an approved security difference, recorded in spec I9 and ADR-0009 decision 5. The `Error` doc tells hosts that error text is tenant content, not for shared logs, and to log Observer records instead.
- **Record times:** stay on the system clock, as message timestamps do (ADR-0009 decision 6).
