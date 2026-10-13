# OpenAI Decisions 现有工单 review 与替代草案

**Status:** needs-triage
**Publication:** proposed，尚未替换原 issues
**Updated:** 2026-10-12
**Baseline:** 8ea61b8d70de6c84fe27be34dc7d21c5c40841d9

## 结论

建议重拆：现有 9 张工单覆盖了主要功能，问题主要是横向契约票、过大的证据票、非必要串行依赖，以及门禁条件未落实到具体反例。替代草案有 13 张，每张有完整失败目录、验收标准、真正阻塞项和可重放 artifact；不增加 spec 范围。

本次完成文档和当前代码的静态 review。没有运行生产 E2E、付费 API、发布 gate 或形成任何 runtime/live/效果 PASS。原 9 票及 spec/design/acceptance/implementation-plan/planning-verification 均保持原字节；其原始 SHA-256 已记录，验证器检查未改动。

依据：[spec](../spec.md)、[设计](../design.md)、[OD01–OD58](../acceptance.md)、[旧计划](../implementation-plan.md)、根 GLOSSARY 和 ADR-0003/0004/0006/0010/0016/0017/0018/0020/0021/0024/0026。当前 HEAD 与设计 baseline 相同。分类执行和 route 支持仍只有 TypeSafe，这与 spec 的未实施说明一致。

## Findings

这里的 P1/P2 是工单执行风险分级；不是已发现并运行复现的生产缺陷。

### F01 / P1：P10 必跑差分缺少精确验收条件

[旧 06](../issues/06-runtime-isolation-and-differential.md)承诺 runner 和零 pending，[旧 09](../issues/09-release-gate-docs-and-rollback.md)提到 P10，但没有明确要求 P10 同时成为差分 Required 子集及“删掉全部 P10 diff 仍必须失败”的反例。

当前 [固定合同](../../../ai/internal/testkit/release/plan.go)的 CheckReleasePlan 明确检查六个子集 P01/P02/P03/P04/P06/P07；[判定](../../../ai/internal/testkit/release/gate.go)仅按 supplied Required 检查各子集。若只添加 P0/追溯，旧协议有 PASS 仍可能满足差分门禁，未执行新协议不会必然被发现。

补充：新 07 证明自己的差分；新 13 固定第七子集 P10，同时修改批准选择器/Hash，并用删除 P10、仅剩普通 E2E、未开启差分等反例验证。Decisions 在冻结 Pi 已有 route，不能登记为 Extension Route 来跳过差分。

### F02 / P1：新增图片容量尚未要求门禁导入实际读数

[旧 06](../issues/06-runtime-isolation-and-differential.md)要求压力，[旧 09](../issues/09-release-gate-docs-and-rollback.md)要求发布，但未指定新 Classifier 图片量测如何成为实际必交字段。

当前 [DesignLoadEvidence](../../../ai/internal/testkit/release/artifacts.go)核验旧混合负载和旧 Classifier 问题/状态限额，字段表没有新图片 count/MaxTotalImageBytes 等。仅生成新 JSON 字段或新增追溯不能证明门禁实际消费它。

补充：新 08 同时交负载、量测与 artifact 导入反例；新 13 接线并验证缺新字段/错 route/旧报告无法通过。限额占用率 <=75% 与剩余余量 >=25% 明确区分，保持旧负载和预算。

### F03 / P1：效果扩展需明确历史快照与全拒答完整性

[旧 08](../issues/08-host-examples-and-chinese-evaluation.md)已有拒答分母、null 指标和历史核验意图，缺少“全拒答有效 COMPLETE + accuracy 0 + 无可评分概率指标 null”的具体报告/导入反例，以及历史 bundle 使用哪个目录快照核验。

当前 [Load](../../../ai/examples/chineseeval/plan.go)固定 Jev；[运行与身份检查](../../../ai/examples/chineseeval/report.go)只把普通 ChoiceAnswer 当 answered；[Verify](../../../ai/examples/chineseeval/verify.go)仅 answered 组成 COMPLETE；[门禁导入](../../../ai/internal/testkit/release/artifacts.go)用当前 BuiltinCatalog 调用 Load。正式 Hash 变化后，原 Jev config/report 不能直接通过当前快照核验。

