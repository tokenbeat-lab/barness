# 初步性能结论

稳定实现来源 `0433459`。一次完整普通验证墙钟 **633.110 秒（10分33秒）**，9个有测试的包全部通过，差分/压力开启。记录54,438条事件，无非事件行；[实际命令](../../.evidence/race-optimization/validation-0433459/ordinary-command.json)、[元数据](../../.evidence/race-optimization/validation-0433459/ordinary.events.jsonl)、[结构化分析](performance.json)。没有执行实际 race，放大比例待人工选择后取得。

## 包时间

包并行执行，不能相加作为总墙钟。最后一列是该包最后一条测试结束事件到包结束的差额，包含测试框架收尾、证据 Finish/Audit、退出/输出等；当前没有各阶段独立计时，不能命名为审计耗时。

| 包 | Go包耗时（秒） | 最后测试后差额（秒） |
| --- | ---: | ---: |
| `e2e` | 631.409 | 88.503 |
| `internal/testkit/release` | 39.124 | 0.012 |
| `internal/testkit/minimaxlive` | 22.425 | 0.001 |
| `release/cmd/releasegate` | 14.621 | 0.000 |
| `examples/visualeval` | 13.243 | 1.961 |
| `internal/testkit/supportmatrix` | 3.500 | 0.001 |
| `internal/testkit/pioracle` | 2.561 | 0.001 |
| `internal/testkit/audit` | 2.110 | 0.001 |
| `github.com/tokenbeat-lab/barness/ai` | 1.573 | 0.001 |

## 顶层热点

只比较顶层测试，不叠加父子测试。子测试明细保存在 performance.json。

| 测试 | 秒 |
| --- | ---: |
| `TestPiDifferential` | 178.590 |
| `TestGLMFormalReports` | 30.320 |
| `TestPolicyPressure` | 27.840 |
| `TestVisualEvaluationFailureRetention` | 22.490 |
| `TestVisualEvaluationIntegrity` | 21.300 |
| `TestFormalRenewalAuthorityProcess` | 17.000 |
| `TestMiniMaxLiveEvidenceTools` | 15.990 |
| `TestProtocolTimeouts` | 11.120 |
| `TestReleaseGatePreflightStopsCommands` | 10.790 |
| `TestClassifierImagesPressureImport` | 9.530 |

Pi差分178.590秒是最大的单项。源码 `ai/internal/testkit/pioracle/oracle.go:142–149` 明确每次 Run 使用空环境的新 Node 进程；这些是实际外部子进程，Go race 不会测量 Node 内部。现有数据没有分开 Node 启动、冻结来源核验、协议交互和断言耗时，不能把178秒全部归因于进程启动。

主E2E包最后测试后的差额88.503秒，足以列为下一轮测量对象。GLM正式导入30.320秒、视觉失败留存/完整性22.490/21.300秒、MiniMax工具15.990秒也值得分段测量。当前证据支持定位这些场景，尚不足以决定复用进程或缓存实现。

## 下一轮建议（未实施）

1. 先分别测量 Oracle 的启动/导入、provenance校验、交互与结果解析；测量使用现有流程，不减少差分、协议或独立实例覆盖。
2. 给现有 TestMain 的 Finish、Audit 增加独立计时，确认88.503秒尾部的组成，再讨论重复解析或夹具复用。
3. 对 GLM/视觉/MiniMax 子进程场景拆开 build/start 与消费耗时；只有确认重复构建为热点，才设计同来源的既有二进制复用。取消、来源与隔离语义继续独立验证。
4. 人工实际选择完整 race 后，补各包及同名测试放大比例，优先优化放大最大的已定位部分。既有5次全量race约37.0–48.9分钟，可作等待估计；本次未取得新race性能结论。

本轮已验证人工触发前实际 race 启动为0，保留正式完整集合。未实施缓存、内容寻址存储、通用调度器、排除race包或改变并行度。单轮提速与总耗时降低比例不作未经测量的承诺。
