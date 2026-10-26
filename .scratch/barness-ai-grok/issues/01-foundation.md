# 身份、候选目录与前置授权

Type: task
Status: ready-for-agent
Blocked by: none

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R01–R04、R21；E01/E02/E04。公开 ProviderXAI；新增可明确装配的候选 Catalog（不激活 BuiltinCatalog），固定 grok-4.7 与全球目标。校验 Host 元数据只能收窄且非法组合在凭据前拒绝。

## Ownership

ai/binding.go、ai/catalog_xai.go、ai/catalog_validate.go、ai/call_runtime.go 等身份/目录/路由文件；对应 xai_foundation E2E。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

待实施。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
