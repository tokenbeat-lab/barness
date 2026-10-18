# 14 — 完整门禁暴露的离线前置缺口

Type: task
Status: resolved
Updated: 2026-10-18
Blocked by: None

## Context

票13的18个实际进程全部关闭，新增245 calls/245 Attempts。来源8fbd170的完整公共门禁已完成，四项数字退出码为0/0/1/1，报告`pass=false`；原始ordinary/race各12439 PASS、8 FAIL、1 NOT_RUN，审计0。所有原结果保留，不将后续修复覆盖原失败。

## Failure inventory before changes

- 已审查的7处metadata场景更名没有同步固定`approvedMappingHash`，公开完整发布合同与反删减控制失败。[原诊断](../evidence/ticket-13/release-plan-gap.json)。工作树7111641/c2604db已修正固定Hash，并先失败后将既有公开合同检查加入付费前置；原五项实际通过和独立审查单独保留。
- 可选历史bundle诊断在检查启用环境变量前注册OD57；默认跳过留下NOT_RUN，被完整门禁正确拒绝。该可选诊断没有mandatory trace引用。仅调整默认skip的注册顺序，显式启用时全部历史断言保持。
- Decisions既有7模式测试夹具未给3个真实JSON文件登记SHA inventory，valid控制在语义校验前失败；相邻wrong-oracle负例也可能因为错误原因通过。修复夹具生产者并要求语义检查前inventory有效，不能放宽严格消费者。

[后两项只读诊断](../evidence/ticket-13/postgate-gap-2.json)同时记录其他非GLM追溯失败均为GLM Responses或race-offline的汇总后果，没有新增旧厂商故障。

## Acceptance

使用既有公共E2E/辅助CLI与必要隔离边界，先记录失败，再做最小修复。完整release包检查和可选诊断的默认/显式行为必须覆盖；前置检查不能在已知离线失败下允许建立付费intent。保留mandatory trace、真实差分、SDK/source/inventory/捕获断言和全部旧selector。

所有修复及前置检查在独立工作树提交后做Standards/Spec两轴复审，完整保留RED、GREEN、环境失败与实际检查退出码；原c260审查不冒充后续源码审查。文档变化与执行来源分别核对。按既有来源规则确定后续最小真实重验集合，禁止刷新旧pins或把完整外层launcher冻结集合自动扩展为所有厂商的执行集合。

本票只授权离线修复与核验，不批准新增付费轮次。全部旧消费545/545保留，未用额度不转移；M3三份正式263文件映射若仍适用不重跑，M3.1没有M Plan权益仍NOT_RUN0/0。

## Comments

2026-10-18：完整门禁09:34:10Z已结束，解除8fbd170执行期间的HEAD冻结；实际报告与原始普通/race封存。票14仍在实现工作树修复，尚未合入集成分支，不授予新完整发布认证。

## Answer

离线修复及完整留存已于集成提交 `7231fdad77bd1fb69bb1f21b934e55db77aab814` 合入。映射固定值、可选诊断默认注册顺序、两个原生夹具 inventory 和六项付费前置检查完成；严格消费者、必需 trace 与旧 selector 保持。源码 `1a0ce345`、合并检查点 `572570bb` 的六项实际无 Key 检查全部退出0，普通/race的语义与进程反例、工具检查、双vet通过；早期缺ZIP与race外层限时失败单独保留。

[映射阶段证据](../evidence/ticket-14/repair/README.md)、[后续缺口与六项检查](../evidence/ticket-14/gap2/README.md)、[独立两轴审查](../evidence/ticket-13/reviews/README.md)完整保留失败先行、源码与归档边界。本票按离线范围关闭，不声称修复后的全仓发布门禁通过；真实重验和最终新全量由[15](15-source-scoped-revalidation.md)承接，无新增付费批准。
