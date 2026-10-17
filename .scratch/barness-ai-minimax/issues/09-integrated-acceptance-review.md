# 09 — 完整验收、双轴审查与交付判定

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 08, 10

## Context

[交付规格](../spec.md)、[任务图](../map.md)。追溯R01–R28、A01–A20、US01–US80。

## Scope and acceptance

执行普通/race公共E2E、两种vet、实际冻结差分、全部原有与新增压力/业务效果、逐包脱敏和完整回滚；汇总US/R/A→实际有限case追溯。保持旧十三组合/十四自身身份的完整当前证据，加入MiniMax三组合/三M3正式身份与M3.1三候选离线状态。依据不可变来源作最终门禁，不重写捕获适用性。

全部实现后按code-review技能对固定基点ecfd4b4运行Standards与Spec并行审查，单一实现subagent修复全部问题，再完成必要重验。只有三份M3正式自身报告、全部门禁和M3.1三协议离线约束完整时才关闭spec与全部票；缺口保留真实状态和可复现指针。清理各实现worktree，报告集成分支。
