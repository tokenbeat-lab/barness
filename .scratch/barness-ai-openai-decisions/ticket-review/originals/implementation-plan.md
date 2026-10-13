# OpenAI Decisions 实施计划

Status: ready-for-agent
Updated: 2026-10-12

关联[spec](spec.md)、[详细设计](design.md)、[验收矩阵](acceptance.md)、[拟议 ADR](../../docs/adr/0026-barness-ai-openai-decisions-classification.md)。本文件负责实施依赖、交付和证据；每个工单独立文件，不用本计划代替工单状态。

## 1. 实施原则

每切片先写公共 E2E 的失败表达，看到可归因的红色证据，再实现；现有 Jev tests 保持。基线工单只固定现状，不因已有用例通过将新功能标通过。通过最高公共 Seam 验证，不为减少困难直接调用私有 parser。

内部过时路径直接替换，TypeSafe 的活跃实现继续使用；不用兼容 wrapper、双写或第二套 runtime。公共新增省略默认是明确合同，不能被误认为内部 shim。协作提交按职责处理，不回退他人修改，生产单文件最多 500 行。

## 2. 工单与顺序

| 编号 | 工单 | 依赖 | 独立交付 |
|---|---|---|---|
| 01 | [Jev 行为基线与失败目录](issues/01-jev-baseline-and-failure-contract.md) | 无 | 精确回归资产、原 Code/Phase/顺序/用量/舍入/回调/目录来源 |
| 02 | [公共契约与安全能力默认](issues/02-public-contracts-and-safe-defaults.md) | 01 | Request/Options/Refusal、Model/Policy/Binding 新字段，旧省略配置继续成立 |
| 03 | [单选协议、授权与最终请求](issues/03-text-protocol-and-final-payload.md) | 02 | fixed route、choice 原生编码/guard/freeze/严格结果、单题 refusal 与基础用量；一次公共调用闭环 |
| 04 | [混合答案、拒答与用量计价](issues/04-answers-refusals-and-usage.md) | 03 | 补齐 score/bool/mixed、分布、混合与全拒答、完整 presence 和阶梯费用 |
| 05 | [图片输入、授权和资源边界](issues/05-classifier-images-and-authorization.md) | 04 | inline 三格式、图片数量/bytes/clone、入口及 Hook 再授权 |
| 06 | [生命周期、隔离、Pi 差分与压力](issues/06-runtime-isolation-and-differential.md) | 05 | 原 runtime 复用证明、所有差分 extension、race/压力与 Jev 兼容全量 |
| 07 | [独立 live 与正式目录/矩阵](issues/07-live-catalog-and-support-matrix.md) | 06 | 候选来源/自己的 live、纳入正式 catalog、正式 hash 报告、第十组合 |
| 08 | [Host 示例与独立中文效果](issues/08-host-examples-and-chinese-evaluation.md) | 07 | 明确拒答迁移、两路线评估隔离、旧 Jev 默认与历史 verifier 保留 |
| 09 | [发布门禁、文档与回滚](issues/09-release-gate-docs-and-rollback.md) | 08 | 十组合完整合同、两路线效果完整性、文档/ADR/traceability、真实发布 verdict |

链式依赖避免半实现功能进入目录。03 的真实单选切片有完整正常/拒答/失败响应，不能用空 parser 或 success stub；尚未实现的 score/bool/图片在 capability 阶段明确拒绝且不读凭据，04/05 各自补齐后替换阶段性支持集合。中间自带目录不是正式全能力目录。工程可以准备后续 fixture/文档，但提交就绪仍按 Dependencies；ready-for-agent 不等于依赖已经 resolved。

## 3. 阶段门槛

### A — 公共契约与完整文本调用（01–04）

全部旧 Jev regression 通过、旧入口拒绝顺序保持；新 route 可用宿主合成目录完成三个题型/混合/拒答/用量。仅离线不纳入内置 Luna，不报 live/图片支持。通用 question 数量上限不扩张，source/source-map 仍唯一 Pi 1.1.0。

### B — 完整图片与运行时证明（05–06）

图片三个授权条件与独立资源政策、最终回调、有效测试资产/128边界、取消/准入/重试/body limits/tenant 并发有效；旧 Jev 和其他 Operation 不回归。差分零 pending；extension 各自有有限理由/path/cases 和真实公共断言；压力不通过调低负载或预算掩盖故障。

