# Host 装配、生命周期、回滚与普通回归

Type: task
Status: ready-for-agent
Blocked by: 02, 03, 07

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R03/R04、R16–R20、R22/R23；E13–E16/E18。完善公共 E2E、有限 live 场景入口及 Host 文档；两轴审查并修复，冻结来源后完成受影响普通包回执。

## Ownership

xAI Host/运行/回滚 E2E、live 候选入口/配置、Host 装配文档；按需协调共享接线。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
