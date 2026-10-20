# 03 — MiniMax Responses 无状态流、工具与可信续接

Status: resolved
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

## Failure inventory (before implementation)

Public Client seam T01, synthetic official schema only. Initial failures: default requests derive OpenAI store/include/none and minimum16; full unsupported efforts can reach credentials; zero sampling, unintegrated controls, mandatory off and precision-losing budgets can reach wire. Subsequent slices cover authoritative/done-only/terminal-only output, malformed or misassociated/duplicate/post-terminal events, EOF without terminal, reasoning-only length, opaque ID collisions and trusted-history projection. Raw output reasoning content must survive in Native State; content→summary is a lossless schema projection hypothesis, not proof of live accepted continuation. Ticket07 owns actual M3 verification and correction if rejected.

## Answer

Implemented independent MiniMax Responses capabilities, omitted defaults, exact budgets, full/simple effort enforcement and final post-hook controls. Added authoritative array SSE/lifecycle validation, raw reasoning preservation, documented lossless candidate input projection and opaque lossless tool IDs with shared trust downgrade policy. Public Client cases cover both candidate models, four entries, tool round trips, cache routing, malformed streams, tenant/account guards and old Responses regressions.

Verification: `go test ./ai/e2e -run 'Responses|DeepSeek|MiniMax' -count=1`, targeted MiniMax race, and `go vet ./ai/...` PASS. See [curated red/green evidence](../evidence/ticket-03/README.md) and [ADR-0039](../../../docs/adr/0039-barness-ai-minimax-responses.md). Whole-suite work is owned by root/final ticket.

No live claim: the schema-based input reasoning projection must independently pass actual M3 thinking/tool continuation in ticket07; M3.1 remains a separate candidate. Actual diagnostics and shared failures/resources remain05/07. No formal catalog admission was made.
