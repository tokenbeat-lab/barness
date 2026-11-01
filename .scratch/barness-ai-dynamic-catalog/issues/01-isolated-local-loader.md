# 逐模型隔离的本地目录加载

Type: task
Status: ready-for-agent
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
