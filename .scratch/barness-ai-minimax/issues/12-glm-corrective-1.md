# 12 — GLM 捕获修正与唯一 prospective corrective-1

Status: ready-for-human
Type: task
Updated: 2026-10-18
Blocked by: final current-source release certification and evidence/document review (ticket19 eleven groups now PASS)

## Context

国际站三条 M3 formal 已有自身通过证据；冻结来源8900034下 GLM Chat、Messages 各22/22 PASS，Responses 为21 PASS/1 FAIL。失败既有512KiB wire截断和17.7MB事件重复快照，又有真实4096输出耗尽、只有thinking而无最终text。原始三路66 calls/66 Attempts、失败捕获、intent及报告不修改。此票不重开原进程、不拼接场景、不变更正式导入或来源规则。[ADR0045](../../../docs/adr/0045-glm-corrective-capture-round.md)。

## Approved one-round contract

只新增一次固定 `glm-corrective-1-{chat,responses,messages}` 三路完整22场景。沿用既定prompts/plan/source manifest、两型号、4096输出、120s单调用/1800s进程、默认零重试，各24 calls/32 Attempts，无自动R2。原66/66不刷新；额外最多72/96，本次续期GLM累计上限138/162；其他旧能力302/302保持，旧能力总体440/464。原unused6/30不转移。人类明确同意前禁止读取凭据或调用；实现只生成PENDING模板，精确code/binary/plan/checks及原收据pins需同时获得人类授权和独立源码审查。

## Acceptance

公开固定formal执行入口使用不能联网的synthetic peer先失败再修正：大于512KiB且未超2MiB的合法SSE须完整保留到尾部用量；4094碎片delta不再重复整个live Partial；所有type/attribution/order/immutable字段、一次final Partial、native/terminal/usage、工具和取消均保留。length/no-text及超限仍FAIL，不能发布支持。新捕获及原历史捕获分别以自身原目录进行严格Result/Observer/wire/cancel回放，replay永不认证当前。

辅助CLI先失败再实现：缺审批/错scope/source/binary/plan/checks、原消费缺失/未知/变动、并发/跨目录/工作树重启、第二round、错canonical pins/混旧新来源均拒绝。正路径必须携带原66/66、固定额外预算、独占intent与原始成功pins。归档只能显式 `--round corrective-1`，仍经当前未放宽Go importer，再生成无付费用量视图。共享/M3-selected Go及三份263文件实际source maps须逐字不变。

## Answer

离线实现完成，生产捕获/审批修正为 `4ff6dca`、`5abc1a5`，受控 fault 单次注入为 `6ff05b8`。仅 GLM 自身 Go 来源及 Python 计划/归档边界改变；共享/M3-selected Go、正式导入、来源选择器、目录和已有真实证据/intent 均未修改。原始三份 M3 formal 公开 verifier 全部 PASS，三个263文件map及报告字节不变。

先保留公开 red 再实现 green：完整大SSE为2,047,892字节，转义exchange为2,114,598字节，4094个delta的事件文件约2.2MiB；当前与历史wire/Result/Observer/cancel回放保持。length/no-text、wire超限及event超限都保持FAIL，后者仍保留已知input20与UsageReporting。后续L计划显式携带原66/66及R1已知/未知消费；本次续期上限440/464、Zhipu138/162，其余302不变，旧GLM任务历史独立保留。

4ff完整GLM普通/race各61项PASS、审计0；最终6ff针对全部新捕获与事件预算的普通/race各4项PASS、审计0，工具CLI35项PASS，普通/live vet通过。原4ff两条Spec发现、5abc修正与6ff fixture独立双轴报告均保留，最终两轴0 open。5abc中间race为4项断言PASS但实现者中断慢速final audit，exit-15/无审计，不能称为Go PASS或timeout；纠正说明原样保留。详见[证据](../evidence/ticket-12/README.md)。

