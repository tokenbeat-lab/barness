# 05 — 三协议隔离、失败重试、回调和资源释放

Status: resolved
Type: task
Updated: 2026-10-18
Blocked by: 02, 03, 04

## Context

[交付规格](../spec.md)、[研究](../research.md)、[任务图](../map.md)。本票继承完整 spec，不以摘要缩减合同。

## Scope and acceptance

以公共 Client E2E 验证跨租户/账户/用途、持久化信任、回调顺序/授权/冻结、永久与瞬态错误、预算精度、取消与全部既有资源策略。复用现有 suite，加入旧 Provider 并发保持原负载与 ≤75% 标准。先登记失败、再 red→green。

追溯：R03、R19–R23；A01–A02、A15–A18；US09–US15、US39–US49、US54–US68。

## Execution

在自己的分支/worktree 从 `codex/minimax-cn` 当前 tip 开始；使用 tdd 技能，测试接缝已由 spec T01 确认。先登记失败方式，再逐垂直切片 red→green，产出可重放合成请求/SSE/事件/Result/Attempt/Observer 与脱敏/来源记录。实现前阅读相关 ADR。不能回退其他工作；单文件 ≤500 行。完成后合入集成分支 tip，在 Answer 附提交、验证及证据指针；真实缺口保持可见。


## Answer

共享生命周期已完成离线验收。Chat 改用显式 `ThinkingSwitchOnly`，M3.1 即使 Host 只
保留 high 仍发送对应 effort，M3 即使 mandatory 仍只有开关语义。MiniMax 的三种请求
回调 panic 均转为私密、可诊断的 callback_failed；HTTP/SSE 不回显任意正文，保留实际
HTTP 状态和文档枚举的 Messages 类型。已知永久 HTTP/消息类别不能被强制 retry 头
放大；未知429、瞬态408/409/5xx仍只走 Binding 的有界初始重试，不猜 base_resp/词义。
精确预算校验已直接改名为公共内部 `outputBudgetNumber`，无旧名 shim。

202 个自身公共 E2E 覆盖六候选身份、13种资源/超时模式、取消与 sibling工具续接、
重复并发失败释放、admission/等待者上限、授权先于秘密读取、冻结回调/目录/凭据/重试
字节、同名租户/账户/用途、原生持久化与跨协议降级、Observer错误/panic和隐私。
202个自身race通过，`go vet ./ai ./ai/internal/testkit/minimax`通过。受影响协议广泛回归
3311用例通过；最终已提交来源 `4ca7a11` 再验270用例（202自身+66旧GLM+2原负载）通过，
所有保留成功包脱敏审计零发现。详见[可重放证据与校验和](../evidence/ticket-05/README.md)。

原本地4/云端8完整饱和批次保持不变；新增本地3/云端6混合批次轮换全部旧route和MiniMax
三协议，维持128KiB历史/16KiB输出和256/512MiB预算。原批次permit为有意100%；新增
批次每项报告限额（含permit）占用≤75%，记录实际内存/峰值，全部资源归零。

兼容协议余额/MPlan永久错误的未知形状及真实 trace/header/body ID 优先级仍按证据边界
由10候选捕获收敛，未使用原生数字错误或自由文本猜测，也未声称本票提供真实支持。
