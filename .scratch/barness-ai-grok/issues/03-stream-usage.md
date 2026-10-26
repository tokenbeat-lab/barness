# 权威流快照、工具合同、用量与估算成本

Type: task
Status: ready-for-agent
Blocked by: 01

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R11/R12、R14/R15、R17/R18/R20；E05/E06/E10–E12。xAI 私有状态绑定 Responses adapter，严格关联、done/final-only、重复与冲突、真实终态及原始数字；独立成本完整性。

## Ownership

Responses 流接线、xAI stream/usage/cost 文件及对应 E2E；不拥有请求/history 文件。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
