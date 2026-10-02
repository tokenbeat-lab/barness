# 18: 宿主契约、路由与最小示例（E10，批次 B 收口）

**What to build:** 上层模块开发者通过三个最小示例理解职责交接：本地显式装配租户与 key、宿主由认证结果建立 scope 并按租户读取历史与验证原生封套、工具往返；宿主汇合多个流时每个事件携带不可变调用归属；下游断开只取消本次生成（spec I2 末段、Testing Decisions §1、User Story 49）。批次 B（Responses + Anthropic）在本票通过后视为验收完成。

**Blocked by:** 07, 12, 13, 14, 17

**Status:** resolved

- [x] 事件经信封或等价不可变引用携带 TenantID、RequestID、BindingID、实际 Provider/API/ModelID；多流汇合用例不依赖到达顺序推断归属
- [x] 本地装配示例：宿主显式读取指定环境变量或秘密文件后注入，核心不自行读取
- [x] 宿主接入示例：认证结果 → CallScope、按 `(tenant_id, session_id)` 读取历史、TrustNativeState、context 取消传播；恶意自报 TenantID / 历史引用由宿主边界拒绝，直接向库提交无效 scope/绑定由库拒绝——两类责任分别断言
- [x] 工具往返示例：宿主执行工具并以新 RequestID 发起下一轮
- [x] 下游断开或发送失败只取消本次生成，其他流不受影响
- [x] Responses 与 Anthropic 分别通过 E01–E09、E11 适用场景（含 12/13 的资源与准入场景）
- [x] 示例本身作为离线 E2E 运行并产出证据包
- [x] 示例给出完整的有限资源策略（spec §9“正式示例”、ADR-0002），每个数值附适用负载、取值依据与调整说明；准入相关数值按下文“准入取值建议”推算，并由突发压力场景的数据确认，不写成已验证的通用默认值

## 准入取值建议

记录于 2026-10-02（工单 13 / ADR-0008 收尾时）。这些是推算方法和起点值，没有压测数据支撑。示例中的最终数值以本票压力场景的结果为准。

### 取值思路

1. **只让可能等到许可的尝试排队。** 等待者只有在 `AdmissionWait` 内有许可释放时才有用，否则必然超时，只是多等了一段。所以上限应接近等待期间预计释放的许可数：

   ```
   MaxAdmissionWaiters ≈ MaxConcurrentProcess × AdmissionWait ÷ 单次调用典型时长（p50）
   ```

   再留少量余量，最少取 1。

2. **LLM 调用时间长，所以这个值通常很小。** 流式生成常常要 10 到 30 秒。例如进程并发 32、等待 2 秒、典型时长 20 秒时，32 × 2 ÷ 20 ≈ 3，取 4 左右即可。设得更大，只会让更多请求等满 `AdmissionWait` 后照样被拒。

3. **按内存兜底。** 准入发生在请求体构造之后（ADR-0008 决策一），每个等待者都持有已编码的请求体，最大可达 `MaxRequestBytes`。排队时的最坏内存约为：

   ```
   MaxAdmissionWaiters × MaxRequestBytes
   ```

   这个值须在实例内存预算之内。

### 建议起点

| 场景 | AdmissionWait | MaxAdmissionWaiters | 说明 |
| --- | --- | --- | --- |
| 本地单用户 | 几秒 | 与 `MaxConcurrentProcess` 相同 | 请求少，排一下体验更顺 |
| 云端交互式 | 0 到 2 秒 | 按公式推算，通常是个位数 | 让客户端尽快收到拒绝 |
| 云端批处理/后台任务 | 0 | 不起作用，但仍须填正数 | 由任务队列负责排队与重试 |

云端宿主拿到 `admission_denied` 时，应向自己的客户端返回 429 与 Retry-After，让排队发生在 barness-ai 之外（客户端或任务队列）。内置等待只用来消化短暂的抖动。

### 验证方式

做一个突发场景：同时发起多于“并发上限 + 等待上限”的调用，记录：