补充：新 11 通过真实公共流程证明拒答统计、完整性、null 指标以及历史 recorded snapshot 核验；新 12 执行独立正式效果；新 13 各 route 各自导入，拒绝同包复用、错身份和 stale 证据。保留原 Jev 命令及报告的外部契约，历史可核验不等于当前可发布。

### F04 / P2：旧 02 是横向公共字段票

[旧 02](../issues/02-public-contracts-and-safe-defaults.md)同时增加类型、目录、policy、Binding，交付时却仍无新的正常 Classify 路线。它虽可通过公共查询验证，但按 to-tickets 的 tracer bullet 规则不应作为单独“字段已落地、行为稍后补齐”的切片。

补充：必要文本类型/能力/Options/拒答并入新 02 单选闭环；图片字段随新 05 图片闭环一起交付。新字段出现即有真实可验证行为及失败关闭边界。

### F05 / P2：旧 04、06、07、08 规模与职责跨度过大

[旧 04](../issues/04-answers-refusals-and-usage.md)包括三题型、拒答、严格答案、presence 和价格；
[旧 06](../issues/06-runtime-isolation-and-differential.md)横跨共享 runtime、混合 race、Node oracle/账本及容量；
[旧 07](../issues/07-live-catalog-and-support-matrix.md)包括来源、候选 live、builtin/schema 迁移、正式 live/矩阵；
[旧 08](../issues/08-host-examples-and-chinese-evaluation.md)将离线 Host 工具和依赖凭据的实际效果绑定。

补充：分别拆为新 03/04、06/07/08、09/10、11/12；每票以公共行为或完整 evidence workflow 验证。无需为重拆先做宽重构或 expand–contract；已有清晰 unary Seam，直接复用。

### F06 / P2：全线串行并非真实阻塞关系

旧 05 依赖旧 04，但图片单选只需最早可用单选；复杂计费也可用单选响应验证。旧 08 依赖正式目录/live，然而 Host CLI/refusal/verifier 可通过自带目录完成离线闭环。运行时证明、Pi 差分和 Host 工具也不存在全线先后必要性。

补充：新 03/04/05 在 02 后独立推进；新 06/07/11 在完整文本/计费/图片后独立推进。候选 live 仍要求完整离线门槛；正式纳入与实际效果保留真实外部依赖，不把“可准备 fixture”当作解除完成门槛。

### F07 / P2：全局 schema 3 迁移需覆盖全部 producer/consumer

[旧 07](../issues/07-live-catalog-and-support-matrix.md)明确迁移矩阵及新报告，但未明确旧九组合 report producer 在全局 schema 升级后的可用性。当前 [schema](../../../ai/internal/testkit/supportmatrix/matrix.go)为 2；[固定身份](../../../ai/internal/testkit/supportmatrix/identity.go)只认识九组合。

补充：新 10 同时验证九组合继续产出/读取新 schema，schema 2 报告拒绝 merge、数据一次性迁移、旧时间/状态保留、批次原子性。新 13 核验正式新 Hash 下全部当前证据，不能手改历史 metadata。

### F08 / P2：tracker 字段与计划验证器需随依赖图更新

现有每票使用 Depends on；显式调用的技能要求 Blocked by。当前[计划核验](../planning-verification.md)硬编码 9 张及每票依赖上一张；这仅能证明原链形式，不能证明必要依赖或新 DAG 正确。

补充：草案使用 Blocked by 数字和标题；新验证器核对前序存在、无环/无传递冗余、元数据与票正文一致、R01–R20/OD01–OD58 无缺口、引用和 500 行约束。批准后同步正式计划/验收 owner/核验脚本，并保留历史 review 记录。

## 提议的 13 张工单

每份完整票在 drafts 中，尚未进入 tracker 的 issues；当前 needs-triage，批准后发布为 ready-for-agent。表中阻塞号均是**新编号**，最后一列是旧票来源。

