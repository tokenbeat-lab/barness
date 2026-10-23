# 10 — 国际站Responses流内引擎过载阻塞候选续验

Type: task
Status: resolved

## Context

- [唯一当前状态](../spec.md#唯一当前状态)
- [真实结果03与脱敏请求/响应](../evidence/candidate-live-result-03.md)
- [范围04](../evidence/candidate-live-review-04.md)
- [测试SOP](../../../docs/agents/verification.md)

## Scope

35000ms原生间隔已落实、审查及14071必需普通验收；国际站Chat9/9及自身资格PASS。Responses第4项HTTP200内SSE error.rate_limit_exceeded明示engine overloaded，随后response.failed、usage=null。范围首失败停止，Responses剩5项和Messages9项NOT_RUN。不能通过吞掉失败、把HTTP200当成功、补造usage或自动重放已开始的流消除本问题。

## Acceptance

由账户/服务维护者确认服务可用时机，并明确选择新的付费范围/预算/截止时间；使用新grant，Responses与Messages各从头完整9项，零重试/首失败停止，当前来源和自身消费者有效。旧Chat当前来源完整证据仍保留；如再修SDK/原生来源，重新评估所有资格，不能沿用旧source。实际状态只维护在spec。

## Comments

2026-10-23：本轮13调用/13Attempt，预算预留US$0.964002非账单，两grant消耗，4份账本原件保留。实际最小间隔≥35秒；不能由该错误推断更具体后端根因。后续scope尚未批准，未自动再发请求。

2026-10-23：用户明确要求“拉大间隔时间重试，如有类似失败请做相同处理”，已直接授权[范围05](../evidence/candidate-live-review-05.md)。原生60秒/新grant完整九项；同类临时失败可120/240/480秒冷却后续轮，SDK不在已开始流内重放。最多4轮/路线及2小时，合计预留4.50包含上轮0.964002；其他失败/额度/未知/证据问题停止。新普通与原始执行由spec唯一状态记录。

## Answer

2026-10-23：按用户最新直接指令，原生逐调用60秒、新grant完整九项，国际站Responses及Messages各第1轮9/9 PASS，native与自身-qualify均exit0。前轮同native来源Chat9/9仍适用，国际三路线候选已完整通过，未拼接局部PASS。18调用/18HTTP Attempt均200，最小间隔60003.497ms；没有再次429/过载，未触发额外退避或整轮重试。当前源与完整普通14071必需PASS/603输入一致，所有原件、失败及账本保留。见[结果04](../evidence/candidate-live-result-04.md)，唯一状态在[spec](../spec.md#唯一当前状态)。本票国际候选续验目标完成；formal、晋升和其他人工阶段仍由07跟进。
