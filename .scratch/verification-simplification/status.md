# 验收精简当前状态

2026-10-27 后续进展：六条缺来源路线已完成离线追溯，6/6 APPLICABLE、原件缺口 0；合并留存索引为 16/17 可沿用，MiniMax 国际 Messages 仍 incomplete/FAIL。新来源、当前验收与测试隔离异常的唯一完整状态见[来源追溯状态](../provider-source-retrace/status.md)及[当前资格索引](../../.evidence/provider-source-retrace/current-qualification-index-v1.json)。以下数字保留精简交付时点；原文已[另存](../../.evidence/provider-source-retrace/simplification-status-before-source-retrace.md)。

2026-10-27：实施完成，普通验证通过；**release.pass=false**。本次未发起付费 live、race、压力或暂停轮次。

基线及受测 HEAD：`69ff4af880e68ef3ca245e5128bdae5e59a2ef7b`。变更仍在未提交工作区，受测身份以回执的 1997 个 `Source.Files` 与 179 项实际输入为准，不能把 HEAD 当成已提交的实现。来源文件集合摘要：`849cf072ce74ef4d1befe728137bb5168a3ccba2ae0f635e4a9d8ad72129fb9e`。

唯一交付事实索引：[implementation-result-v1.json](../../.evidence/verification-simplification/implementation-result-v1.json)。已批准的范围见 [spec](spec.md) 与[精简方案](../provider-integration-audit-20261027/simplification-plan.md)。

## 已完成

- 先采用 [ADR-0055](../../docs/adr/0055-route-scoped-evidence-applicability.md)，再修订 SOP、工具及地区合同；旧 ADR 明确替代范围。
- 原测量、当前适用性和新执行权限分开。保留原完整来源、目录哈希和测量时间；按实际 Provider、协议及已使用型号判定当前适用性。未知产品代码、新子包及未细分共享依赖保守要求审查，不自动授权付费。
- 阶段各自绑定实际输入，可引用其他功能和原提交。旧 live 文档可按原提交核验原字节；找不到原件继续拒绝。普通阶段仍要求当前来源与实际输入一致。
- 当前执行统一新的一次 grant/claim 和本过程账本，保留原预算、一次领取、取消和原生语义；删除旧续期/恢复 fallback、前序递归和历史账本拼接。
- 退役 173 个源文件（78 Go、95 Python，原 21644 行）；计入修改及新代码后净减少 21123 行。含原计划 66 Go + 94 Python、重复 CLI 及历史专属测试/包装，见[逐文件留存索引](../../.evidence/verification-simplification/history-before-retirement/final-retired-source.json)。
- 资格集中到 `releasegate`；来源与模板由 `processgrant` 提供。同次资格判定不再重复构造整包；原件语义验证与独立脱敏审计保留。Go → Python → Go 循环已摘除，原生工具编译后直接执行；真实 xAI 产物消费移至工具进程套件。

## 验证

| 范围 | 实际结果 | 原件 |
| --- | --- | --- |
| 完整受影响普通包 | PASS，exit0，600.258 秒；来源/实际输入执行前后一致 | [最终回执](../../.evidence/verification-simplification/e2e/20261027T072122.475447Z-0n90pm7b/receipt.json) |
| 主 E2E 原生包 | 15141 PASS；214 个原有 opt-in/手工场景 NOT_RUN，未混为通过 | 同上唯一 `./ai/e2e` bundle |
| 辅助包 | GLM 32、MiniMax 原生工具 77、探针 6、visualeval 25 全 PASS，audit 均为空 | 同上各辅助 bundle |
| 工具进程套件 | 56/56 PASS；包括真实选项产物、跨阶段输入、原提交输入、取消/最终写失败及只读消费 | [日志](../../.evidence/verification-simplification/tool-process-suite-05/unittest.log) |
| 静态检查 | 受影响包及两个当前 CLI 的 `go vet` exit0 | [回执](../../.evidence/verification-simplification/static-validation-02/receipt.json) |

原失败场景已转绿；早期 RED、FAIL、CANCELLED、INVALID_EVIDENCE 均独立保留，见[attempt 索引](../../.evidence/verification-simplification/attempts-index-v1.json)和[命令日志](../../.evidence/verification-simplification/command-logs/)。最终回执才代表当前普通验收，零匹配或中断不算完成。

## 精简交付时的旧测量判定（历史记录）

17 条留存路线已生成新的标准 schema2 外层只读消费回执，当前来源与最终普通来源完全一致；原 bundle 全部文件核对前后一致。未更新支持矩阵或原测量时间，未把迁移当成新测量。

| 当前结论 | 路线 | 后续范围 |
| --- | --- | --- |
| 10 条 APPLICABLE | xAI Responses；智谱 Chat/Messages；Z.AI Chat/Messages；MiniMax 国际 Chat/Responses；MiniMax CN Chat/Responses/Messages | 可引用新只读消费回执；没有新增厂商调用 |
| 6 条待补来源原件 | OpenAI Responses/Chat；DeepSeek Responses/Chat；Anthropic Messages；Google Gemini | 旧 bundle 缺可认证原执行来源快照，先找留存来源并离线核验，不默认重买 |
| 1 条仍拒绝 | MiniMax 国际 Messages | 原真实测量 incomplete/FAIL；本次不补执行、不恢复原暂停轮次 |

[逐路线回执与哈希](../../.evidence/verification-simplification/current-applicability-03/report.json)，[具体缺口](../../.evidence/verification-simplification/current-applicability-03/diagnostics.json)。候选复用仅在正式合同确实完整且当前型号适用时成立，原 stage/mode 不改；不据此晋升 Kimi。

## 历史与用户修改

- 原四份历史消费者构建 exit0 后封存，未执行；原 160 个目标文件与基线 Git 逐字节一致，追加退役文件也绑定原提交与 SHA。见[构建回执](../../.evidence/verification-simplification/history-before-retirement/build-receipt.json)。
- 175 项原记录输入及原当前状态可定位，150 项原 live 引用完整；原授权、claim、账本、失败与暂停状态保留，支持矩阵与时间戳未改。修订的历史资格说明，其原字节单独封存，见[留存核对](../../.evidence/verification-simplification/history-retention-check-01.json)。
- 原 17 条 live 回执各 165 项输入均可核验，其中每条 1 项已更新的文档由原提交提供原字节，缺失为零，见[阶段输入核对](../../.evidence/verification-simplification/historical-stage-inputs-check-01.json)。
- 用户原有 AGENTS.md 修改及两个 scratch 目录未动。未做真实隔离、purge 或旧权限恢复。

## 未执行与回滚

付费 live、race、压力均 **NOT_RUN**；暂停轮次 **NOT_RESUMED**。claim/锁属于实际并发影响，race 不伪报不适用。本次实施完成不等于完整发布通过，也不发放下一阶段权限。

回滚按本次变更集成组回退实现、ADR/SOP 和当前入口；保留新增及历史回执，不改已消耗 grant/账本，不恢复旧执行权限，不覆盖用户原有修改。源文件可从留存索引指定的固定 Git 提取。
