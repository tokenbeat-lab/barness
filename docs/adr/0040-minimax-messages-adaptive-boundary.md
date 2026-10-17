---
status: accepted
date: 2026-10-18
---

# MiniMax Messages adaptive 与原厂扩展分离

MiniMax 的 adaptive 开关和逐请求 effort 复用 Messages adapter，但不具有 Claude 的
显示、budget、beta、managed per-turn effort、缓存断点、eager tool streaming 或 browser
access 语义。协议能力将 adaptive 与原厂扩展拆开；型号 compat 只能收紧已实现的形状，
不能顺带激活原厂字段。默认省略保持服务默认；显式 Temperature 在 adaptive 时仍发送。

Model 是可信 Host 的能力声明，正式和候选快照固定已知型号的事实，adapter 不按型号 ID
硬编码。RequiresThinking 只约束开关，ThinkingSwitchOnly 明确区分二元开关与深度控制，
两者独立于支持档位 map。Host 把 M3.1 收窄到 high-only 时仍发送 effort high；把 M3
收紧为必须开启时仍不得发送 Effort。两种真实型号证明，不能用 mandatory 或档位数量
推断另一种语义。新字段是值字段，JSON、Hash 和 Client 快照保留它，零值保持旧型号行为。

MiniMax 自己的最终请求和 SSE 边界严格验证；原厂及 GLM 的既有 wire/排序保持。函数
参数与 schema 内的同名业务属性不作为扩展或文件引用扫描。最终输出预算按 JSON 精确
整数判断；完成块的工具参数必须为完整对象，不用原厂宽松修复把畸形流提升为成功。

签名续接继续使用 Native State Envelope。当前只放行可信签名 thinking；unsigned 可见
思考沿既有规则降级为文本，未验证的 opaque/redacted 结果保留原形并拒绝原生回放。
不借用 GLM 的无签名例外。合成 fixture 仅证明本地协议守卫，实际签名和 start input /
partial_json 组合须由工单 10 的对应型号真实证据单独确认；新证据改变规则时同步本 ADR
与对应 fixture，不能由另一 Provider 或另一个型号的结果推断。

[规格 R05/R15/R16/R20](../../.scratch/barness-ai-minimax/spec.md)、
[工单 04](../../.scratch/barness-ai-minimax/issues/04-messages-protocol.md)、
[失败与重放证据](../../.scratch/barness-ai-minimax/evidence/ticket-04/README.md)。
