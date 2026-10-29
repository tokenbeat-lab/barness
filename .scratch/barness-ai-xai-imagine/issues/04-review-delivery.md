# 集成审查、普通验证与交付

Type: task
Status: resolved
Blocked by: 01, 02, 03

## Scope

调用 code-review 双轴审查并修复；冻结来源；执行完整受影响普通验证。维护唯一当前状态与不可变证据索引。人工阶段未批准保持 NOT_RUN，发布保持 false。

完整契约：[spec](../spec.md)。验证遵循 [SOP](../../../docs/agents/verification.md)。

## Comments

2026-10-29：implement-spec 根据现有规格建立工单依赖图。

## Answer

2026-10-29：双轴审查及同一实现者修复完成，最终完整ai/...普通验证PASS/exit0；首次全包FAIL和全部红绿原件保留。不可变来源/普通索引、只读发布结果、未执行人工阶段和工作区保留原因统一见[spec当前状态](../spec.md#唯一当前状态)。真实候选/正式资格、race、压力和发布不属于本次已完成范围。
