# 07 — 集成验收、双轴审查与发布状态

Type: task
Status: ready-for-agent
Blocked by: 06

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
