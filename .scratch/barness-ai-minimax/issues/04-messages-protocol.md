# 04 — MiniMax Messages adaptive 思考、工具和签名续接

Status: claimed
Type: task
Updated: 2026-10-18
Blocked by: 01

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责 Messages 显式能力拆分、adaptive/effort 与原厂扩展分离、full/simple、最终 wire、流顺序/签名/工具历史、用量。禁止继承 GLM 无签名例外；没有证据不放宽原厂回放。

追溯：R02–R08、R15–R22；A01–A05、A11–A17；US04–US08、US19–US34、US38–US47、US54–US64。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
