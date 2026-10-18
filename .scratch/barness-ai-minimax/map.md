# MiniMax 实施任务图

Status: claimed
Integration branch: `codex/minimax-global`
Historical CN checkpoint: `codex/minimax-cn` at `4295ee8`
Base: `ecfd4b4`

[当前交付：PARTIAL](delivery-status.md)。10、09和spec保持未关闭；自身离线验证通过不替代实际正式报告与完整门禁。

## Notes

2026-10-18 用户调用 implement-spec；原目录只有规格与研究，本次按本地 tracker 补齐任务图。测试接缝由 spec T01 明确确认，不再重复确认。采用独立 worktree 和实现/合并 subagent。

## Frontier

历史任务图01 → {02,03,04} → {05,06,07}、06→08已完成；当前frontier为10 → 09（11已完成国际站实现/离线范围），10同时保留05/06/07依赖。07先建受控证据工具，10等执行源码冻结后才运行真实调用；09汇合最终判定。

- [01 Provider 身份、六个候选型号与思考元数据](issues/01-provider-model-contract.md)；Blocked by: None。
- [02 MiniMax Chat 文本、推理、工具历史与边界](issues/02-chat-protocol.md)；Blocked by: 01。
- [03 MiniMax Responses 无状态流、工具与可信续接](issues/03-responses-protocol.md)；Blocked by: 01。
- [04 MiniMax Messages adaptive 思考、工具和签名续接](issues/04-messages-protocol.md)；Blocked by: 01。
- [05 三协议隔离、失败重试、回调和资源释放](issues/05-isolation-failure-resources.md)；Blocked by: 02, 03, 04。
- [06 固定 MiniMax pi 差分和扩展证据](issues/06-pi-differential.md)；Blocked by: 02, 03, 04。
- [07 独立 Live Harness、预算与可信证据工具](issues/07-live-catalog-release.md)；Blocked by: 02, 03, 04。
- [08 Host 文档、ADR 和回滚实现](issues/08-host-docs-final-acceptance.md)；Blocked by: 06。
- [09 完整验收、双轴审查与交付判定](issues/09-integrated-acceptance-review.md)；Blocked by: 08, 10。

- [10 M3自身真实证据、价格目录与增量发布合同](issues/10-formal-catalog-support.md)；Blocked by: 05, 06, 07, 11。

- [11 国际站身份、美元价格与地区证据切换](issues/11-international-target.md)；Blocked by: 01–08（均已resolved）。

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

- 2026-10-18用户明确改用国际站：01–08的原resolved范围作为CN历史保留；11承接当前目标转换和国际站离线复验，10/09原正式要求不减。使用ProviderMiniMax/minimax、.io端点、直接USD快照；按ADR-0042分开历史预算槽与当前精确证明身份。共享候选消费2/120、2/144，Chat槽2/20、2/24，不重置历史intent或账本。

- 11 resolved（国际站实现/离线范围）：ProviderMiniMax/minimax、.io三端点、直接USD快照、冻结MINIMAX_MODELS、精确目标证明与共享预算槽完成；[Answer](issues/11-international-target.md#answer)、[当前国际站证据](evidence/ticket-11/README.md)。源码e9731a4普通/race各1314 PASS、live-tag race66 PASS、两种vet与无Key检查通过，双轴源码审查0发现；原CN2/2与全部历史完整保留。10继续真实国际站候选/正式报告与旧证据续期，09继续完整门禁及最终审查。

## Fog

- 国际站38b7098三路研究已证明认证与实际M3；Messages三场景通过，Chat等值双推理镜像误拒、Responses本地native编码失败已由10作离线收敛，完整普通/race各1340 PASS、双轴0发现；后续目录预验循环及业务配置边界已修复，来源7605332三路完整实际候选各13/12通过并完成数据目录激活；正式三报告、GLM/旧证据续期和最终门禁继续由10/09完成。原始148文件归档/2个失败状态保留；完整候选/正式仍须R25–R27自身验收。M3.1没有M Plan资格与价格确认，线上三候选保持NOT_RUN，离线仍必交。
- 官方国际站价格512k未定义精确整数；当前512000为明示保守估算策略。正独立缓存写缺价继续失败。三协议结构比较相同不授予线上支持。
- 11完成后按新已提交来源prepare、无Key检查和精确审查，再执行国际站研究。旧十四身份与四效果须在最终目录/源码下实际续期；完整ordinary/race与最终双轴审查仍必交。历史12304 PASS/52 FAIL/exit-15与已通过CN局部验证保持原状态，不能转作国际站完整门禁。

- 8900034国际站M3三份自身正式报告均PASS（39calls/36实际Attempts），完整候选49/45及历史消费、原始失败均封存。GLM初轮Chat/Messages PASS、Responses FAIL，66/66如实保留；[正式证据](evidence/ticket-10/international-formal-glm/README.md)。票10/09未关闭，随后用户批准唯一72/96补正，实际结果见票12。

- [12 GLM捕获修正与唯一补正轮次](issues/12-glm-corrective-1.md)：离线实现、普通/race、双vet、两轴审查及M3源保持均完成；[证据](evidence/ticket-12/README.md)。额外72/96已明确批准且唯一R1完成66/66；连原66/66共132/132。Chat/Messages各22/22 PASS，Responses21 PASS/1 FAIL，完整上游SSE缺正文。当前raw/views续期拒绝，旧四selector不变；无R2。见[实际证据](evidence/ticket-12/actual-corrective-1/README.md)。10/09仍开放。其余15项原授权验证已完成168/168（两项DeepSeek PASS、13项FAIL）；3d3a1fb的完整普通/race各12385 PASS/6 FAIL、审计0，CLI随后读取bundle失败且没有发布报告。全部实际进程已关闭，旧能力本次续期累计300/300，17身份矩阵与完整交付仍未通过；消费者修复及其适用范围见交付状态。

## 当前验收收敛（2026-10-18）

其余15项原授权验证已完成168/168（两项DeepSeek PASS、13项FAIL）；3d3a1fb的完整普通/race各12385 PASS/6 FAIL、审计0，CLI随后读取bundle失败且没有发布报告。全部实际进程已关闭，旧能力本次续期累计300/300，17身份矩阵与完整交付仍未通过；消费者修复及其适用范围见交付状态。

M3 三份自身正式的263文件和 GLM R1三份217文件执行映射保持适用；后续 release 源码修改使 L/Jev 广域来源仅历史适用。最终审查、原始失败、后续定向与诊断结果分别保留，不将 adopted bundle 当作新全量执行认证。详见[交付状态](delivery-status.md)。
