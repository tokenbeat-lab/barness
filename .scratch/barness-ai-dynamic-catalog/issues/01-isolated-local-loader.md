# 逐模型隔离的本地目录加载

Type: task
Status: claimed
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