- 立即被拒的比例；
- 等待后拿到许可的比例；
- 等待后仍超时的比例；
- 内存峰值。

“等待后仍超时”的比例高，说明 `MaxAdmissionWaiters` 或 `AdmissionWait` 设大了。示例配置的数值与说明以这组数据为依据，并随证据包一起保存。

## Comments

**2026-10-02 — from issue 17 (maintainer decision):** call identifiers are globally unique (ADR-0001): the host example must mint RequestIDs that never repeat across tenants or calls, and an injected admission/Observer may key on RequestID or AttemptID alone. BindingID and CredentialID stay unique per tenant only.

**2026-10-02 — implemented.** The evidence bundle has 1625 cases: 1409 PASS and 216 NOT_RUN. The NOT_RUN cases are the pi differential, which needs the `BARNESS_AI_PIDIFF=1` switch; with it set, everything passes. `go test -race ./ai/...` and `go vet ./...` are clean.

- **Event envelope (library):**
  - The immutable call attribution is now a type of its own, `CallAttribution`: TenantID, RequestID, ActorID, JobID, BindingID, Resolved, and once resolved the actual ProviderID, API, ModelID and AccountScopeID. `CallMetadata` embeds it, so the JSON stays flat.
  - `Stream.Envelope()` returns an `EventEnvelope{Call, Event}`. Each event carries the attribution as it stood when the event was published. A terminal refused before resolution has `Resolved=false` and does not echo the requested model.
  - The terminal's `Call` equals the Result's attribution.
  - E2E `ai/e2e/envelope_test.go`: Responses and Anthropic × stream full/simple × success, 401 after resolution, and refusal before resolution.
- **Examples in `ai/examples/`.** Each runs as an offline E2E and leaves an evidence bundle.
  - **`localassembly`** (`example_local_test.go`):
    - The program reads the key itself from an environment variable or secret file it names, and assembles a single tenant and binding.
    - An environment variable that is unset or empty, a file that is missing or empty, and setting zero or two sources all fail at `Open`. Nothing falls back to another source.
    - Throughout, the environment is polluted with decoy keys, endpoints and proxies, and the decoys receive no requests.
    - On both protocols, the output equals the library's own assembly; only the native-state provenance differs, naming the local tenant and account.
  - **`hostintegration`** (`example_host_test.go`, `example_host_streams_test.go`), covering both responsibilities and asserting them separately:
    - Authentication result → `CallScope` (ActorID, a fresh RequestID). History is read by (tenant, session); `MemoryStore` is keyed by tenant. The stored envelope goes back through `TrustNativeState`.
    - On the second turn, the request read back from storage equals the request sent with the in-process message. Downgrades are 0 on both protocols. Anthropic uses a signed thinking block (`block-start-content`).
    - **Refused at the host boundary:** a claimed tenant, another tenant's session, a fabricated session, an unauthenticated caller, a missing prompt. Each case shows zero requests, zero observations and zero admissions.
    - **Refused by the library:** a scope missing TenantID, a scope missing RequestID, an unknown binding, a disallowed model, and `TrustNativeState` naming another tenant. JSON cannot decode into trusted state.
    - **Forged history** (the envelope names another tenant; or the envelope is removed and trust claimed instead): the turn goes ahead, and native state is downgraded with `NoEnvelope=1`.
    - **Downstream send failure or client disconnect:** only that turn is aborted and nothing is stored for it. The other tenant's turn on the same Client, held mid-stream until then, equals its solo run, and its downstream receives every event. Permits, bodies and active calls all return to 0.
    - **Merge:** tenant A on Responses and tenant B on Anthropic, with the Provider interleaving their frames in lockstep. Routing by envelope alone gives each call exactly its solo event sequence.
  - **`toolloop`** (`example_toolloop_test.go`):
    - The host validates calls with `ValidateToolCall` and runs them itself. Each next turn is a new logical call with a new RequestID.
    - Responses: both rounds' request bodies equal the fixture. Anthropic: the signed turn replays natively.
    - Truncated or broken turns (length ×2, disconnect) are never executed. Arguments that fail validation are answered with pi's message as an error result. The loop stops at MaxTurns.
  - **`requestid`:** 128-bit random RequestIDs, globally unique, per the issue 17 decision.
