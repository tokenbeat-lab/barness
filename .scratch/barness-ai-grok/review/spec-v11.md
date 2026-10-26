## Spec

只读审查固定差异 `4c9cd7c6...268faebaba0874248a41539fbf5069e4276e167e`；未运行测试、构建或厂商调用。

- **[P2] 完整 native fixture 的公共 CLI 回归缺失（partial）。** 工单 [17:21](/Users/cyber/RestoX/barness/.scratch/barness-ai-grok/issues/17-declared-unsupported-qualification.md:21) 明确要求“先通过公共CLI建立完整OpenAI Chat native fixture的RED”，且保留其余完整场景与一致预算。新增 CLI 用例仅写入 `live-report.json`（[测试:161](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_qualification_test.go:161)），没有 manifest、case/assertions 或捕获库存，最终只断言统一的拒绝结果（[测试:183](/Users/cyber/RestoX/barness/ai/e2e/generic_unsupported_qualification_test.go:183)）。报告和库存反例分别直接调用两个内部合同函数（同文件 :95、:214）。因此，即使生产消费者恢复旧的全部 PASS 检查，或漏接新增库存检查，这些回归仍可全部通过；现有 RED/GREEN 没有覆盖工单指定的消费者接线。应补齐既有公共边界上的完整回放和可验证分支结果；合成材料最终保持无正式资格，与验证本次接线并不矛盾，也不要求付费重跑。

未发现其他缺失、范围扩大或实现错误。固定身份/型号/原因及零活动例外、其他场景完整 PASS、mandatory xAI history 和支持矩阵不新增 LastPassedAt 的边界均保留。当前普通验证、重新冻结、付费与人工阶段的如实 pending 不计为问题。

合计：1 项；最高 P2。
