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
[encoding-green.log](encoding-green.log) 中全部 18 项通过。encoding-red 的早期 text-key 预期曾误把
字符串 key 视作 TextMarshaler 输出，最终改用结构体 key，按标准库实际语义验收。
[pointer-encoding-red.log](pointer-encoding-red.log) 先复现切片/数组元素的指针接收者编码被误拒，
对应 green 验证保持标准库语义。这两轮复审发现出现于全量运行期间，已主动终止旧代码的运行
（interrupted-full-race*.log 的 terminated）。[raw-element-red.log](raw-element-red.log) 与
[known-pointer-red.log](known-pointer-red.log) 先复现已知 RawMessage/封闭问题被编码回退跳过，
最终实现先计已知长度，再处理未知编码；修复并经两轴复核后运行最终全量。
自定义 marshaler 执行属于宿主，不能预知其输出，其结果仍先通过字节预算。
没有新增单元测试或生产测试接口。

最终专项及全量 race 均已通过。3,730 个证据场景全部 PASS：P07（含差分）294 个，
其中新增 unary 89 个；冻结 pi 差分 623 个，pending 为 0；压力场景 11 个。
86 个资源记录的 body/call/permit/waiter/event/host permit 读数全部为零，
11 个超限回调分配场景最高 24,096 字节（验收上界为宽松的 4 MiB）。
run 与最终交付目录的脱敏审计均为 0 发现；Standards 和 Spec 两轴各三项 P2 已修复，
最终复审无遗留发现。此工单交付离线生命周期，TypeSafe live 与内建型号支持仍由工单 08 验收。

    go test ./ai/e2e -run '^TestUnary' -count=1
    BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(Unary|Classifier)' -count=1
    go vet ./...
    go vet -tags live ./ai/...
    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai/issue07-final" go test -race ./... -count=1

实际 bundle 位于仓库忽略的 .evidence/barness-ai/；最终 report、manifest 和场景索引提供
可核验的 SHA-256、脱敏审计与每个场景的独立回放命令。两轴 code-review 基点为
4f90386fa4c879dddfa3013cce778c44de873acc。


最终测试代码为 0f7b005ce0bb68f31e316d4cec1f55396a00e5e8。完整运行耗时约 496 秒，
实际 bundle 为 .evidence/barness-ai/issue07-final/20261008T024630.958851000Z；
源码、目录快照和输出证据的 SHA-256 见 [report.json](report.json) 与 [manifest.json](manifest.json)。
[unary-cases.json](unary-cases.json) 保存每场景回放命令、fixture/输出哈希、请求数量与断言数；
[unary-resources.json](unary-resources.json) 保存资源读数及分配实测。

在仓库根目录核验提交的证据与当前代码（需要停留在本次代码或其仅修改文档的后继版本）：

    python3 .scratch/barness-ai-pi-1.0/unary-failures-evidence/verify.py
    go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/unary-failures-evidence

发布数据评估见 [release-evaluation.json](release-evaluation.json)：P0 离线、差分、审计、
追溯和快照均 PASS，工单 07 四项新增追溯全部 PASS。该 -bundle 调用只评估已有证据，
不导入前述命令记录，也没有传入真实 live bundle，因此整体发布结论为 false、退出码为 1；
不是本工单离线测试失败。release-before-audit.log 保留一次过早评估：race 进程尚未写出
审计记录，门禁正确拒绝；最终评估已等待 run 审计完成。构建与 race 的实际成功退出码
分别记录于 [commands.json](commands.json)，两轴复核见 [review.md](review.md)。

最终提交仅更新工单状态与证据；生产代码及测试自最终完整运行后没有改变。
