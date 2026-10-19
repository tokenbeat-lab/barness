# MiniMax 实施任务图

Status: resolved
Integration branch: `codex/minimax-global`
Historical CN checkpoint: `codex/minimax-cn` at `4295ee8`
Base: `ecfd4b4`

[当前交付：COMPLETE](delivery-status.md)。全部工单resolved；当前来源11/11、新完整门禁、矩阵发布、证据复核和自建工作树清理均完成。

## Notes

2026-10-18 用户调用 implement-spec；原目录只有规格与研究，本次按本地 tracker 补齐任务图。测试接缝由 spec T01 明确确认，不再重复确认。采用独立 worktree 和实现/合并 subagent。

## Frontier

无开放工单。21→12/15→10→09→spec已按各票[Answer](issues/21-visual-source-renewal.md#answer)关闭；原有限合同的历史失败和额度仍独立保留。最终依据见[当前交付](delivery-status.md)与[最终闭合](evidence/ticket-21/final-closure/README.md)。

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

- 17 resolved（离线源码、文档与证据留存范围）：[Answer](issues/17-final-payload-review-gaps.md#answer)和[最终审查](evidence/ticket-09/final-closure/README.md)闭合最终payload/文档及留存工具问题；完整验收、真实失败和严格当前来源缺口不随本票关闭。

## Historical Fog

- 国际站38b7098三路研究已证明认证与实际M3；Messages三场景通过，Chat等值双推理镜像误拒、Responses本地native编码失败已由10作离线收敛，完整普通/race各1340 PASS、双轴0发现；后续目录预验循环及业务配置边界已修复，来源7605332三路完整实际候选各13/12通过并完成数据目录激活；正式三报告、GLM/旧证据续期和最终门禁继续由10/09完成。原始148文件归档/2个失败状态保留；完整候选/正式仍须R25–R27自身验收。M3.1没有M Plan资格与价格确认，线上三候选保持NOT_RUN，离线仍必交。
- 官方国际站价格512k未定义精确整数；当前512000为明示保守估算策略。正独立缓存写缺价继续失败。三协议结构比较相同不授予线上支持。
- 11完成后按新已提交来源prepare、无Key检查和精确审查，再执行国际站研究。旧十四身份与四效果须在最终目录/源码下实际续期；完整ordinary/race与最终双轴审查仍必交。历史12304 PASS/52 FAIL/exit-15与已通过CN局部验证保持原状态，不能转作国际站完整门禁。

- 8900034国际站M3三份自身正式报告均PASS（39calls/36实际Attempts），完整候选49/45及历史消费、原始失败均封存。GLM初轮Chat/Messages PASS、Responses FAIL，66/66如实保留；[正式证据](evidence/ticket-10/international-formal-glm/README.md)。票10/09未关闭，随后用户批准唯一72/96补正，实际结果见票12。

- [12 GLM捕获修正与唯一补正轮次](issues/12-glm-corrective-1.md)：离线实现、普通/race、双vet、两轴审查及M3源保持均完成；[证据](evidence/ticket-12/README.md)。额外72/96已明确批准且唯一R1完成66/66；连原66/66共132/132。Chat/Messages各22/22 PASS，Responses21 PASS/1 FAIL，完整上游SSE缺正文。当前raw/views续期拒绝，旧四selector不变；无R2。见[实际证据](evidence/ticket-12/actual-corrective-1/README.md)。10/09仍开放。其余15项原授权验证已完成168/168（两项DeepSeek PASS、13项FAIL）；3d3a1fb的完整普通/race各12385 PASS/6 FAIL、审计0，CLI随后读取bundle失败且没有发布报告。全部实际进程已关闭，旧能力本次续期累计300/300，17身份矩阵与完整交付仍未通过；消费者修复及其适用范围见交付状态。

## 历史验收收敛检查点（2026-10-18，六项审批前）

用户新增区域路由验证已执行固定18个付费进程：GLM/DeepSeek直连，国际服务经127.0.0.1:7890。其余15项collection全部PASS，GLM Chat/Messages PASS、Responses19PASS/3FAIL。新增245/245，连此前300/300累计545/545；全部进程关闭，未用额度不转移，不自动另开一轮。Jev自身支持包已派生，17身份默认矩阵原子导入16PASS/1FAIL。

8fbd170完整公共CLI已完成两种vet、ordinary/race，四项数字退出0/0/1/1，两套各12439PASS/8FAIL/1NOT_RUN、审计0；报告pass=false。原始启动中断与完成后的全量运行分开保留；不采用旧bundle冒充新执行。

[14 离线前置缺口](issues/14-release-preflight-gaps.md)承接完整门禁发现的固定映射Hash、可选诊断默认注册和测试夹具inventory修复。14与[16 GLM验收Host预算对齐](issues/16-glm-host-deadline-alignment.md)的离线源码/证据已在7231fdad合入，均按离线范围resolved；13的有限执行与归档也已resolved。当前frontier为[15 最小来源重验](issues/15-source-scoped-revalidation.md)（合同离线实现，待新增有限付费批准） → 10 → 09。独立政策复核确认条件成立时只需GLM3、protocol114和visual2，最高108calls/132Attempts；M3自身映射不变，原generic10/text2保持各自适用来源，不刷新旧pins或时间。新增付费尚未请求或批准，必须先完成新合同与精确来源审查。

历史Fog条目保留原检查点语境，当前结论以[交付状态](delivery-status.md)和机器状态为准；spec/10/09不关闭。

- 13 resolved（固定区域验证/证据范围）：[Answer](issues/13-routed-validation.md#answer)与[证据](evidence/ticket-13/README.md)保留17PASS/1FAIL、245/245及全量false，不关闭spec。
- 14 resolved（离线范围）：[Answer](issues/14-release-preflight-gaps.md#answer)记录映射/诊断/inventory修复、实际六项无Key检查0及双轴审查；旧同号映射子记录保留历史。
- 16 resolved（离线范围）：[Answer](issues/16-glm-host-deadline-alignment.md#answer)记录GLM专属120秒装配、本地91秒RED/GREEN及M3三份263项不变；没有厂商120秒成功证据。

- 15离线实现检查点：09228工具源码与90ec留存已两轴0发现并合入，71工具测试/双vet通过；[证据](evidence/ticket-15/README.md)。ROOT精确计划及六项实际无Key检查仍待执行，新108/132审批未请求，真实六项/最终门禁尚未执行，15、10、09和spec不关闭。


## 当前验收收敛：固定六项已关闭

2026-10-18 最新固定六项真实验收已按用户新增108calls/132HTTPAttempts批准完成：中国区GLM直连，国际protocol114及两项视觉经127.0.0.1:7890。Chat、Responses及三项效果collection通过，Messages21PASS/1FAIL，新增102/102，连原545/545累计647/647。该固定合同全部关闭，未用6calls/30Attempts不转移，无自动下一轮或新增付费授权。严格GLM续期和新四报告装配拒绝，原selector与17身份矩阵SHA保持；原矩阵16PASS/1FAIL为旧Responses记录，最新Messages原始失败未导入。

来源7931的首次完整无Key公共CLI及post-guards均已结束：八项输入校验0，双vet/ordinary/race数字退出0/0/1/1，CLI1、报告pass=false。普通/race各12451项（12445PASS/6FAIL/0NOT_RUN），审计0、868真实PIDIFF PASS；29个pressure命名用例通过，其中固定规则选择15项。六项失败为GLM三协议正式导入/已报告用量检查；未采用旧bundle或绕过缺失Messages报告。

最终双轴审查发现的Messages最终payload形状及文档问题已由票17离线补正，源码2306917、文档295728a均复审零发现。44公共用例为42非法输入零Provider/Attempt拒绝、两合法业务JSON原值发送；定向普通/race各1663PASS/8NOT_RUN，同8项mandatory显式开关另各8PASS，双vet0。实际测量为ca67+WIP，1027编译/JSON/mod/sum文件等于2306，不声称提交后重跑。修复后M3/GLM/文本/protocol/视觉全部11组严格来源当前相等性不成立；历史证据保持，generic10按既有受影响规则另行核对不变。旧门禁不认证当前源码。

[17 最终payload及文档补正](issues/17-final-payload-review-gaps.md)只按离线修复与审查留存范围闭合；[15](issues/15-source-scoped-revalidation.md)及[12](issues/12-glm-corrective-1.md)保留ready-for-human验收缺口，10、09和spec保持claimed/PARTIAL。无当前完整发布认证。原始报告、严格适用性与双轴复审见[最终审查入口](evidence/ticket-09/final-closure/README.md)。历史Fog与审批前条目按原检查点保留，不代表当前待审批状态。

## 2026-10-19 单场景复测检查点

[18 resolved](issues/18-glm-messages-cancel-rerun.md)：用户单独授权的Messages `glm-5.3/cancel` 真实1/1复测PASS，861ms，首个非空ThinkingDelta后取消，canceled/stream、aborted；[证据](evidence/ticket-18/README.md)。原120001ms失败与固定六项647/647保留，累计648/648。该新授权已关闭；全路线formal22/严格来源/装配/完整门禁仍开放，10、09和spec不关闭。

- 2026-10-19：票19十一组真实复验及当前来源公开验证全部通过，默认17身份矩阵全部PASS；[实际证据](evidence/ticket-19/actual/remaining/README.md)。旧0/11诊断、历史失败与原阶段消费均保留。完整门禁及最终审查另行判定。

- 19 resolved：[Answer](issues/19-current-source-eleven-renewal.md#answer)关闭固定十一组复验与结果留存。11/11、189/186及17身份PASS已记录；完整门禁实际FAIL由后续视觉取消证据修复承接。五份有界归档、267820份门禁原件及每套99113引用经独立核对0，来源/矩阵未变；原预算关闭，无新付费授权。

- 2026-10-19：20源码d02精确合入，两轴复审各0；23普通/race与双vet通过，真实WIP源码与提交相等。当前资格8/11，21只准备protocol114/visual-jev/visual-decisions新36/36；未授权、未读Key、未建新付费intent，原预算与门禁失败保持。

- 20 resolved：[Answer](issues/20-visual-unsent-cancellation.md#answer)关闭视觉取消证据离线修复及完整留存；原始RED/失败GREEN与45数字命令保留，203335成员两轴流式核验0，23确定性普通/race和双vet通过。21承接当前三组补证及全量门禁，预算仍未授权。


## 最终闭合上下文

以下旧Fog和检查点描述其当时状态；当前完整交付以本节和顶部COMPLETE为准。来源78843的新完整发布门禁四项实际命令0/0/0/0、CLI0、报告true；普通/race主包各12495PASS、0FAIL、0NOT_RUN、audit0，17份自身支持与4项业务效果PASS。严格当前来源11/11；五份归档、公开读回及固定证据两轴复核均通过，精确e2cf证据已合入8a1163b。默认591323候选实际发布，只更新Jev、保留其他16行对象和时间。自建工作树已移除、分支和中央原件保留，主仓三个安装入口均exit0，1078来源及七项原门禁hash保持。

- [21 resolved](issues/21-visual-source-renewal.md#answer)：直接人类批准的36/36已闭合：protocol114、visual-jev、visual-decisions各12/12 PASS，经127.0.0.1:7890、零重试；M3/GLM/文本八组完整原生来源不变，未重跑。
- [12 resolved](issues/12-glm-corrective-1.md#answer)：本票原R1的66/66及缺正文FAIL保持原始事实，没有自动R2。
- [15 resolved](issues/15-source-scoped-revalidation.md#answer)：原固定六项102/102、Messages120001ms失败及未用6/30保留，不转移旧授权。
- [10 resolved](issues/10-formal-catalog-support.md#answer)：M3国际站三协议均有自身真实候选、正式报告与目录/价格证据；当前三正式PASS39calls/36Attempts，原候选49/45和原正式39/36各自历史不变。
- [09 resolved](issues/09-integrated-acceptance-review.md#answer)：28个R、20个A、80个US均有非空映射，394行追溯与1348个唯一MiniMax必交case保持；全部868个真实PIDIFF、29个命名pressure（固定规则15个）、8个rollback通过。
