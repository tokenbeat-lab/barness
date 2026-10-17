# 07 — 独立 Live Harness、预算与可信证据工具

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: 02, 03, 04

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

先固定六身份场景、研究问题和候选/正式预算；每进程单身份/用途/Key。实现独立 live tag harness、阶段累计预算收据、原始请求/SSE捕获、公共Client场景、可信来源/脱敏/适用性与离线重放/导入校验。复用既有 recorder/Host/evidence 边界，不复制GLM两型号进程假设。用受控进程与篡改失败证明全部证据工具；缺合法配置输出NOT_RUN，candidate/controlled/replay不能正式PASS。本票只建工具并以受控E2E验证，不运行厂商收费调用。05/06完成且执行源码冻结后，10负责实际候选、正式目录和发布门禁，避免付费捕获后再改执行源码。

追溯：R06、R18、R25–R27；A02、A15、A20；US15–US18、US50–US53、US66–US76、US80。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
