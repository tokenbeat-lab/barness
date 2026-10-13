# 12: 取得 Luna 当前正式目录下的独立中文效果报告

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 10: 纳入正式目录并形成第十组合独立支持记录；11: 交付 Host 迁移示例与独立拒答评估流程
**Source tickets:** 08
**Requirements:** R18, R19
**Acceptance cases:** OD55
**Updated:** 2026-10-12

**What to build:** 维护者对同一固定中文任务实际运行 Luna，得到与 Jev 完全独立、基于正式目录的准确率、拒答、延迟和消耗报告；未执行或失败不能变成质量结论。

## Responsibility

负责独立 live 评估执行、当前适用性和真实 evidence bundle；沿用已验证工具，不在此票重写评估引擎。Jev 旧报告按原事实保留，当前发布需其自身适用报告。协作时保留他人改动。

## Failure scenarios first

- fixture、candidate Hash 或 Jev 报告充当正式 Luna 效果。
- 两协议同一进程/错误 key、沿用 Jev threshold 或全拒答被称为高准确率。
- 调用/时间预算超额仍 COMPLETE，漏失败样本，取消继续发送。
- 新目录后修改旧报告日期/Hash，或把协议冒烟当效果评估。

## Acceptance criteria

- [ ] 执行前核验工具票已有的拒答、错 route、预算/取消反例和第一方 source；新增失败方式先由 public process 表达。
- [ ] Luna 正式 route/model/version/Hash、固定 dataset/config/code Hash 与独立 key/账户别名真实一致；记录实际 ResponseModel 和日期。
- [ ] 一进程仅一协议，最多 24 Call/Attempt、零重试、串行、总 600 秒；预算与准确命令可核验。
- [ ] 24 样本均有状态；合法 refusal 计入任务分母、可核验完整性，不计正确，不伪造概率或置信。
- [ ] 两种准确率、混淆、拒答率、confidence/chosen probability、ECE/Brier、延迟与费用独立呈现，null 指标和未上报用量不被零填。
- [ ] 缺 key 形成 NOT_RUN 且不完成；实际失败 FAIL。COMPLETE 仅任务完整性，不设置或降低质量阈值。
- [ ] 新包和旧 Jev bundle 核验分别保留，Jev 当前报告适用性明确；过期则由最终发布票收集自己新证据，不改旧包。
- [ ] actual captures、消费结果、Attempt、Observer、完整样本、manifest/audit 与重放命令齐备，报告不写入支持矩阵。

## Completion

仅该 route 自己的正式实际完整报告与审计成立才完成本票。效果与协议支持保持独立，不能据小样本保证生产质量。
