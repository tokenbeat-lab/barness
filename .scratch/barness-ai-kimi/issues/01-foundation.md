# 01 — 身份、候选目录与成本范围

Type: task
Status: claimed
Blocked by: none

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

R01–R03、R09–R10；K01、K02、K10–K11、K14–K15。拥有 Provider 常量、Kimi 候选目录/验证、Attempt 成本值、共享成本解析与调用快照接线。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：由 foundation implementer 在 `codex/kimi-foundation` 认领；基线 `8678bc951ddaed20b0e5d16314ef791ed19ed156`。公共 Client E2E seam 及验证范围沿 spec；不启动 live/race/压力。
