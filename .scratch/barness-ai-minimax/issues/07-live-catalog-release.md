# 07 — 六身份候选、M3 正式证据和目录发布门禁

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 05, 06

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

先固定六身份场景、研究问题和候选/正式预算；每进程单身份/用途/Key。复用 live tag、可信来源/捕获/脱敏/适用性机制；缺 Key/权益记录 NOT_RUN。M3 只有通过自身候选与可靠 RMB→USD 快照后才能进入正式目录，三正式报告绑定正式 Hash。M3.1 缺可靠价格保持候选。增量加入三组合，不削弱既有十三组合/十四身份门禁。

追溯：R06、R18、R25–R27；A02、A15、A20；US15–US18、US50–US53、US66–US76、US80。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
