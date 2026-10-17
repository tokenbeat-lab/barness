# 01 — Provider 身份、六个候选型号与思考元数据

Status: ready-for-agent
Type: task
Updated: 2026-10-18
Blocked by: None

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责共享 Provider、型号元数据及候选目录。建立零值兼容、序列化/复制/Hash、思考默认标签与 mandatory/binary 语义；不要提前发布未经自身 live 验收的 BuiltinCatalog。三协议条目独立，候选合成费率明确标记。公共 Client E2E 先列失败方式、保留 red/green 证据；为后续协议实现写明共享接口。

追溯：R01–R08、R18；A01–A05、A15；US01、US09–US26、US50–US53。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