| 新号 | 完整草案 / 独立交付 | Blocked by | 旧票来源 |
|---|---|---|---|
| 01 | [固定 Jev 公共行为与可重放回归基线](drafts/01-jev-regression-baseline.md) | 无 | 01 |
| 02 | [打通 Decisions 文本单选与最终请求闭环](drafts/02-decisions-text-choice.md) | 01 | 02, 03 |
| 03 | [交付 score/bool 混合答案、分布与逐题拒答](drafts/03-mixed-answers-and-refusals.md) | 02 | 04 |
| 04 | [证明 Decisions 用量完整性与输入阶梯计费](drafts/04-usage-presence-and-input-pricing.md) | 02 | 04 |
| 05 | [打通分类图片、三重授权与完整资源边界](drafts/05-classifier-inline-images.md) | 02 | 02, 05 |
| 06 | [证明故障、快照隔离与授权撤销行为](drafts/06-runtime-isolation-and-rollback.md) | 03, 04, 05 | 06, 09 |
| 07 | [纳入冻结 Pi 的 Decisions 差分与有限差异账本](drafts/07-frozen-pi-differential.md) | 03, 04, 05 | 06 |
| 08 | [形成分类图片设计负载与可核验容量证据](drafts/08-classifier-image-design-load.md) | 06 | 06, 09 |
| 09 | [取得候选 Decisions 路线自己的有界真实证据](drafts/09-candidate-live-evidence.md) | 07, 08 | 07 |
| 10 | [纳入正式目录并形成第十组合独立支持记录](drafts/10-formal-catalog-and-support-matrix.md) | 09 | 07 |
| 11 | [交付 Host 迁移示例与独立拒答评估流程](drafts/11-host-migration-and-evaluation-tooling.md) | 03, 04, 05 | 08 |
| 12 | [取得 Luna 当前正式目录下的独立中文效果报告](drafts/12-independent-chinese-live-evaluation.md) | 10, 11 | 08 |
| 13 | [闭合十组合发布门禁、迁移文档与回滚证据](drafts/13-ten-route-release-gate.md) | 12 | 09 |

01 是前置回归证据；02 交付首个公共单选；03/04/05 交付互不要求的完整扩展；06/07/11 分别证明运行时、差分和 Host 工作流；08/09/10/12/13 逐级完成容量、自身候选、正式支持、实际效果及最终发布。实际资源或凭据不足不会被状态标签隐去。

## 阻塞理由

- 02 需要 01 的原行为和错误顺序，才能证明新 route 没有收窄 Jev。
- 03、04、05 各只需要 02 的真实协议入口和最终请求机制；不互相等待。
- 06、07、11 需要完整题型、usage 和图片路径，分别核验完整 mixed/snapshot、共同差分、迁移流程。
- 08 需要 06 的并发、取消和资源恢复行为，容量结论不能建立在已知生命周期错误上。
- 09 需要 07 的固定差分和 08 的完整离线资源门槛，候选 live 不绕过 design 的阶段 B。
- 10 需要 09 自己的实际候选证据，不能先纳入 builtin。
- 12 需要 10 正式目录/Hash 及 11 可核验的评估工具，真实效果不能拿 candidate 或 fixture 代替。
- 13 的完成需要 12 实际效果及其全部前置证据；门禁反例与文档可以提前准备，正式验收仍受这些证据约束。

## 原票迁移

| 原编号 | 处理 | 新编号 |
|---|---|---|
| 01 | 保留目标，缩到基线职责 | 01 |
| 02 | 随行为交付，文本/图片分别归入闭环 | 02, 05 |
| 03 | 与最少公共契约合成首个单选闭环 | 02 |
| 04 | 分开答案语义与计费 | 03, 04 |
| 05 | 去掉混合题/完整计费的非必要阻塞 | 05 |
| 06 | 分开运行时、oracle 和容量 | 06, 07, 08 |
| 07 | 分开候选与正式纳入/矩阵 | 09, 10 |
| 08 | 分开无需凭据的工具与实际效果 | 11, 12 |
| 09 | 撤销验证和容量导入前移，最终合同保留 | 06, 08, 13 |

批准后在正式 issues 按依赖顺序发布每票，替换过时原票，保留原 Comments 的来源记录；不关闭或修改父 spec。同步 implementation-plan、acceptance 的工单列、planning-verification 的结构脚本和相关执行顺序引用。此时不重写 design 的业务契约、不把 proposed ADR 改为 accepted。

