# GLM 实施地图

## Notes

- [Spec](spec.md)、[依赖图](ticket-plan.md)、[研究](research.md)。
- 集成分支：`codex/barness-ai-glm`。
- 基线：`f5cf5362610a8bdb8f3e170f408f3858c149e540`。
- 公共 Client E2E 接缝已在 spec T01 获得确认；不另设业务验收入口。

## Decisions-so-far

- 2026-10-17：按 implement-spec 指令发布既有 13 张工单，按阻塞图实施。
- 2026-10-17：[01 已解决](issues/01-provider-capability-prefactor.md)：Provider 自动字段边界完成；既有 Client 前后 10,598 artifact 零未解释差异、适用冻结差分零 pending。见 [ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md) 与[运行索引](evidence/ticket-01/README.md)。

## Fog

- 候选协议、真实标准 Key 冒烟和人民币价格换算证据均待取得；不得以离线通过代替正式支持。
