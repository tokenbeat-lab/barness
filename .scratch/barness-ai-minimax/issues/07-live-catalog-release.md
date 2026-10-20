# 07 — 独立 Live Harness、预算与可信证据工具

Status: resolved
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


## Answer

独立六身份 MiniMax harness、固定13场景、可选3-call研究前沿、工具不足仅重跑一次、单身份与阶段累计持久预算、原始请求/SSE/事件/Result/Attempt/Observer/资源记录、固定来源/源码完整集合、脱敏与密封校验、无网络重放及支持报告导入工具已完成。正式入口要求当前 BuiltinCatalog 自身M3；candidate/controlled/replay不能导入正式PASS。SDK Messages 的`?beta=true`仅记录传输事实，未开放beta功能。

提交：f96faef，合并集成tip164504e，SDK精确绑定de9de39，完整重放来源1dd2330。普通及race共有45受控case产物PASS；辅助ledger/supportmatrix race和live vet通过；独立CLI重验受控包通过。[失败清单、重放命令与当前缺口](../evidence/ticket-07/README.md)、[六身份/重跑/研究/重放原始档案](../evidence/ticket-07/controlled-captures.tar.gz)、[摘要](../evidence/ticket-07/controlled-summary.json)。

本票没有读取.env/真实Key或执行厂商收费调用。M3实际三协议形状、价格正式目录、自身正式报告及最终发布合同仍归10，M3.1线上权益仍独立未知；未将这些缺口改写成正式通过。
