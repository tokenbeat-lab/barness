# OpenAI Decisions 实施计划

Status: ready-for-agent
Updated: 2026-10-12

关联[spec](spec.md)、[详细设计](design.md)、[验收矩阵](acceptance.md)、[review 与迁移记录](ticket-review/README.md)、[拟议 ADR](../../docs/adr/0026-barness-ai-openai-decisions-classification.md)。用户已批准 13 张工单并替换原 9 张；本文件记录实施依赖及交付，不代替实际工单完成状态。

## 1. 实施原则

每切片先列全部失败方式、先写公共 E2E、保存可归因的红色证据，再生产实现。Client.Classify / HookedClient.Classify 是主要执行 Seam；目录发现和完整 artifact 导入是辅助 Interface。旧 Jev 断言不删除或收窄；基线已有失败照实保留。

内部单协议分派直接替换，继续使用 TypeSafe 活跃路径和共享 unary；不新增 registry、SDK、第二套运行时、兼容 wrapper 或双写。必要公共字段随其第一个完整行为切片交付：02 文本，05 图片；新字段落地即失败关闭。生产单文件最多 500 行，协作不回退他人改动。

## 2. 工单与阻塞关系

编号、Blocking 和状态以独立票为准，以下全部使用新编号。Blocked by 表示该票完成所需的前置交付；所有前置 resolved 后才进入实施 frontier，允许提前准备 fixture，但不能绕过验收门槛。

| 编号 | 工单 | Blocked by | 独立交付 |
|---|---|---|---|
| 01 | [固定 Jev 公共行为与可重放回归基线](issues/01-jev-regression-baseline.md) | 无 | 原公共 wire/错误顺序/用量/舍入/回调的基线包 |
| 02 | [打通 Decisions 文本单选与最终请求闭环](issues/02-decisions-text-choice.md) | 01 | 自带目录下文本单选、最终 Hook、严格答案/单题拒答闭环 |
| 03 | [交付 score/bool 混合答案、分布与逐题拒答](issues/03-mixed-answers-and-refusals.md) | 02 | score/bool/mixed、分布与混合/全拒答原子结果 |
| 04 | [证明 Decisions 用量完整性与输入阶梯计费](issues/04-usage-presence-and-input-pricing.md) | 02 | presence、一致性、错误后用量及输入阶梯费用 |
| 05 | [打通分类图片、三重授权与完整资源边界](issues/05-classifier-inline-images.md) | 02 | 有序 inline 三格式图片、三重授权、完整资源边界 |
| 06 | [证明故障、快照隔离与授权撤销行为](issues/06-runtime-isolation-and-rollback.md) | 03, 04, 05 | 共享 runtime 故障/快照/race/Observer 与授权撤销 |
| 07 | [纳入冻结 Pi 的 Decisions 差分与有限差异账本](issues/07-frozen-pi-differential.md) | 03, 04, 05 | 唯一 Pi 1.1.0 的 P10 classify/catalog 差分和有限账本 |
| 08 | [形成分类图片设计负载与可核验容量证据](issues/08-classifier-image-design-load.md) | 06 | 本地/云端分类图片负载、容量读数与 artifact 导入核验 |
| 09 | [取得候选 Decisions 路线自己的有界真实证据](issues/09-candidate-live-evidence.md) | 07, 08 | 固定来源及候选自己的有预算真实服务报告 |
| 10 | [纳入正式目录并形成第十组合独立支持记录](issues/10-formal-catalog-and-support-matrix.md) | 09 | 正式目录/Hash、schema 3、正式 live 和第十组合原子合并 |
| 11 | [交付 Host 迁移示例与独立拒答评估流程](issues/11-host-migration-and-evaluation-tooling.md) | 03, 04, 05 | Host 迁移示例及拒答评估 CLI/fixture/历史快照核验 |
| 12 | [取得 Luna 当前正式目录下的独立中文效果报告](issues/12-independent-chinese-live-evaluation.md) | 10, 11 | 正式 Luna 的独立中文实际效果和完整 evidence |
| 13 | [闭合十组合发布门禁、迁移文档与回滚证据](issues/13-ten-route-release-gate.md) | 12 | 十组合/P10/图片容量/双效果发布合同、真实 verdict 与文档 |

初始 frontier 只有 01；之后为 02。02 完成后 03/04/05 各自可实施；三者完成后 06/07/11 各自可实施。06 放行 08；07 与 08 都完成才放行 09；09 → 10；10 与 11 → 12；12 → 13。后续票保留 ready-for-agent 标签，但未满足阻塞项时不能声明已就绪实施。

图片单选、复杂用量和混合题不互相阻塞；运行时、固定差分和 Host 工具也不强制串行。候选与正式 live 分开，Host fixture 与实际效果分开，外部凭据不足不阻止无需凭据的工具切片。

## 3. 阶段门槛

### A — 基线与文本闭环（01–04）

01 固定旧 Jev；02 完成真实单选请求/最终 guard/严格普通或拒答结果与基础用量。未实现的 score/bool/图片入口在凭据前拒绝，Hook 也不能启用它们。03 补齐三类型、混合/全拒答及完整分布；04 独立证明 presence、唯一 usage 保留和输入阶梯。旧公共 scope 拒绝顺序和目录构造范围保持，不以空 parser 或假成功交付中间路线。

02 后 05 可与 03/04 独立推进。中间宿主自带目录仅声明真实支持子集，不代表 builtin 或 live 支持。

