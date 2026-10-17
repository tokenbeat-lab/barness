# 05 — 三协议隔离、失败重试、回调和资源释放

Status: claimed
Type: task
Updated: 2026-10-18
Blocked by: 02, 03, 04

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

以公共 Client E2E 验证跨租户/账户/用途、持久化信任、回调顺序/授权/冻结、永久与瞬态错误、预算精度、取消与全部既有资源策略。复用现有 suite，加入旧 Provider 并发保持原负载与 ≤75% 标准。先登记失败、再 red→green。

追溯：R03、R19–R23；A01–A02、A15–A18；US09–US15、US39–US49、US54–US68。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