以上为离线实现检查点。用户随后明确批准唯一新增72/96；精确来源5bf2c16、二进制/计划/检查及APPROVED收据均在[实际补正胶囊](../evidence/ticket-12/actual-corrective-1/README.md)保留。实际额外66 calls/66 Attempts，连原66/66为132/132。Chat与Messages各22/22 PASS；Responses为21 PASS/1 FAIL，glm-5.2/native-history的完整HTTP200/SSE只有推理而无正文，适配器未丢失内容。三路均完成且审计通过；没有R2。严格raw/views续期拒绝失败、原四selector不变；公共矩阵可记录捕获和场景清单齐全、但complete=false的真实FAIL。此外部验收阻塞未解决，票12保持ready-for-human，10/09/spec保持未关闭；其余15项原授权验证已完成168/168（两项DeepSeek PASS、13项FAIL）；3d3a1fb的完整普通/race各12385 PASS/6 FAIL、审计0，CLI随后读取bundle失败且没有发布报告。全部实际进程已关闭，旧能力本次续期累计300/300，17身份矩阵与完整交付仍未通过；消费者修复及其适用范围见交付状态。

Accounting scope: all `lifetimeGLM*` fields mean this current renewal original66/66 plus corrective-1. Prior GLM-task paid history (including its separate282-call contract) remains immutable and separately accounted; this extension does not erase or reset it.

2026-10-18补充：便携证据校验器已完成先失败再修复的有界解压验证，原程序和原索引字节另行保留，当前校验器版本独立登记；见[retention-bounds](../evidence/ticket-12/retention-bounds/README.md)。这是离线保留修复，不改变原轮/R1实际结果、消费或审批范围。


2026-10-18票13新增路由合同的实际结果另列：三路65calls/65Attempts，Chat/Messages各22场景PASS，Responses19PASS/3FAIL；两个native-history均PASS，但glm-5.3/tool-call与control-max约90秒超时，tool-result依赖失败0/0。此结果不改写本票原R1缺正文捕获或66/66消费，也不是本票自动R2。严格续期仍拒绝，后续额外付费尚未批准。当前GLM阻断以[票13实际证据](../evidence/ticket-13/actual-glm/README.md)和[交付状态](../delivery-status.md)为准。


### 2026-10-18 固定六项关闭后的最新交付检查点

2026-10-18 最新固定六项真实验收已按人类新增108calls/132HTTPAttempts批准完成：中国区GLM直连，国际protocol114及两项视觉经127.0.0.1:7890。Chat、Responses及三项效果collection通过，Messages21PASS/1FAIL，新增102/102，连原545/545累计647/647。该固定合同全部关闭，未用6calls/30Attempts不转移，无自动下一轮或新增付费授权。严格GLM续期和新四报告装配拒绝，原selector与17身份矩阵SHA保持；原矩阵16PASS/1FAIL为旧Responses记录，最新Messages原始失败未导入。

最新Messages cancel在可观察内容前到达120秒期限，duration120001ms、usage未报告；不证明成功取消或厂商原因。原R1缺正文、区域轮Responses失败与本轮分开保留。本票保持ready-for-human。见[当前交付](../delivery-status.md)、[六项原件](../evidence/ticket-15/new-actual/glm/README.md)、[全量门禁](../evidence/ticket-15/new-actual/final-gate/README.md)及[最终审查](../evidence/ticket-09/final-closure/README.md)。

2026-10-19 当前补充结论：本票原有限轮次的失败、预算与旧拒绝保持历史。票19按新的直接批准执行十一组当前来源复验，GLM三路完整各22/22 PASS（Messages cancel1146ms），严格raw/known-usage/views发布及17份支持输入装配均exit0，默认矩阵17PASS/0FAIL。此前Messages/来源资格阻断已由新独立证据补足；完整发布门禁与最终审查另待完成。见[票19原件](../evidence/ticket-19/actual/remaining/README.md)。