### B — 图片、运行时、固定差分与容量（05–08）

05 同时落地图片 Request/Model/Policy/Binding 契约和实际发送/最终再授权；有效三格式资产、128/129 和 decoded/encoded/whole JSON 边界均证明。06 证明故障、snapshot、mixed/race/Observer、取消与撤销后新旧调用语义。

07 保持唯一 Pi 1.1.0/source pins，Decisions 为 P10 自己的确定性差分，有限 case/path 账本、零 pending，旧 Jev 原比较保持。08 在明确本地/云端负载和预算下加入分类图片量测，并让 artifact 导入核验要求实际新字段。既有压力缺陷不得通过缩小负载或改合同隐藏。

11 的 Host 工具可在此阶段独立实施，用自带目录和 fixture 验证，不要求正式目录或真实 key。

### C — 自身候选与正式支持（09–10）

09 在完整离线阶段 B 后固定来源和候选 catalog，执行自身有界 live。无 key 为 NOT_RUN，厂商失败为 FAIL；不能引用聊天或 Jev 的 PASS，不能为获得拒答/故障循环制造请求。候选报告不更新正式矩阵。

10 在候选实际通过后纳入 builtin、更新正式版本/Hash，并重新形成正式 live；schema 3 一次性迁移保留九行历史，新行初始 NOT_RUN。所有 report producer/consumer 同步，拒绝 schema 2 直接 merge，无运行时迁移 shim。正式 Hash 变化后复核其他路线历史证据，不能重写旧 Hash/时间为当前。

### D — Host 效果与发布闭合（11–13）

11 交付 Host 示例、独立 route/config/report、拒答和 null 指标、旧 Jev 默认与历史 recorded snapshot 核验。12 以正式目录执行 Luna 独立中文评估；拒答纳入总分母而不计正确，COMPLETE 是任务完整性，不是质量保证。

13 固定十组合、P10 P0 与第七必跑差分子集、分类图片容量、两条路线各自当前真实效果及全部审计。完整正式 gate 缺证据即拒绝；不自动部署。撤销证据由 06 提供、最终实际导入；文档和 proposed ADR 仅按实装及真实证据更新交付。

## 4. 需求追溯

各票 Requirements 表示参与子范围，验收矩阵保留精确需求与 OD 关联；一票列出的所有 OD 不自动证明该票的每一需求。一个 OD 分属多票时分别证明单题/混合、入口/Hook、fixture/live 或 artifact 导入等适用变体。

| 需求 | 交付工单 | 公共证据 |
|---|---|---|
| R01 | 02 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R02 | 01, 02, 03, 07 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R03 | 02, 05, 11 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R04 | 02, 03, 11 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R05 | 02, 05, 10 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R06 | 05, 08 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R07 | 05, 06 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R08 | 02, 03, 05, 07 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R09 | 02, 03, 05, 06 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R10 | 02, 03, 06, 11 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R11 | 02, 03, 04 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R12 | 04, 09 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R13 | 02, 06, 08 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R14 | 01, 02, 05, 06 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R15 | 01, 07 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R16 | 09, 10 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R17 | 08, 09, 10, 13 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R18 | 11, 12 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R19 | 01, 02, 03, 04, 05, 06, 07, 08, 09, 10, 11, 12, 13 | 以验收矩阵中本需求对应 OD 的实际变体为准 |
| R20 | 06, 13 | 以验收矩阵中本需求对应 OD 的实际变体为准 |

## 5. 证据、状态与验证

每票保存先红后绿命令、失败原因、fixture/配置、捕获、消费结果/错误、Attempt、Observer、assertions、资源读数、版本/Hash、完整 SHA-256 manifest、审计及重放命令。旧失败不能覆盖；artifact 仅保留声明的合成或固定测试数据，不含真实用户正文/秘密。

09/10 的固定 live 预算和 12 的独立评估预算继续使用设计 §10。一个进程一个组合/key，零 Binding retry，失败及预留入账。省略 usage 不是免费，目录 Hash 不能保证无日期 alias 的服务权重固定。没有实际运行则 NOT_RUN，不勾选完成。

按真实工单证据标 resolved；全部阶段 D 满足才写功能完成。历史报告可用其记录时快照核验，但当前发布需正式 schema/Hash 下适用证据。正常与 race 发布运行均启用差分/压力，两种 vet 保留，十组合与双效果逐包审计。

当前只发布了文档工单，没有生产实现、付费 API 或发布 PASS。运行[文档核验](planning-verification.md)检查 13 票与批准草案一致、DAG、58 个验收归属及链接。

## 6. 旧票迁移及风险

原 9 张已从正式 issues 移除，原文/Comments/Hash 保存在 review 的历史快照；父 spec 不修改。原始编号只出现在来源/历史记录，新实施均使用新编号。

| 原票 | 新票 |
|---|---|
| 01 | 01 |
| 02 | 02, 05 |
| 03 | 02 |
| 04 | 03, 04 |
| 05 | 05 |
| 06 | 06, 07, 08 |
| 07 | 09, 10 |
| 08 | 11, 12 |
| 09 | 06, 08, 13 |

执行风险与具体补漏见 review：P10 未进入必跑差分、图片量测未被门禁读取、全拒答/历史快照、全局 schema producer/consumer。13 必须用删项/缩集/错身份/Hash/预算/null/缺审计等反例证明失败关闭，不能靠追溯映射或文本声明 PASS。
