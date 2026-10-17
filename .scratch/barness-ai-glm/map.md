# GLM 实施地图

## Notes

- [Spec](spec.md)、[依赖图](ticket-plan.md)、[研究](research.md)。
- 集成分支：`codex/barness-ai-glm`。
- 基线：`f5cf5362610a8bdb8f3e170f408f3858c149e540`。
- 公共 Client E2E 接缝已在 spec T01 获得确认；不另设业务验收入口。

## Decisions-so-far

- 2026-10-17：按 implement-spec 指令发布既有 13 张工单，按阻塞图实施。
- 2026-10-17：[01 已解决](issues/01-provider-capability-prefactor.md)：Provider 自动字段边界完成；既有 Client 前后 10,598 artifact 零未解释差异、适用冻结差分零 pending。见 [ADR-0032](../../docs/adr/0032-barness-ai-provider-automatic-fields.md) 与[运行索引](evidence/ticket-01/README.md)。

- 2026-10-17：[02 已解决](issues/02-glm-chat-text.md)：zhipu Chat 独立候选目录、四入口文本/思考、型号与最终请求硬约束、标准 Key Host 示例与 120 秒候选装配完成；最终 166 例、受影响冻结差分 321 例通过。见 [ADR-0033](../../docs/adr/0033-barness-ai-zhipu-chat-candidate.md) 与[证据索引](evidence/ticket-02/README.md)。工具、其他协议和正式支持仍由后续票验收。

- 2026-10-17：[06 已解决](issues/06-glm-messages-text-contract.md)：Messages 独立候选与真实 SSE 文本回放完成，GLM 默认省略思考且不附加 Claude 扩展；66 公共 Client/66 race、104 原生回归、168 原生运行时及 70 冻结差分通过。见 [ADR-0035](../../docs/adr/0035-barness-ai-zhipu-messages-basic.md) 与[证据索引](evidence/ticket-06/README.md)。仅基本文本，07/11/12 继续增强、正式 live 和内置登记。

- 2026-10-17：[04 已解决](issues/04-glm-responses-text-shapes.md)：Responses 原始JSON对象/数组边界、权威done/交错索引和独立候选请求约束完成；原生reasoning保留未转换形状。公共回归1978 PASS，冻结Responses差分335 PASS；真实数组重放与schema对象红绿证据见[索引](evidence/ticket-04/README.md)及[ADR-0034](../../docs/adr/0034-barness-ai-zhipu-responses-boundary.md)。05继续原生历史请求转换。

- 2026-10-17：[03 已解决](issues/03-glm-chat-tool-round-trip.md)：两型号工具分片/Host 往返、同轮推理与信任降级、auto-only/clear_thinking 最终约束及精确输出预算完成；已提交完整 Chat 358 PASS、新增 192 例 race PASS、受影响冻结差分 321 PASS 零 pending。见 [ADR-0033](../../docs/adr/0033-barness-ai-zhipu-chat-candidate.md) 与[证据索引](evidence/ticket-03/README.md)。

- 2026-10-17：[07 已解决](issues/07-glm-messages-tools-thinking-replay.md)：两型号 Messages 工具/原生签名与无签名 thinking 回放、5.2 唯一已证实 budget1024/max4096 full 形状、来源降级及最终能力/图片/SSE 顺序校验完成；193 公共 Client、190+3 race、104 原生回归、168 运行时和70冻结差分通过。见 [ADR-0035](../../docs/adr/0035-barness-ai-zhipu-messages-basic.md) 与[证据索引](evidence/ticket-07/README.md)。无 canonical 强度映射/预算范围/redacted 支持或正式 live 发布声明。

- 2026-10-17：[08 已解决](issues/08-glm-failures-retry-usage.md)：三协议独立错误码、安全诊断、不可恢复429零重试及固定快照有界重试完成；用量缺项/零/缓存/推理/断流保留，GLM不套用Anthropic一小时缓存元数据或溢价。最新合并GLM1108 PASS、原生回归2538 PASS、冻结三协议差分608 PASS零pending；见[证据索引](evidence/ticket-08/README.md)与ADR0006/0010。

## Fog

- 候选协议、真实标准 Key 冒烟和人民币价格换算证据均待取得；不得以离线通过代替正式支持。
