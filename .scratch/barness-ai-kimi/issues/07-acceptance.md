# 07 — 集成验收、双轴审查与发布状态

Type: task
Status: ready-for-human
Blocked by: 02, 03, 04, 05, 06, 08, 09

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

K01–K16；完整受影响普通包、源码冻结、Standards/Spec 审查及修复、唯一当前状态和不可变回执。live/race/压力须按 SOP 独立授权；未执行不得声称发布通过。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：01–06 已 resolved；在集成提交 `ed6d4aba84b340efe99a7b909974537e4779c954` 认领最终验收。双轴审查以工作基线 `68b7da0a3400f909dc535d7f585c93771df1569c` 为固定比较点；修复后冻结完整来源/实际输入并在主工作区执行完整受影响普通包。真实资格、晋升、race 与压力继续按 SOP 等待各自授权。

## Answer

集成实现、同一 implementer 的审查修复、两轴复核与冻结完整普通验收已完成；13956个必需原生用例 PASS，来源/129项输入及整目录集合相等。同一原生读取器接受完整14170-case bundle，路径修正前的失败及编排干扰中断均保留。实际受测 SHA、argv/exit、审计和不可变索引统一见 [唯一当前状态](../spec.md#唯一当前状态)。八个实施 worktree 已清理且原件/分支保留。

本票尚待六路线自身候选与正式资格、审阅晋升及所需人工阶段。六路线候选已获批准；首个中国站 Chat 用例因服务未报缓存读取计数而失败，按批准范围停止其余调用，见 [不可变执行结果](../evidence/candidate-live-result-01.md)。由 [08](08-live-usage-reporting.md) 实施用户批准的可选读取默认0和有界费用修订，随后重新冻结普通验收与自身资格；原 grant 不复用。正式验证、race/压力尚未授权，不能 resolved 或宣称 release.pass。

2026-10-23：用户已明确批准08的合同/代码修订；旧普通与真实结果仅保留其原来源事实，本轮完成红绿/审查/冻结普通后再处理新的付费范围。

2026-10-23：08已resolved：缓存读取默认0和有限成本边界实现、双轴审查及冻结完整普通验收通过（14045必需PASS）。旧真实失败保留；新的六路线候选范围和源绑定构建已备妥，等待本次明确付费批准。正式资格、晋升和人工阶段仍待处理；当前状态及新不可变索引统一见spec。

2026-10-23：修订后授权已执行，三条中国站candidate完整PASS；国际站Chat第4项因账户RPM=3返回429，首失败停止，国际站另两条NOT_RUN。31调用/31Attempt、零重试；见[结果02](../evidence/candidate-live-result-02.md)与新阻塞[09](09-international-rate-limit.md)。正式资格及人工阶段仍未获准。

2026-10-23：按用户要求加入35秒原生间隔，完成双轴复核与新冻结完整普通（14071必需PASS）；当前来源国际站Chat9/9及自身资格PASS。Responses第4项服务端SSE过载，范围停止，Messages未跑；[10](10-responses-engine-overload.md)等待新scope，[11](11-responses-stream-error-diagnostics.md)记录诊断分类待triage。原件见[结果03](../evidence/candidate-live-result-03.md)，当前发布仍false。