- **Resource policies:**
  - **Policies provided:** `localassembly.LocalPolicy`, `hostintegration.CloudInteractivePolicy` and `CloudBatchPolicy`.
  - **Per-value notes:** each value states its intended load, the reasoning behind it and how to adjust it.
  - **Burst scenario:** `E08-admission-burst-pressure-*` (`admission_pressure_test.go`) runs at 1/20 of real time. The process is saturated with calls ending evenly over a typical 20 s duration, then hit with a burst of concurrency + 2 × waiter-cap calls. Results are saved in `pressure.json`:

    | Policy | Waiters | Burst | Admitted after waiting | Refused after full wait | Refused at once |
    | --- | --- | --- | --- | --- | --- |
    | Cloud interactive | 4 (rule of thumb 3.2) | 40 | 3 | 1 | 36 |
    | Same, waiters ×4 | 16 | 64 | 3 | 13 | 48 |
    | Cloud batch (wait 0) | 1 | 34 | — | 0 | ≥ 90% |
    | Local | 8 (rule of thumb 2) | 24 | 2 | 6 | 16 |

    - The data confirms the rule of thumb.
    - LocalPolicy deliberately keeps 8 waiters, per the "本地单用户" row; its comment states what that costs.
    - Under `-race` a call occasionally arrives after the first release and joins the queue, giving 4 admitted / 1 timed out. The scenario therefore asserts shares, not exact counts.
    - This scenario intentionally runs on real time, because it measures the limiter's own timer against call durations; nothing is ordered by sleeping.
  - **Found along the way:** admission happens after the request body is built (ADR-0008), so during a burst every arriving call allocates its body for a moment, including calls refused at once. Peak heap is therefore higher than MaxAdmissionWaiters × MaxRequestBytes; the cloud policy's doc says so.
- **Batch B: Anthropic E07/E08/E09.** Until now these ran on Responses only. They are added in `anthropic_limits_test.go`, `anthropic_admission_test.go` and `anthropic_observe_test.go`:
  - P02-E07: 26 cases.
  - P02-E08: 114 cases — byte limits at boundary and limit+1 on all four entries, held oversize, event queue, timeouts, admission and release.
  - P02-E09: 9 cases.
  - `boundary` gained optional `entries` and `casePrefix` fields; existing Responses cases are unchanged.
  - No library defect surfaced.
  - Not repeated on Anthropic, because the behaviour is protocol-independent core logic: connect timeout, waiter cap, interrupted admission wait, host admission refusal, slow or failing Observer, call timeout during setup.
- **Docs:** GLOSSARY gains Call Attribution and Event Envelope. ADR-0001 gains one consequence. `ai/doc.go` gains sections on attribution and host responsibilities.
- **Code review (two axes):**
  - Fixed: `toolloop` now handles any validation error and rejects duplicate or `Run`-less tools; `Merge` documents that the consumer must read to the end or cancel `ctx`; the burst size in the policy comment was wrong; the environment and file key sources now strip line endings consistently; `encodeUser` no longer swallows its error; the `Stream.Envelope` doc.
  - Not adopted: introducing a per-protocol descriptor in the test helpers (a judgement-call smell).
- **Not applicable:** E10's OpenAI/DeepSeek shared Responses and DeepSeek dual-protocol routing belong to issues 21 and 22 (batch C). A per-call transport override entry point is already covered: `callbacks_test.go` asserts there is none.
- **For maintainer confirmation:**
  - The `CallAttribution` / `EventEnvelope` API shape, and whether `AccountScopeID` belongs in the per-event attribution. The spec requires at least Tenant/Request/Binding/Provider/API/Model; the account is included as well.
  - Whether the examples' numbers are acceptable as deployment starting points, LocalPolicy's waiters = 8 in particular.
