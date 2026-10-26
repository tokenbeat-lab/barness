# 无状态请求、选项、最终守卫与可信历史

Type: task
Status: ready-for-agent
Blocked by: 01

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R05–R10、R12/R13、R16、R19；E02–E04/E07–E09/E13。复用 Responses profile，明确最终字段子集；预算/采样前置；完整 transcript、opaque ID、来源降级与严格图片。

## Ownership

Responses 请求/选项/profile、xAI request/guard/history/ID codec 及对应 E2E；不拥有流与 usage 文件。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
