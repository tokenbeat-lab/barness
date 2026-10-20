# 02 — MiniMax Chat 文本、推理、工具历史与边界

Status: resolved
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


## Answer

Chat 的独立 MiniMax 能力分支和最终请求边界已实现；复用现有 Chat reader，保留
M3 的省略默认与二元 adaptive 开关、M3.1 五档和强制思考；system、max_completion_tokens、
include_usage/reasoning_split 与公共 auto 省略固定，拒绝越界原生能力、精确非法数值与多模态
容器。所有采样/回调之后先授权再验能力，修复中间采样可用且不重选档。多工具独立组装、
冲突身份/终态拒绝，完整 assistant 正文与原生推理续接；晚到用量与 off/缺失/零报告保真。

提交包含 `ai/chat_minimax*.go`、共享 Chat 文件的局部 MiniMax 分支和公共 E2E。
最终 579 个自身用例 + 883 个旧 Chat/DeepSeek/GLM/parity 用例通过，579 个自身 race 通过，
`go vet ./ai` 通过。所有成功包的脱敏审计无发现；[完整 red/green 证据](../evidence/ticket-02/README.md)
保留失败、请求/SSE/事件、Result/Attempt/Usage 和 Observer 生命周期。失败清单与
[来源/候选边界](../../../ai/e2e/testdata/minimax-chat/PROVENANCE.md) 供后续复核。

M3 的真实 reasoning_content/reasoning_details 形状冲突及原生 auto 接受性仍由 07 的
中国区标准按量捕获收敛；合成用例没有被用作真实证明。未知或混合原生推理形状显式失败，
没有扁平化后冒称可回放。05 承担共享错误/trace、回调 panic、并发隔离和资源压力。
