# 全规格集成、双轴审查及普通验证

Type: task
Status: ready-for-agent
Blocked by: 01, 02, 03, 04

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责集成收尾与公共端到端缺口、用户文档、唯一当前状态和可重复证据。以初始提交为固定点按 code-review 技能双轴审查，统一修复，再冻结普通来源并执行完整受影响普通包/构建/vet，记录真实 SHA、哈希和未执行阶段。不得把受控通过写成厂商资格或完整发布通过。race/压力适用时仅列具体人工选择请求，未授权不执行。

## Acceptance and verification

T01–T13；普通证据、全部必需实际场景及 audit，状态/索引一致。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。
