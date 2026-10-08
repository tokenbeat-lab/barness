# 工单 05：TypeSafe 单选 unary 调用验证记录

2026-10-08。Client / HookedClient.Classify 已交付字符串说明的单选切片，调用运行时由聊天与分类共享。
实际 bundle、manifest SHA-256、场景状态、SDK/目录快照和审查结果见 [机器报告](report.json)。

公共测试接缝由 issue 05 明确为现有 scenario fixture runner、公共 Client/HookedClient、受控 HTTP Provider、
宿主替身与既有资源探针；没有新增内部测试接口或隔离单元测试。
[choice fixture](../../../ai/e2e/testdata/typesafe/choice.json) 记录官方来源与核对日期，全部请求、型号、费率、
状态和答案为合成数据。本轮不运行真实 TypeSafe API，也不修改内置目录或声明新的 live 支持。

## 可观察行为

- 类型化单选结果、完整问题集的一次 HTTP 提交、Bearer、ResponseModel 与授权型号分离。
- 空/错型请求、选项数、状态/问题/请求字节、未启用策略、错操作/协议、白名单/目录/租户拒绝。
  预检失败无 Provider 请求和凭据读取；失败结果保留入口 Operation。
- 缺答、多答、错型、非法/缺失概率和 confidence、错误键集、分布和、非最高概率选项、重复 JSON 键；
  允许并列最高项及单选项。先登记明确用量后整体校验；失败答案为空，用量和版本元数据保留。
  重复外层字段被拒绝；重复 usage 或 token 字段不采用覆盖值，用量记为 unreported。
- 回调按 headers → payload → response；最终问题重新解码和冻结。回调改变授权被拒，非法/超限内容
  为 callback_failed，错误原因只通过 Unwrap 保留。请求回调在重试之外执行一次，响应回调只读 HTTP 元数据。
- unary JSON 大于 MaxFrameBytes 但小于 MaxOutputBytes 成功；总输出和错误体超限、不完整 body、
  坏 JSON、读空闲和取消均关闭 body、释放许可。成功响应后不重放。
- 子策略及调用输入独立复制；Observer 操作/调用/尝试/用量完整且不含状态、问题、答案或错误正文，
  拥塞不改变结果。可信回调 panic 继续向宿主传播，但异常退出也释放调用计数、context 和许可。
- 六条聊天路线、pi 1.0.0 差分、既有压力场景保持；Gemini 聊天继续不运行响应回调。
  ADR-0021、ADR-0002/0005/0007/0008/0009/0010、公共契约与 P07 发布追溯同步。

## 先红后绿

先扩充共享 fixture runner 的 classify 入口和完整 JSON 回复，并写成功/失败 fixture、回调、
限额、复制与资源释放 E2E；公共类型尚不存在时 [编译失败](compile-red.log)。
随后补齐三种回调的操作信息，先保留 [操作字段失败](callback-operation-red.log)，再实现入口赋值。
[严格键校验失败](strict-red.log) 证明重复答案/概率/字段被默认 JSON decoder 接受；
实现独立解码边界后通过。[已知长度失败](size-red.log) 证明超大问题键曾进入凭据/请求阶段；
加入复制前的总输入检查后通过。

生命周期抽取后，[panic 失败](panic-red.log) 复现宿主恢复 panic 后许可与调用计数未归还；
增加异常退出也执行的幂等清理，并先红后绿验证。该修复后再次运行定向 race 和全量 race。
两轴审查发现外层响应字段及 usage 的重复键仍会被默认解码覆盖。
[外层重复字段失败](envelope-red.log) 先复现四种公开入口错误，再在 map 转换前识别外层重复键，
排除歧义用量并保留独立有效用量。修复后再次执行定向 race 和全量 race。
红阶段 bundle 保留在本地，日志中的仓库绝对路径已替换为相对路径后进入审计产物。

## 审查与复现

两轴 code-review 以本轮开始前的 10abbd7c90733ed19bc2df627353cb00fd98cf3d 为基点。
最终 Standards 与 Spec 均无未解决 finding；[审查报告与处理记录](review.md) 和机器报告保留初次发现与复核结果。
审查只读当前工单变更。

从仓库根目录重放：

    go test ./ai/e2e -run '^TestClassifier' -count=1
    go test -race ./ai/e2e -run '^(TestClassifier.*|TestTrustedCallbacks|TestChatOperationAuthorization)$' -count=1
    go vet ./...
    go vet -tags live ./ai/...
    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
    go test -race ./... -count=1
    go run ./ai/release/cmd/releasegate -write-snapshot -snapshot /tmp/barness05-catalog.json
    cmp ai/release/catalog-snapshot.json /tmp/barness05-catalog.json
    go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/typesafe-choice-evidence

完整 E2E bundle 位于 .evidence/barness-ai/<run-id>/，包含 fixture、实际请求/回复、结果、断言、
资源读数、观测、脱敏审计和每个场景的精确重放命令。本目录 manifest 固定报告、摘要和红日志的 SHA-256，
完整大负载 bundle 留在本地，以报告中的 manifest 哈希识别。NOT_RUN 的 opt-in 场景不会被写成 PASS；
它们的独立实际通过记录来自启用 pi 差分和压力的完整运行。
