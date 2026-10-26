# 本组合真实资格与捕获回放

Type: task
Status: ready-for-agent
Blocked by: 04, 07

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R21/R24；E17。冻结来源后列出精确场景、账户/地区、预算及所有受影响路线，取得本次明确授权才执行付费验证。保留非零 reasoning、续接、工具及图片捕获；先失败后接入回放，源码变化需重新冻结。

## Ownership

本路线真实证据、捕获夹具与资格记录。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
