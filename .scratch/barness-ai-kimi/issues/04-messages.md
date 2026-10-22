# 04 — Messages 协议

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

R03、R06–R09、R11；K03–K10、K12–K14。拥有 anthropic_* 与原生 thinking 来源接线、自身公共 Client E2E；稳定资源、Bearer、独立 effort 与顶层缓存。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：Messages implementer 已在 codex/kimi-messages 领取；基线集成分支 8678bc9；公共 Client 受控协议 seam；不运行 live/race/压力。

## Answer

已实现并通过普通受控验收；实现提交 `d6a65ae`、`7a2aa91`、`0482300`、`66eab99`。覆盖范围、原始失败/成功回执及未跑项见[工单证据](../evidence/ticket-04.md)。最终受测 `66eab996b860801d188507d071e8e572e355b658`，退出 0，665 个叶测试 PASS（Messages 自身 196），bundle/audit 有效；`./ai` 仅完成该表达式下的编译/包执行，完整共享包验证由集成阶段执行。本工单完成不表示 live/正式资格或发布通过。
