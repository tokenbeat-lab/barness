# 06 — 固定 MiniMax pi 差分和扩展证据

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 02, 03, 04

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

冻结 pi 1.1.0 pins 不变；装配 MiniMax Messages M3 catalog 并运行适用文本/工具差分，逐 case/path 解释差异；Chat/Responses 和 M3.1 登记扩展而非伪差分。旧差分覆盖不减少。

追溯：R05、R24；A19；US71–US75。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
