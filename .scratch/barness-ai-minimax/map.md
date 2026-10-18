# MiniMax 实施任务图

Status: claimed
Integration branch: `codex/minimax-cn`
Base: `ecfd4b4`

## Notes

2026-10-18 用户调用 implement-spec；原目录只有规格与研究，本次按本地 tracker 补齐任务图。测试接缝由 spec T01 明确确认，不再重复确认。采用独立 worktree 和实现/合并 subagent。

## Frontier

01 → {02,03,04}；三协议完成后{05,06,07}独立推进；06→08，{05,06,07}→10，{08,10}→09。07先建受控证据工具，10等执行源码冻结后才运行真实调用；09汇合最终判定。

- [01 Provider 身份、六个候选型号与思考元数据](issues/01-provider-model-contract.md)；Blocked by: None。
- [02 MiniMax Chat 文本、推理、工具历史与边界](issues/02-chat-protocol.md)；Blocked by: 01。
- [03 MiniMax Responses 无状态流、工具与可信续接](issues/03-responses-protocol.md)；Blocked by: 01。
- [04 MiniMax Messages adaptive 思考、工具和签名续接](issues/04-messages-protocol.md)；Blocked by: 01。
- [05 三协议隔离、失败重试、回调和资源释放](issues/05-isolation-failure-resources.md)；Blocked by: 02, 03, 04。
- [06 固定 MiniMax pi 差分和扩展证据](issues/06-pi-differential.md)；Blocked by: 02, 03, 04。
- [07 独立 Live Harness、预算与可信证据工具](issues/07-live-catalog-release.md)；Blocked by: 02, 03, 04。
- [08 Host 文档、ADR 和回滚实现](issues/08-host-docs-final-acceptance.md)；Blocked by: 06。
- [09 完整验收、双轴审查与交付判定](issues/09-integrated-acceptance-review.md)；Blocked by: 08, 10。

- [10 M3自身真实证据、价格目录与增量发布合同](issues/10-formal-catalog-support.md)；Blocked by: 05, 06, 07。

## Decisions so far

- 08的Host示例/回滚依赖已有协议与06差分结论，不依赖05资源增强；移除不必要的05阻塞。最终09仍经10依赖05全部验收。

- 07受控harness无需等待05/06；已将真实候选/正式目录/发布合同独立为10，先冻结执行实现再付费捕获，R26全部预算不变。

- 08文档/回滚不需要等待07真实报告才开始；已将最终验收独立为09，避免无关依赖，全部交付要求保持。

- 实施不得把缺真实凭据改写为正式通过；M3 三协议正式目标与 M3.1 三候选离线约束全部保留。

- 01 resolved：共享 MiniMax 中国身份、默认采样标签及六候选装配完成；[Answer](issues/01-provider-model-contract.md#answer)、[证据](evidence/ticket-01/README.md)、ADR-0037。三协议为下一 frontier，所有候选保持非正式。
- 02 resolved：MiniMax Chat 独立能力、最终守卫、完整工具/推理历史与流末用量完成；[Answer](issues/02-chat-protocol.md#answer)、[证据](evidence/ticket-02/README.md)。579 自身 E2E/race 与 883 旧行为回归通过；真实 M3 分离格式/原生 auto 仍归 10，未宣称已证实。
- 03 resolved (offline scope): independent Responses capability/guards, authoritative stream handling and opaque native/tool history; [Answer](issues/03-responses-protocol.md#answer), [evidence](evidence/ticket-03/README.md), ADR-0039. Lossless reasoning input projection remains provisional until own live10 verifies it; no formal support claim.
- 04 resolved：Messages adaptive/effort、二元元数据、严格 SSE 与可信签名工具续接完成；[Answer](issues/04-messages-protocol.md#answer)、[138 own/race + 297 回归证据](evidence/ticket-04/README.md)、ADR-0040。真实形状与正式支持仍由后续票独立验收。

- 06 resolved：冻结 M3 Messages 六条文本/签名双工具及完整续接实际差分、完整目录与五个扩展身份四入口验收完成；[Answer](issues/06-pi-differential.md#answer)、[29 own/race + 70 旧原生差分证据](evidence/ticket-06/README.md)。41 条有限 case/path/kind、2 条扩展路线；旧317条决策/6路线与全部 pins 不变，pending0，不声明 live 或正式价格。

- 05 resolved：MiniMax 三协议共享错误/回调/重试/冻结/隔离与全部资源边界完成；[Answer](issues/05-isolation-failure-resources.md#answer)、[证据](evidence/ticket-05/README.md)。202 自身 E2E/race、3311 受影响回归、258 已提交来源复验通过；原4/8负载保持，新增3/6批次全部≤75%。真实兼容 quota/trace 继续由10候选证据收敛。

- 07 resolved (controlled tools only)：六身份独立harness、持久阶段预算、研究前驱、原始捕获/密封/重放与正式导入防线完成；[Answer](issues/07-live-catalog-release.md#answer)、[证据](evidence/ticket-07/README.md)。45受控case普通/race通过，无厂商收费请求；实际M3正式目录和自身报告仍由10提供。

- 08 resolved (Host/rollback scope)：完整本地/云端三 Binding/用途示例、双语指南/ADR、配置撤销和7039文件完整软件回滚完成；[Answer](issues/08-host-docs-final-acceptance.md#answer)、[证据](evidence/ticket-08/README.md)。79自身+88旧 Host cases 普通/race、两种软件回滚普通/race与定向 vet PASS；原 GLM receipt 路径泄漏根因已修复且审计保留。实际支持仍待10，09汇合最终状态与全集验收。

- 09文档整理检查点：[有限追溯与历史保存](evidence/ticket-09/README.md)保留28个R、20个A、80个US到1304个既有有限case的精确映射，以及原始1307-case离线观察的来源/状态。03的33次受控red/green捕获及29012文件已原样归档并验证。此整理不代表最终验收；09仍受10阻塞。

## Fog

- 三协议真实推理续接形状与正式支持，按 R25–R27、票10独立验收。用户已确认 MINIMAX_KEY 中国区标准按量；实际M3 Chat研究的两次调用、两次Attempt均返回HTTP401认证失败，研究为FAIL/incomplete，依赖工具续接为NOT_RUN。其余M3实际验收等待该已授权凭据修复，不能用受控或重放报告替代。M3.1未获M Plan资格，三个线上候选保持NOT_RUN。
- 官方价格、汇率与schema已复核；正式价格目录与证据仍待票10生成。09的最终普通/race/压力/差分、旧十四身份和四业务效果更新、门禁及最终双轴审查尚未完成。
