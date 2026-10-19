# 验证策略

决定测试范围、准备付费验证或执行发布门禁时读取本文件。验证结果只证明实际运行的范围；完整发布仍要求现有十项门禁全部通过。

## 开发阶段

行为变更在已有公共 Client 或 CLI 边界先建立失败场景，再实现与运行定向普通 E2E。必须隔离测试时，先列出失败方式再写实现。纯文档、展示清单和审查记录使用引用、格式、状态与来源相等性检查。参与运行、价格、追溯或证据消费的 JSON/目录数据按实际行为选择验证方式。

| 变更 | 普通验证范围 | 记录的建议 race 范围 |
| --- | --- | --- |
| 生命周期、队列、取消、共享状态与资源策略 | 受影响的公共 E2E，含隔离与释放 | 对应并发路径与旧 Provider 回归 |
| 协议编码、解析、历史与用量 | 对应协议 E2E | 涉及的异步、续接与资源路径 |
| 发布调度、证据消费者与测试工具 | 公共 CLI/进程及受影响消费者 | 执行、取消与并发读取路径 |
| 纯文档及展示数据 | 引用、状态、完整来源集合相等性 | 执行来源及选定输入未变时沿用已有事实 |
| 模块、SDK、构建标签或无法限定的共享变更 | 扩大普通验证范围 | 完整集合 |

测试范围由影响决定。合并本身不触发全量重跑；相关行为、测试、依赖或选定输入变化后重新验证受影响范围。并发场景用受控服务、hook 和 channel 确定阶段及取消时点；普通测试应能确定性重现已知时序问题。

## 审查与冻结

付费或长时验收前，先完成关键边界整合审查：最终 payload、原生历史、取消、超时、用量、资源释放及租户隔离。修复后复审，再冻结执行来源、构建工具与选定输入。保留最终审查；已审且字节未变的部分引用原结论。

付费轮次另需指定范围、资格和预算的明确授权。验证计划获批不授予新厂商调用或付费预算。保留实际失败与原账本，按原生来源规则判断重验范围。

## 普通验证与人工 race

默认流程是前置检查 → 双 vet → 完整普通测试（差分和压力开启）→ 原生证据消费及审计 → `READY_FOR_RACE`。前项失败停止后续昂贵执行；诊断模式只能继续普通阶段。

开发、合并、修复、普通通过和收尾均不自动启动 race，包括定向 race。只有人类明确选择实际 race 命令或在聊天中明确要求执行该范围，才运行相应 race；方案实施批准不包含这项选择。失败重跑与扩大范围仍按明确授权处理。

等待人工时交付命令、范围、来源、输入、已有证据和预计耗时，标记 `race: NOT_RUN`、`release.pass: false`。可以交付实现与普通结果，要求完整发布证明的验收项继续开放。人工选择后重新验证普通阶段回执、完整来源集合及原生输入；相关来源或输入变化时拒绝旧结果。完整 race 生成独立 bundle，再由原发布判定验收。

CLI 用法见 [发布指南](../barness-ai/README.md)。正式覆盖保持完整 `go test ./...`、`go test -race ./...`、双 vet、差分、压力、当前真实来源资格及逐包审计。

## 并发覆盖入口

这些既有入口供普通验证及人工选择定向 race 时参考，不自动启动 race：

- MiniMax：`TestMiniMaxSharedTenantRoundTrips`、`TestMiniMaxSharedConcurrentRelease`、`TestMiniMaxSharedAdmissionIsolation`、`TestMiniMaxSharedCancellation`、`TestMiniMaxSharedObserverFailures`、`TestMiniMaxSharedFrozenRetries`。
- 共享逻辑：`TestEventQueueLimits`、`TestEventQueueReleasesAsRead`、`TestProtocolTimeouts`，追加受影响旧 Provider 的取消、Settlement 和隔离场景。
- 视觉取消：`TestVisualUnsentCancellation`、`TestVisualUnsentAfterPin`、`TestVisualSentCancellationBeforeDecode`，以及公共 verify/计量场景。
- 容量策略：原压力集合与 `TestMiniMaxSharedDesignLoad`，明确记录启用开关及未运行项。

## 证据与收尾

每阶段记录实际 argv、数字退出码、起止时间、范围、来源和原件指针。未执行、失败、输入过期分别记录，不互相替代。父子测试和并行包的时间不能相加当作墙钟；包尾部差额也不能直接当作审计时间。

每个功能只维护一份当前状态；其他文档链接该状态。沿用现有原生 bundle、审计和留存工具，以一个最终索引关联事实与审查。文档收尾不新增逐轮复制器；历史失败与旧证明保持原样。
