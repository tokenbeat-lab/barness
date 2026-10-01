# 04: 流生命周期（E01）

**What to build:** 应用开发者可以按任意合法方式消费一次生成：逐事件读取、完全不读事件只等 Result、两者并行，或随时 Close；获得顺序稳定、块索引稳定的事件和可安全并发读取的累计视图（spec I8 前三条、E01）。

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] 解析器被屏障阻塞时，调用者已拿到 Stream 且可 Close，Close 使后台过程结束并释放资源
- [ ] text/thinking 块交错时块索引稳定；成功增量不先于 start；完整消费的流恰有一个终结事件
- [ ] PartialView 以同步保护的只读 Snapshot 提供 live partial 语义；多个事件可引用同一持续更新视图；delta、块索引与 end 数据保持各事件自身值；终态后最终消息稳定
- [ ] Result 不消费事件、不要求先 Next、可与事件消费并行；不返回“流仍活跃”类错误
- [ ] Next/Event 单消费者；Close 可并发、幂等；Err 为 Scanner 风格（终态前为 nil，Next 返回 false 后等于 Result 的 error）
- [ ] Complete 只等待最终结果，不为无人读取的事件建立积压队列（以可观测的方式证明，例如队列计数为零）
- [ ] 调用输入在接收时取得副本：交接后修改源数据不影响在途调用，返回集合与快照不泄漏内部可变存储
- [ ] full/simple × Stream/Complete 四入口各至少一条用例；一条完全不调用 Next 的 Result 用例（输出在队列限额内）
- [ ] 全部用例在 `-race` 下通过；每个等待有截止时间
