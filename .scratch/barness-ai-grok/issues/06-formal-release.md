# 人工验证、正式目录晋升与发布判定

Type: task
Status: ready-for-agent
Blocked by: 05

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R21/R23；E18。live 资格成立后取得人工选择的 race，再取得适用压力选择；满足来源/价格/能力证据后才激活正式目录并保存只读 Release Gate 回执。

## Ownership

正式激活数据、独立发布索引及唯一当前状态。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
