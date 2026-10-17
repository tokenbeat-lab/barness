# MiniMax 实施任务图

Status: claimed
Integration branch: `codex/minimax-cn`
Base: `ecfd4b4`

## Notes

2026-10-18 用户调用 implement-spec；原目录只有规格与研究，本次按本地 tracker 补齐任务图。测试接缝由 spec T01 明确确认，不再重复确认。采用独立 worktree 和实现/合并 subagent。

## Frontier

01 → {02,03,04} → {05,06} → 07 → 08。05 与 06 相互独立。

- [01 Provider 身份、六个候选型号与思考元数据](issues/01-provider-model-contract.md)；Blocked by: None。
- [02 MiniMax Chat 文本、推理、工具历史与边界](issues/02-chat-protocol.md)；Blocked by: 01。
- [03 MiniMax Responses 无状态流、工具与可信续接](issues/03-responses-protocol.md)；Blocked by: 01。
- [04 MiniMax Messages adaptive 思考、工具和签名续接](issues/04-messages-protocol.md)；Blocked by: 01。
- [05 三协议隔离、失败重试、回调和资源释放](issues/05-isolation-failure-resources.md)；Blocked by: 02, 03, 04。
- [06 固定 MiniMax pi 差分和扩展证据](issues/06-pi-differential.md)；Blocked by: 02, 03, 04。
- [07 六身份候选、M3 正式证据和目录发布门禁](issues/07-live-catalog-release.md)；Blocked by: 05, 06。
- [08 Host 文档、ADR、完整验收和回滚](issues/08-host-docs-final-acceptance.md)；Blocked by: 07。

## Decisions so far

- 实施不得把缺真实凭据改写为正式通过；M3 三协议正式目标与 M3.1 三候选离线约束全部保留。

## Fog

- 实际凭据资格、三协议真实推理续接形状和价格来源，按 R25–R27 收敛。

- 01 resolved：共享 MiniMax 中国身份、默认采样标签及六候选装配完成；[Answer](issues/01-provider-model-contract.md#answer)、[证据](evidence/ticket-01/README.md)、ADR-0037。三协议为下一 frontier，所有候选保持非正式。
