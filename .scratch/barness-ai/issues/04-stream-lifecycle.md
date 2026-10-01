# 04: 流生命周期（E01）

**What to build:** 应用开发者可以按任意合法方式消费一次生成：逐事件读取、完全不读事件只等 Result、两者并行，或随时 Close；获得顺序稳定、块索引稳定的事件和可安全并发读取的累计视图（spec I8 前三条、E01）。

**Blocked by:** 01

**Status:** ready-for-agent

- [x] 解析器被屏障阻塞时，调用者已拿到 Stream 且可 Close，Close 使后台过程结束并释放资源
- [x] text/thinking 块交错时块索引稳定；成功增量不先于 start；完整消费的流恰有一个终结事件
- [x] PartialView 以同步保护的只读 Snapshot 提供 live partial 语义；多个事件可引用同一持续更新视图；delta、块索引与 end 数据保持各事件自身值；终态后最终消息稳定
- [x] Result 不消费事件、不要求先 Next、可与事件消费并行；不返回“流仍活跃”类错误
- [x] Next/Event 单消费者；Close 可并发、幂等；Err 为 Scanner 风格（终态前为 nil，Next 返回 false 后等于 Result 的 error）
- [x] Complete 只等待最终结果，不为无人读取的事件建立积压队列（以可观测的方式证明，例如队列计数为零）
- [x] 调用输入在接收时取得副本：交接后修改源数据不影响在途调用，返回集合与快照不泄漏内部可变存储
- [x] full/simple × Stream/Complete 四入口各至少一条用例；一条完全不调用 Next 的 Result 用例（输出在队列限额内）
- [x] 全部用例在 `-race` 下通过；每个等待有截止时间

## Comments

**2026-10-01 — implemented** (lifecycle E2E in `ai/e2e/stream_lifecycle_test.go` with fixture `testdata/responses/interleaved-blocks.json`; differential scenario `PIDIFF-P01-E01-interleaved-*`).

- `PartialView` (`ai/partial.go`) holds the call's message under a RWMutex; the assembler mutates only through it. `StartEvent` and every text/thinking block event carry the same `*PartialView`; `Snapshot()` returns a copy sharing no storage. While running, the view's `StopReason` is `pending` (pi's initial value). After the terminal it equals the final message and never changes. `DoneEvent`/`ErrorEvent` carry their own message copy, as in pi.
- Thinking blocks: `ai.Thinking` plus `thinking_start/delta/end` events, ported from pi's Responses reasoning path (summary and reasoning_text deltas, `\n\n` on summary part done, end replaces the text with the joined summary (else content) text). The parser now keeps one typed slot per output index, so interleaved blocks keep their content index. `thinkingSignature` is the compacted raw item JSON. It matches pi for the fixture; non-ASCII escapes and number spellings are kept as sent where `JSON.stringify` would normalize them. Exact replay encoding and trust belong to 07. So does pi's `backfillReasoningSignatures`, which copies `encrypted_content` from `response.completed` for Azure-style streams that omit it from `output_item.done`. It is not ported, so on such streams the signature lacks it. A done item whose output index holds a block of another kind is ignored, as with pi's `getOrCreateSlot`.
- Complete's no-backlog proof: `ai/internal/probe` counts queued stream events per Client, injected through `Config.Probe` (internal type, so hosts cannot set it; same test-assembly precedent as `AllowLoopbackHTTP`). Complete peaks at 0. The same output on an unread Stream peaks at the full event count, which shows the probe really observes queueing.
- Host double: `HoldBindings` now returns a `Hold` with an `Entered()` barrier, so "blocked in the resolver" is established without sleeping. Every wait in the lifecycle test has a 5s deadline.
- The earlier Stream-only request-copy case in `client_routing_test.go` moved into the lifecycle test and now covers all four entries; outputs (Result, done message, snapshots) are checked for shared storage.
- Mutation-checked: removing the Snapshot lock trips `-race`; dropping `Request.clone` fails the input-copy case.
- Differential: barness events are projected with `partial` = snapshot at receipt, as the pi runner serializes on receipt. The old pending `events[*].partial` entry is replaced by extension decisions limited to the timing-dependent fields (`partial.content` changed/only_barness, `partial.usage`, `partial.stopReason`, `partial.responseId`). These are needed because barness's producer runs ahead of the consumer, while pi's single-threaded runner happens to serialize at event time. A missing `partial` or differing identity fields stay pending. Event types, indexes, deltas, end contents, signatures and the final message match pi exactly. Remaining pending items belong to 05/09/15 only (plus `events[*].partial.usage.cost` → 15).
- Known gap, owned by 12's request-structure checks: pointer variants (`*ai.Text`, `*ai.UserMessage`) satisfy the content/message interfaces through value-receiver markers. `Request.clone` does not deep-copy them, and the Responses adapter silently skips `*Text`. Validation should reject them.

**2026-10-01 — review follow-up** (two-axis code review): every wait in the lifecycle scenarios now has a 5s deadline (`within`/`resultWithin`). The done-item kind mismatch now follows pi. Reading the streamed thinking takes the view's read lock. Test helpers are deduplicated (`sseEvents`, `callResult`). The snapshot check derives the streamed text from the fixture's expected events. Accepted as is: the `Config.Probe` test hook (see above), and the protocol-wide scope of the live-partial ledger waiver. The partial's timing-dependent content cannot be compared across runtimes, so the live-view contract is gated by the offline E01 E2E instead.
