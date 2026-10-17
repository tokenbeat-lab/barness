# 04 — MiniMax Messages adaptive 思考、工具和签名续接

Status: resolved
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


## Answer

`e769362` 实现 MiniMax Messages；已合入集成 tip（验证提交 `f68f38d`）。adaptive/effort
与 Claude 扩展独立，省略默认和显式 Temperature 保留，具名硬约束在读凭据前验证，最终
body 做授权后精确预算/形状守卫。新增 ThinkingSwitchOnly 元数据，二元开关与 mandatory
语义独立；候选快照升级 `.2`，JSON/Hash/Client 快照可追溯，无 ModelID 分支。

自身严格 SSE 顺序/类型和必要字段、独立工具参数、完整 start input、签名与工具历史、
可信续接及降级、opaque 保留/回放拒绝、用量 presence/权威更新与 Observer 均有公共 E2E。
138 own cases + 相同 race 全过，297 原厂/GLM/managed 回归全过，core/vet 通过，≤500 行。

[失败模式与 red/green/race/回归冻结包](../evidence/ticket-04/README.md)、
[ADR-0040](../../../docs/adr/0040-minimax-messages-adaptive-boundary.md)。证据是合成协议
验收；真实签名/start-input/诊断字段与 M3 正式支持仍由 05/07 自身证据收敛，M3.1 不继承
M3 在线结论。不把未测 unsigned/redacted 规则或自动缓存期限升级为支持。
