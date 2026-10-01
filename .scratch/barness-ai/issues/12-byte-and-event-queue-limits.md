# 12: 字节与事件队列限额（E08 资源部分 / D1 执行）

**What to build:** 宿主通过资源策略约束请求与图片字节、单 SSE frame、工具 JSON、错误体、总输出、事件队列条数及字节；限额在读取/解码/分配阶段生效，超限以 resource_limit 终态结束，慢消费者或异常服务不能让进程资源无界增长（spec I9 前三段、ADR-0002、User Stories 33–34）。

**Blocked by:** 05

**Status:** ready-for-agent

- [x] 每项限额覆盖边界值与 limit+1：请求体与图片字节在发送前拒绝；单 frame、工具 JSON、错误体、总输出在读取过程中检测，不先无界读入再检查
- [x] 事件队列超限：结束本次请求并发布 resource_limit 错误，Phase 指明事件队列；已排队事件保留、不静默丢 delta；为终态预留位置；不阻塞生产者、不阻塞 Result
- [x] 只取 Result 而不读事件的 Stream 在长输出时以上述 resource_limit 结束；相同输出用 Complete 正常完成
- [x] 公开文档写明“仅需最终消息应使用 Complete”
- [x] 协议 adapter 内无散落的隐式默认数值；所有容量来自构造时策略
- [x] 超限后响应体关闭、活动请求归零、goroutine 有界收敛（资源探针证明，不断言精确 goroutine 数）
- [x] 限额内的协议行为与 pi 差分一致；超限行为作为扩展单独断言并登记

## Comments

**2026-10-02 — implemented** (E2E `ai/e2e/byte_limits_test.go`, `event_queue_test.go`, `resource_release_test.go`; differential `PIDIFF-P01-E08-*`, 10 cases).

- Where each limit acts (`ai/limits.go`): images in PhaseScope before anything resolves, sized as the image's own bytes computed from the base64 text (not decoded; text that is not base64 is passed on as pi does). The request body is checked after the payload callback, in PhaseRequest, before sending. Every response body is wrapped by SDK middleware before the SDK or the adapter reads it. A 2xx body is bounded per SSE frame (lines split exactly as the SDK's `bufio.ScanLines`) and in total. Any other body, including the SDK's own `io.ReadAll` of an error body and barness's read of a redirect, is bounded by MaxErrorBodyBytes. Tool JSON is checked by the assembler before the adapter grows its copy.
- **Decision for the maintainer: MaxOutputBytes** is the streamed response body of one call as read (after transport decoding), not the message's text. This bounds everything the call reads and may keep (extra blocks, unknown events, repeated content) and is checked while reading. The cost is that for Responses, whose terminal frame repeats the whole response, the text a policy admits is roughly half the configured value, and MaxFrameBytes must hold the largest accepted response. Both are stated on the policy fields.
- **Decision for the maintainer: no retry after a limit.** A 429/5xx whose error body exceeds MaxErrorBodyBytes ends as resource_limit (HTTP status, request id and retry-after are kept) and is not retried, whereas pi would retry it. This follows "超限以 resource_limit 终态结束". It is registered in the ledger for `PIDIFF-P01-E08-error-body-over-limit-stream`.
- Event queue (`ai/stream.go`): the queue bounds count and bytes of non-terminal events. Bytes are the text an event carries: deltas, end contents and tool call fields. Start events count 0 bytes and are bounded by count. The terminal's place is reserved. `push` never blocks. An overflowing event is not queued. The assembler latches the failure (PhaseEventQueue, new `PhaseEventQueue = "event_queue"`), publishing stops, the adapter ends at its next check, and `finish` lets the latched failure win as a backstop. The terminal message keeps everything received, including the content of the event that did not fit, so no delta is dropped silently. Queued events stay readable in order.
- Docs: `ai/doc.go` (new package doc), `Stream`, `Complete` and the `ResourcePolicy` fields say that callers needing only the final message must use Complete.
- Release: `probe.ActiveCalls` is new. Every scenario checks that bodies are closed and no call is active. `TestResourceRelease` runs 24 concurrent unread, unclosed Streams ending at queue, frame and error-body limits (the provider holds the connection) and checks that goroutines settle within baseline+10, not an exact count. The `held` scenarios prove reading-time detection: a frame or body that never ends still fails in time.
- Differential: the at-limit cases (request, image, frame, output, tool JSON, error body) show only ticket 15's `usage.cost`/`usage.reasoning`, the same as every other case. The over-limit cases (frame, output, tool JSON, error body) are registered as case-scoped extensions placed first in the ledger. Request/image over-limit sends nothing, so there is nothing to compare, and the event queue is not observable while the differential drains concurrently. Both are asserted offline only.
- Not here: a host example policy with stress-tested numbers (spec §9 "正式示例") belongs to 18. Admission and timeouts belong to 13.
- Mutation-checked: frame, CRLF frame end, output, error body, request, image off-by-one, queue count off-by-one, queue bytes, queue bytes released on read (the provider pauses via the new `Reply.AfterChunk`), tool JSON, the adapter's failure check, no-retry, and the active-call probe each fail the suite. The `finish` latch survives because it is a backstop the adapter check already covers.
