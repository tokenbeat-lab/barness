# 08 — Host 文档、ADR、完整验收和回滚

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 07

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

完整本地/云端 Host 示例、三 Binding/用途、六身份思考语义、费用/缓存/信任/实际状态；ADR 记录能力拆分和真实续接结论。普通/race E2E、两种 vet、差分、资源/业务效果、逐包审计和完整回滚；更新 US/R/A→有限 case 追溯。全部完成后 code-review 双轴并修复所有问题，按本地 tracker resolved；不能在缺三份 M3 正式自身报告时关闭 spec。

追溯：R01–R28；A01–A20；US01–US80。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