## R01–R20 主责任映射

各草案 Requirements 列包含参与的子范围；下面给出主要交付责任，交叉引用不代表重复实现。

| 需求 | 新票主责任 |
|---|---|
| R01 | 02 |
| R02 | 01, 02, 03, 07 |
| R03 | 02, 05, 11 |
| R04 | 02, 03, 11 |
| R05 | 02, 05, 10 |
| R06 | 05, 08 |
| R07 | 05, 06 |
| R08 | 02, 03, 05, 07 |
| R09 | 02, 03, 05, 06 |
| R10 | 02, 03, 11 |
| R11 | 02, 03, 04 |
| R12 | 04, 09 |
| R13 | 06, 08 |
| R14 | 02, 05, 06 |
| R15 | 07 |
| R16 | 09, 10 |
| R17 | 09, 10, 13 |
| R18 | 11, 12 |
| R19 | 各票独立 artifacts；06 观测、08 容量、13 最终核验 |
| R20 | 06 撤销行为；13 文档/门禁 |

## OD01–OD58 验收责任映射

同一 OD 在多票出现表示不同已明确的变体。例如 02 的单题 refusal / 03 的混合 refusal，11 的 fixture 完整性 / 12 的真实效果 / 13 的发布导入。最终 13 要求所有适用变体实际证据，ID 映射本身不是 PASS。

| 验收 ID | 新票 |
|---|---|
| OD01 | 01, 02 |
| OD02 | 02 |
| OD03 | 01, 02 |
| OD04 | 02 |
| OD05 | 02, 05 |
| OD06 | 02, 05 |
| OD07 | 05 |
| OD08 | 05 |
| OD09 | 01, 02 |
| OD10 | 02 |
| OD11 | 02 |
| OD12 | 02 |
| OD13 | 03 |
| OD14 | 02, 07 |
| OD15 | 05 |
| OD16 | 01, 02 |
| OD17 | 05 |
| OD18 | 05 |
| OD19 | 05 |
| OD20 | 05 |
| OD21 | 05 |
| OD22 | 02, 03, 05 |
| OD23 | 02, 05 |
| OD24 | 05 |
| OD25 | 02, 03, 05 |
| OD26 | 05 |
| OD27 | 02, 03 |
| OD28 | 02, 03 |
| OD29 | 02, 03 |
| OD30 | 02, 03 |
| OD31 | 02, 03 |
| OD32 | 02, 03 |
| OD33 | 03 |
| OD34 | 01, 03 |
| OD35 | 04 |
| OD36 | 04 |
| OD37 | 04 |
| OD38 | 04 |
| OD39 | 03 |
| OD40 | 06 |
| OD41 | 03, 04, 06 |
| OD42 | 05, 06 |
| OD43 | 06 |
| OD44 | 06 |
| OD45 | 06 |
| OD46 | 06 |
| OD47 | 08 |
| OD48 | 07 |
| OD49 | 07 |
| OD50 | 01, 07 |
| OD51 | 09 |
| OD52 | 10 |
| OD53 | 10, 13 |
| OD54 | 09, 10 |
| OD55 | 11, 12 |
| OD56 | 11 |
| OD57 | 08, 13 |
| OD58 | 06, 13 |

## 可重放的文档验证

从 repo root 运行：

~~~sh
python3 .scratch/barness-ai-openai-decisions/ticket-review/verify.py
~~~

输出 validation.json（只表示文档结构核验）及 SHA-256 manifest；不运行生产测试。proposal.json 固定 13 张草案的身份/依赖/覆盖及原文档 Hash。验证原 tracker 未变化、每票失败目录/验收/交付齐全、所有需求/验收 ID 有责任、DAG 无环且没有冗余直接边、文件引用存在。

## 等待的决定

请审阅粒度、真正阻塞关系，以及需要合并/拆分的票。显式调用的 [to-tickets 技能](/Users/cyber/.agents/skills/to-tickets/SKILL.md)在 Quiz the user 阶段要求：“Iterate until the user approves the breakdown.” 因此 review 和全部可审阅草案已完成，正式替换需要确认这份具体拆分。草案仅有发布待确认，未执行的功能/服务证据不填 PASS。
