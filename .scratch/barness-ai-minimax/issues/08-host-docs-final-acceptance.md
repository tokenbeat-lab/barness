# 08 — Host 文档、ADR 和回滚实现

Status: resolved
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

## Answer

2026-10-18：`43970b5` 修复旧 GLM rollback writer 的 stdout/argv 路径脱敏并保留原审计；
`8dc7c9b` 完成本地/云端 MiniMax Host、三 Binding/用途与回滚、双语指南和 ADR 一致性。
最终合入集成验证源码为 `055a55d`。Host 接口只消费现有三个实际操作，Client 与
HookedClient 直接满足；localassembly 显式 Catalog 由 Client 独立复制，没有重复装配层。

[证据与复跑](../evidence/ticket-08/README.md)：79 个自身 Host cases 和88个受影响旧 Host
cases 普通/race PASS；MiniMax 完整软件回滚、原 GLM 回滚普通/race PASS，全部最终
审计 findings 为空，三个受影响 example 包 vet PASS。保留原基线审计失败及四个 red→green
切片，归档后核对2335个 artifact Hash。回滚恢复固定 ecfd4b4 完整7039文件、目录/价格/
矩阵/差分/发布合同；旧路线可用、MiniMax 不存在、拒绝零凭据/请求，用户目录及未跟踪工作
前后一致。所有本票代码/指南文件≤500行。

本票按拆分后的离线范围完成，没有真实请求或正式支持提升。指南明确 M3 实际正式状态
等待10，M3.1 资格/价格/续接证据独立；09 在10完成后同步最终支持状态并执行全套验收。
