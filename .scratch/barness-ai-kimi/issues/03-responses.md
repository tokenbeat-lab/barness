# 03 — Responses 协议

Type: task
Status: resolved
Blocked by: 01

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

R03、R05、R07–R09、R11；K03–K10、K12–K14。拥有 responses_* 与自身公共 Client E2E，summary 输入投影、不透明 ID、终态、缓存及最终校验。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：Responses implementer 已领取，独立分支 codex/kimi-responses 以集成分支 3e25c0e 为基线。公共 Client E2E seam 已确认；保留红例后实现，按影响覆盖 Responses 回归，隔离 checkout 缺少无关历史归档的全库范围由根任务在主 checkout 验证。live/race/压力本票不执行。

## Answer

Implemented Responses protocol and scoped history identity correction on `codex/kimi-responses`. [Immutable receipt](../evidence/ticket-03.md) records exact source, argv, exit0, case inventory, bundle/audit hashes and retained failures; [retention proof](../evidence/ticket-03-retention-20261022T1456.json) binds complete file-set/content equality. Both markets and affected old Responses/cross-protocol histories pass ordinary checks. Root owns final combined acceptance and feature current status; live/race/pressure remain NOT_RUN.
