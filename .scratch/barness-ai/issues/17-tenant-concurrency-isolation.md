# 17: 租户并发隔离（E06）

**What to build:** 两个租户在同一 Client、同一 transport 上使用同名 binding/model/session 并发调用；一方鉴权失败、取消或超时不影响另一方，实际发送的 key、endpoint、结果和费用归属互不串用（spec T03、T05、T06、T08、User Stories 5、30）。

**Blocked by:** 03, 05, 11, 16

**Status:** resolved

- [x] 服务端屏障强制 A/B 交错（不依赖概率碰撞或真实 sleep），分别在 Responses 与 Anthropic 上运行
- [x] 捕获每个请求实际收到的 key 别名与 endpoint，与各自租户绑定一一对应
- [x] A 发生 401、取消、超时、重试时 B 正常完成；B 的事件、Result、usage 不含 A 的内容
- [x] 不从其他租户或环境变量兜底
- [x] 结果元数据与观测的 TenantID/RequestID/AccountScopeID 归属正确
- [x] 在 `-race` 下重复运行稳定

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/isolation_test.go`, `isolation_fixture_test.go`, fixture `testdata/isolation/interleave.json`; 56 cases `E06-<protocol>-<scenario>-<entry>` plus `E06-no-environment-fallback`).

- **Setup:** one Client and one transport (body tracker, probe, host admission, Observer, fake backoff clock). Tenants A and B use the same binding names (`primary`, `claude`), the same model and, on every entry that takes one, the same session id; Anthropic's full options have no session field. Each tenant's binding points at its own path on the shared Provider, so the captured path shows the endpoint. The Provider gained `EnqueueAt(pathPrefix, …)`, which serves each tenant its own script whatever the arrival order. The environment is polluted for the whole test (`polluteEnvironment`).
- **Barriers, no sleeps or probabilistic collisions:**
  - `both-succeed`: the Provider writes A's and B's frames in lockstep (A:0 B:0 A:1 …).
  - Every other scenario: B is held just after its first text frame until A's call has ended. In that window A is refused with a 401 (`a-unauthorized`), canceled while streaming (`a-canceled`) or as its response comes back from the shared transport (`a-canceled-on-arrival`), timed out by its own `timeoutMs` (`a-timed-out`), retried after a 429 (`a-retried`), or has no credential (`a-credential-missing`).
  - Each case runs on both protocols and all four entry points. Every wait is bounded. The order in which frames were written is recorded in the evidence as `wire_trace`.
- **Assertions:**
  - B's events, message (usage and cost included), metadata, attempts and captured request (headers included) equal B's run alone on a Client of its own. A succeeding equals A's run alone.
  - Neither outcome contains the other tenant's markers (text, ids, provider request ids, tenant and account strings).
  - Each endpoint receives only its tenant's key on its own path, as many times as that tenant's attempts. On Responses the shared session id gives each tenant a different `prompt_cache_key` and `session_id`.
  - Result attribution, Observer records (lifecycle, tenant, request, account) and host admission requests are each the call's own.
  - The environment decoy receives nothing.
  - Permits, waiters, bodies and active calls all return to 0.
- **Library bug found and fixed:** when the caller's context ends just as the response arrives, openai-go v3.66.0 and anthropic-sdk-go v1.75.0 return `ctx.Err()` without handing the response back or closing its body (`internal/requestconfig`), so the body leaked. The new `deliveredBody` SDK middleware (`ai/sdk_middleware.go`) remembers the delivered body behind a close-once wrapper, and both adapters close it on a failed attempt. `a-canceled-on-arrival` reproduces the leak deterministically: without the fix all 40 runs leak, with it none do. Removing the workaround once the SDKs fix this is tracked as issue 29.
- **Removed:** `P01-E06-shared-client-concurrent-tenants` (client_routing_test.go). It relied on probabilistic collision with identical FIFO replies, so it could not detect content crossing tenants; E06 now covers it strictly.
- **Mutation checks**, each breaking the suite:
  - the cache key ignoring tenant and account;
  - a process-wide "last good key" fallback when a credential does not resolve;
  - one call's cancellation canceling the other calls in flight;
  - attempt records keeping process-wide attribution;
  - the deliveredBody close removed.
- **Race:** `go test -race -count=30 -run TestTenantIsolation ./ai/e2e` is clean. The full suite passes with `-race`, and the pi differential (`BARNESS_AI_PIDIFF=1`) passes.

**2026-10-02 — maintainer decisions.**

- **Call identifiers are globally unique.** RequestID, and the AttemptID derived from it, never repeat across tenants or calls. The host guarantees this and the library cannot check it, so an injected admission or Observer may key on either alone. Configuration identifiers (BindingID, CredentialID) stay unique per tenant, as same-named bindings require. This is recorded in ADR-0001, GLOSSARY (Logical Call) and the docs of `CallScope.RequestID`, `Attempt.AttemptID` and `AdmissionRequest.AttemptID`, and noted on issue 18 for the host example. E06 now gives each tenant its own RequestID; previously both shared one.
- **`deliveredBody` needs no ADR.** It works around SDK defects and is not an architecture decision; the code comment plus issue 29 is enough to track it.
- **`a-timed-out` keeps a real ~100 ms timer.** The interleaving is still forced by barriers: B is held until A ends. The wait is the timeout under test, not a sleep for ordering. Protocol timeouts stay off the fake clock (ADR-0008).

