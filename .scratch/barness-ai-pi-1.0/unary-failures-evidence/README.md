# 工单 07：unary 故障与资源释放

2026-10-08。公共验收边界由工单指定，失败方式在实现前记录于 [FAILURES.md](FAILURES.md)。
新增 TestUnary E2E 与已有 TestClassifier 场景共同验证 TypeSafe 首条 unary 生命周期；
所有 Provider、型号、价格及宿主故障均为合成数据。

本次修复：开始前已结束的 context 在复制/解析前拒绝；解析器成功返回也检查 context；
绑定的可变重试配置在解析边界复制，凭据后端接收独立副本；直连 HTTP 的连接失败正确进入
显式重试；AttemptStarted 用量完整性明确为 unreported。两轴审查另发现回调容量检查晚于复制；现已在序列化前检查普通 JSON
的可取得尺寸，最终请求字节预算在独立问题解码前检查。

验证覆盖初始 HTTP 分类和耗尽、固定快照与冻结回调、准入拒绝/等待/迟到许可、退避中断、
连接/响应头/读空闲时限、读取期间取消和截止时间、精确输出/错误体上限、无效答案、
回调错误链、无效回调、子策略及 Observer 故障。成功体之后的失败不重放；body 关闭时仍
持有许可，各失败路径资源探针归零，并验证其他租户与后续调用继续获得许可。

[context-red.log](context-red.log)、[resolution-red.log](resolution-red.log)、
[retry-red.log](retry-red.log)、[binding-red.log](binding-red.log)、
[observation-red.log](observation-red.log) 保留修复前的公共失败；对应 green 日志保存修复后通过。
retry-red 还保留一处测试预期路径写错（实际沿用绑定 /v1），该预期已改正；生产重试修复
仅针对实际连接失败。[allocation-red.log](allocation-red.log) 复现一个 8 MiB 宿主值造成 75–294 MiB 库内分配；
[allocation-green.log](allocation-green.log) 验证提前拒绝、公共调用分配小于宽松的 4 MiB 上界。
复审的 [encoding-red.log](encoding-red.log) 另复现数字/空字符串数组造成 54–73 MiB 分配；
[encoding-semantics-red.log](encoding-semantics-red.log) 复现匿名字段冲突、TextMarshaler 和自定义问题编码被误拒。
[string-key-red.log](string-key-red.log) 验证字符串 map key 不采用 TextMarshaler，超大原始键须提前拒绝。
[encoding-green.log](encoding-green.log) 中全部 11 项通过。encoding-red 的早期 text-key 预期曾误把
字符串 key 视作 TextMarshaler 输出，最终改用结构体 key，按标准库实际语义验收。
自定义 marshaler 执行属于宿主，不能预知其输出，其结果仍先通过字节预算。
没有新增单元测试或生产测试接口。

专项 race 和冻结 pi 差分已通过，命令如下。全量 race、门禁数据核验、审查报告和哈希索引
在最终证据中记录。此工单仅交付离线生命周期，不提供 TypeSafe live 或新增内建型号支持声明。

    go test ./ai/e2e -run '^TestUnary' -count=1
    BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(Unary|Classifier)' -count=1
    go vet ./...
    go vet -tags live ./ai/...
    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1

实际 bundle 位于仓库忽略的 .evidence/barness-ai/；最终 report、manifest 和场景索引提供
可核验的 SHA-256、脱敏审计与每个场景的独立回放命令。两轴 code-review 基点为
4f90386fa4c879dddfa3013cce778c44de873acc。
