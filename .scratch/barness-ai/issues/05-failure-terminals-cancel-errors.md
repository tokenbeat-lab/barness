# 05: 失败终态、取消与错误分类（E02 / P01 终态）

**What to build:** 设置失败、鉴权失败、限流、服务端错误、流内协议错误、异常断流、截断、取消和超时都返回完整的最终 assistant 消息，StopReason 与带 Code/Phase 的分类错误区分正常停止、截断、错误和取消，已收到的部分内容和 usage 不丢失（spec I8、I9 错误部分、E02、P01 终态）。

**Blocked by:** 04

**Status:** ready-for-agent

- [x] start 前设置失败可直接 error 事件；建立 HTTP 失败产生错误 assistant 消息
- [x] 401/403 → upstream_auth、429 → rate_limited、5xx、transport、protocol 各自分类；保留可安全公开的 HTTP status、厂商 request ID、Retry-After；401/403 不触发换 key 或身份兜底
- [x] Responses 分别处理 response.completed / incomplete / failed；流内 error、半个 JSON、无协议终态的 EOF 均为 StopReason=error，SDK 无错误 EOF 不得视为成功
- [x] 正常 length 为明确截断终态（incomplete-length 场景）
- [x] StopReason 为 error/aborted 时 Complete / Result 返回非 nil error 且 Result 完整有效，Err() 与之相同；成功终态 error 为 nil
- [x] 取消阶段语义保留基线：lazy setup 期间取消形成 error，进入 adapter 后形成 aborted；附加分类区分 canceled 与 deadline_exceeded
- [x] Close 未完成时取消本次设置或 I/O，已完成时不改写结果；所有路径关闭响应体并释放等待者，不依赖再次 Next
- [x] 流中断保留部分内容、usage 与 errorMessage；涉及秘密的 errorMessage 脱敏并登记差异
- [ ] 未完成的工具参数不被标记为可执行：工具调用尚未解析（function_call 被忽略），由 06 的“截断（length / error）中的工具调用不能通过助手得到可执行结论”验收
- [x] 适用用例接入 02 的 pi 差分，无待处理差异

## Comments

**2026-10-01 — implemented** (E2E `ai/e2e/failure_terminals_test.go` with fixture `testdata/responses/failures.json`, 19 scenarios × the four entry points, plus cancel/deadline in setup, request and stream; differential `PIDIFF-P01-E02-*`).

- Classification (`ai/errors.go`): 401/403 → `upstream_auth`, 429 → `rate_limited`, 5xx/408/409 → **new `upstream_error`**, other 4xx → `invalid_request`, a redirect (never followed) → `protocol`. `upstream_error` also covers provider-reported failures inside the stream: an `error` event, a gateway frame with a top-level `error` object, `response.failed`, and an `incomplete` response for a reason other than `max_output_tokens`. The spec's code list is "至少", so this adds one code rather than overloading `protocol` (stream broke the protocol: half JSON, no terminal, unknown status) or `transport` (connection failed or was cut). `Error` gains `ProviderRequestID` (`x-request-id`) and `RetryAfter` (`retry-after-ms`, then `retry-after` seconds). An HTTP-date `retry-after` is left at 0: converting it needs the replaceable clock ticket 11 introduces. The 401 case asserts exactly one request with the tenant's own key, so there is no key switch or identity fallback.
- Messages follow pi exactly where pi's logic decides them. That covers openai-node's HTTP error message plus pi's `formatProviderError`, `Error Code …`, `response.failed`, the incomplete and unknown-status texts, `Connection error.`, `Request was aborted.` / `Request timed out.` in the request phase, and pi's "ended before a terminal response event" for a cancel mid-stream (the SDK ends quietly on abort, which the differential confirmed). Where pi's text comes from its JS runtime (JSON.parse for half JSON, undici's `terminated` for a cut connection), barness uses fixed texts, recorded as case-scoped extensions.
- Redaction: provider error text has the call's key and any `sk-…` key-shaped token replaced with `[REDACTED]` before it reaches a message or error. This is a case-scoped security extension in the ledger (http-401 masked key, http-403 full key echoed by a gateway). Error-body size bounding is still 12's.
- `AssistantMessage.RawStopReason` ports pi's `rawStopReason` (`status`, `status.reason`, the failed status). The ledger's 05 items are now `fixed`. As in pi, a `response.failed` response's usage is not taken over; usage from `completed`/`incomplete` is kept even when the stream later breaks (disconnect-after-terminal).
- Stage semantics: a cancel or deadline during resolution is `error` + `canceled`/`deadline_exceeded` in its preflight phase. Once inside the adapter it is `aborted` in phase request or stream. Close from another goroutine releases a consumer blocked in `Next` (cancel-after-first-frames). Response bodies are proven closed on every path by a counting transport wrapper (`provider.TrackBodies`), so no internal hook is needed.
- Testkit: provider replies can end `abort` (connection cut mid-body) or `hold` (kept open until the client leaves), and a raw tail can be appended. The pi runner takes `abortAfterEvents`. Ledger decisions can be limited to case IDs (`cases`), so an approved difference stays pending everywhere else. The comparator accepts observations where neither side sent a request.
- Truncated tool arguments: function_call items are not parsed yet (ticket 06), so nothing can be marked executable. 06 owns the "截断中的工具调用不可执行" check.
- Mutation-checked: dropping the stream Close fails the body-closed checks, and dropping redaction fails http-401/403. Differential: 23 cases, and the only pending items belong to 09/15.

**2026-10-01 — review follow-up** (two-axis code review):

- Fixed: the redaction pattern lacked a word boundary and mangled words like `task-runner`. It is now `\bsk-…`, guarded by the 429 scenario, whose text matches pi exactly. 408/409 are `upstream_error` (the baseline retries them), with an `http-409` scenario. The canceled/deadline choice is one helper (`interrupted`) shared by setup and adapter. `statusCode` is renamed `codeForStatus`, and the unbounded error-body read now states its risk and removal condition.
- Open, needs maintainer confirmation: the new public code `upstream_error`. The spec list is "至少", but this is a new stable classification.
- Kept for pi parity: a `response.failed` response's usage is not taken over. This matches pi (differential-equal), though spec §8's wording "保留…usage" could be read otherwise. An HTTP-date `retry-after` reports `RetryAfter` 0 until ticket 11's clock lands.
- Accepted as is: the small test-helper duplication (close-after-N in the differential vs. `drainClosingAfter`; the differential needs per-event projection at receipt).
