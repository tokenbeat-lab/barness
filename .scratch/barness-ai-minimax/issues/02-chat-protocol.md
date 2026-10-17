# 02 — MiniMax Chat 文本、推理、工具历史与边界

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 01

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责 Chat adapter 的 MiniMax 显式能力、full/simple、最终 wire 校验、文本/推理/多工具流、完整同轮历史、晚到用量及错误终态。保留旧 Provider 行为。只支持 spec 子集；分离格式的真实未知不能用合成证据宣称已解决。

追溯：R02–R14、R17–R22；A01–A08、A13–A17；US02–US08、US19–US34、US41–US47、US54–US64。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
