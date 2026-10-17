# 08 — Host 文档、ADR 和回滚实现

Status: claimed
Type: task
Updated: 2026-10-18
Blocked by: 06

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

负责完整本地/云端 Host 示例、三 Binding/用途、六身份思考语义、费用/缓存/信任与如实支持状态；ADR 记录能力拆分和有来源的续接结论。实现并先验完整软件回滚、禁用 Binding 零凭据/请求和旧路线恢复。修复基线既有 GLM 回滚 commands.json 绝对路径泄漏，不减弱脱敏审计。与共享验证/证据工具并行；10结果到达后同步真实支持状态。最终全集验收与工单关闭由09汇合。

追溯：R01–R28；A01–A20；US01–US80。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。
