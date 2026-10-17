# 03 — MiniMax Responses 无状态流、工具与可信续接

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 01

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责 Responses adapter 的显式能力、full/simple、最终 wire、SSE 终态/权威输出、原生历史、不透明 ID 和缓存路由。必须保留原始 reasoning，输出/输入不对称问题依来源记录；未真实验证不得标正式续接通过。

追溯：R02–R08、R12–R14、R17–R22；A01–A05、A09–A10、A13–A17；US02、US05–US08、US27、US29–US37、US39–US52、US54–US64。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
