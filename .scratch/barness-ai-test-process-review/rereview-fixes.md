# 第 1 节三项复审修复

基线：`fce5f387e375fa3f7901700f5f525fdb64c6804f`。意见见 [复审报告](reviews/20261022-section1-rereview.md)。当前状态统一见 [status.md](status.md)。本轮只修复第 1 节及受影响普通验收。

沿用已批准的 plan.md §1.2 接缝：公共 CLI/子进程、临时文件系统和原生证据消费者。所有清理验证使用 TemporaryDirectory 假仓库，所有 live 资格验证使用合成工件，不请求厂商；不运行 race、压力、付费 live、真实证据清理或后续计划。Go 父子进程固定 GOENV=off、空 GOFLAGS，并移除重负载开关与凭据。

先列失败方式，再逐项先红后绿：

1. 失败原件引用 `../unused` 目录、无扩展名文件或非 JSON 工件时，字符串过滤丢失依赖；依赖内部的传递引用同样必须保留。来源相对路径须先解析到候选范围，不能根据扩展名决定是否保护。普通状态/哈希标量不能膨胀为不存在的路径。
2. 普通 chat 只有 PASS 标签、零 calls/attempts、空预算或错误身份也可取得资格；即使补上正数，缺原件或原件与报告不一致也须拒绝。检查覆盖固定路由、官方目标/型号、预算、真实调用与捕获/结果/Observer 的相互一致；保持非 chat 及历史失败语义，合成场景不产生真实支持资格。
3. 消费者在回执、日志或最终元数据写失败以及输出管道失败时，不能留下可冒充成功的数字退出与日志。正常成功、资格拒绝与终结 I/O 失败均需由实际进程输出和退出事实验证，回执失败须可诊断且不能被发布判定消费为 PASS。

每次 RED/GREEN 写入新的独立命令回执，保留实际 argv、数字退出、日志、工具与源码快照。最终提交冻结后再运行受影响普通验收；状态变化另保存不可变快照及只读判定，旧索引和失败原件不覆盖。

## 修复与原件

以下 attempt 均位于 `.evidence/barness-ai-test-process-review/e2e/`，含实际命令/数字退出/时间、源码快照及日志；不覆盖任何失败原件。

| 意见 | RED 原件 | 实现与普通结果 |
| --- | --- | --- |
| 相对目录依赖丢失 | rereview-dependency-red-20261022T034908.075916Z | 先解析来源相对路径，再保护目录、二进制与无扩展名文件；覆盖失败原件→目录→传递目录，完整临时 move/verify/观察期/purge 后均保留。12 个清理场景 GREEN：rereview-dependency-green3-20261022T034956.628331Z。 |
| 普通 chat 零调用取得资格 | rereview-chat-red-20261022T035048.957835Z | 在原生发布资格边界补固定官方身份/型号/SDK、预算与调用约束；按已认证库存核对捕获、结果、终结 Observer、唯一请求、尝试与用量。13 种独立篡改经公共 CLI 拒绝，完整合成输入只读消费通过。GREEN：rereview-chat-green-20261022T035133.079297Z。 |
| 消费者收尾失败日志/退出不一致 | rereview-consumer-finalize-red-20261022T035241.127698Z；rereview-consumer-outer-red-20261022T035300.842340Z | 删除原生进程自采集 stdout/stderr/exit 格式。外层 consume_live.py 实际采集唯一原生子进程，再原子保存 schema 2；发布判定核对外层+原生资格+实际日志。原生最终写入失败记录实际退出1；外层终结写失败保留 RUNNING 非通过记录，即使子进程退出0也不能取得 PASS。11 个消费者场景 GREEN：rereview-consumer-finalization-green-20261022T035555.693835Z。 |

保持支持矩阵 Merge 的历史部分报告/失败报告语义；完整发布资格位于原生消费者边界，避免为了拒绝零调用而破坏历史报告的读取与合并。仅六条通用 chat 使用该固定合同，地区、GLM、MiniMax、Jev 和图像/分类仍走各自原生规则。消费者内部格式直接迁移到外层 schema 2，旧自采集格式拒绝，无兼容层。

中间 GREEN 尝试亦保留：dependency-green/green2 的新场景没有可隔离 payload，因而在 purge 阶段失败；补上无引用的成功对照目录后验证通过，未修改清理实现以放行空批次。consumer-outer-green 漏接外层 elapsedSeconds schema，原生成功但发布判定拒绝；补齐类型并保留严格未知字段拒绝。consumer-outer-green2 的文件限额只比基准小一字节，时间戳精度差异使注入不稳定；改为预留64字节差值，未放松 I/O 失败断言。RED 之后的公共进程测试全部保留，不以重跑覆盖。

本轮全部临时清理和故障注入均在假仓库进行；脱敏、取消与 Go 配置隔离继续由完整工具场景验证。正式普通回执、来源冻结、只读发布判定及当前状态在收尾填写，未开展后续阶段。

最终冻结 `3bce7a61cb85662e25faf13137feee723c226503`：47 组工具进程 `.evidence/barness-ai-test-process-review/e2e/rereview-final-process-20261022T035646.133488Z/command-receipt.json` exit0；完整普通 `.evidence/barness-ai-test-process-review/e2e/20261022T035647.866871Z-jvpjotru/receipt.json` PASS / exit0、12890 实际通过节点、原生 {'PASS': 12086, 'NOT_RUN': 214}、audit0。普通通过 case 集合不变，1247 份执行源码和 8636 份历史输入相等。双 vet 与 live-tag 仅编译通过；当前状态和只读判定统一见 status.md / verification.json。
