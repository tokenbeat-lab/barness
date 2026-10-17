# MiniMax 实施任务图

Status: claimed
Integration branch: `codex/minimax-cn`
Base: `ecfd4b4`

## Notes

2026-10-18 用户调用 implement-spec；原目录只有规格与研究，本次按本地 tracker 补齐任务图。测试接缝由 spec T01 明确确认，不再重复确认。采用独立 worktree 和实现/合并 subagent。

## Frontier

01 → {02,03,04} → {05,06} → {07,08} → 09。05与06相互独立；07真实证据/发布和08文档/回滚独立推进，09汇合最终判定。

- [01 Provider 身份、六个候选型号与思考元数据](issues/01-provider-model-contract.md)；Blocked by: None。
- [02 MiniMax Chat 文本、推理、工具历史与边界](issues/02-chat-protocol.md)；Blocked by: 01。
- [03 MiniMax Responses 无状态流、工具与可信续接](issues/03-responses-protocol.md)；Blocked by: 01。
- [04 MiniMax Messages adaptive 思考、工具和签名续接](issues/04-messages-protocol.md)；Blocked by: 01。
- [05 三协议隔离、失败重试、回调和资源释放](issues/05-isolation-failure-resources.md)；Blocked by: 02, 03, 04。
- [06 固定 MiniMax pi 差分和扩展证据](issues/06-pi-differential.md)；Blocked by: 02, 03, 04。
- [07 六身份候选、M3 正式证据和目录发布门禁](issues/07-live-catalog-release.md)；Blocked by: 05, 06。
- [08 Host 文档、ADR 和回滚实现](issues/08-host-docs-final-acceptance.md)；Blocked by: 05, 06。
- [09 完整验收、双轴审查与交付判定](issues/09-integrated-acceptance-review.md)；Blocked by: 07, 08。

## Decisions so far

- 08文档/回滚不需要等待07真实报告才开始；已将最终验收独立为09，避免无关依赖，全部交付要求保持。

- 实施不得把缺真实凭据改写为正式通过；M3 三协议正式目标与 M3.1 三候选离线约束全部保留。

## Fog

- 实际凭据资格、三协议真实推理续接形状和价格来源，按 R25–R27 收敛。

- 01 resolved：共享 MiniMax 中国身份、默认采样标签及六候选装配完成；[Answer](issues/01-provider-model-contract.md#answer)、[证据](evidence/ticket-01/README.md)、ADR-0037。三协议为下一 frontier，所有候选保持非正式。
- 02 resolved：MiniMax Chat 独立能力、最终守卫、完整工具/推理历史与流末用量完成；[Answer](issues/02-chat-protocol.md#answer)、[证据](evidence/ticket-02/README.md)。579 自身 E2E/race 与 883 旧行为回归通过；真实 M3 分离格式/原生 auto 仍归 07，未宣称已证实。

- 03 resolved (offline scope): independent Responses capability/guards, authoritative stream handling and opaque native/tool history; [Answer](issues/03-responses-protocol.md#answer), [evidence](evidence/ticket-03/README.md), ADR-0039. Lossless reasoning input projection remains provisional until own live07 verifies it; no formal support claim.
