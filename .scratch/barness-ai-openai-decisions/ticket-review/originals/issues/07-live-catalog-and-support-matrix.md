# 07 — 自己的真实服务证据、正式目录与第十组合

Status: ready-for-agent
Type: task
Depends on: 06
Updated: 2026-10-12

## What to build

固定第一方事实 fixture 和 candidate catalog，通过真实公共 Client 的有预算验证后纳入 Luna classifier，再取得正式 hash 报告。新增 openai-decisions-classifier 的矩阵 identity/report/merge 合同。依据 [R12/R16/R17](../spec.md)、[设计 §10](../design.md)、[OD51–54](../acceptance.md)。

## Ownership and context

负责官方 capability/price source fixture、candidate 与 builtin data/catalog snapshot、该 combo 的 live/scenario/budget/report 和独立 supportmatrix validator。Jev 的 expected/counters/model/hash 与历史状态保持；不修改历史报告为新日期/hash。协作不回退他人编辑。

## Failure scenarios first

- 候选不能跑因为内置未纳入；未 live 先列入正式目录；candidate Hash 直接冒充正式 Hash。
- 用 chat context/输出缓存价、造日期 ID、把 2–10/格式子集说成厂商上限。
- 自己未持有 key 却引用 Jev/chat PASS；endpoint override 或另一 combo key 注入进程。
- 将三格式/常规必交缺口标 UNSUPPORTED，少跑能力集/报告改身份也能 merge。
- 目录新 hash 使旧报告不适用，却只替换 hash 或 lastPassedAt。

## Acceptance criteria

- [ ] 先写 source/candidate/merge/report/budget 的 public artifact 失败反例，再实施 live/matrix 接缝。
- [ ] candidate/official model 身份、922000、input-only .1/.2 tiers、说明格式/模态/subset 来源固定有日期/hash。
- [ ] 使用真实官方 endpoint、一个 combo 一个低权限 key/account/region alias，零 Binding retry；预算设计 §10 有计数与超额拒绝。
- [ ] 文本/混合/三格式图片/usage/ID/归属/取消自己的真实证据；不强求线上特定 refusal 或人工制造 429/504。
- [ ] 候选通过后才正式 version/hash；正式报告重新形成，candidate 不能更新正式矩阵 PASS。
- [ ] 矩阵/新报告 schema=3，schema2 矩阵一次性迁移保留九行历史、新增 NOT_RUN 行；schema2 报告拒绝直接 merge，自己 validator 不扩大 Jev validator。
- [ ] source/正规 snapshot 和旧 Jev 内容精确；无新字段省略态被随意 materialize，无日期 alias 风险明确记录。
- [ ] 缺条件诚实 NOT_RUN，厂商故障 FAIL；完整发布还依赖 08/09，不能在本工单宣布正式发布。

## Verification and handoff

候选 public fixture replay/预算反例 → candidate actual live → official inclusion/source/hash checks → official live → 正式矩阵 merge。保存 bounded raw wire、资源/预算、版本与审计；缓存账单或区域费用未观察明确记录。向 08/09 交付正式报告和当前适用性，不手填支持矩阵时间。

## Comments

2026-10-12：本次只是计划，没有新厂商付费调用或正式目录纳入。
