# 测试流程重置当前状态

日期：2026-10-22（Asia/Shanghai）。入口迁移、普通修复与后续流程规范已启用；实际人工阶段及真实清理分开记录。

## 流程启用

- 已完成 plan.md 第 2–5、7 节流程规则调整，以及第 8 节规范交付部分；步骤与完成条件统一在[测试流程 SOP](../../docs/agents/verification.md)，AGENTS.md 已明确要求读取。
- [逐节落实记录](rollout.md)列明日常先失败后实现、审查冻结、分阶段验收、五项只读判定、证据保留和旧入口退役。地区无 Key 示例改为隔离环境的普通回执入口；真实原生资格另接外层 schema 2 消费者。
- 本轮仅修改 Markdown 与本功能状态/验证投影，没有修改执行源码或实际输入。文档检查通过：175 个本地链接/锚点、7 份 shell 模板、9 个步骤的完成条件、AGENTS 原有测试原则及 CLI 参数合同；执行来源和选定输入的集合/哈希均相等。
- **新流程首次完整大功能验收：待验收**。三项真实维护修复已有先红后绿实例；下一真实大功能的 live/人工阶段闭环尚未发生，不能以 SOP 交付冒充完成。
- **真实隔离与存量清理：NOT_RUN**。没有真实 cleanup scan/plan/move/verify/restore/purge，没有观察期起点、PURGED 或回收量；清理工具仍只通过临时假仓库场景。

## 保留的普通验收事实

- 实际受测/实现冻结：`3bce7a61cb85662e25faf13137feee723c226503`；三项复审修复收尾 `bfcd88eb9a4707c55a6e8fe0b79bbf39420ac95f`。修复见 [rereview-fixes.md](rereview-fixes.md)，更早七项与脱敏见 [review-fixes.md](review-fixes.md)。本轮没有重新执行 Go 测试、vet、live-tag 编译或工具测试，不把文档提交写成新测试。
- 适用普通回执仍为 47 组工具进程 exit0；完整普通 Go exit0、12,890 个通过节点、原生 12,086 PASS / 214 NOT_RUN、audit0、墙钟375.15秒；两种 vet、live-tag 仅编译通过。来源1,247份，选定历史输入8,636份与当前相关集合/字节相等。
- race：NOT_RUN；正式压力：NOT_RUN。此前误跑旧负载的事实仍在 [execution-log.md](execution-log.md)，本轮无新增执行。
- 付费 live：NOT_RUN；本功能不改厂商运行行为，当前索引 live 有依据地 NOT_APPLICABLE。合成与离线回放不产生真实资格。
- release.pass：false；本功能需要的人工 race、正式压力仍缺失。新只读判定仅消费现有事实，不启动阶段。

[验证索引](verification.json)保存旧普通事实及本轮文档检查；[新的不可变索引](../../.evidence/barness-ai-test-process-review/record-check-sop-rollout-20261022T042545Z-3k9_kp8y/record.json)、[只读判定](../../.evidence/barness-ai-test-process-review/record-check-sop-rollout-20261022T042545Z-3k9_kp8y/report/release-report.md)、[收尾来源核对](../../.evidence/barness-ai-test-process-review/record-check-sop-rollout-20261022T042545Z-3k9_kp8y/final-source-check.json)绑定当前状态。上一状态和验证索引另存本轮目录的 previous-status.md / previous-verification.json；旧失败、旧报告和已消费授权保持不变。

地区资格和恢复7继续由其原规格唯一当前状态维护；没有恢复该轮、创建恢复8、发行或付费授权。本功能的[执行方案](plan.md)保留原验收合同，日常操作按 SOP。
