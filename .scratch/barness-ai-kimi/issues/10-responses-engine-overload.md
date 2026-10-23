# 10 — 国际站Responses流内引擎过载阻塞候选续验

Type: task
Status: ready-for-human

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
