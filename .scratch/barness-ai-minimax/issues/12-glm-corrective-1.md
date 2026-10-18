# 12 — GLM 捕获修正与唯一 prospective corrective-1

Status: ready-for-human
Type: task
Updated: 2026-10-18
Blocked by: explicit human approval of additional paid budget

## Context

国际站三条 M3 formal 已有自身通过证据；冻结来源8900034下 GLM Chat、Messages 各22/22 PASS，Responses 为21 PASS/1 FAIL。失败既有512KiB wire截断和17.7MB事件重复快照，又有真实4096输出耗尽、只有thinking而无最终text。原始三路66 calls/66 Attempts、失败捕获、intent及报告不修改。此票不重开原进程、不拼接场景、不变更正式导入或来源规则。[ADR0045](../../../docs/adr/0045-glm-corrective-capture-round.md)。

## Prospective contract — PENDING HUMAN APPROVAL

只新增一次固定 `glm-corrective-1-{chat,responses,messages}` 三路完整22场景。沿用既定prompts/plan/source manifest、两型号、4096输出、120s单调用/1800s进程、默认零重试，各24 calls/32 Attempts，无自动R2。原66/66不刷新；额外最多72/96，本次续期GLM累计上限138/162；其他旧能力302/302保持，旧能力总体440/464。原unused6/30不转移。人类明确同意前禁止读取凭据或调用；实现只生成PENDING模板，精确code/binary/plan/checks及原收据pins需同时获得人类授权和独立源码审查。

## Acceptance

公开固定formal执行入口使用不能联网的synthetic peer先失败再修正：大于512KiB且未超2MiB的合法SSE须完整保留到尾部用量；4094碎片delta不再重复整个live Partial；所有type/attribution/order/immutable字段、一次final Partial、native/terminal/usage、工具和取消均保留。length/no-text及超限仍FAIL，不能发布支持。新捕获及原历史捕获分别以自身原目录进行严格Result/Observer/wire/cancel回放，replay永不认证当前。

辅助CLI先失败再实现：缺审批/错scope/source/binary/plan/checks、原消费缺失/未知/变动、并发/跨目录/工作树重启、第二round、错canonical pins/混旧新来源均拒绝。正路径必须携带原66/66、固定额外预算、独占intent与原始成功pins。归档只能显式 `--round corrective-1`，仍经当前未放宽Go importer，再生成无付费用量视图。共享/M3-selected Go及三份263文件实际source maps须逐字不变。

## Answer

离线实现完成，生产捕获/审批修正为 `4ff6dca`、`5abc1a5`，受控 fault 单次注入为 `6ff05b8`。仅 GLM 自身 Go 来源及 Python 计划/归档边界改变；共享/M3-selected Go、正式导入、来源选择器、目录和已有真实证据/intent 均未修改。原始三份 M3 formal 公开 verifier 全部 PASS，三个263文件map及报告字节不变。

先保留公开 red 再实现 green：完整大SSE为2,047,892字节，转义exchange为2,114,598字节，4094个delta的事件文件约2.2MiB；当前与历史wire/Result/Observer/cancel回放保持。length/no-text、wire超限及event超限都保持FAIL，后者仍保留已知input20与UsageReporting。后续L计划显式携带原66/66及R1已知/未知消费；本次续期上限440/464、Zhipu138/162，其余302不变，旧GLM任务历史独立保留。

4ff完整GLM普通/race各61项PASS、审计0；最终6ff针对全部新捕获与事件预算的普通/race各4项PASS、审计0，工具CLI35项PASS，普通/live vet通过。原4ff两条Spec发现、5abc修正与6ff fixture独立双轴报告均保留，最终两轴0 open。5abc中间race为4项断言PASS但实现者中断慢速final audit，exit-15/无审计，不能称为Go PASS或timeout；纠正说明原样保留。详见[证据](../evidence/ticket-12/README.md)。

付费预算扩展仍为 **PENDING HUMAN APPROVAL**。本票没有读取用户.env/真实Key、调用厂商、创建APPROVED或新的canonical R1 intent，也未改写任何真实历史。原Responses失败、GLM当前来源续期及全规格未完成状态保持；父任务将在集成后生成精确PENDING源码/二进制/计划/检查pins供人类审批。

Accounting scope: all `lifetimeGLM*` fields mean this current renewal original66/66 plus corrective-1. Prior GLM-task paid history (including its separate282-call contract) remains immutable and separately accounted; this extension does not erase or reset it.
