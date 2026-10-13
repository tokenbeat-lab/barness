# 10: 纳入正式目录并形成第十组合独立支持记录

**What to build:** candidate 实际通过后，Host 能从正式目录发现 Luna classifier；维护者用正式版本重新验证该路线，并经原子报告合并得到第十组合的独立支持记录。

**Blocked by:** 09: 取得候选 Decisions 路线自己的有界真实证据

**Status:** resolved
**Type:** task
**Source tickets:** 07
**Requirements:** R05, R16, R17, R19
**Acceptance cases:** OD52, OD53, OD54
**Updated:** 2026-10-13

## Responsibility

负责正式 builtin/价格快照、schema 3 报告和矩阵、独立组合 identity/expected/validator、一次性数据迁移及正式 live。旧 Jev validator 不扩大。协作时保留他人改动。

## Failure scenarios first

- candidate Hash 被改写为正式 Hash，chat 价格串入，旧 Jev 省略态被物化。
- schema 2 报告直接 merge；一次性迁移重写九行历史时间、能力或状态。
- 矩阵与报告同时错身份，缩小/重复 expected，跨行批次部分写入。
- 新全局 schema 使其他九组合继续只能产 schema 2 或报错。
- 自己 FAIL/NOT_RUN 借另一路线 PASS 放行，缺图能力降为 UNSUPPORTED。

## Acceptance criteria

- [x] 先写正式 catalog/source/Hash 与 public merge/process 的错身份、缩集、旧报告、预算和原子性反例。
- [x] candidate 全部门槛满足才纳入；正式版本/Hash 新生成，能力和独立价格来源准确，旧 Jev 内容及聊天同 ID 均精确保持。
- [x] 一次性将检入 schema 2 矩阵迁移到 3，九行所有历史状态/时间/Hash 保留，新行 NOT_RUN；无运行时迁移、双写或兼容 shim。
- [x] 各现有组合报告 producer/consumer 同步使用新 schema；schema 2 报告拒绝 merge，历史包作为原始证据保留。
- [x] 固定 classifier/openai/openai-decisions/gpt-6-luna 及自己完整 expected 集合，报告与矩阵都错不能相互证明；缺、重复、降级必交能力拒绝。
- [x] batch merge 任一拒绝时文件不变；NOT_RUN 不更新旧 PASS，FAIL 保留历史通过时间但阻止当前发布。
- [x] 以正式版本/Hash 重新执行完整新路线 live，沿用候选预算与权限约束；candidate 报告不可更新正式 PASS。
- [x] 已有九组合是否适用新正式 Hash 明确复核，不能手改历史 Hash/时间；最终完整发布所需重跑由发布票收集。
- [x] 输出正式快照、一次性迁移前后对照、merge 反例和实际正式 report/manifest/audit。

## Completion

仅自己的当前完整正式报告可标第十组合 PASS。无正式真实报告保持未完成；此票不宣布全十组合发布成功。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 07

2026-10-12：本次只是计划，没有新厂商付费调用或正式目录纳入。


## Implementation delivery (2026-10-13)

正式 `classifier/openai/openai-decisions/gpt-6-luna` 已纳入 BuiltinCatalog，
版本 `2026-10-13.1`、Hash
`sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
先验证 09 候选自己的 9/9 真实 PASS、83 项 inventory 和审计，再纳入；
候选 Hash/报告不改写，Jev 的省略态及同 ID 聊天型号精确保留。

公共 Catalog、Merge/CLI、独立进程失败反例先 RED 后实现。一次性迁移 schema 2
到 3，九行身份、能力、状态、时间和 note 中历史 Hash 逐值一致，第十行九项
NOT_RUN 初始化；新路线独立 expected/validator，Jev validator 未扩大。
各旧组合 producer/consumer 同步新 schema；旧报告无法 merge，历史包保留。
NOT_RUN 不覆盖历史 PASS，FAIL 更新当前能力且保留历史通过时间，batch 任一
拒绝时文件字节不变。负 token、预算、错身份、缩集/重复、缺项、图片降级均拒绝。

两轴审查初见 2 项 Standards、后复审见 1 项 Spec，均先写公共反例再修复。
最终 accepting 报告在 audit/inventory 完成、同步之后原子发布；审计前 pending
为独立封套，公共 Report decoder 拒绝，两处 abrupt exit 不遗留可接受 PASS。
最终 Standards 0 / Spec 0。依据新全局目录刷新 Jev 默认评估配置 pin/版本和
usage fixture pin，模型/任务/提示词/预算及历史效果不改变，不宣告 11/12 完成。

沿用 09 的用户授权：`.env:OPENAI_KEY`、127.0.0.1:7890、不能核实的信息如实
unknown。最终执行源码 `f9bd770` 的正式 Hash 真实 9/9 PASS，9 Logical Call /
9 Attempt / 13 题 / 480 state bytes / 3 图 / 1161 decoded bytes，零重试。
八次正常结果实际 Model=gpt-6-luna、usage complete、已报告 input=1608/output=0；
成功响应头后的取消为 aborted/usage unreported，不伪造未知消耗。完整 84 项
inventory、audit、manifest 与实际请求/响应、结果、归属和 Observer 均核验。
只用该正式报告经公共 CLI 更新自己的第十行；最终无 key 9 NOT_RUN、零消费，
CLI merge 后正式 PASS 文件字节不变。

全仓最终检查含 Pi 1.1.0 差分和压力：8234 PASS / 69 SKIP / 0 FAIL，7 个执行包
PASS；新接缝最终 race 53 PASS / 1 SKIP / 0 FAIL；`go vet -tags live ./...` 通过。
SKIP 未解释为通过。首次 full 的旧 Jev 配置和 usage pin 导致 25 FAIL，原始
输出/清单保留；修复后独立完整重跑通过。最终 full Go 源与 f9bd770 一致，
usage fixture 的元数据在 9be8fa5 刷新，完整 case 记录了实际消费的 fixture SHA。

[证据说明、正式报告、迁移对照、反例及收据](../formal-support-evidence/README.md)
和 [两轴复审](../formal-support-evidence/REVIEW.md) 可独立重放。首次与中间正式
运行、对应源码和报告均原样保留，不改时间/Hash 或用它们替代最终正式包。
九条旧路线明确不自动适用于新 Hash，所需重跑由工单 13 收集；此票仅完成自身
正式支持记录，没有宣布全十组合发布成功。
