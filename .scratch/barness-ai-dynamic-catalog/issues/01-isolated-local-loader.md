# 逐模型隔离的本地目录加载

Type: task
Status: resolved
Blocked by: none

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责 catalog/ 的公共加载入口、报告及 ai 中复用约束的最小公共校验边界。严格验证完整原始 hash、顶层结构，逐模型严格解码、通用身份及实际协议约束，重复组全部失败，成功集合保持顺序/来源版本并计算有效 Hash。不得修改默认目录迁移文件。

## Acceptance and verification

T08、T09、T10、T12；公共加载 → 公共 Client 的受控调用、报告及副本隔离。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

已确认受测 seam：catalog.Load/LoadFile → Config.Catalog → 公共 Client。失败方式：完整 JSON/顶层/hash 损坏必须整份拒绝；单条未知字段、错误类型/值、未知协议/能力、身份缺失及非对象必须隔离；重复完整身份的所有成员必须失败；全失败返回完整报告和空集合；原始摘要覆盖未知字段且仅移除外部空白；装配后来源/报告/查询变化不能改变调用或权限。


## Answer

已实现公共本地加载/报告、ai 复用校验和 T08/T09/T10/T12 的受控公共调用场景。最终定向普通回执 PASS/exit0，所有新增 CAT case 与 audit 见 [验证记录](../loader-verification.md) 和 [不可变索引](../loader-verification-20261101.json)。初始全包回归因历史来源适用性严格拒绝共享 runtime 文件的等价提取而 FAIL/exit1，原件未覆盖；集成 05 继续处理有限来源审查，不能将定向通过称为全包或发布通过。

已合并当时集成 tip 后测试。厂商事实、Binding 权限及历史 pin 未改写；本工单没有执行 live/race/压力。已知 compat 的实际接纳与来源价格完整性仍在维护/候选合约边界，详见验证记录的适用性说明。
