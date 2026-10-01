# 12: 字节与事件队列限额（E08 资源部分 / D1 执行）

**What to build:** 宿主通过资源策略约束请求与图片字节、单 SSE frame、工具 JSON、错误体、总输出、事件队列条数及字节；限额在读取/解码/分配阶段生效，超限以 resource_limit 终态结束，慢消费者或异常服务不能让进程资源无界增长（spec I9 前三段、ADR-0002、User Stories 33–34）。

**Blocked by:** 05

**Status:** ready-for-agent

- [ ] 每项限额覆盖边界值与 limit+1：请求体与图片字节在发送前拒绝；单 frame、工具 JSON、错误体、总输出在读取过程中检测，不先无界读入再检查
- [ ] 事件队列超限：结束本次请求并发布 resource_limit 错误，Phase 指明事件队列；已排队事件保留、不静默丢 delta；为终态预留位置；不阻塞生产者、不阻塞 Result
- [ ] 只取 Result 而不读事件的 Stream 在长输出时以上述 resource_limit 结束；相同输出用 Complete 正常完成
- [ ] 公开文档写明“仅需最终消息应使用 Complete”
- [ ] 协议 adapter 内无散落的隐式默认数值；所有容量来自构造时策略
- [ ] 超限后响应体关闭、活动请求归零、goroutine 有界收敛（资源探针证明，不断言精确 goroutine 数）
- [ ] 限额内的协议行为与 pi 差分一致；超限行为作为扩展单独断言并登记