### C — 候选与正式服务验收（07）

先 candidate、后 source+自身能力验收、再版本化内置 Catalog，最后正式 Hash 下的完整新路线报告；不能循环要求内置目录才能测 candidate，也不能拿 candidate 通过替代正式报告。旧矩阵历史状态保留，正式目录范围/价格 Hash 的变化使旧证据进入适用性复核。

first-party 官方和固定 Pi 元数据进入 test fixture，注明日期/hash 和支持子集；没有 account key/region alias 时 NOT_RUN，型号不可达/厂商失败为 FAIL。功能实现与服务证据完成分别记录，不把不足转换为 UNSUPPORTED。

### D — Host 效果与正式发布（08–09）

旧 Jev 工具用法、评估默认配置和历史 bundle 验证不回归；新 Luna 以独立协议/权限/凭据/配置/报告运行同标注任务。合法拒答可完成评估但不能计为正确分类；不为通过设置或降低业务准确率门槛。正式 gate 固定十组合和必须 artifact，缺证据停止发布。

## 4. 需求追溯

| 需求 | 主交付工单 | 公共证据 |
|---|---|---|
| R01 | 02,03 | OD02–04, OD16, OD48 |
| R02 | 01,06 | OD01,09–12,16,29,34,39,50 |
| R03 | 02,03,08 | OD04,09,15,56 |
| R04 | 03 | OD09–14 |
| R05 | 02,07 | OD01,05,06,20,52 |
| R06 | 05,06 | OD07,17–21,26,47 |
| R07 | 02,05,09 | OD08,21,23,24,45,58 |
| R08 | 03,05 | OD11–17,25 |
| R09 | 03,05,06 | OD22–26,45 |
| R10 | 04,08 | OD28,29,41,54–56 |
| R11 | 04 | OD27,29–36,39,43 |
| R12 | 04,07 | OD35–38,51,52 |
| R13 | 06 | OD40–43,45,47 |
| R14 | 02,06 | OD02,06,22,44,45 |
| R15 | 06 | OD14,48–50 |
| R16 | 07 | OD51,52 |
| R17 | 07,09 | OD53,54,57 |
| R18 | 08 | OD55,56 |
| R19 | 06,09 | OD46,47,57 |
| R20 | 09 | OD58 和文档/ADR/合同核验 |

## 5. 证据、状态和交接

每工单保存先红后绿命令、失败的原因和实际 artifacts 指针，勾验收条目时给可核验结果。共用实际 evidence bundle 基础设施，feat 笔记只引用包，不复制 secret/body；原始失败不能覆盖。只修改已执行工单为 resolved，整个功能满足 D 阶段才写完成结论。

候选/正式 live 的预算在设计 §10 固定；Host 的元数据/代码版本变化后核对旧证据是否可用，不修改历史报告为新 Hash。如果适用性不能建立，就形成新正式报告；全量旧九组合要求按 gate 契约办理，不能豁免 Jev。

发布前检查：public Interface/document 与 E2E 一致；目录/价格/shape 证据有来源；scope/metadata/observer 审计完整；Pi 固定 pins；十组合完整报告和独立效果；所有 file hash、source refs、rollback smoke 可复现。

## 6. 风险与明确处置

| 风险 | 处置 / 工单 |
|---|---|
| Jev 的 scope 错误顺序被重构改变 | 01 固定 public baseline；02–06 持续对照，不新增兼容 fallback |
| 新字段使旧目录构造失败或默许图片 | 02 定省略态、旧字段接受范围、新权限 false；fixture 全覆盖 |
| 真实响应精度不符合初始严格校验 | 04/07 先记录 wire 回放红例，限定 Decisions 调整，不扩大 Jev 容差 |
| 大 context/score API 边界未知 | 声明本期子集，不启用无界输入，不为常规验收主动发大 token 请求 |
| live 缺权限/账单/区域形状 | 07 记录 NOT_RUN/FAIL 或未观察项；不纳入未验收范围，不伪造服务承诺 |
| 目录 Hash 变化使历史真实报告过期 | 07/09 核验适用性和正式重跑，不重新标记旧时间 |
| nil error 被宿主当作普通答案保证 | 08 明确 RefusalAnswer 分支、包含拒答的任务完整性和新结果迁移 |
| 十组合或第二效果报告被 gate 忽略 | 09 固定合同反例，删行/缩集/错 model/hash 必须失败 |
