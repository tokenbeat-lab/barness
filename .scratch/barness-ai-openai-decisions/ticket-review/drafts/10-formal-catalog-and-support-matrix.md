# 10: 纳入正式目录并形成第十组合独立支持记录

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 09: 取得候选 Decisions 路线自己的有界真实证据
**Source tickets:** 07
**Requirements:** R05, R16, R17, R19
**Acceptance cases:** OD52, OD53, OD54
**Updated:** 2026-10-12

**What to build:** candidate 实际通过后，Host 能从正式目录发现 Luna classifier；维护者用正式版本重新验证该路线，并经原子报告合并得到第十组合的独立支持记录。

## Responsibility

负责正式 builtin/价格快照、schema 3 报告和矩阵、独立组合 identity/expected/validator、一次性数据迁移及正式 live。旧 Jev validator 不扩大。协作时保留他人改动。

## Failure scenarios first

- candidate Hash 被改写为正式 Hash，chat 价格串入，旧 Jev 省略态被物化。
- schema 2 报告直接 merge；一次性迁移重写九行历史时间、能力或状态。
- 矩阵与报告同时错身份，缩小/重复 expected，跨行批次部分写入。
- 新全局 schema 使其他九组合继续只能产 schema 2 或报错。
- 自己 FAIL/NOT_RUN 借另一路线 PASS 放行，缺图能力降为 UNSUPPORTED。

## Acceptance criteria

- [ ] 先写正式 catalog/source/Hash 与 public merge/process 的错身份、缩集、旧报告、预算和原子性反例。
- [ ] candidate 全部门槛满足才纳入；正式版本/Hash 新生成，能力和独立价格来源准确，旧 Jev 内容及聊天同 ID 均精确保持。
- [ ] 一次性将检入 schema 2 矩阵迁移到 3，九行所有历史状态/时间/Hash 保留，新行 NOT_RUN；无运行时迁移、双写或兼容 shim。
- [ ] 各现有组合报告 producer/consumer 同步使用新 schema；schema 2 报告拒绝 merge，历史包作为原始证据保留。
- [ ] 固定 classifier/openai/openai-decisions/gpt-6-luna 及自己完整 expected 集合，报告与矩阵都错不能相互证明；缺、重复、降级必交能力拒绝。
- [ ] batch merge 任一拒绝时文件不变；NOT_RUN 不更新旧 PASS，FAIL 保留历史通过时间但阻止当前发布。
- [ ] 以正式版本/Hash 重新执行完整新路线 live，沿用候选预算与权限约束；candidate 报告不可更新正式 PASS。
- [ ] 已有九组合是否适用新正式 Hash 明确复核，不能手改历史 Hash/时间；最终完整发布所需重跑由发布票收集。
- [ ] 输出正式快照、一次性迁移前后对照、merge 反例和实际正式 report/manifest/audit。

## Completion

仅自己的当前完整正式报告可标第十组合 PASS。无正式真实报告保持未完成；此票不宣布全十组合发布成功。
