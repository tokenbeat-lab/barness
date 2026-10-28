# 人工验证、正式目录晋升与发布判定

Type: task
Status: ready-for-agent
Blocked by: 05

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R21/R23；E18。按维护者已确认的[未发布分支顺序](../formal-qualification-plan.md)先准备并审查最终目录/资格源码，再冻结最终来源取得正式live；有效live后分别取得人工race和适用压力选择，全部通过后才接受/发布正式晋升并保存只读Release Gate。目录源码接线不代表已经正式支持。

## Ownership

正式激活数据、独立发布索引及唯一当前状态。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
